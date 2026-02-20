//go:build js && wasm
// +build js,wasm

package api

import (
	"encoding/binary"
	"syscall/js"

	"github.com/MHSarmadi/Umbra/Client/crypto"
	"github.com/MHSarmadi/Umbra/Client/tools"
	"golang.org/x/crypto/argon2"
)

const (
	sessTokCiphKeyMemoryMB    = 12
	sessTokCiphKeyParallelism = 1
	sessTokCiphKeyIterations  = 24
)

func CheckoutCaptcha() {
	js.Global().Set("CheckoutCaptcha", js.FuncOf(func(this js.Value, args []js.Value) any {
		// expected args: captcha_solution: number, session_token_ciphered: uint8array, session_token_cipher_key_salt: uint8array, session_id: uint8array
		if len(args) < 4 {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("At least 4 parameters are required: captcha_solution, session_token_ciphered, session_token_cipher_key_salt, session_id")
				return nil
			}))
		}

		captcha_solution := uint64(args[0].Int())
		captcha_solution_bytes := make([]byte, 8)
		binary.BigEndian.PutUint64(captcha_solution_bytes, captcha_solution)

		session_token_ciphered_pack, err := tools.JsValueToByteSlice(args[1])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid session_token_ciphered_pack:" + err.Error())
				return nil
			}))
		}
		session_token_cipher_key_salt, err := tools.JsValueToByteSlice(args[2])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid session_token_cipher_key_salt:" + err.Error())
				return nil
			}))
		}

		session_id, err := tools.JsValueToByteSlice(args[3])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid session_id:" + err.Error())
				return nil
			}))
		}

		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke("Panic occurred: " + r.(string))
					}
				}()

				session_token_cipher_key := argon2.IDKey(captcha_solution_bytes, session_token_cipher_key_salt, sessTokCiphKeyIterations, sessTokCiphKeyMemoryMB*1024, sessTokCiphKeyParallelism, 32)

				session_token_salt := session_token_ciphered_pack[:12]
				session_token_tag := session_token_ciphered_pack[12 : 12+16]
				session_token_ciphered := session_token_ciphered_pack[12+16:]

				session_token, valid, err := crypto.MACE_Decrypt_MIXIN_AEAD(session_token_cipher_key, session_token_ciphered, session_id, session_token_salt, session_token_tag, "@SESSION-TOKEN", 2)
				if !valid {
					reject.Invoke("Wrong captcha solution")
					return
				}
				if err != nil {
					reject.Invoke("Failed to decrypt session token: " + err.Error())
					return
				}
				if len(session_token) != 24 {
					reject.Invoke("Invalid session token length: expected 24 bytes")
					return
				}

				resolve.Invoke(b64(session_token))
			}()
			return nil
		}))
	}))
}
