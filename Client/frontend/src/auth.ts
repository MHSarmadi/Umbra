import { useSecureVault } from "./db";
import { decodeBufferIntoDate } from "./tools/base64";

const { retrieveSecret, storeSecret } = useSecureVault();

type Buffer = Uint8Array<ArrayBuffer>;

export class Sensitive {
	value?: Buffer
	constructor(value: Buffer) {
		this.value = value;
	}

	destroy() {
		this.value?.fill(0); // Zero out the sensitive data
		delete this.value; // Remove reference
		this.value = new Uint8Array(0); // Clear reference
	}

	exportB64() {
		return btoa(String.fromCharCode(...this.value || new Uint8Array(0)));
	}

	static fromB64(b64: string): Sensitive {
		const bytes = Uint8Array.from(atob(b64), c => c.charCodeAt(0));
		return new Sensitive(bytes);
	}
}

const Auth = {
	session: {
		async init(): Promise<void> {
			if (!(await this.ready())) {
				await Promise.all([
					this.setId(new Sensitive(new Uint8Array(0))),
					this.setToken(new Sensitive(new Uint8Array(0))),
					this.setSoul(new Sensitive(new Uint8Array(0)))
				]);
			}
		},
		async expiryUnix(): Promise<Date | null> {
			const expiry_unix_millisec = await retrieveSecret("session_expiry_unix_millisec");
			if (expiry_unix_millisec && expiry_unix_millisec.length > 0) {
				return decodeBufferIntoDate(expiry_unix_millisec);
			}
			return null;
		},
		async setExpiryUnix(expiry_unix: Sensitive): Promise<void> {
			await storeSecret("session_expiry_unix_millisec", expiry_unix.value!);
			expiry_unix.destroy(); // Zero out the sensitive data after storing
		},
		async id(): Promise<Sensitive | null> {
			const id = await retrieveSecret("session_id");
			if (id && id.length > 0) {
				return new Sensitive(id);
			}
			return null;
		},
		async setId(id: Sensitive): Promise<void> {
			await storeSecret("session_id", id.value!);
			id.destroy(); // Zero out the sensitive data after storing
		},
		async token(): Promise<Sensitive | null> {
			const token = await retrieveSecret("session_token");
			if (token && token.length > 0) {
				return new Sensitive(token);
			}
			return null;
		},
		async setToken(token: Sensitive): Promise<void> {
			await storeSecret("session_token", token.value!);
			token.destroy(); // Zero out the sensitive data after storing
		},
		async soul(): Promise<Sensitive | null> {
			const soul = await retrieveSecret("session_soul");
			if (soul && soul.length > 0) {
				return new Sensitive(soul);
			}
			return null;
		},
		async setSoul(soul: Sensitive): Promise<void> {
			await storeSecret("session_soul", soul.value!);
			soul.destroy(); // Zero out the sensitive data after storing
		},
		server: {
			async EdPubKey(): Promise<Sensitive | null> {
				const key = await retrieveSecret("server_ed_pubkey");
				if (key && key.length > 0) {
					return new Sensitive(key);
				}
				return null;
			},
			async setEdPubKey(key: Sensitive): Promise<void> {
				await storeSecret("server_ed_pubkey", key.value!);
				key.destroy(); // Zero out the sensitive data after storing
			},
			async XPubKey(): Promise<Sensitive | null> {
				const key = await retrieveSecret("server_x_pubkey");
				if (key && key.length > 0) {
					return new Sensitive(key);
				}
				return null;
			},
			async setXPubKey(key: Sensitive): Promise<void> {
				await storeSecret("server_x_pubkey", key.value!);
				key.destroy(); // Zero out the sensitive data after storing
			}
		},
		temp: {
			async tokenCiphered(): Promise<Sensitive | null> {
				const data = await retrieveSecret("token_ciphered");
				if (data) {
					return new Sensitive(data);
				}
				return null;
			},
			async setTokenCiphered(token: Sensitive): Promise<void> {
				await storeSecret("token_ciphered", token.value!);
				token.destroy(); // Zero out the sensitive data after storing
			},
			async tokenCipherKeySalt(): Promise<Sensitive | null> {
				const data = await retrieveSecret("token_cipher_key_salt");
				if (data) {
					return new Sensitive(data);
				}
				return null;
			},
			async setTokenCipherKeySalt(salt: Sensitive): Promise<void> {
				await storeSecret("token_cipher_key_salt", salt.value!);
				salt.destroy(); // Zero out the sensitive data after storing
			},
			clear() {
				this.setTokenCiphered(new Sensitive(new Uint8Array(0)));
				this.setTokenCipherKeySalt(new Sensitive(new Uint8Array(0)));
			},
		},
		async ready() {
			const id = await this.id();
			const token = await this.token();
			const soul = await this.soul();
			return id !== null && id.value!.length > 0
				&& token !== null && token.value!.length > 0
				&& soul !== null && soul.value!.length > 0
		},
		async clear() {
			await Promise.all([
				this.setId(new Sensitive(new Uint8Array(0))),
				this.setToken(new Sensitive(new Uint8Array(0))),
				this.setSoul(new Sensitive(new Uint8Array(0))),
				this.server.setEdPubKey(new Sensitive(new Uint8Array(0))),
				this.server.setXPubKey(new Sensitive(new Uint8Array(0)))
			]);
			this.temp.clear();
		},
		async logout() {
			this.clear();
		},
	},
}

export function useAuth() {
	return Auth;
}