<script lang="ts">
	import { onDestroy } from 'svelte';
	import { page } from '$app/stores';

	let errorMessage = '';
	let showError = false;
	let unsubscribe: (() => void) | null = null;

	// Listen untuk SvelteKit error events
	function handleError(event: ErrorEvent) {
		errorMessage = event.message || 'Terjadi kesalahan';
		showError = true;
		setTimeout(() => (showError = false), 5000);
	}

	function handleUnhandledRejection(event: PromiseRejectionEvent) {
		errorMessage = 'Koneksi bermasalah atau server tidak merespons';
		showError = true;
		setTimeout(() => (showError = false), 5000);
	}

	window.addEventListener('error', handleError);
	window.addEventListener('unhandledrejection', handleUnhandledRejection);

	onDestroy(() => {
		window.removeEventListener('error', handleError);
		window.removeEventListener('unhandledrejection', handleUnhandledRejection);
	});

	function dismiss() {
		showError = false;
	}

	function reload() {
		window.location.reload();
	}
</script>

{#if showError}
	<div
		class="fixed inset-x-0 top-0 z-[9999] border-b border-error/40 bg-error/10 px-4 py-2 text-center"
		role="alert"
		aria-live="assertive"
	>
		<div class="mx-auto flex max-w-2xl items-center justify-between gap-3">
			<span class="text-sm text-error">
				<svg class="mr-2 inline h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
				</svg>
				{errorMessage}
			</span>
			<div class="flex gap-2">
				<button class="btn btn-ghost btn-xs" on:click={reload}>Muat ulang</button>
				<button class="btn btn-ghost btn-xs btn-square" on:click={dismiss} aria-label="Tutup error">
					<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" d="M6 6l12 12M18 6L6 18" />
					</svg>
				</button>
			</div>
		</div>
	</div>
{/if}
