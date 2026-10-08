package finance

import (
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"strings"
	"time"
	"uuid"
)

type repository interface {
	CreateTransaction(uuid.UUID, TransactionDraft) (TransactionDetail, error)
	ListTransactions(uuid.UUID, TransactionFilter) ([]TransactionDetail, error)
	FindTransaction(uuid.UUID, uuid.UUID) (TransactionDetail, error)
	ListObligations(uuid.UUID, ObligationFilter) ([]ObligationDetail, error)
	FindObligation(uuid.UUID, uuid.UUID) (ObligationDetail, error)
	CreatePayment(uuid.UUID, uuid.UUID, PaymentDraft) (ObligationDetail, error)
	PartyHasRole(uuid.UUID, uuid.UUID, string) (bool, error)
	ItemExists(uuid.UUID, uuid.UUID) (bool, error)
}

type Service struct{ repo repository }

var (
	InvalidTransactionTypeErr = errors.New("the transaction type is invalid")
	InvalidTransactionLineErr = errors.New("the transaction line is invalid")
	InvalidPaymentErr         = errors.New("the payment is invalid")
	InvalidCurrencyErr        = errors.New("the currency is invalid")
	InvalidFilterErr          = errors.New("the filter is invalid")
	PartyRoleConflictErr      = errors.New("the party does not have the required role")
	PaymentExceedsBalanceErr  = errors.New("the payment exceeds the outstanding balance")
	TransactionConflictErr    = errors.New("the transaction conflicts with an existing record")
	TransactionNotFoundErr    = errors.New("transaction not found")
	ObligationNotFoundErr     = errors.New("obligation not found")
	FinanceInternalErr        = errors.New("an internal error has occurred")
)

func NewService(repo Repo) *Service { return &Service{repo: repo} }

func (s Service) CreateTransaction(orgID uuid.UUID, actor *uuid.UUID, request CreateTransactionRequest) (TransactionDetail, error) {
	if request.Type != Sale && request.Type != Purchase && request.Type != Expense {
		return TransactionDetail{}, InvalidTransactionTypeErr
	}
	if len(request.Lines) == 0 {
		return TransactionDetail{}, InvalidTransactionLineErr
	}
	currency := strings.ToUpper(strings.TrimSpace(request.Currency))
	if currency == "" {
		currency = "BRL"
	}
	if len(currency) != 3 {
		return TransactionDetail{}, InvalidCurrencyErr
	}
	if request.OccurredAt.IsZero() {
		request.OccurredAt = time.Now()
	}
	if request.Type != Expense && request.PartyID == nil {
		return TransactionDetail{}, PartyRoleConflictErr
	}
	if request.PartyID != nil {
		requiredRole := "SUPPLIER"
		if request.Type == Sale {
			requiredRole = "CUSTOMER"
		}
		if ok, err := s.repo.PartyHasRole(orgID, *request.PartyID, requiredRole); err != nil {
			slog.Error("finance.partyRole", "err", err)
			return TransactionDetail{}, FinanceInternalErr
		} else if !ok {
			return TransactionDetail{}, PartyRoleConflictErr
		}
	}
	lines := make([]TransactionLine, 0, len(request.Lines))
	var total int64
	for _, line := range request.Lines {
		if line.Quantity <= 0 || line.UnitPriceCents < 0 {
			return TransactionDetail{}, InvalidTransactionLineErr
		}
		if request.Type != Expense && line.ItemID == nil {
			return TransactionDetail{}, InvalidTransactionLineErr
		}
		if line.ItemID == nil && strings.TrimSpace(line.Description) == "" {
			return TransactionDetail{}, InvalidTransactionLineErr
		}
		if line.ItemID != nil {
			if exists, err := s.repo.ItemExists(orgID, *line.ItemID); err != nil {
				return TransactionDetail{}, FinanceInternalErr
			} else if !exists {
				return TransactionDetail{}, InvalidTransactionLineErr
			}
		}
		if line.UnitPriceCents != 0 && line.Quantity > math.MaxInt64/line.UnitPriceCents {
			return TransactionDetail{}, InvalidTransactionLineErr
		}
		lineTotal := line.Quantity * line.UnitPriceCents
		if total > math.MaxInt64-lineTotal {
			return TransactionDetail{}, InvalidTransactionLineErr
		}
		total += lineTotal
		lines = append(lines, TransactionLine{ItemID: line.ItemID, Description: line.Description, Quantity: line.Quantity, UnitPriceCents: line.UnitPriceCents, TotalCents: lineTotal})
	}
	tx := Transaction{OrganizationID: orgID, PartyID: request.PartyID, Type: request.Type, Status: Confirmed, OccurredAt: request.OccurredAt, Description: request.Description, Currency: currency, TotalCents: total}
	direction := Payable
	if request.Type == Sale {
		direction = Receivable
	}
	obligation := Obligation{OrganizationID: orgID, PartyID: request.PartyID, Direction: direction, Status: Open, AmountCents: total, Currency: currency, DueDate: request.DueDate}
	if total <= 0 {
		return TransactionDetail{}, InvalidTransactionLineErr
	}
	draft := TransactionDraft{Transaction: tx, Lines: lines, Obligation: obligation, ActorUserID: actor}
	if request.InitialPayment != nil {
		payment, err := paymentFromRequest(*request.InitialPayment, total)
		if err != nil {
			return TransactionDetail{}, err
		}
		payment.OrganizationID = orgID
		draft.InitialPayment = &payment
		draft.Obligation.PaidAmountCents = payment.AmountCents
		draft.Obligation.Status = statusFor(total, payment.AmountCents)
	}
	detail, err := s.repo.CreateTransaction(orgID, draft)
	if err != nil {
		if errors.Is(err, TransactionConflictErr) {
			return TransactionDetail{}, err
		}
		slog.Error("finance.CreateTransaction", "err", err)
		return TransactionDetail{}, FinanceInternalErr
	}
	return detail, nil
}

