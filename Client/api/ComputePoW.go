//go:build js && wasm
// +build js,wasm

package api

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"syscall/js"

	"github.com/MHSarmadi/Umbra/Client/models"
	"github.com/MHSarmadi/Umbra/Client/tools"
	"golang.org/x/crypto/argon2"
)

const targetPoWFailProbability = 0.0001 // 0.01%

func ComputePoW(progressChan chan models.ProgressReport) {
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
		challenge, err := tools.JsValueToByteSlice(args[1])
		if err != nil {
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
				reject := promArgs[1]
				reject.Invoke("Invalid challenge: " + err.Error())
				return nil
			}))
		}

		salt, err := tools.JsValueToByteSlice(args[2])
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
				case progressChan <- models.ProgressReport{
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
						case progressChan <- models.ProgressReport{
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
}
