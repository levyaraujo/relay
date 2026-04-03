package transactions

import (
	"time"

	"github.com/levyaraujo/relay/shared"

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

func (t TransactionType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

type Transaction struct {
	shared.Model
	CompanyId   uuid.UUID       `db:"company_id" json:"company"`
	CreatorId   uuid.UUID       `db:"creator_id" json:"user"`
	AccountId   uuid.UUID       `db:"account_id" json:"account"`
	VendorId    uuid.UUID       `db:"vendor_id" json:"vendor"`
	Type        TransactionType `db:"type" json:"type"`
	Amount      float64         `db:"amount" json:"amount"`
	Description string          `db:"description" json:"description"`
	DueDate     time.Time       `db:"due_date" json:"due_date"`
	PaidDate    time.Time       `db:"paid_date" json:"paid_date"`
	Origin      string          `db:"origin" json:"origin"`
}

type CashFlowPoint struct {
	Label   time.Time `db:"label" json:"label"`
	Income  float64   `db:"income" json:"income"`
	Expense float64   `db:"expense" json:"expense"`
}
