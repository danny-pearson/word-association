package game

import "errors"

var (
	ErrGameNotFound = errors.New("game: not found")
	ErrOutOfPuzzles = errors.New("game: out of puzzles")
	ErrGameComplete = errors.New("game: already complete")
)
