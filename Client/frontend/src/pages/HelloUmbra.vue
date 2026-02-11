<script setup lang="ts">
import { inject, ref, type Ref } from 'vue';
import LargeButton from '../components/LargeButton.vue';
// import { useRouter } from 'vue-router';

// const $router = useRouter();

const wasmWorker = inject('wasmWorker') as Worker;
const workerRouter = inject('workerRouter') as Ref<{ [key: string]: (data: any) => void }>;

type Pages = 'HelloUmbra' | 'SessionInit';
const current_page = ref<Pages>('HelloUmbra');

// 0: 'keypair_gen' | 1: 'send_to_server' | 2: 'pow';
const current_step = ref<number|null>(null)
const step_failed = ref<boolean>(false);
const failure_message = ref<string>('');

const pow_percent = ref<number>(0);

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
	}
}

workerRouter.value['SendSessionKeypair'] = (event: MessageEvent) => {
	if (current_step.value !== 1 && event.data.success) {
		console.warn("Not in the right step for SendSessionKeypair response. Ignoring.")
		return; // Ignore if not in right step
	}
	if (event.data.success) {
		current_step.value = 2;

		// TODO: Proceed to next step, e.g., proof of work or human verification
		setTimeout(() => {
			// Simulate proof of work progress
			const powInterval = setInterval(() => {
				if (pow_percent.value >= 100) {
					clearInterval(powInterval);
					// Session initialization complete, proceed to next page or functionality
					console.log("Session initialization complete!");
					current_step.value = 3;
				} else if (pow_percent.value < 180) {
					pow_percent.value += Math.floor(Math.random() * 10) + 5; // Increment by random value for demo
				} else {
					// clearInterval(powInterval);
					// step_failed.value = true;
					// failure_message.value = 'Cryptographic proof of you not being a robot failed. Please refresh the page and try again. If the problem persists, please contact support.';
					// console.error('Proof of work failed during session initialization.');
				}
			}, 500);
		}, 1000);
	} else {
		console.error('Failed to send session key pair:', event.data.error);
		step_failed.value = true;
		failure_message.value = 'Failed to send session key pair to the server. Please check your internet connection and try again. If the problem persists, please contact support.';
	}
}


function goto_session() {
	current_page.value = 'SessionInit';
	current_step.value = 0;

	// STEP 1: Generate Keypair
	wasmWorker.postMessage({
		type: 'SessionKeypair'
	});
}
</script>

