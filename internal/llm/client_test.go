package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient starts a server running handler and returns a Client pointed at
// it, so tests exercise the real request and response path without network access.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return New(server.URL, "test-key", "test-model", 0.9)
}

// respondWith returns a handler that writes status and body to every request.
func respondWith(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

func TestSendReturnsCompletion(t *testing.T) {
	client := newTestClient(t, respondWith(http.StatusOK,
		`{"choices":[{"message":{"role":"assistant","content":"a completion"}}]}`,
	))

	got, err := client.Send(context.Background(), "a prompt")

	if err != nil {
		t.Fatalf("Send() returned error: %v", err)
	}

	if got != "a completion" {
		t.Errorf("Send() = %q, want %q", got, "a completion")
	}
}

func TestSendBuildsRequest(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotAuth   string
		gotType   string
		gotBody   chatRequest
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotType = r.Header.Get("Content-Type")

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}

		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	})

	if _, err := client.Send(context.Background(), "a prompt"); err != nil {
		t.Fatalf("Send() returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}

	if gotPath != "/chat/completions" {
		t.Errorf("path = %q, want %q", gotPath, "/chat/completions")
	}

	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}

	if gotType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotType, "application/json")
	}

	if gotBody.Model != "test-model" {
		t.Errorf("model = %q, want %q", gotBody.Model, "test-model")
	}

	if gotBody.Temperature != 0.9 {
		t.Errorf("temperature = %v, want %v", gotBody.Temperature, 0.9)
	}

	if len(gotBody.Messages) != 1 {
		t.Fatalf("got %d messages, want 1", len(gotBody.Messages))
	}

	if gotBody.Messages[0].Role != "user" {
		t.Errorf("role = %q, want %q", gotBody.Messages[0].Role, "user")
	}

	if gotBody.Messages[0].Content != "a prompt" {
		t.Errorf("content = %q, want %q", gotBody.Messages[0].Content, "a prompt")
	}
}

func TestSendStatusErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{
			name:   "rate limited",
			status: http.StatusTooManyRequests,
			body:   `{"error":{"message":"slow down"}}`,
			want:   ErrRateLimited,
		},
		{
			name:   "unauthorized",
			status: http.StatusUnauthorized,
			body:   `{"error":{"message":"bad key"}}`,
			want:   ErrUnauthorized,
		},
		{
			name:   "payment required",
			status: http.StatusPaymentRequired,
			body:   `{"error":{"message":"no credit"}}`,
			want:   ErrNoCredit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(t, respondWith(tc.status, tc.body))

			_, err := client.Send(context.Background(), "a prompt")

			if !errors.Is(err, tc.want) {
				t.Errorf("Send() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestSendRateLimitErrorIncludesBody(t *testing.T) {
	client := newTestClient(t, respondWith(http.StatusTooManyRequests,
		`{"error":{"metadata":{"raw":"temporarily rate-limited upstream"}}}`,
	))

	_, err := client.Send(context.Background(), "a prompt")

	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("Send() = %v, want ErrRateLimited", err)
	}

	if !strings.Contains(err.Error(), "temporarily rate-limited upstream") {
		t.Errorf("error %q does not include the response body", err)
	}
}

func TestSendUnexpectedStatus(t *testing.T) {
	client := newTestClient(t, respondWith(http.StatusInternalServerError, "upstream exploded"))

	_, err := client.Send(context.Background(), "a prompt")

	if err == nil {
		t.Fatal("Send() = nil, want an error")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error %q does not mention the status code", err)
	}

	if !strings.Contains(err.Error(), "upstream exploded") {
		t.Errorf("error %q does not include the response body", err)
	}
}

func TestSendEmptyChoices(t *testing.T) {
	client := newTestClient(t, respondWith(http.StatusOK, `{"choices":[]}`))

	_, err := client.Send(context.Background(), "a prompt")

	if !errors.Is(err, ErrEmptyResponse) {
		t.Errorf("Send() = %v, want ErrEmptyResponse", err)
	}
}

// A 200 response carrying an error object is reported rather than treated as a
// completion, since OpenRouter does not always signal failures with a status code.
func TestSendErrorInBody(t *testing.T) {
	client := newTestClient(t, respondWith(http.StatusOK,
		`{"error":{"message":"model unavailable","code":503}}`,
	))

	_, err := client.Send(context.Background(), "a prompt")

	if err == nil {
		t.Fatal("Send() = nil, want an error")
	}

	if !strings.Contains(err.Error(), "model unavailable") {
		t.Errorf("error %q does not include the reported message", err)
	}
}

func TestSendMalformedJSON(t *testing.T) {
	client := newTestClient(t, respondWith(http.StatusOK, `{"choices":[`))

	_, err := client.Send(context.Background(), "a prompt")

	if err == nil {
		t.Fatal("Send() = nil, want an error")
	}
}

func TestSendCancelledContext(t *testing.T) {
	client := newTestClient(t, respondWith(http.StatusOK,
		`{"choices":[{"message":{"content":"ok"}}]}`,
	))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Send(ctx, "a prompt")

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Send() = %v, want context.Canceled", err)
	}
}
