/// <reference lib="webworker" />

const response = await fetch("/wasm_exec.js");
const wasmExecCode = await response.text();
eval(wasmExecCode); // Defines 'Go' class

let go = null, initialized = false;

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

	if (event.data.type === 'encrypt') {
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
	}
}