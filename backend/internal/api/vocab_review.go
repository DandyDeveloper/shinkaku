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

// getVocabReviewQueue returns all vocab cards due today.
func getVocabReviewQueue(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := database.QueryContext(r.Context(), `
			SELECT vc.id, vc.vocab_word_id, vc.interval, vc.repetitions, vc.e_factor,
			       vc.due_date, vc.last_reviewed,
			       vw.jlpt_level, vw.word, vw.reading, vw.meaning, vw.part_of_speech,
			       vw.example_jp, vw.example_en, vw.notes
			FROM vocab_review_cards vc
			JOIN vocab_words vw ON vw.id = vc.vocab_word_id
			WHERE vc.due_date <= datetime('now')
			ORDER BY vc.due_date ASC
		`)
		if err != nil {
			jsonError(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type QueueItem struct {
			models.VocabReviewCard
			VocabWord models.VocabWord `json:"vocab_word"`
		}

		var items []QueueItem
		for rows.Next() {
			var item QueueItem
			var lastReviewed *time.Time
			if err := rows.Scan(
				&item.VocabReviewCard.ID,
				&item.VocabReviewCard.VocabWordID,
				&item.VocabReviewCard.Interval,
				&item.VocabReviewCard.Repetitions,
				&item.VocabReviewCard.EFactor,
				&item.VocabReviewCard.DueDate,
				&lastReviewed,
				&item.VocabWord.JLPTLevel,
				&item.VocabWord.Word,
				&item.VocabWord.Reading,
				&item.VocabWord.Meaning,
				&item.VocabWord.PartOfSpeech,
				&item.VocabWord.ExampleJP,
				&item.VocabWord.ExampleEN,
				&item.VocabWord.Notes,
			); err != nil {
				jsonError(w, "scan failed", http.StatusInternalServerError)
				return
			}
			item.VocabReviewCard.LastReviewed = lastReviewed
			item.VocabWord.ID = item.VocabReviewCard.VocabWordID
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

// submitVocabGrade receives a grade (0–5) for a vocab card, runs SM-2, and persists.
func submitVocabGrade(database *db.DB) http.HandlerFunc {
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

		var card models.VocabReviewCard
		row := database.QueryRowContext(r.Context(),
			`SELECT id, vocab_word_id, interval, repetitions, e_factor, due_date, last_reviewed
			 FROM vocab_review_cards WHERE id = ?`, cardID)
		if err := row.Scan(&card.ID, &card.VocabWordID, &card.Interval,
			&card.Repetitions, &card.EFactor, &card.DueDate, &card.LastReviewed); err != nil {
			jsonError(w, "card not found", http.StatusNotFound)
			return
		}

		state := srs.CardState{
			Interval:    card.Interval,
			Repetitions: card.Repetitions,
			EFactor:     card.EFactor,
		}
		now := time.Now().UTC()
		newState, nextDue := srs.Review(state, req.Grade, now)

		_, err = database.ExecContext(r.Context(),
			`UPDATE vocab_review_cards SET interval=?, repetitions=?, e_factor=?, due_date=?, last_reviewed=?
			 WHERE id=?`,
			newState.Interval, newState.Repetitions, newState.EFactor,
			nextDue.Format(time.RFC3339), now.Format(time.RFC3339), cardID)
		if err != nil {
			jsonError(w, "update failed", http.StatusInternalServerError)
			return
		}

		_, err = database.ExecContext(r.Context(),
			`INSERT INTO vocab_review_history (card_id, grade, reviewed_at) VALUES (?, ?, ?)`,
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
