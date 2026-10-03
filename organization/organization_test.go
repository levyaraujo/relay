package organization

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestOrganization(t *testing.T) {
	t.Run("cnpj validation", func(t *testing.T) {
		tests := []struct {
			cnpj string
			want bool
		}{
			{"VVJ6HCZA000112", true},
			{"57082641000150", true},
			{"WWY8WVB3000171", false},
		}

		for _, tt := range tests {
			got := ValidateCNPJ(tt.cnpj)

			if got != tt.want {
				t.Errorf("got %t, want %t", got, tt.want)
			}
		}
	})

	t.Run("cpf validation", func(t *testing.T) {
		tests := []struct {
			cpf  string
			want bool
		}{
			{"52998224725", true},
			{"529.982.247-25", true},
			{"52998224724", false},
			{"5299822472", false},
			{"5299822472A", false},
			{"11111111111", false},
		}

		for _, tt := range tests {
			got := ValidateCPF(tt.cpf)

			if got != tt.want {
				t.Errorf("got %t, want %t", got, tt.want)
			}
		}
	})

	t.Run("validate organization type", func(t *testing.T) {
		tests := []struct {
			oType string
			want  error
		}{
			{"nada", InvalidOrgTypeErr},
			{"MARKETING", InvalidOrgTypeErr},
			{"SERVICES", nil},
			{"PRODUCTS", nil},
		}

		for _, tt := range tests {
			got := ValidateOrgType(OrgType(tt.oType))

			if got != tt.want {
				t.Errorf("got %t, want %t", got, tt.want)
			}
		}
	})
}

type recordingRepository struct {
	called  int
	payload CreateOrganizationPayload
}

func (r *recordingRepository) CreateWithOwner(payload CreateOrganizationPayload) error {
	r.called++
	r.payload = payload
	return nil
}

func TestCreateRegistersOrganizationAndOwnerTogether(t *testing.T) {
	repo := &recordingRepository{}
	service := Service{repo: repo}
	payload := CreateOrganizationPayload{
		Organization: OrganizationPayload{
			Name:  "Acme",
			Email: "org@example.com",
			TaxID: "57082641000150",
			Type:  PRODUCTS,
		},
		User: UserPayload{
			Name:     "Owner",
			Email:    "owner@example.com",
			Password: "a secure password",
			Document: "52998224725",
		},
	}

	if err := service.Create(payload); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repo.called != 1 {
		t.Fatalf("CreateWithOwner() called %d times, want 1", repo.called)
	}
	if repo.payload.Organization.Name != "Acme" {
		t.Errorf("organization name = %q, want %q", repo.payload.Organization.Name, "Acme")
	}
	if repo.payload.User.Email != "owner@example.com" {
		t.Errorf("owner email = %q, want %q", repo.payload.User.Email, "owner@example.com")
	}
	if repo.payload.User.Password == "a secure password" {
		t.Error("owner password was not hashed before persistence")
	}
}

func TestCreateOrganizationAcceptsOrganizationAndOwnerInOneRequest(t *testing.T) {
	repo := &recordingRepository{}
	handler := NewHandler(
		validator.New(validator.WithRequiredStructEnabled()),
		&Service{repo: repo},
	)
	request := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBufferString(`{
		"organization": {
			"name": "Acme",
			"email": "org@example.com",
			"taxId": "57082641000150",
			"type": "SERVICES"
		},
		"user": {
			"name": "Owner",
			"email": "owner@example.com",
			"password": "a secure password",
			"document": "52998224725",
			"phone": "5511999999999"
		}
	}`))
	response := httptest.NewRecorder()

	handler.CreateOrganization(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if repo.called != 1 {
		t.Fatalf("CreateWithOwner() called %d times, want 1", repo.called)
	}
}
