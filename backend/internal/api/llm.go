package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/shinkaku/backend/internal/db"
	"github.com/user/shinkaku/backend/internal/llm"
	"github.com/user/shinkaku/backend/internal/models"
)

func loadGrammarPoint(database *db.DB, r *http.Request, grammarPointID int64) (*models.GrammarPoint, error) {
	var gp models.GrammarPoint
	row := database.QueryRowContext(r.Context(),
		`SELECT id, jlpt_level, pattern, meaning, example_jp, example_en, notes, source, created_at
		 FROM grammar_points WHERE id = ?`, grammarPointID)
	if err := row.Scan(&gp.ID, &gp.JLPTLevel, &gp.Pattern, &gp.Meaning,
		&gp.ExampleJP, &gp.ExampleEN, &gp.Notes, &gp.Source, &gp.CreatedAt); err != nil {
		return nil, err
	}
	return &gp, nil
}

// gradeChallenge receives a user sentence and grammar point ID, sends them to
// Ollama for grading, and returns structured LLM feedback.
func gradeChallenge(database *db.DB, ollamaClient *llm.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ChallengeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if req.UserSentence == "" || req.GrammarPointID == 0 {
			jsonError(w, "grammar_point_id and user_sentence are required", http.StatusBadRequest)
			return
		}

		gp, err := loadGrammarPoint(database, r, req.GrammarPointID)
		if err != nil {
			jsonError(w, "grammar point not found", http.StatusNotFound)
			return
		}

		grade, err := ollamaClient.GradeSentence(r.Context(), *gp, req.UserSentence)
		if err != nil {
			jsonError(w, "llm grading failed: "+err.Error(), http.StatusBadGateway)
			return
		}

		jsonOK(w, grade)
	}
}

// generateConversationPrompt creates a short roleplay setup for a grammar point.
func generateConversationPrompt(database *db.DB, ollamaClient *llm.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ConversationPromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if req.GrammarPointID == 0 {
			jsonError(w, "grammar_point_id is required", http.StatusBadRequest)
			return
		}

		gp, err := loadGrammarPoint(database, r, req.GrammarPointID)
		if err != nil {
			jsonError(w, "grammar point not found", http.StatusNotFound)
			return
		}

		prompt, err := ollamaClient.GenerateConversationPrompt(r.Context(), *gp, req.IncludeFurigana)
		if err != nil {
			jsonError(w, "llm conversation prompt failed: "+err.Error(), http.StatusBadGateway)
			return
		}

		jsonOK(w, prompt)
	}
}

// gradeConversationReply grades a learner's reply in a grammar-focused roleplay.
func gradeConversationReply(database *db.DB, ollamaClient *llm.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ConversationReplyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if req.GrammarPointID == 0 || req.Scenario == "" || req.AssistantMessage == "" || req.UserReply == "" {
			jsonError(w, "grammar_point_id, scenario, assistant_message, and user_reply are required", http.StatusBadRequest)
			return
		}

		gp, err := loadGrammarPoint(database, r, req.GrammarPointID)
		if err != nil {
			jsonError(w, "grammar point not found", http.StatusNotFound)
			return
		}

		grade, err := ollamaClient.GradeConversationReply(r.Context(), *gp, req.Scenario, req.AssistantMessage, req.UserReply, req.IncludeFurigana)
		if err != nil {
			jsonError(w, "llm conversation grading failed: "+err.Error(), http.StatusBadGateway)
			return
		}

		jsonOK(w, grade)
	}
}

// jsonOK writes a 200 JSON response.
func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// jsonError writes an error JSON response.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
