package controllers

import (
	"context"

	"github.com/MHSarmadi/Umbra/Server/database"
)

type Controller struct {
	ctx     context.Context
	storage *database.BadgerStore
}

func NewController(ctx context.Context, storage *database.BadgerStore) *Controller {
	return &Controller{
		ctx:     ctx,
		storage: storage,
	}
}
