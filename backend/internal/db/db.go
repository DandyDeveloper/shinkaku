package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps a *sql.DB with application-level helpers.
type DB struct {
	*sql.DB
}

// Open opens (or creates) the SQLite database at path and runs schema migrations.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode and foreign keys at the connection level.
	if _, err := sqlDB.Exec(`PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("set pragmas: %w", err)
	}

	db := &DB{sqlDB}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// migrate runs the embedded schema.sql then applies any column additions needed
// for databases created before those columns were introduced.
func (db *DB) migrate() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}
	// Column back-fills: safe to run on every start; no-ops if column exists.
	for _, m := range []struct{ col, def string }{
		{"external_id", "TEXT NOT NULL DEFAULT ''"},
		{"tags", "TEXT NOT NULL DEFAULT ''"},
		{"audio_url", "TEXT NOT NULL DEFAULT ''"},
	} {
		if err := db.addColIfMissing("grammar_points", m.col, m.def); err != nil {
			return fmt.Errorf("migrate grammar_points.%s: %w", m.col, err)
		}
	}
	return nil
}

// addColIfMissing adds a column to table only when it does not already exist.
func (db *DB) addColIfMissing(table, col, def string) error {
	var n int
	// pragma_table_info is a SQLite virtual table available since 3.16 (2017).
	err := db.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name=?", table), col,
	).Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, def))
	return err
}
