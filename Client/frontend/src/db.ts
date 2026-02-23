// src/composables/useSecureVault.ts
import { ref } from 'vue'

const DB_NAME = 'umbraVault'
const DB_VERSION = 1
const KEY_STORE = 'vaultKeys'
const SECRETS_STORE = 'secrets'
const KEY_ID = 'main-aes-gcm-key'

// ────────────────────────────────────────────────
// Types
// ────────────────────────────────────────────────

interface EncryptedPayload {
	id: string
	iv: number[]
	ciphertext: number[]
}

interface VaultKeyEntry {
	id: string
	key: CryptoKey
}

// ────────────────────────────────────────────────
// IndexedDB Helpers
// ────────────────────────────────────────────────

function openVaultDb(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		const req = indexedDB.open(DB_NAME, DB_VERSION)

		req.onupgradeneeded = (event) => {
			const db = (event.target as IDBOpenDBRequest).result
			//   const oldVersion = event.oldVersion

			// Create vaultKeys store if needed (for the non-extractable CryptoKey)
			if (!db.objectStoreNames.contains(KEY_STORE)) {
				db.createObjectStore(KEY_STORE, { keyPath: 'id' })
			}

			// Create secrets store if needed (for encrypted payloads)
			if (!db.objectStoreNames.contains(SECRETS_STORE)) {
				db.createObjectStore(SECRETS_STORE, { keyPath: 'id' })
			}
		}

		req.onsuccess = (e) => resolve((e.target as IDBOpenDBRequest).result)
		req.onerror = (e) => reject((e.target as IDBOpenDBRequest).error)
	})
}

// ────────────────────────────────────────────────
// Vault Key (non-extractable CryptoKey) Management
// ────────────────────────────────────────────────

async function saveVaultKey(key: CryptoKey): Promise<void> {
	const db = await openVaultDb()
	return new Promise((resolve, reject) => {
		const tx = db.transaction(KEY_STORE, 'readwrite')
		const store = tx.objectStore(KEY_STORE)
		const req = store.put({ id: KEY_ID, key } as VaultKeyEntry)

		req.onsuccess = () => {
			db.close()
			resolve()
		}
		req.onerror = (e) => {
			db.close()
			reject((e.target as IDBRequest).error)
		}
	})
}

async function loadVaultKey(): Promise<CryptoKey | null> {
	const db = await openVaultDb()
	return new Promise((resolve, reject) => {
		const tx = db.transaction(KEY_STORE, 'readonly')
		const store = tx.objectStore(KEY_STORE)
		const req = store.get(KEY_ID)

		req.onsuccess = () => {
			const entry = req.result as VaultKeyEntry | undefined
			db.close()
			resolve(entry?.key ?? null)
		}
		req.onerror = (e) => {
			db.close()
			reject((e.target as IDBRequest).error)
		}
	})
}

async function createOrGetVaultKey(): Promise<CryptoKey> {
	let key = await loadVaultKey()

	if (key) {
		return key
	}

	if (crypto && crypto.subtle) { 
		key = await crypto.subtle.generateKey(
			{
				name: 'AES-GCM',
				length: 256
			},
			false,                    // non-extractable
			['encrypt', 'decrypt']
		)

		await saveVaultKey(key)
		return key
	} else {
		throw new Error("No Web Crypto Support by this device.")
	}
}

// ────────────────────────────────────────────────
// Encrypt / Decrypt
// ────────────────────────────────────────────────

async function encryptSecret(
	secret: Uint8Array<ArrayBuffer>,
	key: CryptoKey
): Promise<Omit<EncryptedPayload, 'id'>> {
	if (secret.byteLength === 0) {
		throw new Error('Cannot encrypt empty secret')
	}

	const iv = crypto.getRandomValues(new Uint8Array(12))

	const ciphertextBuffer = await crypto.subtle.encrypt(
		{ name: 'AES-GCM', iv },
		key,
		secret
	)

	return {
		iv: Array.from(iv),
		ciphertext: Array.from(new Uint8Array(ciphertextBuffer))
	}
}

async function decryptSecret(
	payload: EncryptedPayload,
	key: CryptoKey
): Promise<Uint8Array<ArrayBuffer>> {
	const iv = new Uint8Array(payload.iv)
	const ciphertext = new Uint8Array(payload.ciphertext)

	const decryptedBuffer = await crypto.subtle.decrypt(
		{ name: 'AES-GCM', iv },
		key,
		ciphertext.buffer
	)

	return new Uint8Array(decryptedBuffer)
}

// ────────────────────────────────────────────────
// Public API
// ────────────────────────────────────────────────

export function useSecureVault() {
	const error = ref<string | null>(null)

	async function storeSecret(id: string, secret: Uint8Array<ArrayBuffer>): Promise<void> {
		try {
			const vaultKey = await createOrGetVaultKey()
			if (!secret.length) {
				secret = new Uint8Array([0]);
			}
			const encrypted = await encryptSecret(secret, vaultKey)

			// Zero original secret immediately
			secret.fill(0)

			const db = await openVaultDb()
			await new Promise<void>((resolve, reject) => {
				const tx = db.transaction(SECRETS_STORE, 'readwrite')
				const store = tx.objectStore(SECRETS_STORE)
				const req = store.put({ id, ...encrypted })

				req.onsuccess = () => resolve()
				req.onerror = (e) => reject((e.target as IDBRequest).error)
			})

			db.close()
		} catch (err) {
			error.value = err instanceof Error ? err.message : 'Unknown error'
			throw err
		}
	}

	async function retrieveSecret(id: string): Promise<Uint8Array<ArrayBuffer> | null> {
		try {
			const vaultKey = await loadVaultKey()
			if (!vaultKey) {
				throw new Error('Vault key not found')
			}

			const db = await openVaultDb()
			const payload = await new Promise<EncryptedPayload | undefined>((resolve, reject) => {
				const tx = db.transaction(SECRETS_STORE, 'readonly')
				const store = tx.objectStore(SECRETS_STORE)
				const req = store.get(id)

				req.onsuccess = () => resolve(req.result)
				req.onerror = (e) => reject((e.target as IDBRequest).error)
			})

			db.close()

			if (!payload) return null

			const decrypted = await decryptSecret(payload, vaultKey)
			
			return (decrypted.length > 1 || (decrypted.length == 1 && decrypted[0] !== 0)) ? decrypted : new Uint8Array(0);
		} catch (err) {
			error.value = err instanceof Error ? err.message : 'Unknown error'
			throw err
		}
	}

	return {
		storeSecret,
		retrieveSecret,
		error
	}
}