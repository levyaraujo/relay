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

func (r Repo) CreateWithOwner(payload CreateOrganizationPayload) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var organizationID string
	organization := payload.Organization
	orgInsert := `INSERT INTO organizations (name, website, phone, email, tax_id, currency, type, description) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err = tx.QueryRow(
		orgInsert,
		organization.Name,
		organization.Website,
		organization.Phone,
		organization.Email,
		organization.TaxID,
		organization.Currency,
		organization.Type.String(),
		organization.Description,
	).Scan(&organizationID)
	if err != nil {
		return err
	}

	user := payload.User
	userInsert := `INSERT INTO users (name, document, phone, email, password, organization_id) VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err = tx.Exec(userInsert, user.Name, user.Document, user.Phone, user.Email, user.Password, organizationID); err != nil {
		return err
	}

	return tx.Commit()
}

func CreateRepo(db *sql.DB) Repo {
	return Repo{db}
}
