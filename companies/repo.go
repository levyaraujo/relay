package companies

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CompanyRepository interface {
	Create(c *Company) (*Company, error)
	Update(c *Company) (*Company, error)
	GetByID(id uuid.UUID) (*Company, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) CompanyRepository {
	return &repository{db: db}
}

func (r *repository) Create(c *Company) (*Company, error) {
	var created Company
	rows, err := r.db.NamedQuery(`
		INSERT INTO companies
			(id, name, cnpj, created_at, updated_at, deleted)
		VALUES
			(:id, :name, :cnpj, :created_at, :updated_at, :deleted)
		RETURNING *
	`, c)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&created); err != nil {
			return nil, err
		}
	}
	return &created, nil
}

func (r *repository) Update(c *Company) (*Company, error) {
	var updated Company
	rows, err := r.db.NamedQuery(`
		UPDATE companies SET
			name       = :name,
			cnpj       = :cnpj,
			updated_at = :updated_at,
			deleted    = :deleted
		WHERE id = :id
		RETURNING *
	`, c)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&updated); err != nil {
			return nil, err
		}
	}
	return &updated, nil
}

func (r *repository) GetByID(id uuid.UUID) (*Company, error) {
	var c Company
	err := r.db.Get(&c, `SELECT * FROM companies WHERE id = $1 AND deleted = false`, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
