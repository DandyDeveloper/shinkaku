package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/user/nihongo-sensei/backend/internal/db"
)

type importLogEntry struct {
	ID         int64      `json:"id"`
	Source     string     `json:"source"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Imported   int        `json:"imported"`
	Skipped    int        `json:"skipped"`
	Errors     int        `json:"errors"`
	Status     string     `json:"status"`
	Message    string     `json:"message,omitempty"`
}

// listImportLog returns the 100 most recent import_log rows newest-first.
func listImportLog(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := database.QueryContext(r.Context(),
			`SELECT id, source, started_at, finished_at, imported, skipped, errors, status, message
			   FROM import_log
			  ORDER BY id DESC
			  LIMIT 100`)
		if err != nil {
			jsonError(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var entries []importLogEntry
		for rows.Next() {
			var e importLogEntry
			var finishedAt sql.NullTime
			if err := rows.Scan(&e.ID, &e.Source, &e.StartedAt, &finishedAt,
				&e.Imported, &e.Skipped, &e.Errors, &e.Status, &e.Message); err != nil {
				jsonError(w, "scan failed", http.StatusInternalServerError)
				return
			}
			if finishedAt.Valid {
				e.FinishedAt = &finishedAt.Time
			}
			entries = append(entries, e)
		}
		if entries == nil {
			entries = []importLogEntry{}
		}
		jsonOK(w, entries)
	}
}
