package db

import (
	"path/filepath"
	"testing"
)

func countStarterPointsUpTo(level string) int {
	normalized, ok := normalizeJLPTLevel(level)
	if !ok {
		return 0
	}
	total := 0
	for _, point := range starterGrammarPoints {
		if point.jlptLevel == normalized {
			total++
		}
	}
	return total
}

func TestOpenDoesNotAutoSeedStarterGrammarOnEmptyDatabase(t *testing.T) {
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
	if grammarCount != 0 {
		t.Fatalf("grammar count = %d, want 0", grammarCount)
	}
}

func TestSeedStarterContentForJLPTSeedsOnlySelectedLevel(t *testing.T) {
	t.Parallel()

	databasePath := filepath.Join(t.TempDir(), "seed-by-level.db")
	database, err := Open(databasePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.Close()

	inserted, err := database.SeedStarterContentForJLPT("N3")
	if err != nil {
		t.Fatalf("SeedStarterContentForJLPT() error = %v", err)
	}

	expected := countStarterPointsUpTo("N3")
	if inserted != expected {
		t.Fatalf("inserted = %d, want %d", inserted, expected)
	}

	var grammarCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM grammar_points`).Scan(&grammarCount); err != nil {
		t.Fatalf("count grammar_points: %v", err)
	}
	if grammarCount != expected {
		t.Fatalf("grammar count = %d, want %d", grammarCount, expected)
	}

	var reviewCardCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM review_cards`).Scan(&reviewCardCount); err != nil {
		t.Fatalf("count review_cards: %v", err)
	}
	if reviewCardCount != expected {
		t.Fatalf("review card count = %d, want %d", reviewCardCount, expected)
	}
}

func TestSeedStarterContentForJLPTFailsWhenTableIsNotEmpty(t *testing.T) {
	t.Parallel()

	databasePath := filepath.Join(t.TempDir(), "seeded-once.db")
	database, err := Open(databasePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.Close()

	if _, err := database.SeedStarterContentForJLPT("N5"); err != nil {
		t.Fatalf("first seed error = %v", err)
	}

	_, err = database.SeedStarterContentForJLPT("N4")
	if err == nil {
		t.Fatalf("second seed expected error, got nil")
	}
}