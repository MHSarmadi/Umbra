//go:build js && wasm
// +build js,wasm

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"slices"
	"syscall/js"

	"github.com/MHSarmadi/Umbra/Client/crypto"
	"golang.org/x/crypto/argon2"
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

type ProgressReport struct {
	Type       string
	ID         string
	Percentage float64
}

func main() {
	progressChan := make(chan ProgressReport, 1000)

	go func() {
		for report := range progressChan {
			jsProgressCallback := js.Global().Get("onProgressMade")
			if !jsProgressCallback.IsUndefined() {
				jsProgressCallback.Invoke(report.Type, report.ID, report.Percentage)
			}
		}
	}()

	js.Global().Set("umbraReady", js.FuncOf(func(this js.Value, args []js.Value) any {
		return "Umbra WASM initialized"
	}))

	js.Global().Set("SessionKeypair", js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(js.Error{Value: js.ValueOf(fmt.Sprintf("Panic occurred: %v", r))})
					}
				}()

				// Generate soul
				var soul [32]byte
				if _, err := rand.Read(soul[:]); err != nil {
					reject.Invoke(js.Error{Value: js.ValueOf("Could not read entropy for soul generation")})
					return
				}

				// Derive keys
				edPubKey := crypto.DeriveEd25519PubKey(soul[:])
				xPubKey, err := crypto.DeriveX25519PubKey(soul[:])
				if err != nil {
					reject.Invoke(js.Error{Value: js.ValueOf("Could not derive X25519 public key: " + err.Error())})
					return
				}
				xPubKeySign := crypto.Sign(soul[:], xPubKey)

				// Prepare response
				response := js.Global().Get("Object").New()
				response.Set("ed_pubkey", base64.RawURLEncoding.EncodeToString(edPubKey))
				response.Set("x_pubkey", base64.RawURLEncoding.EncodeToString(xPubKey))
				response.Set("x_pubkey_sign", base64.RawURLEncoding.EncodeToString(xPubKeySign))
				response.Set("soul", js.Global().Get("Uint8Array").New(32))
				js.CopyBytesToJS(response.Get("soul"), soul[:])

				resolve.Invoke(response)
			}()

			return nil
		}))
	}))

	js.Global().Set("ComputePoW", js.FuncOf(func(this js.Value, args []js.Value) any {
		// expected args: progress_id, challenge, salt, memory_mb, iterations, parallelism
		// return: Promise<error|number>
		if len(args) < 6 {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke(js.Error{Value: js.ValueOf("At least 6 parameters are required: progress_id, challenge, salt, memory_mb, iterations, parallelism")})
				return nil
			}))
		}

		progressID := args[0].String()
		challenge, err := jsValueToByteSlice(args[1])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke(js.Error{Value: js.ValueOf("Invalid challenge: " + err.Error())})
				return nil
			}))
		}

		salt, err := jsValueToByteSlice(args[2])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke(js.Error{Value: js.ValueOf("Invalid salt: " + err.Error())})
				return nil
			}))
		}

		memoryMB := args[3].Int()
		iterations := args[4].Int()
		parallelism := args[5].Int()

		if memoryMB <= 0 || iterations <= 0 || parallelism <= 0 {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke(js.Error{Value: js.ValueOf("Memory, iterations, and parallelism must be greater than 0")})
				return nil
			}))
		}

		theoretical_max_attempts := 1 << (8 * len(challenge))
		attempts_per_report := theoretical_max_attempts / 100

		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(js.Error{Value: js.ValueOf(fmt.Sprintf("Panic occurred: %v", r))})
					}
				}()

				// initial report
				select {
				case progressChan <- ProgressReport{
					Type:       "pow",
					ID:         progressID,
					Percentage: 0,
				}:
				default:
				}

				// brute-force the PoW challenge
				var nonce uint64 = 0
				var nonce_bytes [8]byte = [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
				attempt := 0
				for attempt < 2*theoretical_max_attempts {
					attempt++
					if attempt%attempts_per_report == 0 {
						select {
						case progressChan <- ProgressReport{
							Type:       "pow",
							ID:         progressID,
							Percentage: float64(attempt) / float64(theoretical_max_attempts) * 100,
						}:
						default:
						}
					}

					// compute the hash
					hash := argon2.IDKey(nonce_bytes[:], salt, uint32(iterations), uint32(memoryMB*1024), uint8(parallelism), 32)
					if slices.Equal(hash[:len(challenge)], challenge) {
						resolve.Invoke(nonce)
						return
					} else {
						nonce++
						binary.BigEndian.PutUint64(nonce_bytes[:], nonce)
					}
				}
				reject.Invoke(js.Error{Value: js.ValueOf("No valid nonce found after the theoretical maximum twice attempts")})
			}()

			return nil
		}))

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
