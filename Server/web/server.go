package web

import (
	"context"
	"net/http"
	"time"

	"github.com/MHSarmadi/Umbra/Server/controllers"
	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/MHSarmadi/Umbra/Server/router"
	"github.com/gogearbox/gearbox"
)

type Server struct {
	app        gearbox.Gearbox
	httpServer *http.Server
	storage    *database.BadgerStore
}

func NewServer(ctx context.Context, address string, storage *database.BadgerStore) *Server {
	app := gearbox.New()

	c := controllers.NewController(ctx, storage)

	// register routes using our router package
	router.SetupRoutes(app, c)

	srv := &http.Server{
		Addr: address,
		// Handler:      app.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{app: app, httpServer: srv, storage: storage}
}

func (s *Server) Run() error {
	// return s.httpServer.ListenAndServe()
	return s.app.Start(s.httpServer.Addr)
}

func (s *Server) ShutDown(ctx context.Context) error {
	// return s.httpServer.Shutdown(ctx)
	return s.app.Stop()
}
