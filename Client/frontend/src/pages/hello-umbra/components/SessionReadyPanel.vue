<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import LargeButton from '../../../components/LargeButton.vue';
import Accordion from '../../../components/Accordion.vue';

const props = defineProps<{
	sessionExpiry: Date | null;
	sessionExpiredNotice: boolean;
	reloadCountdown: number | null;
}>();
const emit = defineEmits<{
	(event: 'session-expired'): void;
}>();

const nowMs = ref(Date.now());
let ticker: ReturnType<typeof setInterval> | null = null;
const didEmitExpiry = ref(false);

const expiryMs = computed(() => props.sessionExpiry?.getTime() ?? null);
const remainingMs = computed(() => {
	if (expiryMs.value === null) {
		return null;
	}
	return Math.max(0, expiryMs.value - nowMs.value);
});

const isExpired = computed(() => remainingMs.value !== null && remainingMs.value <= 0);
const isUnderOneMinute = computed(() => remainingMs.value !== null && remainingMs.value > 0 && remainingMs.value < 60000);

const remainingTimeLabel = computed(() => {
	if (remainingMs.value === null) {
		return '--:--:--';
	}

	const totalSeconds = Math.floor(remainingMs.value / 1000);
	const hours = Math.floor(totalSeconds / 3600);
	const minutes = Math.floor((totalSeconds % 3600) / 60);
	const seconds = totalSeconds % 60;

	return [hours, minutes, seconds].map((part) => String(part).padStart(2, '0')).join(':');
});
const remainingShortTimeLabel = computed(() => {
	if (remainingMs.value === null) {
		return '--:--';
	}

	const totalSeconds = Math.floor(remainingMs.value / 1000);
	const hours = Math.floor(totalSeconds / 3600);
	const minutes = Math.floor((totalSeconds % 3600) / 60);
	const seconds = totalSeconds % 60;

	return [hours, minutes, seconds].map((part) => String(part).padStart(2, '0')).filter((n,i) => i > 0 || n !== '00').join(':');
});

const roundedTenMinuteLabel = computed(() => {
	if (remainingMs.value === null) {
		return 'Expiration time unavailable';
	}
	if (props.sessionExpiredNotice || remainingMs.value <= 0) {
		return 'Session expired';
	}

	const totalMinutes = Math.round(remainingMs.value / 60000);
	const rounded = Math.floor(totalMinutes);
	return `After ~${rounded} minutes you will no longer be able to use this session. Please make sure to Login or Signup before then to avoid repeating the initialization process.`;
});

const reloadNoticeLabel = computed(() => {
	if (!props.sessionExpiredNotice || props.reloadCountdown === null) {
		return '';
	}
	return `This page will reload in ${props.reloadCountdown}s.`;
});

watch(remainingMs, (value) => {
	if (value === null || value > 0 || didEmitExpiry.value) {
		return;
	}
	didEmitExpiry.value = true;
	emit('session-expired');
}, { immediate: true });

onMounted(() => {
	ticker = setInterval(() => {
		nowMs.value = Date.now();
	}, 1000);
});

onUnmounted(() => {
	if (ticker) {
		clearInterval(ticker);
		ticker = null;
	}
});
</script>

<template>
	<div class="session-ready page-panel animate-in">
		<h1 style="margin-top: 0;">
			<svg class="inline" fill="#e0e0e0">
				<use href="../../../assets/check2.svg" />
			</svg>
			Session Ready
		</h1>
		<div class="expiry-banner" :class="{ expired: isExpired, urgent: isUnderOneMinute }">
			<p class="expiry-banner-title">Session timeout</p>
			<p class="expiry-banner-timer">{{ remainingTimeLabel }}</p>
			<p class="expiry-banner-rounded">{{ roundedTenMinuteLabel }}</p>
			<p v-if="sessionExpiredNotice" class="expiry-banner-warning">
				This session has expired. Please initialize a new session to continue.
			</p>
			<p v-if="sessionExpiredNotice" class="expiry-banner-reload">{{ reloadNoticeLabel }}</p>
		</div>

		<p>
			Your secure session has been established successfully. You are ready for the next part of the flow.
			No account is attached yet. The session is active yet, but not authenticated.
		</p>
		<h4 class="warning">
			Please consider logging in or signing up quickly before it expires.
		</h4>
		<h2>Next Step</h2>
		<p>You may now:</p>
		<ul>
			<li>Login to an existing Umbra account</li>
			<li>Sign Up to create a new account</li>
		</ul>
		<large-button style="margin-right: 15px">Login</large-button>
		<large-button>Signup</large-button>
		<h2>Technical Notes</h2>
		<accordion :collapsed-height="220">
			<p>
				The initialization phase is complete. Temporary session keys have been generated,
				and this device is now operating within a protected communication channel. All further interactions
				with Umbra will take place inside this secured context.
			</p>
			<p>
				At this point, you only have a temporary session that is not yet associated with any account.
				This session is going to be expired in <b>{{ remainingShortTimeLabel }}</b> if you do not proceed with authentication.
			</p>
			<p>
				Once authentication is completed, this session becomes cryptographically bound to the account you
				choose.
				From that point forward, all actions performed during this session are associated exclusively with
				that identity.
			</p>
			<hr />
			<h2>Session Binding Model</h2>
			<q class="block">Umbra does not reuse sessions across logins.</q>
			<p>
				If you log out, the current session is immediately terminated.
				<br />
				To authenticate again — even with the same account — a new secure session must be initialized.
			</p>
			<p>This ensures that:</p>
			<ul>
				<li>Each authentication begins with fresh cryptographic state</li>
				<li>Sessions remain isolated between accounts</li>
				<li>No residual security context persists after logout</li>
			</ul>
		</accordion>
	</div>
