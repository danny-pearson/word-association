package game

type Game struct {
	ID    string   `json:"id"`
	Clues []string `json:"clues"`
}

type GuessResult struct {
	Clues  []string `json:"clues"`
	Answer string   `json:"answer"`
	HasWon bool     `json:"has_won"`
}
