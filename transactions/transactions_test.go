package transactions

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- mock repo ---

type mockTransactionRepo struct {
	store map[uuid.UUID]*Transaction
}

func newMockTransactionRepo() *mockTransactionRepo {
	return &mockTransactionRepo{store: make(map[uuid.UUID]*Transaction)}
}

func (m *mockTransactionRepo) Create(t *Transaction) (*Transaction, error) {
	m.store[t.Id] = t
	return t, nil
}

func (m *mockTransactionRepo) Update(t *Transaction) (*Transaction, error) {
	if _, ok := m.store[t.Id]; !ok {
		return nil, sql.ErrNoRows
	}
	m.store[t.Id] = t
	return t, nil
}

func (m *mockTransactionRepo) GetByID(id uuid.UUID) (*Transaction, error) {
	t, ok := m.store[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return t, nil
}

// --- controller tests ---

func TestControllerCreate(t *testing.T) {
	repo := newMockTransactionRepo()
	ctrl := NewController(repo)

	t.Run("valid transaction", func(t *testing.T) {
		tx := &Transaction{
			CompanyId:   uuid.New(),
			Type:        TransactionTypeDebit,
			Amount:      1000,
			Description: "office supplies",
			Origin:      "manual",
			CreatorId:   uuid.New(),
		}
		created, err := ctrl.Create(tx)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if created.Id == uuid.Nil {
			t.Error("Id should be populated")
		}
	})

	t.Run("negative amount returns error", func(t *testing.T) {
		tx := &Transaction{Amount: -100}
		_, err := ctrl.Create(tx)
		if !errors.Is(err, ErrNegativeAmount) {
			t.Errorf("Create() error = %v, want %v", err, ErrNegativeAmount)
		}
	})

	t.Run("zero amount is valid", func(t *testing.T) {
		tx := &Transaction{
			CompanyId: uuid.New(),
			Amount:    0,
			CreatorId: uuid.New(),
		}
		_, err := ctrl.Create(tx)
		if err != nil {
			t.Errorf("Create() error = %v, want nil", err)
		}
	})
}

func TestControllerUpdate(t *testing.T) {
	repo := newMockTransactionRepo()
	ctrl := NewController(repo)

	tx := &Transaction{
		CompanyId:   uuid.New(),
		Type:        TransactionTypeCredit,
		Amount:      500,
		Description: "payment",
		Origin:      "manual",
		CreatorId:   uuid.New(),
	}
	created, _ := ctrl.Create(tx)

	t.Run("valid update", func(t *testing.T) {
		created.Amount = 750
		updated, err := ctrl.UpdateTransaction(created)
		if err != nil {
			t.Fatalf("UpdateTransaction() error = %v", err)
		}
		if updated.Amount != 750 {
			t.Errorf("Amount = %v, want %v", updated.Amount, 750.0)
		}
	})

	t.Run("negative amount returns error", func(t *testing.T) {
		created.Amount = -1
		_, err := ctrl.UpdateTransaction(created)
		if !errors.Is(err, ErrNegativeAmount) {
			t.Errorf("UpdateTransaction() error = %v, want %v", err, ErrNegativeAmount)
		}
	})
}

// --- HTTP handler tests ---

func TestTransactionHandler(t *testing.T) {
	repo := newMockTransactionRepo()
	ctrl := NewController(repo)
	handler := NewHandler(ctrl)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("create transaction returns 201", func(t *testing.T) {
		tx := Transaction{
			CompanyId:   uuid.New(),
			Type:        TransactionTypeDebit,
			Amount:      1000,
			Description: "test",
			DueDate:     time.Now().Add(24 * time.Hour),
			Origin:      "test",
			CreatorId:   uuid.New(),
		}
		body, _ := json.Marshal(tx)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusCreated, res.Code)
	})

	t.Run("get transaction by id returns 200", func(t *testing.T) {
		tx := &Transaction{
			CompanyId: uuid.New(),
			Amount:    500,
			CreatorId: uuid.New(),
		}
		created, _ := ctrl.Create(tx)

		req := httptest.NewRequest(http.MethodGet, "/transactions/"+created.Id.String(), nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusOK, res.Code)
	})

	t.Run("get with invalid uuid returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/transactions/not-a-uuid", nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusBadRequest, res.Code)
	})

	t.Run("get non-existent returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/transactions/"+uuid.New().String(), nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		assertStatus(t, http.StatusNotFound, res.Code)
	})
}

func assertStatus(t testing.TB, want, got int) {
	t.Helper()
	if got != want {
		t.Errorf("status: got %d, want %d", got, want)
	}
}
