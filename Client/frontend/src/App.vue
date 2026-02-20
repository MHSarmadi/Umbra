<script setup lang="ts">
import { RouterView } from 'vue-router';
import WorkerPool from './workers/worker-pool?worker';
import { provide, ref } from 'vue';

const workerPool = new WorkerPool();
const workerRouter = ref<{ [key: string]: (event: MessageEvent) => void }>({});
const progressPercentages = ref<{ [key: string]: (id: string) => (percentage: number) => void }>({});
provide('worker-pool', workerPool);
provide('workerRouter', workerRouter);
provide('progressPercentages', progressPercentages);

workerRouter.value["progress"] = (event: MessageEvent) => {
	const { progressType, id, percentage } = event.data;
	if (typeof percentage === 'number' && typeof id === 'string') {
		progressPercentages.value[progressType]?.(id)?.(percentage);
	} else {
		console.warn(`Invalid progress message: ${JSON.stringify(event.data)}`);
	}
};


workerPool.onmessage = (event: MessageEvent) => {
	if (typeof workerRouter.value[event.data?.type ?? 'default'] === 'function') {
		workerRouter.value[event.data?.type ?? 'default']!(event);
	} else {
		console.warn(`No handler for worker message type: ${event.data?.type ?? 'default'}`);
	}
}

console.log(
	"%c⚠️  DANGER ZONE  ⚠️\n" +
	"STOP pasting random code here!\n" +
	"It can STEAL your private keys, tokens & passwords\n" +
	"Only use code you personally understand and trust 100%",
	"color:#ffeb3b; background:#d32f2f; font-size:18px; font-weight:bold; " +
	"padding:8px 16px; line-height:1.6; border-left:8px solid #ffeb3b; display:block;"
);

console.log("%cNEVER PASTE ANYTHING HERE UNLESS YOU WROTE IT YOURSELF",
	"color:#ff1744; background-color:#ffeb3b; padding:8px 16px; font-size:17px; font-weight:900; line-height:1.6;"
);


</script>

<template>
	<router-view />
</template>

<style scoped lang="less"></style>
