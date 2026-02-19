/// <reference lib="webworker" />

import WasmWorker from "./wasm-worker?worker";

function collectTransferables(message: unknown): Transferable[] {
	if (!message || typeof message !== "object") {
		return [];
	}
	const seen = new Set<ArrayBuffer>();
	const transferables: Transferable[] = [];
	for (const value of Object.values(message)) {
		if (value instanceof ArrayBuffer) {
			if (!seen.has(value)) {
				seen.add(value);
				transferables.push(value);
			}
			continue;
		}
		if (!ArrayBuffer.isView(value)) {
			continue;
		}
		const buffer = value.buffer;
		if (buffer instanceof ArrayBuffer && !seen.has(buffer)) {
			seen.add(buffer);
			transferables.push(buffer);
		}
	}
	return transferables;
}

class WorkerInstance {
	id: number;
	worker: Worker;
	busy: boolean;
	router: Map<string, (data: any) => void>;
	ready: boolean;
	private readonly baseURL: string;
	private readonly initPromise: Promise<void>;
	private readonly baseURLPromise: Promise<void>;
	private readonly startupTimeoutMs: number;
	private startupFailed: boolean;
	private resolveInit!: () => void;
	private rejectInit!: (reason?: any) => void;
	private resolveBaseURL!: () => void;
	private rejectBaseURL!: (reason?: any) => void;
	private readonly onFreed: () => void;
	private readonly onFatal: (reason: Error) => void;

	constructor(worker: Worker, id: number, baseURL: string, onFreed: () => void, onFatal: (reason: Error) => void, startupTimeoutMs = 90000) {
		this.id = id
		this.worker = worker;
		this.busy = false;
		this.ready = false;
		this.baseURL = baseURL;
		this.startupTimeoutMs = startupTimeoutMs;
		this.startupFailed = false;
		this.onFreed = onFreed;
		this.onFatal = onFatal;
		this.router = new Map();
		this.initPromise = new Promise((resolve, reject) => {
			this.resolveInit = resolve;
			this.rejectInit = reject;
		});
		this.baseURLPromise = new Promise((resolve, reject) => {
			this.resolveBaseURL = resolve;
			this.rejectBaseURL = reject;
		});

		this.router.set("init", (data: any) => {
			this.busy = false;
			if (data?.success) {
				console.log(`WASM Worker ${this.id} initialized successfully`);
				this.resolveInit();
				return;
			}
			this.rejectInit(new Error(`WASM worker ${this.id} init failed: ${String(data?.error ?? "Unknown error")}`));
		})
		this.router.set("setBaseURL", (data: any) => {
			this.busy = false;
			if (data?.success) {
				console.log(`WASM Worker ${this.id} base URL set successfully`);
				this.resolveBaseURL();
				return;
			}
			this.rejectBaseURL(new Error(`WASM worker ${this.id} setBaseURL failed: ${String(data?.error ?? "Unknown error")}`));
		})
		this.router.set("freed", (data: any) => {
			console.log(this.id, "FREED:", data.processType)
			this.busy = false;
			this.onFreed();
		});

		this.worker.onerror = (event: ErrorEvent) => {
			this.handleFatal(new Error(`Worker ${this.id} runtime error: ${event.message}`));
		};
		this.worker.onmessageerror = () => {
			this.handleFatal(new Error(`Worker ${this.id} message deserialization failed`));
		};

		this.worker.onmessage = (event: MessageEvent) => {
			console.log(this.id, "RECEIVE", event.data)
			const { type } = event.data;
			if (!type) {
				console.warn("Received message without type:", event.data);
			} else if (this.router.has(type)) {
				this.router.get(type)!(event.data);
			} else {
				self.postMessage(event.data);
			}
		}		
	}

	private failStartup(reason: Error) {
		if (this.startupFailed) {
			return;
		}
		this.startupFailed = true;
		if (!this.ready) {
			this.rejectInit(reason);
			this.rejectBaseURL(reason);
		}
	}

	private handleFatal(reason: Error) {
		this.ready = false;
		this.busy = false;
		this.failStartup(reason);
		this.onFatal(reason);
	}

	post(message: any) {
		console.log(this.id, "SEND", message)
		if (this.busy) {
			return false;
		}
		this.busy = true;
		const transferables = collectTransferables(message);
		this.worker.postMessage(message, transferables);
		return true;
	}

