package finance

import (
	"time"
	"uuid"
)

type TransactionType string
type TransactionStatus string
type ObligationDirection string
type ObligationStatus string
type PaymentMethod string

const (
	Sale     TransactionType = "SALE"
	Purchase TransactionType = "PURCHASE"
	Expense  TransactionType = "EXPENSE"

	Confirmed TransactionStatus = "CONFIRMED"
	Cancelled TransactionStatus = "CANCELLED"

	Receivable ObligationDirection = "RECEIVABLE"
	Payable    ObligationDirection = "PAYABLE"

	Open                ObligationStatus = "OPEN"
	PartiallyPaid       ObligationStatus = "PARTIALLY_PAID"
	Paid                ObligationStatus = "PAID"
	ObligationCancelled ObligationStatus = "CANCELLED"

	PIX          PaymentMethod = "PIX"
	Cash         PaymentMethod = "CASH"
	BankTransfer PaymentMethod = "BANK_TRANSFER"
	Card         PaymentMethod = "CARD"
	Other        PaymentMethod = "OTHER"
)

type Transaction struct {
	ID             uuid.UUID         `json:"id"`
	OrganizationID uuid.UUID         `json:"organizationId"`
	PartyID        *uuid.UUID        `json:"partyId,omitempty"`
	Type           TransactionType   `json:"type"`
	Status         TransactionStatus `json:"status"`
	OccurredAt     time.Time         `json:"occurredAt"`
	Description    string            `json:"description"`
	Currency       string            `json:"currency"`
	TotalCents     int64             `json:"totalCents"`
}

type TransactionLine struct {
	ID             uuid.UUID  `json:"id"`
	TransactionID  uuid.UUID  `json:"transactionId"`
	ItemID         *uuid.UUID `json:"itemId,omitempty"`
	Description    string     `json:"description"`
	Quantity       int64      `json:"quantity"`
	UnitPriceCents int64      `json:"unitPriceCents"`
	TotalCents     int64      `json:"totalCents"`
}

type Obligation struct {
	ID              uuid.UUID           `json:"id"`
	OrganizationID  uuid.UUID           `json:"organizationId"`
	TransactionID   uuid.UUID           `json:"transactionId"`
	PartyID         *uuid.UUID          `json:"partyId,omitempty"`
	Direction       ObligationDirection `json:"direction"`
	Status          ObligationStatus    `json:"status"`
	AmountCents     int64               `json:"amountCents"`
	PaidAmountCents int64               `json:"paidAmountCents"`
	Currency        string              `json:"currency"`
	DueDate         *time.Time          `json:"dueDate,omitempty"`
}

type Payment struct {
	ID             uuid.UUID     `json:"id"`
	ObligationID   uuid.UUID     `json:"obligationId"`
	OrganizationID uuid.UUID     `json:"organizationId"`
	AmountCents    int64         `json:"amountCents"`
	PaidAt         time.Time     `json:"paidAt"`
	Method         PaymentMethod `json:"method"`
	Reference      string        `json:"reference"`
}

type TransactionDetail struct {
	Transaction Transaction       `json:"transaction"`
	Lines       []TransactionLine `json:"lines"`
	Obligation  *Obligation       `json:"obligation,omitempty"`
	Payments    []Payment         `json:"payments,omitempty"`
}

type ObligationDetail struct {
	Obligation Obligation `json:"obligation"`
	Payments   []Payment  `json:"payments"`
}

type TransactionDraft struct {
	Transaction    Transaction
	Lines          []TransactionLine
	Obligation     Obligation
	InitialPayment *Payment
	ActorUserID    *uuid.UUID
}

type PaymentDraft struct {
	AmountCents int64
	PaidAt      time.Time
	Method      PaymentMethod
	Reference   string
	ActorUserID *uuid.UUID
}
