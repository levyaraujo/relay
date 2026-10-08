package finance

import (
	"database/sql"
	"encoding/json"
	"errors"
	"relay/audit"
	"uuid"
)

type Repo struct {
	db    *sql.DB
	audit audit.Repo
}

func (r Repo) PartyHasRole(orgID, partyID uuid.UUID, role string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM parties p JOIN party_roles pr ON pr.party_id = p.id WHERE p.organization_id = $1 AND p.id = $2 AND pr.role = $3)`, orgID, partyID, role).Scan(&exists)
	return exists, err
}
func (r Repo) ItemExists(orgID, itemID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM items WHERE organization_id = $1 AND id = $2)`, orgID, itemID).Scan(&exists)
	return exists, err
}

func (r Repo) CreateTransaction(orgID uuid.UUID, draft TransactionDraft) (TransactionDetail, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return TransactionDetail{}, err
	}
	defer tx.Rollback()
	var transactionID uuid.UUID
	transaction := draft.Transaction
	err = tx.QueryRow(`INSERT INTO transactions (organization_id, party_id, type, status, occurred_at, description, currency, total_cents) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, orgID, transaction.PartyID, transaction.Type, transaction.Status, transaction.OccurredAt, transaction.Description, transaction.Currency, transaction.TotalCents).Scan(&transactionID)
	if err != nil {
		return TransactionDetail{}, mapRepositoryError(err)
	}
	transaction.ID = transactionID
	lines := make([]TransactionLine, 0, len(draft.Lines))
	for _, line := range draft.Lines {
		var id uuid.UUID
		err = tx.QueryRow(`INSERT INTO transaction_lines (transaction_id, item_id, description, quantity, unit_price_cents, total_cents) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, transactionID, line.ItemID, line.Description, line.Quantity, line.UnitPriceCents, line.TotalCents).Scan(&id)
		if err != nil {
			return TransactionDetail{}, err
		}
		line.ID, line.TransactionID = id, transactionID
		lines = append(lines, line)
	}
	obligation := draft.Obligation
	var obligationID uuid.UUID
	err = tx.QueryRow(`INSERT INTO obligations (organization_id, transaction_id, party_id, direction, status, amount_cents, paid_amount_cents, currency, due_date) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, orgID, transactionID, obligation.PartyID, obligation.Direction, obligation.Status, obligation.AmountCents, obligation.PaidAmountCents, obligation.Currency, obligation.DueDate).Scan(&obligationID)
	if err != nil {
		return TransactionDetail{}, err
	}
	obligation.ID, obligation.TransactionID = obligationID, transactionID
	payments := make([]Payment, 0, 1)
	if draft.InitialPayment != nil {
		payment := *draft.InitialPayment
		payment.ObligationID, payment.OrganizationID = obligationID, orgID
		var paymentID uuid.UUID
		err = tx.QueryRow(`INSERT INTO payments (organization_id, obligation_id, amount_cents, paid_at, method, reference) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, orgID, obligationID, payment.AmountCents, payment.PaidAt, payment.Method, payment.Reference).Scan(&paymentID)
		if err != nil {
			return TransactionDetail{}, err
		}
		payment.ID = paymentID
		payments = append(payments, payment)
	}
	payload, err := json.Marshal(TransactionDetail{Transaction: transaction, Lines: lines, Obligation: &obligation, Payments: payments})
	if err != nil {
		return TransactionDetail{}, err
	}
	if err := r.audit.Enqueue(tx, audit.NewEvent(orgID, draft.ActorUserID, "TRANSACTION_CREATED", "TRANSACTION", transactionID, "CREATE", payload)); err != nil {
		return TransactionDetail{}, err
	}
	if err := tx.Commit(); err != nil {
		return TransactionDetail{}, err
	}
	return TransactionDetail{Transaction: transaction, Lines: lines, Obligation: &obligation, Payments: payments}, nil
}

