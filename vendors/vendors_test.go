package vendors

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/levyaraujo/relay/companies"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/testutil"
	"github.com/levyaraujo/relay/users"
)

// --- HTTP handler tests (integration, real DB) ---

func TestVendorHandler(t *testing.T) {
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)

	t.Run("create returns 201", func(t *testing.T) {
		_, company := seedVendor(t)
		v := Vendor{
			CompanyId:    company.Id,
			Name:         "New Vendor",
			CNPJ:         "99888777000100",
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
		v := Vendor{Name: ""}
		body, _ := json.Marshal(v)
		req := httptest.NewRequest(http.MethodPost, "/vendors", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusUnprocessableEntity, res.Code)
	})

	t.Run("get by id returns 200", func(t *testing.T) {
		seeded, _ := seedVendor(t)
		req := httptest.NewRequest(http.MethodGet, "/vendors/"+seeded.Id.String(), nil)
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
		seeded, _ := seedVendor(t)
		seeded.Name = "Updated Vendor"
		body, _ := json.Marshal(seeded)
		req := httptest.NewRequest(http.MethodPut, "/vendors/"+seeded.Id.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusOK, res.Code)
	})
}

// --- helpers ---

func assertStatus(t testing.TB, want, got int) {
	t.Helper()
	if got != want {
		t.Errorf("status: got %d, want %d", got, want)
	}
}

var (
	testDB      *sqlx.DB
	testHandler *Handler
	vendorRepo  VendorRepository
	companyRepo companies.CompanyRepository
	userRepo    users.UserRepository
)

func TestMain(m *testing.M) {
	testDB = shared.Connect()

	companyRepo = companies.NewRepository(testDB)
	userRepo = users.NewRepository(testDB)
	vendorRepo = NewRepository(testDB)

	ctrl := &Controller{repo: vendorRepo}
	testHandler = NewHandler(ctrl)

	defer func() {
		testutil.CleanTables(nil, testDB, "vendors", "users", "companies")
	}()

	os.Exit(m.Run())
}

func seedVendor(t *testing.T) (*Vendor, *companies.Company) {
	t.Helper()
	company, _ := testutil.SeedCompanyAndUser(companyRepo, userRepo)

	v := &Vendor{
		CompanyId:    company.Id,
		Name:         "Test Vendor",
		CNPJ:         "11222333000181",
		Email:        uuid.New().String() + "@vendor.com",
		Phone:        "11999999999",
		PaymentTerms: 30,
	}
	v.Id = uuid.New()
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()

	created, err := vendorRepo.Create(v)
	if err != nil {
		t.Fatalf("seed vendor: %v", err)
	}

	t.Cleanup(func() {
		testutil.CleanTables(t, testDB, "vendors", "users", "companies")
	})

	return created, company
}
