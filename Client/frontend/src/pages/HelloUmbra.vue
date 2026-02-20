<script setup lang="ts">
import { computed, inject, onUnmounted, ref, watch, type Ref } from 'vue';
import LargeButton from '../components/LargeButton.vue';
import ProgressBar from '../components/ProgressBar.vue';
import { decodeBase64 } from '../tools/base64';
import InputField from '../components/InputField.vue';
import jQuery from 'jquery';

const workerPool = inject<Worker>('worker-pool')!;
const workerRouter = inject<Ref<{ [key: string]: (data: any) => void }>>('workerRouter')!;
const progressPercentages = inject<Ref<{ [key: string]: (id: string) => (percentage: number) => void }>>('progressPercentages')!;

type Pages = 'HelloUmbra' | 'SessionInit' | 'SessionReady';
const current_page = ref<Pages>('HelloUmbra');

// 0: 'keypair_gen' | 1: 'send_to_server' | 2: 'pow';
const current_step = ref<number | null>(null)
const step_failed = ref<boolean>(false);
const failure_message = ref<string>('');

const pow_percent = ref<number>(0);
const pow_id = ref<string>('');

const captcha_challenge_image = ref<string>('');
const captcha_input = ref<string>('');
const captcha_error_msg = ref<string>('');
const captcha_loading = ref<boolean>(false);
const captcha_verified = ref<boolean>(false);
const captcha_success_msg = ref<string>('');
const show_captcha_box = ref<boolean>(true);
let captchaSuccessTimer: ReturnType<typeof setTimeout> | null = null;

const isPowCompleted = computed<boolean>(() => current_step.value !== null && current_step.value > 2);
const canProceedToNext = computed<boolean>(() => isPowCompleted.value && captchaPanelState.value === 'verified' && !step_failed.value);
const shouldShowCaptchaBox = computed<boolean>(() => current_step.value !== null && current_step.value >= 2 && show_captcha_box.value);
const captchaPanelState = computed<'form' | 'verified' | 'none'>(() => {
	if (shouldShowCaptchaBox.value) return 'form';
	else if (captcha_verified.value) return 'verified';
	else return 'none';
});

function resetCaptchaSuccessTimer() {
	if (captchaSuccessTimer) {
		clearTimeout(captchaSuccessTimer);
		captchaSuccessTimer = null;
	}
}

watch(captcha_input, () => {
	if (captcha_error_msg.value.length) {
		captcha_error_msg.value = '';
	}
})

workerRouter.value['SessionKeypair'] = (event: MessageEvent) => {
	if (current_step.value !== 0 && event.data.success) {
		console.warn("Not in the right step for SessionKeypair response. Ignoring.")
		return; // Ignore if not in right step
	}
	if (event.data.success) {
		current_step.value = 1;
	} else {
		console.error('Session key pair generation failed:', event.data.error)
		step_failed.value = true;
		failure_message.value = 'Session key pair generation failed. Please refresh the page and try again. If the problem persists, please contact support.';
		// failure_message.value = event.data.error
	}
}

