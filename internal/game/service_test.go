package game

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/danny-pearson/word-association/internal/db"
	"github.com/danny-pearson/word-association/internal/puzzle"
)

// fakeStore is an in-memory Store holding a single puzzle and a single game,
// which is all any one call to the service touches. Each field mirrors what the
// real store would return, and the Err fields force the failure paths.
type fakeStore struct {
	puzzle     puzzle.Puzzle
	clues      []string
	game       db.Game
	nextGameID string

	randomErr    error
	getGameErr   error
	getPuzzleErr error

	// Recorded so tests can assert on writes rather than only on return values.
	increments  int
	completedAt string
	completedAs bool
	completed   bool
}

func (f *fakeStore) GetRandomUnplayedPuzzle(ctx context.Context, playerID string) (int64, error) {
	if f.randomErr != nil {
		return 0, f.randomErr
	}
	return f.puzzle.ID, nil
}

func (f *fakeStore) CreateGame(ctx context.Context, playerID string, puzzleID int64) (string, error) {
	return f.nextGameID, nil
}

func (f *fakeStore) GetGame(ctx context.Context, id string, playerID string) (db.Game, error) {
	if f.getGameErr != nil {
		return db.Game{}, f.getGameErr
	}
	return f.game, nil
}

func (f *fakeStore) CompleteGame(ctx context.Context, id string, completedAt string, won bool) (int64, error) {
	f.completed = true
	f.completedAt = completedAt
	f.completedAs = won
	return 1, nil
}

func (f *fakeStore) IncrementCluesShown(ctx context.Context, id string) (int64, error) {
	f.increments++
	return 1, nil
}

func (f *fakeStore) GetPuzzle(ctx context.Context, id int64) (puzzle.Puzzle, error) {
	if f.getPuzzleErr != nil {
		return puzzle.Puzzle{}, f.getPuzzleErr
	}
	return f.puzzle, nil
}

func (f *fakeStore) GetPuzzleCluesUpTo(ctx context.Context, id int64, position int64) ([]string, error) {
	if position > int64(len(f.clues)) {
		position = int64(len(f.clues))
	}
	return f.clues[:position], nil
}

// newFake returns a store holding one puzzle with five clues and a game that has
// been started but not played, which is the state most tests begin from.
func newFake() *fakeStore {
	return &fakeStore{
		puzzle: puzzle.Puzzle{
			ID:               1,
			Answer:           "Bell",
			AnswerNormalized: "bell",
		},
		clues:      []string{"church", "brass", "school", "ring", "tower"},
		nextGameID: "game-1",
		game: db.Game{
			ID:         "game-1",
			PuzzleID:   1,
			CluesShown: 1,
		},
	}
}

func TestStartRevealsOneClue(t *testing.T) {
	store := newFake()
	service := New(store)

	game, err := service.Start(context.Background(), "player-a")

	if err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}

	if game.ID != "game-1" {
		t.Errorf("ID = %q, want %q", game.ID, "game-1")
	}

	if !slices.Equal(game.Clues, []string{"church"}) {
		t.Errorf("clues = %v, want one clue", game.Clues)
	}
}

func TestStartOutOfPuzzles(t *testing.T) {
	store := newFake()
	store.randomErr = db.ErrNotFound

	_, err := New(store).Start(context.Background(), "player-a")

	if !errors.Is(err, ErrOutOfPuzzles) {
		t.Errorf("Start() = %v, want ErrOutOfPuzzles", err)
	}
}

// A storage failure is not the player running out of puzzles, and must not be
// reported as though it were.
func TestStartPropagatesStoreErrors(t *testing.T) {
	store := newFake()
	store.randomErr = errors.New("database is on fire")

	_, err := New(store).Start(context.Background(), "player-a")

	if errors.Is(err, ErrOutOfPuzzles) {
		t.Errorf("Start() = %v, want the underlying error", err)
	}

	if err == nil {
		t.Error("Start() = nil, want an error")
	}
}

