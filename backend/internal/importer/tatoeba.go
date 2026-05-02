package importer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/user/nihongo-sensei/backend/internal/db"
)

// TatoebaImporter streams sentences from a Tatoeba sentences.tsv file.
//
// Download the file from https://tatoeba.org/en/downloads (Sentences in language).
// The TSV columns are: sentence_id \t lang \t text
// Only Japanese (jpn) rows are imported; all others are skipped.
type TatoebaImporter struct{}

func (t *TatoebaImporter) Source() string { return "tatoeba" }

func (t *TatoebaImporter) Import(ctx context.Context, database *db.DB, opts Options) (Result, error) {
	if opts.File == "" {
		return Result{}, fmt.Errorf("--file is required for tatoeba source (download sentences.tsv from tatoeba.org/en/downloads)")
	}
	f, err := os.Open(opts.File)
	if err != nil {
		return Result{}, fmt.Errorf("open %s: %w", opts.File, err)
	}
	defer f.Close()

	var result Result
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if opts.Limit > 0 && result.Imported+result.Skipped >= opts.Limit {
			break
		}

		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		sentenceID, lang, text := parts[0], parts[1], parts[2]
		if lang != "jpn" {
			continue
		}

		res, err := database.ExecContext(ctx,
			`INSERT OR IGNORE INTO grammar_points
			    (source, external_id, jlpt_level, pattern, meaning, example_jp, tags)
			 VALUES (?, ?, '', ?, '', ?, 'sentence')`,
			"tatoeba", sentenceID, text, text)
		if err != nil {
			result.Errors++
			continue
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			result.Skipped++
		} else {
			result.Imported++
		}
	}
	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("scan: %w", err)
	}
	return result, nil
}
