<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { MENU_ITEMS, menuForApp, activeAppFor, activeMenuHref, type AppId } from './menu';
	import { iconFor } from './icons';
	import NotificationBell from '$lib/components/NotificationBell.svelte';
	import { page } from '$app/stores';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	export let contextOpen = true;

	$: currentPath = $page.url.pathname;
	$: activeApp = activeAppFor(currentPath);
	$: visible = menuForApp(MENU_ITEMS, activeApp, $authStore.user?.role?.name);
	$: activeLabel =
		visible.find((i) => i.href === activeMenuHref(currentPath, visible))
			?.label ?? (activeApp === 'inventory' ? 'Inventory Management' : 'Stock Opname');
	$: user = $authStore.user;
	$: initial = (user?.full_name || user?.username || '?')?.[0]?.toUpperCase() ?? '?';

	const MANAGER = ['admin', 'supervisor'];
	$: isManager = MANAGER.includes($authStore.user?.role?.name ?? '');
	$: otherApp = activeApp === 'inventory'
		? { href: '/', label: 'Stock Opname' }
		: { href: '/inventory', label: 'Inventory' };

	function toggleContext() {
		contextOpen = !contextOpen;
	}
</script>

<div class="workspace-root workspace" class:workspace--context={contextOpen}>
	<!-- Rail: ikon saja, tooltip via title -->
	<aside class="rail glass" aria-label="Navigasi cepat">
		<div class="rail__brand" title="SOPQ">S</div>
		{#if isManager}
			<a href={otherApp.href} class="rail__btn press" title="Buka {otherApp.label}">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
					<path d="M4 7h13m0 0l-4-4m4 4l-4 4M20 17H7m0 0l4 4m-4-4l4-4" />
				</svg>
			</a>
		{/if}
		<nav class="rail__nav">
			{#each visible as item (item.href)}
				<a
					href={item.href}
					class="rail__btn press"
					class:active={activeMenuHref(currentPath, visible) === item.href}
					title={item.label}
					aria-label={item.label}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
						<path d={iconFor(item.href)} />
					</svg>
				</a>
			{/each}
		</nav>
		<div class="rail__foot" title={user?.full_name || user?.username || 'User'}>{initial}</div>
	</aside>

	<!-- Topbar tipis -->
	<header class="wtop glass">
		<span class="wtop__title">{activeLabel}</span>
		<span class="wtop__spacer"></span>
		<NotificationBell />
		<button class="rail__btn press" aria-label="Toggle panel konteks" title="Panel konteks" on:click={toggleContext}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round">
				<rect x="3" y="4" width="18" height="16" rx="2" /><path d="M15 4v16" />
			</svg>
		</button>
	</header>

	<main class="wmain">
		<div class="wmain__inner">
			<slot />
		</div>
	</main>

	<!-- Panel konteks: konten dikirim halaman lewat <slot name="context"> -->
	{#if contextOpen}
		<aside class="context glass" in:fly={{ x: 16, duration: 240, easing: cubicOut }}>
			<div class="context__head">
				<span>Konteks</span>
				<button class="press" aria-label="Tutup panel" title="Tutup" on:click={toggleContext}>✕</button>
			</div>
			<div class="context__body">
				<slot name="context">
					<p class="text-xs text-base-content/50">Halaman ini belum menyediakan panel konteks.</p>
				</slot>
			</div>
		</aside>
	{/if}
</div>

<style>
	.workspace-root {
		min-height: 100dvh;
		background: oklch(var(--b2));
	}
	.rail {
		position: fixed;
		top: 0.75rem;
		left: 0.75rem;
		bottom: 0.75rem;
		z-index: 50;
		width: 3.6rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.35rem;
		padding: 0.6rem 0.4rem;
		border-radius: 1.2rem;
	}
	.rail__brand {
		width: 2.2rem;
		height: 2.2rem;
		display: grid;
		place-items: center;
		font-weight: 800;
		color: oklch(var(--p));
	}
	.rail__nav {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}
	.rail__btn {
		width: 2.6rem;
		height: 2.6rem;
		display: grid;
		place-items: center;
		border: 0;
		border-radius: 0.75rem;
		background: transparent;
		color: oklch(var(--bc) / 0.6);
		cursor: pointer;
		text-decoration: none;
	}
	.rail__btn:hover { background: oklch(var(--bc) / 0.08); color: oklch(var(--bc)); }
	.rail__btn.active { background: oklch(var(--p) / 0.14); color: oklch(var(--p)); }
	.rail__btn svg { width: 1.1rem; height: 1.1rem; }
	.rail__foot {
		width: 2rem;
		height: 2rem;
		display: grid;
		place-items: center;
		border-radius: 9999px;
		background: oklch(var(--p) / 0.16);
		color: oklch(var(--p));
		font-size: 0.75rem;
		font-weight: 700;
	}

	.wtop {
		position: fixed;
		top: 0.75rem;
		left: 4.85rem;
		right: 0.75rem;
		z-index: 40;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		min-height: 2.8rem;
		padding: 0.35rem 0.75rem;
		border-radius: 1rem;
	}
	.wtop__title { font-size: 0.85rem; font-weight: 600; }
	.wtop__spacer { flex: 1; }

	.wmain {
		min-height: 100dvh;
		padding: 4.5rem 1rem 2rem 5rem;
	}
	.wmain__inner { max-width: 80rem; margin: 0 auto; }

	.context {
		position: fixed;
		top: 4rem;
		right: 0.75rem;
		bottom: 0.75rem;
		z-index: 45;
		width: 17rem;
		display: flex;
		flex-direction: column;
		border-radius: 1.1rem;
		overflow: hidden;
	}
	.context__head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem;
		font-size: 0.8rem;
		font-weight: 700;
		border-bottom: 1px solid oklch(var(--bc) / 0.1);
	}
	.context__head button {
		border: 0;
		background: transparent;
		color: oklch(var(--bc) / 0.5);
		cursor: pointer;
	}
	.context__body { flex: 1; overflow-y: auto; padding: 0.75rem; }

	/* Panel konteks hanya mengambil ruang di desktop besar */
	@media (max-width: 1279px) {
		.workspace-root.workspace--context .wmain { padding-right: 1rem; }
		.context { display: none; }
	}
	@media (min-width: 1280px) {
		.workspace-root.workspace--context .wmain { padding-right: 18.5rem; }
	}
</style>
