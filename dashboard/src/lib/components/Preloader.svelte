<script lang="ts">
	import { onMount } from 'svelte';
	import { loaderStore } from '$lib/stores/loader';

	// Preloader layar penuh. Hanya sekali per sesi tab (sessionStorage),
	// dan bisa dimatikan di Settings.
	let visible = false;
	let fading = false;

	const SESSION_KEY = 'stock-opname-preloader-shown';

	onMount(() => {
		let alreadyShown = false;
		try {
			alreadyShown = sessionStorage.getItem(SESSION_KEY) === '1';
		} catch {
			// sessionStorage bisa diblokir; anggap belum tampil.
		}

		if (alreadyShown || !$loaderStore.enabled) return;

		visible = true;
		try {
			sessionStorage.setItem(SESSION_KEY, '1');
		} catch {
			// abaikan
		}

		// Minimal 900 ms supaya animasi tidak berkedip, maksimal 2 s.
		const start = Date.now();
		const done = () => {
			const wait = Math.max(0, 900 - (Date.now() - start));
			setTimeout(() => {
				fading = true;
				setTimeout(() => (visible = false), 320);
			}, wait);
		};
		if (document.readyState === 'complete') done();
		else window.addEventListener('load', done, { once: true });
	});
</script>

{#if visible}
	<div
		class="preloader"
		class:preloader--fading={fading}
		role="status"
		aria-live="polite"
		aria-label="Memuat aplikasi"
	>
		{#if $loaderStore.preloader === 'rings'}
			<div class="preloader__rings">
				<span></span><span></span><span></span>
			</div>
		{:else if $loaderStore.preloader === 'boxes'}
			<div class="preloader__boxes">
				<span></span><span></span><span></span><span></span>
			</div>
		{:else if $loaderStore.preloader === 'orbit'}
			<div class="preloader__orbit">
				<span class="preloader__orbit-core"></span>
				<span class="preloader__orbit-dot"></span>
			</div>
		{:else if $loaderStore.preloader === 'skeleton'}
			<div class="preloader__skeleton">
				<i class="preloader__skel-title"></i>
				<i class="preloader__skel-row"></i>
				<i class="preloader__skel-row"></i>
				<i class="preloader__skel-row"></i>
			</div>
		{:else}
			<!-- scan: garis pemindai di atas kotak QR -->
			<div class="preloader__scan">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6">
					<path d="M3 8V5a2 2 0 0 1 2-2h3M16 3h3a2 2 0 0 1 2 2v3M21 16v3a2 2 0 0 1-2 2h-3M8 21H5a2 2 0 0 1-2-2v-3" />
					<path d="M7 12h10" stroke-linecap="round" />
				</svg>
				<span class="preloader__scan-line"></span>
			</div>
		{/if}

		<p class="preloader__title">Stock Opname</p>
		<p class="preloader__sub">Menyiapkan data gudang…</p>
	</div>
{/if}

<style>
	.preloader {
		position: fixed;
		inset: 0;
		z-index: 9999;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1.25rem;
		background: oklch(var(--b1));
		color: oklch(var(--bc));
		transition: opacity 320ms ease;
	}

	.preloader--fading {
		opacity: 0;
		pointer-events: none;
	}

	.preloader__title {
		font-size: 1rem;
		font-weight: 600;
		letter-spacing: -0.01em;
		margin: 0;
	}

	.preloader__sub {
		font-size: 0.8125rem;
		color: oklch(var(--bc) / 0.55);
		margin: -0.75rem 0 0;
	}

	/* 1. scan */
	.preloader__scan {
		position: relative;
		width: 84px;
		height: 84px;
		color: oklch(var(--p));
		overflow: hidden;
	}
	.preloader__scan svg {
		width: 100%;
		height: 100%;
	}
	.preloader__scan-line {
		position: absolute;
		left: 8%;
		right: 8%;
		height: 2px;
		top: 12%;
		background: oklch(var(--p));
		box-shadow: 0 0 12px oklch(var(--p));
		animation: scan-move 1.6s cubic-bezier(0.4, 0, 0.2, 1) infinite;
	}

	/* 2. rings */
	.preloader__rings {
		position: relative;
		width: 84px;
		height: 84px;
	}
	.preloader__rings span {
		position: absolute;
		inset: 0;
		border-radius: 9999px;
		border: 2px solid oklch(var(--p) / 0.55);
		animation: ring-pulse 1.8s ease-out infinite;
	}
	.preloader__rings span:nth-child(2) {
		animation-delay: 0.6s;
	}
	.preloader__rings span:nth-child(3) {
		animation-delay: 1.2s;
	}

	/* 3. boxes */
	.preloader__boxes {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 8px;
		width: 76px;
		height: 76px;
	}
	.preloader__boxes span {
		border-radius: 8px;
		background: oklch(var(--p));
		animation: box-flip 1.4s ease-in-out infinite;
	}
	.preloader__boxes span:nth-child(2) {
		animation-delay: 0.15s;
		opacity: 0.75;
	}
	.preloader__boxes span:nth-child(3) {
		animation-delay: 0.3s;
		opacity: 0.55;
	}
	.preloader__boxes span:nth-child(4) {
		animation-delay: 0.45s;
		opacity: 0.35;
	}

	/* 4. orbit */
	.preloader__orbit {
		position: relative;
		width: 84px;
		height: 84px;
	}
	.preloader__orbit-core {
		position: absolute;
		inset: 32%;
		border-radius: 9999px;
		background: oklch(var(--p) / 0.25);
	}
	.preloader__orbit-dot {
		position: absolute;
		inset: 0;
		border-radius: 9999px;
		border: 1px dashed oklch(var(--p) / 0.4);
		animation: orbit-spin 2.4s linear infinite;
	}
	.preloader__orbit-dot::after {
		content: '';
		position: absolute;
		top: -5px;
		left: calc(50% - 5px);
		width: 10px;
		height: 10px;
		border-radius: 9999px;
		background: oklch(var(--p));
	}

	/* 5. skeleton */
	.preloader__skeleton {
		display: flex;
		flex-direction: column;
		gap: 10px;
		width: min(72vw, 280px);
	}
	.preloader__skeleton i {
		display: block;
		height: 12px;
		border-radius: 6px;
		background: linear-gradient(
			90deg,
			oklch(var(--bc) / 0.06) 0%,
			oklch(var(--bc) / 0.16) 50%,
			oklch(var(--bc) / 0.06) 100%
		);
		background-size: 200% 100%;
		animation: skel 1.3s linear infinite;
	}
	.preloader__skel-title {
		height: 20px !important;
		width: 55%;
	}

	@keyframes scan-move {
		0% {
			top: 12%;
		}
		50% {
			top: 86%;
		}
		100% {
			top: 12%;
		}
	}
	@keyframes ring-pulse {
		0% {
			transform: scale(0.35);
			opacity: 0.9;
		}
		100% {
			transform: scale(1);
			opacity: 0;
		}
	}
	@keyframes box-flip {
		0%,
		100% {
			transform: scale(0.85) rotate(0deg);
		}
		50% {
			transform: scale(1) rotate(90deg);
		}
	}
	@keyframes orbit-spin {
		to {
			transform: rotate(360deg);
		}
	}
	@keyframes skel {
		to {
			background-position: -200% 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.preloader__scan-line,
		.preloader__rings span,
		.preloader__boxes span,
		.preloader__orbit-dot,
		.preloader__skeleton i {
			animation: none;
		}
	}
</style>
