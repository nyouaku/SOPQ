<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { MENU_ITEMS, menuForApp, activeAppFor, type AppId } from './menu';
	import { iconFor } from './icons';
	import NotificationBell from '$lib/components/NotificationBell.svelte';
	import BottomBar from './BottomBar.svelte';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	export let currentPath: string = '';

	// Aplikasi aktif dari path. Sidebar hanya tampilkan item milik aplikasi aktif.
	$: activeApp: AppId = activeAppFor(currentPath);
	$: visible = menuForApp(MENU_ITEMS, activeApp, $authStore.user?.role?.name);
	$: activeLabel =
		visible.find((i) => currentPath === i.href || (i.href !== '/' && currentPath.startsWith(i.href)))
			?.label ?? (activeApp === 'inventory' ? 'Inventory Management' : 'Stock Opname');

	// Aplikasi lain untuk tombol pindah (hanya manager; staff tak punya akses inventory).
	const MANAGER = ['admin', 'supervisor'];
	$: isManager = MANAGER.includes($authStore.user?.role?.name ?? '');
	$: otherApp = activeApp === 'inventory'
		? { href: '/', label: 'Stock Opname' }
		: { href: '/inventory', label: 'Inventory' };

	let drawerOpen = false;
	let lastPath = '';
	$: if (currentPath !== lastPath) {
		lastPath = currentPath;
		drawerOpen = false;
	}

	$: user = $authStore.user;
	$: initial = (user?.full_name || user?.username || '?')?.[0]?.toUpperCase() ?? '?';
	$: roleName = user?.role?.name ?? '';

	function toggleDrawer() {
		drawerOpen = !drawerOpen;
	}

	async function logout() {
		await authStore.logout();
	}
</script>

<!-- Apple-inspired layout: glass sidebar (desktop) / glass drawer (mobile),
     premium translucent top bar, soft shadows, hairline separators.
     Animasi transform/opacity saja (GPU-friendly); hormati prefers-reduced-motion. -->
