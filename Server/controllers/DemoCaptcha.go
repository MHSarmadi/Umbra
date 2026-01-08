package controllers

import (
	"net/http"

	"github.com/gogearbox/gearbox"

	"github.com/MHSarmadi/Umbra/Server/captcha"
	math_tools "github.com/MHSarmadi/Umbra/Server/math"
)

func (c *Controller) DemoCaptcha(ctx gearbox.Context) {
	// Demo number (hardcoded for now)
	number := math_tools.RandomDecimalString(6)

	pngBytes, err := captcha.GenerateNumericCaptcha(number)
	if err != nil {
		ctx.Status(gearbox.StatusInternalServerError).SendString("failed to generate captcha")
		return
	}

	ctx.Set("Content-Type", "image/png")
	ctx.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	ctx.Set("Pragma", "no-cache")
	ctx.Set("Expires", "0")
	ctx.Status(http.StatusOK)

	_ = ctx.SendBytes(pngBytes)
}
