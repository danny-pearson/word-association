package db

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/danny-pearson/word-association/internal/puzzle"
)

// savePuzzle stores a puzzle with generated clues and returns its id, so tests
// that care about games rather than puzzles can set up a library in one line.
func savePuzzle(t *testing.T, store *Store, answer string) int64 {
	t.Helper()

	id, err := store.SavePuzzle(context.Background(), puzzle.GeneratedPuzzle{
		Answer: answer,
		Clues: []string{
			answer + " one",
			answer + " two",
			answer + " three",
			answer + " four",
			answer + " five",
		},
	}, "test-model")

	if err != nil {
		t.Fatalf("SavePuzzle(%q) returned error: %v", answer, err)
	}

	return id
}

// completeGame plays a game to completion so that its puzzle counts as played.
func completeGame(t *testing.T, store *Store, playerID string, puzzleID int64) {
	t.Helper()

	ctx := context.Background()

	gameID, err := store.CreateGame(ctx, playerID, puzzleID)

	if err != nil {
		t.Fatalf("CreateGame() returned error: %v", err)
	}

	if _, err := store.CompleteGame(ctx, gameID, "2026-01-01T00:00:00Z", true); err != nil {
		t.Fatalf("CompleteGame() returned error: %v", err)
	}
}

// A puzzle one player has finished stays available to everyone else, because
// the library is consumed per player rather than globally.
func TestGetRandomUnplayedPuzzleExcludesOnlyThatPlayersGames(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	played := savePuzzle(t, store, "bell")
	completeGame(t, store, "player-a", played)

	if _, err := store.GetRandomUnplayedPuzzle(ctx, "player-a"); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetRandomUnplayedPuzzle() = %v, want ErrNotFound", err)
	}

	got, err := store.GetRandomUnplayedPuzzle(ctx, "player-b")

	if err != nil {
		t.Fatalf("GetRandomUnplayedPuzzle() returned error: %v", err)
	}

	if got != played {
		t.Errorf("puzzle for second player = %d, want %d", got, played)
	}
}

// An abandoned game does not consume its puzzle, so a player who never finishes
// can still be given it again.
func TestGetRandomUnplayedPuzzleIgnoresIncompleteGames(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	id := savePuzzle(t, store, "bell")

	if _, err := store.CreateGame(ctx, "player-a", id); err != nil {
		t.Fatalf("CreateGame() returned error: %v", err)
	}

	got, err := store.GetRandomUnplayedPuzzle(ctx, "player-a")

	if err != nil {
		t.Fatalf("GetRandomUnplayedPuzzle() returned error: %v", err)
	}

	if got != id {
		t.Errorf("puzzle = %d, want %d", got, id)
	}
}

func TestGetRandomUnplayedPuzzleEmptyLibrary(t *testing.T) {
	store := testStore(t)

	_, err := store.GetRandomUnplayedPuzzle(context.Background(), "player-a")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetRandomUnplayedPuzzle() = %v, want ErrNotFound", err)
	}
}

func TestGetGame(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	puzzleID := savePuzzle(t, store, "bell")
	gameID, err := store.CreateGame(ctx, "player-a", puzzleID)

	if err != nil {
		t.Fatalf("CreateGame() returned error: %v", err)
	}

	game, err := store.GetGame(ctx, gameID, "player-a")

	if err != nil {
		t.Fatalf("GetGame() returned error: %v", err)
	}

	if game.PuzzleID != puzzleID {
		t.Errorf("PuzzleID = %d, want %d", game.PuzzleID, puzzleID)
	}

	if game.CluesShown != 1 {
		t.Errorf("CluesShown = %d, want 1", game.CluesShown)
	}

	if game.Completed {
		t.Error("Completed = true, want false")
	}
}

// A game is only reachable by the player it belongs to. Someone holding another
// player's game id is told the game does not exist rather than that it is theirs.
func TestGetGameRejectsOtherPlayers(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	puzzleID := savePuzzle(t, store, "bell")
	gameID, err := store.CreateGame(ctx, "player-a", puzzleID)

	if err != nil {
		t.Fatalf("CreateGame() returned error: %v", err)
	}

	if _, err := store.GetGame(ctx, gameID, "player-b"); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetGame() for another player = %v, want ErrNotFound", err)
	}

	if _, err := store.GetGame(ctx, "no-such-game", "player-a"); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetGame() for unknown id = %v, want ErrNotFound", err)
	}
}

