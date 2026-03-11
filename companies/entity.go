package companies

import "lucrerp/shared"

type Company struct {
	shared.Model
	Name string `db:"name"`
	CNPJ string `db:"cnpj"`
}
