/// <reference lib="webworker" />

import WasmWorker from "./wasm-worker?worker";

class WorkerInstance {
	id: number;
	worker: Worker;
	busy: boolean;
	router: Map<string, (data: any) => void>;
	onload: () => void;

	constructor(worker: Worker, id: number, then: () => void = () => {}) {
		this.id = id
		this.worker = worker;
		this.busy = false;
		this.router = new Map();
		this.onload = then

		this.router.set("init", () => {
			console.log("WASM Worker initialized successfully")
			this.onload()
		})
		this.router.set("freed", (data: any) => {
			console.log(this.id, "FREED:", data.processType)
			this.busy = false;
		});

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

	post(message: any) {
		console.log(this.id, "SEND", message)
		if (this.busy) {
			return false;
		}
		this.busy = true;
		this.worker.postMessage(message);
		return true;
	}
}
const workerPool: WorkerInstance[] = [];

let workerCounter = 0;
async function createWorker(): Promise<WorkerInstance> {
	++workerCounter
	console.log("===== GENERATING WORKER " + workerCounter + " =====")
	const worker = new WasmWorker();
	const workerInstance = new WorkerInstance(worker, workerCounter);
	await new Promise(resolve => setTimeout(resolve, 1000)); // Ensure Worker is fully initialized
	workerInstance.post({ type: 'init' })
	await new Promise(resolve => setTimeout(resolve, 1000)); // Ensure Prev post is fully computed
	workerInstance.post({
		type: 'setBaseURL',
		url: "http://localhost:8888" // TODO: Replace with environment variable
	})
	await new Promise(resolve => setTimeout(resolve, 1000)); // Ensure Prev post is fully computed
	workerPool.push(workerInstance);
	return workerInstance;
}
createWorker();

async function post(message: any) {
	for (const workerInstance of workerPool) {
		if (!workerInstance.busy) {
			workerInstance.post(message);
			return;
		}
	}
	(await createWorker()).post(message);
}

self.onmessage = async (event: MessageEvent) => {
	if (!event.data || !event.data.type) {
		console.warn("Received message without type:", event.data);
		return;
	}

	await post(event.data);
}