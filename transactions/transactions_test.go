package transactions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/levyaraujo/relay/companies"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/testutil"
	"github.com/levyaraujo/relay/users"

	"github.com/google/uuid"
)

func TestTransactionHandler(t *testing.T) {
	mux := http.NewServeMux()
	testHandler.RegisterRoutes(mux)

	t.Run("create transaction", func(t *testing.T) {
		seeded := seedTransaction(t)
		newTransaction := *seeded
		newTransaction.Id = uuid.New()

		body, _ := json.Marshal(newTransaction)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusCreated, res.Code)
	})

	t.Run("get transaction by id", func(t *testing.T) {
		seeded := seedTransaction(t)
		req := httptest.NewRequest(http.MethodGet, "/transactions/"+seeded.Id.String(), nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusOK, res.Code)
	})
}

var (
	testDB                *sqlx.DB
	testHandler           *Handler
	transactionRepo       TransactionRepository
	transactionController *TransactionController
	companyRepo           companies.CompanyRepository
	userRepo              users.UserRepository
)

func TestMain(m *testing.M) {
	testDB = shared.Connect()

	companyRepo = companies.NewRepository(testDB)
	userRepo = users.NewRepository(testDB)
	transactionRepo = NewRepository(testDB)
	transactionController = NewController(transactionRepo)

	testHandler = NewHandler(transactionController)

	defer func() {
		testutil.CleanTables(nil, testDB, "transactions", "users", "companies")
	}()

	os.Exit(m.Run())
}

// seedTransaction creates a company, user, and transaction for a single test.
func seedTransaction(t *testing.T) *Transaction {
	t.Helper()
	company, user := testutil.SeedCompanyAndUser(companyRepo, userRepo)

	transaction := &Transaction{
		CompanyId:   company.Id,
		Type:        TransactionTypeDebit,
		Amount:      1000,
		Description: "test",
		Origin:      "test",
		CreatorId:   user.Id,
	}
	transaction.Id = uuid.New()
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()

	created, err := transactionRepo.Create(transaction)
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}

	t.Cleanup(func() {
		testutil.CleanTables(t, testDB, "transactions", "users", "companies")
	})

	return created
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
