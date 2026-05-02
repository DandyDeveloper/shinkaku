package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/shinkaku/backend/internal/db"
	"github.com/user/shinkaku/backend/internal/models"
	"github.com/user/shinkaku/backend/internal/srs"
)

// getReviewQueue returns all cards due today (due_date <= now).
func getReviewQueue(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := database.QueryContext(r.Context(), `
			SELECT rc.id, rc.grammar_point_id, rc.interval, rc.repetitions, rc.e_factor,
			       rc.due_date, rc.last_reviewed,
			       gp.jlpt_level, gp.pattern, gp.meaning, gp.example_jp, gp.example_en, gp.notes
			FROM review_cards rc
			JOIN grammar_points gp ON gp.id = rc.grammar_point_id
			WHERE rc.due_date <= datetime('now')
			ORDER BY rc.due_date ASC
		`)
		if err != nil {
			jsonError(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type QueueItem struct {
			models.ReviewCard
			GrammarPoint models.GrammarPoint `json:"grammar_point"`
		}

		var items []QueueItem
		for rows.Next() {
			var item QueueItem
			var lastReviewed *time.Time
			if err := rows.Scan(
				&item.ReviewCard.ID,
				&item.ReviewCard.GrammarPointID,
				&item.ReviewCard.Interval,
				&item.ReviewCard.Repetitions,
				&item.ReviewCard.EFactor,
				&item.ReviewCard.DueDate,
				&lastReviewed,
				&item.GrammarPoint.JLPTLevel,
				&item.GrammarPoint.Pattern,
				&item.GrammarPoint.Meaning,
				&item.GrammarPoint.ExampleJP,
				&item.GrammarPoint.ExampleEN,
				&item.GrammarPoint.Notes,
			); err != nil {
				jsonError(w, "scan failed", http.StatusInternalServerError)
				return
			}
			item.ReviewCard.LastReviewed = lastReviewed
			item.GrammarPoint.ID = item.ReviewCard.GrammarPointID
			items = append(items, item)
		}
		if items == nil {
			items = []QueueItem{}
		}
		jsonOK(w, map[string]any{
			"cards": items,
			"total": len(items),
		})
	}
}

// submitGrade receives a grade (0–5) for a card, runs SM-2, and persists the result.
func submitGrade(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, "invalid card id", http.StatusBadRequest)
			return
		}

		var req models.GradeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if req.Grade < 0 || req.Grade > 5 {
			jsonError(w, "grade must be 0–5", http.StatusBadRequest)
			return
		}

		// Load current card state.
		var card models.ReviewCard
		row := database.QueryRowContext(r.Context(),
			`SELECT id, grammar_point_id, interval, repetitions, e_factor, due_date, last_reviewed
			 FROM review_cards WHERE id = ?`, cardID)
		if err := row.Scan(&card.ID, &card.GrammarPointID, &card.Interval,
			&card.Repetitions, &card.EFactor, &card.DueDate, &card.LastReviewed); err != nil {
			jsonError(w, "card not found", http.StatusNotFound)
			return
		}

		// Run SM-2.
		state := srs.CardState{
			Interval:    card.Interval,
			Repetitions: card.Repetitions,
			EFactor:     card.EFactor,
		}
		now := time.Now().UTC()
		newState, nextDue := srs.Review(state, req.Grade, now)

		// Persist updated card.
		_, err = database.ExecContext(r.Context(),
			`UPDATE review_cards SET interval=?, repetitions=?, e_factor=?, due_date=?, last_reviewed=?
			 WHERE id=?`,
			newState.Interval, newState.Repetitions, newState.EFactor,
			nextDue.Format(time.RFC3339), now.Format(time.RFC3339), cardID)
		if err != nil {
			jsonError(w, "update failed", http.StatusInternalServerError)
			return
		}

		// Record history.
		_, err = database.ExecContext(r.Context(),
			`INSERT INTO review_history (card_id, grade, reviewed_at) VALUES (?, ?, ?)`,
			cardID, req.Grade, now.Format(time.RFC3339))
		if err != nil {
			jsonError(w, "history insert failed", http.StatusInternalServerError)
			return
		}

		jsonOK(w, map[string]any{
			"card_id":      cardID,
			"new_interval": newState.Interval,
			"repetitions":  newState.Repetitions,
			"e_factor":     newState.EFactor,
			"next_due":     nextDue,
		})
	}
}
