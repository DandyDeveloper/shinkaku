package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/shinkaku/backend/config"
	"github.com/user/shinkaku/backend/internal/api"
	"github.com/user/shinkaku/backend/internal/auth"
	"github.com/user/shinkaku/backend/internal/db"
	"github.com/user/shinkaku/backend/internal/llm"
)

func main() {
	cfg := config.Load()

	// Open database (auto-migrates on first run).
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Build Ollama client.
	ollamaClient := llm.New(cfg.OllamaURL, cfg.OllamaModel)

	// Build auth handler (fetches Google OIDC discovery document at startup).
	authHandler, err := auth.NewHandler(cfg)
	if err != nil {
		log.Fatalf("failed to initialize auth: %v", err)
	}

	// Build router.
	router := api.NewRouter(database, ollamaClient, authHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second, // generous for LLM calls
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so we can listen for shutdown signals.
	go func() {
		log.Printf("shinkaku backend listening on http://localhost:%s", cfg.Port)
		log.Printf("ollama: %s  model: %s", cfg.OllamaURL, cfg.OllamaModel)
		log.Printf("database: %s", cfg.DBPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for SIGINT or SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
