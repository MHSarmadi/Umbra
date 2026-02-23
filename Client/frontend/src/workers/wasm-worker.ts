/// <reference lib="webworker" />

import { decodeBase64 } from "../tools/base64";
import { useAuth, Sensitive } from "../auth";

const Auth = useAuth();

let goRuntimePromise: Promise<void> | null = null;
async function ensureGoRuntimeLoaded(): Promise<void> {
	if (goRuntimePromise) {
		return goRuntimePromise;
	}
	goRuntimePromise = (async () => {
		const response = await fetch("/wasm_exec.js");
		const wasmExecCode = await response.text();
		eval(wasmExecCode); // Defines 'Go' class
	})();
	return goRuntimePromise;
}

let go = null, initialized = false;

let baseURL: URL|null = null;
async function getBaseURL(): Promise<URL> {
	return new Promise((resolve, reject) => {
		const interval_delay = 5
		const timeout = 10**4
		const max_iterations = timeout/interval_delay

		let n = 0;
		const interval = setInterval(() => {
			if (baseURL) {
				clearInterval(interval)
				resolve(baseURL)
			} else if (n >= max_iterations) {
				clearInterval(interval)
				reject("Base URL not specified in the first 10 seconds")
			}
			n++
		}, interval_delay)
	})
}

declare global {
	interface Window {
		onProgressMade?: (type: string, id: string, percentage: number) => void;
		SessionKeypair?: () => Promise<{
			ed_pubkey: string,
			x_pubkey: string,
			x_pubkey_sign: string,
			soul: Uint8Array<ArrayBuffer>
		}>;
		IntroduceServer?: (
			soul: Uint8Array<ArrayBuffer>,
			server_ed_pubkey: string,
			server_x_pubkey: string,
			server_x_pubkey_sign: string,
			payload: string,
			signature: string,
		) => Promise<{
			session_id: string,
			captcha_challenge: string,
			pow_challenge: string,
			pow_salt: string,
			pow_params: {
				memory_mb: number,
				iterations: number,
				parallelism: number
			},
			session_token_ciphered: string,
			session_token_cipher_key_salt: string
		}>;

		// expected args: progress_id, challenge, salt, memory_mb, iterations, parallelism
		// return: Promise<error|number>
		ComputePoW?: (progress_id: string, challenge: Uint8Array<ArrayBuffer>, salt: Uint8Array<ArrayBuffer>, memory_mb: number, iterations: number, parallelism: number) => Promise<number>;
		
		// expected args: captcha_challenge_numeric, session_token_ciphered, session_id
		// return: Promise<string> which is the decipehred session_token
		CheckoutCaptcha?: (captcha_solution_numeric: number, session_token_ciphered: Uint8Array<ArrayBuffer>, session_token_cipher_key_salt: Uint8Array<ArrayBuffer>, session_id: Uint8Array<ArrayBuffer>) => Promise<string>;
	}
}

