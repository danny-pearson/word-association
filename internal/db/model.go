package db

type Game struct {
	ID         string
	PuzzleID   int64
	CluesShown int64
	Completed  bool
	Won        *bool
}

type PuzzleSupply struct {
	Unplayed int64
	Players  int64
}
