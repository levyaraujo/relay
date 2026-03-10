package transactions

import (
	"lucrerp/shared"
	"time"

	"github.com/google/uuid"
)

type TransactionType int

const (
	TransactionTypeUnknown TransactionType = iota
	TransactionTypeDebit
	TransactionTypeCredit
)

func (t TransactionType) String() string {
	switch t {
	case TransactionTypeDebit:
		return "debit"
	case TransactionTypeCredit:
		return "credit"
	default:
		return "unknown"
	}
}

type Transaction struct {
	shared.Model
	CompanyId   uuid.UUID       `db:"company_id"`
	Type        TransactionType `db:"type"`
	Amount      float64         `db:"amount"`
	Description string          `db:"description"`
	DueDate     time.Time       `db:"due_date"`
	PaidDate    time.Time       `db:"paid_date"`
	Origin      string          `db:"origin"`
	CreatorId   uuid.UUID       `db:"creator_id"`
}
