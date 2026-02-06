/// <reference lib="webworker" />

const response = await fetch("/wasm_exec.js");
const wasmExecCode = await response.text();
eval(wasmExecCode); // Defines 'Go' class

let go = null, initialized = false;

declare global {
	interface Window {
		onProgressMade?: (type: string, id: string, percentage: number) => void;
		
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