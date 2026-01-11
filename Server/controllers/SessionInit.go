package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MHSarmadi/Umbra/Server/captcha"
	"github.com/MHSarmadi/Umbra/Server/crypto"
	math_tools "github.com/MHSarmadi/Umbra/Server/math"
	"github.com/MHSarmadi/Umbra/Server/models"
	models_requests "github.com/MHSarmadi/Umbra/Server/models/requests"
)

const Expiry_Offset = 300 * time.Second

var (
	b64url  = base64.RawURLEncoding.EncodeToString
	db64url = base64.RawURLEncoding.DecodeString
)

func (c *Controller) SessionInit(w http.ResponseWriter, r *http.Request) {
	var (
		err          error
		body_encoded models_requests.SessionInitRequestEncoded
		body_decoded models_requests.SessionInitRequestDecoded
	)
	json.NewDecoder(r.Body).Decode(&body_encoded)
	
	if body_decoded.ClientEdPubKey, err = db64url(body_encoded.ClientEdPubKey); err != nil {
		http.Error(w, "invalid client_ed_pubkey base64url encoding", http.StatusBadRequest)
	} else if body_decoded.ClientXPubKey, err = db64url(body_encoded.ClientXPubKey); err != nil {
		http.Error(w, "invalid client_x_pubkey base64url encoding", http.StatusBadRequest)
	} else if body_decoded.ClientXPubKeySignature, err = db64url(body_encoded.ClientXPubKeySignature); err != nil {
		http.Error(w, "invalid client_x_pubkey_sign base64url encoding", http.StatusBadRequest)
	} else if crypto.Verify(body_decoded.ClientEdPubKey, body_decoded.ClientXPubKey, body_decoded.ClientXPubKeySignature) == false {
		http.Error(w, "invalid signature over client_x_pubkey", http.StatusBadRequest)
	} else {
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
		}
		server_x_pubkey_sign := crypto.Sign(server_soul[:], server_x_pubkey)

		var pow_challenge [2]byte
		if _, err := rand.Read(pow_challenge[:]); err != nil {
			http.Error(w, "could not read entropy", http.StatusInternalServerError)
			return
		}
		pow_params := models.PowParamsType{
			MemoryMB:    256,
			Iterations:  4,
			Parallelism: 1,
		}

		captcha_solution := math_tools.RandomDecimalString(6)
		var captcha_solution_numeric uint64 = 0
		for _, c := range captcha_solution {
			captcha_solution_numeric *= 10
			captcha_solution_numeric += uint64(c - '0')
		}
		captcha_png, err := captcha.GenerateNumericCaptcha(captcha_solution)
		if err != nil {
			http.Error(w, "could not draw captcha", http.StatusInternalServerError)
		}

		session := models.Session{
			UUID: session_id,

			CreatedAt: time.Now().UTC(),
			ExpiresAt: time.Now().Add(Expiry_Offset).UTC(),

			ClientEdPubKey: [32]byte(body_decoded.ClientEdPubKey),
			ClientXPubKey:  [32]byte(body_decoded.ClientXPubKey),

			ServerSoul: server_soul,

			PoWChallenge:    pow_challenge,
			PoWParams:       pow_params,
			CaptchaSolution: captcha_solution_numeric,
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
		}
		type SessionInitRawPayload struct {
			CaptchaChallenge string               `json:"captcha_challenge"`
			PoWChallenge     string               `json:"pow_challenge"`
			PowParams        models.PowParamsType `json:"pow_params"`
		}

		payload_raw := SessionInitRawPayload{
			CaptchaChallenge: b64url(captcha_png),
			PoWChallenge:     b64url(session.PoWChallenge[:]),
			PowParams:        session.PoWParams,
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
		payload_ciphered, payload_salt := crypto.MACE_Encrypt(shared_key, payload_encoded, "@RESPONSE-PAYLOAD", 1, false)
		payload := append(payload_salt, payload_ciphered...) // payload_salt is always exactly 12 bytes
		signature := crypto.Sign(server_soul[:], payload)

		response := SessionInitResponse{
			Status:                 "ok",
			SessionUUID:            b64url(session.UUID[:]),
			ServerEdPubKey:         b64url(server_ed_pubkey),
			ServerXPubKey:          b64url(server_x_pubkey),
			ServerXPubKeySignature: b64url(server_x_pubkey_sign),
			Payload:                b64url(payload),
			Signature:              b64url(signature),
		}
		json.NewEncoder(w).Encode(response)
	}
}
