package organization

import (
	"database/sql"
)

type Repo struct {
	db *sql.DB
}

func (r Repo) Create(o OrganizationPayload) *sql.Row {
	insert := `INSERT INTO organizations (name, website, phone, email, tax_id, currency, type, description) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	return r.db.QueryRow(insert, o.Name, o.Website, o.Phone, o.Email, o.TaxID, o.Currency, o.Type.String(), o.Description)
}

func CreateRepo(db *sql.DB) Repo {
	return Repo{db}
}
