package api

import (
	"encoding/json"
	"net/http"
)

type Handler struct {}

type TestResponse struct {
    Message string
}

func NewHandler() *Handler {
	return &Handler{}
}

func (handler Handler) Test(writer http.ResponseWriter, req *http.Request) {
	message := TestResponse{ "Working" }

	writer.Header().Set("Content-Type", "application/json")

	json.NewEncoder(writer).Encode(message)
}