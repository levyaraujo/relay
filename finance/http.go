package finance

import (
	"encoding/json"
	"errors"
	"net/http"
	"relay/auth"
	"relay/shared"
	"strings"
	"time"
	"uuid"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	validate *validator.Validate
	service  *Service
}

type CreateTransactionRequest struct {
	Type           TransactionType                `json:"type" validate:"required,oneof=SALE PURCHASE EXPENSE"`
	PartyID        *uuid.UUID                     `json:"partyId"`
	OccurredAt     time.Time                      `json:"occurredAt"`
	DueDate        *time.Time                     `json:"dueDate"`
	Description    string                         `json:"description"`
	Currency       string                         `json:"currency" validate:"omitempty,len=3"`
	Lines          []CreateTransactionLineRequest `json:"lines" validate:"required,min=1,dive"`
	InitialPayment *CreatePaymentRequest          `json:"initialPayment"`
}

type CreateTransactionLineRequest struct {
	ItemID         *uuid.UUID `json:"itemId"`
	Description    string     `json:"description"`
	Quantity       int64      `json:"quantity" validate:"gte=1"`
	UnitPriceCents int64      `json:"unitPriceCents" validate:"gte=0"`
}

type CreatePaymentRequest struct {
	AmountCents int64         `json:"amountCents" validate:"gt=0"`
	PaidAt      time.Time     `json:"paidAt"`
	Method      PaymentMethod `json:"method" validate:"required,oneof=PIX CASH BANK_TRANSFER CARD OTHER"`
	Reference   string        `json:"reference"`
}

type TransactionFilter struct {
	Type   *TransactionType
	Status *TransactionStatus
}

type ObligationFilter struct {
	Direction *ObligationDirection
	Status    *ObligationStatus
	PartyID   *uuid.UUID
	DueBefore *time.Time
	DueAfter  *time.Time
}

func NewHandler(validate *validator.Validate, service *Service) *Handler {
	return &Handler{validate: validate, service: service}
}

func (h Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	orgID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	var request CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.validate.Struct(request); err != nil {
		shared.JSONResponse(w, http.StatusBadRequest, shared.ValidationResponse(err, financeErrorMessage))
		return
	}
	var actor *uuid.UUID
	if id, ok := auth.UserIDFromContext(r.Context()); ok {
		actor = &id
	}
	detail, err := h.service.CreateTransaction(orgID, actor, request)
	if err != nil {
		writeFinanceError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusCreated, detail)
}

func (h Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	orgID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	filter := TransactionFilter{}
	if value := strings.TrimSpace(r.URL.Query().Get("type")); value != "" {
		v := TransactionType(value)
		if v != Sale && v != Purchase && v != Expense {
			shared.JSONError(w, http.StatusBadRequest, InvalidFilterErr)
			return
		}
		filter.Type = &v
	}
	if value := strings.TrimSpace(r.URL.Query().Get("status")); value != "" {
		v := TransactionStatus(value)
		if v != Confirmed && v != Cancelled {
			shared.JSONError(w, http.StatusBadRequest, InvalidFilterErr)
			return
		}
		filter.Status = &v
	}
	details, err := h.service.ListTransactions(orgID, filter)
	if err != nil {
		writeFinanceError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusOK, details)
}

func (h Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	orgID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}
	detail, err := h.service.FindTransaction(orgID, id)
	if err != nil {
		writeFinanceError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusOK, detail)
}

func (h Handler) ListObligations(w http.ResponseWriter, r *http.Request) {
	orgID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	filter := ObligationFilter{}
	if value := strings.TrimSpace(r.URL.Query().Get("direction")); value != "" {
		v := ObligationDirection(value)
		if v != Receivable && v != Payable {
			shared.JSONError(w, http.StatusBadRequest, InvalidFilterErr)
			return
		}
		filter.Direction = &v
	}
	if value := strings.TrimSpace(r.URL.Query().Get("status")); value != "" {
		v := ObligationStatus(value)
		if v != Open && v != PartiallyPaid && v != Paid && v != ObligationCancelled {
			shared.JSONError(w, http.StatusBadRequest, InvalidFilterErr)
			return
		}
		filter.Status = &v
	}
	if value := strings.TrimSpace(r.URL.Query().Get("partyId")); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			shared.JSONError(w, http.StatusBadRequest, err)
			return
		}
		filter.PartyID = &id
	}
	if value := strings.TrimSpace(r.URL.Query().Get("dueBefore")); value != "" {
		date, err := time.Parse("2006-01-02", value)
		if err != nil {
			shared.JSONError(w, http.StatusBadRequest, err)
			return
		}
		filter.DueBefore = &date
	}
	if value := strings.TrimSpace(r.URL.Query().Get("dueAfter")); value != "" {
		date, err := time.Parse("2006-01-02", value)
		if err != nil {
			shared.JSONError(w, http.StatusBadRequest, err)
			return
		}
		filter.DueAfter = &date
	}
	details, err := h.service.ListObligations(orgID, filter)
	if err != nil {
		writeFinanceError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusOK, details)
}

func (h Handler) GetObligation(w http.ResponseWriter, r *http.Request) {
	orgID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}
	detail, err := h.service.FindObligation(orgID, id)
	if err != nil {
		writeFinanceError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusOK, detail)
}

func (h Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	orgID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}
	var request CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.validate.Struct(request); err != nil {
		shared.JSONResponse(w, http.StatusBadRequest, shared.ValidationResponse(err, financeErrorMessage))
		return
	}
	var actor *uuid.UUID
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		actor = &userID
	}
	detail, err := h.service.CreatePayment(orgID, id, actor, request)
	if err != nil {
		writeFinanceError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusCreated, detail)
}

func financeErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "oneof":
		return "This field has an invalid value"
	case "gte":
		return "This field must be greater than or equal to the minimum"
	case "gt":
		return "This field must be greater than zero"
	case "len":
		return "Currency must have three characters"
	default:
		return "This field is invalid"
	}
}

func writeFinanceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, InvalidTransactionTypeErr), errors.Is(err, InvalidTransactionLineErr), errors.Is(err, InvalidPaymentErr), errors.Is(err, InvalidCurrencyErr), errors.Is(err, InvalidFilterErr):
		status = http.StatusBadRequest
	case errors.Is(err, TransactionNotFoundErr), errors.Is(err, ObligationNotFoundErr):
		status = http.StatusNotFound
	case errors.Is(err, PartyRoleConflictErr), errors.Is(err, PaymentExceedsBalanceErr), errors.Is(err, TransactionConflictErr):
		status = http.StatusConflict
	}
	shared.JSONError(w, status, err)
}
