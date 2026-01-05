package controllers

import (
	"context"
	"net/http"

	"github.com/MHSarmadi/Umbra/Server/database"
)

type Controller struct {
	ctx context.Context
	storage *database.BadgerStore
}

func NewController(ctx context.Context, storage *database.BadgerStore) *Controller {
	return &Controller{
		ctx: ctx,
		storage: storage,
	}
}

func BadRequest(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{"status":"error","message":"` + message + `"}`))
}

func InternalServerError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"error","message":"` + message + `"}`))
}
