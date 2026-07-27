package puzzle

import "fmt"

type RejectionReason string

const (
	ReasonIncorrectClueCount RejectionReason = "incorrect_clue_count"
	ReasonDuplicateClue      RejectionReason = "duplicate_clue"
	ReasonClueContainsAnswer RejectionReason = "clue_contains_answer"
	ReasonClueTooManyWords   RejectionReason = "clue_too_many_words"
	ReasonEmptyNormalized    RejectionReason = "empty_normalized"
	ReasonDuplicateAnswer    RejectionReason = "duplicate_answer"
	ReasonInvalidText        RejectionReason = "invalid_text"
)

type ValidationError struct {
	Reason RejectionReason
	Detail string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Reason, e.Detail)
}
