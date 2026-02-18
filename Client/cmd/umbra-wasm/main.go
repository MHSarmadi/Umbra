//go:build js && wasm
// +build js,wasm

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"syscall/js"
	"time"

	"github.com/MHSarmadi/Umbra/Client/crypto"
	"golang.org/x/crypto/argon2"
)

var (
	b64  = base64.RawStdEncoding.EncodeToString
	db64 = base64.RawStdEncoding.DecodeString
)

func jsValueToByteSlice(v js.Value) ([]byte, error) {
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

type ProgressReport struct {
	Type       string
	ID         string
	Percentage float64
}

const targetPoWFailProbability = 0.0001 // 0.01%

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

	js.Global().Set("IntroduceServer", js.FuncOf(func(this js.Value, args []js.Value) any {
		// expected args: soul, server_ed_pubkey: base64, server_x_pubkey: base64, server_x_pubkey_sign: base64, payload: base64, signature: base64
		if len(args) < 6 {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("At least 6 parameters are required: soul, server_ed_pubkey, server_x_pubkey, server_x_pubkey_sign, payload, signature")
				return nil
			}))
		}
		soul, err := jsValueToByteSlice(args[0])
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

	js.Global().Set("ComputePoW", js.FuncOf(func(this js.Value, args []js.Value) any {
		// expected args: progress_id, challenge, salt, memory_mb, iterations, parallelism
		// return: Promise<error|number>
		if len(args) < 6 {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("At least 6 parameters are required: progress_id, challenge, salt, memory_mb, iterations, parallelism")
				return nil
			}))
		}

		progressID := args[0].String()
		challenge, err := jsValueToByteSlice(args[1])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid challenge: " + err.Error())
				return nil
			}))
		}

		salt, err := jsValueToByteSlice(args[2])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid salt: " + err.Error())
				return nil
			}))
		}

		memoryMB := args[3].Int()
		iterations := args[4].Int()
		parallelism := args[5].Int()

		if memoryMB <= 0 || iterations <= 0 || parallelism <= 0 {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Memory, iterations, and parallelism must be greater than 0")
				return nil
			}))
		}

		// Per-attempt success probability for prefix matching:
		// p = 1 / 2^(8*len(challenge)).
		perAttemptSuccessProb := math.Exp2(-8.0 * float64(len(challenge)))
		if perAttemptSuccessProb <= 0 || perAttemptSuccessProb >= 1 {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid PoW challenge length for probabilistic solver")
				return nil
			}))
		}

		// Solve N from: (1-p)^N <= targetPoWFailProbability.
		maxAttemptsFloat := math.Ceil(math.Log(targetPoWFailProbability) / math.Log1p(-perAttemptSuccessProb))
		if maxAttemptsFloat <= 0 || maxAttemptsFloat > float64(^uint64(0)) {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("PoW max attempts overflowed for this challenge length")
				return nil
			}))
		}
		maxAttempts := uint64(maxAttemptsFloat)
		attemptsPerReport := maxAttempts/1000 + 1
		targetSuccessProb := 1.0 - targetPoWFailProbability

		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(fmt.Sprintf("Panic occurred: %v", r))
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
				var attempt uint64 = 0
				for attempt < maxAttempts {
					attempt++
					if attempt%attemptsPerReport == 0 {
						successProb := 1.0 - math.Pow(1.0-perAttemptSuccessProb, float64(attempt))
						percentage := 100.0 * successProb / targetSuccessProb
						if percentage > 100.0 {
							percentage = 100.0
						}
						select {
						case progressChan <- ProgressReport{
							Type:       "pow",
							ID:         progressID,
							Percentage: percentage,
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
				reject.Invoke(fmt.Sprintf(
					"No valid nonce found after %d attempts (target fail probability %.5f%%)",
					maxAttempts,
					targetPoWFailProbability*100.0,
				))
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
