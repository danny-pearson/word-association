package puzzle

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Normalize converts text into the canonical form used for matching and
// deduplication. It strips diacritics, lowercases, removes any leading
// article ("the", "a", "an"), and discards everything that is not a letter
// or digit — including spaces, hyphens, and apostrophes.
//
// Both stored values and incoming guesses must be normalized with this
// function; comparing forms produced by different rules will silently fail
// to match. Normalize returns an error only if text is not valid UTF-8.
func Normalize(text string) (string, error) {
	transformer := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)

	result, _, err := transform.String(transformer, text)

	if err != nil {
		return "", err
	}

	result = strings.ToLower(result)
	fields := strings.Fields(result)

	if len(fields) > 1 && isArticle(fields[0]) {
		fields = fields[1:]
	}

	// Space-collapsing may as well happen here seeing as it's a join.
	result = strings.Join(fields, "")

	result = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, result)

	return result, nil
}

func isArticle(word string) bool {
	return word == "the" ||
		word == "a" ||
		word == "an"
}
