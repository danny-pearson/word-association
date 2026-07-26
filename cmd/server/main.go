package main

import (
	"net/http"

	"github.com/danny-pearson/word-association/internal/api"
)

func main() {
	handler := api.NewHandler()
	mux := http.NewServeMux()

	api.RegisterRoutes(mux, handler)

	server := api.NewServer(mux)

	server.Start()
}