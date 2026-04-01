package reports

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/levyaraujo/relay/shared/types"
)

type ReportsRepository interface {
	ReceivablesByInterval(interval types.Interval, co uuid.UUID) ([]Receivables, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) ReportsRepository {
	return &repository{db: db}
}

func (r *repository) ReceivablesByInterval(interval types.Interval, co uuid.UUID) ([]Receivables, error) {
	var receivables []Receivables

	err := r.db.Get(&receivables, `
		SELECT amount, due_date, description, account_id, vendor_id FROM transactions
		WHERE due_date >= $1 AND due_date <= $2 AND deleted = FALSE AND paid_date IS NULL
		AND company_id = $3;
	`, interval.Start, interval.End, co)

	if err != nil {
		return nil, err
	}
	return receivables, nil
}
