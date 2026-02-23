<script setup lang="ts">
import LargeButton from '../../../components/LargeButton.vue';
import Accordion from '../../../components/Accordion.vue';

defineEmits<{
	(e: 'start'): void;
}>();
</script>

<template>
	<div class="hello-umbra page-panel animate-in">
		<h1 style="margin-top: 0;">
			<svg class="inline" fill="#e0e0e0">
				<use href="../../../assets/locked.svg" />
			</svg>
			Welcome to <i style="color: var(--main-highlight-color-3);">Umbra</i>
		</h1>
		<h2>Welcome — and thank you for being here.</h2>
		<p>
			Umbra is a privacy-first communication platform built for people who believe that private
			conversations should actually be private.
			No trackers, no hidden data collection, no silent compromises. Umbra is designed from the ground up
			with one clear principle:
			<q class="block">Your data belongs to you — not to servers, companies, or intermediaries.</q>
			Whether you’re here out of curiosity, concern for privacy, or a desire for stronger security, you’re
			in the right place.
		</p>
		<h2 style="color: var(--main-highlight-color-4)">
			By continuing, you agree to begin Umbra’s secure session initialization process.
		</h2>
		<large-button @click="$emit('start')">
			Let’s Go!
		</large-button>
		<h2>Technical Notes</h2>
		<accordion :collapsed-height="320">
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
			<h2>What Makes Umbra Different?</h2>
			<p>
				Umbra is not just encrypted chat. It is a <q>Security Architecture</q>.
			</p>
			<p>
				Here’s what that means for you — in simple terms:
			</p>
			<ul>
				<li>
					<h3>
						<svg class="inline" fill="#e0e0e0">
							<use href="../../../assets/key.svg" />
						</svg>
						<b style="color: var(--main-highlight-color-3)">End-to-End Encryption</b> by Design
					</h3>
					Messages are encrypted on your device and can only be decrypted by the intended recipient. Even
					Umbra’s servers cannot read your messages.
				</li>

				<li>
					<h3>Client-Side Key Ownership</h3>
					Your private cryptographic keys are generated and stored on your device. They are never stored in
					plaintext on any server.
				</li>

				<li>
					<h3>Decentralization-Friendly Architecture</h3>
					Umbra avoids centralized trust wherever possible and supports integrity verification mechanisms
					inspired by distributed systems.
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
					<h3>
						<svg class="inline" fill="#e0e0e0">
							<use href="../../../assets/vault.svg" />
						</svg>
						Strong Encryption
					</h3>
					Messages are protected using our <b>powerful unique encryption algorithm</b> called <q>MACE</q>.
					You can find more information about it in the
					<a href="https://github.com/MHSarmadi/MACE" target="_blank">MACE GitHub repository</a>.
				</li>
				<li>
					<h3>Digital Signatures</h3>
					Every message can be cryptographically signed (Ed25519), allowing clients to verify authenticity
					of <b>others</b> and <b>the server</b>.
				</li>
				<li>
					<h3>Key Derivation & Protection</h3>
					Passwords and secrets are hardened using modern memory-hard algorithms (Argon2) to resist
					brute-force attacks.
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
		</accordion>
	</div>
</template>

<style scoped lang="less">
@import url(../../../style.less);

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
