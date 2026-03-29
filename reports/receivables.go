package reports

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared/types"
	"golang.org/x/text/currency"
)

type Receivables struct {
	Amount      currency.Amount
	DueDate     time.Time
	Description string
	AccountId   uuid.UUID
	VendorId    uuid.UUID
}

type Report struct {
	repo ReportsRepository
}

var ErrFailReport = errors.New("report failed")

func (r *Report) Transactions(interval types.Interval, co uuid.UUID) ([]Receivables, error) {
	re, err := r.repo.ReceivablesByInterval(interval, co)

	if err != nil {
		return nil, ErrFailReport
	}
	return re, nil
}
