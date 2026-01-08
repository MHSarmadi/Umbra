package controllers

import (
	"net/http"

	"github.com/gogearbox/gearbox"
)

type Response struct {
	Msg string `json:"msg"`
}

// func (c *Controller) HelloWorld(w http.ResponseWriter, r *http.Request) {
func (c *Controller) HelloWorld(ctx gearbox.Context) {
	ctx.Status(http.StatusOK)
	ctx.SendBytes([]byte(`{"msg": "Hello, World!"}`))
}
