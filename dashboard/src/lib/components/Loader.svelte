<script lang="ts">
	import { loaderStore } from '$lib/stores/loader';

	/** Ukuran kotak loader (px). */
	export let size = 32;
	/** Override gaya; default mengikuti pengaturan di Settings. */
	export let style: string | null = null;
	export let label = '';

	$: current = style || $loaderStore.loader;
</script>

<!--
  Loader inline. Lima varian, warna mengikuti token DaisyUI
  (currentColor / --p) agar otomatis cocok di semua tema.
-->
<span
	class="loader"
	style="--size:{size}px"
	role="status"
	aria-live="polite"
	aria-label={label || 'Memuat'}
>
	{#if current === 'dots'}
		<span class="loader__dots">
			<i></i><i></i><i></i>
		</span>
	{:else if current === 'bars'}
		<span class="loader__bars">
			<i></i><i></i><i></i><i></i>
		</span>
	{:else if current === 'ring'}
		<span class="loader__ring"></span>
	{:else if current === 'barcode'}
		<!-- Barcode berjalan: sesuai konteks stok/scan -->
		<span class="loader__barcode">
			<i></i><i></i><i></i><i></i><i></i><i></i><i></i>
		</span>
	{:else}
		<span class="loader__spinner"></span>
	{/if}
	{#if label}<span class="sr-only">{label}</span>{/if}
</span>

<style>
	.loader {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: var(--size);
		height: var(--size);
		color: oklch(var(--p));
		flex: none;
	}

	.loader__spinner,
	.loader__ring {
		display: block;
		width: 100%;
		height: 100%;
		border-radius: 9999px;
		border: max(2px, calc(var(--size) / 10)) solid currentColor;
		border-right-color: transparent;
		animation: loader-spin 0.7s linear infinite;
	}

	.loader__ring {
		border-width: calc(var(--size) / 12);
		border-color: color-mix(in srgb, currentColor 22%, transparent);
		border-top-color: currentColor;
		animation-duration: 0.9s;
	}

	.loader__dots,
	.loader__bars,
	.loader__barcode {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: calc(var(--size) / 8);
		width: 100%;
		height: 100%;
	}

	.loader__dots i {
		width: calc(var(--size) / 4);
		height: calc(var(--size) / 4);
		border-radius: 9999px;
		background: currentColor;
		animation: loader-bounce 1s ease-in-out infinite;
	}
	.loader__dots i:nth-child(2) {
		animation-delay: 0.15s;
	}
	.loader__dots i:nth-child(3) {
		animation-delay: 0.3s;
	}

	.loader__bars i {
		width: calc(var(--size) / 7);
		height: 60%;
		border-radius: 2px;
		background: currentColor;
		animation: loader-scale 1s ease-in-out infinite;
	}
	.loader__bars i:nth-child(2) {
		animation-delay: 0.12s;
	}
	.loader__bars i:nth-child(3) {
		animation-delay: 0.24s;
	}
	.loader__bars i:nth-child(4) {
		animation-delay: 0.36s;
	}

	.loader__barcode i {
		width: calc(var(--size) / 12);
		height: 100%;
		background: currentColor;
		animation: loader-sweep 1.1s ease-in-out infinite;
	}
	.loader__barcode i:nth-child(2) {
		width: calc(var(--size) / 18);
		animation-delay: 0.05s;
	}
	.loader__barcode i:nth-child(3) {
		animation-delay: 0.1s;
	}
	.loader__barcode i:nth-child(4) {
		width: calc(var(--size) / 18);
		animation-delay: 0.15s;
	}
	.loader__barcode i:nth-child(5) {
		animation-delay: 0.2s;
	}
	.loader__barcode i:nth-child(6) {
		width: calc(var(--size) / 18);
		animation-delay: 0.25s;
	}
	.loader__barcode i:nth-child(7) {
		animation-delay: 0.3s;
	}

	@keyframes loader-spin {
		to {
			transform: rotate(360deg);
		}
	}
	@keyframes loader-bounce {
		0%,
		100% {
			transform: translateY(0);
			opacity: 0.45;
		}
		50% {
			transform: translateY(-45%);
			opacity: 1;
		}
	}
	@keyframes loader-scale {
		0%,
		100% {
			transform: scaleY(0.4);
			opacity: 0.5;
		}
		50% {
			transform: scaleY(1);
			opacity: 1;
		}
	}
	@keyframes loader-sweep {
		0%,
		100% {
			transform: scaleY(0.55);
			opacity: 0.5;
		}
		50% {
			transform: scaleY(1);
			opacity: 1;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.loader__spinner,
		.loader__ring,
		.loader__dots i,
		.loader__bars i,
		.loader__barcode i {
			animation-duration: 2s;
		}
	}
</style>
