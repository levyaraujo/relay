package user

import (
	"encoding/json"
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

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user UserPayload

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(user)

	if err != nil {
		shared.JSONResponse(
			w,
			http.StatusBadRequest,
			shared.ValidationResponse(err, errorMessage),
		)
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
