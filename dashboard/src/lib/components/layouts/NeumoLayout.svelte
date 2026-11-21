<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';
	import { MENU_ITEMS, menuForApp, activeAppFor } from './menu';
	import { NEUMO_ICONS, neumoIconFor } from './neumo-icons';

	$: currentPath = $page.url.pathname;
	$: activeApp = activeAppFor(currentPath);
	$: visible = menuForApp(MENU_ITEMS, activeApp, $authStore.user?.role?.name);
	$: activeLabel =
		visible.find((i) => currentPath === i.href || (i.href !== '/' && currentPath.startsWith(i.href)))
			?.label ?? (activeApp === 'inventory' ? 'Inventory' : 'Stock Opname');
	$: user = $authStore.user;
	$: initial = (user?.full_name || user?.username || '?')?.[0]?.toUpperCase() ?? '?';

	let compact = false;
	let mobileMenu = false;
	$: if (currentPath) mobileMenu = false;

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') mobileMenu = false;
	}
</script>

<svelte:window on:keydown={onKeydown} />

<div class="neumo-root" class:neumo--compact={compact}>
	<!-- ===== DESKTOP: centered header ===== -->
	<header class="n-top">
		<div class="n-brand">
			<span class="brand-orb">S</span>
			<div>
				<div class="brand-name">SOPQ</div>
				<div class="brand-sub">SOFT CONTROL PANEL</div>
			</div>
		</div>
		<div class="n-top-center">
			<span class="n-pill">{activeApp === 'inventory' ? 'INVENTORY' : 'STOCK OPNAME'}</span>
			<span class="n-pill outline">{visible.length} MODUL</span>
		</div>
		<div class="n-top-right">
			<button class="n-icon press" aria-label="Notifikasi" title="Notifikasi">
				<svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.bell} /></svg>
			</button>
			<a href="/settings" class="n-icon press" aria-label="Pengaturan" title="Pengaturan">
				<svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.settings} /></svg>
			</a>
			<span class="n-avatar" title={user?.full_name || user?.username || 'User'}>{initial}</span>
		</div>
	</header>

	<!-- Desktop composition -->
	<div class="n-desktop">
		<!-- compact floating dock -->
		<nav class="n-dock soft-raise" aria-label="Navigasi cepat">
			{#each visible.slice(0, 8) as item (item.href)}
				<a
					href={item.href}
					class="dock-btn press"
					class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}
					title={item.label}
					aria-label={item.label}
				>
					<svg viewBox="0 0 24 24" class="n-ico"><path d={neumoIconFor(item.href)} /></svg>
				</a>
			{/each}
		</nav>

		<!-- central workspace -->
		<div class="n-center">
			<section class="soft-raise n-panel n-hero">
				<div class="n-hero-row">
					<div>
						<span class="eyebrow">WORKSPACE / {activeApp.toUpperCase()}</span>
						<h1>{activeLabel}</h1>
						<p>Panel kontrol utama — permukaan lembut menyatu dengan dasar.</p>
					</div>
					<div class="n-hero-stat">
						<svg viewBox="0 0 24 24" class="n-ico"><path d={neumoIconFor(currentPath)} /></svg>
						<span>{visible.length}</span>
						<small>MODUL</small>
					</div>
				</div>
			</section>

			<section class="soft-raise n-panel n-main">
				<div class="n-panel-head">
					<span class="eyebrow">KONTEN UTAMA</span>
					<button class="n-toggle press" class:on={compact} on:click={() => (compact = !compact)} aria-pressed={compact}>
						<span class="n-toggle__knob"></span>
						<span>Kompak</span>
					</button>
				</div>
				<div class="n-inset n-scroll"><slot /></div>
			</section>
		</div>

		<!-- right quick statistics -->
		<aside class="n-right">
			<section class="soft-raise n-panel n-stats">
				<span class="eyebrow">STATISTIK CEPAT</span>
				<div class="stat-row">
					<span class="stat-icon"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.dashboard} /></svg></span>
					<div><strong>{visible.length}</strong><small>Menu tersedia</small></div>
				</div>
				<div class="stat-row">
					<span class="stat-icon accent"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.chart} /></svg></span>
					<div><strong>100%</strong><small>Sistem siap</small></div>
				</div>
				<div class="n-progress"><span style="width: 88%"></span></div>
				<p class="progress-label">Kesehatan sistem <b>88%</b></p>
				<div class="n-stat-actions">
					<a href="/scan" class="n-icon press" title="Scan" aria-label="Scan"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.search} /></svg></a>
					<a href="/sessions" class="n-icon press" title="Sesi" aria-label="Sesi"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.calendar} /></svg></a>
					<a href="/reports" class="n-icon press" title="Laporan" aria-label="Laporan"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.chart} /></svg></a>
				</div>
			</section>
			<section class="soft-raise n-panel n-context">
				<span class="eyebrow">KONTEKS</span>
				<slot name="context">
					<p>{user?.role?.name || 'USER'} · {activeApp.toUpperCase()}</p>
				</slot>
			</section>
		</aside>
	</div>

	<footer class="n-status">
		<span>● SISTEM SIAP</span>
		<span>{activeApp.toUpperCase()} MODE</span>
		<span>{user?.username || 'USER'}</span>
	</footer>

	<!-- ===== MOBILE: compact tactile ===== -->
	<header class="n-mobile-top">
		<span class="brand-orb">S</span>
		<span class="mobile-brand">SOPQ</span>
		<button class="n-icon press" aria-label="Menu" title="Menu" on:click={() => (mobileMenu = !mobileMenu)}>
			<svg viewBox="0 0 24 24" class="n-ico"><path d={mobileMenu ? NEUMO_ICONS.close : NEUMO_ICONS.menu} /></svg>
		</button>
	</header>

	{#if mobileMenu}
		<nav class="n-mobile-menu soft-raise" aria-label="Menu">
			{#each visible as item (item.href)}
				<a href={item.href} class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}>
					<svg viewBox="0 0 24 24" class="n-ico"><path d={neumoIconFor(item.href)} /></svg>
					{item.label}
				</a>
			{/each}
		</nav>
	{/if}

	<main class="n-mobile-main">
		<h1 class="n-mobile-title">{activeLabel}</h1>
		<section class="soft-raise n-mobile-hero">
			<span class="eyebrow">FOKUS UTAMA</span>
			<div class="mobile-hero__num">{visible.length}</div>
			<p>modul di {activeApp.toUpperCase()}</p>
		</section>

		<section class="n-tiles">
			<a href="/scan" class="soft-raise press"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.search} /></svg><span>Scan</span></a>
			<a href="/sessions" class="soft-raise press"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.calendar} /></svg><span>Sesi</span></a>
			<a href="/reports" class="soft-raise press"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.chart} /></svg><span>Laporan</span></a>
			<a href="/settings" class="soft-raise press"><svg viewBox="0 0 24 24" class="n-ico"><path d={NEUMO_ICONS.settings} /></svg><span>Atur</span></a>
		</section>

		<section class="soft-raise n-mobile-panel">
			<div class="n-panel-head"><span class="eyebrow">KONTEN</span></div>
			<div class="n-inset"><slot /></div>
		</section>
	</main>

	<nav class="n-mobile-dock soft-raise" aria-label="Navigasi bawah">
		{#each visible.slice(0, 5) as item (item.href)}
			<a href={item.href} class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))} aria-label={item.label} title={item.label}>
				<svg viewBox="0 0 24 24" class="n-ico"><path d={neumoIconFor(item.href)} /></svg>
			</a>
		{/each}
	</nav>
