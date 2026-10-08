package catalog

import (
	"testing"
	"uuid"
)

type recordingRepository struct {
	request        CreateItemRequest
	organizationID uuid.UUID
	item           Item
}

func (r *recordingRepository) Create(organizationID uuid.UUID, _ *uuid.UUID, request CreateItemRequest) (Item, error) {
	r.organizationID, r.request = organizationID, request
	return r.item, nil
}
func (r *recordingRepository) List(uuid.UUID) ([]Item, error)              { return nil, nil }
func (r *recordingRepository) FindByID(uuid.UUID, uuid.UUID) (Item, error) { return r.item, nil }

func TestCreateDefaultsCurrencyAndScopesOrganization(t *testing.T) {
	organizationID := uuid.New()
	repo := &recordingRepository{item: Item{ID: uuid.New(), OrganizationID: organizationID}}
	service := Service{repo: repo}
	_, err := service.Create(organizationID, nil, CreateItemRequest{Name: "Consulting", Kind: ServiceKind})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repo.organizationID != organizationID {
		t.Fatalf("organization ID = %s, want %s", repo.organizationID, organizationID)
	}
	if repo.request.Currency != "BRL" {
		t.Fatalf("currency = %q, want BRL", repo.request.Currency)
	}
}

func TestCreateRejectsNegativePrice(t *testing.T) {
	service := Service{repo: &recordingRepository{}}
	if _, err := service.Create(uuid.New(), nil, CreateItemRequest{Kind: ProductKind, DefaultPriceCents: -1}); err != InvalidPriceErr {
		t.Fatalf("error = %v, want %v", err, InvalidPriceErr)
	}
}
