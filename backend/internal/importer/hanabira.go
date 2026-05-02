// Package importer provides importers for external grammar data sources.
package importer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/user/shinkaku/backend/internal/db"
	"github.com/user/shinkaku/backend/internal/models"
)

// HanabiraGrammarItem maps the JSON structure from Hanabira.org grammar exports.
// TODO: Update field names to match the actual Hanabira JSON schema once confirmed.
type HanabiraGrammarItem struct {
	Grammar    string `json:"grammar_point"`
	Meaning    string `json:"meaning"`
	Level      string `json:"level"`
	ExampleJP  string `json:"example"`
	ExampleEN  string `json:"example_en"`
	Caution    string `json:"caution"`
}

// ImportHanabiraFile reads a Hanabira-format JSON file and inserts all grammar
// points into the database. Existing patterns are skipped (upsert by pattern).
//
// TODO: Verify the actual Hanabira JSON export format and adjust field mappings.
// TODO: Add deduplication by (pattern, jlpt_level) pair.
// TODO: Support batch inserts for large files.
func ImportHanabiraFile(path string, database *db.DB) (imported int, skipped int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var items []HanabiraGrammarItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return 0, 0, fmt.Errorf("decode json: %w", err)
	}

	for _, item := range items {
		gp := models.GrammarPoint{
			Pattern:   item.Grammar,
			Meaning:   item.Meaning,
			JLPTLevel: normaliseLevel(item.Level),
			ExampleJP: item.ExampleJP,
			ExampleEN: item.ExampleEN,
			Notes:     item.Caution,
			Source:    "hanabira",
		}

		res, execErr := database.Exec(
			`INSERT OR IGNORE INTO grammar_points (jlpt_level, pattern, meaning, example_jp, example_en, notes, source)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			gp.JLPTLevel, gp.Pattern, gp.Meaning, gp.ExampleJP, gp.ExampleEN, gp.Notes, gp.Source,
		)
		if execErr != nil {
			return imported, skipped, fmt.Errorf("insert %q: %w", item.Grammar, execErr)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			skipped++
		} else {
			imported++
			// Auto-create SRS card.
			gpID, _ := res.LastInsertId()
			database.Exec(
				`INSERT OR IGNORE INTO review_cards (grammar_point_id, interval, repetitions, e_factor, due_date)
				 VALUES (?, 1, 0, 2.5, datetime('now'))`, gpID,
			)
		}
	}
	return imported, skipped, nil
}

// normaliseLevel converts Hanabira level strings to our canonical format (N1–N5).
func normaliseLevel(level string) string {
	switch level {
	case "N1", "N2", "N3", "N4", "N5":
		return level
	case "n1":
		return "N1"
	case "n2":
		return "N2"
	case "n3":
		return "N3"
	case "n4":
		return "N4"
	case "n5":
		return "N5"
	default:
		return ""
	}
}
