package catalog

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

type CreateItemRequest struct {
	Name              string `json:"name" validate:"required"`
	Description       string `json:"description"`
	Kind              Kind   `json:"kind" validate:"required,oneof=PRODUCT SERVICE"`
	DefaultPriceCents int64  `json:"defaultPriceCents" validate:"gte=0"`
	Currency          string `json:"currency" validate:"omitempty,len=3"`
}

func NewHandler(validate *validator.Validate, service *Service) *Handler {
	return &Handler{validate: validate, service: service}
}

func (h Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	var request CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.validate.Struct(request); err != nil {
		shared.JSONResponse(w, http.StatusBadRequest, shared.ValidationResponse(err, itemErrorMessage))
		return
	}
	var actorUserID *uuid.UUID
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		actorUserID = &userID
	}
	item, err := h.service.Create(organizationID, actorUserID, request)
	if err != nil {
		writeItemError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusCreated, item)
}

func (h Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	items, err := h.service.List(organizationID)
	if err != nil {
		writeItemError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusOK, items)
}

func (h Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := auth.OrganizationIDFromContext(r.Context())
	if !ok {
		shared.JSONError(w, http.StatusUnauthorized, errors.New("organization context is missing"))
		return
	}
	itemID, err := uuid.Parse(strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}
	item, err := h.service.FindByID(organizationID, itemID)
	if err != nil {
		writeItemError(w, err)
		return
	}
	shared.JSONResponse(w, http.StatusOK, item)
}

func itemErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "oneof":
		return "Kind must be PRODUCT or SERVICE"
	case "gte":
		return "This field must be greater than or equal to zero"
	case "len":
		return "Currency must have three characters"
	default:
		return "This field is invalid"
	}
}

func writeItemError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, InvalidKindErr), errors.Is(err, InvalidPriceErr), errors.Is(err, InvalidCurrencyErr):
		status = http.StatusBadRequest
	case errors.Is(err, ItemNotFoundErr):
		status = http.StatusNotFound
	case errors.Is(err, ItemConflictErr):
		status = http.StatusConflict
	}
	shared.JSONError(w, status, err)
}
