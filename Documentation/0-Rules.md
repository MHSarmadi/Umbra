# Doc - GROUND RULES
This document lists basic rules of **Umbra**'s development (*Now only v1*)

> Scope: This document applies only to **v1**. Features like algorithm selection, blockchain integration, and multi-platform UI are deferred to later versions.

- Platform:
	> I'm sure I will use Go programming language for both **Client** and **Server** side.
- General Algorithms:
	> I will certainly use those safe and fast algorithms:
	> - **Digital Signatures**: Ed25519 Curve of ECC
	> - **Safe Key Exchanges**: X25519 Curve of ECC
	> - **Hash, MAC and simple KDF**: BLAKE3 (1)
	> - **Symmetric Encryption**: MACE-BLAKE3-MIXIN-AEAD (1) (2)
	> - **Memory-Safe KDF**: Argon2id
	> (1): *Those algo's are all for v1. Later, user will decide on his own algorithm.*
	> (2): MACE-BLAKE3-MIXIN-AEAD: A planned custom AEAD mode using BLAKE3 as the MAC engine in a SIV-like construction for misuse-resistant encryption. More details are there in ~~the related documentation~~.
- Protocols:
	> I will connect **Client** side to the **Server** with a **WebSocket** and every message is going to be formatted with **ProtoBuf**
	> **ProtoBuf** (*proto3*) is my choice because its a compact binary encoding and fits **Golang** very well.

You will find more information about all these in other **Documentation Files** within the same directory.

*Last Update: Aug 9th, 2025*