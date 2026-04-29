package middleware

import (
	"net/http"

	chicors "github.com/go-chi/cors"
)

// CORS returns the configured cross-origin middleware.
func CORS(origin string) func(http.Handler) http.Handler {
	return chicors.Handler(chicors.Options{
		AllowedOrigins:   []string{origin},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "x-n8n-secret"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}
