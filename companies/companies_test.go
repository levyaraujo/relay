package companies

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/levyaraujo/relay/accounts"
)

// --- controller tests ---

func TestControllerCreate(t *testing.T) {
	repo := newMockCompanyRepo()
	ctrl := NewController(repo)
	accRepo := &mockAccountRepo{}

	t.Run("valid company", func(t *testing.T) {
		co := &Company{Name: "Acme", CNPJ: "96100041000129"}
		created, err := ctrl.Create(co, accRepo)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if created.Name != "Acme" {
			t.Errorf("Name = %q, want %q", created.Name, "Acme")
		}
		if created.Id == uuid.Nil {
			t.Error("Id should be populated")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		co := &Company{Name: "", CNPJ: "96100041000129"}
		_, err := ctrl.Create(co, accRepo)
		if !errors.Is(err, ErrNameRequired) {
			t.Errorf("Create() error = %v, want %v", err, ErrNameRequired)
		}
	})

	t.Run("invalid CNPJ returns error", func(t *testing.T) {
		co := &Company{Name: "Bad Co", CNPJ: "00000000000000"}
		_, err := ctrl.Create(co, accRepo)
		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Create() error = %v, want %v", err, ErrInvalidCNPJ)
		}
	})
}

func TestControllerUpdate(t *testing.T) {
	repo := newMockCompanyRepo()
	ctrl := NewController(repo)
	accRepo := &mockAccountRepo{}

	co := &Company{Name: "Original", CNPJ: "96100041000129"}
	created, _ := ctrl.Create(co, accRepo)

	t.Run("valid update", func(t *testing.T) {
		created.Name = "Updated"
		updated, err := ctrl.Update(created)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != "Updated" {
			t.Errorf("Name = %q, want %q", updated.Name, "Updated")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		created.Name = ""
		_, err := ctrl.Update(created)
		if !errors.Is(err, ErrNameRequired) {
			t.Errorf("Update() error = %v, want %v", err, ErrNameRequired)
		}
	})

	t.Run("invalid CNPJ returns error", func(t *testing.T) {
		created.Name = "Good Name"
		created.CNPJ = "00000000000000"
		_, err := ctrl.Update(created)
		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Update() error = %v, want %v", err, ErrInvalidCNPJ)
		}
	})
}

// --- HTTP handler tests ---

func TestCompanyHandler(t *testing.T) {
	repo := newMockCompanyRepo()
	ctrl := NewController(repo)
	handler := NewHandler(ctrl, &mockAccountRepo{})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("create returns 201", func(t *testing.T) {
		c := Company{Name: "New Co", CNPJ: "96100041000129"}
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusCreated, res.Code)
	})

	t.Run("create with empty name returns 422", func(t *testing.T) {
		c := Company{Name: "", CNPJ: "96100041000129"}
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusUnprocessableEntity, res.Code)
	})

	t.Run("create with invalid CNPJ returns 422", func(t *testing.T) {
		c := Company{Name: "Bad CNPJ Co", CNPJ: "00000000000000"}
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/companies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusUnprocessableEntity, res.Code)
	})

	t.Run("get by id returns 200", func(t *testing.T) {
		// seed via controller so it's in the mock store
		co := &Company{Name: "Lookup Co", CNPJ: "96100041000129"}
		created, _ := ctrl.Create(co, &mockAccountRepo{})

		req := httptest.NewRequest(http.MethodGet, "/companies/"+created.Id.String(), nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusOK, res.Code)
	})

	t.Run("get with invalid uuid returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/companies/not-a-uuid", nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusBadRequest, res.Code)
	})

	t.Run("get non-existent returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/companies/"+uuid.New().String(), nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusNotFound, res.Code)
	})

	t.Run("update returns 200", func(t *testing.T) {
		co := &Company{Name: "To Update", CNPJ: "96100041000129"}
		created, _ := ctrl.Create(co, &mockAccountRepo{})

		created.Name = "Updated Co"
		body, _ := json.Marshal(created)
		req := httptest.NewRequest(http.MethodPut, "/companies/"+created.Id.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusOK, res.Code)
	})
}

func assertStatus(t testing.TB, want, got int) {
	t.Helper()
	if got != want {
		t.Errorf("status: got %d, want %d", got, want)
	}
}

// --- mock repos ---

type mockCompanyRepo struct {
	store map[uuid.UUID]*Company
}

func newMockCompanyRepo() *mockCompanyRepo {
	return &mockCompanyRepo{store: make(map[uuid.UUID]*Company)}
}

func (m *mockCompanyRepo) Create(c *Company) (*Company, error) {
	m.store[c.Id] = c
	return c, nil
}

func (m *mockCompanyRepo) Update(c *Company) (*Company, error) {
	if _, ok := m.store[c.Id]; !ok {
		return nil, sql.ErrNoRows
	}
	m.store[c.Id] = c
	return c, nil
}

func (m *mockCompanyRepo) GetByID(id uuid.UUID) (*Company, error) {
	c, ok := m.store[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return c, nil
}

type mockAccountRepo struct{}

func (m *mockAccountRepo) Create(a *accounts.Account) (*accounts.Account, error) {
	return a, nil
}
func (m *mockAccountRepo) GetByID(id uuid.UUID) (*accounts.Account, error) {
	return nil, sql.ErrNoRows
}
func (m *mockAccountRepo) Update(a *accounts.Account) (*accounts.Account, error) {
	return a, nil
}
