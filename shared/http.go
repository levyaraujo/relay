package shared

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func JSONResponse[T any](w http.ResponseWriter, status int, response T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func JSONError(w http.ResponseWriter, status int, err error) {
	JSONResponse(w, status, ErrorResponse{
		Error: err.Error(),
	})
}

func JSONSuccess(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, SuccessResponse{Message: message})
}
