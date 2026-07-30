package puzzle

type Puzzle struct {
	ID               int64  `json:"id"`
	Answer           string `json:"answer"`
	AnswerNormalized string `json:"answer_normalized"`
}

type GeneratedPuzzle struct {
	Answer string   `json:"answer"`
	Clues  []string `json:"clues"`
}