func TestGuessCorrect(t *testing.T) {
	store := newFake()
	service := New(store)

	result, err := service.Guess(context.Background(), "game-1", "player-a", "bell")

	if err != nil {
		t.Fatalf("Guess() returned error: %v", err)
	}

	if !result.HasWon {
		t.Error("HasWon = false, want true")
	}

	if result.Answer != "Bell" {
		t.Errorf("Answer = %q, want %q", result.Answer, "Bell")
	}

	if len(result.Clues) != puzzle.ClueCount {
		t.Errorf("got %d clues, want all %d", len(result.Clues), puzzle.ClueCount)
	}

	if store.increments != 0 {
		t.Errorf("clues revealed on a winning guess = %d, want 0", store.increments)
	}

	if !store.completed || !store.completedAs {
		t.Errorf("game completed = %v, won = %v, want both true", store.completed, store.completedAs)
	}
}

// A guess is matched against the stored answer in normalized form, so casing,
// punctuation and a leading article never decide whether a player is right.
func TestGuessNormalizesInput(t *testing.T) {
	for _, guess := range []string{"bell", "BELL", "Bell", " bell ", "the bell", "béll"} {
		t.Run(guess, func(t *testing.T) {
			result, err := New(newFake()).Guess(context.Background(), "game-1", "player-a", guess)

			if err != nil {
				t.Fatalf("Guess(%q) returned error: %v", guess, err)
			}

			if !result.HasWon {
				t.Errorf("Guess(%q) was not accepted", guess)
			}
		})
	}
}

func TestGuessWrongRevealsNextClue(t *testing.T) {
	store := newFake()

	result, err := New(store).Guess(context.Background(), "game-1", "player-a", "church")

	if err != nil {
		t.Fatalf("Guess() returned error: %v", err)
	}

	if result.HasWon {
		t.Error("HasWon = true, want false")
	}

	// The answer stays hidden while the game is still playable.
	if result.Answer != "" {
		t.Errorf("Answer = %q, want it withheld", result.Answer)
	}

	if len(result.Clues) != 2 {
		t.Errorf("got %d clues, want 2", len(result.Clues))
	}

	if store.increments != 1 {
		t.Errorf("increments = %d, want 1", store.increments)
	}

	if store.completed {
		t.Error("game was completed on a wrong guess with clues remaining")
	}
}

// Guessing wrong on the last clue ends the game and reveals the answer, rather
// than leaving the player stuck with nothing left to reveal.
func TestGuessWrongOnFinalClueLoses(t *testing.T) {
	store := newFake()
	store.game.CluesShown = puzzle.ClueCount

	result, err := New(store).Guess(context.Background(), "game-1", "player-a", "wrong")

	if err != nil {
		t.Fatalf("Guess() returned error: %v", err)
	}

	if result.HasWon {
		t.Error("HasWon = true, want false")
	}

	if result.Answer != "Bell" {
		t.Errorf("Answer = %q, want it revealed", result.Answer)
	}

	if len(result.Clues) != puzzle.ClueCount {
		t.Errorf("got %d clues, want all %d", len(result.Clues), puzzle.ClueCount)
	}

	if store.increments != 0 {
		t.Errorf("increments past the last clue = %d, want 0", store.increments)
	}

	if !store.completed || store.completedAs {
		t.Errorf("game completed = %v, won = %v, want completed and lost", store.completed, store.completedAs)
	}
}

func TestGuessOnCompletedGame(t *testing.T) {
	store := newFake()
	store.game.Completed = true

	_, err := New(store).Guess(context.Background(), "game-1", "player-a", "bell")

	if !errors.Is(err, ErrGameComplete) {
		t.Errorf("Guess() = %v, want ErrGameComplete", err)
	}
}

func TestGuessUnknownGame(t *testing.T) {
	store := newFake()
	store.getGameErr = db.ErrNotFound

	_, err := New(store).Guess(context.Background(), "game-1", "player-a", "bell")

	if !errors.Is(err, ErrGameNotFound) {
		t.Errorf("Guess() = %v, want ErrGameNotFound", err)
	}
}

// A game pointing at a puzzle that is not there is broken state rather than
// anything the player did, so it must not be dressed up as a client error.
func TestGuessMissingPuzzleIsNotAClientError(t *testing.T) {
	store := newFake()
	store.getPuzzleErr = db.ErrNotFound

	_, err := New(store).Guess(context.Background(), "game-1", "player-a", "bell")

	if err == nil {
		t.Fatal("Guess() = nil, want an error")
	}

	if errors.Is(err, ErrGameNotFound) || errors.Is(err, ErrGameComplete) {
		t.Errorf("Guess() = %v, want the underlying error", err)
	}
}
