package companies

import "github.com/levyaraujo/relay/shared"

type Company struct {
	shared.Model
	Name string `db:"name"`
	CNPJ string `db:"cnpj"`
}

// ValidateCNPJ checks whether the company's CNPJ is structurally valid.
func (c Company) ValidateCNPJ() bool {
	return shared.ValidateCNPJ(c.CNPJ)
}