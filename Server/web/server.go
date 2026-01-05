package web

import (
	"context"
	"net/http"
	"time"

	"github.com/MHSarmadi/Umbra/Server/controllers"
	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/gorilla/mux"
)

type Server struct {
	httpServer *http.Server
	storage *database.BadgerStore
}

func NewServer(ctx context.Context, address string, storage *database.BadgerStore) *Server {
	r := mux.NewRouter()
	
	c := controllers.NewController(ctx, storage)

	r.HandleFunc("/hello-world", c.HelloWorld).Methods("GET", "POST")
	r.HandleFunc("/demo/captcha", c.DemoCaptcha).Methods("GET")
	r.HandleFunc("/session/init", c.SessionInit).Methods("POST")

	srv := &http.Server{
		Addr: address,
		Handler: r,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	return &Server{httpServer: srv, storage: storage}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}