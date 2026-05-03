package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/shinkaku/backend/internal/db"
)

type onboardingSeedRequest struct {
	StartingJLPT string `json:"starting_jlpt"`
}

func getOnboardingStatus(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		grammarCount, err := database.StarterGrammarCount()
		if err != nil {
			jsonError(w, "failed to read onboarding status", http.StatusInternalServerError)
			return
		}

		jsonOK(w, map[string]any{
			"needs_onboarding": grammarCount == 0,
			"grammar_count":    grammarCount,
		})
	}
}

func completeOnboarding(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req onboardingSeedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}

		inserted, err := database.SeedStarterContentForJLPT(req.StartingJLPT)
		if err != nil {
			jsonError(w, "failed to complete onboarding: "+err.Error(), http.StatusBadRequest)
			return
		}

		jsonOK(w, map[string]any{
			"inserted": inserted,
		})
	}
}