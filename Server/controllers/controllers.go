package controllers

import "github.com/MHSarmadi/Umbra/Server/database"

type Controller struct {
	storage *database.BadgerStore
}

func NewController(storage *database.BadgerStore) *Controller {
	return &Controller{
		storage: storage,
	}
}