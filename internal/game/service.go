package game

import (
	"context"
	"errors"
	"time"

	"github.com/danny-pearson/word-association/internal/db"
	"github.com/danny-pearson/word-association/internal/puzzle"
)

type Store interface {
	GetRandomUnplayedPuzzle(ctx context.Context, playerID string) (int64, error)
	CreateGame(ctx context.Context, playerID string, puzzleID int64) (string, error)
	GetGame(ctx context.Context, id string, playerID string) (db.Game, error)
	CompleteGame(ctx context.Context, id string, completedAt string, won bool) (int64, error)
	IncrementCluesShown(ctx context.Context, id string) (int64, error)
	GetPuzzle(ctx context.Context, id int64) (puzzle.Puzzle, error)
	GetPuzzleCluesUpTo(ctx context.Context, id int64, position int64) ([]string, error)
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Start(ctx context.Context, playerID string) (Game, error) {
	var game Game

	puzzleID, err := s.store.GetRandomUnplayedPuzzle(ctx, playerID)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return game, ErrOutOfPuzzles
		}
		return game, err
	}

	gameID, err := s.store.CreateGame(ctx, playerID, puzzleID)

	if err != nil {
		return game, err
	}

	clues, err := s.store.GetPuzzleCluesUpTo(ctx, puzzleID, 1)

	if err != nil {
		return game, err
	}

	game = Game{
		ID:    gameID,
		Clues: clues,
	}

	return game, nil
}

func (s *Service) Guess(ctx context.Context, gameID string, playerID string, guess string) (GuessResult, error) {
	var result GuessResult

	game, err := s.store.GetGame(ctx, gameID, playerID)

	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return result, ErrGameNotFound
		}
		return result, err
	}

	if game.Completed {
		return result, ErrGameComplete
	}

	p, err := s.store.GetPuzzle(ctx, game.PuzzleID)

	if err != nil {
		return result, err
	}

	normalizedGuess, err := puzzle.Normalize(guess)

	if err != nil {
		return result, err
	}

	isCorrect := normalizedGuess == p.AnswerNormalized

	var (
		revealedClues []string
		answer        string
		isComplete    bool
		hasWon        bool
	)

	if isCorrect {
		revealedClues, err = s.store.GetPuzzleCluesUpTo(ctx, p.ID, puzzle.ClueCount)

		if err != nil {
			return result, err
		}

		isComplete = true
		hasWon = true
	} else {
		if game.CluesShown < puzzle.ClueCount {
			if _, err := s.store.IncrementCluesShown(ctx, gameID); err != nil {
				return result, err
			}
		} else {
			isComplete = true
		}

		revealedClues, err = s.store.GetPuzzleCluesUpTo(ctx, p.ID, min(game.CluesShown+1, puzzle.ClueCount))

		if err != nil {
			return result, err
		}
	}

	if isComplete {
		_, err := s.store.CompleteGame(
			ctx,
			gameID,
			time.Now().UTC().Format(time.RFC3339),
			hasWon,
		)

		if err != nil {
			return result, err
		}

		answer = p.Answer
	}

	result = GuessResult{
		Clues:  revealedClues,
		Answer: answer,
		HasWon: hasWon,
	}

	return result, nil
}
