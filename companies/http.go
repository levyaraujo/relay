package companies

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/levyaraujo/relay/accounts"
)

// Handler holds HTTP handlers for company endpoints.
type Handler struct {
	ctrl    *Controller
	accRepo accounts.AccountRepository
}

// NewHandler returns a Handler wired to the given controller.
func NewHandler(ctrl *Controller, accRepo accounts.AccountRepository) *Handler {
	return &Handler{ctrl: ctrl, accRepo: accRepo}
}

// RegisterRoutes registers company routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /companies", h.create)
	mux.HandleFunc("GET /companies/{id}", h.getByID)
	mux.HandleFunc("PUT /companies/{id}", h.update)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var co Company
	if err := json.NewDecoder(r.Body).Decode(&co); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.ctrl.Create(&co, h.accRepo)
	if err != nil {
		if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrInvalidCNPJ) {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	co, err := h.ctrl.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(co)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var co Company
	if err := json.NewDecoder(r.Body).Decode(&co); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	co.Id = id

	updated, err := h.ctrl.Update(&co)
	if err != nil {
		if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrInvalidCNPJ) {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
