package api

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	http *http.Server
}

func NewServer(handler http.Handler) *Server {
	return &Server{
		http: &http.Server{
			Addr:         ":8080",
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
