<script lang="ts">
	import { onDestroy } from 'svelte';
	import { fly } from 'svelte/transition';

	// beforeinstallprompt event type belum ada di TS lama
	interface BeforeInstallPromptEvent extends Event {
		prompt(): Promise<void>;
		userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
	}

	let deferredPrompt: BeforeInstallPromptEvent | null = null;
	let showBanner = false;
	let hideTimer: ReturnType<typeof setTimeout> | null = null;

	function scheduleHide() {
		if (hideTimer) clearTimeout(hideTimer);
		hideTimer = setTimeout(() => {
			showBanner = false;
			hideTimer = null;
		}, 5000);
	}

	function onBeforeInstallPrompt(e: Event) {
		e.preventDefault();
		deferredPrompt = e as BeforeInstallPromptEvent;
		const dismissed = localStorage.getItem('pwa-install-dismissed');
		if (!dismissed) {
			showBanner = true;
			scheduleHide();
		}
	}

	function onAppInstalled() {
		deferredPrompt = null;
		if (hideTimer) clearTimeout(hideTimer);
		hideTimer = null;
		showBanner = false;
	}

	window.addEventListener('beforeinstallprompt', onBeforeInstallPrompt);
	window.addEventListener('appinstalled', onAppInstalled);

	async function install() {
		if (!deferredPrompt) return;
		await deferredPrompt.prompt();
		const { outcome } = await deferredPrompt.userChoice;
		if (outcome === 'accepted') {
			showBanner = false;
		}
		deferredPrompt = null;
	}

	function dismiss() {
		if (hideTimer) clearTimeout(hideTimer);
		hideTimer = null;
		showBanner = false;
		localStorage.setItem('pwa-install-dismissed', '1');
	}

	onDestroy(() => {
		if (hideTimer) clearTimeout(hideTimer);
		window.removeEventListener('beforeinstallprompt', onBeforeInstallPrompt);
		window.removeEventListener('appinstalled', onAppInstalled);
	});
</script>

{#if showBanner}
	<div
		class="fixed inset-x-0 bottom-0 z-50 border-t border-base-300 bg-base-100 p-4 shadow-2xl sm:inset-x-auto sm:right-4 sm:bottom-4 sm:rounded-2xl sm:border"
		transition:fly={{ y: 60, duration: 300 }}
	>
		<div class="mx-auto flex max-w-lg items-center gap-3">
			<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-content">
				<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
				</svg>
			</div>
			<div class="min-w-0 flex-1">
				<p class="font-semibold">Pasang Stock Opname</p>
				<p class="text-xs text-base-content/60">Akses cepat dari home screen, bisa offline</p>
			</div>
			<button class="btn btn-primary btn-sm shrink-0" on:click={install}>Pasang</button>
			<button
				class="btn btn-ghost btn-sm btn-square shrink-0"
				aria-label="Tutup"
				on:click={dismiss}
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" d="M6 6l12 12M18 6L6 18" />
				</svg>
			</button>
		</div>
	</div>
{/if}
