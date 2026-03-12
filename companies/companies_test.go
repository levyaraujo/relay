package companies

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
	"github.com/levyaraujo/relay/shared"
)

// --- HTTP handler tests (integration, real DB) ---

func TestCompanyHandler(t *testing.T) {
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)

	t.Run("create returns 201", func(t *testing.T) {
		t.Cleanup(func() {
			testDB.MustExec("DELETE FROM companies")
		})
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
		seeded := seedCompany(t)
		req := httptest.NewRequest(http.MethodGet, "/companies/"+seeded.Id.String(), nil)
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
		seeded := seedCompany(t)
		seeded.Name = "Updated Co"
		body, _ := json.Marshal(seeded)
		req := httptest.NewRequest(http.MethodPut, "/companies/"+seeded.Id.String(), bytes.NewReader(body))
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

var (
	testDB      *sqlx.DB
	testHandler *Handler
	companyRepo CompanyRepository
)

func TestMain(m *testing.M) {
	testDB = shared.Connect()

	companyRepo = NewRepository(testDB)
	ctrl := NewController(companyRepo)
	testHandler = NewHandler(ctrl)

	defer func() {
		testDB.MustExec("DELETE FROM companies")
	}()

	os.Exit(m.Run())
}

func seedCompany(t *testing.T) *Company {
	t.Helper()
	c := &Company{
		Name: "Test Company",
		CNPJ: "96100041000129",
	}
	c.Id = uuid.New()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	created, err := companyRepo.Create(c)
	if err != nil {
		t.Fatalf("seed company: %v", err)
	}

	t.Cleanup(func() {
		testDB.MustExec("DELETE FROM companies")
	})

	return created
}