func (r Repo) FindTransaction(orgID, id uuid.UUID) (TransactionDetail, error) {
	var d TransactionDetail
	var tx Transaction
	err := r.db.QueryRow(`SELECT id, organization_id, party_id, type, status, occurred_at, description, currency, total_cents FROM transactions WHERE organization_id = $1 AND id = $2`, orgID, id).Scan(&tx.ID, &tx.OrganizationID, &tx.PartyID, &tx.Type, &tx.Status, &tx.OccurredAt, &tx.Description, &tx.Currency, &tx.TotalCents)
	if err != nil {
		return d, err
	}
	d.Transaction = tx
	d.Lines, err = r.lines(id)
	if err != nil {
		return d, err
	}
	d.Obligation, err = r.obligation(orgID, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return d, err
	}
	if d.Obligation != nil {
		d.Payments, err = r.payments(orgID, d.Obligation.ID)
		if err != nil {
			return d, err
		}
	}
	return d, nil
}
func (r Repo) ListTransactions(orgID uuid.UUID, filter TransactionFilter) ([]TransactionDetail, error) {
	query := `SELECT id FROM transactions WHERE organization_id = $1`
	args := []any{orgID}
	n := 2
	if filter.Type != nil {
		query += ` AND type = $` + itoa(n)
		args = append(args, *filter.Type)
		n++
	}
	if filter.Status != nil {
		query += ` AND status = $` + itoa(n)
		args = append(args, *filter.Status)
	}
	query += ` ORDER BY occurred_at DESC, id`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []TransactionDetail
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		d, err := r.FindTransaction(orgID, id)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (r Repo) CreatePayment(orgID, obligationID uuid.UUID, draft PaymentDraft) (ObligationDetail, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return ObligationDetail{}, err
	}
	defer tx.Rollback()
	var obligation Obligation
	err = tx.QueryRow(`SELECT id, organization_id, transaction_id, party_id, direction, status, amount_cents, paid_amount_cents, currency, due_date FROM obligations WHERE organization_id = $1 AND id = $2 FOR UPDATE`, orgID, obligationID).Scan(&obligation.ID, &obligation.OrganizationID, &obligation.TransactionID, &obligation.PartyID, &obligation.Direction, &obligation.Status, &obligation.AmountCents, &obligation.PaidAmountCents, &obligation.Currency, &obligation.DueDate)
	if err != nil {
		return ObligationDetail{}, err
	}
	if draft.AmountCents <= 0 || draft.AmountCents > obligation.AmountCents-obligation.PaidAmountCents {
		return ObligationDetail{}, PaymentExceedsBalanceErr
	}
	var payment Payment
	err = tx.QueryRow(`INSERT INTO payments (organization_id, obligation_id, amount_cents, paid_at, method, reference) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, orgID, obligationID, draft.AmountCents, draft.PaidAt, draft.Method, draft.Reference).Scan(&payment.ID)
	if err != nil {
		return ObligationDetail{}, err
	}
	payment.OrganizationID, payment.ObligationID, payment.AmountCents, payment.PaidAt, payment.Method, payment.Reference = orgID, obligationID, draft.AmountCents, draft.PaidAt, draft.Method, draft.Reference
	obligation.PaidAmountCents += draft.AmountCents
	obligation.Status = statusFor(obligation.AmountCents, obligation.PaidAmountCents)
	if _, err := tx.Exec(`UPDATE obligations SET paid_amount_cents = $1, status = $2 WHERE organization_id = $3 AND id = $4`, obligation.PaidAmountCents, obligation.Status, orgID, obligationID); err != nil {
		return ObligationDetail{}, err
	}
	payload, err := json.Marshal(payment)
	if err != nil {
		return ObligationDetail{}, err
	}
	if err := r.audit.Enqueue(tx, audit.NewEvent(orgID, draft.ActorUserID, "PAYMENT_CREATED", "PAYMENT", payment.ID, "CREATE", payload)); err != nil {
		return ObligationDetail{}, err
	}
	if err := tx.Commit(); err != nil {
		return ObligationDetail{}, err
	}
	payments, err := r.payments(orgID, obligationID)
	if err != nil {
		return ObligationDetail{}, err
	}
	return ObligationDetail{Obligation: obligation, Payments: payments}, nil
}

func (r Repo) ListObligations(orgID uuid.UUID, filter ObligationFilter) ([]ObligationDetail, error) {
	query := `SELECT id FROM obligations WHERE organization_id = $1`
	args := []any{orgID}
	n := 2
	if filter.Direction != nil {
		query += ` AND direction = $` + itoa(n)
		args = append(args, *filter.Direction)
		n++
	}
	if filter.Status != nil {
		query += ` AND status = $` + itoa(n)
		args = append(args, *filter.Status)
		n++
	}
	if filter.PartyID != nil {
		query += ` AND party_id = $` + itoa(n)
		args = append(args, *filter.PartyID)
		n++
	}
	if filter.DueBefore != nil {
		query += ` AND due_date <= $` + itoa(n)
		args = append(args, *filter.DueBefore)
		n++
	}
	if filter.DueAfter != nil {
		query += ` AND due_date >= $` + itoa(n)
		args = append(args, *filter.DueAfter)
	}
	query += ` ORDER BY due_date NULLS LAST, id`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ObligationDetail
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		d, err := r.FindObligation(orgID, id)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}
func (r Repo) FindObligation(orgID, id uuid.UUID) (ObligationDetail, error) {
	obligation, err := r.obligationByID(orgID, id)
	if err != nil {
		return ObligationDetail{}, err
	}
	payments, err := r.payments(orgID, id)
	if err != nil {
		return ObligationDetail{}, err
	}
	return ObligationDetail{Obligation: obligation, Payments: payments}, nil
}

func (r Repo) obligation(orgID, transactionID uuid.UUID) (*Obligation, error) {
	o, err := r.obligationByIDQuery(`SELECT id, organization_id, transaction_id, party_id, direction, status, amount_cents, paid_amount_cents, currency, due_date FROM obligations WHERE organization_id = $1 AND transaction_id = $2`, orgID, transactionID)
	return o, err
}
func (r Repo) obligationByID(orgID, id uuid.UUID) (Obligation, error) {
	o, err := r.obligationByIDQuery(`SELECT id, organization_id, transaction_id, party_id, direction, status, amount_cents, paid_amount_cents, currency, due_date FROM obligations WHERE organization_id = $1 AND id = $2`, orgID, id)
	if o == nil {
		return Obligation{}, err
	}
	return *o, err
}
func (r Repo) obligationByIDQuery(query string, args ...any) (*Obligation, error) {
	var o Obligation
	err := r.db.QueryRow(query, args...).Scan(&o.ID, &o.OrganizationID, &o.TransactionID, &o.PartyID, &o.Direction, &o.Status, &o.AmountCents, &o.PaidAmountCents, &o.Currency, &o.DueDate)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
func (r Repo) lines(transactionID uuid.UUID) ([]TransactionLine, error) {
	rows, err := r.db.Query(`SELECT id, transaction_id, item_id, description, quantity, unit_price_cents, total_cents FROM transaction_lines WHERE transaction_id = $1 ORDER BY id`, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []TransactionLine
	for rows.Next() {
		var l TransactionLine
		if err := rows.Scan(&l.ID, &l.TransactionID, &l.ItemID, &l.Description, &l.Quantity, &l.UnitPriceCents, &l.TotalCents); err != nil {
			return nil, err
		}
		result = append(result, l)
	}
	return result, rows.Err()
}
func (r Repo) payments(orgID, obligationID uuid.UUID) ([]Payment, error) {
	rows, err := r.db.Query(`SELECT id, obligation_id, organization_id, amount_cents, paid_at, method, reference FROM payments WHERE organization_id = $1 AND obligation_id = $2 ORDER BY paid_at, id`, orgID, obligationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.ObligationID, &p.OrganizationID, &p.AmountCents, &p.PaidAt, &p.Method, &p.Reference); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
func itoa(value int) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	b := make([]byte, 0, 4)
	for value > 0 {
		b = append([]byte{digits[value%10]}, b...)
		value /= 10
	}
	return string(b)
}
func mapRepositoryError(err error) error {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
		return TransactionConflictErr
	}
	return err
}
func CreateRepo(db *sql.DB) Repo { return Repo{db: db, audit: audit.CreateRepo(db)} }
