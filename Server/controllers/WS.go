package controllers

import (
	"net/http"

	"github.com/olahol/melody"
)

func (c *Controller) WS(w http.ResponseWriter, r *http.Request) {
	m := melody.New()
	m.HandleRequest(w, r)
}
