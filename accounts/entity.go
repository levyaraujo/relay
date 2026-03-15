package accounts

import (
	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
)

type AccountType int

const (
	AccountAsset AccountType = iota + 1
	AccountLiability
	AccountEquity
	AccountRevenue
	AccountExpense
)

type Account struct {
	shared.Model
	CompanyId uuid.UUID   `db:"company_id"`
	Code      string      `db:"code"`
	Name      string      `db:"name"`
	Type      AccountType `db:"type"`
	ParentId  *uuid.UUID  `db:"parent_id"`
}
