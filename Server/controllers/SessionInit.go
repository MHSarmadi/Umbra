package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MHSarmadi/Umbra/Server/crypto"
	"github.com/MHSarmadi/Umbra/Server/models"
	models_requests "github.com/MHSarmadi/Umbra/Server/models/requests"
)

const Expiry_Offset = 300 * time.Second

func (c *Controller) SessionInit(w http.ResponseWriter, r *http.Request) {
	var body_encoded models_requests.SessionInitRequestEncoded
	json.NewDecoder(r.Body).Decode(&body_encoded)

	var err error
	var body_decoded models_requests.SessionInitRequestDecoded
	if body_decoded.ClientEdPubKey, err = base64.RawURLEncoding.DecodeString(body_encoded.ClientEdPubKey); err != nil {
		BadRequest(w, "invalid client_ed_pubkey base64url encoding")
	} else if body_decoded.ClientXPubKey, err = base64.RawURLEncoding.DecodeString(body_encoded.ClientXPubKey); err != nil {
		BadRequest(w, "invalid client_x_pubkey base64url encoding")
	} else if body_decoded.ClientXPubKeySignature, err = base64.RawURLEncoding.DecodeString(body_encoded.ClientXPubKeySignature); err != nil {
		BadRequest(w, "invalid client_x_pubkey_sign base64url encoding")
	} else if crypto.Verify(body_decoded.ClientEdPubKey, body_decoded.ClientXPubKey, body_decoded.ClientXPubKeySignature) == false {
		BadRequest(w, "invalid signature over client_x_pubkey")
	} else {
		if len(body_decoded.ClientEdPubKey) != 32 || len(body_decoded.ClientXPubKey) != 32 {
			BadRequest(w, "invalid ed-pubkey or x-pubkey")
			return
		}
		var session_id [24]byte
		if _, err := rand.Read(session_id[:]); err != nil {
			InternalServerError(w, "unexpected server error")
			return
		}

		var server_soul [32]byte
		if _, err := rand.Read(server_soul[:]); err != nil {
			InternalServerError(w, "unexpected server error")
			return
		}

		server_ed_pubkey := crypto.DeriveEd25519PubKey(server_soul[:])
		server_x_pubkey, err := crypto.DeriveX25519PubKey(server_soul[:])
		if err != nil {
			panic("unexpected error while deriving x25519 pubkey")
		}
		server_x_pubkey_sign := crypto.Sign(server_soul[:], server_x_pubkey)

		var pow_challenge [2]byte
		if _, err := rand.Read(pow_challenge[:]); err != nil {
			InternalServerError(w, "unexpected server error")
			return
		}
		pow_params := models.PowParamsType{
			MemoryMB:    256,
			Iterations:  4,
			Parallelism: 1,
		}

		var captcha_solution [8]byte
		if _, err := rand.Read(captcha_solution[:]); err != nil {
			InternalServerError(w, "unexpected server error")
			return
		}

		session := models.Session{
			UUID: session_id,

			CreatedAt: time.Now().UTC(),
			ExpiresAt: time.Now().Add(Expiry_Offset).UTC(),

			ClientEdPubKey: [32]byte(body_decoded.ClientEdPubKey),
			ClientXPubKey:  [32]byte(body_decoded.ClientXPubKey),

			ServerSoul: server_soul,

			PoWChallenge: pow_challenge,
			PoWParams:    pow_params,
		}
		binary.BigEndian.PutUint64(captcha_solution[:], session.CaptchaSolution)

		if err := c.storage.PutSession(c.ctx, &session); err != nil {
			InternalServerError(w, "unexpected server error")
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
			CaptchaChallenge string `json:"captcha_challenge"`
			PoWChallenge string `json:"pow_challenge"`
			PowParams models.PowParamsType `json:"pow_params"`
		}
		payload_raw := SessionInitRawPayload{
			CaptchaChallenge: "",
			PoWChallenge: base64.RawURLEncoding.EncodeToString(session.PoWChallenge[:]),
			PowParams: session.PoWParams,
		}
		payload_encoded, err := json.Marshal(payload_raw)
		if err != nil {
			InternalServerError(w, "unexpected server error")
			return
		}
		shared_secret, err := crypto.ComputeSharedSecret(server_soul[:], body_decoded.ClientXPubKey)
		if err != nil {
			InternalServerError(w, "unexpected server error")
			return
		}
		shared_key := crypto.KDF(shared_secret, "@SESSION-SHARED-KEY", 32)
		payload_ciphered, payload_salt := crypto.MACE_Encrypt(shared_key, payload_encoded, "@RESPONSE-PAYLOAD", 1, false)
		payload := append(payload_ciphered, payload_salt...) // payload_salt is always exactly 12 bytes
		signature := crypto.Sign(server_soul[:], payload)

		response := SessionInitResponse{
			Status: "ok",
			SessionUUID: base64.RawURLEncoding.EncodeToString(session.UUID[:]),
			ServerEdPubKey: base64.RawURLEncoding.EncodeToString(server_ed_pubkey),
			ServerXPubKey: base64.RawURLEncoding.EncodeToString(server_x_pubkey),
			ServerXPubKeySignature: base64.RawURLEncoding.EncodeToString(server_x_pubkey_sign),
			Payload: base64.RawURLEncoding.EncodeToString(payload),
			Signature: base64.RawURLEncoding.EncodeToString(signature),
		}
		json.NewEncoder(w).Encode(response)
	}
}
