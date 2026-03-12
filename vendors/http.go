package vendors

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// Handler holds HTTP handlers for vendor endpoints.
type Handler struct {
	ctrl *Controller
}

// NewHandler returns a Handler wired to the given controller.
func NewHandler(ctrl *Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

// RegisterRoutes registers vendor routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /vendors", h.create)
	mux.HandleFunc("GET /vendors/{id}", h.getByID)
	mux.HandleFunc("PUT /vendors/{id}", h.update)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var v Vendor
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.ctrl.Create(&v)
	if err != nil {
		if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrInvalidPaymentTerms) || errors.Is(err, ErrInvalidCNPJ) {
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

	v, err := h.ctrl.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var v Vendor
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	v.Id = id

	updated, err := h.ctrl.Update(&v)
	if err != nil {
		if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrInvalidPaymentTerms) {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
