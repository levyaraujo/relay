package finance

import (
	"testing"
	"uuid"
)

type recordingRepository struct {
	draft      TransactionDraft
	partyRole  bool
	itemExists bool
}

func (r *recordingRepository) CreateTransaction(_ uuid.UUID, draft TransactionDraft) (TransactionDetail, error) {
	r.draft = draft
	return TransactionDetail{Transaction: draft.Transaction}, nil
}
func (r *recordingRepository) ListTransactions(uuid.UUID, TransactionFilter) ([]TransactionDetail, error) {
	return nil, nil
}
func (r *recordingRepository) FindTransaction(uuid.UUID, uuid.UUID) (TransactionDetail, error) {
	return TransactionDetail{}, nil
}
func (r *recordingRepository) ListObligations(uuid.UUID, ObligationFilter) ([]ObligationDetail, error) {
	return nil, nil
}
func (r *recordingRepository) FindObligation(uuid.UUID, uuid.UUID) (ObligationDetail, error) {
	return ObligationDetail{}, nil
}
func (r *recordingRepository) CreatePayment(uuid.UUID, uuid.UUID, PaymentDraft) (ObligationDetail, error) {
	return ObligationDetail{}, nil
}
func (r *recordingRepository) PartyHasRole(uuid.UUID, uuid.UUID, string) (bool, error) {
	return r.partyRole, nil
}
func (r *recordingRepository) ItemExists(uuid.UUID, uuid.UUID) (bool, error) {
	return r.itemExists, nil
}

func TestCreateSaleCalculatesLineAndHeaderTotals(t *testing.T) {
	itemID, partyID := uuid.New(), uuid.New()
	repo := &recordingRepository{partyRole: true, itemExists: true}
	service := Service{repo: repo}
	detail, err := service.CreateTransaction(uuid.New(), nil, CreateTransactionRequest{Type: Sale, PartyID: &partyID, Lines: []CreateTransactionLineRequest{{ItemID: &itemID, Quantity: 2, UnitPriceCents: 1500}}})
	if err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}
	if detail.Transaction.TotalCents != 3000 || repo.draft.Lines[0].TotalCents != 3000 {
		t.Fatalf("totals = %d/%d, want 3000", detail.Transaction.TotalCents, repo.draft.Lines[0].TotalCents)
	}
	if repo.draft.Obligation.Direction != Receivable {
		t.Fatalf("direction = %s, want %s", repo.draft.Obligation.Direction, Receivable)
	}
}

func TestCreatePurchaseRequiresSupplierRole(t *testing.T) {
	partyID, itemID := uuid.New(), uuid.New()
	service := Service{repo: &recordingRepository{partyRole: false, itemExists: true}}
	_, err := service.CreateTransaction(uuid.New(), nil, CreateTransactionRequest{Type: Purchase, PartyID: &partyID, Lines: []CreateTransactionLineRequest{{ItemID: &itemID, Quantity: 1, UnitPriceCents: 1}}})
	if err != PartyRoleConflictErr {
		t.Fatalf("error = %v, want %v", err, PartyRoleConflictErr)
	}
}

func TestCreatePaymentRejectsOverpayment(t *testing.T) {
	if _, err := paymentFromRequest(CreatePaymentRequest{AmountCents: 101, Method: PIX}, 100); err != PaymentExceedsBalanceErr {
		t.Fatalf("error = %v, want %v", err, PaymentExceedsBalanceErr)
	}
}

func TestCreateTransactionAllowsZeroUnitPriceWithoutPanicking(t *testing.T) {
	itemID, partyID := uuid.New(), uuid.New()
	service := Service{repo: &recordingRepository{partyRole: true, itemExists: true}}
	if _, err := service.CreateTransaction(uuid.New(), nil, CreateTransactionRequest{
		Type: Sale, PartyID: &partyID,
		Lines: []CreateTransactionLineRequest{{ItemID: &itemID, Quantity: 1, UnitPriceCents: 0}},
	}); err != InvalidTransactionLineErr {
		t.Fatalf("error = %v, want %v", err, InvalidTransactionLineErr)
	}
}
