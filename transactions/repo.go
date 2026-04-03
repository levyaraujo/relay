package transactions

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/levyaraujo/relay/shared/types"
)

type TransactionRepository interface {
	Create(t *Transaction) (*Transaction, error)
	Update(t *Transaction) (*Transaction, error)
	GetByID(id uuid.UUID) (*Transaction, error)
	ByCompanyAndDateRange(co uuid.UUID, interval types.Interval) ([]Transaction, error)
	CashFlow(co uuid.UUID, interval types.Interval, groupBy string) ([]CashFlowPoint, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) TransactionRepository {
	return &repository{db: db}
}

func (r *repository) Create(t *Transaction) (*Transaction, error) {
	var created Transaction
	rows, err := r.db.NamedQuery(`
		INSERT INTO transactions
			(id, company_id, type, amount, description, due_date, paid_date, origin, creator_id, created_at, updated_at, deleted)
		VALUES
			(:id, :company_id, :type, :amount, :description, :due_date, :paid_date, :origin, :creator_id, :created_at, :updated_at, :deleted)
		RETURNING *
	`, t)
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

func (r *repository) Update(t *Transaction) (*Transaction, error) {
	var transaction Transaction
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
		if err := rows.StructScan(&transaction); err != nil {
			return nil, err
		}
	}
	return &transaction, nil
}

func (r *repository) GetByID(id uuid.UUID) (*Transaction, error) {
	var t Transaction
	err := r.db.Get(&t, `SELECT * FROM transactions WHERE id = $1 AND deleted = false`, id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) ByCompanyAndDateRange(co uuid.UUID, interval types.Interval) ([]Transaction, error) {
	var txs []Transaction

	err := r.db.Select(&txs, `
		SELECT id, company_id, creator_id, account_id, vendor_id, type, amount,
		       description, due_date, paid_date, origin, created_at
		FROM transactions
		WHERE company_id = $1
		  AND paid_date >= $2
		  AND paid_date < $3
		  AND deleted = false`,
		co, interval.Start, interval.End,
	)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *repository) CashFlow(co uuid.UUID, interval types.Interval, groupBy string) ([]CashFlowPoint, error) {
	var points []CashFlowPoint

	err := r.db.Select(&points, `
		SELECT
			date_trunc($1, paid_date) AS label,
			COALESCE(SUM(amount) FILTER (WHERE type = 1), 0) AS income,
			COALESCE(SUM(amount) FILTER (WHERE type = 2), 0) AS expense
		FROM transactions
		WHERE company_id = $2
		  AND paid_date >= $3
		  AND paid_date < $4
		  AND deleted = false
		GROUP BY date_trunc($1, paid_date)
		ORDER BY date_trunc($1, paid_date)`,
		groupBy, co, interval.Start, interval.End,
	)
	if err != nil {
		return nil, err
	}
	return points, nil
}
