package accounts

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// AccountRepository defines persistence operations for GL accounts.
type AccountRepository interface {
	Create(a *Account) (*Account, error)
	GetByID(id uuid.UUID) (*Account, error)
	Update(a *Account) (*Account, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository returns an AccountRepository backed by sqlx.
func NewRepository(db *sqlx.DB) AccountRepository {
	return &repository{db: db}
}

func (r *repository) Create(a *Account) (*Account, error) {
	var created Account
	rows, err := r.db.NamedQuery(`
		INSERT INTO accounts
			(id, company_id, code, name, type, parent_id, created_at, updated_at, deleted)
		VALUES
			(:id, :company_id, :code, :name, :type, :parent_id, :created_at, :updated_at, :deleted)
		RETURNING *
	`, a)
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

func (r *repository) GetByID(id uuid.UUID) (*Account, error) {
	var a Account
	err := r.db.Get(&a, `SELECT * FROM accounts WHERE id = $1 AND deleted = false`, id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) Update(a *Account) (*Account, error) {
	var updated Account
	rows, err := r.db.NamedQuery(`
		UPDATE accounts SET
			code       = :code,
			name       = :name,
			type       = :type,
			parent_id  = :parent_id,
			updated_at = :updated_at,
			deleted    = :deleted
		WHERE id = :id
		RETURNING *
	`, a)
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