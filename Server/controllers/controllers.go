package controllers

import (
	"context"

	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/olahol/melody"
)

type Controller struct {
	ctx     context.Context
	storage *database.BadgerStore
	ws      *melody.Melody
}

func NewController(ctx context.Context, storage *database.BadgerStore) *Controller {
	return &Controller{
		ctx:     ctx,
		storage: storage,
		ws:      melody.New(),
	}
}