func (s Service) ListTransactions(orgID uuid.UUID, filter TransactionFilter) ([]TransactionDetail, error) {
	details, err := s.repo.ListTransactions(orgID, filter)
	if err != nil {
		return nil, FinanceInternalErr
	}
	return details, nil
}
func (s Service) FindTransaction(orgID, id uuid.UUID) (TransactionDetail, error) {
	detail, err := s.repo.FindTransaction(orgID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return TransactionDetail{}, TransactionNotFoundErr
	}
	if err != nil {
		return TransactionDetail{}, FinanceInternalErr
	}
	return detail, nil
}
func (s Service) ListObligations(orgID uuid.UUID, filter ObligationFilter) ([]ObligationDetail, error) {
	details, err := s.repo.ListObligations(orgID, filter)
	if err != nil {
		return nil, FinanceInternalErr
	}
	return details, nil
}
func (s Service) FindObligation(orgID, id uuid.UUID) (ObligationDetail, error) {
	detail, err := s.repo.FindObligation(orgID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ObligationDetail{}, ObligationNotFoundErr
	}
	if err != nil {
		return ObligationDetail{}, FinanceInternalErr
	}
	return detail, nil
}

func (s Service) CreatePayment(orgID, obligationID uuid.UUID, actor *uuid.UUID, request CreatePaymentRequest) (ObligationDetail, error) {
	payment, err := paymentFromRequest(request, 0)
	if err != nil {
		return ObligationDetail{}, err
	}
	detail, err := s.repo.CreatePayment(orgID, obligationID, PaymentDraft{AmountCents: payment.AmountCents, PaidAt: payment.PaidAt, Method: payment.Method, Reference: payment.Reference, ActorUserID: actor})
	if errors.Is(err, sql.ErrNoRows) {
		return ObligationDetail{}, ObligationNotFoundErr
	}
	if errors.Is(err, PaymentExceedsBalanceErr) {
		return ObligationDetail{}, err
	}
	if err != nil {
		slog.Error("finance.CreatePayment", "err", err)
		return ObligationDetail{}, FinanceInternalErr
	}
	return detail, nil
}

func paymentFromRequest(request CreatePaymentRequest, max int64) (Payment, error) {
	if request.AmountCents <= 0 || !validPaymentMethod(request.Method) || (max > 0 && request.AmountCents > max) {
		if max > 0 && request.AmountCents > max {
			return Payment{}, PaymentExceedsBalanceErr
		}
		return Payment{}, InvalidPaymentErr
	}
	if request.PaidAt.IsZero() {
		request.PaidAt = time.Now()
	}
	return Payment{AmountCents: request.AmountCents, PaidAt: request.PaidAt, Method: request.Method, Reference: request.Reference}, nil
}
func validPaymentMethod(method PaymentMethod) bool {
	return method == PIX || method == Cash || method == BankTransfer || method == Card || method == Other
}
func statusFor(total, paid int64) ObligationStatus {
	if paid == 0 {
		return Open
	}
	if paid == total {
		return Paid
	}
	return PartiallyPaid
}
