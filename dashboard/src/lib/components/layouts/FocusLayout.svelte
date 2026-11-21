<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { MENU_ITEMS, menuForApp, activeAppFor } from './menu';
	import { iconFor } from './icons';
	import NotificationBell from '$lib/components/NotificationBell.svelte';
	import { page } from '$app/stores';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	$: currentPath = $page.url.pathname;
	$: activeApp = activeAppFor(currentPath);
	$: visible = menuForApp(MENU_ITEMS, activeApp, $authStore.user?.role?.name);
	$: activeLabel =
		visible.find((i) => currentPath === i.href || (i.href !== '/' && currentPath.startsWith(i.href)))
			?.label ?? (activeApp === 'inventory' ? 'Inventory Management' : 'Stock Opname');
	$: user = $authStore.user;

	let showMenu = false;
	$: if (currentPath) showMenu = false;

	function exitFocus() {
		themeStore.setLayout('apple');
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && !showMenu) exitFocus();
	}
</script>

<svelte:window on:keydown={onKeydown} />

<div class="focus-root">
	<!-- Floating mini-bar: judul + aksi + keluar. Tidak menghalangi konten. -->
	<div class="focus-minibar glass" in:fly={{ y: -8, duration: 300, easing: cubicOut }}>
		<button class="mini-btn press" aria-label="Buka menu cepat" title="Menu cepat" on:click={() => (showMenu = !showMenu)}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
				<path d="M4 7h16M4 12h16M4 17h16" />
			</svg>
		</button>
		<span class="mini-title">{activeLabel}</span>
		<span class="mini-spacer"></span>
		<NotificationBell />
		<button class="mini-btn press" aria-label="Keluar mode fokus" title="Keluar mode fokus (Esc)" on:click={exitFocus}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
				<path d="M9 4H5a2 2 0 00-2 2v12a2 2 0 002 2h4M15 8l4 4-4 4M20 12H9" />
			</svg>
		</button>
	</div>

	{#if showMenu}
		<div class="focus-menu glass" transition:fly={{ y: -8, duration: 200 }}>
			{#each visible as item (item.href)}
				<a href={item.href} class="focus-menu__item press" class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
						<path d={iconFor(item.href)} />
					</svg>
					<span>{item.label}</span>
				</a>
			{/each}
			<div class="focus-menu__user">
				{user?.full_name || user?.username || 'User'} · {user?.role?.name || ''}
			</div>
		</div>
	{/if}

	<main class="focus-main">
		<div class="focus-inner">
			<slot />
		</div>
	</main>
</div>

<style>
	.focus-root {
		position: relative;
		min-height: 100dvh;
		background: oklch(var(--b2));
	}
	.focus-minibar {
		position: fixed;
		top: 0.75rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 60;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		max-width: calc(100vw - 2rem);
		padding: 0.35rem 0.6rem;
		border-radius: 9999px;
	}
	.mini-title {
		font-size: 0.8rem;
		font-weight: 600;
		white-space: nowrap;
	}
	.mini-spacer { width: 0.25rem; }
	.mini-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.9rem;
		height: 1.9rem;
		border-radius: 9999px;
		border: 0;
		background: oklch(var(--bc) / 0.06);
		color: oklch(var(--bc));
		cursor: pointer;
	}
	.mini-btn:hover { background: oklch(var(--bc) / 0.12); }
	.mini-btn svg { width: 1rem; height: 1rem; }

	.focus-menu {
		position: fixed;
		top: 3.6rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 61;
		display: grid;
		grid-template-columns: repeat(3, minmax(7rem, 1fr));
		gap: 0.25rem;
		max-height: 60vh;
		overflow: auto;
		padding: 0.6rem;
		border-radius: 1rem;
	}
	.focus-menu__item {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.5rem;
		border-radius: 0.6rem;
		font-size: 0.78rem;
		color: oklch(var(--bc) / 0.8);
		text-decoration: none;
	}
	.focus-menu__item:hover { background: oklch(var(--bc) / 0.08); }
	.focus-menu__item.active { background: oklch(var(--p) / 0.14); color: oklch(var(--p)); }
	.focus-menu__item svg { width: 0.95rem; height: 0.95rem; flex: none; }
	.focus-menu__user {
		grid-column: 1 / -1;
		padding: 0.4rem 0.5rem;
		font-size: 0.7rem;
		color: oklch(var(--bc) / 0.5);
		border-top: 1px solid oklch(var(--bc) / 0.1);
	}

	.focus-main { min-height: 100dvh; }
	.focus-inner {
		max-width: 100rem;
		margin: 0 auto;
		padding: 4.5rem 1rem 2.5rem;
	}
	@media (min-width: 768px) {
		.focus-inner { padding: 4.5rem 2rem 2.5rem; }
	}
	@media (prefers-reduced-motion: reduce) {
		.focus-minibar, .focus-menu { transition: none; }
	}
</style>
