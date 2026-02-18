package web

import (
	"context"
	"net/http"
	"time"

	"github.com/MHSarmadi/Umbra/Server/database"
)

type Server struct {
	httpServer *http.Server
	storage    *database.BadgerStore
}

func NewServer(ctx context.Context, address string, storage *database.BadgerStore) *Server {
	r := buildRouter(ctx, storage)

	// Wrap the router with Logger and CORS middleware
	handler := LoggerMiddleware(corsMiddleware(r))

	srv := &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{httpServer: srv, storage: storage}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
