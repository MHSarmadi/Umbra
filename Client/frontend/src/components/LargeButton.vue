<script setup lang="ts">

defineEmits(['click'])

</script>

<template>
	<button class="large-button" @click="$emit('click')">
		<slot></slot>
	</button>
</template>

<style scoped lang="less">
@import "../style.less";

.large-button {
	background-color: var(--main-highlight-color);
	color: var(--text-color);
	border: none;
	border-radius: var(--border-radius-lg);
	padding: 12px 24px;
	font-size: 1.3em;
	.bold();
	cursor: pointer;
	transition: background-color 0.3s ease;
	&:hover {
		background-color: var(--main-highlight-color-2);
	}
	&:active {
		background-color: var(--secondary-highlight-color);
	}
}
</style>
