//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

func main() {
	js.Global().Set("umbraReady", js.FuncOf(func(this js.Value, args []js.Value) any {
		return "Umbra WASM initialized"
	}))

	select {}
}
