package transactions

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/shared/types"
)

var (
	ErrNegativeAmount = errors.New("amount cannot be negative")
	ErrTxsNotFound    = errors.New("transactions not found")
	ErrInternal       = errors.New("an error occured while searching transactions")
)

type TransactionController struct {
	repo TransactionRepository
}

func NewController(repo TransactionRepository) *TransactionController {
	return &TransactionController{repo}
}

func (c *TransactionController) Create(t *Transaction) (*Transaction, error) {
	if t.Amount < 0 {
		return nil, ErrNegativeAmount
	}

	t.Model = shared.NewModel()

	return c.repo.Create(t)
}

func (c *TransactionController) GetTransactionByID(id uuid.UUID) (*Transaction, error) {
	return c.repo.GetByID(id)
}

func (c *TransactionController) TransactionsByDateRange(co uuid.UUID, dateRange types.Interval) ([]Transaction, error) {
	txs, err := c.repo.ByCompanyAndDateRange(co, dateRange)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTxsNotFound
		}

		return nil, ErrInternal
	}

	return txs, nil
}

// UpdateTransaction validates and persists changes to a transaction.
func (c *TransactionController) UpdateTransaction(t *Transaction) (*Transaction, error) {
	if t.Amount < 0 {
		return nil, ErrNegativeAmount
	}
	return c.repo.Update(t)
}

type dashboardSummary struct {
	TotalIncome       float64 `json:"totalIncome"`
	TotalExpenses     float64 `json:"totalExpenses"`
	Balance           float64 `json:"balance"`
	TotalTransactions int     `json:"totalTransactions"`
}

func (c *TransactionController) DashboardSummary(co uuid.UUID, dateRange types.Interval) (*dashboardSummary, error) {
	txs, err := c.repo.ByCompanyAndDateRange(co, dateRange)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTxsNotFound
		}
		return nil, ErrInternal
	}

	totInc := 0.0
	for _, tx := range txs {
		if tx.Type == TransactionTypeCredit {
			totInc += tx.Amount
		}
	}

	totExp := 0.0
	for _, tx := range txs {
		if tx.Type == TransactionTypeDebit {
			totExp += tx.Amount
		}
	}
	balance := totInc - totExp
	totTxs := len(txs)

	return &dashboardSummary{
		TotalIncome:       totInc,
		TotalExpenses:     totExp,
		Balance:           balance,
		TotalTransactions: totTxs,
	}, nil
}
