/// <reference lib="webworker" />

import { useSecureVault } from "../db";

const { storeSecret } = useSecureVault();

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
		}>
		// expected args: progress_id, challenge, salt, memory_mb, iterations, parallelism
		// return: Promise<error|number>
		ComputePoW?: (progress_id: string, challenge: Uint8Array, salt: Uint8Array, memory_mb: number, iterations: number, parallelism: number) => Promise<string>;
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
				self.postMessage({ type: 'init', success: false, error: (err as Error).message });
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
			// self.postMessage({ type: 'setBaseURL', success: false, error: "" });
		} catch (err) {
			console.error('Error setting base URL:', err);
			self.postMessage({ type: 'setBaseURL', success: false, error: (err as Error).message });
		}
	} else if (event.data.type === 'encrypt') {
		try {
			const { key, data, context, difficulty } = event.data.payload as {
				key?: Uint8Array,
				data?: Uint8Array,
				context?: string,
				difficulty?: number
			};
			if (!key || !(key instanceof Uint8Array)) {
				throw new Error("Invalid key: must be a Uint8Array");
			}
			if (!data || !(data instanceof Uint8Array)) {
				throw new Error("Invalid data: must be a Uint8Array");
			}
			if (typeof context !== 'string') {
				throw new Error("Invalid context: must be a string");
			}
			if (typeof difficulty !== 'number' || difficulty < 0) {
				throw new Error("Invalid difficulty: must be a non-negative number");
			}
			const result = (globalThis as any).MACE_Encrypt(key, data, context, difficulty);
			if (!result || !(result.cipher instanceof Uint8Array) || !(result.salt instanceof Uint8Array)) {
				throw new Error("MACE_Encrypt did not return valid result");
			}
			self.postMessage({
				type: 'encrypt',
				success: true,
				cipher: result.cipher.buffer,
				salt: result.salt.buffer
			}, [
				result.cipher.buffer,
				result.salt.buffer
			]);
		} catch (err) {
			console.error('Error during encryption:', err);
			self.postMessage({ type: 'encrypt', success: false, error: (err as Error).message });
		}
	} else if (event.data.type === 'SessionKeypair') {
		let payload = ""
		try {
			const pubkeys = await self.SessionKeypair?.()
			if (typeof pubkeys !== 'object') {
				throw new Error("SessionKeypair did not return a valid result");
			}
			await storeSecret("session_soul", pubkeys.soul);
			pubkeys.soul.fill(0);
			payload = JSON.stringify({
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
					body: payload
				})
				if (!result.ok) {
					throw new Error(`Failed to send session initialization: ${result.status} ${result.statusText}`);
				}
				self.postMessage({ type: 'SendSessionKeypair', success: result.ok, response: await result.json() });
			} catch (err) {
				console.error('Error during sending session initialization:', err);
				self.postMessage({ type: 'SendSessionKeypair', success: false, error: (err as Error).message });
			}
		} catch (err) {
			console.error('Error during session key pair generation:', err);
			self.postMessage({ type: 'SessionKeypair', success: false, error: (err as Error).message });
		}
	} else if (event.data.type === 'PoW') {
		try {
			const result = await self.ComputePoW?.(
				event.data.progress_id,
				new Uint8Array(event.data.challenge),
				new Uint8Array(event.data.salt),
				event.data.memory_mb,
				event.data.iterations,
				event.data.parallelism
			);
			if (typeof result !== 'number') {
				throw new Error("ComputePoW did not return a valid result");
			}
			self.postMessage({ type: 'PoW', success: true, result });
		} catch (err) {
			console.error('Error during PoW computation:', err);
			self.postMessage({ type: 'PoW', success: false, error: (err as Error).message });
		}
	} else {
		console.warn('Unknown message type:', event.data.type);
	}
}


self.onProgressMade = (type: string, id: string, percentage: number) => {
	self.postMessage({ type: 'progress', success: true, progressType: type, id, percentage });
}