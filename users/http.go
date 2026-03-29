package users

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
)

type Handler struct {
	ctrl *Controller
}

func NewHandler(controller *Controller) *Handler {
	return &Handler{ctrl: controller}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/me", h.UserInfo)
}

func (h *Handler) UserInfo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(shared.UserID).(uuid.UUID)

	u, err := h.ctrl.GetUser(userID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}
