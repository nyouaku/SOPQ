<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import BottomBar from './BottomBar.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';

	export let currentPath: string = '';
</script>

<!-- Bottom Bar Layout — mobile-first modern -->
<!--
  h-dvh + overflow-hidden: dokumen tidak ikut menggulir, hanya <main>.
  Penyebab bottom bar "hilang sampai digulir" di Chrome Android adalah
  dokument yang scroll dengan min-h-screen: fixed bottom:0 mengacu ke
  layout viewport yang lebih tinggi dari area tampak (address bar).
-->
<div class="flex h-screen flex-col overflow-hidden bg-base-100 [height:100dvh]">
	<!-- Top Navbar (ringkas) -->
	<header class="navbar bg-base-200 shadow-lg sticky top-0 z-30">
		<div class="flex-1">
			<a href="/" class="flex items-center gap-2 px-4">
				<span class="text-2xl">📦</span>
				<div>
					<h1 class="text-xl font-bold text-primary">Stock Opname</h1>
					<p class="text-xs text-base-content/70 hidden sm:block">WMS</p>
				</div>
			</a>
		</div>
		<div class="navbar-end gap-2">
			{#if $authStore.user}
				<NotificationBell />
				<div class="hidden md:flex items-center gap-3 mr-2">
					<div class="text-right">
						<div class="text-sm font-semibold">
							{$authStore.user.full_name || $authStore.user.username}
						</div>
						<div class="text-xs text-base-content/70 capitalize">
							{$authStore.user.role?.name || $authStore.user.role}
						</div>
					</div>
					<div class="avatar placeholder">
						<div class="bg-primary text-primary-content rounded-full w-10">
							<span class="text-lg"
								>{($authStore.user.full_name || $authStore.user.username)?.[0]}</span
							>
						</div>
					</div>
				</div>
				<button class="btn btn-ghost btn-sm" on:click={() => authStore.logout()}>
					<span>🚪</span>
					<span class="hidden md:inline">Logout</span>
				</button>
			{/if}
		</div>
	</header>

	<!-- Main Content: padding bawah agar tak tertutup bottom bar -->
	<main class="flex-1 overflow-y-auto">
		<slot />
	</main>

	<!-- Bottom Navigation (Apple Dock style) -->
	<BottomBar {currentPath} />
</div>
