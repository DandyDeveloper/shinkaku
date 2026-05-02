package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/user/nihongo-sensei/backend/internal/db"
)

// hanabiraDefaultURL points to the community-maintained Hanabira grammar JSON.
// Override via HanabiraImporter.URL for testing or a local mirror.
const hanabiraDefaultURL = "https://raw.githubusercontent.com/aiko-chan-ai/hanabira.org/main/grammar/grammar_en.json"

// HanabiraImporter fetches grammar points from Hanabira.org's JSON export.
type HanabiraImporter struct {
	URL string // optional override; defaults to hanabiraDefaultURL
}

func (h *HanabiraImporter) Source() string { return "hanabira" }

type hanabiraItem struct {
	GrammarPoint string `json:"grammar_point"`
	Meaning      string `json:"meaning"`
	Level        string `json:"level"`
	ExampleJP    string `json:"example"`
	ExampleEN    string `json:"example_en"`
	Caution      string `json:"caution"`
}

func (h *HanabiraImporter) Import(ctx context.Context, database *db.DB, opts Options) (Result, error) {
	url := h.URL
	if url == "" {
		url = hanabiraDefaultURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, fmt.Errorf("build request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("unexpected HTTP %d from %s", resp.StatusCode, url)
	}

	var items []hanabiraItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return Result{}, fmt.Errorf("decode json: %w", err)
	}

	var result Result
	for i, item := range items {
		if opts.Limit > 0 && i >= opts.Limit {
			break
		}
		res, err := database.ExecContext(ctx,
			`INSERT OR IGNORE INTO grammar_points
			    (source, external_id, jlpt_level, pattern, meaning, example_jp, example_en, notes)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"hanabira", item.GrammarPoint,
			normaliseLevel(item.Level), item.GrammarPoint,
			item.Meaning, item.ExampleJP, item.ExampleEN, item.Caution)
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

// normaliseLevel converts any capitalisation variant (n3, N3) to canonical form.
func normaliseLevel(level string) string {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "N1":
		return "N1"
	case "N2":
		return "N2"
	case "N3":
		return "N3"
	case "N4":
		return "N4"
	case "N5":
		return "N5"
	default:
		return ""
	}
}
