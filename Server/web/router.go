package web

import (
	"context"
	"net/http"

	"github.com/MHSarmadi/Umbra/Server/controllers"
	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/gorilla/mux"
)

func buildRouter(ctx context.Context, storage *database.BadgerStore) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(mux.CORSMethodMiddleware(r))
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	c := controllers.NewController(ctx, storage)

	demo := r.PathPrefix("/demo").Subrouter()
	demo.HandleFunc("/captcha", c.DemoCaptcha).Methods(http.MethodGet)

	r.HandleFunc("/hello-world", c.HelloWorld).Methods(http.MethodGet, http.MethodPost)

	session := r.PathPrefix("/session").Subrouter()
	session.HandleFunc("/init", c.SessionInit).Methods(http.MethodPost)

	r.HandleFunc("/ws", c.WS).Methods(http.MethodGet)

	return r
}
