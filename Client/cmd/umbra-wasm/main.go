//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/MHSarmadi/Umbra/Client/crypto"
)

func jsValueToByteSlice(v js.Value) ([]byte, error) {
	if v.IsNull() || v.IsUndefined() {
		return nil, fmt.Errorf("value is null or undefined")
	}

	// Usually people check typeof first, but for simplicity:
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

func main() {
	js.Global().Set("umbraReady", js.FuncOf(func(this js.Value, args []js.Value) any {
		return "Umbra WASM initialized"
	}))

	js.Global().Set("MACE_Encrypt", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 4 {
			err := js.Error{
				Value: js.ValueOf("At least 4 parameters"),
			}
			js.Global().Get("console").Call("error", err)
			return err
		}

		key, err := jsValueToByteSlice(args[0])
		if err != nil {
			js.Global().Get("console").Call("error", err.Error())
			return err.Error()
		}

		data, err := jsValueToByteSlice(args[1])
		if err != nil {
			js.Global().Get("console").Call("error", err.Error())
			return err.Error()
		}

		context := args[2].String()

		difficulty := args[3].Int()

		cipher, salt := crypto.MACE_Encrypt(key, data, context, uint16(difficulty), true)
		jsCipher := js.Global().Get("Uint8Array").New(len(cipher))
		jsSalt := js.Global().Get("Uint8Array").New(len(salt))

		js.CopyBytesToJS(jsCipher, cipher)
		js.CopyBytesToJS(jsSalt, salt)

		result := js.Global().Get("Object").New()
		result.Set("cipher", jsCipher)
		result.Set("salt", jsSalt)

		return result
	}))

	select {}
}