workerRouter.value['SendSessionKeypair'] = (event: MessageEvent) => {
	if (current_step.value !== 1 && event.data.success) {
		console.warn("Not in the right step for SendSessionKeypair response. Ignoring.")
		return; // Ignore if not in right step
	}
	if (event.data.success) {

	} else {
		console.error('Failed to send session key pair:', event.data.error);
		step_failed.value = true;
		failure_message.value = 'Failed to send session key pair to the server. Please check your internet connection and try again. If the problem persists, please contact support.';
	}
}
workerRouter.value['IntroduceServer'] = (event: MessageEvent) => {
	if (current_step.value !== 1 && event.data.success) {
		console.warn("Not in the right step for IntroduceServer response. Ignoring.")
		return; // Ignore if not in right step
	}
	if (event.data.success) {
		current_step.value = 2;
		console.log("Server introduced successfully:", event.data.payload);
		captcha_challenge_image.value = `data:image/png;base64,${event.data.payload.captcha_challenge}`
		show_captcha_box.value = true;
		captcha_verified.value = false;
		captcha_success_msg.value = '';
		resetCaptchaSuccessTimer();
		pow_id.value = Math.floor(Math.random() * 36 ** 8).toString(36); // Generate random ID for this proof of work session
		const challenge = decodeBase64(event.data.payload.pow_challenge), salt = decodeBase64(event.data.payload.pow_salt)
		workerPool.postMessage({
			type: "PoW",
			progress_id: pow_id.value,
			challenge: challenge.buffer,
			salt: salt.buffer,
			memory_mb: event.data.payload.pow_params.memory_mb,
			iterations: event.data.payload.pow_params.iterations,
			parallelism: event.data.payload.pow_params.parallelism
		}, [
			challenge.buffer,
			salt.buffer
		])
	} else {
		console.error('Failed to introduce server during session initialization:', event.data.error);
		step_failed.value = true;
		failure_message.value = 'Failed to establish a secure connection with the server. Please check your internet connection and try again. If the problem persists, please contact support.';
	}
}
workerRouter.value['PoW'] = (event: MessageEvent) => {
	if (current_step.value !== 2) {
		console.warn("Not in the right step for PoW response. Ignoring.")
		return; // Ignore if not in right step
	}
	if (event.data.success) {
		// Session initialization complete, proceed to next page or functionality
		console.log("PoW result:", event.data.result);
		current_step.value = 3;
	} else {
		console.error('Proof of work failed during session initialization:', event.data.error);
		step_failed.value = true;
		failure_message.value = 'Cryptographic proof of you not being a robot failed. Please refresh the page and try again. If the problem persists, please contact support.';
	}
}

workerRouter.value['CheckoutCaptcha'] = (event: MessageEvent) => {
	if (!current_step.value || current_step.value < 2) {
		console.warn("Not in the right step for Checking out the CAPTCHA. Ignoring.");
		return;
	}
	captcha_loading.value = false;
	if (event.data.success) {
		captcha_verified.value = true;
		captcha_error_msg.value = '';
		captcha_success_msg.value = "Correct CAPTCHA. Verification complete.";
		captcha_input.value = '';
		resetCaptchaSuccessTimer();
		captchaSuccessTimer = setTimeout(() => {
			show_captcha_box.value = false;
		}, 2400);
	} else if (event.data.error !== 'Wrong captcha solution') {
		console.error("Decrypting the Session Token failed:", event.data.error);
	} else {
		captcha_verified.value = false;
		captcha_error_msg.value = 'Incorrect. Retry?';
		jQuery("#captcha_input").trigger("focus");
	}
}

const previous_pow_progress_function = progressPercentages.value['pow'];
progressPercentages.value['pow'] = (id: string) => {
	if (id !== pow_id.value) {
		if (previous_pow_progress_function)
			return previous_pow_progress_function?.(id)
		else
			return (_: number) => {
				console.warn(`Received progress update for unknown proof of work session ID ${id}. Ignoring.`);
			};
	}
	return (percentage: number) => {
		pow_percent.value = Math.round(percentage * 100) / 100;
	}
}

onUnmounted(() => {
	delete workerRouter.value['SessionKeypair'];
	delete workerRouter.value['SendSessionKeypair'];
	delete workerRouter.value['IntroduceServer'];
	delete workerRouter.value['PoW'];
	delete workerRouter.value['CheckoutCaptcha'];
	resetCaptchaSuccessTimer();
	if (previous_pow_progress_function) {
		progressPercentages.value['pow'] = previous_pow_progress_function;
	} else {
		delete progressPercentages.value['pow'];
	}
});

function onCaptchaCheckout(value: string | number) {
	if (typeof value !== 'string' || value.length !== 6 || !/^\d+$/.test(value)) {
		captcha_error_msg.value = "Please enter a valid 6-digit numeric code from the CAPTCHA challenge.";
		return;
	}
	captcha_loading.value = true;
	workerPool.postMessage({
		type: 'CheckoutCaptcha',
		captcha_response: value
	});
}

