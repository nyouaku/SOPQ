<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { MENU_ITEMS, menuForApp, activeAppFor } from './menu';
	import { RETRO_ICONS, retroIconFor } from './retro-icons';

	$: currentPath = $page.url.pathname;
	$: activeApp = activeAppFor(currentPath);
	$: visible = menuForApp(MENU_ITEMS, activeApp, $authStore.user?.role?.name);
	$: activeLabel =
		visible.find((i) => currentPath === i.href || (i.href !== '/' && currentPath.startsWith(i.href)))
			?.label ?? (activeApp === 'inventory' ? 'Inventory Management' : 'Stock Opname');
	$: user = $authStore.user;

	let now = '';
	let online = true;
	let showStart = false;
	let windowMinimized = false;
	let windowMaximized = false;

	// Drag window (desktop only): offset relatif terhadap viewport.
	let dragOffset = { x: 0, y: 0 };
	let dragging = false;
	let startPointer = { x: 0, y: 0 };

	function onPointerDown(e: PointerEvent) {
		if (windowMaximized || e.pointerType === 'touch') return;
		dragging = true;
		startPointer = { x: e.clientX - dragOffset.x, y: e.clientY - dragOffset.y };
		(e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId);
	}
	function onPointerMove(e: PointerEvent) {
		if (!dragging) return;
		dragOffset = { x: e.clientX - startPointer.x, y: e.clientY - startPointer.y };
	}
	function onPointerUp() {
		dragging = false;
	}
	function resetWindow() {
		dragOffset = { x: 0, y: 0 };
		windowMinimized = false;
		windowMaximized = false;
	}

	onMount(() => {
		const tick = () => {
			now = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
		};
		tick();
		const t = setInterval(tick, 15000);
		const onOnline = () => (online = true);
		const onOffline = () => (online = false);
		window.addEventListener('online', onOnline);
		window.addEventListener('offline', onOffline);
		return () => {
			clearInterval(t);
			window.removeEventListener('online', onOnline);
			window.removeEventListener('offline', onOffline);
		};
	});

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') showStart = false;
	}
</script>

<svelte:window on:keydown={onKeydown} />