// Revealing clues stops at the last one rather than failing the position check,
// so a player who keeps guessing after the final clue is simply told nothing new.
func TestIncrementCluesShownSaturates(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	puzzleID := savePuzzle(t, store, "bell")
	gameID, err := store.CreateGame(ctx, "player-a", puzzleID)

	if err != nil {
		t.Fatalf("CreateGame() returned error: %v", err)
	}

	// Four increments take a new game from one clue to all five.
	for i := range puzzle.ClueCount - 1 {
		rows, err := store.IncrementCluesShown(ctx, gameID)

		if err != nil {
			t.Fatalf("IncrementCluesShown() returned error on call %d: %v", i+1, err)
		}

		if rows != 1 {
			t.Errorf("rows affected on call %d = %d, want 1", i+1, rows)
		}
	}

	rows, err := store.IncrementCluesShown(ctx, gameID)

	if err != nil {
		t.Fatalf("IncrementCluesShown() returned error: %v", err)
	}

	if rows != 0 {
		t.Errorf("rows affected past the last clue = %d, want 0", rows)
	}

	game, err := store.GetGame(ctx, gameID, "player-a")

	if err != nil {
		t.Fatalf("GetGame() returned error: %v", err)
	}

	if game.CluesShown != puzzle.ClueCount {
		t.Errorf("CluesShown = %d, want %d", game.CluesShown, puzzle.ClueCount)
	}
}

// Completing a finished game leaves the original result intact, so a repeated
// request cannot turn a loss into a win.
func TestCompleteGameIsIdempotent(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	puzzleID := savePuzzle(t, store, "bell")
	gameID, err := store.CreateGame(ctx, "player-a", puzzleID)

	if err != nil {
		t.Fatalf("CreateGame() returned error: %v", err)
	}

	rows, err := store.CompleteGame(ctx, gameID, "2026-01-01T00:00:00Z", false)

	if err != nil {
		t.Fatalf("CompleteGame() returned error: %v", err)
	}

	if rows != 1 {
		t.Errorf("rows affected = %d, want 1", rows)
	}

	rows, err = store.CompleteGame(ctx, gameID, "2026-01-02T00:00:00Z", true)

	if err != nil {
		t.Fatalf("CompleteGame() returned error: %v", err)
	}

	if rows != 0 {
		t.Errorf("rows affected on second call = %d, want 0", rows)
	}

	game, err := store.GetGame(ctx, gameID, "player-a")

	if err != nil {
		t.Fatalf("GetGame() returned error: %v", err)
	}

	if !game.Completed {
		t.Error("Completed = false, want true")
	}

	if game.Won == nil || *game.Won {
		t.Errorf("Won = %v, want false", game.Won)
	}
}

func TestGetPuzzleCluesUpTo(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	puzzleID := savePuzzle(t, store, "bell")

	clues, err := store.GetPuzzleCluesUpTo(ctx, puzzleID, 3)

	if err != nil {
		t.Fatalf("GetPuzzleCluesUpTo() returned error: %v", err)
	}

	want := []string{"bell one", "bell two", "bell three"}

	if !slices.Equal(clues, want) {
		t.Errorf("clues = %v, want %v", clues, want)
	}
}

func TestGetPuzzleNotFound(t *testing.T) {
	store := testStore(t)

	_, err := store.GetPuzzle(context.Background(), 404)

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetPuzzle() = %v, want ErrNotFound", err)
	}
}

// PuzzleSupply drives the generator, so it counts puzzles nobody has finished
// and players who have started at least one game.
func TestPuzzleSupply(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	first := savePuzzle(t, store, "bell")
	savePuzzle(t, store, "church")
	savePuzzle(t, store, "school")

	completeGame(t, store, "player-a", first)

	// An abandoned game still makes player-b a player, but consumes no puzzle.
	if _, err := store.CreateGame(ctx, "player-b", savePuzzle(t, store, "brass")); err != nil {
		t.Fatalf("CreateGame() returned error: %v", err)
	}

	supply, err := store.PuzzleSupply(ctx)

	if err != nil {
		t.Fatalf("PuzzleSupply() returned error: %v", err)
	}

	if supply.Unplayed != 3 {
		t.Errorf("Unplayed = %d, want 3", supply.Unplayed)
	}

	if supply.Players != 2 {
		t.Errorf("Players = %d, want 2", supply.Players)
	}
}
