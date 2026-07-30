package token

import (
	"encoding/base64"
	"testing"
)

// Tokens identify a player and authorize resuming a game, so a token that
// repeats would hand one player another's session. This does not measure the
// quality of the randomness — that is crypto/rand's job — but it does catch a
// token that was never randomized at all.
func TestNewIsUnique(t *testing.T) {
	const count = 1000

	seen := make(map[string]bool, count)

	for i := range count {
		got, err := New()

		if err != nil {
			t.Fatalf("New() returned error on call %d: %v", i+1, err)
		}

		if seen[got] {
			t.Fatalf("New() returned a duplicate token on call %d: %q", i+1, got)
		}

		seen[got] = true
	}
}

// The encoding has to survive a round trip through a cookie and a URL, so it
// must be the unpadded URL-safe alphabet rather than the standard one.
func TestNewIsURLSafeBase64(t *testing.T) {
	got, err := New()

	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(got)

	if err != nil {
		t.Fatalf("New() = %q, which is not raw URL-safe base64: %v", got, err)
	}

	if len(decoded) != 16 {
		t.Errorf("New() decoded to %d bytes, want 16", len(decoded))
	}
}
