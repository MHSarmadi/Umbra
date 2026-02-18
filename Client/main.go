//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/MHSarmadi/Umbra/Client/api"
	"github.com/MHSarmadi/Umbra/Client/models"
)

func main() {
	progressChan := make(chan models.ProgressReport, 1000)

	go func() {
		for report := range progressChan {
			jsProgressCallback := js.Global().Get("onProgressMade")
			if !jsProgressCallback.IsUndefined() {
				jsProgressCallback.Invoke(report.Type, report.ID, report.Percentage)
			}
		}
	}()

	api.UmbraReady()

	api.SessionKeyPair()

	api.IntroduceServer()

	api.ComputePoW(progressChan)

	api.CheckoutCaptcha()

	select {}
}
