package api

import (
	"encoding/json"
	"net/http"
)

type Handler struct{}

type TestResponse struct {
	Message string
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h Handler) Test(w http.ResponseWriter, r *http.Request) {
	message := TestResponse{"Working"}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(message)
}
