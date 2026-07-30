package puzzle

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// Validate reports whether p is a well-formed puzzle. It returns nil if the
// puzzle is usable, or a ValidationError carrying the RejectionReason of the
// first rule that failed. Rules are checked in a fixed order, so a puzzle that
// breaks several only reports one.
//
// Validate is pure: it does not check whether the answer already exists in the
// library, since that requires the stored puzzles. Callers must apply that
// check separately, using ReasonDuplicateAnswer.
func Validate(p GeneratedPuzzle) error {
	if len(p.Clues) != 5 {
		return ValidationError{
			Reason: ReasonIncorrectClueCount,
			Detail: fmt.Sprintf("expected 5 clues, got %d", len(p.Clues)),
		}
	}

	answerWords, err := normalizeWords(p.Answer)

	if err != nil {
		return ValidationError{
			Reason: ReasonInvalidText,
			Detail: fmt.Sprintf("answer %q could not be normalized: %v", p.Answer, err),
		}
	}

	if len(answerWords) == 0 {
		return ValidationError{
			Reason: ReasonEmptyNormalized,
			Detail: fmt.Sprintf("answer %q is empty after normalizing words", p.Answer),
		}
	}

	answerStems := singularizeAll(answerWords)

	normalizedClues := make([]string, 0, len(p.Clues))

	for i, clue := range p.Clues {
		clueWords, err := normalizeWords(clue)

		if err != nil {
			return ValidationError{
				Reason: ReasonInvalidText,
				Detail: fmt.Sprintf("clue %q could not be normalized: %v", clue, err),
			}
		}

		if len(clueWords) == 0 {
			return ValidationError{
				Reason: ReasonEmptyNormalized,
				Detail: fmt.Sprintf("clue %q is empty after normalizing words", clue),
			}
		}

		if len(clueWords) > 2 {
			return ValidationError{
				Reason: ReasonClueTooManyWords,
				Detail: fmt.Sprintf("expected no more than 2 words, got %d", len(clueWords)),
			}
		}

		if word, ok := sharedWord(answerStems, singularizeAll(clueWords)); ok {
			return ValidationError{
				Reason: ReasonClueContainsAnswer,
				Detail: fmt.Sprintf("clue %d shares the word %q with the answer %q", i+1, word, p.Answer),
			}
		}

		normalizedClues = append(normalizedClues, strings.Join(clueWords, " "))
	}

	if first, second, ok := hasDuplicates(normalizedClues); ok {
		return ValidationError{
			Reason: ReasonDuplicateClue,
			Detail: fmt.Sprintf("clues %d and %d are both %q", first+1, second+1, normalizedClues[first]),
		}
	}

	return nil
}

// normalizeWords splits text on anything that is not a letter or digit and
// normalizes each word individually. It exists because Normalize collapses
// whitespace, so "apple pie" becomes "applepie" and the individual words can
// no longer be recovered from its output.
//
// Words that normalize to nothing are dropped, so text consisting only of
// punctuation yields an empty slice rather than an error.
func normalizeWords(text string) ([]string, error) {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	normalized := make([]string, 0, len(words))

	for _, w := range words {
		n, err := Normalize(w)

		if err != nil {
			return nil, err
		}

		if n != "" {
			normalized = append(normalized, n)
		}
	}
	return normalized, nil
}

// sharedWord returns the first word present in both a and b. Word-level
// comparison is used rather than substring matching so that an answer such as
// "ear" is not treated as leaking into a clue like "hearing aid".
func sharedWord(a, b []string) (string, bool) {
	for _, v := range a {
		if slices.Contains(b, v) {
			return v, true
		}
	}
	return "", false
}

// hasDuplicates returns the indices of the first repeated value in clues,
// earlier index first.
func hasDuplicates(clues []string) (first, second int, found bool) {
	seen := make(map[string]int, len(clues))

	for i, c := range clues {
		if j, ok := seen[c]; ok {
			return j, i, true
		}
		seen[c] = i
	}
	return -1, -1, false
}

// singularize strips a common English plural suffix from word. It is
// deliberately naive: it exists only so that a clue like "church bells" is
// recognized as leaking the answer "bell", and is not general morphology.
// Irregular plurals ("mice", "children") and verb forms ("ringing") are not
// handled — stripping "-ing" would wreck ordinary words such as "ring".
func singularize(word string) string {
	switch {
	case strings.HasSuffix(word, "ies") && len(word) > 4:
		return strings.TrimSuffix(word, "ies") + "y"
	case strings.HasSuffix(word, "es") && len(word) > 3:
		return strings.TrimSuffix(word, "es")
	case strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") && len(word) > 2:
		return strings.TrimSuffix(word, "s")
	}
	return word
}

// singularizeAll applies singularize to every word.
func singularizeAll(words []string) []string {
	singularWords := make([]string, 0, len(words))

	for _, w := range words {
		singularWords = append(singularWords, singularize(w))
	}
	return singularWords
}
