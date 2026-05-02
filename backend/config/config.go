package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Port               string
	DBPath             string
	OllamaURL          string
	OllamaModel        string
	GoogleClientID     string
	GoogleClientSecret string
	AllowedEmail       string
	// Set SESSION_SECRET to a random 32-byte hex value in production. Never commit this value.
	SessionSecret string
	BaseURL       string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	sessionSecret := getEnv("SESSION_SECRET", "")
	if sessionSecret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatal("failed to generate session secret")
		}
		sessionSecret = hex.EncodeToString(b)
		log.Println("WARNING: SESSION_SECRET not set — using ephemeral random value. Sessions will not survive restarts. Set SESSION_SECRET to a random 32-byte hex value in production.")
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBPath:             getEnv("DB_PATH", "./nihongo.db"),
		OllamaURL:          getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:        getEnv("OLLAMA_MODEL", "llama3.2"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		AllowedEmail:       getEnv("ALLOWED_EMAIL", ""),
		SessionSecret:      sessionSecret,
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
