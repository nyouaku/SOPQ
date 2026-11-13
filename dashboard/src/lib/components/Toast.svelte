<script lang="ts">
	import { toastStore } from '$lib/stores/toast';
</script>

<!-- Notifikasi pojok kanan atas; auto-dismiss dari store (3 detik). -->
<div class="toasts" aria-live="polite" aria-atomic="false">
	{#each $toastStore as toast (toast.id)}
		<div class="toast toast-item toast-item--{toast.type}" role="status">
			<svg class="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
				{#if toast.type === 'success'}
					<path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round" />
				{:else if toast.type === 'error'}
					<path d="M6 6l12 12M18 6L6 18" stroke-linecap="round" />
				{:else}
					<circle cx="12" cy="12" r="9" />
					<path d="M12 11v5M12 7.5v.5" stroke-linecap="round" />
				{/if}
			</svg>
			<span class="min-w-0 flex-1">{toast.message}</span>
			<button
				class="toast-close"
				aria-label="Tutup notifikasi"
				on:click={() => toastStore.dismiss(toast.id)}
			>
				✕
			</button>
		</div>
	{/each}
</div>

<style>
	.toasts {
		position: fixed;
		top: 0;
		inset-inline-end: 0;
		z-index: 9998;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding: max(0.75rem, env(safe-area-inset-top)) 0.75rem 0 0.75rem;
		width: min(100%, 22rem);
		pointer-events: none;
	}

	.toast-item {
		pointer-events: auto;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.625rem 0.75rem;
		border-radius: 0.85rem;
		border: 1px solid hsl(var(--b3));
		background: color-mix(in srgb, hsl(var(--b2)) 94%, transparent);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		box-shadow: 0 10px 30px hsl(var(--bc) / 0.18);
		font-size: 0.8125rem;
		line-height: 1.35;
		color: hsl(var(--bc));
		animation: toast-in 220ms cubic-bezier(0.32, 0.72, 0, 1);
	}

	.toast-item--success {
		border-color: hsl(var(--su) / 0.5);
		color: hsl(var(--su));
	}
	.toast-item--error {
		border-color: hsl(var(--er) / 0.5);
		color: hsl(var(--er));
	}

	.toast-close {
		border: 0;
		background: transparent;
		color: inherit;
		opacity: 0.55;
		cursor: pointer;
		font-size: 0.75rem;
		padding: 0.15rem;
	}
	.toast-close:hover {
		opacity: 1;
	}

	@keyframes toast-in {
		from {
			opacity: 0;
			transform: translateY(-8px) scale(0.98);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.toast-item {
			animation: none;
		}
	}
</style>
