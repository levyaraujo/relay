package party

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"relay/auth"
	"testing"
	"uuid"

	"github.com/go-playground/validator/v10"
)

type recordingRepository struct {
	called         int
	organizationID uuid.UUID
	request        CreatePartyRequest
	party          Party
}

func (r *recordingRepository) Create(organizationID uuid.UUID, _ *uuid.UUID, request CreatePartyRequest) (Party, error) {
	r.called++
	r.organizationID = organizationID
	r.request = request
	return r.party, nil
}

func (r *recordingRepository) List(uuid.UUID, *Role) ([]Party, error) {
	return nil, nil
}

func (r *recordingRepository) FindByID(uuid.UUID, uuid.UUID) (Party, error) {
	return r.party, nil
}

func TestCreatePartyNormalizesDocumentAndUsesAuthenticatedOrganization(t *testing.T) {
	organizationID := uuid.New()
	repo := &recordingRepository{party: Party{ID: uuid.New(), OrganizationID: organizationID, Name: "João"}}
	service := Service{repo: repo}

	_, err := service.Create(organizationID, nil, CreatePartyRequest{
		Name:     "João",
		Document: "529.982.247-25",
		Roles:    []Role{CustomerRole, SupplierRole},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repo.organizationID != organizationID {
		t.Fatalf("organization ID = %s, want %s", repo.organizationID, organizationID)
	}
	if repo.request.Document != "52998224725" {
		t.Fatalf("document = %q, want %q", repo.request.Document, "52998224725")
	}
	if len(repo.request.Roles) != 2 {
		t.Fatalf("roles = %v, want customer and supplier", repo.request.Roles)
	}
}

func TestCreatePartyHandlerUsesOrganizationContext(t *testing.T) {
	organizationID := uuid.New()
	repo := &recordingRepository{party: Party{ID: uuid.New(), OrganizationID: organizationID, Name: "AWS"}}
	handler := NewHandler(validator.New(validator.WithRequiredStructEnabled()), &Service{repo: repo})
	request := httptest.NewRequest(http.MethodPost, "/parties", bytes.NewBufferString(`{
		"name": "AWS",
		"document": "57082641000150",
		"roles": ["SUPPLIER"]
	}`))
	request = request.WithContext(auth.WithOrganizationID(request.Context(), organizationID))
	response := httptest.NewRecorder()

	handler.CreateParty(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if repo.called != 1 {
		t.Fatalf("Create() called %d times, want 1", repo.called)
	}
}
