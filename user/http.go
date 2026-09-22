package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"relay/shared"
	"strings"

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

type UserPayload struct {
	Name     string `validate:"required" json:"name"`
	Email    string `validate:"required,email" json:"email"`
	Password string `validate:"required,gte=8,lte=64" json:"password"`
	Document string `validate:"required" json:"document"`
	Phone    string `validate:"number" json:"phone"`
}

type SuccessResponse struct {
	Email string `json:"email"`
}

type ErrorResponse struct {
	Title      string            `json:"title"`
	Details    string            `json:"details"`
	Validation map[string]string `json:"validation"`
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user UserPayload

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(user)

	if err != nil {
		errRes := errorResponse(err)

		shared.JSONResponse(w, http.StatusBadRequest, errRes)
		return
	}

	email, err := h.service.Register(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := SuccessResponse{Email: email}
	shared.JSONResponse(w, http.StatusCreated, res)
}

func errorResponse(err error) ErrorResponse {
	var ve validator.ValidationErrors
	validation := make(map[string]string)

	if errors.As(err, &ve) {
		for _, fe := range ve {
			msg := errorMessage(fe)
			field := strings.ToLower(fe.Field())
			validation[field] = msg
		}
	}
	errRes := ErrorResponse{
		Title:      "The payload is invalid",
		Details:    "The payload has one or more validation errors, please fix them and try again.",
		Validation: validation,
	}
	return errRes
}

func errorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("The field %s is required", strings.ToLower(fe.Field()))

	case "email":
		return "Please provide a valid email address"

	default:
		return fmt.Sprintf("Field validation failed on '%s'", fe.Tag())
	}
}
