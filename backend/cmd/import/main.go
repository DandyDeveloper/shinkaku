// Command import populates the grammar_points table from an external data source.
//
// Usage:
//
//	go run ./cmd/import --source=hanabira
//	go run ./cmd/import --source=tatoeba  --file=sentences.tsv  --limit=5000
//	go run ./cmd/import --source=jmdict   --file=JMdict_e.gz
//	go run ./cmd/import --source=custom   --file=my_grammar.json
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/user/nihongo-sensei/backend/internal/db"
	"github.com/user/nihongo-sensei/backend/internal/importer"
)

func main() {
	source := flag.String("source", "", "data source: hanabira | tatoeba | jmdict | custom")
	limit := flag.Int("limit", 0, "max records to import (0 = unlimited)")
	file := flag.String("file", "", "path to local data file (required for tatoeba, jmdict, custom)")
	dbPath := flag.String("db", envOr("DB_PATH", "./nihongo.db"), "path to SQLite database")
	flag.Parse()

	if *source == "" {
		fmt.Fprintln(os.Stderr,
			"usage: import --source=<hanabira|tatoeba|jmdict|custom> [--limit=N] [--file=PATH] [--db=PATH]")
		os.Exit(1)
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db %s: %v", *dbPath, err)
	}
	defer database.Close()

	opts := importer.Options{Limit: *limit, File: *file}

	var imp importer.Importer
	switch *source {
	case "hanabira":
		imp = &importer.HanabiraImporter{}
	case "tatoeba":
		imp = &importer.TatoebaImporter{}
	case "jmdict":
		imp = &importer.JMdictImporter{}
	case "custom":
		imp = &importer.CustomImporter{}
	default:
		log.Fatalf("unknown source %q; valid values: hanabira, tatoeba, jmdict, custom", *source)
	}

	result, err := importer.RunImport(context.Background(), imp, database, opts)
	if err != nil {
		log.Fatalf("import failed: %v", err)
	}
	fmt.Printf("source=%s  imported=%d  skipped=%d  errors=%d\n",
		*source, result.Imported, result.Skipped, result.Errors)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
