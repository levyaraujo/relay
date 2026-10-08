package party

import (
	"encoding/json"
	"errors"
	"net/http"
	"relay/auth"
	"relay/shared"
	"strings"
	"uuid"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	validate *validator.Validate
	service  *Service
}

type CreatePartyRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"omitempty,email"`
	Phone    string `json:"phone"`
	Document string `json:"document" validate:"required"`
	Roles    []Role `json:"roles" validate:"required,min=1,dive,oneof=CUSTOMER SUPPLIER"`
}

type PartyResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Document       string    `json:"document"`
	Roles          []Role    `json:"roles"`
}

func NewHandler(validate *validator.Validate, service *Service) *Handler {
	return &Handler{validate: validate, service: service}
}

func (h Handler) CreateParty(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}

	var request CreatePartyRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.validate.Struct(request); err != nil {
		shared.JSONResponse(w, http.StatusBadRequest, shared.ValidationResponse(err, errorMessage))
		return
	}

	var actorUserID *uuid.UUID
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		actorUserID = &userID
	}
	party, err := h.service.Create(organizationID, actorUserID, request)
	if err != nil {
		writePartyError(w, err)
		return
	}

	shared.JSONResponse(w, http.StatusCreated, toResponse(party))
}

func (h Handler) ListParties(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}

	var role *Role
	if value := strings.TrimSpace(r.URL.Query().Get("role")); value != "" {
		candidate := Role(value)
		if err := validateRole(candidate); err != nil {
			writePartyError(w, err)
			return
		}
		role = &candidate
	}

	parties, err := h.service.List(organizationID, role)
	if err != nil {
		writePartyError(w, err)
		return
	}

	responses := make([]PartyResponse, 0, len(parties))
	for _, party := range parties {
		responses = append(responses, toResponse(party))
	}
	shared.JSONResponse(w, http.StatusOK, responses)
}

func (h Handler) GetParty(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}

	partyID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}

	party, err := h.service.FindByID(organizationID, partyID)
	if err != nil {
		writePartyError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusOK, toResponse(party))
}

func errorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Please provide a valid email address"
	case "oneof":
		return "Role must be CUSTOMER or SUPPLIER"
	default:
		return "This field is invalid"
	}
}

func writePartyError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, InvalidDocumentErr), errors.Is(err, InvalidRoleErr):
		status = http.StatusBadRequest
	case errors.Is(err, PartyNotFoundErr):
		status = http.StatusNotFound
	case errors.Is(err, PartyConflictErr):
		status = http.StatusConflict
	}
	shared.JSONError(w, status, err)
}

func toResponse(party Party) PartyResponse {
	return PartyResponse{
		ID:             party.ID,
		OrganizationID: party.OrganizationID,
		Name:           party.Name,
		Email:          party.Email,
		Phone:          party.Phone,
		Document:       party.Document,
		Roles:          party.Roles,
	}
}
