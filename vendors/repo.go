package vendors

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// VendorRepository defines persistence operations for vendors.
type VendorRepository interface {
	Create(v *Vendor) (*Vendor, error)
	GetByID(id uuid.UUID) (*Vendor, error)
	Update(v *Vendor) (*Vendor, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository returns a VendorRepository backed by sqlx.
func NewRepository(db *sqlx.DB) VendorRepository {
	return &repository{db: db}
}

func (r *repository) Create(v *Vendor) (*Vendor, error) {
	var created Vendor
	rows, err := r.db.NamedQuery(`
		INSERT INTO vendors
			(id, company_id, name, cnpj, email, phone, payment_terms, created_at, updated_at, deleted)
		VALUES
			(:id, :company_id, :name, :cnpj, :email, :phone, :payment_terms, :created_at, :updated_at, :deleted)
		RETURNING *
	`, v)
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

func (r *repository) GetByID(id uuid.UUID) (*Vendor, error) {
	var v Vendor
	err := r.db.Get(&v, `SELECT * FROM vendors WHERE id = $1 AND deleted = false`, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *repository) Update(v *Vendor) (*Vendor, error) {
	var updated Vendor
	rows, err := r.db.NamedQuery(`
		UPDATE vendors SET
			name          = :name,
			cnpj          = :cnpj,
			email         = :email,
			phone         = :phone,
			payment_terms = :payment_terms,
			updated_at    = :updated_at,
			deleted       = :deleted
		WHERE id = :id
		RETURNING *
	`, v)
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
