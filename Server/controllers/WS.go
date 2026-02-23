package controllers

import (
	"net/http"
)

func (c *Controller) WS(w http.ResponseWriter, r *http.Request) {
	if c.ws == nil {
		http.Error(w, "websocket server unavailable", http.StatusInternalServerError)
		return
	}
	if err := c.ws.HandleRequest(w, r); err != nil {
		http.Error(w, "websocket handshake failed", http.StatusBadRequest)
	}
}