</template>

<style scoped lang="less">
@import url(../../../style.less);

.page-panel {
	position: relative;
	overflow: hidden;
}

.session-ready {
	background:
		radial-gradient(120% 90% at 100% -10%, #26a69a66 0%, transparent 62%),
		radial-gradient(100% 100% at -10% 105%, #007fff4d 0%, transparent 72%),
		linear-gradient(165deg, #151515 0%, #0d0d0d 100%);
	@w1: calc(100vw - 60px);
	@w2: max(40vw, 640px);
	width: min(@w1, @w2);
	@h1: calc(100dvh - 60px);
	@h2: max(40dvh, 800px);
	height: min(@h1, @h2);
	padding: 20px;
	.scroll-container();
	border-radius: var(--border-radius-lg);
	box-shadow: 0 18px 42px #00000066;
	background-size: 150% 150%, 130% 130%, 140% 140%;
	animation: panel-gradient-drift 13s ease-in-out infinite alternate;

	h1 {
		border-bottom: 0;
		padding-bottom: 0;
	}
	
	.warning {
		margin: 22px 0 12px;
		padding: 12px 14px;
		border: 1px solid #e6c300cc;
		border-radius: 14px;
		background: linear-gradient(125deg, #2a1f0f9c 0%, #2a1f0fcc 100%);
		box-shadow: inset 0 0 0 1px #ffb3b314;
		color: #ffd2d2;
	}
}


.expiry-banner {
	margin: 12px 0 22px;
	padding: 12px 14px;
	border: 1px solid #26a69ab3;
	border-radius: 14px;
	background: linear-gradient(125deg, #0f2a2b9c 0%, #0f202dcc 100%);
	box-shadow: inset 0 0 0 1px #ffffff14;
}

.expiry-banner.expired {
	border-color: #e65555cc;
	background: linear-gradient(125deg, #3211119c 0%, #2a0f0fcc 100%);
}

.expiry-banner.urgent {
	border-color: #e65555cc;
	background: linear-gradient(125deg, #3211119c 0%, #2a0f0fcc 100%);
}

.expiry-banner-title {
	margin: 0;
	font-size: 0.85rem;
	text-transform: uppercase;
	letter-spacing: 0.08em;
	color: #a9d8d3;
}

.expiry-banner-timer {
	margin: 4px 0;
	font-size: clamp(1.2rem, 1.45vw, 1.55rem);
	font-weight: 700;
	font-variant-numeric: tabular-nums;
	color: #f6fffd;
}

.expiry-banner-rounded {
	margin: 0;
	font-size: 0.88rem;
	color: #c8e7e4;
}

.expiry-banner-warning {
	margin: 8px 0 0;
	font-size: 0.92rem;
	font-weight: 600;
	color: #ffd2d2;
}

.expiry-banner-reload {
	margin: 4px 0 0;
	font-size: 0.82rem;
	color: #ffb3b3;
}

.animate-in {
	& > * {
		opacity: 0;
		animation: content-rise 0.5s ease forwards;
	}

	& > :nth-child(1) { animation-delay: 0.05s; }
	& > :nth-child(2) { animation-delay: 0.1s; }
	& > :nth-child(3) { animation-delay: 0.15s; }
	& > :nth-child(4) { animation-delay: 0.2s; }
	& > :nth-child(5) { animation-delay: 0.25s; }
	& > :nth-child(6) { animation-delay: 0.3s; }
	& > :nth-child(7) { animation-delay: 0.35s; }
	& > :nth-child(8) { animation-delay: 0.4s; }
	& > :nth-child(9) { animation-delay: 0.45s; }
	& > :nth-child(10) { animation-delay: 0.5s; }
	& > :nth-child(11) { animation-delay: 0.55s; }
	& > :nth-child(12) { animation-delay: 0.6s; }
	& > :nth-child(13) { animation-delay: 0.65s; }
	& > :nth-child(14) { animation-delay: 0.7s; }
}

@keyframes content-rise {
	from {
		opacity: 0;
		transform: translateY(12px);
	}

	to {
		opacity: 1;
		transform: translateY(0);
	}
}

@keyframes panel-gradient-drift {
	0% {
		background-position: 0% 0%, 100% 100%, 50% 50%;
	}

	50% {
		background-position: 35% 45%, 70% 40%, 60% 75%;
	}

	100% {
		background-position: 100% 100%, 0% 0%, 45% 20%;
	}
}
</style>
