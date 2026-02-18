//go:build js && wasm
// +build js,wasm

package tools

import (
	"fmt"
	"syscall/js"
)

func JsValueToByteSlice(v js.Value) ([]byte, error) {
	if v.IsNull() || v.IsUndefined() {
		return nil, fmt.Errorf("value is null or undefined")
	}

	// Usually people check typeof first, but for simplicity:
	if v.Get("byteLength").IsUndefined() {
		return nil, fmt.Errorf("not a valid typed array (byteLength missing)")
	}
	length := v.Get("byteLength").Int()
	if length < 0 {
		return nil, fmt.Errorf("not a valid typed array (byteLength missing or invalid)")
	}

	dst := make([]byte, length)

	n := js.CopyBytesToGo(dst, v)
	if n != length {
		return nil, fmt.Errorf("only copied %d of %d bytes", n, length)
	}

	return dst, nil
}
