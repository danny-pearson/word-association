package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/danny-pearson/word-association/internal/db/gen"
	"github.com/danny-pearson/word-association/internal/puzzle"
	"github.com/danny-pearson/word-association/internal/token"
)

type Store struct {
	db      *sql.DB
	queries *gen.Queries
}

func NewStore(conn *sql.DB) *Store {
	return &Store{
		db:      conn,
		queries: gen.New(conn),
	}
}

func (s *Store) RecentAnswers(ctx context.Context, limit int64) ([]string, error) {
	return s.queries.RecentAnswers(ctx, limit)
}

func (s *Store) FindExistingAnswers(ctx context.Context, answers []string) ([]string, error) {
	return s.queries.FindExistingAnswers(ctx, answers)
}

func (s *Store) GetPuzzle(ctx context.Context, id int64) (puzzle.Puzzle, error) {
	var result puzzle.Puzzle

	row, err := s.queries.GetPuzzle(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, ErrNotFound
		}
		return result, err
	}

	result = puzzle.Puzzle{
		ID:               row.ID,
		Answer:           row.Answer,
		AnswerNormalized: row.AnswerNormalized,
	}

	return result, nil
}

func (s *Store) GetRandomUnplayedPuzzle(ctx context.Context, playerID string) (int64, error) {
	id, err := s.queries.GetRandomUnplayedPuzzle(ctx, playerID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}

	return id, err
}

func (s *Store) GetPuzzleClue(ctx context.Context, id int64, position int64) (string, error) {
	return s.queries.GetPuzzleClue(ctx, gen.GetPuzzleClueParams{
		PuzzleID: id,
		Position: position,
	})
}

func (s *Store) GetPuzzleCluesUpTo(ctx context.Context, id int64, position int64) ([]string, error) {
	return s.queries.GetPuzzleCluesUpTo(ctx, gen.GetPuzzleCluesUpToParams{
		PuzzleID: id,
		Position: position,
	})
}

func (s *Store) PuzzleSupply(ctx context.Context) (gen.PuzzleSupplyRow, error) {
	return s.queries.PuzzleSupply(ctx)
}

func (s *Store) CreateGame(ctx context.Context, playerID string, puzzleID int64) (string, error) {
	gameID, err := token.New()

	if err != nil {
		return "", err
	}

	if err = s.queries.CreateGame(ctx, gen.CreateGameParams{
		ID:        gameID,
		PlayerID:  playerID,
		PuzzleID:  puzzleID,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		return "", err
	}

	return gameID, nil
}

func (s *Store) GetGame(ctx context.Context, id string, playerID string) (gen.GetGameRow, error) {
	row, err := s.queries.GetGame(ctx, gen.GetGameParams{
		ID:       id,
		PlayerID: playerID,
	})

	if errors.Is(err, sql.ErrNoRows) {
		return row, ErrNotFound
	}

	return row, err
}

func (s *Store) CompleteGame(ctx context.Context, id string, completedAt string, won bool) (int64, error) {
	return s.queries.CompleteGame(ctx, gen.CompleteGameParams{
		ID:          id,
		CompletedAt: &completedAt,
		Won:         &won,
	})
}

func (s *Store) IncrementCluesShown(ctx context.Context, id string) (int64, error) {
	return s.queries.IncrementCluesShown(ctx, id)
}
