/// <reference lib="webworker" />

import { decodeBase64 } from "../tools/base64";
import { useSecureVault } from "../db";

const { storeSecret, retrieveSecret } = useSecureVault();

const response = await fetch("/wasm_exec.js");
const wasmExecCode = await response.text();
eval(wasmExecCode); // Defines 'Go' class

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
			session_token_ciphered: string
		}>;

		// expected args: progress_id, challenge, salt, memory_mb, iterations, parallelism
		// return: Promise<error|number>
		ComputePoW?: (progress_id: string, challenge: Uint8Array<ArrayBuffer>, salt: Uint8Array<ArrayBuffer>, memory_mb: number, iterations: number, parallelism: number) => Promise<string>;
		
		// expected args: captcha_challenge_numeric, session_token_ciphered, session_id
		// return: Promise<string> which is the decipehred session_token
		CheckoutCaptcha?: (captcha_solution_numeric: number, session_token_ciphered: Uint8Array<ArrayBuffer>, session_id: Uint8Array<ArrayBuffer>) => Promise<string>;
	}
}

self.onmessage = async (event: MessageEvent) => {
	if (!initialized || event.data.type === 'init') {
		try {
			go = new Go();
			const result = await WebAssembly.instantiateStreaming(fetch('/umbra.wasm'), go.importObject);
			go.run(result.instance);
			await new Promise(resolve => setTimeout(resolve, 100)); // Ensure WASM is fully initialized
			if ((globalThis as any).umbraReady?.() !== "Umbra WASM initialized") {
				throw new Error("WASM module did not initialize correctly");
			}
			initialized = true;
			console.log("WASM module initialized successfully");
			if (event.data.type === 'init') {
				self.postMessage({ type: 'init', success: true });
			}
		} catch (err) {
			console.error('Error initializing WASM module:', err);
			if (event.data.type === 'init') {
				self.postMessage({ type: 'init', success: false, error: err });
			}
			return;
		}
	}
	if (event.data.type === 'setBaseURL') {	
		try {
			if (typeof event.data.url !== 'string') {
				throw new Error("Invalid base URL: must be a string");
			}
			baseURL = new URL(event.data.url);
			self.postMessage({ type: 'setBaseURL', success: true, url: baseURL.toString() });
		} catch (err) {
			console.error('Error setting base URL:', err);
			self.postMessage({ type: 'setBaseURL', success: false, error: err });
		}
	} else if (event.data.type === 'SessionKeypair') {
		let request_payload = ""
		try {
			const pubkeys = await self.SessionKeypair?.()
			if (typeof pubkeys !== 'object') {
				throw new Error("SessionKeypair did not return a valid result");
			}

			await storeSecret("session_soul", pubkeys.soul);

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

				const server_ed_pubkey_bytes = decodeBase64(server_ed_pubkey);
				const server_x_pubkey_bytes = decodeBase64(server_x_pubkey);
				const server_x_pubkey_sign_bytes = decodeBase64(server_x_pubkey_sign);

				await storeSecret("server_ed_pubkey", server_ed_pubkey_bytes);
				await storeSecret("server_x_pubkey", server_x_pubkey_bytes);
				await storeSecret("server_x_pubkey_sign", server_x_pubkey_sign_bytes);

				const soul = await retrieveSecret("session_soul");
				if (!soul) {
					throw new Error("Session soul not found in vault");
				}
				
				const deciphered_payload = await self.IntroduceServer?.(soul, server_ed_pubkey, server_x_pubkey, server_x_pubkey_sign, payload, signature);

				// Remove sensitive data from memory
				soul.fill(0);

				if (!deciphered_payload || typeof deciphered_payload !== 'object' || !('session_token_ciphered' in deciphered_payload)) {
					throw new Error("IntroduceServer did not return a valid session token ciphered");
				}

				await storeSecret("session_token_ciphered", decodeBase64(deciphered_payload.session_token_ciphered));
				await storeSecret("session_id", decodeBase64(deciphered_payload.session_id));

				deciphered_payload.session_token_ciphered = "";
				deciphered_payload.session_id = "";

				self.postMessage({ type: 'IntroduceServer', success: true, payload: deciphered_payload });
			} catch (err) {
				console.error('Error during sending session initialization:', err);
				self.postMessage({ type: 'SendSessionKeypair', success: false, error: err });
			}
		} catch (err) {
			console.error('Error during session key pair generation:', err);
			self.postMessage({ type: 'SessionKeypair', success: false, error: err });
		}
	} else if (event.data.type === 'PoW') {
		self.ComputePoW?.(
			event.data.progress_id,
			new Uint8Array(event.data.challenge),
			new Uint8Array(event.data.salt),
			event.data.memory_mb,
			event.data.iterations,
			event.data.parallelism
		)?.then((result: string) => {;
			if (typeof result !== 'number') {
				throw new Error("ComputePoW did not return a valid result");
			}
			
			self.postMessage({ type: 'PoW', success: true, result });
		})?.catch((err: Error) => {
			console.error('Error during PoW computation:', err);
			self.postMessage({ type: 'PoW', success: false, error: err.message });
		});
	} else if (event.data.type === 'CheckoutCaptcha') {
		const { captcha_response } = event.data;
		if (typeof captcha_response !== 'string' || captcha_response.length !== 6 || !/^\d{6}$/.test(captcha_response)) {
			console.warn("Invalid CAPTCHA response:", captcha_response);
			self.postMessage({ type: 'CheckoutCaptcha', success: false, error: "Invalid CAPTCHA response" });
			return;
		}
		const captcha_response_numeric = parseInt(captcha_response);

		try {
			const session_token_ciphered = await retrieveSecret("session_token_ciphered");
			if (!session_token_ciphered) {
				throw new Error("Session token ciphered not found in vault");
			}
			const session_id = await retrieveSecret("session_id");
			if (!session_id) {
				throw new Error("Session ID not found in vault");
			}

			// Decipher Session Token
			const session_token = await self.CheckoutCaptcha?.(captcha_response_numeric, session_token_ciphered, session_id);
			if (typeof session_token !== 'string') {
				throw new Error("CheckoutCaptcha did not return a valid session token");
			}

			console.log(session_token)

			await storeSecret("session_token", decodeBase64(session_token))
			
			self.postMessage({ type: 'CheckoutCaptcha', success: true });
		} catch (err) {
			console.error('Error retrieving session token ciphered:', err);
			self.postMessage({ type: 'CheckoutCaptcha', success: false, error: err });
			return;
		}
	} else {
		console.warn('Unknown message type:', event.data.type);
	}
}


self.onProgressMade = (type: string, id: string, percentage: number) => {
	self.postMessage({ type: 'progress', success: true, progressType: type, id, percentage });
}
