# Doc - CRYPTO TOOLS
This document lists and explains crypto tools that are implemented for **Umbra** (*Now only v1*)

> Scope: This document applies only to **v1**. More Cryptography algorithms will be added later. Now, we are working with BLAKE3 only.

- Hash (@/crypto/hash.go)
	> Same as blake3.Sum512 - outputs [64]bytes.

- KDF (@/crypto/kdf.go)
	> Same as blake3.DeriveKey with special context prefix which is related to current version of Umbra.

- MAC (@/crypto/mac.go)
	> Same as blake3.Keyed, And internally uses blake3.DeriveKey with special context prefix (which is related to current version of Umbra) to be used as MAC's key.

- Symmetric Encryption (@/crypto/encryption.go)
	Based on MACE-BLAKE3 (see [github.com/MHSarmadi/MACE](https://github.com/MHSarmadi/MACE))
	
	> Has 4 modes:
	>	- Simple: `MACE_<Encrypt/Decrypt>`
	>	- MIXIN: `MACE_<Encrypt/Decrypt>_MIXIN`
	>	- AEAD: `MACE_<Encrypt/Decrypt>_AEAD`
	>	- MIXIN+AEAD: `MACE_<Encrypt/Decrypt>_MIXIN_AEAD`

	More details and important warnings are there in 1.1-MACE.md