function goto_next_page() {
	current_page.value = 'SessionReady';
}

function goto_session() {
	current_page.value = 'SessionInit';
	current_step.value = 0;

	// STEP 1: Generate Keypair
	workerPool.postMessage({
		type: 'SessionKeypair'
	});
}
</script>

<template>
	<transition name="page-swap" mode="out-in">
		<div :key="current_page">
			<div class="hello-umbra page-panel animate-in" v-if="current_page == 'HelloUmbra'">
				<h1 style="margin-top: 0;">
					<svg class="inline" fill="#e0e0e0">
						<use href="../assets/locked.svg" />
					</svg>
					Welcome to <i style="color: var(--main-highlight-color-3);">Umbra</i>
				</h1>
				<h2>Welcome — and thank you for being here.</h2>
				<p>
					Umbra is a privacy-first communication platform built for people who believe that private
					conversations
					should actually be private.
					No trackers, no hidden data collection, no silent compromises. Umbra is designed from the ground up
					with one
					clear principle:
					<q class="block">Your data belongs to you — not to servers, companies, or intermediaries.</q>
					Whether you’re here out of curiosity, concern for privacy, or a desire for stronger security, you’re
					in the
					right place.
				</p>
				<h2>What Makes Umbra Different?</h2>
				<p>
					Umbra is not just encrypted chat. It is a <q>Security Architecture</q>.
				</p>
				<p>
					Here’s what that means for you — in simple terms:
				</p>
				<ul>
					<li>
						<h3><svg class="inline" fill="#e0e0e0">
								<use href="../assets/key.svg" />
							</svg> <b style="color: var(--main-highlight-color-3)">End-to-End Encryption</b> by Design
						</h3>
						Messages are encrypted on your device and can only be decrypted by the intended recipient. Even
						Umbra’s
						servers cannot read your messages.
					</li>

					<li>
						<h3>Client-Side Key Ownership</h3>
						Your private cryptographic keys are generated and stored on your device. They are never stored
						in
						plaintext on any server.
					</li>

					<li>
						<h3>Decentralization-Friendly Architecture</h3>
						Umbra avoids centralized trust wherever possible and supports integrity verification mechanisms
						inspired
						by distributed systems.
					</li>

					<li>
						<h3>Minimal Metadata Exposure</h3>
						Umbra is designed to reduce what can be inferred about who talks to whom, when, and how often.
					</li>

					<li>
						<h3>Open and Auditable</h3>
						Umbra’s core technologies are open to inspection:
						<q class="block">Transparency is a security feature.</q>
					</li>
				</ul>
				<h2>Technologies & Cryptography (Plain Language)</h2>
				<p>
					Umbra uses modern, well-studied cryptographic tools
				</p>
				<p>
					You don’t need to understand these to use Umbra, but if you’re curious:
				</p>

				<ul>
					<li>
						<h3>Secure Key Exchange</h3>
						We use elliptic-curve cryptography (X25519) so a group of users can safely agree on shared
						secrets —
						even over the public internet.
					</li>
					<li>
						<h3><svg class="inline" fill="#e0e0e0">
								<use href="../assets/vault.svg" />
							</svg> Strong Encryption</h3>
						Messages are protected using our <b>powerful unique encryption algorithm</b> called <q>MACE</q>.
						You can
						find more information about it in the <a href="https://github.com/MHSarmadi/MACE"
							target="_blank">MACE
							GitHub repository</a>.
					</li>
					<li>
						<h3>Digital Signatures</h3>
						Every message can be cryptographically signed (Ed25519), allowing clients to verify authenticity
						of
						<b>others</b> and <b>the server</b>.
					</li>
					<li>
						<h3>Key Derivation & Protection</h3>
						Passwords and secrets are hardened using modern memory-hard algorithms (Argon2) to resist
						brute-force
						attacks.
					</li>
					<li>
						<h3>Integrity & Verification Layers</h3>
						Group messaging and message ordering rely on cryptographic validation to prevent silent
						manipulation.
					</li>
				</ul>
				<p>
					All of this happens quietly in the background — your experience remains simple.
				</p>

				<h2>What’s Next?</h2>
				<p>
					Before you can start using Umbra, your client needs to establish a secure session.
					<br />
					This process will:
				</p>
				<ul>
					<li>Generate a unique key pair for your current session</li>
					<li>Establish a secure connection with the Umbra server</li>
					<li>Ensure that future requests are authenticated and protected</li>
				</ul>
				<p>
					No messages are sent and no data is shared until this process completes.
				</p>
				<h2 style="color: var(--main-highlight-color-4)">
					By continuing, you agree to begin Umbra’s secure session initialization process.
				</h2>
				<large-button @click="goto_session()">
					Let’s Go!
				</large-button>
			</div>
			<div class="session-init page-panel animate-in" v-else-if="current_page == 'SessionInit'">
				<h1 style="margin-top: 0;"><svg class="inline" fill="#e0e0e0">
						<use href="../assets/locked.svg" />
					</svg> Establishing Secure Session...</h1>
				<p>
					Umbra is now setting up your secure session. This process may take a moment as we generate
					cryptographic
					keys and prove you are not a robot.
					<br /><br />
					Please wait while we ensure that your communication will be private and secure.
				</p>
				<ul class="steps">
					<li
						:class="`${current_step! > 0 ? 'done' : ((step_failed && current_step == 0) ? 'failed' : (!step_failed ? 'loading' : ''))}${current_step == 0 ? ' active' : ''}`">
						Generating session key pairs...</li>
					<li
						:class="`${current_step! > 1 ? 'done' : ((step_failed && current_step == 1) ? 'failed' : (!step_failed ? 'loading' : ''))}${current_step == 1 ? ' active' : ''}`">
						Executing first handshake with the server...</li>
					<li
						:class="`${current_step! > 2 ? 'done' : ((step_failed && current_step == 2) ? 'failed' : (!step_failed ? 'loading' : ''))}${current_step == 2 ? ' active' : ''}`">
						Cryptographic level of anti-bot assurance...
						<progress-bar v-if="current_step == 2" style="margin-left: 20px;" :percentage="pow_percent"
							size="large" />
					</li>
				</ul>
				<transition name="fade-rise">
					<hr />
				</transition>
				<div class="captcha-transition-host">
					<transition name="captcha-swap" mode="out-in">
						<div v-if="captchaPanelState === 'form'" key="captcha-form" class="captcha-box">
							<p>Meanwhile, please solve the CAPTCHA below to additionally prove you are a human:</p>
							<img :src="captcha_challenge_image" alt="CAPTCHA Challenge" class="captcha-image"
								@contextmenu.prevent="" @drag.prevent="" @dragstart.prevent="" />

							<input-field id="captcha_input" v-model="captcha_input" inputmode="numeric" :maxlength="6"
								v-if="!captcha_verified" style="width: 350px;" label="What's written in the box?"
								:checkoutable="captcha_input.length == 6 && !captcha_error_msg.length && !captcha_loading && !captcha_verified"
								:clearable="!!captcha_error_msg.length && !captcha_loading && !captcha_verified"
								:loading="captcha_loading" :disabled="captcha_loading || captcha_verified"
								:readonly="captcha_loading || captcha_verified"
								@checkout="onCaptchaCheckout(captcha_input)" @enter="onCaptchaCheckout(captcha_input)"
								:error-text="captcha_error_msg.length ? captcha_error_msg : undefined" />
							<transition name="fade-rise">
								<p v-if="captcha_verified" class="captcha-success">{{ captcha_success_msg }}</p>
							</transition>
						</div>
						<p v-else-if="captchaPanelState === 'verified'" key="captcha-verified"
							class="captcha-success compact">
							CAPTCHA Verified.
						</p>
					</transition>
				</div>

				<div v-if="canProceedToNext" class="next-action">
					<h3 class="next-hint">All anti-bot checks are complete. You can continue now.</h3>
					<large-button @click="goto_next_page()">Next</large-button>
				</div>
				<p v-if="step_failed" class="failure-message">
					{{ failure_message || 'Session initialization failed. Please refresh the page and try again.' }}
				</p>
			</div>
			<div class="session-ready page-panel animate-in" v-else-if="current_page == 'SessionReady'">
				<h1 style="margin-top: 0;">
					<svg class="inline" fill="#e0e0e0">
						<use href="../assets/check2.svg" />
					</svg>
					Session Ready
				</h1>

				<p>
					Your secure session has been established successfully. You are ready for the next part of the flow.
				</p>
				<p>
					The initialization phase is complete. Temporary session keys have been generated,
					and this device is now operating within a protected communication channel. All further interactions
					with Umbra will take place inside this secured context.
				</p>
				<p>
					No account is attached yet. The session is active, but not authenticated.
				</p>
				<h2>Next Step</h2>
				<p>You may now:</p>
				<ul>
					<li>Login to an existing Umbra account</li>
					<li>Sign Up to create a new account</li>
				</ul>
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
				<hr />
				<h3>You are now ready to continue.</h3>
				<large-button style="margin-right: 15px">Login</large-button>
				<large-button>Signup</large-button>
			</div>
		</div>
	</transition>
