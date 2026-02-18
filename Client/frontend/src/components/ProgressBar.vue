<!-- ProgressBar.vue -->
<script lang="ts" setup>
import { computed } from 'vue'

interface Props {
	percentage: number | string
	showLabel?: boolean
	size?: 'small' | 'medium' | 'large'
}

const props = withDefaults(defineProps<Props>(), {
	showLabel: true,
	size: "medium"
})

const clampedPercentage = computed<number>(() => {
	const val = Number(props.percentage)
	if (isNaN(val)) return 0
	return Math.max(0, Math.min(100, val))
})

const displayPercentage = computed<number>(() => {
	return Math.round(clampedPercentage.value)
})

</script>

<template>
	<div class="progress-bar-container loading" :class="{ small: size === 'small', large: size === 'large' }">
		<div class="progress-bar-outer">
			<div class="progress-bar-inner" :style="{
				width: clampedPercentage + '%'
			}"></div>

			<div v-if="showLabel" class="progress-label">
				{{ displayPercentage }}%
			</div>
		</div>
	</div>
</template>

<style lang="less" scoped>
.progress-bar-container {
	width: 300px;
	padding-top: 0.2em;
	display: inline-block;

	.progress-bar-outer {
		position: relative;
		height: 1.1em;
		background: var(--dark-secondary-bg);
		border-radius: var(--border-radius-sm);
		overflow: hidden;
		box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.4);

		.progress-bar-inner {
			height: 100%;
			width: 0;
			background: var(--main-highlight-color-3);
			border-radius: var(--border-radius-sm);
			transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
			box-shadow:
				0 0 8px var(--main-highlight-color-2),
				inset 0 1px 1px rgba(255, 255, 255, 0.15);
		}

		.progress-label {
			position: absolute;
			inset: 0;
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 1em;
			font-weight: 600;
			color: var(--text-color);
			text-shadow: 0 1px 3px #0008;
			letter-spacing: 0.4px;
			mix-blend-mode: hard-light;
			pointer-events: none;
			user-select: none;
		}
	}

	&.loading .progress-bar-inner {
		background: linear-gradient(90deg,
			var(--main-highlight-color) 0%,
			var(--main-highlight-color-2) 20%,
			var(--main-highlight-color-3) 30%,
			var(--main-highlight-color-4) 50%,
			var(--main-highlight-color-3) 70%,
			var(--main-highlight-color-2) 80%,
			var(--main-highlight-color) 100%);
		background-size: 200% 100%;
		animation: loading-pulse 2.2s linear infinite;
	}
}

@keyframes loading-pulse {
	0% {
		background-position: 200% 0;
	}

	100% {
		background-position: -200% 0;
	}
}
</style>