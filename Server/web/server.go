package web

import (
	"context"
	"net/http"
	"time"

	"github.com/MHSarmadi/Umbra/Server/controllers"
	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/MHSarmadi/Umbra/Server/logger"
	"github.com/gorilla/mux"
)

type Server struct {
	httpServer *http.Server
	storage    *database.BadgerStore
}

// CORS Middleware - wraps the router
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Your frontend URL
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Pass to the next handler
		next.ServeHTTP(w, r)
	})
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.Debugf("%s %s from [%s] responded in %d microseconds", r.Method, r.RequestURI, r.RemoteAddr, duration.Microseconds())
	})
}

func NewServer(ctx context.Context, address string, storage *database.BadgerStore) *Server {
	r := mux.NewRouter()

	c := controllers.NewController(ctx, storage)

	r.HandleFunc("/hello-world", c.HelloWorld).Methods("GET", "POST")
	r.HandleFunc("/demo/captcha", c.DemoCaptcha).Methods("GET")
	r.HandleFunc("/session/init", c.SessionInit).Methods("POST")

	// Wrap the router with CORS middleware
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
