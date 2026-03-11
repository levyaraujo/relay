# CLAUDE.md — github.com/levyaraujo/relay / relay

> Guidance for Claude when working in this repository.

---

## Project Overview

**relay** is a tiny ERP backend written in Go. It is designed to be simple, robust, and adaptable to AI and conversational workflows. It exposes a JSON HTTP API backed by PostgreSQL and follows a clean, package-per-domain structure.

---

## Tech Stack

| Concern         | Choice                              |
|-----------------|-------------------------------------|
| Language        | Go 1.26+                            |
| HTTP            | `net/http` stdlib (no framework)    |
| Database        | PostgreSQL 17 via `pgx/v5` + `sqlx` |
| Migrations      | `golang-migrate` (CLI + SQL files)  |
| Config          | `godotenv` (`.env` file)            |
| UUIDs           | `google/uuid`                       |
| Testing         | `testing` stdlib + `httptest`       |
| Container       | Docker Compose (local dev)          |

---

## Repository Layout

```
.
├── main.go                    # Entry point: connect, migrate, wire, serve
├── shared/
│   ├── model.go               # Base model (Id, CreatedAt, UpdatedAt, Deleted)
│   ├── db.go                  # DB connection singleton
│   └── config.go              # Settings / .env loader (sync.Once)
├── companies/                 # Domain: companies
├── users/                     # Domain: users
├── transactions/              # Domain: transactions
│   ├── entity.go              # Transaction struct + TransactionType enum
│   ├── repo.go                # TransactionRepository interface + implementation
│   ├── controllers.go         # Business logic (validation, orchestration)
│   ├── http.go                # HTTP handlers (routes → controller/repo)
│   ├── migrations.go          # In-process migration helper (shared.DB)
│   └── transactions_test.go   # Integration tests with real DB
├── testutil/                  # Shared test seed helpers
├── database/
│   └── migrations/            # SQL migration files (up/down)
├── docker-compose.yml
├── Makefile
└── .env                       # Not committed — see .gitignore
```

Each domain package is **self-contained**: entity, repo interface, repo implementation, controller, HTTP handler, and migration all live together.

---

## Core Conventions

### 1. Domain Package Structure

Every domain follows this internal layout:

```
<domain>/
├── entity.go        # Structs, enums, constants
├── repo.go          # Repository interface + sqlx implementation
├── controllers.go   # Business logic — calls repo, validates, orchestrates
├── http.go          # HTTP handlers — decodes request, calls controller/repo, encodes response
├── migrations.go    # Programmatic migration (for legacy/fast iteration)
└── <domain>_test.go # Integration tests
```

> **Rule:** HTTP handlers must not contain business logic. Controllers must not know about HTTP. Repositories must not know about either.

### 2. The Repository Pattern

- Define an **interface** in `repo.go`:
  ```go
  type FooRepository interface {
      Create(f *Foo) (*Foo, error)
      GetByID(id uuid.UUID) (*Foo, error)
      Update(f *Foo) (*Foo, error)
  }
  ```
- Provide a concrete `struct` implementation using `*sqlx.DB`.
- Return `(*Entity, error)` from all repo methods — never panic.
- Use `NamedQuery` + `StructScan` for INSERT/UPDATE RETURNING patterns.
- Soft-delete: always filter `deleted = false` in SELECT queries.

### 3. Controllers (Business Logic Layer)

- Accept and return domain structs, never `http.Request`/`http.ResponseWriter`.
- Validate inputs early and return sentinel errors (e.g. `ErrNegativeAmount`).
- Example:
  ```go
  var ErrNegativeAmount = errors.New("amount cannot be negative")

  func (c *TransactionController) CreateTransaction(t *Transaction) (*Transaction, error) {
      if t.Amount < 0 {
          return nil, ErrNegativeAmount
      }
      return c.repo.Create(t)
  }
  ```

### 4. HTTP Handlers

- Use stdlib `net/http` — no framework.
- Route registration via `mux.HandleFunc("METHOD /path", handler)`.
- Parse path params with `r.PathValue("param")` (Go 1.22+ stdlib router).
- JSON in / JSON out. Always set `Content-Type: application/json`.
- Map sentinel errors to HTTP status codes in the handler layer:
  ```go
  if errors.Is(err, ErrNegativeAmount) {
      http.Error(w, err.Error(), http.StatusUnprocessableEntity)
      return
  }
  ```

