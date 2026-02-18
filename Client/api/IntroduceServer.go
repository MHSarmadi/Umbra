//go:build js && wasm
// +build js,wasm

package api

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/MHSarmadi/Umbra/Client/crypto"
	"github.com/MHSarmadi/Umbra/Client/tools"
)

func IntroduceServer() {
	js.Global().Set("IntroduceServer", js.FuncOf(func(this js.Value, args []js.Value) any {
		// expected args: soul, server_ed_pubkey: base64, server_x_pubkey: base64, server_x_pubkey_sign: base64, payload: base64, signature: base64
		if len(args) < 6 {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("At least 6 parameters are required: soul, server_ed_pubkey, server_x_pubkey, server_x_pubkey_sign, payload, signature")
				return nil
			}))
		}
		soul, err := tools.JsValueToByteSlice(args[0])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid soul: " + err.Error())
				return nil
			}))
		}
		server_ed_pubkey, err := db64(args[1].String())
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid server_ed_pubkey base64 encoding: " + err.Error())
				return nil
			}))
		}
		server_x_pubkey, err := db64(args[2].String())
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid server_x_pubkey base64 encoding: " + err.Error())
				return nil
			}))
		}
		server_x_pubkey_sign, err := db64(args[3].String())
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid server_x_pubkey_sign base64 encoding: " + err.Error())
				return nil
			}))
		}
		payload, err := db64(args[4].String())
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid payload base64 encoding: " + err.Error())
				return nil
			}))
		}
		signature, err := db64(args[5].String())
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid signature base64 encoding: " + err.Error())
				return nil
			}))
		}
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(fmt.Sprintf("Panic occurred: %v", r))
					}
				}()

				// 1. verifying signatures
				if !crypto.Verify(server_ed_pubkey, server_x_pubkey, server_x_pubkey_sign) {
					reject.Invoke("Invalid server session keys")
					return
				}

				if !crypto.Verify(server_ed_pubkey, payload, signature) {
					reject.Invoke("Invalid signature over payload")
				}

				// 2. derive shared secret and shared key
				shared_secret, err := crypto.ComputeSharedSecret(soul, server_x_pubkey)
				if err != nil {
					reject.Invoke("Failed to compute shared secret: " + err.Error())
					return
				}
				shared_key := crypto.KDF(shared_secret, "@SESSION-SHARED-KEY", 32)

				// 3. decipher payload
				payload_salt := payload[:12]
				payload_tag := payload[12 : 12+16]
				payload_ciphered := payload[12+16:]
				now := time.Now()
				payload_deciphered, valid, err := crypto.MACE_Decrypt_AEAD(shared_key, payload_ciphered, payload_salt, payload_tag, "@RESPONSE-PAYLOAD", 8)
				duration := time.Since(now)
				if !valid {
					reject.Invoke("AEAD failed during deciphering payload.")
					return
				}
				if err != nil {
					reject.Invoke("Failed to decipher payload.")
					return
				}

				// 4. parse JSON
				type SessionInitRawPayload struct {
					CaptchaChallenge string         `json:"captcha_challenge"`
					PoWChallenge     string         `json:"pow_challenge"`
					PowParams        map[string]any `json:"pow_params"`
					PoWSalt          string         `json:"pow_salt"`
					SessionToken     string         `json:"session_token_ciphered"`
				}
				var payloadData SessionInitRawPayload
				if err := json.Unmarshal(payload_deciphered, &payloadData); err != nil {
					reject.Invoke("Failed to parse payload JSON: " + err.Error())
					return
				}

				// 5. resolve results
				result := js.Global().Get("Object").New()
				result.Set("captcha_challenge", payloadData.CaptchaChallenge)
				result.Set("pow_challenge", payloadData.PoWChallenge)
				result.Set("pow_params", payloadData.PowParams)
				result.Set("pow_salt", payloadData.PoWSalt)
				result.Set("session_token_ciphered", payloadData.SessionToken)
				result.Set("took_microseconds", duration.Microseconds())

				resolve.Invoke(result)
			}()
			return nil
		}))
	}))
}