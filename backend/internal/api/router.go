package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/user/shinkaku/backend/internal/auth"
	"github.com/user/shinkaku/backend/internal/db"
	"github.com/user/shinkaku/backend/internal/llm"
)

// NewRouter wires all handlers onto a chi Mux and returns it.
func NewRouter(database *db.DB, ollamaClient *llm.Client, authHandler *auth.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(corsMiddleware)

	// Auth routes — no session middleware on these.
	r.Get("/auth/login", authHandler.Login)
	r.Get("/auth/callback", authHandler.Callback)
	r.Get("/auth/logout", authHandler.Logout)

	r.Route("/api", func(r chi.Router) {
		r.Use(authHandler.RequireAuth)

		// Grammar points
		r.Get("/grammar", listGrammar(database))
		r.Post("/grammar", createGrammar(database))
		r.Get("/grammar/{id}", getGrammar(database))

		// SRS review
		r.Get("/review/queue", getReviewQueue(database))
		r.Post("/review/{id}/grade", submitGrade(database))

		// LLM challenge
		r.Post("/challenge/grade", gradeChallenge(database, ollamaClient))
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// OpenAPI spec and Swagger UI — no auth required
	r.Get("/openapi.yaml", serveOpenAPISpec())
	r.Get("/docs", serveSwaggerUI())

	return r
}

// corsMiddleware adds permissive CORS headers for local development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
