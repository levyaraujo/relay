package transactions

import (
	"lucrerp/companies"
	"lucrerp/shared"
	"lucrerp/users"
	"time"

	"github.com/google/uuid"
)

type Stringer interface {
	String() string
}

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
	CompanyId   uuid.UUID       `gorm:"not null"`
	Type        TransactionType `gorm:"not null"`
	Amount      float64         `gorm:"type:decimal(15,2);not null"`
	Description string          `gorm:"not null"`
	DueDate     time.Time
	PaidDate    time.Time
	Origin      string            `gorm:"not null"`
	CreatorId   uuid.UUID         `gorm:"not null"`
	Creator     users.User        `gorm:"foreignKey:CreatorId"`
	Company     companies.Company `gorm:"foreignKey:CompanyId"`
}
