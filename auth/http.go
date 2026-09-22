package auth

import (
	"encoding/json"
	"net/http"
	"relay/shared"

	"github.com/go-playground/validator/v10"
)

type LoginPayload struct {
	Email    string `validate:"required,email" json:"email"`
	Password string `validate:"required" json:"password"`
}

type LoginErrorResponse struct {
	Message string `json:"message"`
}

type JWTResponse struct {
	AccessToken string `json:"accessToken"`
}

type Handler struct {
	v *validator.Validate
	s *Service
}

func NewHandler(v *validator.Validate, s *Service) *Handler {
	return &Handler{
		v,
		s,
	}
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var l LoginPayload

	err := json.NewDecoder(r.Body).Decode(&l)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.v.Struct(l)

	if err != nil {
		res := LoginErrorResponse{Message: "Please, provide valid email and password"}

		shared.JSONResponse(w, http.StatusBadRequest, res)
		return
	}

	t, err := h.s.Login(l)

	if err != nil {
		res := LoginErrorResponse{Message: "Something went wrong. Please try again later."}
		shared.JSONResponse(w, http.StatusInternalServerError, res)
		return
	}

	res := JWTResponse{AccessToken: t}

	shared.JSONResponse(w, http.StatusOK, res)
}
