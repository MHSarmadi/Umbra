package web

import (
	"context"

	"github.com/MHSarmadi/Umbra/Server/controllers"
	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/gorilla/mux"
)

func buildRouter(ctx context.Context, storage *database.BadgerStore) *mux.Router {
	r := mux.NewRouter()

	c := controllers.NewController(ctx, storage)

	// === DEMO ===
	r.HandleFunc("/demo/captcha", c.DemoCaptcha).Methods("GET")

	// === PING ===
	r.HandleFunc("/hello-world", c.HelloWorld).Methods("GET", "POST")

	// === SESSION ===
	r.HandleFunc("/session/init", c.SessionInit).Methods("POST")

	// === WS ===
	r.HandleFunc("/ws", c.WS).Methods("GET")

	return r
}
