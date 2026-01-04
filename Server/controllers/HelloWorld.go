package controllers

import "net/http"

func (c *Controller) HelloWorld(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"msg":"Hello, World!"}`))
}