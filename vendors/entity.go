package vendors

import (
	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
)

// Vendor represents a supplier or service provider that the company pays.
type Vendor struct {
	shared.Model
	CompanyId    uuid.UUID `db:"company_id" json:"company_id"`
	Name         string    `db:"name" json:"name"`
	CNPJ         string    `db:"cnpj" json:"cnpj"`
	Email        string    `db:"email" json:"email"`
	Phone        string    `db:"phone" json:"phone"`
	PaymentTerms int       `db:"payment_terms" json:"payment_terms"` // days, e.g. 30 = Net 30
}

func (v *Vendor) ValidateCNPJ() bool {
	return shared.ValidateCNPJ(v.CNPJ)
}
