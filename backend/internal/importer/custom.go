package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/user/nihongo-sensei/backend/internal/db"
)

// CustomImporter reads grammar points from a local JSON file.
//
// The file must contain a JSON array of objects.  Required fields: "pattern"
// and "meaning".  All other fields are optional and match the grammar_points
// column names.
type CustomImporter struct{}

func (c *CustomImporter) Source() string { return "custom" }

type customItem struct {
	ExternalID string `json:"external_id"`
	JLPTLevel  string `json:"jlpt_level"`
	Pattern    string `json:"pattern"`
	Meaning    string `json:"meaning"`
	ExampleJP  string `json:"example_jp"`
	ExampleEN  string `json:"example_en"`
	Notes      string `json:"notes"`
	Tags       string `json:"tags"`
	AudioURL   string `json:"audio_url"`
}

func (c *CustomImporter) Import(ctx context.Context, database *db.DB, opts Options) (Result, error) {
	if opts.File == "" {
		return Result{}, fmt.Errorf("--file is required for custom source")
	}
	f, err := os.Open(opts.File)
	if err != nil {
		return Result{}, fmt.Errorf("open %s: %w", opts.File, err)
	}
	defer f.Close()

	var items []customItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return Result{}, fmt.Errorf("decode json: %w", err)
	}

	var result Result
	for i, item := range items {
		if opts.Limit > 0 && i >= opts.Limit {
			break
		}
		if item.Pattern == "" || item.Meaning == "" {
			result.Errors++
			continue
		}

		res, err := database.ExecContext(ctx,
			`INSERT OR IGNORE INTO grammar_points
			    (source, external_id, jlpt_level, pattern, meaning,
			     example_jp, example_en, notes, tags, audio_url)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"custom", item.ExternalID,
			normaliseLevel(item.JLPTLevel), item.Pattern, item.Meaning,
			item.ExampleJP, item.ExampleEN, item.Notes,
			item.Tags, item.AudioURL)
		if err != nil {
			result.Errors++
			continue
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			result.Skipped++
			continue
		}
		result.Imported++
		gpID, _ := res.LastInsertId()
		database.ExecContext(ctx, //nolint:errcheck
			`INSERT OR IGNORE INTO review_cards (grammar_point_id, interval, repetitions, e_factor, due_date)
			 VALUES (?, 1, 0, 2.5, datetime('now'))`, gpID)
	}
	return result, nil
}