### 5. Shared Base Model

All entities embed `shared.Model`:

```go
type Model struct {
    Id        uuid.UUID `db:"id"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
    Deleted   bool      `db:"deleted"`
}
```

Always populate `Id`, `CreatedAt`, `UpdatedAt` before inserting. Use `uuid.New()`.

### 6. Database & Migrations

- SQL migration files live in `database/migrations/` as `<timestamp>_<name>.up.sql` / `.down.sql`.
- The in-process `Migrate()` helpers in each domain package use `shared.DB.MustExec` for quick iteration but **SQL files are the source of truth for production**.
- Run migrations: `make migrate-up`.
- Reset DB: `make db-reset`.

### 7. Configuration

- Load from `.env` via `godotenv`. Never hardcode credentials.
- Access settings via `shared.LoadConfig()` (thread-safe `sync.Once`).
- Required env var: `DATABASE_URL`.

---

## Testing Philosophy — TDD First

### Rules

1. **Write the test before the implementation** (red → green → refactor).
2. Tests live in the same package (`package transactions`) — white-box testing.
3. Use **real database** for integration tests. No mocks for the repository layer; mock only at the controller boundary when unit-testing business logic.
4. Use `TestMain` for setup/teardown:
   - Migrate schema.
   - Seed required entities via `testutil` helpers.
   - `defer` cleanup (`DELETE FROM ...`).
   - Call `os.Exit(m.Run())`.
5. Use `httptest.NewRequest` + `httptest.NewRecorder` for HTTP handler tests.
6. Keep `assertStatus` and similar helpers as unexported functions in the test file.

### Test Structure

```go
func TestFooHandler(t *testing.T) {
    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)

    t.Run("create foo returns 201", func(t *testing.T) { ... })
    t.Run("get foo by id returns 200", func(t *testing.T) { ... })
    t.Run("invalid id returns 400", func(t *testing.T) { ... })
}
```

### Running Tests

```bash
make test          # go test ./...
```

Tests require a running PostgreSQL instance. Use `docker compose up -d` first.

---

## Adding a New Domain

Follow these steps **in order** (TDD):

1. Create `<domain>/entity.go` — define the struct embedding `shared.Model`.
2. Create `<domain>/repo.go` — define the interface and stub the implementation.
3. Write `<domain>/<domain>_test.go` — write failing tests first.
4. Implement the repo methods until tests pass.
5. Create `<domain>/controllers.go` — write controller tests, then implement.
6. Create `<domain>/http.go` — write HTTP tests, then implement.
7. Create `<domain>/migrations.go` and the corresponding SQL file in `database/migrations/`.
8. Wire everything in `main.go`.

---

## Code Style & Go Idioms

- **Errors**: return `error` as the last return value; use `errors.New` for sentinel errors; wrap with `fmt.Errorf("...: %w", err)` when adding context.
- **No global state** except `shared.DB` (initialized once at startup).
- **Interfaces**: define them in the consumer package (repo interface lives next to controller that uses it).
- **Exported names only** in public API surface (repository interface, handler, controller constructor).
- **Short variable names** for local scope (`t` for transaction, `r` for request/repo, `w` for writer).
- **No `init()`** functions.
- **Defer `rows.Close()`** immediately after `NamedQuery` calls.
- Run `go vet ./...` and `go fmt ./...` before committing.

---

## Makefile Reference

| Target           | Description                                  |
|------------------|----------------------------------------------|
| `make run`       | `go run .`                                   |
| `make test`      | `go test ./...`                              |
| `make migrate-up`| Apply all pending migrations                 |
| `make migrate-down` | Roll back last migration                  |
| `make migrate-create name=<name>` | Create new migration file   |
| `make migrate-force version=<v>` | Force migration version      |
| `make db-reset`  | Drop all tables and re-apply all migrations  |

---

## Environment Setup

```bash
# 1. Start Postgres
docker compose up -d

# 2. Create .env
echo 'DATABASE_URL=postgres://dev:xhgz0AhrUm88Qydjm7ug@localhost:5432/github.com/levyaraujo/relay' > .env

# 3. Run migrations
make migrate-up

# 4. Run server
make run

# 5. Run tests
make test
```
