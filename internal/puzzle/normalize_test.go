package puzzle

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercases", "BELL", "bell"},
		{"strips leading article", "The Beatles", "beatles"},
		{"strips hyphens", "Spider-Man", "spiderman"},
		{"strips apostrophes", "don't", "dont"},
		{"collapses spaces", "ice cream", "icecream"},
		{"strips diacritics", "café", "cafe"},
		{"keeps article prefix within a word", "theatre", "theatre"},
		{"strips only the first article", "The The", "the"},
		{"keeps a lone article", "the", "the"},
		{"punctuation only normalizes to empty", "...", ""},
		{"whitespace only normalizes to empty", "  ", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Normalize(tc.input)

			if err != nil {
				t.Fatalf("Normalize(%q) returned error: %v", tc.input, err)
			}

			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
