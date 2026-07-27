package puzzle

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		puzzle Puzzle
		want   RejectionReason
	}{
		{
			name: "valid puzzle",
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church", "brass", "school", "ring", "tower"},
			},
		},
		{
			name: "too few clues",
			want: ReasonIncorrectClueCount,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church", "brass"},
			},
		},
		{
			name: "too many clues",
			want: ReasonIncorrectClueCount,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church", "brass", "school", "ring", "tower", "curve"},
			},
		},
		{
			name: "duplicate clue",
			want: ReasonDuplicateClue,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church", "church", "school", "ring", "tower"},
			},
		},
		{
			name: "duplicate clue differing only by case",
			want: ReasonDuplicateClue,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church", "Church", "school", "ring", "tower"},
			},
		},
		{
			name: "clue contains answer",
			want: ReasonClueContainsAnswer,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church bell", "brass", "school", "ring", "tower"},
			},
		},
		{
			name: "clue contains answer across a hyphen",
			want: ReasonClueContainsAnswer,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church", "brass", "school", "ring", "Bell-Tower"},
			},
		},
		{
			name: "clue contains a plural of the answer",
			want: ReasonClueContainsAnswer,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church bells", "brass", "school", "ring", "tower"},
			},
		},
		{
			name: "clue is a word from a multi-word answer",
			want: ReasonClueContainsAnswer,
			puzzle: Puzzle{
				Answer: "apple pie",
				Clues:  []string{"apple", "dessert", "crumble", "pastry", "custard"},
			},
		},
		{
			name: "clue shares a word with a multi-word answer",
			want: ReasonClueContainsAnswer,
			puzzle: Puzzle{
				Answer: "ice cream",
				Clues:  []string{"cold", "cream cheese", "cone", "sundae", "vanilla"},
			},
		},
		{
			name: "answer is a substring of a clue word",
			puzzle: Puzzle{
				Answer: "ear",
				Clues:  []string{"hearing aid", "lobe", "drum", "wax", "muffs"},
			},
		},
		{
			name: "clue has too many words",
			want: ReasonClueTooManyWords,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church tower bell", "brass", "school", "ring", "tower"},
			},
		},
		{
			name: "answer normalizes to nothing",
			want: ReasonEmptyNormalized,
			puzzle: Puzzle{
				Answer: "...",
				Clues:  []string{"church", "brass", "school", "ring", "tower"},
			},
		},
		{
			name: "clue normalizes to nothing",
			want: ReasonEmptyNormalized,
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church", "---", "school", "ring", "tower"},
			},
		},
		{
			name: "two word clues are allowed",
			puzzle: Puzzle{
				Answer: "bell",
				Clues:  []string{"church tower", "brass band", "school yard", "door frame", "wedding day"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.puzzle)

			if tc.want == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}

			var vErr ValidationError

			if !errors.As(err, &vErr) {
				t.Fatalf("Validate() = %v, want ValidationError", err)
			}

			if vErr.Reason != tc.want {
				t.Errorf("Reason = %q, want %q", vErr.Reason, tc.want)
			}
		})
	}
}
