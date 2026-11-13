<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { MENU_ITEMS, menuForRole } from './menu';
	import BottomBar from './BottomBar.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';

	export let currentPath: string = '';

	$: menuItems = menuForRole(MENU_ITEMS, $authStore.user?.role?.name);

	function handleLogout() {
		authStore.logout();
	}
</script>

<!-- Top Navigation Layout -->
<div class="min-h-screen bg-base-100">
	<!-- Top Navbar -->
	<header class="navbar bg-base-200 shadow-lg sticky top-0 z-30">
		<!-- Logo -->
		<div class="navbar-start">
			<a href="/" class="flex items-center gap-2 px-4">
				<span class="text-2xl">📦</span>
				<div>
					<h1 class="text-xl font-bold text-primary">Stock Opname</h1>
					<p class="text-xs text-base-content/70">WMS</p>
				</div>
			</a>
		</div>

		<!-- Center Navigation (desktop) -->
		<div class="navbar-center hidden lg:flex">
			<ul class="menu menu-horizontal px-1 gap-1">
				{#each menuItems as item}
					<li>
						<a
							href={item.href}
							class="flex items-center gap-2 px-4"
							class:active={currentPath === item.href}
							class:bg-primary={currentPath === item.href}
							class:text-primary-content={currentPath === item.href}
						>
							<span class="text-lg">{item.icon}</span>
							<span class="font-medium">{item.label}</span>
						</a>
					</li>
				{/each}
			</ul>
		</div>

		<!-- Mobile: placeholder agar judul tetap di tengah -->
		<div class="navbar-center lg:hidden">
			<span class="text-sm font-semibold text-base-content/70">WMS</span>
		</div>

		<!-- User Info -->
		<div class="navbar-end gap-2">
			{#if $authStore.user}
				<NotificationBell />
				<div class="hidden md:flex items-center gap-3 mr-2">
					<div class="text-right">
						<div class="text-sm font-semibold">{$authStore.user.full_name || $authStore.user.username}</div>
						<div class="text-xs text-base-content/70 capitalize">{$authStore.user.role?.name || $authStore.user.role}</div>
					</div>
					<div class="avatar placeholder">
						<div class="bg-primary text-primary-content rounded-full w-10">
							<span class="text-lg">{$authStore.user.username?.[0]?.toUpperCase()}</span>
						</div>
					</div>
				</div>
				<button class="btn btn-ghost btn-sm" on:click={handleLogout}>
					<span>🚪</span>
					<span class="hidden md:inline">Logout</span>
				</button>
			{/if}
		</div>
	</header>

	<!-- Main Content -->
	<main class="container mx-auto p-4 md:p-6 pb-24 md:pb-28">
		<slot />
	</main>

	<!-- Bottom Navigation: semua ukuran (desktop ikut bottom bar sesuai permintaan) -->
	<BottomBar {currentPath} />
</div>
