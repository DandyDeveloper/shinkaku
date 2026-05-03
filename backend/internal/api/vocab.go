package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/user/shinkaku/backend/internal/db"
	"github.com/user/shinkaku/backend/internal/models"
)

func listVocab(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jlpt := r.URL.Query().Get("jlpt")

		var (
			rows dbRows
			err  error
		)

		if jlpt != "" {
			rows, err = database.QueryContext(r.Context(),
				`SELECT id, jlpt_level, word, reading, meaning, part_of_speech,
				        example_jp, example_en, notes, source, created_at
				 FROM vocab_words WHERE jlpt_level = ? ORDER BY id ASC`, jlpt)
		} else {
			rows, err = database.QueryContext(r.Context(),
				`SELECT id, jlpt_level, word, reading, meaning, part_of_speech,
				        example_jp, example_en, notes, source, created_at
				 FROM vocab_words ORDER BY id ASC`)
		}
		if err != nil {
			jsonError(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var words []models.VocabWord
		for rows.Next() {
			var v models.VocabWord
			if err := rows.Scan(&v.ID, &v.JLPTLevel, &v.Word, &v.Reading, &v.Meaning,
				&v.PartOfSpeech, &v.ExampleJP, &v.ExampleEN, &v.Notes, &v.Source, &v.CreatedAt); err != nil {
				jsonError(w, "scan failed", http.StatusInternalServerError)
				return
			}
			words = append(words, v)
		}
		if words == nil {
			words = []models.VocabWord{}
		}
		jsonOK(w, words)
	}
}

func getVocab(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, "invalid id", http.StatusBadRequest)
			return
		}

		var v models.VocabWord
		row := database.QueryRowContext(r.Context(),
			`SELECT id, jlpt_level, word, reading, meaning, part_of_speech,
			        example_jp, example_en, notes, source, created_at
			 FROM vocab_words WHERE id = ?`, id)
		if err := row.Scan(&v.ID, &v.JLPTLevel, &v.Word, &v.Reading, &v.Meaning,
			&v.PartOfSpeech, &v.ExampleJP, &v.ExampleEN, &v.Notes, &v.Source, &v.CreatedAt); err != nil {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonOK(w, v)
	}
}

func createVocab(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var v models.VocabWord
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if v.Word == "" || v.Meaning == "" {
			jsonError(w, "word and meaning are required", http.StatusBadRequest)
			return
		}

		res, err := database.ExecContext(r.Context(),
			`INSERT INTO vocab_words (jlpt_level, word, reading, meaning, part_of_speech,
			                          example_jp, example_en, notes, source)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			v.JLPTLevel, v.Word, v.Reading, v.Meaning, v.PartOfSpeech,
			v.ExampleJP, v.ExampleEN, v.Notes, v.Source)
		if err != nil {
			jsonError(w, "insert failed", http.StatusInternalServerError)
			return
		}

		vocabID, _ := res.LastInsertId()
		v.ID = vocabID

		// Auto-create an SRS review card for this vocab word.
		_, err = database.ExecContext(r.Context(),
			`INSERT INTO vocab_review_cards (vocab_word_id) VALUES (?)`, vocabID)
		if err != nil {
			jsonError(w, "card creation failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		jsonOK(w, v)
	}
}
