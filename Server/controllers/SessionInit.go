package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MHSarmadi/Umbra/Server/captcha"
	"github.com/MHSarmadi/Umbra/Server/crypto"
	"github.com/MHSarmadi/Umbra/Server/logger"
	math_tools "github.com/MHSarmadi/Umbra/Server/math"
	"github.com/MHSarmadi/Umbra/Server/models"
	models_requests "github.com/MHSarmadi/Umbra/Server/models/requests"
	"golang.org/x/crypto/argon2"
)

const Expiry_Offset = 300 * time.Second

const (
	sessionInitWindow             = 10 * time.Minute
	sessionInitMaxRequestsPerWind = 32
	sessionInitTrackerTTL         = 30 * time.Minute
	trustForwardedIdentityHeaders = false

	powChallengeSize = 1
	powMemoryMB      = 12
	powParallelism   = 1
	powIterationsMin = 2
	powIterationsMax = 7

	sessTokCiphKeyMemoryMB    = 12
	sessTokCiphKeyParallelism = 1
	sessTokCiphKeyIterations  = 24
)

var (
	b64  = base64.RawStdEncoding.EncodeToString
	db64 = base64.RawStdEncoding.DecodeString
)

func sessionInitIdentityHash(r *http.Request) string {
	identityRaw := clientIP(r)
	sum := crypto.Sum([]byte(identityRaw))
	return b64(sum[:16])
}

