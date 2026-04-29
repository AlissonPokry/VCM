package middleware

import "net/http"

// N8NAuth validates the x-n8n-secret header for all n8n routes.
func N8NAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" || r.Header.Get("x-n8n-secret") != secret {
				RespondError(w, http.StatusUnauthorized, "Unauthorized", "INVALID_N8N_SECRET")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
