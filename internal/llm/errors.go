package llm

import "errors"

var (
	ErrRateLimited   = errors.New("llm: rate limited")
	ErrUnauthorized  = errors.New("llm: unauthorized")
	ErrNoCredit      = errors.New("llm: insufficient credit")
	ErrEmptyResponse = errors.New("llm: no choices in response")
)