<div class="apple-root">
	<!-- Ambient gradient background ( sangat halus, blur lembut). -->
	<div class="ambient" aria-hidden="true">
		<div class="ambient__blob ambient__blob--a"></div>
		<div class="ambient__blob ambient__blob--b"></div>
	</div>

	<!-- Mobile top bar -->
	<header class="topbar glass">
		<button class="iconbtn press" aria-label="Buka menu" on:click={toggleDrawer}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
				<path d="M4 7h16M4 12h16M4 17h16" />
			</svg>
		</button>
		<div class="topbar__title">{activeLabel}</div>
		<div class="topbar__actions">
			<NotificationBell />
			<button class="iconbtn press" aria-label="Keluar" on:click={logout}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
					<path d="M15 12H4m0 0l3-3m-3 3l3 3M10 4h8a2 2 0 012 2v12a2 2 0 01-2 2h-8" />
				</svg>
			</button>
		</div>
	</header>

	<!-- Sidebar desktop + drawer mobile -->
	{#if drawerOpen}
		<div class="drawer-backdrop" transition:fly={{ duration: 200, opacity: 1 }} on:click={toggleDrawer}
			role="button" tabindex="0" aria-label="Tutup menu" on:keydown={(e) => e.key === 'Escape' && toggleDrawer()}></div>
	{/if}

	<aside class="sidebar glass" class:sidebar--open={drawerOpen} aria-label="Navigasi utama">
		<div class="sidebar__brand hairline-b">
			<div class="brandmark">SOPQ</div>
			<div class="brandsub">
				{activeApp === 'inventory' ? 'Inventory Management' : 'Stock Opname'}
			</div>
			{#if isManager}
				<a href={otherApp.href} class="switchapp press" title="Buka {otherApp.label}">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"
						stroke-linecap="round" stroke-linejoin="round">
						<path d="M4 7h13m0 0l-4-4m4 4l-4 4M20 17H7m0 0l4 4m-4-4l4-4" />
					</svg>
					<span>{otherApp.label}</span>
				</a>
			{/if}
		</div>

		<nav class="sidebar__nav">
			{#each visible as item (item.href)}
				<a
					href={item.href}
					class="navitem press"
					class:navitem--active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}
				>
					<svg class="navitem__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor"
						stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
						<path d={iconFor(item.href)} />
					</svg>
					<span class="navitem__label">{item.label}</span>
				</a>
			{/each}
		</nav>

		<div class="sidebar__footer hairline-t">
			<div class="usercard">
				<div class="avatar">{initial}</div>
				<div class="usercard__meta">
					<div class="usercard__name">{user?.full_name || user?.username || 'User'}</div>
					<div class="usercard__role">{roleName}</div>
				</div>
			</div>
			<button class="logout press" on:click={logout}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round">
					<path d="M15 12H4m0 0l3-3m-3 3l3 3M10 4h8a2 2 0 012 2v12a2 2 0 01-2 2h-8" />
				</svg>
				<span>Keluar</span>
			</button>
		</div>
	</aside>

	<!-- Main content -->
	<main class="apple-main">
		<div class="apple-main__inner" in:fly={{ y: 8, duration: 360, easing: cubicOut }}>
			<slot />
		</div>
	</main>

	<!-- Bottom bar (mobile) -->
	<BottomBar {currentPath} />
</div>

<style>
	.apple-root {
		min-height: 100dvh;
		position: relative;
		display: flex;
		flex-direction: column;
		background: hsl(var(--b2));
		color: hsl(var(--bc));
	}

	/* Ambient gradient — dua blob blur besar, bergerak sangat lambat. */
	.ambient {
		position: fixed;
		inset: 0;
		pointer-events: none;
		overflow: hidden;
		z-index: 0;
	}
	.ambient__blob {
		position: absolute;
		border-radius: 50%;
		filter: blur(80px);
		opacity: 0.35;
		will-change: transform;
	}
	.ambient__blob--a {
		width: 42rem;
		height: 42rem;
		top: -12rem;
		right: -10rem;
		background: hsl(var(--p) / 0.25);
		animation: floatA 24s var(--ease-apple) infinite alternate;
	}
	.ambient__blob--b {
		width: 34rem;
		height: 34rem;
		bottom: -12rem;
		left: -10rem;
		background: hsl(var(--s) / 0.18);
		animation: floatB 30s var(--ease-apple) infinite alternate;
	}
	@keyframes floatA {
		from { transform: translate3d(0, 0, 0) scale(1); }
		to { transform: translate3d(-6rem, 4rem, 0) scale(1.1); }
	}
	@keyframes floatB {
		from { transform: translate3d(0, 0, 0) scale(1); }
		to { transform: translate3d(5rem, -4rem, 0) scale(1.08); }
	}

	/* Top bar (mobile) */
	.topbar {
		position: sticky;
		top: 0;
		z-index: 40;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.6rem 0.9rem;
		border-bottom: none;
	}
	.topbar__title {
		flex: 1;
		font-weight: 600;
		font-size: 1rem;
		letter-spacing: -0.01em;
	}
	.topbar__actions {
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}
	.iconbtn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2.25rem;
		height: 2.25rem;
		border-radius: 0.7rem;
		color: hsl(var(--bc));
	}
	.iconbtn svg {
		width: 1.25rem;
		height: 1.25rem;
	}

	/* Sidebar */
	.sidebar {
		position: fixed;
		z-index: 50;
		top: 0;
		left: 0;
		height: 100dvh;
		width: 16rem;
		display: flex;
		flex-direction: column;
		border-radius: 0;
		transform: translateX(-100%);
		transition: transform 320ms var(--ease-apple);
	}
	.sidebar--open {
		transform: translateX(0);
	}
	.drawer-backdrop {
		position: fixed;
		inset: 0;
		z-index: 45;
		background: rgba(0, 0, 0, 0.35);
	}
	.sidebar__brand {
		padding: 1.4rem 1.25rem 1.1rem;
	}
	.brandmark {
		font-size: 1.35rem;
		font-weight: 700;
		letter-spacing: -0.02em;
	}
	.brandsub {
		font-size: 0.72rem;
		color: hsl(var(--bc) / 0.55);
		margin-top: 0.1rem;
	}
	.switchapp {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		margin-top: 0.7rem;
		padding: 0.4rem 0.6rem;
		border-radius: 0.6rem;
		font-size: 0.76rem;
		font-weight: 500;
		background: hsl(var(--p) / 0.1);
		color: hsl(var(--p));
		transition: background var(--dur-hover) var(--ease-apple);
	}
	.switchapp:hover {
		background: hsl(var(--p) / 0.18);
	}
	.switchapp svg {
		width: 0.95rem;
		height: 0.95rem;
		flex-shrink: 0;
	}
	.sidebar__nav {
		flex: 1;
		overflow-y: auto;
		padding: 0.5rem 0.75rem;
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}
	.navitem {
		display: flex;
		align-items: center;
		gap: 0.7rem;
		padding: 0.6rem 0.75rem;
		border-radius: 0.7rem;
		color: hsl(var(--bc) / 0.8);
		transition: background var(--dur-hover) var(--ease-apple), color var(--dur-hover) var(--ease-apple);
	}
	.navitem:hover {
		background: hsl(var(--bc) / 0.06);
		color: hsl(var(--bc));
	}
	.navitem--active {
		background: hsl(var(--p) / 0.14);
		color: hsl(var(--p));
		font-weight: 600;
	}
	.navitem__icon {
		width: 1.2rem;
		height: 1.2rem;
		flex-shrink: 0;
	}
	.navitem__label {
		font-size: 0.9rem;
	}
	.sidebar__footer {
		padding: 0.85rem 0.75rem;
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}
	.hairline-t {
		border-top: 1px solid color-mix(in srgb, hsl(var(--bc)) 10%, transparent);
	}
	.usercard {
		display: flex;
		align-items: center;
		gap: 0.6rem;
	}
	.avatar {
		width: 2.1rem;
		height: 2.1rem;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: hsl(var(--p) / 0.16);
		color: hsl(var(--p));
		font-weight: 600;
		font-size: 0.85rem;
		flex-shrink: 0;
	}
	.usercard__name {
		font-size: 0.85rem;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.usercard__role {
		font-size: 0.72rem;
		color: hsl(var(--bc) / 0.55);
		capitalize: true;
	}
	.logout {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.7rem;
		border-radius: 0.6rem;
		font-size: 0.85rem;
		color: hsl(var(--bc) / 0.7);
	}
	.logout:hover {
		background: hsl(var(--er) / 0.12);
		color: hsl(var(--er));
	}
	.logout svg {
		width: 1.05rem;
		height: 1.05rem;
	}

	/* Main */
	.apple-main {
		flex: 1;
		position: relative;
		z-index: 1;
		padding-top: 3.6rem; /* ruang topbar mobile */
		padding-left: 0;
	}
	.apple-main__inner {
		padding: 1rem 1rem 6rem;
		max-width: 84rem;
		margin: 0 auto;
	}

	/* Desktop */
	@media (min-width: 1024px) {
		.apple-root {
			flex-direction: row;
		}
		.topbar {
			display: none;
		}
		.sidebar {
			transform: translateX(0);
			width: 15rem;
			border-radius: 0 1.25rem 1.25rem 0;
			top: 0.75rem;
			left: 0.75rem;
			height: calc(100dvh - 1.5rem);
		}
		.drawer-backdrop {
			display: none;
		}
		.apple-main {
			padding-top: 0;
			padding-left: 15.75rem;
		}
		.apple-main__inner {
			padding: 1.75rem 2rem 2rem;
		}
	}
</style>
