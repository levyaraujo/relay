package organization

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"relay/shared"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	validate *validator.Validate
	service  *Service
}

func NewHandler(v *validator.Validate, s *Service) *Handler {
	return &Handler{
		validate: v,
		service:  s,
	}
}

type ValidationErrResponse struct {
	Title      string            `json:"title"`
	Details    string            `json:"details"`
	Validation map[string]string `json:"validation"`
}

type OrganizationPayload struct {
	Name        string  `json:"name" validate:"required"`
	Website     string  `json:"website"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email" validate:"required,email"`
	TaxID       string  `json:"taxId" validate:"required"`
	Currency    string  `json:"currency"`
	Type        OrgType `json:"type"`
	Description string  `json:"description"`
}

type UserPayload struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=8,lte=64"`
	Document string `json:"document" validate:"required"`
	Phone    string `json:"phone" validate:"number"`
}

type CreateOrganizationPayload struct {
	Organization OrganizationPayload `json:"organization" validate:"required"`
	User         UserPayload         `json:"user" validate:"required"`
}

func (h Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var payload CreateOrganizationPayload

	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		shared.JSONError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validate.Struct(payload); err != nil {
		shared.JSONResponse(w, http.StatusBadRequest, validationResponse(err))
		return
	}

	if err := h.service.Create(payload); err != nil {
		writeOrganizationError(w, err)
		return
	}

	shared.JSONSuccess(
		w,
		http.StatusCreated,
		fmt.Sprintf("The organization %s was created successfully!", payload.Organization.Name),
	)
}

func errorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("The field %s is required", fe.Field())

	case "email":
		return "Please provide a valid email address"

	case "type":
		return "Please provide a valid organization type: SERVICES or PRODUCTS"

	default:
		return fmt.Sprintf("Field validation failed on '%s'", fe.Tag())
	}
}

func validationResponse(err error) ValidationErrResponse {
	var validationErrors validator.ValidationErrors
	validation := make(map[string]string)

	if errors.As(err, &validationErrors) {
		for _, fieldError := range validationErrors {
			validation[fieldError.Field()] = errorMessage(fieldError)
		}
	}

	return ValidationErrResponse{
		Title:      "The payload is invalid",
		Details:    "The payload has one or more validation errors, please fix them and try again.",
		Validation: validation,
	}
}

func writeOrganizationError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, InvalidTaxIDErr),
		errors.Is(err, InvalidOrgTypeErr):
		status = http.StatusBadRequest
	}

	shared.JSONError(w, status, err)
}
