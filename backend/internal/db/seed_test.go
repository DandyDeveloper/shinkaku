package db

import (
	"path/filepath"
	"testing"
)

func TestOpenSeedsStarterGrammarOnEmptyDatabase(t *testing.T) {
	t.Parallel()

	databasePath := filepath.Join(t.TempDir(), "seeded.db")
	database, err := Open(databasePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.Close()

	var grammarCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM grammar_points`).Scan(&grammarCount); err != nil {
		t.Fatalf("count grammar_points: %v", err)
	}
	if grammarCount != len(starterGrammarPoints) {
		t.Fatalf("grammar count = %d, want %d", grammarCount, len(starterGrammarPoints))
	}

	var reviewCardCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM review_cards`).Scan(&reviewCardCount); err != nil {
		t.Fatalf("count review_cards: %v", err)
	}
	if reviewCardCount != len(starterGrammarPoints) {
		t.Fatalf("review card count = %d, want %d", reviewCardCount, len(starterGrammarPoints))
	}
}

func TestOpenDoesNotDuplicateStarterGrammar(t *testing.T) {
	t.Parallel()

	databasePath := filepath.Join(t.TempDir(), "seeded-once.db")
	firstOpen, err := Open(databasePath)
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	firstOpen.Close()

	secondOpen, err := Open(databasePath)
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	defer secondOpen.Close()

	var grammarCount int
	if err := secondOpen.QueryRow(`SELECT COUNT(*) FROM grammar_points`).Scan(&grammarCount); err != nil {
		t.Fatalf("count grammar_points: %v", err)
	}
	if grammarCount != len(starterGrammarPoints) {
		t.Fatalf("grammar count after reopen = %d, want %d", grammarCount, len(starterGrammarPoints))
	}
}