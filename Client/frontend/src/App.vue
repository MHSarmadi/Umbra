<script setup lang="ts">
import { RouterView } from 'vue-router';
import WasmWorker from './workers/wasm-worker?worker';
import { provide, ref } from 'vue';

const wasmWorker = new WasmWorker();
const workerRouter = ref<{[key: string]: (event: MessageEvent) => void}>({});
provide('wasmWorker', wasmWorker);
provide('workerRouter', workerRouter);

wasmWorker.onmessage = (event: MessageEvent) => {
	if (typeof workerRouter.value[event.data?.type ?? 'default'] === 'function') {
		workerRouter.value[event.data?.type ?? 'default']?.(event);
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
