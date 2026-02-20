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
	position: relative;
	isolation: isolate;
	overflow: hidden;
	color: var(--text-color);
	border: 1px solid #66cfc152;
	border-radius: var(--border-radius-lg);
	padding: 12px 28px;
	font-size: 1.3em;
	.bold();
	cursor: pointer;
	background:
		linear-gradient(135deg, #4db6ac33 0%, #00897b45 38%, #007fff2e 100%),
		rgba(15, 27, 25, 0.42);
	backdrop-filter: blur(8px) saturate(118%);
	-webkit-backdrop-filter: blur(8px) saturate(118%);
	box-shadow:
		0 10px 24px #0000005e,
		inset 0 1px 0 #ffffff3a,
		inset 0 -1px 0 #00000042;
	transition:
		transform 0.24s ease-in-out,
		box-shadow 0.24s ease,
		filter 0.24s ease,
		background-color 0.24s ease,
		border-color 0.24s ease,
		backdrop-filter 0.24s ease;

	&::before {
		content: "";
		position: absolute;
		inset: 0;
		background: linear-gradient(120deg, transparent 0%, #ffffff1f 44%, transparent 100%);
		transform: translateX(-130%);
		z-index: -1;
	}

	&::after {
		content: "";
		position: absolute;
		inset: 0;
		background: radial-gradient(75% 90% at 15% 0%, #4db6ac2e 0%, transparent 72%);
		z-index: -2;
		transition: opacity 0.24s ease;
	}

	&:hover {
		transform: translateY(-1px) scale(100.5%);
		filter: saturate(1.12);
		border-color: #7de6d783;
		background:
			linear-gradient(135deg, #4db6ac45 0%, #00897b62 42%, #007fff42 100%),
			rgba(19, 38, 34, 0.58);
		backdrop-filter: saturate(128%);
		-webkit-backdrop-filter: blur(11px) saturate(128%);
		box-shadow:
			0 16px 32px #00000069,
			0 0 0 1px #4db6ac42,
			inset 0 1px 0 #ffffff47,
			inset 0 -1px 0 #00000033;

		&::before {
			transition: transform 0.6s ease;
			transform: translateX(130%);
		}

		&::after {
			opacity: 1.2;
		}
	}

	&:active {
		transform: translateY(1px) scale(0.985);
		filter: saturate(0.94);
		border-color: #4db6ac75;
		background:
			linear-gradient(135deg, #4db6ac2e 0%, #00897b45 40%, #007fff2a 100%),
			rgba(12, 20, 19, 0.36);
		backdrop-filter: blur(6px) saturate(105%);
		-webkit-backdrop-filter: blur(6px) saturate(105%);
		box-shadow:
			0 8px 14px #00000052,
			inset 0 1px 0 #ffffff2a,
			inset 0 -1px 0 #00000052;
	}

	&:focus-visible {
		outline: 2px solid #4db6ac8f;
		outline-offset: 2px;
	}
}
</style>
