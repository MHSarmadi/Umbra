<script setup lang="ts">
import LargeButton from '../../../components/LargeButton.vue';
import Accordion from '../../../components/Accordion.vue';
</script>

<template>
	<div class="session-ready page-panel animate-in">
		<h1 style="margin-top: 0;">
			<svg class="inline" fill="#e0e0e0">
				<use href="../../../assets/check2.svg" />
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
		<accordion :collapsed-height="220">
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
		<h3>You are now ready to continue.</h3>
		<large-button style="margin-right: 15px">Login</large-button>
		<large-button>Signup</large-button>
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