	async ensureReady(): Promise<void> {
		if (this.ready) {
			return;
		}
		const waitWithTimeout = async (promise: Promise<void>, stage: string) => {
			let timeoutHandle: ReturnType<typeof setTimeout> | undefined;
			const timeoutPromise = new Promise<never>((_, reject) => {
				timeoutHandle = setTimeout(() => {
					reject(new Error(`Worker ${this.id} timed out while waiting for ${stage}`));
				}, this.startupTimeoutMs);
			});
			try {
				await Promise.race([promise, timeoutPromise]);
			} finally {
				if (timeoutHandle !== undefined) {
					clearTimeout(timeoutHandle);
				}
			}
		};

		if (!this.post({ type: 'init' })) {
			throw new Error(`Worker ${this.id} is unexpectedly busy during init`);
		}
		await waitWithTimeout(this.initPromise, "init response");
		if (!this.post({ type: 'setBaseURL', url: this.baseURL })) {
			throw new Error(`Worker ${this.id} is unexpectedly busy during setBaseURL`);
		}
		await waitWithTimeout(this.baseURLPromise, "setBaseURL response");
		this.ready = true;
	}
}
const workerPool: WorkerInstance[] = [];
type Job = {
	message: any;
	resolve: () => void;
	reject: (reason?: any) => void;
};
const pendingJobs: Job[] = [];
let schedulerRunning = false;

let workerCounter = 0;
let createWorkerPromise: Promise<WorkerInstance> | null = null;
const BASE_URL = "http://localhost:8888"; // TODO: Replace with environment variable
async function createWorker(): Promise<WorkerInstance> {
	if (createWorkerPromise) {
		return createWorkerPromise;
	}
	createWorkerPromise = (async () => {
	++workerCounter
	const workerId = workerCounter;
	console.log("===== GENERATING WORKER " + workerId + " =====")
	const worker = new WasmWorker();
	const workerInstance = new WorkerInstance(
		worker,
		workerId,
		BASE_URL,
		() => {
			void runScheduler();
		},
		(reason: Error) => {
			const index = workerPool.findIndex((w) => w.id === workerId);
			if (index !== -1) {
				workerPool.splice(index, 1);
			}
			worker.terminate();
			console.error(`Worker ${workerId} removed from pool due to fatal error:`, reason);
			self.postMessage({ type: "WorkerPoolError", success: false, error: String(reason) });
			void runScheduler();
		}
	);
	try {
		await workerInstance.ensureReady();
		workerPool.push(workerInstance);
	} catch (err) {
		worker.terminate();
		throw err;
	}
	return workerInstance;
	})();
	try {
		return await createWorkerPromise;
	} finally {
		createWorkerPromise = null;
	}
}
void createWorker().catch((err) => {
	console.error("Initial worker prewarm failed:", err);
});

async function runScheduler() {
	if (schedulerRunning) {
		return;
	}
	schedulerRunning = true;
	try {
		while (pendingJobs.length > 0) {
			const nextJob = pendingJobs[0];
			if (!nextJob) {
				break;
			}
			let selectedWorker: WorkerInstance | undefined;

			for (const workerInstance of workerPool) {
				if (!workerInstance.ready || workerInstance.busy) {
					continue;
				}
				selectedWorker = workerInstance;
				break;
			}

			if (!selectedWorker) {
				try {
					selectedWorker = await createWorker();
				} catch (err) {
					nextJob.reject(err);
					pendingJobs.shift();
					continue;
				}
			}

			if (!selectedWorker.post(nextJob.message)) {
				continue;
			}
			nextJob.resolve();
			pendingJobs.shift();
		}
	} finally {
		schedulerRunning = false;
	}
}

async function post(message: any): Promise<void> {
	return new Promise<void>((resolve, reject) => {
		pendingJobs.push({ message, resolve, reject });
		void runScheduler();
	});
}

self.onmessage = async (event: MessageEvent) => {
	if (!event.data || !event.data.type) {
		console.warn("Received message without type:", event.data);
		return;
	}

	try {
		await post(event.data);
	} catch (err) {
		console.error("Failed to dispatch message to worker pool:", err);
		self.postMessage({ type: event.data.type, success: false, error: String(err) });
	}
}
