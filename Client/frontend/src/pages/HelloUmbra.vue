<script setup lang="ts">
import { computed, inject, onUnmounted, ref, watch, type Ref } from 'vue';
import { decodeBase64 } from '../tools/base64';
import HelloIntroPanel from './hello-umbra/components/HelloIntroPanel.vue';
import SessionInitPanel from './hello-umbra/components/SessionInitPanel.vue';
import SessionReadyPanel from './hello-umbra/components/SessionReadyPanel.vue';
import { useAuth } from '../auth';
import { encodeBase64 } from "../tools/base64";

const Auth = useAuth();

type PageName = 'Unspecified' | 'HelloUmbra' | 'SessionInit' | 'SessionReady';
type CaptchaPanelState = 'form' | 'verified' | 'none';

type WorkerRouterMap = Record<string, (event: MessageEvent) => void>;
type ProgressMap = Record<string, (id: string) => (percentage: number) => void>;

const workerPool = inject<Worker>('worker-pool');
const workerRouter = inject<Ref<WorkerRouterMap>>('workerRouter');
const progressPercentages = inject<Ref<ProgressMap>>('progressPercentages');

if (!workerPool || !workerRouter || !progressPercentages) {
	throw new Error('HelloUmbra.vue requires worker-related providers from App.vue.');
}

const worker = workerPool;
const router = workerRouter;
const progressByType = progressPercentages;

const currentPage = ref<PageName>('Unspecified');
const sessionExpiry = ref<Date | null>(null);
watch(currentPage, async (newPage) => {
	if (newPage === 'SessionReady') {
		sessionExpiry.value = await Auth.session.expiryUnix();
		console.log('Session expiry time:', sessionExpiry.value);
	}
});
Auth.session.ready().then(ready => {
	if (ready) {
		currentPage.value = 'SessionReady';
	} else {
		currentPage.value = 'HelloUmbra';
	}
}).catch(err => {
	console.error("Failed to check auth session readiness:", err);
	currentPage.value = 'HelloUmbra'; // Fallback to HelloUmbra if auth session check fails
})
Promise.all([Auth.session.id(), Auth.session.token()]).then(([id, token]) => {
	if (id) {
		console.log('Session ID:', encodeBase64(id.value!));
	} else {
		console.warn('No session ID found.');
	}
	if (token) {
		console.log('Session Token:', encodeBase64(token.value!));
	} else {
		console.warn('No session token found.');
	}
})

// 0: keypair_gen, 1: send_to_server, 2: pow, 3: done
const currentStep = ref<number | null>(null);
const stepFailed = ref(false);
const failureMessage = ref('');

const powPercent = ref(0);
const powId = ref('');

const captchaChallengeImage = ref('');
const captchaInput = ref('');
const captchaErrorMsg = ref('');
const captchaLoading = ref(false);
const captchaVerified = ref(false);
const captchaSuccessMsg = ref('');
const showCaptchaBox = ref(true);
let captchaSuccessTimer: ReturnType<typeof setTimeout> | null = null;
const sessionExpiredNotice = ref(false);
const sessionReloadCountdown = ref<number | null>(null);
let sessionReloadTimer: ReturnType<typeof setInterval> | null = null;

const isPowCompleted = computed<boolean>(() => currentStep.value !== null && currentStep.value > 2);
const shouldShowCaptchaBox = computed<boolean>(() => currentStep.value !== null && currentStep.value >= 2 && showCaptchaBox.value);
const captchaPanelState = computed<CaptchaPanelState>(() => {
	if (shouldShowCaptchaBox.value) {
		return 'form';
	}
	if (captchaVerified.value) {
		return 'verified';
	}
	return 'none';
});
const canProceedToNext = computed<boolean>(() => {
	return isPowCompleted.value && captchaPanelState.value === 'verified' && !stepFailed.value;
});

function resetCaptchaSuccessTimer() {
	if (!captchaSuccessTimer) {
		return;
	}
	clearTimeout(captchaSuccessTimer);
	captchaSuccessTimer = null;
}

