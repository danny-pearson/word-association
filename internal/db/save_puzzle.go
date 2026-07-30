package db

import (
	"context"
	"time"

	"github.com/danny-pearson/word-association/internal/db/gen"
	"github.com/danny-pearson/word-association/internal/puzzle"
)

func (s *Store) SavePuzzle(ctx context.Context, p puzzle.GeneratedPuzzle, model string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)

	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	normalizedAnswer, err := puzzle.Normalize(p.Answer)

	if err != nil {
		return 0, err
	}

	puzzleID, err := qtx.CreatePuzzle(ctx, gen.CreatePuzzleParams{
		Answer:           p.Answer,
		AnswerNormalized: normalizedAnswer,
		SourceModel:      &model,
		CreatedAt:        time.Now().UTC().Format(time.RFC3339),
	})

	if err != nil {
		return 0, err
	}

	for i, c := range p.Clues {
		normalizedClue, err := puzzle.Normalize(c)

		if err != nil {
			return 0, err
		}

		clueID, err := qtx.UpsertClue(ctx, gen.UpsertClueParams{
			Text:           c,
			TextNormalized: normalizedClue,
		})

		if err != nil {
			return 0, err
		}

		if err := qtx.LinkClue(ctx, gen.LinkClueParams{
			PuzzleID: puzzleID,
			ClueID:   clueID,
			Position: int64(i + 1),
		}); err != nil {
			return 0, err
		}
	}

	return puzzleID, tx.Commit()
}
