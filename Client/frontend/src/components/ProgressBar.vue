<!-- ProgressBar.vue -->
<script lang="ts" setup>
import { computed } from 'vue'

interface Props {
	percentage: number | string
	showLabel?: boolean
	size?: 'small' | 'medium' | 'large'
	color?: string
}

const props = withDefaults(defineProps<Props>(), {
	showLabel: true,
	size: "medium",
	color: undefined
})

const clampedPercentage = computed<number>(() => {
	const val = Number(props.percentage)
	if (isNaN(val)) return 0
	return Math.max(0, Math.min(100, val))
})

const displayPercentage = computed<number>(() => {
	return Math.round(clampedPercentage.value)
})

// const progressColor = computed<string>(() => {
// 	return props.color || 'var(--main-highlight-color-3)'
// })
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
	// padding: 4px 0;
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

	/* Optional indeterminate / loading style (add .loading class to container) */
	&.loading .progress-bar-inner {
		background: linear-gradient(90deg,
			var(--main-highlight-color) 0%,
			var(--main-highlight-color-2) 20%,
			var(--main-highlight-color-3) 30%,
			var(--main-highlight-color-4) 50%,
			var(--main-highlight-color-3) 70%,
			var(--main-highlight-color-2) 80%,
			var(--main-highlight-color) 100%);
		// background: linear-gradient(90deg,
		// 	var(--main-color) 0%,
		// 	var(--main-color) 4%,
		// 	var(--main-highlight-color-2) 4.01%,
		// 	var(--main-highlight-color-2) 6%,
		// 	var(--main-color) 6.01%,
		// 	var(--main-color) 10%,
		// 	var(--main-highlight-color-2) 10.01%,
		// 	var(--main-highlight-color-2) 12%,
		// 	var(--main-color) 12.01%,
		// 	var(--main-color) 16%,
		// 	var(--main-highlight-color-2) 16.01%,
		// 	var(--main-highlight-color-2) 18%,
		// 	var(--main-color) 18.01%,
		// 	var(--main-color) 22%,
		// 	var(--main-highlight-color-2) 22.01%,
		// 	var(--main-highlight-color-2) 24%,
		// 	var(--main-color) 24.01%,
		// 	var(--main-color) 28%,
		// 	var(--main-highlight-color-2) 28.01%,
		// 	var(--main-highlight-color-2) 30%,
		// 	var(--main-color) 30.01%,
		// 	var(--main-color) 34%,
		// 	var(--main-highlight-color-2) 34.01%,
		// 	var(--main-highlight-color-2) 36%,
		// 	var(--main-color) 36.01%,
		// 	var(--main-color) 40%,
		// 	var(--main-highlight-color-2) 40.01%,
		// 	var(--main-highlight-color-2) 42%,
		// 	var(--main-color) 42.01%,
		// 	var(--main-color) 46%,
		// 	var(--main-highlight-color-2) 46.01%,
		// 	var(--main-highlight-color-2) 48%,
		// 	var(--main-color) 48.01%,
		// 	var(--main-color) 52%,
		// 	var(--main-highlight-color-2) 52.01%,
		// 	var(--main-highlight-color-2) 54%,
		// 	var(--main-color) 54.01%,
		// 	var(--main-color) 58%,
		// 	var(--main-highlight-color-2) 58.01%,
		// 	var(--main-highlight-color-2) 60%,
		// 	var(--main-color) 60.01%,
		// 	var(--main-color) 64%,
		// 	var(--main-highlight-color-2) 64.01%,
		// 	var(--main-highlight-color-2) 66%,
		// 	var(--main-color) 66.01%,
		// 	var(--main-color) 70%,
		// 	var(--main-highlight-color-2) 70.01%,
		// 	var(--main-highlight-color-2) 72%,
		// 	var(--main-color) 72.01%,
		// 	var(--main-color) 76%,
		// 	var(--main-highlight-color-2) 76.01%,
		// 	var(--main-highlight-color-2) 78%,
		// 	var(--main-color) 78.01%,
		// 	var(--main-color) 82%,
		// 	var(--main-highlight-color-2) 82.01%,
		// 	var(--main-highlight-color-2) 84%,
		// 	var(--main-color) 84.01%,
		// 	var(--main-color) 88%,
		// 	var(--main-highlight-color-2) 88.01%,
		// 	var(--main-highlight-color-2) 90%,
		// 	var(--main-color) 90.01%,
		// 	var(--main-color) 94%,
		// 	var(--main-highlight-color-2) 94.01%,
		// 	var(--main-highlight-color-2) 96%,
		// 	var(--main-color) 96.01%,
		// 	var(--main-color) 100%
		// );
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