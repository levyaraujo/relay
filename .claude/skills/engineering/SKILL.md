---
name: engineering
description: >
  Engineering skill for the relay Go ERP project. Use this skill whenever
  working on this codebase — adding domains, writing tests, implementing handlers,
  repo methods, controllers, migrations, or reviewing code for correctness and
  Go idioms. Triggers on: new feature, new domain, write test, add endpoint,
  add migration, implement repo, implement controller, review code, fix bug,
  refactor, or any task involving the relay project files.
---

# relay-golang Skill

Go engineering playbook for the **relay** ERP project.
Stack: Go 1.26 · stdlib `net/http` · PostgreSQL 17 · `sqlx` · `pgx/v5` · `golang-migrate`.

---

## Quick Reference

| Layer        | File              | Responsibility                                      |
|--------------|-------------------|-----------------------------------------------------|
| Entity       | `entity.go`       | Struct, enums, constants. Embeds `shared.Model`.    |
| Repository   | `repo.go`         | Interface + `sqlx` implementation. DB only.         |
| Controller   | `controllers.go`  | Business logic. Calls repo. No HTTP knowledge.      |
| HTTP Handler | `http.go`         | Decode → call controller/repo → encode. No logic.   |
| Migration    | `migrations.go`   | In-process `shared.DB.MustExec` schema helper.      |
| SQL Files    | `database/migrations/` | Source of truth for production schema.         |
| Tests        | `<domain>_test.go`| Integration tests, real DB, `TestMain` lifecycle.   |

---

## Workflow: Adding a New Domain (TDD Order)

Always go in this order — **write tests before implementation**.

```
1. entity.go         → define struct + enums
2. repo.go           → define interface (stub implementation)
3. <domain>_test.go  → write FAILING tests
4. repo.go           → implement until tests pass
5. controllers.go    → write controller tests, then implement
6. http.go           → write HTTP tests, then implement
7. migrations.go     → in-process migration helper
8. database/migrations/<ts>_<name>.up.sql + .down.sql
9. main.go           → wire and register routes
```

---

## Patterns to Follow

### Entity

```go
// entity.go
package foos

import (
    "github.com/levyaraujo/relay/shared"
    "github.com/google/uuid"
)

type FooType int

const (
    FooTypeUnknown FooType = iota
    FooTypeBar
)

type Foo struct {
    shared.Model
    CompanyId uuid.UUID `db:"company_id"`
    Name      string    `db:"name"`
    Type      FooType   `db:"type"`
}
```

### Repository

```go
// repo.go
type FooRepository interface {
    Create(f *Foo) (*Foo, error)
    GetByID(id uuid.UUID) (*Foo, error)
    Update(f *Foo) (*Foo, error)
}

type repository struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) FooRepository { return &repository{db: db} }

func (r *repository) Create(f *Foo) (*Foo, error) {
    var created Foo
    rows, err := r.db.NamedQuery(`INSERT INTO foos (...) VALUES (...) RETURNING *`, f)
    if err != nil { return nil, err }
    defer rows.Close()
    if rows.Next() {
        if err := rows.StructScan(&created); err != nil { return nil, err }
    }
    return &created, nil
}

// GetByID must always filter: WHERE id = $1 AND deleted = false
// Update uses RETURNING * and NamedQuery
```

### Controller

```go
// controllers.go
var ErrFooInvalid = errors.New("foo: name is required")

type FooController struct{ repo FooRepository }

func NewFooController(db *sqlx.DB) *FooController {
    return &FooController{NewRepository(db)}
}

func (c *FooController) CreateFoo(f *Foo) (*Foo, error) {
    if f.Name == "" {
        return nil, ErrFooInvalid
    }
    f.Id = uuid.New()
    f.CreatedAt = time.Now()
    f.UpdatedAt = time.Now()
    return c.repo.Create(f)
}
```

### HTTP Handler

```go
// http.go
type Handler struct{ ctrl *FooController }

func NewHandler(ctrl *FooController) *Handler { return &Handler{ctrl} }

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("POST /foos", h.create)
    mux.HandleFunc("GET /foos/{id}", h.getByID)
    mux.HandleFunc("PUT /foos/{id}", h.update)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
    var f Foo
    if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    created, err := h.ctrl.CreateFoo(&f)
    if err != nil {
        // Map sentinel errors to HTTP status codes here
        if errors.Is(err, ErrFooInvalid) {
            http.Error(w, err.Error(), http.StatusUnprocessableEntity)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(created)
}
```