func clientIP(r *http.Request) string {
	if trustForwardedIdentityHeaders {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			first := strings.TrimSpace(parts[0])
			if first != "" {
				return first
			}
		}
		if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
			return xrip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func dynamicPoWIterations(requestCount int) uint {
	if requestCount < 1 {
		requestCount = 1
	}
	density := float64(requestCount) / float64(sessionInitMaxRequestsPerWind)
	if density > 1 {
		density = 1
	}

	// Logistic curve normalized to [0,1] over density range [0,1].
	const k float64 = 10.0
	const mid float64 = 0.55
	raw := 1.0 / (1.0 + math.Exp(-k*(density-mid)))
	lo := 1.0 / (1.0 + math.Exp(-k*(0-mid)))
	hi := 1.0 / (1.0 + math.Exp(-k*(1-mid)))
	normalized := (raw - lo) / (hi - lo)

	span := float64(powIterationsMax - powIterationsMin)
	it := float64(powIterationsMin) + normalized*span
	rounded := uint(math.Round(it))
	if rounded < powIterationsMin {
		return powIterationsMin
	}
	if rounded > powIterationsMax {
		return powIterationsMax
	}
	return rounded
}

func (c *Controller) SessionInit(w http.ResponseWriter, r *http.Request) {
	reqStart := time.Now()
	logger.Verbosef("session init started method=%s path=%s remote=%s", r.Method, r.URL.Path, r.RemoteAddr)

	var (
		err          error
		body_encoded models_requests.SessionInitRequestEncoded
		body_decoded models_requests.SessionInitRequestDecoded
	)
	if err := json.NewDecoder(r.Body).Decode(&body_encoded); err != nil {
		logger.Debugf("session init rejected: malformed json body remote=%s err=%v", r.RemoteAddr, err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	logger.Tracef(
		"session init payload field lengths (base64 chars): ed=%d x=%d sign=%d",
		len(body_encoded.ClientEdPubKey),
		len(body_encoded.ClientXPubKey),
		len(body_encoded.ClientXPubKeySignature),
	)
	if body_decoded.ClientEdPubKey, err = db64(body_encoded.ClientEdPubKey); err != nil {
		logger.Debugf("session init rejected: invalid client_ed_pubkey encoding err=%v", err)
		http.Error(w, "invalid client_ed_pubkey base64 encoding", http.StatusBadRequest)
		return
	} else if body_decoded.ClientXPubKey, err = db64(body_encoded.ClientXPubKey); err != nil {
		logger.Debugf("session init rejected: invalid client_x_pubkey encoding err=%v", err)
		http.Error(w, "invalid client_x_pubkey base64 encoding", http.StatusBadRequest)
		return
	} else if body_decoded.ClientXPubKeySignature, err = db64(body_encoded.ClientXPubKeySignature); err != nil {
		logger.Debugf("session init rejected: invalid client_x_pubkey_sign encoding err=%v", err)
		http.Error(w, "invalid client_x_pubkey_sign base64 encoding", http.StatusBadRequest)
		return
	} else if len(body_decoded.ClientEdPubKey) != 32 || len(body_decoded.ClientXPubKey) != 32 {
		logger.Debugf("session init rejected: invalid pubkey lengths ed=%d x=%d", len(body_decoded.ClientEdPubKey), len(body_decoded.ClientXPubKey))
		http.Error(w, "invalid ed-pubkey or x-pubkey length", http.StatusBadRequest)
		return
	} else if crypto.Verify(body_decoded.ClientEdPubKey, body_decoded.ClientXPubKey, body_decoded.ClientXPubKeySignature) == false {
		logger.Debugf("session init rejected: client signature verification failed")
		http.Error(w, "invalid signature over client_x_pubkey", http.StatusBadRequest)
		return
	} else {
		logger.Tracef("session init: client cryptographic identity verified")
		now := time.Now().UTC()
		trackerID := sessionInitIdentityHash(r)
		requestCount, limited, retryAfter, err := c.storage.RegisterSessionInitRequest(
			c.ctx,
			trackerID,
			now,
			sessionInitWindow,
			sessionInitMaxRequestsPerWind,
			sessionInitTrackerTTL,
		)
		if err != nil {
			logger.Errorf("session init tracker update failed for identity=%s: %v", trackerID, err)
			http.Error(w, "could not update session-init tracker", http.StatusInternalServerError)
			return
		}
		if limited {
			logger.Infof("session init rate-limited for identity=%s retry_after=%ds", trackerID, int64(retryAfter.Seconds()))
			w.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
			http.Error(w, "too many session initialization requests", http.StatusTooManyRequests)
			return
		}
		powIterations := dynamicPoWIterations(requestCount)
		logger.Debugf("session init identity=%s request_count=%d pow_iterations=%d", trackerID, requestCount, powIterations)

		if len(body_decoded.ClientEdPubKey) != 32 || len(body_decoded.ClientXPubKey) != 32 {
			logger.Errorf("session init internal invariant failed: pubkey lengths changed ed=%d x=%d", len(body_decoded.ClientEdPubKey), len(body_decoded.ClientXPubKey))
			http.Error(w, "invalid ed-pubkey or x-pubkey", http.StatusBadRequest)
			return
		}
		var session_id [24]byte
		if _, err := rand.Read(session_id[:]); err != nil {
			logger.Errorf("session init entropy read failed for session id: %v", err)
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}

		var server_soul [32]byte
		if _, err := rand.Read(server_soul[:]); err != nil {
			logger.Errorf("session init entropy read failed for server soul: %v", err)
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}

		deriveStart := time.Now()
		server_ed_pubkey := crypto.DeriveEd25519PubKey(server_soul[:])
		server_x_pubkey, err := crypto.DeriveX25519PubKey(server_soul[:])
		if err != nil {
			logger.Errorf("session init failed deriving server x25519 pubkey: %v", err)
			http.Error(w, "could not derive pubkey", http.StatusInternalServerError)
			return
		}
		server_x_pubkey_sign := crypto.Sign(server_soul[:], server_x_pubkey)
		logger.Tracef("session init: server key derivation/sign complete in %d microseconds", time.Since(deriveStart).Microseconds())

		var pow_challenge [powChallengeSize]byte
		if _, err := rand.Read(pow_challenge[:]); err != nil {
			logger.Errorf("session init entropy read failed for pow challenge: %v", err)
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}
		pow_params := models.PowParamsType{
			MemoryMB:    powMemoryMB,
			Iterations:  powIterations,
			Parallelism: powParallelism,
		}
		var pow_salt [12]byte
		if _, err := rand.Read(pow_salt[:]); err != nil {
			logger.Errorf("session init entropy read failed for pow salt: %v", err)
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}
		logger.Verbosef("session init pow params memory_mb=%d iterations=%d parallelism=%d challenge_bytes=%d salt_bytes=%d", pow_params.MemoryMB, pow_params.Iterations, pow_params.Parallelism, len(pow_challenge), len(pow_salt))

		captcha_solution := math_tools.RandomDecimalString(6)
		var captcha_solution_numeric uint64 = 0
		for _, c := range captcha_solution {
			captcha_solution_numeric *= 10
			captcha_solution_numeric += uint64(c - '0')
		}
		captcha_solution_bytes := make([]byte, 8)
		binary.BigEndian.PutUint64(captcha_solution_bytes, captcha_solution_numeric)
		captcha_png, err := captcha.GenerateNumericCaptcha(captcha_solution)
		if err != nil {
			logger.Errorf("session init captcha generation failed: %v", err)
			http.Error(w, "could not draw captcha", http.StatusInternalServerError)
			return
		}
		logger.Tracef("session init captcha generated bytes=%d", len(captcha_png))

		session_token_cipher_key_salt := make([]byte, 12)
		if _, err := rand.Read(session_token_cipher_key_salt[:]); err != nil {
			logger.Errorf("session init entropy read failed for session token cipher key salt: %v", err)
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}
		session_token_cipher_key := argon2.IDKey(captcha_solution_bytes, session_token_cipher_key_salt, sessTokCiphKeyIterations, sessTokCiphKeyMemoryMB*1024, sessTokCiphKeyParallelism, 32)

		var session_token [24]byte
		if _, err := rand.Read(session_token[:]); err != nil {
			logger.Errorf("session init entropy read failed for session token: %v", err)
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}
		session_token_ciphered, session_token_salt, session_token_tag := crypto.MACE_Encrypt_MIXIN_AEAD(session_token_cipher_key, session_token[:], session_id[:], "@SESSION-TOKEN", 2, false)
		logger.Tracef("session init session-token encryption produced cipher_bytes=%d salt_bytes=%d tag_bytes=%d", len(session_token_ciphered), len(session_token_salt), len(session_token_tag))

		session_token_ciphered_pack := append(session_token_salt, session_token_tag...)
		session_token_ciphered_pack = append(session_token_ciphered_pack, session_token_ciphered...) // session_token_salt is always exactly 12 bytes and session_token_tag is always exactly 16 bytes

		session := models.Session{
			UUID: session_id,

			CreatedAt: now,
			ExpiresAt: now.Add(Expiry_Offset).UTC(),

			ClientEdPubKey: [32]byte(body_decoded.ClientEdPubKey),
			ClientXPubKey:  [32]byte(body_decoded.ClientXPubKey),

			ServerSoul: server_soul,

			SessionToken:              session_token,
			SessionTokenCipherKeySalt: [12]byte(session_token_cipher_key_salt),

			PoWChallenge: pow_challenge,
			PoWParams:    pow_params,
			PoWSalt:      pow_salt,
		}

		if err := c.storage.PutSession(c.ctx, &session); err != nil {
			logger.Errorf("session init failed persisting session identity=%s: %v", trackerID, err)
			http.Error(w, "could not store seesion", http.StatusInternalServerError)
			return
		}
		logger.Tracef("session init session persisted session_id_b64_len=%d", len(b64(session.UUID[:])))

		type SessionInitResponse struct {
			Status                 string `json:"status"`
			ServerEdPubKey         string `json:"server_ed_pubkey"`
			ServerXPubKey          string `json:"server_x_pubkey"`
			ServerXPubKeySignature string `json:"server_x_pubkey_sign"`
			Payload                string `json:"payload"`
			Signature              string `json:"signature"`
		}
		type SessionInitRawPayload struct {
			SessionUUID               string               `json:"session_id"`
			CaptchaChallenge          string               `json:"captcha_challenge"`
			PoWChallenge              string               `json:"pow_challenge"`
			PowParams                 models.PowParamsType `json:"pow_params"`
			PoWSalt                   string               `json:"pow_salt"`
			SessionToken              string               `json:"session_token_ciphered"`
			SessionTokenCipherKeySalt string               `json:"session_token_cipher_key_salt"`
		}

		payload_raw := SessionInitRawPayload{
			SessionUUID:               b64(session.UUID[:]),
			PoWChallenge:              b64(session.PoWChallenge[:]),
			PowParams:                 session.PoWParams,
			PoWSalt:                   b64(session.PoWSalt[:]),
			SessionToken:              b64(session_token_ciphered_pack),
			SessionTokenCipherKeySalt: b64(session_token_cipher_key_salt),
			CaptchaChallenge:          b64(captcha_png),
		}
		payload_encoded, err := json.Marshal(payload_raw)
		if err != nil {
			logger.Errorf("session init failed marshaling payload json: %v", err)
			http.Error(w, "could not marshal to json", http.StatusInternalServerError)
			return
		}
		logger.Tracef("session init payload marshaled bytes=%d", len(payload_encoded))
		shared_secret, err := crypto.ComputeSharedSecret(server_soul[:], body_decoded.ClientXPubKey)
		if err != nil {
			logger.Errorf("session init failed computing shared secret: %v", err)
			http.Error(w, "could not compute shared secret", http.StatusInternalServerError)
			return
		}
		shared_key := crypto.KDF(shared_secret, "@SESSION-SHARED-KEY", 32)
		payload_ciphered, payload_salt, payload_tag := crypto.MACE_Encrypt_AEAD(shared_key, payload_encoded, "@RESPONSE-PAYLOAD", 8, false)
		payload := append(payload_salt, payload_tag...)
		payload = append(payload, payload_ciphered...) // payload_salt is always exactly 12 bytes and payload_tag is always exactly 16 bytes
		signature := crypto.Sign(server_soul[:], payload)
		logger.Tracef("session init response cryptography complete payload_bytes=%d signature_bytes=%d", len(payload), len(signature))

		response := SessionInitResponse{
			Status:                 "ok",
			ServerEdPubKey:         b64(server_ed_pubkey),
			ServerXPubKey:          b64(server_x_pubkey),
			ServerXPubKeySignature: b64(server_x_pubkey_sign),
			Payload:                b64(payload),
			Signature:              b64(signature),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Errorf("session init response encode failed: %v", err)
			return
		}
		logger.Verbosef("session init completed identity=\"%s\" duration_ms=%d", trackerID, time.Since(reqStart).Milliseconds())
	}
}