self.onmessage = async (event: MessageEvent) => {
	if (!initialized) {
		try {
			await ensureGoRuntimeLoaded();
			go = new Go();
			const result = await WebAssembly.instantiateStreaming(fetch('/umbra.wasm'), go.importObject);
			go.run(result.instance);
			await new Promise(resolve => setTimeout(resolve, 100)); // Ensure WASM is fully initialized
			if ((globalThis as any).umbraReady?.() !== "Umbra WASM initialized") {
				throw new Error("WASM module did not initialize correctly");
			}
			initialized = true;
			console.log("WASM module initialized successfully");
		} catch (err) {
			console.error('Error initializing WASM module:', err);
			self.postMessage({ type: 'init', success: false, error: err });
			return;
		}
	}
	if (event.data.type === 'init') {
		self.postMessage({ type: 'init', success: true });
		postMessage({ type: "freed", processType: 'init' })
	} else if (event.data.type === 'setBaseURL') {	
		try {
			if (typeof event.data.url !== 'string') {
				throw new Error("Invalid base URL: must be a string");
			}
			baseURL = new URL(event.data.url);
			self.postMessage({ type: 'setBaseURL', success: true });
		} catch (err) {
			console.error('Error setting base URL:', err);
			self.postMessage({ type: 'setBaseURL', success: false, error: err });
		} finally {
			postMessage({ type: "freed", processType: event.data.type })
		}
	} else if (event.data.type === 'SessionKeypair') {
		let request_payload = ""
		try {
			const pubkeys = await self.SessionKeypair?.()
			if (typeof pubkeys !== 'object') {
				throw new Error("SessionKeypair did not return a valid result");
			}

			await Auth.session.setSoul(new Sensitive(pubkeys.soul));
			
			// Remove sensitive data from memory
			pubkeys.soul.fill(0);
			
			request_payload = JSON.stringify({
				"client_ed_pubkey": pubkeys.ed_pubkey,
				"client_x_pubkey": pubkeys.x_pubkey,
				"client_x_pubkey_sign": pubkeys.x_pubkey_sign
			})

			postMessage({ type: 'SessionKeypair', success: true });

			try {
				const headers = new Headers
				headers.append('Content-Type', 'application/json')

				const result = await fetch(new URL("/session/init", await getBaseURL()), {
					method: "POST",
					headers,
					body: request_payload
				})
				if (!result.ok) {
					throw new Error(`Failed to send session initialization: ${result.status} ${result.statusText}`);
				}
				
				const response = await result.json();
				if (response.status !== 'ok') {
					throw new Error(`Session initialization failed: ${response.error || "Unknown error"}`);
				}

				self.postMessage({ type: 'SendSessionKeypair', success: result.ok, response });

				const { payload, server_ed_pubkey, server_x_pubkey, server_x_pubkey_sign, signature } = response;

				if (typeof payload !== 'string' || typeof server_ed_pubkey !== 'string' || typeof server_x_pubkey !== 'string' || typeof server_x_pubkey_sign !== 'string' || typeof signature !== 'string') {
					throw new Error("Invalid PoW response: missing or invalid fields");
				}

				await Promise.all([
					Auth.session.server.setEdPubKey(new Sensitive(decodeBase64(server_ed_pubkey))),
					Auth.session.server.setXPubKey(new Sensitive(decodeBase64(server_x_pubkey))),
				])

				const soul = await Auth.session.soul();
				if (!soul) {
					throw new Error("Session soul not found in vault");
				}
				
				const deciphered_payload = await self.IntroduceServer?.(soul.value!, server_ed_pubkey, server_x_pubkey, server_x_pubkey_sign, payload, signature);

				// Remove sensitive data from memory
				soul.destroy();
				response.server_ed_pubkey = "";
				delete response.server_ed_pubkey;
				response.server_x_pubkey = "";
				delete response.server_x_pubkey;
				response.server_x_pubkey_sign = "";
				delete response.server_x_pubkey_sign;

				if (!deciphered_payload || typeof deciphered_payload !== 'object' || !('session_token_ciphered' in deciphered_payload)) {
					throw new Error("IntroduceServer did not return a valid session token ciphered");
				}

				await Promise.all([
					Auth.session.temp.setTokenCiphered(new Sensitive(decodeBase64(deciphered_payload.session_token_ciphered))),
					Auth.session.temp.setTokenCipherKeySalt(new Sensitive(decodeBase64(deciphered_payload.session_token_cipher_key_salt))),
					Auth.session.setId(new Sensitive(decodeBase64(deciphered_payload.session_id))),
				])

				// Remove sensitive data from memory
				deciphered_payload.session_token_ciphered = "";
				deciphered_payload.session_token_cipher_key_salt = "";
				deciphered_payload.session_id = "";

				self.postMessage({ type: 'IntroduceServer', success: true, payload: deciphered_payload });
			} catch (err) {
				console.error('Error during sending session initialization:', err);
				self.postMessage({ type: 'SendSessionKeypair', success: false, error: err });
			}
		} catch (err) {
			console.error('Error during session key pair generation:', err);
			self.postMessage({ type: 'SessionKeypair', success: false, error: err });
		} finally {
			postMessage({ type: "freed", processType: event.data.type })
		}
	} else if (event.data.type === 'PoW') {
		self.ComputePoW?.(
			event.data.progress_id,
			new Uint8Array(event.data.challenge),
			new Uint8Array(event.data.salt),
			event.data.memory_mb,
			event.data.iterations,
			event.data.parallelism
		)?.then((result: number) => {;
			if (typeof result !== 'number') {
				throw new Error("ComputePoW did not return a valid result");
			}

			self.postMessage({ type: 'PoW', success: true, result });
		})?.catch((err: Error) => {
			console.error('Error during PoW computation:', err);
			self.postMessage({ type: 'PoW', success: false, error: err.message });
		})?.finally(() => {
			postMessage({ type: "freed", processType: event.data.type })
		});
	} else if (event.data.type === 'CheckoutCaptcha') {
		const { captcha_response } = event.data;
		if (typeof captcha_response !== 'string' || captcha_response.length !== 6 || !/^\d{6}$/.test(captcha_response)) {
			console.warn("Invalid CAPTCHA response:", captcha_response);
			self.postMessage({ type: 'CheckoutCaptcha', success: false, error: "Invalid CAPTCHA response" });
			postMessage({ type: "freed", processType: event.data.type })
			return;
		}
		const captcha_response_numeric = parseInt(captcha_response);

		try {
			const [ session_token_ciphered, session_token_cipher_key_salt, session_id ] = await Promise.all([
				Auth.session.temp.tokenCiphered(),
				Auth.session.temp.tokenCipherKeySalt(),
				Auth.session.id()
			]);

			if (!session_token_ciphered || !session_token_cipher_key_salt || !session_id) {
				throw new Error("Session token ciphered, key salt, or ID not found in vault");
			}

			// Decipher Session Token
			const session_token = await self.CheckoutCaptcha?.(captcha_response_numeric, session_token_ciphered.value!, session_token_cipher_key_salt.value!, session_id.value!);
			if (typeof session_token !== 'string') {
				throw new Error("CheckoutCaptcha did not return a valid session token");
			}

			// Remove sensitive data from memory
			session_token_ciphered.destroy();
			session_token_cipher_key_salt.destroy();
			session_id.destroy();

			Auth.session.temp.clear(); // Clear temporary session data

			console.log(session_token)

			await Auth.session.setToken(new Sensitive(decodeBase64(session_token)));
			
			self.postMessage({ type: 'CheckoutCaptcha', success: true });
		} catch (err) {
			console.error('Error validating CAPTCHA:', err);
			self.postMessage({ type: 'CheckoutCaptcha', success: false, error: err });
			return;
		} finally {
			postMessage({ type: "freed", processType: event.data.type })
		}
	} else {
		console.warn('Unknown message type:', event.data.type);
		postMessage({ type: "freed", processType: event.data.type })
	}
}


self.onProgressMade = (type: string, id: string, percentage: number) => {
	self.postMessage({ type: 'progress', success: true, progressType: type, id, percentage });
}
