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
	if err := db.seedStarterContent(); err != nil {
		return nil, fmt.Errorf("seed starter content: %w", err)
	}
	return db, nil
}

// migrate runs the embedded schema.sql (idempotent — uses CREATE IF NOT EXISTS).
func (db *DB) migrate() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}
	return nil
}
