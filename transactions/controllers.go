package transactions

import (
	"errors"

	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
)

var ErrNegativeAmount = errors.New("amount cannot be negative")

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

// UpdateTransaction validates and persists changes to a transaction.
func (c *TransactionController) UpdateTransaction(t *Transaction) (*Transaction, error) {
	if t.Amount < 0 {
		return nil, ErrNegativeAmount
	}
	return c.repo.Update(t)
}
