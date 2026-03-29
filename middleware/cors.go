package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

func Cors(next http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	return c.Handler(next)
}
