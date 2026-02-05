<script setup lang="ts">
import LargeButton from '../components/LargeButton.vue';
import { useRouter } from 'vue-router';

const $router = useRouter();

function goto_session() {
	$router.push({ name: 'SessionInit' });
}
</script>

<template>
	<div class="hello-umbra">
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
		</p>
		<h2>Technologies & Cryptography (Plain Language)</h2>
		<p>
			Umbra uses modern, well-studied cryptographic tools
		</p>
		<p>
			You don’t need to understand these to use Umbra, but if you’re curious:

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
			Let's Go!
		</large-button>
	</div>
</template>

<style scoped lang="less">
@import url(../style.less);
.hello-umbra {
	background-color: var(--secondary-bg);
	border-radius: var(--border-radius-lg);
	// border: 2px solid var(--hover-bg);
	box-shadow: 0 0 10px var(--shadow-color);
	width: min(calc(100vw - 100px), max(40vw, 600px));
	height: min(calc(100vh - 100px), max(40vw, 800px));
	padding: 20px;
	border-radius: var(--border-radius-md);
	.scroll-container();

	li {
		text-align: justify;
	}
}
</style>