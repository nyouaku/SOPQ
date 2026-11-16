<script lang="ts">
	// Modal konfirmasi berbahaya (hapus, dll.) — mengganti window.confirm() bawaan
	// browser agar konsisten dengan tema dan bisa di-custom per kasus.
	import { createEventDispatcher } from 'svelte';

	export let isOpen = false;
	export let title = 'Konfirmasi';
	export let message = '';
	export let confirmLabel = 'Konfirmasi';
	export let cancelLabel = 'Batal';
	export let danger = false;
	export let isLoading = false;

	const dispatch = createEventDispatcher<{ confirm: void; cancel: void }>();

	function onConfirm() {
		if (isLoading) return;
		dispatch('confirm');
	}
	function onCancel() {
		if (isLoading) return;
		dispatch('cancel');
	}
	function onKeydown(e: KeyboardEvent) {
		if (!isOpen) return;
		if (e.key === 'Escape') onCancel();
	}
</script>

<svelte:window on:keydown={onKeydown} />

{#if isOpen}
	<div class="modal modal-open" role="dialog" aria-modal="true" aria-labelledby="confirm-title">
		<div class="modal-box max-w-sm">
			<h3 id="confirm-title" class="text-lg font-semibold">{title}</h3>
			<p class="mt-3 text-sm whitespace-pre-line">{message}</p>
			<!-- Slot opsional: field input tambahan (mis. textarea alasan penolakan). -->
			<slot />
			<div class="modal-action mt-6">
				<button class="btn btn-ghost btn-sm" on:click={onCancel} disabled={isLoading}>
					{cancelLabel}
				</button>
				<button
					class="btn btn-sm {danger ? 'btn-error' : 'btn-primary'}"
					on:click={onConfirm}
					disabled={isLoading}
				>
					{#if isLoading}
						<span class="loading loading-spinner loading-sm"></span>
					{/if}
					{confirmLabel}
				</button>
			</div>
		</div>
	</div>
	<button class="modal-backdrop" aria-label="Tutup" on:click={onCancel}></button>
{/if}
