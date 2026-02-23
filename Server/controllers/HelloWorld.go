package controllers

import (
	"encoding/json"
	"net/http"
)

func (c *Controller) HelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"msg": "Hello, World!"})
}
