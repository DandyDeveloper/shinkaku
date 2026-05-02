// Package importer defines the shared interface for all data-source importers
// and the RunImport wrapper that logs each run to the import_log table.
package importer

import (
	"context"
	"fmt"
	"log"

	"github.com/user/nihongo-sensei/backend/internal/db"
)

// Options are passed to every Importer.Import call.
type Options struct {
	Limit int    // max records to process (0 = unlimited)
	File  string // path to local data file (required by file-based sources)
}

// Result summarises a completed import run.
type Result struct {
	Imported int
	Skipped  int
	Errors   int
}

// Importer is implemented by every data source.
type Importer interface {
	// Source returns the canonical source name stored in grammar_points.source
	// and import_log.source.
	Source() string
	// Import performs the import and returns row counts.  Implementations must
	// not write to import_log themselves — RunImport owns that.
	Import(ctx context.Context, database *db.DB, opts Options) (Result, error)
}

// RunImport executes imp, wraps the run with import_log rows, and returns the
// aggregated result.  A failed import still records counts for the rows that
// did succeed before the error.
func RunImport(ctx context.Context, imp Importer, database *db.DB, opts Options) (Result, error) {
	res, err := database.ExecContext(ctx,
		`INSERT INTO import_log (source, status) VALUES (?, 'running')`, imp.Source())
	if err != nil {
		return Result{}, fmt.Errorf("log start: %w", err)
	}
	logID, _ := res.LastInsertId()

	log.Printf("[import] starting source=%s limit=%d file=%q", imp.Source(), opts.Limit, opts.File)
	result, importErr := imp.Import(ctx, database, opts)
	log.Printf("[import] done source=%s imported=%d skipped=%d errors=%d err=%v",
		imp.Source(), result.Imported, result.Skipped, result.Errors, importErr)

	status, msg := "done", ""
	if importErr != nil {
		status = "failed"
		msg = importErr.Error()
	}
	database.ExecContext(ctx, //nolint:errcheck
		`UPDATE import_log
		    SET finished_at=datetime('now'), imported=?, skipped=?, errors=?, status=?, message=?
		  WHERE id=?`,
		result.Imported, result.Skipped, result.Errors, status, msg, logID)

	return result, importErr
}
