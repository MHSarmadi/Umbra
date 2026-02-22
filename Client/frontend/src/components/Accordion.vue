<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';

interface Props {
	collapsedHeight?: number;
	initiallyExpanded?: boolean;
	showMoreText?: string;
	showLessText?: string;
}

const props = withDefaults(defineProps<Props>(), {
	collapsedHeight: 260,
	initiallyExpanded: false,
	showMoreText: 'Show More',
	showLessText: 'Show Less'
});

const expanded = ref(props.initiallyExpanded);
const contentBodyRef = ref<HTMLElement | null>(null);
const contentHeight = ref(0);
let resizeObserver: ResizeObserver | null = null;

const canCollapse = computed<boolean>(() => contentHeight.value > props.collapsedHeight + 4);

const currentMaxHeight = computed<string>(() => {
	if (!canCollapse.value || expanded.value) {
		return `${contentHeight.value + 2}px`;
	}
	return `${props.collapsedHeight}px`;
});

const isCollapsed = computed<boolean>(() => canCollapse.value && !expanded.value);
const canToggle = computed<boolean>(() => canCollapse.value);

function updateHeight() {
	if (!contentBodyRef.value) {
		contentHeight.value = 0;
		return;
	}
	contentHeight.value = contentBodyRef.value.scrollHeight;
}

function toggleExpansion() {
	if (!canCollapse.value) {
		return;
	}
	expanded.value = !expanded.value;
}

function collapse() {
	if (!canCollapse.value) {
		return;
	}
	expanded.value = false;
}

function onContainerClick(event: MouseEvent) {
	if (!canCollapse.value) {
		return;
	}
	if (isCollapsed.value) {
		expanded.value = true;
		return;
	}
	if (event.ctrlKey || event.metaKey) {
		collapse();
	}
}

function onDoubleClick() {
	toggleExpansion();
}

watch(
	() => expanded.value,
	async () => {
		await nextTick();
		updateHeight();
	}
);

onMounted(async () => {
	await nextTick();
	updateHeight();
	resizeObserver = new ResizeObserver(() => {
		updateHeight();
	});
	if (contentBodyRef.value) {
		resizeObserver.observe(contentBodyRef.value);
	}
});

onBeforeUnmount(() => {
	resizeObserver?.disconnect();
});
</script>

<template>
	<section
		class="accordion focusable"
		:class="{ expanded: !isCollapsed, collapsed: isCollapsed, collapsible: canToggle }"
		@click="onContainerClick"
		@dblclick="onDoubleClick"
	>
		<div class="accordion-shell">
			<div class="accordion-viewport">
				<div class="accordion-content" :style="{ maxHeight: currentMaxHeight }">
					<div ref="contentBodyRef" class="accordion-content-body">
						<slot />
					</div>
				</div>
				<div class="fade-shadow" :class="{ hidden: !isCollapsed }"></div>
			</div>
			<div v-if="canCollapse" class="accordion-footer">
				<button
					type="button"
					class="accordion-toggle"
					:aria-expanded="expanded"
					@click.stop="toggleExpansion"
				>
					{{ expanded ? showLessText : showMoreText }}
				</button>
			</div>
		</div>
	</section>
</template>

<style scoped lang="less">
@import "../style.less";