</div>

<style>
	.neumo-root {
		--n-bg: #e8e9ee;
		--n-ink: #33384a;
		--n-muted: #6c7286;
		--n-accent: #6f7fc1;
		--n-sage: #7fa98b;
		--n-light: #ffffff;
		--n-dark: #c4c7d1;
		min-height: 100dvh;
		background: var(--n-bg);
		color: var(--n-ink);
		font-family: 'Inter', system-ui, sans-serif;
	}
	.soft-raise {
		background: var(--n-bg);
		border: 1px solid rgba(255,255,255,0.6);
		box-shadow: 8px 8px 18px var(--n-dark), -8px -8px 18px var(--n-light);
		border-radius: 1.2rem;
	}
	.n-inset {
		background: var(--n-bg);
		border-radius: 0.9rem;
		box-shadow: inset 5px 5px 11px var(--n-dark), inset -5px -5px 11px var(--n-light);
	}
	.n-ico {
		width: 1.1rem; height: 1.1rem; fill: none; stroke: currentColor;
		stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round;
	}
	.eyebrow { font-size: 0.62rem; font-weight: 800; letter-spacing: 0.12em; text-transform: uppercase; color: var(--n-muted); }
	.press:active { box-shadow: inset 5px 5px 11px var(--n-dark), inset -5px -5px 11px var(--n-light); transform: translateY(1px); }

	/* Top */
	.n-top {
		display: grid; grid-template-columns: 1fr auto 1fr; align-items: center;
		gap: 1rem; padding: 0.8rem 1.2rem;
	}
	.n-brand { display: flex; align-items: center; gap: 0.6rem; }
	.brand-orb {
		width: 2.2rem; height: 2.2rem; border-radius: 9999px;
		display: grid; place-items: center; font-weight: 800; color: var(--n-accent);
		box-shadow: inset 3px 3px 7px var(--n-dark), inset -3px -3px 7px var(--n-light);
	}
	.brand-name { font-weight: 900; letter-spacing: -0.03em; }
	.brand-sub { font-size: 0.6rem; color: var(--n-muted); letter-spacing: 0.1em; }
	.n-top-center { display: flex; gap: 0.5rem; }
	.n-pill { padding: 0.35rem 0.7rem; border-radius: 9999px; font-size: 0.68rem; font-weight: 800; background: var(--n-bg); box-shadow: inset 3px 3px 7px var(--n-dark), inset -3px -3px 7px var(--n-light); }
	.n-pill.outline { color: var(--n-accent); }
	.n-top-right { display: flex; align-items: center; justify-content: flex-end; gap: 0.5rem; }
	.n-icon {
		width: 2.3rem; height: 2.3rem; border-radius: 9999px; border: 0;
		display: grid; place-items: center; background: var(--n-bg); color: var(--n-ink);
		box-shadow: 4px 4px 9px var(--n-dark), -4px -4px 9px var(--n-light);
		cursor: pointer; text-decoration: none;
	}
	.n-icon:hover { color: var(--n-accent); }
	.n-avatar { width: 2.3rem; height: 2.3rem; border-radius: 9999px; display: grid; place-items: center; font-weight: 800; color: #fff; background: var(--n-accent); box-shadow: 3px 3px 8px var(--n-dark); }

	/* Desktop */
	.n-desktop {
		display: grid; grid-template-columns: 4.5rem minmax(0,1fr) 16rem;
		gap: 1.2rem; padding: 0.4rem 1.2rem 3.6rem;
		max-width: 92rem; margin: 0 auto;
	}
	.n-dock {
		display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
		padding: 0.7rem 0.5rem; align-self: start;
	}
	.dock-btn {
		width: 2.8rem; height: 2.8rem; border-radius: 0.9rem;
		display: grid; place-items: center; color: var(--n-muted); text-decoration: none;
	}
	.dock-btn:hover { color: var(--n-ink); }
	.dock-btn.active {
		color: var(--n-accent);
		box-shadow: inset 4px 4px 8px var(--n-dark), inset -4px -4px 8px var(--n-light);
	}
	.n-center { display: flex; flex-direction: column; gap: 1.2rem; min-width: 0; }
	.n-panel { padding: 1.1rem; }
	.n-hero-row { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
	.n-hero h1 { font-size: clamp(1.6rem, 3vw, 2.8rem); font-weight: 800; letter-spacing: -0.04em; margin: 0.4rem 0; }
	.n-hero p { font-size: 0.78rem; color: var(--n-muted); margin: 0; }
	.n-hero-stat {
		width: 5.5rem; height: 5.5rem; border-radius: 1.5rem; flex: none;
		display: flex; flex-direction: column; align-items: center; justify-content: center;
		box-shadow: inset 4px 4px 10px var(--n-dark), inset -4px -4px 10px var(--n-light);
		color: var(--n-accent);
	}
	.n-hero-stat span { font-size: 1.6rem; font-weight: 900; line-height: 1; color: var(--n-ink); }
	.n-hero-stat small { font-size: 0.58rem; letter-spacing: 0.1em; color: var(--n-muted); }
	.n-panel-head { display: flex; align-items: center; justify-content: space-between; gap: 0.6rem; margin-bottom: 0.8rem; }
	.n-toggle {
		display: inline-flex; align-items: center; gap: 0.4rem;
		padding: 0.3rem 0.6rem; border: 0; border-radius: 9999px;
		background: var(--n-bg); color: var(--n-muted); font-size: 0.7rem; font-weight: 700;
		box-shadow: inset 3px 3px 7px var(--n-dark), inset -3px -3px 7px var(--n-light); cursor: pointer;
	}
	.n-toggle__knob { width: 0.7rem; height: 0.7rem; border-radius: 9999px; background: var(--n-muted); transition: transform 180ms ease; }
	.n-toggle.on { color: var(--n-accent); }
	.n-toggle.on .n-toggle__knob { transform: translateX(0.35rem); background: var(--n-accent); }
	.n-scroll { padding: 0.8rem; max-height: calc(100dvh - 20rem); overflow: auto; }
	.neumo--compact .n-hero { display: none; }
	.neumo--compact .n-scroll { max-height: calc(100dvh - 10rem); }

	.n-right { display: flex; flex-direction: column; gap: 1.2rem; }
	.n-stats, .n-context { padding: 1rem; }
	.stat-row { display: flex; align-items: center; gap: 0.7rem; padding: 0.6rem 0; }
	.stat-icon {
		width: 2.2rem; height: 2.2rem; border-radius: 0.8rem; flex: none;
		display: grid; place-items: center; color: var(--n-muted);
		box-shadow: inset 3px 3px 6px var(--n-dark), inset -3px -3px 6px var(--n-light);
	}
	.stat-icon.accent { color: var(--n-sage); }
	.stat-row strong { display: block; font-size: 1.05rem; }
	.stat-row small { font-size: 0.66rem; color: var(--n-muted); }
	.n-progress { height: 0.6rem; border-radius: 9999px; margin-top: 0.6rem; box-shadow: inset 3px 3px 6px var(--n-dark), inset -3px -3px 6px var(--n-light); overflow: hidden; }
	.n-progress span { display: block; height: 100%; border-radius: 9999px; background: linear-gradient(90deg, var(--n-accent), var(--n-sage)); }
	.progress-label { font-size: 0.68rem; color: var(--n-muted); margin: 0.4rem 0 0.8rem; }
	.n-stat-actions { display: flex; gap: 0.5rem; }
	.n-stat-actions .n-icon { width: 2rem; height: 2rem; }
	.n-context p { font-size: 0.75rem; color: var(--n-muted); }
	.n-status {
		position: fixed; left: 0; right: 0; bottom: 0; z-index: 40;
		display: flex; justify-content: space-between; gap: 1rem;
		padding: 0.4rem 1.2rem; font-size: 0.66rem; color: var(--n-muted);
		background: var(--n-bg); border-top: 1px solid rgba(255,255,255,0.6);
		box-shadow: 0 -4px 12px rgba(120,125,145,0.18);
	}

	/* Mobile */
	.n-mobile-top, .n-mobile-main, .n-mobile-dock, .n-mobile-menu, .n-mobile-hero, .n-mobile-panel, .n-tiles { display: none; }
	@media (max-width: 900px) {
		.n-top, .n-desktop, .n-status { display: none; }
		.n-mobile-top {
			position: sticky; top: 0; z-index: 50;
			display: flex; align-items: center; gap: 0.6rem;
			padding: 0.7rem 0.9rem;
		}
		.mobile-brand { font-weight: 900; margin-right: auto; }
		.n-mobile-menu {
			position: sticky; top: 4rem; z-index: 49;
			display: grid; grid-template-columns: 1fr 1fr; gap: 0.4rem;
			margin: 0 0.7rem; padding: 0.7rem;
		}
		.n-mobile-menu a { display: flex; align-items: center; gap: 0.4rem; padding: 0.5rem; border-radius: 0.7rem; color: var(--n-muted); text-decoration: none; font-size: 0.72rem; }
		.n-mobile-menu a.active { color: var(--n-accent); box-shadow: inset 3px 3px 7px var(--n-dark), inset -3px -3px 7px var(--n-light); }
		.n-mobile-main { display: block; padding: 0.7rem 0.7rem 6rem; }
		.n-mobile-title { font-size: 1.8rem; font-weight: 800; letter-spacing: -0.04em; margin: 0.7rem 0; }
		.n-mobile-hero { display: block; padding: 1.1rem; text-align: center; margin-bottom: 1rem; }
		.mobile-hero__num { font-size: 3.4rem; font-weight: 900; line-height: 1; color: var(--n-accent); }
		.n-mobile-hero p { font-size: 0.72rem; color: var(--n-muted); margin: 0.3rem 0 0; }
		.n-tiles { display: grid; grid-template-columns: 1fr 1fr; gap: 0.8rem; margin-bottom: 1rem; }
		.n-tiles a { display: flex; flex-direction: column; align-items: center; gap: 0.4rem; padding: 0.9rem 0.4rem; color: var(--n-ink); text-decoration: none; font-size: 0.74rem; font-weight: 700; }
		.n-tiles a svg { color: var(--n-accent); }
		.n-mobile-panel { display: block; padding: 0.9rem; }
		.n-mobile-panel .n-inset { padding: 0.7rem; }
		.n-mobile-dock {
			position: fixed; left: 0.7rem; right: 0.7rem; bottom: 0.7rem; z-index: 55;
			display: flex; justify-content: space-around; padding: 0.5rem;
		}
		.n-mobile-dock a {
			width: 2.7rem; height: 2.4rem; border-radius: 0.9rem;
			display: grid; place-items: center; color: var(--n-muted); text-decoration: none;
		}
		.n-mobile-dock a.active { color: var(--n-accent); box-shadow: inset 4px 4px 8px var(--n-dark), inset -4px -4px 8px var(--n-light); }
	}

	@media (prefers-reduced-motion: reduce) {
		.n-toggle__knob { transition: none; }
		.press:active { transform: none; }
	}
</style>
