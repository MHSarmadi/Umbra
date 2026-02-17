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
)

var (
	b64url  = base64.RawURLEncoding.EncodeToString
	db64url = base64.RawURLEncoding.DecodeString
)

func sessionInitIdentityHash(r *http.Request) string {
	identityRaw := clientIP(r)
	sum := crypto.Sum([]byte(identityRaw))
	return b64url(sum[:16])
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
	var (
		err          error
		body_encoded models_requests.SessionInitRequestEncoded
		body_decoded models_requests.SessionInitRequestDecoded
	)
	if err := json.NewDecoder(r.Body).Decode(&body_encoded); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body_decoded.ClientEdPubKey, err = db64url(body_encoded.ClientEdPubKey); err != nil {
		http.Error(w, "invalid client_ed_pubkey base64url encoding", http.StatusBadRequest)
		return
	} else if body_decoded.ClientXPubKey, err = db64url(body_encoded.ClientXPubKey); err != nil {
		http.Error(w, "invalid client_x_pubkey base64url encoding", http.StatusBadRequest)
		return
	} else if body_decoded.ClientXPubKeySignature, err = db64url(body_encoded.ClientXPubKeySignature); err != nil {
		http.Error(w, "invalid client_x_pubkey_sign base64url encoding", http.StatusBadRequest)
		return
	} else if len(body_decoded.ClientEdPubKey) != 32 || len(body_decoded.ClientXPubKey) != 32 {
		// Check lengths BEFORE calling Verify
		http.Error(w, "invalid ed-pubkey or x-pubkey length", http.StatusBadRequest)
		return
	} else if crypto.Verify(body_decoded.ClientEdPubKey, body_decoded.ClientXPubKey, body_decoded.ClientXPubKeySignature) == false {
		http.Error(w, "invalid signature over client_x_pubkey", http.StatusBadRequest)
		return
	} else {
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
			http.Error(w, "invalid ed-pubkey or x-pubkey", http.StatusBadRequest)
			return
		}
		var session_id [24]byte
		if _, err := rand.Read(session_id[:]); err != nil {
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}

		var server_soul [32]byte
		if _, err := rand.Read(server_soul[:]); err != nil {
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}

		server_ed_pubkey := crypto.DeriveEd25519PubKey(server_soul[:])
		server_x_pubkey, err := crypto.DeriveX25519PubKey(server_soul[:])
		if err != nil {
			http.Error(w, "could not derive pubkey", http.StatusInternalServerError)
			return
		}
		server_x_pubkey_sign := crypto.Sign(server_soul[:], server_x_pubkey)

		var pow_challenge [powChallengeSize]byte
		if _, err := rand.Read(pow_challenge[:]); err != nil {
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
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}

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
			http.Error(w, "could not draw captcha", http.StatusInternalServerError)
			return
		}

		var session_token [24]byte
		if _, err := rand.Read(session_token[:]); err != nil {
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}
		session_token_ciphered, session_token_salt := crypto.MACE_Encrypt(captcha_solution_bytes, session_token[:], "@SESSION-TOKEN", 2, false)

		session := models.Session{
			UUID: session_id,

			CreatedAt: now,
			ExpiresAt: now.Add(Expiry_Offset).UTC(),

			ClientEdPubKey: [32]byte(body_decoded.ClientEdPubKey),
			ClientXPubKey:  [32]byte(body_decoded.ClientXPubKey),

			ServerSoul: server_soul,

			SessionToken: session_token,

			PoWChallenge: pow_challenge,
			PoWParams:    pow_params,
			PoWSalt:      pow_salt,
		}

		if err := c.storage.PutSession(c.ctx, &session); err != nil {
			http.Error(w, "could not store seesion", http.StatusInternalServerError)
			return
		}

		type SessionInitResponse struct {
			Status                 string `json:"status"`
			SessionUUID            string `json:"session_id"`
			ServerEdPubKey         string `json:"server_ed_pubkey"`
			ServerXPubKey          string `json:"server_x_pubkey"`
			ServerXPubKeySignature string `json:"server_x_pubkey_sign"`
			Payload                string `json:"payload"`
			Signature              string `json:"signature"`
			// SharedSecret           string `json:"shared_secret"` // ONLY FOR TESTING PURPOSES. REMOVE THIS IN PRODUCTION.
		}
		type SessionInitRawPayload struct {
			CaptchaChallenge string               `json:"captcha_challenge"`
			PoWChallenge     string               `json:"pow_challenge"`
			PowParams        models.PowParamsType `json:"pow_params"`
			PoWSalt          string               `json:"pow_salt"`
			SessionToken     string               `json:"session_token_ciphered"`
		}

		payload_raw := SessionInitRawPayload{
			PoWChallenge:     b64url(session.PoWChallenge[:]),
			PowParams:        session.PoWParams,
			PoWSalt:          b64url(session.PoWSalt[:]),
			SessionToken:     b64url(append(session_token_salt, session_token_ciphered...)), // session_token_salt is always exactly 12 bytes
			CaptchaChallenge: b64url(captcha_png),
		}
		payload_encoded, err := json.Marshal(payload_raw)
		if err != nil {
			http.Error(w, "could not marshal to json", http.StatusInternalServerError)
			return
		}
		shared_secret, err := crypto.ComputeSharedSecret(server_soul[:], body_decoded.ClientXPubKey)
		if err != nil {
			http.Error(w, "could not compute shared secret", http.StatusInternalServerError)
			return
		}
		shared_key := crypto.KDF(shared_secret, "@SESSION-SHARED-KEY", 32)
		payload_ciphered, payload_salt, payload_tag := crypto.MACE_Encrypt_AEAD(shared_key, payload_encoded, "@RESPONSE-PAYLOAD", 8, false)
		payload := append(payload_salt, payload_tag...)
		payload = append(payload, payload_ciphered...) // payload_salt is always exactly 12 bytes and payload_tag is always exactly 16 bytes
		signature := crypto.Sign(server_soul[:], payload)

		response := SessionInitResponse{
			Status:                 "ok",
			SessionUUID:            b64url(session.UUID[:]),
			ServerEdPubKey:         b64url(server_ed_pubkey),
			ServerXPubKey:          b64url(server_x_pubkey),
			ServerXPubKeySignature: b64url(server_x_pubkey_sign),
			Payload:                b64url(payload),
			Signature:              b64url(signature),
			// SharedSecret:           b64url(shared_key), // ONLY FOR TESTING PURPOSES. REMOVE THIS IN PRODUCTION.
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
