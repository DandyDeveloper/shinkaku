package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type contextKey string

// EmailKey is the context key under which the authenticated user's email is stored.
const EmailKey contextKey = "email"

// EmailFromContext retrieves the authenticated email from a request context.
func EmailFromContext(ctx context.Context) string {
	v, _ := ctx.Value(EmailKey).(string)
	return v
}

// RequireAuth validates the session cookie. On failure it returns 401 JSON
// {"error": "unauthorized"}. On success it attaches the email to the request context.
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			writeUnauthorized(w)
			return
		}

		var session sessionData
		if err := h.sc.Decode(sessionCookieName, cookie.Value, &session); err != nil {
			writeUnauthorized(w)
			return
		}

		if time.Now().After(session.Expires) {
			writeUnauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), EmailKey, session.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
