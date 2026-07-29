package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL     string
	apiKey      string
	model       string
	temperature float64
	http        *http.Client
}

func New(baseURL, apiKey, model string, temperature float64) *Client {
	return &Client{
		baseURL:     baseURL,
		apiKey:      apiKey,
		model:       model,
		temperature: temperature,
		http:        &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Send(ctx context.Context, prompt string) (string, error) {
	messages := []chatMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	payload, err := json.Marshal(
		chatRequest{
			Model:       c.model,
			Messages:    messages,
			Temperature: c.temperature,
		},
	)

	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(payload),
	)

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	res, err := c.http.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	switch status := res.StatusCode; status {
	case http.StatusTooManyRequests:
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))

		return "", fmt.Errorf("%w: %s", ErrRateLimited, body)
	case http.StatusUnauthorized:
		return "", ErrUnauthorized
	case http.StatusPaymentRequired:
		return "", ErrNoCredit
	}

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))

		return "", fmt.Errorf("llm: status %d: %s", res.StatusCode, body)
	}

	var out chatResponse

	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}

	if out.Error != nil {
		return "", fmt.Errorf("llm: %s", out.Error.Message)
	}

	if len(out.Choices) == 0 {
		return "", ErrEmptyResponse
	}

	return out.Choices[0].Message.Content, nil
}
