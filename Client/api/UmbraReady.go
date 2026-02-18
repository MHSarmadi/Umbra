//go:build js && wasm
// +build js,wasm

package api

import "syscall/js"

func UmbraReady() {
	js.Global().Set("umbraReady", js.FuncOf(func(this js.Value, args []js.Value) any {
		return "Umbra WASM initialized"
	}))
}