.accordion {
	margin: 12px 0 18px;
	border-radius: var(--border-radius-lg);
	background:
		radial-gradient(120% 120% at 0% 0%, #26a69a29 0%, transparent 62%),
		radial-gradient(130% 120% at 100% 100%, #007fff2f 0%, transparent 68%),
		linear-gradient(155deg, #101313 0%, #0a0f10 100%);
	border: 1px solid #5fcab24d;
	box-shadow:
		0 10px 26px #0000004f,
		inset 0 1px 0 #ffffff18;
	transition:
		transform 0.22s ease,
		border-color 0.22s ease,
		box-shadow 0.22s ease,
		background-position 0.5s ease;
	background-size: 150% 150%, 140% 140%, 120% 120%;
	animation: accordion-gradient-drift 11s ease-in-out infinite alternate;

	&.collapsible {
		cursor: pointer;
	}

	&:hover {
		transform: translateY(-1px);
		border-color: #7de6d783;
		box-shadow:
			0 14px 30px #00000069,
			0 0 0 1px #4db6ac42,
			0 0 20px #4db6ac1f,
			inset 0 1px 0 #ffffff22;
	}

	&:active {
		transform: translateY(1px) scale(0.998);
	}
}

.accordion-shell {
	position: relative;
	border-radius: inherit;
	padding: 14px 14px 0;
}

.accordion-viewport {
	position: relative;
	border-radius: calc(var(--border-radius-lg) - 3px);
	overflow: hidden;
	background: linear-gradient(180deg, #ffffff03 0%, #00000026 100%);
}

.accordion-content {
	overflow: hidden;
	transition: max-height 0.5s cubic-bezier(0.18, 0.85, 0.28, 1);
	will-change: max-height;
}

.accordion-content-body {
	display: flow-root;
	position: relative;
	padding: 0 10px 12px;
}

.accordion.collapsed .accordion-content-body {
	user-select: none;
	-webkit-user-select: none;
}

.accordion.expanded .accordion-content-body {
	user-select: text;
	-webkit-user-select: text;
	cursor: text;
}

.fade-shadow {
	position: absolute;
	left: 0;
	right: 0;
	bottom: 0;
	height: 72px;
	pointer-events: none;
	background: linear-gradient(180deg, transparent 0%, #0c1112cc 62%, #0c1112f7 100%);
	opacity: 1;
	transition: opacity 0.24s ease;

	&.hidden {
		opacity: 0;
	}
}

.accordion-footer {
	display: flex;
	justify-content: center;
	padding: 10px 0 14px;
}

.accordion-toggle {
	margin: 0;
	min-width: 128px;
	padding: 8px 14px;
	border-radius: var(--border-radius-md);
	border: 1px solid #66cfc164;
	color: var(--text-color);
	background:
		linear-gradient(135deg, #4db6ac3a 0%, #00897b4d 42%, #007fff30 100%),
		rgba(15, 27, 25, 0.5);
	box-shadow:
		0 8px 16px #00000058,
		inset 0 1px 0 #ffffff2f,
		inset 0 -1px 0 #00000035;
	cursor: pointer;
	transition:
		transform 0.22s ease-in-out,
		box-shadow 0.22s ease,
		filter 0.22s ease,
		background-color 0.22s ease,
		border-color 0.22s ease;
	.bold();

	&:hover {
		transform: translateY(-1px) scale(1.01);
		border-color: #7de6d789;
		filter: saturate(1.1);
		box-shadow:
			0 13px 22px #00000066,
			0 0 0 1px #4db6ac3a,
			inset 0 1px 0 #ffffff3a,
			inset 0 -1px 0 #0000002e;
	}

	&:active {
		transform: translateY(1px) scale(0.985);
		filter: saturate(0.95);
	}

	&:focus-visible {
		outline: 2px solid #4db6ac8f;
		outline-offset: 2px;
	}
}

.accordion:hover .accordion-toggle,
.accordion:focus-within .accordion-toggle {
	transform: translateY(-1px) scale(1.01);
	border-color: #7de6d789;
	filter: saturate(1.1);
	box-shadow:
		0 13px 22px #00000066,
		0 0 0 1px #4db6ac3a,
		inset 0 1px 0 #ffffff3a,
		inset 0 -1px 0 #0000002e;
}

@keyframes accordion-gradient-drift {
	0% {
		background-position: 0% 0%, 100% 100%, 50% 50%;
	}

	50% {
		background-position: 40% 30%, 80% 45%, 62% 70%;
	}

	100% {
		background-position: 100% 100%, 10% 5%, 40% 25%;
	}
}
</style>
