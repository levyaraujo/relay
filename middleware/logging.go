package middleware

import (
	"log"
	"net/http"
	"time"
)

type wrappedWriter struct {
	http.ResponseWriter
	status int
}

func (w *wrappedWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := &wrappedWriter{ResponseWriter: w, status: http.StatusOK}

		start := time.Now()
		next.ServeHTTP(wrapped, r)

		params := r.URL.Query().Encode()
		if params != "" {
			log.Printf("[%s] %s?%s %d | %v\n", r.Method, r.URL.Path, params, wrapped.status, time.Since(start))
		} else {
			log.Printf("[%s] %s %d | %v\n", r.Method, r.URL.Path, wrapped.status, time.Since(start))
		}
	})
}
