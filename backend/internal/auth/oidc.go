package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"

	"github.com/user/nihongo-sensei/backend/config"
)

const (
	sessionCookieName = "nihongo_session"
	stateCookieName   = "nihongo_oauth_state"
	sessionDuration   = 24 * time.Hour
)

// Handler holds OIDC auth state and handles login/callback/logout HTTP routes.
type Handler struct {
	cfg      *config.Config
	oauth2   *oauth2.Config
	verifier *oidc.IDTokenVerifier
	sc       *securecookie.SecureCookie
}

type sessionData struct {
	Email   string
	Expires time.Time
}

// NewHandler initializes the OIDC provider and returns a Handler.
// Returns an error if GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET are not set.
func NewHandler(cfg *config.Config) (*Handler, error) {
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	keyBytes, err := hex.DecodeString(cfg.SessionSecret)
	if err != nil || len(keyBytes) < 32 {
		// Fall back to raw bytes, padded to 32.
		raw := []byte(cfg.SessionSecret)
		keyBytes = make([]byte, 32)
		copy(keyBytes, raw)
	}

	// Use same 32-byte key for HMAC signing and AES-256 encryption.
	// In production, consider deriving separate keys.
	sc := securecookie.New(keyBytes, keyBytes)

	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("creating OIDC provider: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.BaseURL + "/auth/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.GoogleClientID})

	return &Handler{
		cfg:      cfg,
		oauth2:   oauth2Config,
		verifier: verifier,
		sc:       sc,
	}, nil
}

// Login generates a random state, stores it in a short-lived cookie, and
// redirects the browser to Google's OAuth2 consent screen.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	state := base64.RawURLEncoding.EncodeToString(b)

	encoded, err := h.sc.Encode(stateCookieName, state)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    encoded,
		MaxAge:   300, // 5-minute window to complete the OAuth flow
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production (requires HTTPS)
		Path:     "/",
	})

	http.Redirect(w, r, h.oauth2.AuthCodeURL(state), http.StatusFound)
}

// Callback validates the OAuth2 state (CSRF guard), exchanges the code for
// tokens, verifies the ID token, enforces the email whitelist, and issues a
// signed session cookie on success.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	// Validate state before accepting the callback — prevents CSRF on the OAuth flow.
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		http.Error(w, "missing state cookie", http.StatusForbidden)
		return
	}

	var storedState string
	if err := h.sc.Decode(stateCookieName, stateCookie.Value, &storedState); err != nil {
		http.Error(w, "invalid state cookie", http.StatusForbidden)
		return
	}

	if r.URL.Query().Get("state") != storedState {
		http.Error(w, "state mismatch", http.StatusForbidden)
		return
	}

	// Clear the state cookie immediately after verification.
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	code := r.URL.Query().Get("code")
	token, err := h.oauth2.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("auth: token exchange failed: %v", err)
		http.Error(w, "token exchange failed", http.StatusForbidden)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in response", http.StatusForbidden)
		return
	}

	idToken, err := h.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		log.Printf("auth: ID token verification failed: %v", err)
		http.Error(w, "token verification failed", http.StatusForbidden)
		return
	}

	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "failed to extract claims", http.StatusForbidden)
		return
	}

	if claims.Email != h.cfg.AllowedEmail {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	session := sessionData{
		Email:   claims.Email,
		Expires: time.Now().Add(sessionDuration),
	}
	encoded, err := h.sc.Encode(sessionCookieName, session)
	if err != nil {
		http.Error(w, "session creation failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production (requires HTTPS)
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

// Logout clears the session cookie and redirects to /.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
