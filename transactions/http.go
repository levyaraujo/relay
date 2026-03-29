package transactions

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/shared/types"
)

type Handler struct {
	ctrl *TransactionController
}

func NewHandler(controller *TransactionController) *Handler {
	return &Handler{ctrl: controller}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /transactions", h.create)
	mux.HandleFunc("GET /transactions", h.handleListTransactions)
	mux.HandleFunc("GET /transactions/{id}", h.getByID)
	mux.HandleFunc("PUT /transactions/{id}", h.update)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var t Transaction
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.ctrl.Create(&t)
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

	t, err := h.ctrl.GetTransactionByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	start, _ := time.Parse(time.RFC3339, query.Get("from"))
	end, _ := time.Parse(time.RFC3339, query.Get("to"))
	interval := types.Interval{Start: start, End: end}
	coID := r.Context().Value(shared.CompanyID).(uuid.UUID)

	txs, err := h.ctrl.TransactionsByDateRange(coID, interval)

	if err != nil {
		if errors.Is(err, ErrTxsNotFound) {
			shared.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		shared.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	shared.JSON(w, http.StatusOK, map[string]interface{}{"transactions": txs})
	return
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

	updated, err := h.ctrl.UpdateTransaction(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
