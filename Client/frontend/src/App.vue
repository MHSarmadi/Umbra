<script setup lang="ts">
import { RouterView } from 'vue-router';
import WasmWorker from './workers/wasm-worker?worker';
import { onMounted, provide, ref } from 'vue';

const wasmWorker = new WasmWorker();
const workerRouter = ref<{[key: string]: (event: MessageEvent) => void}>({});
const progressPercentages = ref<{[key: string]: (id: string) => (percentage: number) => void}>({});
provide('wasmWorker', wasmWorker);
provide('workerRouter', workerRouter);
provide('progressPercentages', progressPercentages);


var interval: number, counter = 0;
workerRouter.value["setBaseURL"] = (event: MessageEvent) => {
	if (!event.data.success && ++counter > 20) {
		console.error('Failed to set base URL:', event.data.error);
	} else if (event.data.success) {
		console.log('Base URL set successfully');
		clearInterval(interval);
	} else {
		console.warn('Failed to set base URL, retrying...', event.data.error);
	}
};
workerRouter.value["progress"] = (event: MessageEvent) => {
	const { progressType, id, percentage } = event.data;
	if (typeof percentage === 'number' && typeof id === 'string') {
		progressPercentages.value[progressType]?.(id)?.(percentage);
	} else {
		console.warn(`Invalid progress message: ${JSON.stringify(event.data)}`);
	}
};

onMounted(() => {
	interval = setInterval(() => {
		// console.log("try to set base URL #" + (counter + 1));
		wasmWorker.postMessage({
			type: 'setBaseURL',
			url: "http://localhost:8888" // TODO: Replace with environment variable
		})
	}, 500);
});

wasmWorker.onmessage = (event: MessageEvent) => {
	if (typeof workerRouter.value[event.data?.type ?? 'default'] === 'function') {
		workerRouter.value[event.data?.type ?? 'default']!(event);
	} else {
		console.warn(`No handler for worker message type: ${event.data?.type ?? 'default'}`);
	}
}


</script>

<template>
	<router-view />
</template>

<style scoped lang="less">

</style>
