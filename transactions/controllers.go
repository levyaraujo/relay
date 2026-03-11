package transactions

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrNegativeAmount = errors.New("amount cannot be negative")

type TransactionController struct {
	repo TransactionRepository
}

func NewTransactionController(db *sqlx.DB) *TransactionController {
	return &TransactionController{NewRepository(db)}
}

func (c *TransactionController) CreateTransaction(t *Transaction) (*Transaction, error) {
	if t.Amount < 0 {
		return nil, ErrNegativeAmount
	}

	return c.repo.Create(t)
}
