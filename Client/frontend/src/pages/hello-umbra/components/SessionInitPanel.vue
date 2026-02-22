<script setup lang="ts">
import { computed } from 'vue';
import InputField from '../../../components/InputField.vue';
import LargeButton from '../../../components/LargeButton.vue';
import ProgressBar from '../../../components/ProgressBar.vue';

interface Props {
	currentStep: number | null;
	stepFailed: boolean;
	failureMessage: string;
	powPercent: number;
	captchaPanelState: 'form' | 'verified' | 'none';
	captchaChallengeImage: string;
	captchaInput: string;
	captchaErrorMsg: string;
	captchaLoading: boolean;
	captchaVerified: boolean;
	captchaSuccessMsg: string;
	canProceedToNext: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
	(e: 'update:captchaInput', value: string): void;
	(e: 'captchaCheckout'): void;
	(e: 'next'): void;
}>();

const captchaInputModel = computed<string>({
	get: () => props.captchaInput,
	set: (value: string | number | null) => {
		emit('update:captchaInput', String(value ?? ''));
	}
});

function stepClass(step: number): string {
	const done = props.currentStep !== null && props.currentStep > step;
	const failed = props.stepFailed && props.currentStep === step;
	const loading = !props.stepFailed && !done;
	const active = props.currentStep === step;

	return [
		done ? 'done' : '',
		failed ? 'failed' : '',
		loading ? 'loading' : '',
		active ? 'active' : ''
	].filter(Boolean).join(' ');
}
</script>

<template>
	<div class="session-init page-panel animate-in">
		<h1 style="margin-top: 0;">
			<svg class="inline" fill="#e0e0e0">
				<use href="../../../assets/locked.svg" />
			</svg>
			Establishing Secure Session...
		</h1>
		<p>
			Umbra is now setting up your secure session. This process may take a moment as we generate
			cryptographic keys and prove you are not a robot.
			<br /><br />
			Please wait while we ensure that your communication will be private and secure.
		</p>
		<ul class="steps">
			<li :class="stepClass(0)">Generating session key pairs...</li>
			<li :class="stepClass(1)">Executing first handshake with the server...</li>
			<li :class="stepClass(2)">
				Cryptographic level of anti-bot assurance...
				<progress-bar
					v-if="currentStep === 2"
					style="margin-left: 20px;"
					:percentage="powPercent"
					size="large"
				/>
			</li>
		</ul>
		<transition name="fade-rise">
			<hr />
		</transition>
		<div class="captcha-transition-host">
			<transition name="captcha-swap" mode="out-in">
				<div v-if="captchaPanelState === 'form'" key="captcha-form" class="captcha-box">
					<p>Meanwhile, please solve the CAPTCHA below to additionally prove you are a human:</p>
					<img
						:src="captchaChallengeImage"
						alt="CAPTCHA Challenge"
						class="captcha-image"
						@contextmenu.prevent=""
						@drag.prevent=""
						@dragstart.prevent=""
					/>

					<input-field
						id="captcha_input"
						v-model="captchaInputModel"
						inputmode="numeric"
						:maxlength="6"
						v-if="!captchaVerified"
						style="width: 350px;"
						label="What's written in the box?"
						:checkoutable="captchaInput.length === 6 && !captchaErrorMsg.length && !captchaLoading && !captchaVerified"
						:clearable="!!captchaErrorMsg.length && !captchaLoading && !captchaVerified"
						:loading="captchaLoading"
						:disabled="captchaLoading || captchaVerified"
						:readonly="captchaLoading || captchaVerified"
						@checkout="$emit('captchaCheckout')"
						@enter="$emit('captchaCheckout')"
						:error-text="captchaErrorMsg.length ? captchaErrorMsg : undefined"
					/>
					<transition name="fade-rise">
						<p v-if="captchaVerified" class="captcha-success">{{ captchaSuccessMsg }}</p>
					</transition>
				</div>
				<p v-else-if="captchaPanelState === 'verified'" key="captcha-verified" class="captcha-success compact">
					CAPTCHA Verified.
				</p>
			</transition>
		</div>

		<div v-if="canProceedToNext" class="next-action">
			<h3 class="next-hint">All anti-bot checks are complete. You can continue now.</h3>
			<large-button @click="$emit('next')">Next</large-button>
		</div>
		<p v-if="stepFailed" class="failure-message">
			{{ failureMessage || 'Session initialization failed. Please refresh the page and try again.' }}
		</p>
	</div>
</template>

<style scoped lang="less">
@import url(../../../style.less);

.page-panel {
	position: relative;
	overflow: hidden;
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
			background-image: url('/icons/check2.svg');
			top: 0.1em;
			left: -1em;
			font-size: 1.5em;
		}

		&.failed::before {
			background-image: url('/icons/danger.svg');
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

.fade-rise-enter-active {
	animation: fade-rise-in 0.28s ease;
}

.fade-rise-leave-active {
	animation: fade-rise-out 0.22s ease;
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
