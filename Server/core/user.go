package core

import (
	"context"
	"crypto/rand"

	umbra_crypto "github.com/MHSarmadi/Umbra/Server/crypto"
	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/MHSarmadi/Umbra/Server/models"
)

func NewUser(ctx context.Context, s *database.BadgerStore, username, password string, recovery_key []byte) (uuid []byte, err error) {
	soul := make([]byte, 32)
	rand.Read(soul)

	e_pub_key := umbra_crypto.DeriveEd25519PubKey(soul)
	x_pub_key, err := umbra_crypto.DeriveX25519PubKey(soul)
	if err != nil {
		return nil, err
	}

	soul_key := umbra_crypto.Sum([]byte(password))
	soul_cipher, soul_salt, soul_tag := umbra_crypto.MACE_Encrypt_AEAD(soul_key[:32], soul, "@SOUL-ENCRYPTION", 4, false)

	true_recovery_key := umbra_crypto.Sum(recovery_key)
	recovery_cipher, recovery_salt, recovery_tag := umbra_crypto.MACE_Encrypt_AEAD(true_recovery_key[:32], soul_key[:], "@SOUL-ENCRYPTION-@RECOVERY-KEY", 16, false)

	uuid = make([]byte, 32)
	rand.Read(uuid)

	user := models.User{
		UUID:               uuid,
		Username:           username,
		XPublicKey:         x_pub_key,
		EPublicKey:         e_pub_key,
		EncipheredSoul:     soul_cipher,
		EncipheredSoulSalt: soul_salt,
		EncipheredSoulTag:  soul_tag,
		SoulRecovery:       recovery_cipher,
		SoulRecoverySalt:   recovery_salt,
		SoulRecoveryTag:    recovery_tag,
	}

	s.PutUser(ctx, &user)

	return
}
