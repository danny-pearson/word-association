package db

import (
	"context"
	"embed"
	"slices"
	"testing"

	"github.com/danny-pearson/word-association/internal/puzzle"
	"github.com/pressly/goose/v3"
)

// Migrations are embedded for tests only. The application applies them with the
// goose CLI, so this exists purely to give each test a schema to run against.
//
//go:embed migrations/*.sql
var migrationFS embed.FS

// testStore returns a Store backed by its own migrated database, opened exactly
// as the application opens it so tests run against the same pragmas.
func testStore(t *testing.T) *Store {
	t.Helper()

	conn, err := Open(t.TempDir() + "/test.db")

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { conn.Close() })

	goose.SetBaseFS(migrationFS)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}

	if err := goose.Up(conn, "migrations"); err != nil {
		t.Fatal(err)
	}

	return NewStore(conn)
}

func TestSavePuzzle(t *testing.T) {
	store := testStore(t)

	puzzleID, err := store.SavePuzzle(context.Background(), puzzle.GeneratedPuzzle{
		Answer: "bell",
		Clues:  []string{"church", "brass", "school", "ring", "tower"},
	}, "test-model")

	if err != nil {
		t.Fatalf("SavePuzzle() returned error: %v", err)
	}

	var answer, normalizedAnswer, sourceModel string

	err = store.db.QueryRow(`
		SELECT answer, answer_normalized, source_model FROM puzzles
			WHERE id = ?
	`, puzzleID).Scan(&answer, &normalizedAnswer, &sourceModel)

	if err != nil {
		t.Fatalf("querying puzzle: %v", err)
	}

	if answer != "bell" {
		t.Errorf("answer = %q, want %q", answer, "bell")
	}

	if normalizedAnswer != "bell" {
		t.Errorf("answer_normalized = %q, want %q", normalizedAnswer, "bell")
	}

	if sourceModel != "test-model" {
		t.Errorf("source_model = %q, want %q", sourceModel, "test-model")
	}

	clues := readClues(t, store, puzzleID)
	want := []string{"church", "brass", "school", "ring", "tower"}

	if !slices.Equal(clues, want) {
		t.Errorf("clues = %v, want %v", clues, want)
	}
}

// A clue used by more than one puzzle is stored once and linked twice, which is
// what lets "bell" belong to church, brass instrument and school alike.
func TestSavePuzzleReusesClues(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	firstID, err := store.SavePuzzle(ctx, puzzle.GeneratedPuzzle{
		Answer: "church",
		Clues:  []string{"bell", "pew", "altar", "spire", "hymn"},
	}, "test-model")

	if err != nil {
		t.Fatalf("SavePuzzle() returned error: %v", err)
	}

	secondID, err := store.SavePuzzle(ctx, puzzle.GeneratedPuzzle{
		Answer: "school",
		Clues:  []string{"bell", "desk", "pupil", "term", "playground"},
	}, "test-model")

	if err != nil {
		t.Fatalf("SavePuzzle() returned error: %v", err)
	}

	var clueCount int

	if err := store.db.QueryRow(`
		SELECT COUNT(*) FROM clues
	`).Scan(&clueCount); err != nil {
		t.Fatalf("counting clues: %v", err)
	}

	// Nine distinct clues across two puzzles, because "bell" appears in both.
	if clueCount != 9 {
		t.Errorf("clue count = %d, want 9", clueCount)
	}

	var bellLinks int

	err = store.db.QueryRow(`
		SELECT COUNT(*) FROM puzzle_clues pc
			JOIN clues c ON c.id = pc.clue_id
			WHERE c.text_normalized = 'bell'
	`).Scan(&bellLinks)

	if err != nil {
		t.Fatalf("counting links: %v", err)
	}

	if bellLinks != 2 {
		t.Errorf("links to %q = %d, want 2", "bell", bellLinks)
	}

	if got := readClues(t, store, firstID); got[0] != "bell" {
		t.Errorf("first puzzle clue 1 = %q, want %q", got[0], "bell")
	}

	if got := readClues(t, store, secondID); got[0] != "bell" {
		t.Errorf("second puzzle clue 1 = %q, want %q", got[0], "bell")
	}
}

// A puzzle whose answer is already stored is rejected by the unique constraint,
// and the transaction must leave nothing behind.
func TestSavePuzzleRollsBackOnDuplicateAnswer(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	if _, err := store.SavePuzzle(ctx, puzzle.GeneratedPuzzle{
		Answer: "bell",
		Clues:  []string{"church", "brass", "school", "ring", "tower"},
	}, "test-model"); err != nil {
		t.Fatalf("SavePuzzle() returned error: %v", err)
	}

	// Same answer, different clues: the answer collides once normalized.
	_, err := store.SavePuzzle(ctx, puzzle.GeneratedPuzzle{
		Answer: "Bell",
		Clues:  []string{"boxing", "cow", "liberty", "door", "dumb"},
	}, "test-model")

	if err == nil {
		t.Fatal("SavePuzzle() = nil, want a unique constraint error")
	}

	var puzzleCount int

	if err := store.db.QueryRow(`
		SELECT COUNT(*) FROM puzzles
	`).Scan(&puzzleCount); err != nil {
		t.Fatalf("counting puzzles: %v", err)
	}

	if puzzleCount != 1 {
		t.Errorf("puzzle count = %d, want 1", puzzleCount)
	}

	// The second puzzle's clues were upserted before the failure, so they must
	// have been rolled back along with it.
	var clueCount int

	if err := store.db.QueryRow(`
		SELECT COUNT(*) FROM clues
	`).Scan(&clueCount); err != nil {
		t.Fatalf("counting clues: %v", err)
	}

	if clueCount != puzzle.ClueCount {
		t.Errorf("clue count = %d, want %d", clueCount, puzzle.ClueCount)
	}
}

// readClues returns a puzzle's clue texts ordered by their reveal position.
func readClues(t *testing.T, store *Store, puzzleID int64) []string {
	t.Helper()

	rows, err := store.db.Query(`
		SELECT c.text FROM puzzle_clues pc
			JOIN clues c ON c.id = pc.clue_id
			WHERE pc.puzzle_id = ?
			ORDER BY pc.position
	`, puzzleID)

	if err != nil {
		t.Fatalf("querying clues: %v", err)
	}

	defer rows.Close()

	var clues []string

	for rows.Next() {
		var text string

		if err := rows.Scan(&text); err != nil {
			t.Fatalf("scanning clue: %v", err)
		}

		clues = append(clues, text)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterating clues: %v", err)
	}

	return clues
}
