package transactions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"lucrerp/companies"
	"lucrerp/shared"
	"lucrerp/testutil"
	"lucrerp/users"

	"github.com/google/uuid"
)

func TestTransactionController(t *testing.T) {
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)

	t.Run("create transaction", func(t *testing.T) {
		newTransaction := *testTransaction
		newTransaction.Id = uuid.New()

		body, _ := json.Marshal(newTransaction)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusCreated, res.Code)
	})

	t.Run("get transaction by id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/transactions/"+testTransaction.Id.String(), nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusOK, res.Code)
	})
}

var (
	testHandler     *Handler
	testTransaction *Transaction
)

func TestMain(m *testing.M) {
	db := shared.Connect()

	companies.Migrate()
	users.Migrate()
	Migrate()

	companyRepo := companies.NewRepository(db)
	userRepo := users.NewRepository(db)
	txRepo := NewRepository(db)

	company, user := testutil.SeedCompanyAndUser(companyRepo, userRepo)

	testHandler = NewHandler(txRepo)

	t := &Transaction{
		CompanyId:   company.Id,
		Type:        TransactionTypeDebit,
		Amount:      1000,
		Description: "test",
		Origin:      "test",
		CreatorId:   user.Id,
	}
	t.Id = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	var err error
	testTransaction, err = txRepo.Create(t)
	if err != nil {
		panic("failed to seed test transaction: " + err.Error())
	}

	defer func() {
		db.MustExec("DELETE FROM transactions")
		db.MustExec("DELETE FROM users")
		db.MustExec("DELETE FROM companies")
	}()

	os.Exit(m.Run())
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
