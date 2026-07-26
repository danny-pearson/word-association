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
			Addr: ":8080",
			Handler: handler,
			ReadTimeout: 5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

func (server *Server) Start() error {
	return server.http.ListenAndServe();
}

func (server *Server) Shutdown(ctx context.Context) error {
	return server.http.Shutdown(ctx)
}
