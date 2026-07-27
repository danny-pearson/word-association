package puzzle

type Puzzle struct {
	Answer string   `json:"answer"`
	Clues  []string `json:"clues"`
}