<div class="retro-root">
	<div class="retro-noise" aria-hidden="true"></div>

	<!-- ===== TITLE BAR (desktop) ===== -->
	<header class="titlebar">
		<div class="titlebar__left">
			<button class="tb-icon press" aria-label="Menu sistem" title="Menu sistem" on:click={() => (showStart = !showStart)}>
				<svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.gear} /></svg>
			</button>
			<span class="titlebar__title">SOPQ — {activeLabel.toUpperCase()}</span>
		</div>
		<div class="titlebar__right">
			<span class="tb-badge">v1.0</span>
			<span class="tb-clock">{now}</span>
			<button class="tb-icon press" aria-label="Keluar" title="Keluar" on:click={() => authStore.logout()}>
				<svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.power} /></svg>
			</button>
		</div>
	</header>

	{#if showStart}
		<div class="start-menu" role="dialog" aria-label="Menu sistem">
			<div class="start-menu__head">SOPQ SYSTEM MENU</div>
			<button class="start-menu__item" on:click={() => (showStart = false)}>▶ Lanjutkan kerja</button>
			<button class="start-menu__item" on:click={() => { themeStore.setLayout('apple'); showStart = false; }}>▶ Ganti ke layout Apple</button>
			<button class="start-menu__item" on:click={() => authStore.logout()}>▶ Keluar akun</button>
		</div>
	{/if}

	<!-- ===== DESKTOP: nav kiri + workspace + panel kanan ===== -->
	<div class="retro-desktop">
		<nav class="nav-panel" aria-label="Navigasi utama">
			<div class="nav-panel__head">PROGRAM</div>
			{#each visible as item (item.href)}
				<a
					href={item.href}
					class="nav-item press"
					class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}
				>
					<svg viewBox="0 0 16 16" class="px"><path d={retroIconFor(item.href)} /></svg>
					<span>{item.label}</span>
				</a>
			{/each}
		</nav>

		<div class="workspace" class:window-max={windowMaximized}>
			<section
				class="retro-window"
				class:minimized={windowMinimized}
				style="--dx: {dragOffset.x}px; --dy: {dragOffset.y}px"
				aria-label="Workspace utama"
			>
				<header
					class="window-titlebar"
					on:pointerdown={onPointerDown}
					on:pointermove={onPointerMove}
					on:pointerup={onPointerUp}
					on:pointercancel={onPointerUp}
					role="presentation"
				>
					<svg viewBox="0 0 16 16" class="px"><path d={retroIconFor(currentPath)} /></svg>
					<span>SOPQ.EXE — {activeLabel}</span>
					<span class="window-controls">
						<button class="wc press" aria-label="Minimize" title="Minimize" on:pointerdown|stopPropagation on:click={() => (windowMinimized = !windowMinimized)}>
							<svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.minimize} /></svg>
						</button>
						<button class="wc press" aria-label="Maximize" title="Maximize" on:pointerdown|stopPropagation on:click={() => { windowMaximized = !windowMaximized; dragOffset = { x: 0, y: 0 }; }}>
							<svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.maximize} /></svg>
						</button>
						<button class="wc press" aria-label="Reset posisi" title="Reset posisi" on:pointerdown|stopPropagation on:click={resetWindow}>
							<svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.close} /></svg>
						</button>
					</span>
				</header>
				<div class="window-body">
					<div class="window-scroll">
						<slot />
					</div>
				</div>
				<footer class="window-status">
					<span>SIAP</span><span>{activeApp.toUpperCase()} MODE</span><span>{visible.length} MENU</span>
				</footer>
			</section>
		</div>

		<aside class="context-panel" aria-label="Informasi sistem">
			<div class="context-head">SYSTEM INFO</div>
			<div class="context-body">
				<div class="ctx-row"><svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.user} /></svg><span>{user?.full_name || user?.username || 'USER'}</span></div>
				<div class="ctx-row"><svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.folder} /></svg><span>{activeApp.toUpperCase()}</span></div>
				<div class="ctx-row"><svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.network} /></svg><span>{online ? 'ONLINE' : 'OFFLINE'}</span></div>
				<div class="ctx-box">
					<div class="ctx-box__title">AKTIVITAS</div>
					<p>Stock opname & inventory dalam satu sistem.</p>
					<p class="muted">Gunakan menu di samping untuk berpindah modul.</p>
				</div>
				<div class="ctx-box">
					<div class="ctx-box__title">SHORTCUT</div>
					<p>ESC — tutup menu</p>
					<p>ALT+1..9 — (rencana)</p>
				</div>
				<slot name="context">
					<!-- halaman bisa mengirim panel kontekstual sendiri -->
				</slot>
			</div>
		</aside>
	</div>

	<!-- ===== TASKBAR (desktop) ===== -->
	<footer class="taskbar">
		<button class="taskbar__start press" on:click={() => (showStart = !showStart)}>
			<svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.gear} /></svg> START
		</button>
		<div class="taskbar__tasks">
			<button class="task press" class:active={!windowMinimized} on:click={() => (windowMinimized = false)}>
				<svg viewBox="0 0 16 16" class="px"><path d={retroIconFor(currentPath)} /></svg>
				SOPQ.EXE
			</button>
		</div>
		<div class="taskbar__tray">
			<svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.network} /></svg>
			<span>{now}</span>
			<span class="tray-user">{user?.username || 'USER'}</span>
		</div>
	</footer>

	<!-- ===== MOBILE: compact title + single window + dock + status strip ===== -->
	<header class="mobile-titlebar">
		<button class="tb-icon press" aria-label="Menu" title="Menu" on:click={() => (showStart = !showStart)}>
			<svg viewBox="0 0 16 16" class="px"><path d={RETRO_ICONS.gear} /></svg>
		</button>
		<span class="mobile-title">{activeLabel}</span>
		<span class="tb-clock">{now}</span>
	</header>

	<main class="mobile-workspace">
		<div class="mobile-window">
			<div class="mobile-window__head">
				<svg viewBox="0 0 16 16" class="px"><path d={retroIconFor(currentPath)} /></svg>
				<span>{activeLabel}</span>
			</div>
			<div class="mobile-window__body"><slot /></div>
		</div>
	</main>

	<nav class="mobile-dock" aria-label="Navigasi cepat">
		{#each visible.slice(0, 5) as item (item.href)}
			<a href={item.href} class="dock-item press" class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))} aria-label={item.label} title={item.label}>
				<svg viewBox="0 0 16 16" class="px"><path d={retroIconFor(item.href)} /></svg>
			</a>
		{/each}
	</nav>
	<footer class="mobile-status">
		<span>● {online ? 'ONLINE' : 'OFFLINE'}</span>
		<span>{user?.username || 'USER'}</span>
	</footer>
