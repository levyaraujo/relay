# Testing Guide

## Philosophy

- **TDD**: Write the failing test first, then write the minimum code to make it pass, then refactor.
- **Real database**: Integration tests use a real PostgreSQL instance. No mocking the repo layer.
- **Unit tests for controllers**: Mock the repository interface to test business logic in isolation.
- **Clean state**: `TestMain` is responsible for seeding and cleaning up all test data.

---

## TestMain Lifecycle

```go
func TestMain(m *testing.M) {
    // 1. Connect to DB (reads DATABASE_URL from .env)
    db := shared.Connect()

    // 2. Run migrations (all dependent domains first)
    companies.Migrate()
    users.Migrate()
    Migrate() // current domain last

    // 3. Build repos and handlers
    repo := NewRepository(db)
    testHandler = NewHandler(repo)

    // 4. Seed required fixtures using testutil helpers
    company, user := testutil.SeedCompanyAndUser(
        companies.NewRepository(db),
        users.NewRepository(db),
    )

    // 5. Create test entity
    var err error
    testFoo, err = repo.Create(&Foo{
        Id:        uuid.New(),
        CompanyId: company.Id,
        CreatorId: user.Id,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Name:      "fixture",
    })
    if err != nil {
        panic("seed failed: " + err.Error())
    }

    // 6. Defer cleanup (order matters — children before parents)
    defer func() {
        db.MustExec("DELETE FROM foos")
        db.MustExec("DELETE FROM users")
        db.MustExec("DELETE FROM companies")
    }()

    // 7. Run tests
    os.Exit(m.Run())
}
```

---

## HTTP Handler Tests

```go
func TestFooHandler(t *testing.T) {
    mux := http.NewServeMux()
    testHandler.RegisterRoutes(mux)

    t.Run("create returns 201", func(t *testing.T) {
        payload := &Foo{
            Id:        uuid.New(),
            CompanyId: testFoo.CompanyId,
            CreatorId: testFoo.CreatorId,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
            Name:      "new item",
        }
        body, _ := json.Marshal(payload)
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

    t.Run("get with invalid uuid returns 400", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/foos/not-a-uuid", nil)
        res := httptest.NewRecorder()
        mux.ServeHTTP(res, req)
        assertStatus(t, http.StatusBadRequest, res.Code)
    })

    t.Run("get non-existent returns 404", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/foos/"+uuid.New().String(), nil)
        res := httptest.NewRecorder()
        mux.ServeHTTP(res, req)
        assertStatus(t, http.StatusNotFound, res.Code)
    })
}
```

---

## Controller Unit Tests (with mocked repo)

When testing business logic independently of the DB, implement a fake repository:

```go
type mockFooRepo struct {
    foos map[uuid.UUID]*Foo
}

func (m *mockFooRepo) Create(f *Foo) (*Foo, error) {
    m.foos[f.Id] = f
    return f, nil
}

func (m *mockFooRepo) GetByID(id uuid.UUID) (*Foo, error) {
    f, ok := m.foos[id]
    if !ok {
        return nil, sql.ErrNoRows
    }
    return f, nil
}

func (m *mockFooRepo) Update(f *Foo) (*Foo, error) {
    m.foos[f.Id] = f
    return f, nil
}

func TestFooController_CreateFoo(t *testing.T) {
    ctrl := &FooController{repo: &mockFooRepo{foos: make(map[uuid.UUID]*Foo)}}

    t.Run("rejects empty name", func(t *testing.T) {
        _, err := ctrl.CreateFoo(&Foo{})
        if !errors.Is(err, ErrFooInvalid) {
            t.Errorf("expected ErrFooInvalid, got %v", err)
        }
    })

    t.Run("creates valid foo", func(t *testing.T) {
        got, err := ctrl.CreateFoo(&Foo{Name: "hello"})
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if got.Name != "hello" {
            t.Errorf("name: got %q, want %q", got.Name, "hello")
        }
    })
}
```

---

## Table-Driven Tests

Use table-driven tests for validation logic and multiple input cases:

```go
func TestCreateFoo_Validation(t *testing.T) {
    ctrl := &FooController{repo: &mockFooRepo{foos: make(map[uuid.UUID]*Foo)}}

    tests := []struct {
        name    string
        input   *Foo
        wantErr error
    }{
        {"empty name", &Foo{}, ErrFooInvalid},
        {"negative amount", &Foo{Name: "x", Amount: -1}, ErrNegativeAmount},
        {"valid", &Foo{Name: "x", Amount: 10}, nil},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            _, err := ctrl.CreateFoo(tc.input)
            if !errors.Is(err, tc.wantErr) {
                t.Errorf("got %v, want %v", err, tc.wantErr)
            }
        })
    }
}
```

---

## Test Helpers

Keep helpers in the `testutil` package for cross-domain seeds:

```go
// testutil/seed.go
package testutil

func SeedCompanyAndUser(
    companyRepo companies.CompanyRepository,
    userRepo users.UserRepository,
) (*companies.Company, *users.User) {
    company, err := companyRepo.Create(&companies.Company{
        Id:        uuid.New(),
        Name:      "Test Corp",
        CNPJ:      "00000000000000",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    })
    if err != nil { panic(err) }

    user, err := userRepo.Create(&users.User{
        Id:        uuid.New(),
        Name:      "Test User",
        Email:     uuid.New().String() + "@test.com", // unique
        Password:  "hashed",
        CompanyId: company.Id,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    })
    if err != nil { panic(err) }

    return company, user
}
```

Keep `assertStatus` and similar assertion helpers as unexported functions local to each test file — don't share trivial helpers across packages.

---

## Running Tests

```bash
# All tests
make test

# Single package
go test ./transactions/...

# Verbose
go test -v ./transactions/...

# Single test function
go test -v -run TestTransactionController ./transactions/...
```

Requires: running PostgreSQL (`docker compose up -d`) and a valid `.env` file.