<template>
	<div class="hello-umbra" v-if="current_page == 'HelloUmbra'">
		<h1 style="margin-top: 0;"><svg class="inline" fill="#e0e0e0"><use href="../assets/locked.svg" /></svg> Welcome to <i style="color: var(--main-highlight-color-3);">Umbra</i></h1>
		<h2>Welcome — and thank you for being here.</h2>
		<p>
			Umbra is a privacy-first communication platform built for people who believe that private conversations should actually be private.
			No trackers, no hidden data collection, no silent compromises. Umbra is designed from the ground up with one clear principle:
			<q class="block">Your data belongs to you — not to servers, companies, or intermediaries.</q>
			Whether you’re here out of curiosity, concern for privacy, or a desire for stronger security, you’re in the right place.
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
				<h3><svg class="inline" fill="#e0e0e0"><use href="../assets/key.svg" /></svg> <b style="color: var(--main-highlight-color-3)">End-to-End Encryption</b> by Design</h3>
				Messages are encrypted on your device and can only be decrypted by the intended recipient. Even Umbra’s servers cannot read your messages.
			</li>

			<li>
				<h3>Client-Side Key Ownership</h3>
				Your private cryptographic keys are generated and stored on your device. They are never stored in plaintext on any server.
			</li>

			<li>
				<h3>Decentralization-Friendly Architecture</h3>
				Umbra avoids centralized trust wherever possible and supports integrity verification mechanisms inspired by distributed systems.
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
				We use elliptic-curve cryptography (X25519) so a group of users can safely agree on shared secrets — even over the public internet.
			</li>
			<li>
				<h3><svg class="inline" fill="#e0e0e0"><use href="../assets/vault.svg" /></svg> Strong Encryption</h3>
				Messages are protected using our <b>powerful unique encryption algorithm</b> called <q>MACE</q>. You can find more information about it in the <a href="https://github.com/MHSarmadi/MACE" target="_blank">MACE GitHub repository</a>.
			</li>
			<li>
				<h3>Digital Signatures</h3>
				Every message can be cryptographically signed (Ed25519), allowing clients to verify authenticity of <b>others</b> and <b>the server</b>.
			</li>
			<li>
				<h3>Key Derivation & Protection</h3>
				Passwords and secrets are hardened using modern memory-hard algorithms (Argon2) to resist brute-force attacks.
			</li>
			<li>
				<h3>Integrity & Verification Layers</h3>
				Group messaging and message ordering rely on cryptographic validation to prevent silent manipulation.
			</li>
		</ul>
		<p>
			All of this happens quietly in the background — your experience remains simple.
		</p>

		<h2>What’s Next?</h2>
		<p>
			Before you can start using Umbra, your client needs to establish a secure session.
			<br/>
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
	<div class="session-init" v-else-if="current_page == 'SessionInit'">
		<h1 style="margin-top: 0;"><svg class="inline" fill="#e0e0e0"><use href="../assets/locked.svg" /></svg> Establishing Secure Session...</h1>
		<p>
			Umbra is now setting up your secure session. This process may take a moment as we generate cryptographic keys and prove you are not a robot.
			<br/><br/>
			Please wait while we ensure that your communication will be private and secure.

		</p>
		<br/>
		<ul class="steps">
			<li :class="`${current_step! > 0 ? 'done' : ((step_failed && current_step == 0) ? 'failed' : (!step_failed ? 'loading' : ''))}${current_step == 0 ? ' active' : ''}`">Generating session key pairs...</li>
			<li :class="`${current_step! > 1 ? 'done' : ((step_failed && current_step == 1) ? 'failed' : (!step_failed ? 'loading' : ''))}${current_step == 1 ? ' active' : ''}`">Sending request to the server...</li>
			<li :class="`${current_step! > 2 ? 'done' : ((step_failed && current_step == 2) ? 'failed' : (!step_failed ? 'loading' : ''))}${current_step == 2 ? ' active' : ''}`">Proving that it's not a robot... <span v-if="current_step == 2">({{ pow_percent }}%)</span></li>
		</ul>
		<p v-if="step_failed" class="failure-message">
			{{ failure_message || 'Session initialization failed. Please refresh the page and try again.' }}
		</p>
	</div>
</template>

<style scoped lang="less">

@import url(../style.less);

.hello-umbra {
	background-color: var(--secondary-bg);
	border-radius: var(--border-radius-lg);
	box-shadow: 0 0 10px var(--shadow-color);
	@w1: calc(100vw - 60px);
	@w2: max(40vw, 650px);
	width: min(@w1, @w2);
	@h1: calc(100dvh - 60px);
	@h2: max(40dvh, 800px);
	height: min(@h1, @h2);
	padding: 20px;
	border-radius: var(--border-radius-md);
	.scroll-container();

	li {
		text-align: justify;
	}
}

.session-init {
	@w1: calc(100vw - 60px);
	@w2: max(40vw, 650px);
	width: min(@w1, @w2);
	@h1: calc(100dvh - 60px);
	@h2: max(40dvh, 800px);
	height: min(@h1, @h2);
	padding: 20px;
	border-radius: var(--border-radius-md);
	.scroll-container();
}

.steps {
	list-style: none;
	padding-left: 1.5em;

	li {
		position: relative;
		margin-bottom: 0.8em;

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

</style>