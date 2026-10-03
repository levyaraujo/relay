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

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(payload)

	if err != nil {
		errRes := errorResponse(err)

		shared.JSONResponse(w, http.StatusBadRequest, errRes)
		return
	}

	err = h.service.Create(payload)

	if err != nil {
		errMsg, status := statusFromOrganizationError(err)
		res := ErrorResponse{Error: errMsg}

		shared.JSONResponse(w, status, res)
		return
	}

	res := SuccessResponse{Message: fmt.Sprintf("The organization %s was created successfully!", payload.Organization.Name)}
	shared.JSONResponse(w, http.StatusCreated, res)
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

func errorResponse(err error) ValidationErrResponse {
	var ve validator.ValidationErrors
	validation := make(map[string]string)

	if errors.As(err, &ve) {
		for _, fe := range ve {
			msg := errorMessage(fe)
			field := fe.Field()
			validation[field] = msg
		}
	}
	errRes := ValidationErrResponse{
		Title:      "The payload is invalid",
		Details:    "The payload has one or more validation errors, please fix them and try again.",
		Validation: validation,
	}
	return errRes
}

func statusFromOrganizationError(err error) (string, int) {
	switch {
	case errors.Is(err, InvalidTaxIDErr),
		errors.Is(err, InvalidOrgTypeErr):
		return err.Error(), http.StatusBadRequest

	case errors.Is(err, OrgRegistrationErr):
		return err.Error(), http.StatusInternalServerError

	default:
		return err.Error(), http.StatusInternalServerError
	}
}