</div>

<style>
	/* ===== Retro theme tokens (scoped layout) ===== */
	.retro-root {
		--r-bg: #d8d0bb;
		--r-panel: #e8e2cf;
		--r-ink: #2b2a22;
		--r-olive: #6b7043;
		--r-orange: #b5651d;
		--r-line: #8b8570;
		--r-shadow: #4a4638;
		min-height: 100dvh;
		background: var(--r-bg);
		color: var(--r-ink);
		font-family: 'Courier New', 'Lucida Console', monospace;
		position: relative;
		overflow-x: hidden;
	}
	.retro-noise {
		position: fixed;
		inset: 0;
		z-index: 90;
		pointer-events: none;
		opacity: 0.16;
		background: repeating-linear-gradient(to bottom, rgba(0,0,0,0.12) 0 1px, transparent 1px 3px);
	}
	.px {
		width: 1rem;
		height: 1rem;
		fill: none;
		stroke: currentColor;
		stroke-width: 1.4;
		stroke-linecap: square;
		stroke-linejoin: miter;
		shape-rendering: crispEdges;
	}
	.press:active { transform: translate(1px, 1px); box-shadow: inset 1px 1px 0 var(--r-shadow); }

	/* Title bar */
	.titlebar {
		position: relative;
		z-index: 60;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		padding: 0.3rem 0.6rem;
		background: linear-gradient(180deg, #f2eddb, #cfc7ad);
		border-bottom: 2px solid var(--r-shadow);
		box-shadow: inset 0 1px 0 #fff, inset 0 -1px 0 var(--r-line);
		font-size: 0.8rem;
		font-weight: bold;
		letter-spacing: 0.04em;
	}
	.titlebar__left, .titlebar__right { display: flex; align-items: center; gap: 0.5rem; }
	.tb-icon {
		display: inline-grid;
		place-items: center;
		width: 1.6rem;
		height: 1.5rem;
		border: 2px solid;
		border-color: #fff var(--r-shadow) var(--r-shadow) #fff;
		background: var(--r-panel);
		cursor: pointer;
	}
	.tb-badge { border: 1px solid var(--r-line); padding: 0 0.3rem; font-size: 0.65rem; }
	.tb-clock { font-variant-numeric: tabular-nums; }

	/* Start menu */
	.start-menu {
		position: fixed;
		left: 0.5rem;
		bottom: 2.4rem;
		z-index: 80;
		width: 13rem;
		background: var(--r-panel);
		border: 2px solid;
		border-color: #fff var(--r-shadow) var(--r-shadow) #fff;
		box-shadow: 4px 4px 0 var(--r-shadow);
	}
	.start-menu__head {
		padding: 0.4rem 0.6rem;
		background: var(--r-olive);
		color: #f4f1e2;
		font-size: 0.72rem;
		letter-spacing: 0.08em;
	}
	.start-menu__item {
		display: block;
		width: 100%;
		text-align: left;
		padding: 0.5rem 0.6rem;
		border: 0;
		background: transparent;
		color: var(--r-ink);
		font-family: inherit;
		font-size: 0.75rem;
		cursor: pointer;
	}
	.start-menu__item:hover { background: var(--r-olive); color: #fff; }

	/* Desktop grid */
	.retro-desktop {
		display: grid;
		grid-template-columns: 9.5rem 1fr 13rem;
		gap: 0.5rem;
		padding: 0.5rem 0.5rem 2.7rem;
		min-height: calc(100dvh - 5rem);
	}
	.nav-panel, .context-panel {
		background: var(--r-panel);
		border: 2px solid;
		border-color: #fff var(--r-shadow) var(--r-shadow) #fff;
		box-shadow: inset 1px 1px 0 #fff, 3px 3px 0 rgba(74,70,56,0.25);
	}
	.nav-panel__head, .context-head {
		padding: 0.35rem 0.5rem;
		background: var(--r-olive);
		color: #f4f1e2;
		font-size: 0.68rem;
		letter-spacing: 0.1em;
	}
	.nav-item {
		display: flex;
		align-items: center;
		gap: 0.45rem;
		padding: 0.4rem 0.5rem;
		margin: 0.25rem;
		border: 1px solid transparent;
		color: var(--r-ink);
		text-decoration: none;
		font-size: 0.72rem;
	}
	.nav-item:hover { border-color: var(--r-line); background: #ded6bd; }
	.nav-item.active { border-color: var(--r-olive); background: var(--r-olive); color: #fff; }
	.nav-item .px { width: 0.9rem; height: 0.9rem; flex: none; }

	.context-body { padding: 0.5rem; font-size: 0.7rem; }
	.ctx-row { display: flex; align-items: center; gap: 0.4rem; padding: 0.25rem 0; border-bottom: 1px dotted var(--r-line); }
	.ctx-box { margin-top: 0.6rem; padding: 0.4rem; border: 2px solid; border-color: var(--r-line) #fff #fff var(--r-line); }
	.ctx-box__title { font-weight: bold; color: var(--r-orange); margin-bottom: 0.25rem; }
	.ctx-box p { margin: 0.15rem 0; }
	.ctx-box .muted { color: #6b6550; }

	/* Window */
	.workspace { min-width: 0; }
	.retro-window {
		transform: translate(var(--dx), var(--dy));
		background: var(--r-panel);
		border: 2px solid;
		border-color: #fff var(--r-shadow) var(--r-shadow) #fff;
		box-shadow: 5px 5px 0 var(--r-shadow);
		display: flex;
		flex-direction: column;
		max-height: calc(100dvh - 7rem);
	}
	.retro-window.minimized .window-body { display: none; }
	.window-max .retro-window { transform: none; height: calc(100dvh - 7rem); }
	.window-titlebar {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.3rem 0.5rem;
		background: var(--r-olive);
		color: #f4f1e2;
		font-size: 0.72rem;
		cursor: move;
		user-select: none;
	}
	.window-controls { margin-left: auto; display: flex; gap: 0.2rem; }
	.wc {
		width: 1.4rem;
		height: 1.3rem;
		display: grid;
		place-items: center;
		border: 1px solid;
		border-color: #fff var(--r-shadow) var(--r-shadow) #fff;
		background: #cfc7ad;
		color: var(--r-ink);
		cursor: pointer;
	}
	.window-body { flex: 1; min-height: 0; overflow: hidden; }
	.window-scroll { height: 100%; overflow: auto; padding: 0.75rem; }
	.window-status {
		display: flex;
		justify-content: space-between;
		gap: 0.5rem;
		padding: 0.25rem 0.5rem;
		background: #cfc7ad;
		border-top: 2px solid;
		border-color: #fff var(--r-shadow);
		font-size: 0.65rem;
	}

	/* Taskbar */
	.taskbar {
		position: fixed;
		left: 0;
		right: 0;
		bottom: 0;
		z-index: 70;
		display: flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.25rem 0.5rem;
		background: linear-gradient(180deg, #e8e2cf, #c9c1a6);
		border-top: 2px solid var(--r-shadow);
		font-size: 0.7rem;
	}
	.taskbar__start, .task {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
		padding: 0.25rem 0.5rem;
		border: 2px solid;
		border-color: #fff var(--r-shadow) var(--r-shadow) #fff;
		background: var(--r-panel);
		font-family: inherit;
		font-size: 0.7rem;
		cursor: pointer;
	}
	.task.active { box-shadow: inset 2px 2px 0 var(--r-shadow); }
	.taskbar__tasks { display: flex; gap: 0.3rem; }
	.taskbar__tray {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: 0.45rem;
		padding: 0.2rem 0.5rem;
		border: 2px solid;
		border-color: var(--r-shadow) #fff #fff var(--r-shadow);
		font-size: 0.68rem;
	}
	.tray-user { color: var(--r-orange); font-weight: bold; }

	/* ===== MOBILE: komposisi beda, bukan sekadar perkecil ===== */
	.mobile-titlebar, .mobile-workspace, .mobile-dock, .mobile-status { display: none; }
	@media (max-width: 767px) {
		.titlebar, .retro-desktop, .taskbar { display: none; }
		.mobile-titlebar {
			position: sticky;
			top: 0;
			z-index: 60;
			display: flex;
			align-items: center;
			gap: 0.5rem;
			padding: 0.3rem 0.5rem;
			background: var(--r-olive);
			color: #f4f1e2;
			border-bottom: 3px solid var(--r-shadow);
			font-size: 0.75rem;
			font-weight: bold;
		}
		.mobile-title { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
		.mobile-workspace { display: block; padding: 0.5rem 0.5rem 5.5rem; }
		.mobile-window {
			background: var(--r-panel);
			border: 2px solid;
			border-color: #fff var(--r-shadow) var(--r-shadow) #fff;
			box-shadow: 4px 4px 0 var(--r-shadow);
		}
		.mobile-window__head {
			display: flex;
			align-items: center;
			gap: 0.4rem;
			padding: 0.35rem 0.5rem;
			background: #d3cbb0;
			border-bottom: 2px solid var(--r-line);
			font-size: 0.72rem;
		}
		.mobile-window__body { padding: 0.6rem; }
		.mobile-dock {
			position: fixed;
			left: 0.4rem;
			right: 0.4rem;
			bottom: 1.7rem;
			z-index: 70;
			display: flex;
			justify-content: space-around;
			padding: 0.35rem;
			background: #cfc7ad;
			border: 2px solid;
			border-color: #fff var(--r-shadow) var(--r-shadow) #fff;
			box-shadow: 3px 3px 0 var(--r-shadow);
		}
		.dock-item {
			display: grid;
			place-items: center;
			width: 2.6rem;
			height: 2.3rem;
			border: 2px solid transparent;
			color: var(--r-ink);
		}
		.dock-item.active { border-color: var(--r-olive); background: var(--r-olive); color: #fff; }
		.mobile-status {
			position: fixed;
			left: 0;
			right: 0;
			bottom: 0;
			z-index: 71;
			display: flex;
			justify-content: space-between;
			padding: 0.2rem 0.5rem;
			background: #2b2a22;
			color: #d8d0bb;
			font-size: 0.62rem;
		}
		/* START menu jadi panel bawah penuh di mobile */
		.start-menu {
			left: 0; right: 0; bottom: 1.7rem; width: auto;
			border-radius: 0.8rem 0.8rem 0 0;
			box-shadow: 0 -4px 0 var(--r-shadow);
		}
		.start-menu__item { padding: 0.75rem 0.6rem; min-height: 44px; }
	}

	@media (prefers-reduced-motion: reduce) {
		.press:active { transform: none; }
	}
</style>
