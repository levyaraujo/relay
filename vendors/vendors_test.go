package vendors

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// --- mock repo ---

type mockVendorRepo struct {
	store map[uuid.UUID]*Vendor
}

func newMockVendorRepo() *mockVendorRepo {
	return &mockVendorRepo{store: make(map[uuid.UUID]*Vendor)}
}

func (m *mockVendorRepo) Create(v *Vendor) (*Vendor, error) {
	m.store[v.Id] = v
	return v, nil
}

func (m *mockVendorRepo) Update(v *Vendor) (*Vendor, error) {
	if _, ok := m.store[v.Id]; !ok {
		return nil, sql.ErrNoRows
	}
	m.store[v.Id] = v
	return v, nil
}

func (m *mockVendorRepo) GetByID(id uuid.UUID) (*Vendor, error) {
	v, ok := m.store[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return v, nil
}

// --- controller tests ---

func TestControllerCreate(t *testing.T) {
	repo := newMockVendorRepo()
	ctrl := NewController(repo)

	t.Run("valid vendor", func(t *testing.T) {
		v := &Vendor{
			CompanyId:    uuid.New(),
			Name:         "Supplier Co",
			CNPJ:         "96100041000129",
			Email:        "contact@supplier.com",
			Phone:        "11999999999",
			PaymentTerms: 30,
		}
		created, err := ctrl.Create(v)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if created.Id == uuid.Nil {
			t.Error("Id should be populated")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		v := &Vendor{Name: "", CNPJ: "96100041000129"}
		_, err := ctrl.Create(v)
		if !errors.Is(err, ErrNameRequired) {
			t.Errorf("Create() error = %v, want %v", err, ErrNameRequired)
		}
	})

	t.Run("invalid CNPJ returns error", func(t *testing.T) {
		v := &Vendor{Name: "Bad CNPJ", CNPJ: "00000000000000"}
		_, err := ctrl.Create(v)
		if !errors.Is(err, ErrInvalidCNPJ) {
			t.Errorf("Create() error = %v, want %v", err, ErrInvalidCNPJ)
		}
	})

	t.Run("negative payment terms returns error", func(t *testing.T) {
		v := &Vendor{
			Name:         "Negative Terms",
			CNPJ:         "96100041000129",
			PaymentTerms: -1,
		}
		_, err := ctrl.Create(v)
		if !errors.Is(err, ErrInvalidPaymentTerms) {
			t.Errorf("Create() error = %v, want %v", err, ErrInvalidPaymentTerms)
		}
	})
}

func TestControllerUpdate(t *testing.T) {
	repo := newMockVendorRepo()
	ctrl := NewController(repo)

	v := &Vendor{
		CompanyId:    uuid.New(),
		Name:         "Original Vendor",
		CNPJ:         "96100041000129",
		Email:        "original@vendor.com",
		Phone:        "11999999999",
		PaymentTerms: 30,
	}
	created, _ := ctrl.Create(v)

	t.Run("valid update", func(t *testing.T) {
		created.Name = "Updated Vendor"
		updated, err := ctrl.Update(created)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Name != "Updated Vendor" {
			t.Errorf("Name = %q, want %q", updated.Name, "Updated Vendor")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		created.Name = ""
		_, err := ctrl.Update(created)
		if !errors.Is(err, ErrNameRequired) {
			t.Errorf("Update() error = %v, want %v", err, ErrNameRequired)
		}
	})

	t.Run("negative payment terms returns error", func(t *testing.T) {
		created.Name = "Good Name"
		created.PaymentTerms = -5
		_, err := ctrl.Update(created)
		if !errors.Is(err, ErrInvalidPaymentTerms) {
			t.Errorf("Update() error = %v, want %v", err, ErrInvalidPaymentTerms)
		}
	})
}

// --- HTTP handler tests ---

func TestVendorHandler(t *testing.T) {
	repo := newMockVendorRepo()
	ctrl := NewController(repo)
	handler := NewHandler(ctrl)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("create returns 201", func(t *testing.T) {
		v := Vendor{
			CompanyId:    uuid.New(),
			Name:         "New Vendor",
			CNPJ:         "96100041000129",
			Email:        "new@vendor.com",
			Phone:        "11888888888",
			PaymentTerms: 15,
		}
		body, _ := json.Marshal(v)
		req := httptest.NewRequest(http.MethodPost, "/vendors", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusCreated, res.Code)
	})

	t.Run("create with empty name returns 422", func(t *testing.T) {
		v := Vendor{Name: "", CNPJ: "96100041000129"}
		body, _ := json.Marshal(v)
		req := httptest.NewRequest(http.MethodPost, "/vendors", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusUnprocessableEntity, res.Code)
	})

	t.Run("get by id returns 200", func(t *testing.T) {
		v := &Vendor{
			CompanyId:    uuid.New(),
			Name:         "Lookup Vendor",
			CNPJ:         "96100041000129",
			PaymentTerms: 30,
		}
		created, _ := ctrl.Create(v)

		req := httptest.NewRequest(http.MethodGet, "/vendors/"+created.Id.String(), nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusOK, res.Code)
	})

	t.Run("get with invalid uuid returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vendors/not-a-uuid", nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusBadRequest, res.Code)
	})

	t.Run("get non-existent returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vendors/"+uuid.New().String(), nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusNotFound, res.Code)
	})

	t.Run("update returns 200", func(t *testing.T) {
		v := &Vendor{
			CompanyId:    uuid.New(),
			Name:         "To Update",
			CNPJ:         "96100041000129",
			PaymentTerms: 30,
		}
		created, _ := ctrl.Create(v)

		created.Name = "Updated Vendor"
		body, _ := json.Marshal(created)
		req := httptest.NewRequest(http.MethodPut, "/vendors/"+created.Id.String(), bytes.NewReader(body))
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
