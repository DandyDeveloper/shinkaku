package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/user/nihongo-sensei/backend/internal/db"
	"github.com/user/nihongo-sensei/backend/internal/models"
)

func listGrammar(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jlpt := r.URL.Query().Get("jlpt")

		var (
			rows dbRows
			err  error
		)

		if jlpt != "" {
			rows, err = database.QueryContext(r.Context(),
				`SELECT id, jlpt_level, pattern, meaning, example_jp, example_en, notes, source, created_at
				 FROM grammar_points WHERE jlpt_level = ? ORDER BY id ASC`, jlpt)
		} else {
			rows, err = database.QueryContext(r.Context(),
				`SELECT id, jlpt_level, pattern, meaning, example_jp, example_en, notes, source, created_at
				 FROM grammar_points ORDER BY id ASC`)
		}
		if err != nil {
			jsonError(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var points []models.GrammarPoint
		for rows.Next() {
			var gp models.GrammarPoint
			if err := rows.Scan(&gp.ID, &gp.JLPTLevel, &gp.Pattern, &gp.Meaning,
				&gp.ExampleJP, &gp.ExampleEN, &gp.Notes, &gp.Source, &gp.CreatedAt); err != nil {
				jsonError(w, "scan failed", http.StatusInternalServerError)
				return
			}
			points = append(points, gp)
		}
		if points == nil {
			points = []models.GrammarPoint{}
		}
		jsonOK(w, points)
	}
}

func getGrammar(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, "invalid id", http.StatusBadRequest)
			return
		}

		var gp models.GrammarPoint
		row := database.QueryRowContext(r.Context(),
			`SELECT id, jlpt_level, pattern, meaning, example_jp, example_en, notes, source, created_at
			 FROM grammar_points WHERE id = ?`, id)
		if err := row.Scan(&gp.ID, &gp.JLPTLevel, &gp.Pattern, &gp.Meaning,
			&gp.ExampleJP, &gp.ExampleEN, &gp.Notes, &gp.Source, &gp.CreatedAt); err != nil {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonOK(w, gp)
	}
}

func createGrammar(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var gp models.GrammarPoint
		if err := json.NewDecoder(r.Body).Decode(&gp); err != nil {
			jsonError(w, "invalid body", http.StatusBadRequest)
			return
		}
		if gp.Pattern == "" || gp.Meaning == "" {
			jsonError(w, "pattern and meaning are required", http.StatusBadRequest)
			return
		}

		res, err := database.ExecContext(r.Context(),
			`INSERT INTO grammar_points (jlpt_level, pattern, meaning, example_jp, example_en, notes, source)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			gp.JLPTLevel, gp.Pattern, gp.Meaning, gp.ExampleJP, gp.ExampleEN, gp.Notes, gp.Source)
		if err != nil {
			jsonError(w, "insert failed", http.StatusInternalServerError)
			return
		}

		gpID, _ := res.LastInsertId()
		gp.ID = gpID

		// Auto-create an SRS review card for this grammar point.
		_, err = database.ExecContext(r.Context(),
			`INSERT INTO review_cards (grammar_point_id, interval, repetitions, e_factor, due_date)
			 VALUES (?, 1, 0, 2.5, datetime('now'))`, gpID)
		if err != nil {
			jsonError(w, "create card failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		jsonOK(w, gp)
	}
}

// dbRows is an alias to avoid import cycles when the db package is used directly.
type dbRows = interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
}
