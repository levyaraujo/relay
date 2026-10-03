package shared

import (
	"encoding/json"
	"net/http"
)

func JSONResponse[T any](w http.ResponseWriter, status int, res T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