</template>

<style scoped lang="less">
@import url(../style.less);

.page-panel {
	position: relative;
	overflow: hidden;
}

.hello-umbra {
	background:
		radial-gradient(120% 90% at 100% -10%, #26a69a66 0%, transparent 62%),
		radial-gradient(100% 100% at -10% 105%, #007fff4d 0%, transparent 72%),
		linear-gradient(165deg, #151515 0%, #0d0d0d 100%);
	border-radius: var(--border-radius-lg);
	box-shadow: 0 20px 44px #0000006e;
	@w1: calc(100vw - 60px);
	@w2: max(40vw, 640px);
	width: min(@w1, @w2);
	@h1: calc(100dvh - 60px);
	@h2: max(40dvh, 800px);
	height: min(@h1, @h2);
	padding: 20px;
	.scroll-container();
	background-size: 150% 150%, 160% 160%, 140% 140%;
	animation: panel-gradient-drift 14s ease-in-out infinite alternate;

	h1 {
		border-bottom: 0;
		padding-bottom: 0;
	}

	li {
		text-align: justify;
	}
}

.session-init {
	background:
		radial-gradient(120% 90% at 100% -10%, #26a69a66 0%, transparent 62%),
		radial-gradient(100% 100% at -10% 105%, #007fff4d 0%, transparent 72%),
		linear-gradient(165deg, #151515 0%, #0d0d0d 100%);
	@w1: calc(100vw - 60px);
	@w2: max(40vw, 720px);
	width: min(@w1, @w2);
	@h1: calc(100dvh - 60px);
	@h2: max(40dvh, 560px);
	height: min(@h1, @h2);
	padding: 20px;
	border-radius: var(--border-radius-lg);
	overflow: hidden;
	box-shadow: 0 14px 40px #00000057;
	background-size: 140% 140%, 130% 130%;
	animation: panel-gradient-drift 16s ease-in-out infinite alternate-reverse;
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
}

.captcha-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 10px;
}

.captcha-transition-host {
	min-height: 42px;
}

.captcha-image {
	width: 350px;
	margin-top: 10px;
	border-radius: var(--border-radius-md);
	pointer-events: none;
	user-select: none;
}

.captcha-success {
	margin: 0;
	color: var(--main-highlight-color-4);
	.bolder();
}

.captcha-success.compact {
	margin-top: 8px;
}

.next-action {
	margin-top: 16px;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	gap: 10px;
}

.next-hint {
	margin: 0;
	color: var(--text-color);
}

.animate-in {
	&>* {
		opacity: 0;
		animation: content-rise 0.5s ease forwards;
	}

	&> :nth-child(1) {
		animation-delay: 0.05s;
	}

	&> :nth-child(2) {
		animation-delay: 0.1s;
	}

	&> :nth-child(3) {
		animation-delay: 0.15s;
	}

	&> :nth-child(4) {
		animation-delay: 0.2s;
	}

	&> :nth-child(5) {
		animation-delay: 0.25s;
	}

	&> :nth-child(6) {
		animation-delay: 0.3s;
	}

	&> :nth-child(7) {
		animation-delay: 0.35s;
	}

	&> :nth-child(8) {
		animation-delay: 0.4s;
	}

	&> :nth-child(9) {
		animation-delay: 0.45s;
	}

	&> :nth-child(10) {
		animation-delay: 0.5s;
	}

	&> :nth-child(11) {
		animation-delay: 0.55s;
	}

	&> :nth-child(12) {
		animation-delay: 0.6s;
	}

	&> :nth-child(13) {
		animation-delay: 0.65s;
	}

	&> :nth-child(14) {
		animation-delay: 0.7s;
	}
}


.steps {
	list-style: none;
	padding-left: 1.5em;

	li {
		position: relative;
		margin-bottom: 0.8em;
		display: flex;
		flex-direction: row;
		flex-wrap: nowrap;
		align-items: center;
		justify-content: flex-start;

		&:not(.active) {
			color: var(--comment-color);
		}

		&.active {
			.bolder();
		}

		&::before {
			content: "";
			position: absolute;
			width: 0.8em;
			height: 0.8em;
			background-size: contain;
			background-repeat: no-repeat;
			background-position: center;
		}

		&.loading::before {
			border: 2px solid #ccc;
			border-top-color: #555;
			border-radius: 50%;
			top: 0.4em;
			left: -1.5em;
			animation: spin 1s linear infinite;
		}

		&.done::before {
			background-image: url("/icons/check2.svg");
			top: 0.1em;
			left: -1em;
			font-size: 1.5em;
		}

		&.failed::before {
			background-image: url("/icons/danger.svg");
			top: 0.1em;
			left: -1em;
			font-size: 1.5em;
		}
	}
}

@keyframes spin {
	from {
		transform: rotate(0deg);
	}

	to {
		transform: rotate(360deg);
	}
}

@keyframes fade-rise-in {
	from {
		opacity: 0;
		transform: translateY(10px);
	}

	to {
		opacity: 1;
		transform: translateY(0);
	}
}

@keyframes fade-rise-out {
	from {
		opacity: 1;
		transform: translateY(0);
	}

	to {
		opacity: 0;
		transform: translateY(-8px);
	}
}

@keyframes fade-rise-lazy-in {
	from {
		opacity: 0;
		transform: translateY(10px);
	}

	30% {
		opacity: 0;
		transform: translateY(10px);
	}

	to {
		opacity: 1;
		transform: translateY(0);
	}
}

@keyframes fade-rise-lazy-out {
	from {
		opacity: 1;
		transform: translateY(0);
	}

	30% {
		opacity: 1;
		transform: translateY(0);
	}

	to {
		opacity: 0;
		transform: translateY(-8px);
	}
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

.page-swap-enter-active,
.page-swap-leave-active {
	transition: opacity 0.32s ease, transform 0.32s ease;
}

.page-swap-enter-from,
.page-swap-leave-to {
	opacity: 0;
	transform: translateY(10px) scale(0.99);
}

.fade-rise-enter-active {
	animation: fade-rise-in 0.28s ease;
}

.fade-rise-leave-active {
	animation: fade-rise-out 0.22s ease;
}

.fade-rise-lazy-enter-active {
	animation: fade-rise-lazy-in calc(0.28s * 2) ease;
}

.fade-rise-lazy-leave-active {
	animation: fade-rise-lazy-out calc(0.28s * 2) ease;
}

.captcha-swap-enter-active,
.captcha-swap-leave-active {
	transition: opacity 0.24s ease, transform 0.24s ease;
}

.captcha-swap-enter-from,
.captcha-swap-leave-to {
	opacity: 0;
	transform: translateY(8px);
}
</style>