### Tests (Integration, Real DB)

```go
// foos_test.go
package foos

import (
    "bytes"; "encoding/json"; "net/http"; "net/http/httptest"
    "os"; "testing"
    "github.com/levyaraujo/relay/shared"; "github.com/levyaraujo/relay/testutil"
    "github.com/google/uuid"
)

var (
    testHandler *Handler
    testFoo     *Foo
)

func TestMain(m *testing.M) {
    db := shared.Connect()
    Migrate()
    // ... other domain migrations if FK deps exist

    repo := NewRepository(db)
    ctrl := &FooController{repo}
    testHandler = NewHandler(ctrl)

    // Seed test data
    testFoo, _ = repo.Create(&Foo{ /* ... */ })

    defer func() {
        db.MustExec("DELETE FROM foos")
    }()
    os.Exit(m.Run())
}

func TestFooHandler(t *testing.T) {
    mux := http.NewServeMux()
    testHandler.RegisterRoutes(mux)

    t.Run("create returns 201", func(t *testing.T) {
        body, _ := json.Marshal(&Foo{Name: "test"})
        req := httptest.NewRequest(http.MethodPost, "/foos", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        res := httptest.NewRecorder()
        mux.ServeHTTP(res, req)
        assertStatus(t, http.StatusCreated, res.Code)
    })

    t.Run("get by id returns 200", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/foos/"+testFoo.Id.String(), nil)
        res := httptest.NewRecorder()
        mux.ServeHTTP(res, req)
        assertStatus(t, http.StatusOK, res.Code)
    })

    t.Run("invalid uuid returns 400", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/foos/not-a-uuid", nil)
        res := httptest.NewRecorder()
        mux.ServeHTTP(res, req)
        assertStatus(t, http.StatusBadRequest, res.Code)
    })
}

func assertStatus(t testing.TB, want, got int) {
    t.Helper()
    if got != want {
        t.Errorf("status: got %d, want %d", got, want)
    }
}
```

### Migration SQL

```sql
-- database/migrations/<timestamp>_create_foos.up.sql
CREATE TABLE IF NOT EXISTS foos (
    id         UUID PRIMARY KEY,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted    BOOLEAN        NOT NULL DEFAULT FALSE,
    company_id UUID           NOT NULL,
    name       VARCHAR(255)   NOT NULL,
    type       INTEGER        NOT NULL DEFAULT 0,
    CONSTRAINT fk_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_foos_company_id ON foos(company_id);
```

```sql
-- database/migrations/<timestamp>_create_foos.down.sql
DROP TABLE IF EXISTS foos;
```

---

## Rules Checklist

Before finishing any task, verify:

- [ ] No business logic in HTTP handlers
- [ ] No HTTP knowledge in controllers or repos
- [ ] All INSERT/UPDATE set `Id`, `CreatedAt`, `UpdatedAt`
- [ ] All SELECT queries filter `deleted = false`
- [ ] Sentinel errors defined and mapped to HTTP status in handler
- [ ] `defer rows.Close()` after every `NamedQuery`
- [ ] Tests written before or alongside implementation (TDD)
- [ ] `TestMain` handles schema setup and cleanup
- [ ] SQL migration file created alongside code changes
- [ ] `go vet ./...` passes

---

## Common Mistakes to Avoid

| Mistake                              | Correct approach                              |
|--------------------------------------|-----------------------------------------------|
| Logic in HTTP handler                | Move to controller                            |
| Calling `http.Error` in controller   | Return sentinel error, map in handler         |
| Forgetting `defer rows.Close()`      | Add immediately after `NamedQuery`            |
| Not filtering `deleted = false`      | Always add to WHERE clause in SELECT          |
| Skipping `Id`/timestamps on insert   | Set in controller before calling `repo.Create`|
| Using `DB.Get` for NamedQuery inserts| Use `NamedQuery` + `StructScan` for RETURNING |
| Writing tests after implementation   | Write failing tests first (TDD)               |

---

## Detailed References

Read these files when you need deeper context:

- `references/repo-patterns.md` — Advanced sqlx patterns, error handling, NULL columns
- `references/testing-guide.md` — TestMain lifecycle, testutil seeds, table-driven tests