function resetSessionExpiryFlow() {
	sessionExpiredNotice.value = false;
	sessionReloadCountdown.value = null;
	if (sessionReloadTimer) {
		clearInterval(sessionReloadTimer);
		sessionReloadTimer = null;
	}
}

function failStep(message: string, error?: unknown) {
	stepFailed.value = true;
	failureMessage.value = message;
	if (error) {
		console.error(error);
	}
}

function resetCaptchaState() {
	captchaInput.value = '';
	captchaErrorMsg.value = '';
	captchaLoading.value = false;
	captchaVerified.value = false;
	captchaSuccessMsg.value = '';
	showCaptchaBox.value = true;
	resetCaptchaSuccessTimer();
}

watch(captchaInput, () => {
	if (captchaErrorMsg.value.length) {
		captchaErrorMsg.value = '';
	}
});

router.value.SessionKeypair = (event: MessageEvent) => {
	if (event.data.success && currentStep.value !== 0) {
		console.warn('Not in the right step for SessionKeypair response. Ignoring.');
		return;
	}

	if (event.data.success) {
		currentStep.value = 1;
		return;
	}

	failStep(
		'Session key pair generation failed. Please refresh the page and try again. If the problem persists, please contact support.',
		event.data.error
	);
};

router.value.SendSessionKeypair = (event: MessageEvent) => {
	if (event.data.success && currentStep.value !== 1) {
		console.warn('Not in the right step for SendSessionKeypair response. Ignoring.');
		return;
	}

	if (!event.data.success) {
		failStep(
			'Failed to send session key pair to the server. Please check your internet connection and try again. If the problem persists, please contact support.',
			event.data.error
		);
	}
};

router.value.IntroduceServer = (event: MessageEvent) => {
	if (event.data.success && currentStep.value !== 1) {
		console.warn('Not in the right step for IntroduceServer response. Ignoring.');
		return;
	}

	if (!event.data.success) {
		failStep(
			'Failed to establish a secure connection with the server. Please check your internet connection and try again. If the problem persists, please contact support.',
			event.data.error
		);
		return;
	}

	currentStep.value = 2;
	captchaChallengeImage.value = `data:image/png;base64,${event.data.payload.captcha_challenge}`;
	resetCaptchaState();

	powId.value = Math.floor(Math.random() * 36 ** 8).toString(36);
	const challenge = decodeBase64(event.data.payload.pow_challenge);
	const salt = decodeBase64(event.data.payload.pow_salt);
	worker.postMessage(
		{
			type: 'PoW',
			progress_id: powId.value,
			challenge: challenge.buffer,
			salt: salt.buffer,
			memory_mb: event.data.payload.pow_params.memory_mb,
			iterations: event.data.payload.pow_params.iterations,
			parallelism: event.data.payload.pow_params.parallelism
		},
		[challenge.buffer, salt.buffer]
	);
};

router.value.PoW = (event: MessageEvent) => {
	if (currentStep.value !== 2) {
		console.warn('Not in the right step for PoW response. Ignoring.');
		return;
	}

	if (event.data.success) {
		currentStep.value = 3;
		return;
	}

	failStep(
		'Cryptographic proof of you not being a robot failed. Please refresh the page and try again. If the problem persists, please contact support.',
		event.data.error
	);
};

router.value.CheckoutCaptcha = (event: MessageEvent) => {
	if (!currentStep.value || currentStep.value < 2) {
		console.warn('Not in the right step for Checking out the CAPTCHA. Ignoring.');
		return;
	}

	captchaLoading.value = false;

	if (event.data.success) {
		captchaVerified.value = true;
		captchaErrorMsg.value = '';
		captchaSuccessMsg.value = 'Correct CAPTCHA. Verification complete.';
		captchaInput.value = '';
		resetCaptchaSuccessTimer();
		captchaSuccessTimer = setTimeout(() => {
			showCaptchaBox.value = false;
		}, 2400);
		return;
	}

	if (event.data.error !== 'Wrong captcha solution') {
		console.error('Decrypting the Session Token failed:', event.data.error);
		return;
	}

	captchaVerified.value = false;
	captchaErrorMsg.value = 'Incorrect. Retry?';
	const captchaInputElement = document.getElementById('captcha_input') as HTMLInputElement | null;
	captchaInputElement?.focus();
};

