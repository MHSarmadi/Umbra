//go:build js && wasm
// +build js,wasm

package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"syscall/js"

	"github.com/MHSarmadi/Umbra/Client/crypto"
)

var (
	b64  = base64.RawStdEncoding.EncodeToString
	db64 = base64.RawStdEncoding.DecodeString
)

func SessionKeyPair() {
	js.Global().Set("SessionKeypair", js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(fmt.Sprintf("Panic occurred: %v", r))
					}
				}()

				// Generate soul
				var soul [32]byte
				if _, err := rand.Read(soul[:]); err != nil {
					reject.Invoke("Could not read entropy for soul generation")
					return
				}

				// Derive keys
				edPubKey := crypto.DeriveEd25519PubKey(soul[:])
				xPubKey, err := crypto.DeriveX25519PubKey(soul[:])
				if err != nil {
					reject.Invoke("Could not derive X25519 public key: " + err.Error())
					return
				}
				xPubKeySign := crypto.Sign(soul[:], xPubKey)

				// Prepare response
				response := js.Global().Get("Object").New()
				response.Set("ed_pubkey", b64(edPubKey))
				response.Set("x_pubkey", b64(xPubKey))
				response.Set("x_pubkey_sign", b64(xPubKeySign))
				response.Set("soul", js.Global().Get("Uint8Array").New(32))
				js.CopyBytesToJS(response.Get("soul"), soul[:])

				resolve.Invoke(response)
			}()

			return nil
		}))
	}))
}
