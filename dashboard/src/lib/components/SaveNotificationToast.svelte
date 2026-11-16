<script lang="ts">
	import { fly } from 'svelte/transition';

	export let message = 'Data tersimpan';
	export let scannedCount = 0;
	export let totalProducts = 0;
</script>

<div
	class="notification"
	role="status"
	aria-live="polite"
	transition:fly={{ y: -40, duration: 300 }}
>
	<div class="content">
		<div class="icon">
			<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
				<path
					fill-rule="evenodd"
					d="M16.704 4.296a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
					clip-rule="evenodd"
				/>
			</svg>
		</div>
		<div class="text">
			<div class="title">{message}</div>
			{#if totalProducts > 0}
				<div class="subtitle">Scan {scannedCount}/{totalProducts}</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.notification {
		position: fixed;
		top: 1.5rem;
		left: 50%;
		background: linear-gradient(135deg, #10b981 0%, #059669 100%);
		border-radius: 0.75rem;
		padding: 0.75rem 1rem;
		box-shadow:
			0 10px 40px rgba(16, 185, 129, 0.3),
			0 2px 8px rgba(0, 0, 0, 0.15);
		z-index: 9999;
		min-width: 280px;
		max-width: 420px;
	}

	.content {
		display: flex;
		gap: 0.75rem;
		align-items: center;
	}

	.icon {
		width: 2rem;
		height: 2rem;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.25);
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		flex-shrink: 0;
	}

	.text {
		flex: 1;
		min-width: 0;
	}

	.title {
		color: white;
		font-weight: 600;
		font-size: 0.875rem;
	}

	.subtitle {
		color: rgba(255, 255, 255, 0.85);
		font-size: 0.75rem;
		margin-top: 0.125rem;
	}
</style>
