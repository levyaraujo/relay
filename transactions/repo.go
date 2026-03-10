package transactions

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TransactionRepository interface {
	Create(t *Transaction) error
	Update(t *Transaction) (*Transaction, error)
	GetByID(id uuid.UUID) (*Transaction, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) TransactionRepository {
	return &repository{db: db}
}

func (r *repository) Create(t *Transaction) error {
	_, err := r.db.NamedExec(`
		INSERT INTO transactions
			(id, company_id, type, amount, description, due_date, paid_date, origin, creator_id, created_at, updated_at, deleted)
		VALUES
			(:id, :company_id, :type, :amount, :description, :due_date, :paid_date, :origin, :creator_id, :created_at, :updated_at, :deleted)
	`, t)
	return err
}

func (r *repository) Update(t *Transaction) (*Transaction, error) {
	var updated Transaction
	rows, err := r.db.NamedQuery(`
		UPDATE transactions SET
			company_id  = :company_id,
			type        = :type,
			amount      = :amount,
			description = :description,
			due_date    = :due_date,
			paid_date   = :paid_date,
			origin      = :origin,
			updated_at  = :updated_at,
			deleted     = :deleted
		WHERE id = :id
		RETURNING *
	`, t)
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

func (r *repository) GetByID(id uuid.UUID) (*Transaction, error) {
	var t Transaction
	err := r.db.Get(&t, `SELECT * FROM transactions WHERE id = $1 AND deleted = false`, id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
