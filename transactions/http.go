package transactions

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	repo TransactionRepository
}

func NewHandler(repo TransactionRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /transactions", h.create)
	mux.HandleFunc("GET /transactions/{id}", h.getByID)
	mux.HandleFunc("PUT /transactions/{id}", h.update)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var t Transaction
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.repo.Create(&t)
	if err != nil {
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

	t, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var transaction Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	transaction.Id = id

	updated, err := h.repo.Update(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
