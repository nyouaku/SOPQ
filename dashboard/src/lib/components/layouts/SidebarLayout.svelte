<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { MENU_ITEMS, menuForApp, activeAppFor } from './menu';
	import BottomBar from './BottomBar.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';

	export let currentPath: string = '';

	$: menuItems = menuForApp(MENU_ITEMS, activeAppFor(currentPath), $authStore.user?.role?.name);

	// Sidebar di mobile: default tersembunyi, dibuka lewat hamburger.
	// Di desktop (lg+): selalu tampil, tidak ada fitur hide.
	let mobileOpen = false;

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

<div class="min-h-screen lg:grid lg:grid-cols-[16rem_1fr] bg-base-100">
	<!-- Overlay (mobile saja) -->
	{#if mobileOpen}
		<div
			class="fixed inset-0 bg-black/50 z-40 lg:hidden"
			aria-hidden="true"
			on:click={() => (mobileOpen = false)}
		></div>
	{/if}

	<!-- Sidebar (desktop: permanen; mobile: drawer) -->
	<aside
		class="w-64 bg-base-200 shadow-xl flex flex-col fixed inset-y-0 left-0 z-50 transition-transform duration-200 lg:static lg:translate-x-0 {!mobileOpen ? '-translate-x-full' : 'translate-x-0'}"
		aria-label="Navigasi utama"
	>
		<!-- Logo -->
		<div class="p-6 border-b border-base-300 flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold text-primary">📦 Stock Opname</h1>
				<p class="text-sm text-base-content/70 mt-1">Warehouse Management</p>
			</div>
			<button
				class="btn btn-ghost btn-sm lg:hidden"
				aria-label="Tutup menu"
				on:click={() => (mobileOpen = false)}
			>
				✕
			</button>
		</div>

		<!-- Navigation -->
		<nav class="flex-1 p-4 space-y-2 overflow-y-auto">
			{#each menuItems as item}
				<a
					href={item.href}
					class="flex items-center gap-3 px-4 py-3 rounded-lg transition-all hover:bg-base-300"
					class:bg-primary={currentPath === item.href}
					class:text-primary-content={currentPath === item.href}
					class:font-semibold={currentPath === item.href}
				>
					<span class="text-xl">{item.icon}</span>
					<span>{item.label}</span>
				</a>
			{/each}
		</nav>

		<!-- User Info -->
		{#if $authStore.user}
			<div class="p-4 border-t border-base-300">
				<div class="flex items-center gap-3 mb-3">
					<div class="avatar placeholder">
						<div class="bg-primary text-primary-content rounded-full w-10">
							<span class="text-lg">{$authStore.user.username?.[0]?.toUpperCase()}</span>
						</div>
					</div>
					<div class="flex-1 min-w-0">
						<div class="font-semibold truncate">
							{$authStore.user.full_name || $authStore.user.username}
						</div>
						<div class="text-xs text-base-content/70 capitalize">
							{$authStore.user.role?.name || $authStore.user.role}
						</div>
					</div>
				</div>
				<button class="btn btn-ghost btn-sm w-full" on:click={handleLogout}>
					<span>🚪</span> Logout
				</button>
			</div>
		{/if}
	</aside>

	<!-- Topbar (mobile): hamburger + judul + notifikasi -->
	<div class="lg:hidden sticky top-0 z-30 flex items-center gap-2 border-b border-base-300 bg-base-100 px-4 py-3">
		<button
			class="btn btn-ghost btn-sm p-1"
			aria-label="Buka menu"
			aria-expanded={mobileOpen}
			on:click={() => (mobileOpen = true)}
		>
			<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
			</svg>
		</button>
		<span class="flex-1 font-bold">📦 Stock Opname</span>
		{#if $authStore.user}
			<NotificationBell />
		{/if}
	</div>

	<!-- Main Content -->
	<main class="flex-1 overflow-y-auto min-h-screen">
		<div>
			<slot />
		</div>
	</main>

	<!-- Bottom Bar di semua ukuran.
	     Diletakkan di luar <main> (bukan di dalam container scroll):
	     Chrome Android tidak me-render fixed descendant container scroll
	     sampai halaman digulir. -->
	<BottomBar {currentPath} />
</div>
