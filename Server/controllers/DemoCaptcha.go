package controllers

import (
	"net/http"

	"github.com/MHSarmadi/Umbra/Server/captcha"
	math_tools "github.com/MHSarmadi/Umbra/Server/math"
)

func (c *Controller) DemoCaptcha(w http.ResponseWriter, r *http.Request) {
	// Demo number (hardcoded for now)
	number := math_tools.RandomDecimalString(6)

	pngBytes, err := captcha.GenerateNumericCaptcha(number)
	if err != nil {
		http.Error(w, "failed to generate captcha", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(pngBytes)
}