const previousPowProgressFunction = progressByType.value.pow;
progressByType.value.pow = (id: string) => {
	if (id !== powId.value) {
		if (previousPowProgressFunction) {
			return previousPowProgressFunction(id);
		}
		return (_: number) => {
			console.warn(`Received progress update for unknown proof of work session ID ${id}. Ignoring.`);
		};
	}

	return (percentage: number) => {
		powPercent.value = Math.round(percentage * 100) / 100;
	};
};

onUnmounted(() => {
	delete router.value.SessionKeypair;
	delete router.value.SendSessionKeypair;
	delete router.value.IntroduceServer;
	delete router.value.PoW;
	delete router.value.CheckoutCaptcha;

	resetCaptchaSuccessTimer();
	resetSessionExpiryFlow();

	if (previousPowProgressFunction) {
		progressByType.value.pow = previousPowProgressFunction;
	} else {
		delete progressByType.value.pow;
	}
});

function onCaptchaCheckout() {
	if (captchaInput.value.length !== 6 || !/^\d+$/.test(captchaInput.value)) {
		captchaErrorMsg.value = 'Please enter a valid 6-digit numeric code from the CAPTCHA challenge.';
		return;
	}

	captchaLoading.value = true;
	worker.postMessage({
		type: 'CheckoutCaptcha',
		captcha_response: captchaInput.value
	});
}

function goToNextPage() {
	resetSessionExpiryFlow();
	currentPage.value = 'SessionReady';
}

async function onSessionExpired() {
	if (sessionExpiredNotice.value) {
		return;
	}

	sessionExpiredNotice.value = true;
	sessionReloadCountdown.value = 12;

	try {
		await Auth.session.logout();
	} catch (error) {
		console.error('Failed to logout expired session:', error);
	}

	sessionReloadTimer = setInterval(() => {
		if (sessionReloadCountdown.value === null) {
			return;
		}
		sessionReloadCountdown.value -= 1;
		if (sessionReloadCountdown.value <= 0) {
			if (sessionReloadTimer) {
				clearInterval(sessionReloadTimer);
				sessionReloadTimer = null;
			}
			window.location.reload();
		}
	}, 1000);
}

function goToSessionInit() {
	resetSessionExpiryFlow();
	currentPage.value = 'SessionInit';
	currentStep.value = 0;
	stepFailed.value = false;
	failureMessage.value = '';
	powPercent.value = 0;
	resetCaptchaState();

	worker.postMessage({
		type: 'SessionKeypair'
	});
}
</script>

<template>
	<transition name="page-swap" mode="out-in">
		<div :key="currentPage">
			<hello-intro-panel v-if="currentPage === 'HelloUmbra'" @start="goToSessionInit" />
			<session-init-panel v-else-if="currentPage === 'SessionInit'" :current-step="currentStep"
				:step-failed="stepFailed" :failure-message="failureMessage" :pow-percent="powPercent"
				:captcha-panel-state="captchaPanelState" :captcha-challenge-image="captchaChallengeImage"
				:captcha-input="captchaInput" :captcha-error-msg="captchaErrorMsg" :captcha-loading="captchaLoading"
				:captcha-verified="captchaVerified" :captcha-success-msg="captchaSuccessMsg"
				:can-proceed-to-next="canProceedToNext" @update:captcha-input="captchaInput = $event"
				@captcha-checkout="onCaptchaCheckout" @next="goToNextPage" />
			<session-ready-panel v-else-if="currentPage === 'SessionReady'" :session-expiry="sessionExpiry"
				:session-expired-notice="sessionExpiredNotice" :reload-countdown="sessionReloadCountdown"
				@session-expired="onSessionExpired" />
		</div>
	</transition>
</template>

<style scoped lang="less">
.page-swap-enter-active,
.page-swap-leave-active {
	transition: opacity 0.32s ease, transform 0.32s ease;
}

.page-swap-enter-from,
.page-swap-leave-to {
	opacity: 0;
	transform: translateY(10px) scale(0.99);
}
</style>
