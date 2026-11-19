<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { onMount } from 'svelte';
	import { MENU_ITEMS, menuForApp, activeAppFor } from './menu';
	import BottomBar from './BottomBar.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';

	export let currentPath: string = '';

	$: menuItems = menuForApp(MENU_ITEMS, activeAppFor(currentPath), $authStore.user?.role?.name);

	let currentTime = '';
	let systemStatus = 'ONLINE';

	// Nav panel di mobile: default tersembunyi, dibuka lewat tombol [MENU].
	let mobileOpen = false;

	onMount(() => {
		const updateTime = () => {
			const now = new Date();
			currentTime = now.toLocaleTimeString('id-ID', {
				hour: '2-digit',
				minute: '2-digit',
				second: '2-digit'
			});
		};
		updateTime();
		const interval = setInterval(updateTime, 1000);
		return () => clearInterval(interval);
	});

	function handleLogout() {
		authStore.logout();
	}

	// Tutup drawer setiap kali berpindah halaman.
	let lastPath = '';
	$: if (currentPath !== lastPath) {
		lastPath = currentPath;
		mobileOpen = false;
	}
</script>

<!-- Command Center Layout -->
<div class="min-h-screen bg-base-100 font-mono lg:grid lg:grid-cols-[16rem_1fr]">
	<!-- Overlay (mobile saja) -->
	{#if mobileOpen}
		<div
			class="fixed inset-0 bg-black/50 z-40 lg:hidden"
			aria-hidden="true"
			on:click={() => (mobileOpen = false)}
		></div>
	{/if}

	<!-- Left Panel - Navigation (desktop: permanen; mobile: drawer) -->
	<aside
		class="col-span-1 bg-base-200 border-r-2 border-base-300 p-4 overflow-y-auto fixed inset-y-0 left-0 z-50 w-72 max-w-[85vw] transition-transform duration-200 lg:static lg:translate-x-0 lg:w-auto {!mobileOpen ? '-translate-x-full' : 'translate-x-0'}"
		aria-label="Navigasi modul"
	>
		<div class="mb-4 pb-4 border-b border-base-300 flex items-center justify-between">
			<div>
				<div class="text-xs text-primary font-bold mb-1">[NAVIGATION]</div>
				<div class="text-xs text-base-content/50">Select Module:</div>
			</div>
			<button
				class="btn btn-ghost btn-sm lg:hidden"
				aria-label="Tutup menu"
				on:click={() => (mobileOpen = false)}
			>
				✕
			</button>
		</div>

		<nav class="space-y-1">
			{#each menuItems as item}
				<a
					href={item.href}
					class="block px-3 py-2 rounded transition-all text-sm hover:bg-base-300"
					class:bg-primary={currentPath === item.href}
					class:text-primary-content={currentPath === item.href}
					class:font-bold={currentPath === item.href}
				>
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2">
							<span class="text-lg">{item.icon}</span>
							<span>{item.label}</span>
						</div>
						<kbd class="kbd kbd-xs hidden lg:inline-flex">{item.shortcut}</kbd>
					</div>
				</a>
			{/each}
		</nav>

		<!-- System Info -->
		<div class="mt-6 pt-4 border-t border-base-300">
			<div class="text-xs text-primary font-bold mb-2">[SYSTEM]</div>
			<div class="space-y-1 text-xs text-base-content/70">
				<div class="flex justify-between">
					<span>API:</span>
					<span class="text-success">●</span>
				</div>
				<div class="flex justify-between">
					<span>DB:</span>
					<span class="text-success">●</span>
				</div>
				<div class="flex justify-between">
					<span>Cache:</span>
					<span class="text-success">●</span>
				</div>
			</div>
		</div>

		<!-- Logout Button -->
		<button class="btn btn-error btn-sm w-full mt-4 font-mono text-xs" on:click={handleLogout}>
			[LOGOUT]
		</button>
	</aside>

	<!-- Right area: status bar + content + status panel -->
	<div class="col-span-1 flex flex-col min-h-screen">
		<!-- Top Status Bar -->
		<div
			class="bg-base-300 border-b-2 border-primary px-4 py-2 flex items-center justify-between text-xs sticky top-0 z-30"
		>
			<div class="flex items-center gap-4">
				<!-- Tombol buka nav panel (mobile) -->
				<button
					class="btn btn-ghost btn-xs px-2 lg:hidden"
					aria-label="Buka menu"
					aria-expanded={mobileOpen}
					on:click={() => (mobileOpen = true)}
				>
					[MENU]
				</button>
				<div class="flex items-center gap-2">
					<div class="w-2 h-2 rounded-full bg-success animate-pulse"></div>
					<span class="text-success font-bold hidden sm:inline">{systemStatus}</span>
				</div>
				<div class="text-base-content/70 hidden sm:block">
					<span class="text-primary">█</span> STOCK OPNAME CONTROL
				</div>
			</div>
			<div class="flex items-center gap-4">
				{#if $authStore.user}
					<div class="text-base-content/70 hidden md:block">
						USER: <span class="text-primary font-bold">{$authStore.user.username?.toUpperCase()}</span>
					</div>
					<div class="text-base-content/70 hidden md:block">
						ROLE: <span class="text-warning font-bold"
							>{$authStore.user.role?.name?.toUpperCase() || 'N/A'}</span
						>
					</div>
				{/if}
				<div class="text-base-content/70">
					TIME: <span class="text-info font-bold">{currentTime}</span>
				</div>
				{#if $authStore.user}
					<NotificationBell />
				{/if}
			</div>
		</div>

		<!-- Content row: main + status panel (stack di mobile, sejajar di xl) -->
		<div class="flex-1 flex flex-col xl:grid xl:grid-cols-[1fr_18rem]">
			<!-- Center Panel - Main Content -->
			<main class="bg-base-100 overflow-y-auto">
				<div class="p-4 md:p-6">
					<!-- Content Header -->
					<div class="border-l-4 border-primary pl-4 mb-6">
						<div class="text-xs text-primary font-bold mb-1">[MAIN TERMINAL]</div>
						<div class="text-sm text-base-content/50">Active Workspace</div>
					</div>

					<!-- Page Content -->
					<slot />
				</div>
			</main>

			<!-- Right Panel - Status Panel (hanya xl ke atas; mobile pakai bottom bar) -->
			<aside class="hidden xl:block bg-base-200 border-l-2 border-base-300 p-4 overflow-y-auto">
				<div class="mb-4 pb-4 border-b border-base-300">
					<div class="text-xs text-primary font-bold mb-1">[STATUS PANEL]</div>
					<div class="text-xs text-base-content/50">Real-time Monitor</div>
				</div>

				<!-- User Card -->
				{#if $authStore.user}
					<div class="card bg-base-300 mb-4">
						<div class="card-body p-3">
							<div class="text-xs text-primary font-bold mb-2">[OPERATOR]</div>
							<div class="flex items-center gap-2 mb-2">
								<div class="avatar placeholder">
									<div class="bg-primary text-primary-content rounded w-8 text-xs">
										<span>{$authStore.user.username?.[0]?.toUpperCase()}</span>
									</div>
								</div>
								<div class="text-xs">
									<div class="font-bold">
										{$authStore.user.full_name || $authStore.user.username}
									</div>
									<div class="text-base-content/50 uppercase">
										{$authStore.user.role?.name || 'N/A'}
									</div>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Quick Stats -->
				<div class="card bg-base-300 mb-4">
					<div class="card-body p-3">
						<div class="text-xs text-primary font-bold mb-2">[QUICK STATS]</div>
						<div class="text-xs text-base-content/50">Belum ada data</div>
					</div>
				</div>

				<!-- Activity Log -->
				<div class="card bg-base-300">
					<div class="card-body p-3">
						<div class="text-xs text-primary font-bold mb-2">[ACTIVITY LOG]</div>
						<div class="space-y-1 text-xs text-base-content/50">
							<div>> System initialized</div>
							<div>> User authenticated</div>
							<div>> Dashboard loaded</div>
							<div class="text-success">> Ready for input</div>
						</div>
					</div>
				</div>

				<!-- Keyboard Shortcuts -->
				<div class="card bg-base-300 mt-4">
					<div class="card-body p-3">
						<div class="text-xs text-primary font-bold mb-2">[SHORTCUTS]</div>
						<div class="space-y-1 text-xs text-base-content/50">
							{#each menuItems.slice(0, 4) as item}
								<div class="flex justify-between">
									<span>{item.label}</span>
									<kbd class="kbd kbd-xs">{item.shortcut}</kbd>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</aside>
		</div>
	</div>

	<!-- Bottom Navigation: semua ukuran -->
	<BottomBar {currentPath} />
</div>

<style>
	/* Terminal-like cursor animation */
	@keyframes blink {
		0%, 50% { opacity: 1; }
		51%, 100% { opacity: 0; }
	}
</style>
