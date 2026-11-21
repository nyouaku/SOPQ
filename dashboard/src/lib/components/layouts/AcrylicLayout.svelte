<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { MENU_ITEMS, menuForApp, activeAppFor } from './menu';
	import { ACRYLIC_ICONS, acrylicIconFor } from './acrylic-icons';

	$: currentPath = $page.url.pathname;
	$: activeApp = activeAppFor(currentPath);
	$: visible = menuForApp(MENU_ITEMS, activeApp, $authStore.user?.role?.name);
	$: activeLabel =
		visible.find((i) => currentPath === i.href || (i.href !== '/' && currentPath.startsWith(i.href)))
			?.label ?? (activeApp === 'inventory' ? 'Inventory' : 'Stock Opname');
	$: user = $authStore.user;
	$: initial = (user?.full_name || user?.username || '?')?.[0]?.toUpperCase() ?? '?';

	let inspectorOpen = true;
	let mobileMenu = false;
	let search = '';
	let now = '';
	$: if (currentPath) mobileMenu = false;

	function searchSubmit(e: Event) {
		e.preventDefault();
		// Halaman Products sudah punya field search; arahkan ke sana.
		// (Layout tidak boleh mengubah logika pencarian bisnis.)
		goto('/products');
	}

	onMount(() => {
		const tick = () => (now = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }));
		tick();
		const t = setInterval(tick, 15000);
		return () => clearInterval(t);
	});

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') mobileMenu = false;
	}
</script>

<svelte:window on:keydown={onKeydown} />

<div class="acr-root">
	<!-- Atmospheric background -->
	<div class="acr-atmosphere" aria-hidden="true">
		<span class="orb orb-a"></span><span class="orb orb-b"></span><span class="orb orb-c"></span>
		<span class="grid-overlay"></span>
	</div>

	<!-- ===== DESKTOP: layered control center ===== -->
	<div class="acr-desktop">
		<!-- floating nav rail -->
		<nav class="acr-rail acrylic" aria-label="Navigasi utama">
			<div class="rail-logo" title="SOPQ">S</div>
			{#each visible.slice(0, 8) as item (item.href)}
				<a
					href={item.href}
					class="rail-btn"
					class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}
					title={item.label}
					aria-label={item.label}
				>
					<svg viewBox="0 0 24 24" class="acr-ico"><path d={acrylicIconFor(item.href)} /></svg>
				</a>
			{/each}
			<div class="rail-spacer"></div>
			<a href="/settings" class="rail-btn" title="Settings" aria-label="Settings">
				<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.settings} /></svg>
			</a>
			<div class="rail-user" title={user?.full_name || user?.username || 'User'}>{initial}</div>
		</nav>

		<!-- command bar -->
		<form class="acr-command acrylic" on:submit={searchSubmit}>
			<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.search} /></svg>
			<input bind:value={search} placeholder="Cari modul, produk, sesi…" aria-label="Pencarian" />
			<span class="cmd-hint">⌘ K</span>
			<a href="/activity" class="cmd-icon" aria-label="Aktivitas & notifikasi" title="Aktivitas & notifikasi">
				<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.bell} /></svg>
			</a>
			<a href="/settings" class="cmd-icon" aria-label="Profil & pengaturan" title="Profil & pengaturan">
				<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.profile} /></svg>
			</a>
		</form>

		<!-- layered workspace -->
		<div class="acr-workspace" class:with-inspector={inspectorOpen}>
			<section class="acr-panel acrylic acr-hero-layer">
				<div class="panel-head">
					<span class="eyebrow">WORKSPACE / {activeApp.toUpperCase()}</span>
					<span class="layer-chip">LAYER 01</span>
				</div>
				<h1 class="acr-title">{activeLabel}</h1>
				<p class="acr-sub">Lapisan kerja utama — permukaan transparan di atas atmosfer sistem.</p>
			</section>

			<section class="acr-panel acrylic acr-main-layer">
				<div class="panel-head">
					<span class="eyebrow">KONTEN</span>
					<button class="ghost-btn" on:click={() => (inspectorOpen = !inspectorOpen)} title="Toggle inspector">
						<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.overview} /></svg>
						Inspector
					</button>
				</div>
				<div class="acr-scroll"><slot /></div>
			</section>

			{#if inspectorOpen}
				<aside class="acr-inspector acrylic" aria-label="Inspector">
					<div class="panel-head">
						<span class="eyebrow">INSPECTOR</span>
						<button class="ghost-btn" aria-label="Tutup inspector" title="Tutup" on:click={() => (inspectorOpen = false)}>
							<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.close} /></svg>
						</button>
					</div>
					<div class="inspector-body">
						<div class="inspector-row">
							<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.profile} /></svg>
							<span>{user?.full_name || user?.username || 'User'}</span>
						</div>
						<div class="inspector-row">
							<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.cloud} /></svg>
							<span>Sinkron aktif</span>
						</div>
						<div class="inspector-row">
							<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.shield} /></svg>
							<span>Akses terverifikasi</span>
						</div>
						<div class="inspector-card">
							<span class="eyebrow">AKTIVITAS</span>
							<p>Data opname & inventory terkonsolidasi dalam satu permukaan.</p>
						</div>
						<slot name="context" />
					</div>
				</aside>
			{/if}
		</div>

		<!-- floating quick-access toolbar -->
		<div class="acr-quickbar acrylic">
			<a href="/scan" class="quick-btn" title="Scan" aria-label="Scan"><svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.search} /></svg></a>
			<a href="/sessions" class="quick-btn" title="Sesi" aria-label="Sesi"><svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.overview} /></svg></a>
			<a href="/reports" class="quick-btn" title="Laporan" aria-label="Laporan"><svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.chart} /></svg></a>
			<span class="quick-time">{now}</span>
		</div>
	</div>

	<!-- ===== MOBILE: floating surfaces ===== -->
	<header class="acr-mobile-top acrylic">
		<button class="ghost-btn" aria-label="Buka menu" title="Menu" on:click={() => (mobileMenu = !mobileMenu)}>
			<svg viewBox="0 0 24 24" class="acr-ico"><path d={mobileMenu ? ACRYLIC_ICONS.close : ACRYLIC_ICONS.menu} /></svg>
		</button>
		<span class="mobile-brand">SOPQ</span>
		<span class="quick-time">{now}</span>
	</header>

	{#if mobileMenu}
		<button class="sheet-backdrop" aria-label="Tutup menu" on:click={() => (mobileMenu = false)}></button>
		<nav class="acr-mobile-menu acrylic" aria-label="Menu" role="dialog" aria-modal="true">
			<div class="sheet-handle" aria-hidden="true"></div>
			<div class="sheet-title">MENU / {activeApp.toUpperCase()}</div>
			{#each visible as item (item.href)}
				<a href={item.href} class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}>
					<svg viewBox="0 0 24 24" class="acr-ico"><path d={acrylicIconFor(item.href)} /></svg>
					{item.label}
				</a>
			{/each}
			<a href="/settings" class="sheet-foot">
				<svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.profile} /></svg>
				Profil & pengaturan
			</a>
		</nav>
	{/if}

	<main class="acr-mobile-main">
		<section class="acr-hero acrylic">
			<span class="eyebrow">LAYER 01</span>
			<h1>{activeLabel}</h1>
			<p>Permukaan kerja utama</p>
			<div class="hero-badges">
				<span class="status-badge">{activeApp.toUpperCase()}</span>
				<span class="status-badge alt">{visible.length} MENU</span>
			</div>
		</section>

		<section class="acr-card acrylic acr-mobile-content">
			<div class="panel-head"><span class="eyebrow">KONTEN</span></div>
			<slot />
		</section>

		<div class="acr-actions" aria-label="Aksi cepat">
			<a href="/scan" class="acrylic"><svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.search} /></svg><span>Scan</span></a>
			<a href="/sessions" class="acrylic"><svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.overview} /></svg><span>Sesi</span></a>
			<a href="/reports" class="acrylic"><svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.chart} /></svg><span>Laporan</span></a>
			<a href="/settings" class="acrylic"><svg viewBox="0 0 24 24" class="acr-ico"><path d={ACRYLIC_ICONS.settings} /></svg><span>Atur</span></a>
		</div>
	</main>

	<nav class="acr-mobile-nav acrylic" aria-label="Navigasi bawah">
		{#each visible.slice(0, 5) as item (item.href)}
			<a href={item.href} class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))} aria-label={item.label} title={item.label}>
				<svg viewBox="0 0 24 24" class="acr-ico"><path d={acrylicIconFor(item.href)} /></svg>
			</a>
		{/each}
	</nav>
</div>

<style>
	.acr-root {
		--a-text: #f7fbff;
		--a-line: rgba(255,255,255,0.28);
		--a-accent: #7dd3fc;
		min-height: 100dvh;
		position: relative;
		overflow-x: hidden;
		background: linear-gradient(135deg, #0b1b33 0%, #122b52 40%, #2b1f5c 100%);
		color: var(--a-text);
		font-family: 'Inter', system-ui, sans-serif;
	}
	.acr-atmosphere { position: fixed; inset: 0; z-index: 0; pointer-events: none; overflow: hidden; }
	.orb { position: absolute; border-radius: 50%; filter: blur(60px); opacity: 0.55; }
	.orb-a { width: 34rem; height: 34rem; left: -8rem; top: -10rem; background: #2563eb; }
	.orb-b { width: 28rem; height: 28rem; right: -6rem; top: 20%; background: #7c3aed; }
	.orb-c { width: 30rem; height: 30rem; left: 30%; bottom: -12rem; background: #06b6d4; }
	.grid-overlay {
		position: absolute; inset: 0; opacity: 0.12;
		background-image: linear-gradient(rgba(255,255,255,.15) 1px, transparent 1px),
			linear-gradient(90deg, rgba(255,255,255,.15) 1px, transparent 1px);
		background-size: 44px 44px;
	}

	/* Acrylic material + fallback */
	.acrylic {
		background: rgba(255,255,255,0.10);
		border: 1px solid var(--a-line);
		box-shadow: 0 18px 40px rgba(3,10,30,0.35), inset 0 1px 0 rgba(255,255,255,0.22);
		backdrop-filter: blur(18px) saturate(150%);
		-webkit-backdrop-filter: blur(18px) saturate(150%);
	}
	@supports not ((backdrop-filter: blur(1px)) or (-webkit-backdrop-filter: blur(1px))) {
		.acrylic { background: rgba(16,32,64,0.88); }
	}
	.acr-ico {
		width: 1.05rem; height: 1.05rem; flex: none;
		fill: none; stroke: currentColor; stroke-width: 1.7;
		stroke-linecap: round; stroke-linejoin: round;
	}
	.eyebrow { font-size: 0.62rem; font-weight: 800; letter-spacing: 0.14em; text-transform: uppercase; color: var(--a-accent); }

	/* Desktop */
	.acr-desktop { position: relative; z-index: 1; min-height: 100dvh; padding: 1rem 1.2rem 5rem; }
	.acr-rail {
		position: fixed; left: 1rem; top: 50%; transform: translateY(-50%); z-index: 40;
		display: flex; flex-direction: column; align-items: center; gap: 0.35rem;
		padding: 0.7rem 0.5rem; border-radius: 1.4rem;
	}
	.rail-logo { width: 2rem; height: 2rem; display: grid; place-items: center; font-weight: 900; color: var(--a-accent); }
	.rail-btn {
		width: 2.5rem; height: 2.5rem; display: grid; place-items: center;
		border-radius: 0.8rem; color: rgba(255,255,255,0.72); text-decoration: none;
	}
	.rail-btn:hover { background: rgba(255,255,255,0.12); color: #fff; }
	.rail-btn.active { background: rgba(125,211,252,0.22); color: #fff; box-shadow: inset 0 0 0 1px rgba(125,211,252,0.5); }
	.rail-spacer { height: 0.5rem; }
	.rail-user {
		width: 2rem; height: 2rem; display: grid; place-items: center;
		border-radius: 9999px; background: rgba(124,58,237,0.55); font-size: 0.75rem; font-weight: 800;
	}

	.acr-command {
		position: relative; z-index: 20;
		display: flex; align-items: center; gap: 0.6rem;
		max-width: 46rem; margin: 0 auto 1.2rem;
		padding: 0.5rem 0.8rem; border-radius: 9999px;
	}
	.acr-command input {
		flex: 1; min-width: 0; border: 0; outline: 0; background: transparent;
		color: #fff; font-size: 0.82rem;
	}
	.acr-command input::placeholder { color: rgba(255,255,255,0.55); }
	.cmd-hint { font-size: 0.65rem; padding: 0.15rem 0.4rem; border: 1px solid var(--a-line); border-radius: 0.4rem; }
	.cmd-icon { display: inline-flex; align-items: center; border: 0; background: transparent; color: rgba(255,255,255,0.8); cursor: pointer; padding: 0.2rem; text-decoration: none; }
	.cmd-icon:hover { color: #fff; }

	.acr-workspace {
		display: grid;
		grid-template-columns: 1fr;
		gap: 1rem;
		max-width: 78rem;
		margin: 0 auto;
		padding-left: 4.5rem;
	}
	.acr-workspace.with-inspector { grid-template-columns: minmax(0, 1fr) 17rem; }
	.acr-panel, .acr-inspector { border-radius: 1.3rem; padding: 1rem 1.1rem; }
	.acr-hero-layer { position: relative; overflow: hidden; min-height: 9rem; }
	.acr-title { font-size: clamp(1.8rem, 3.4vw, 3.1rem); font-weight: 800; letter-spacing: -0.04em; margin: 0.5rem 0 0.3rem; }
	.acr-sub { font-size: 0.8rem; color: rgba(255,255,255,0.7); margin: 0; }
	.layer-chip { font-size: 0.62rem; padding: 0.2rem 0.5rem; border: 1px solid var(--a-line); border-radius: 9999px; color: var(--a-accent); }
	.panel-head { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
	.ghost-btn {
		display: inline-flex; align-items: center; gap: 0.35rem;
		border: 1px solid var(--a-line); border-radius: 0.6rem;
		background: rgba(255,255,255,0.06); color: #fff;
		padding: 0.3rem 0.5rem; font-size: 0.72rem; cursor: pointer;
	}
	.ghost-btn:hover { background: rgba(255,255,255,0.16); }
	.acr-main-layer { min-width: 0; }
	.acr-scroll { margin-top: 0.8rem; max-height: calc(100dvh - 17rem); overflow: auto; padding-right: 0.2rem; }

	.acr-inspector { grid-row: 1 / span 2; }
	.inspector-body { margin-top: 0.8rem; font-size: 0.76rem; }
	.inspector-row { display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem 0; border-bottom: 1px solid rgba(255,255,255,0.12); }
	.inspector-card { margin-top: 0.8rem; padding: 0.7rem; border-radius: 0.8rem; background: rgba(255,255,255,0.08); }
	.inspector-card p { font-size: 0.72rem; color: rgba(255,255,255,0.8); margin: 0.4rem 0 0; }

	.acr-quickbar {
		position: fixed; bottom: 1.2rem; left: 50%; transform: translateX(-50%); z-index: 40;
		display: flex; align-items: center; gap: 0.4rem;
		padding: 0.4rem 0.7rem; border-radius: 9999px;
	}
	.quick-btn {
		width: 2.3rem; height: 2.3rem; display: grid; place-items: center;
		border-radius: 9999px; color: #fff; text-decoration: none;
	}
	.quick-btn:hover { background: rgba(255,255,255,0.14); }
	.quick-time { font-size: 0.7rem; color: rgba(255,255,255,0.7); font-variant-numeric: tabular-nums; }

	/* Mobile */
	.acr-mobile-top, .acr-mobile-main, .acr-mobile-nav, .acr-mobile-menu, .acr-hero, .acr-card, .acr-actions { display: none; }
	@media (max-width: 900px) {
		.acr-desktop { display: none; }
		.acr-mobile-top {
			position: sticky; top: 0.5rem; z-index: 50;
			display: flex; align-items: center; gap: 0.6rem;
			margin: 0.5rem; padding: 0.5rem 0.7rem; border-radius: 1rem;
		}
		.mobile-brand { font-weight: 900; }
		.acr-mobile-top .quick-time { margin-left: auto; }
		.acr-mobile-menu {
			position: fixed; left: 0; right: 0; bottom: 0; z-index: 62;
			display: grid; grid-template-columns: 1fr 1fr; gap: 0.4rem;
			padding: 0.7rem 0.8rem calc(5.2rem + env(safe-area-inset-bottom, 0px));
			border-radius: 1.4rem 1.4rem 0 0;
			max-height: 75dvh; overflow-y: auto;
			animation: sheet-in 200ms ease-out;
		}
		@keyframes sheet-in { from { transform: translateY(28px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
		.sheet-handle { grid-column: 1 / -1; justify-self: center; width: 2.6rem; height: 0.28rem; border-radius: 9999px; background: rgba(255,255,255,0.45); }
		.sheet-title { grid-column: 1 / -1; font-size: 0.62rem; font-weight: 800; letter-spacing: 0.14em; color: var(--a-accent); padding-bottom: 0.3rem; }
		.sheet-foot { grid-column: 1 / -1; border-top: 1px solid var(--a-line); margin-top: 0.3rem; padding-top: 0.6rem !important; }
		.sheet-backdrop { position: fixed; inset: 0; z-index: 61; border: 0; background: rgba(4,10,28,0.55); cursor: default; }
		.acr-mobile-menu a {
			display: flex; align-items: center; gap: 0.4rem;
			padding: 0.5rem; border-radius: 0.6rem;
			color: #fff; text-decoration: none; font-size: 0.74rem;
		}
		.acr-mobile-menu a.active { background: rgba(125,211,252,0.24); }
		.acr-mobile-main { display: block; padding: 0.5rem 0.5rem 6.5rem; }
		.acr-hero {
			display: block; border-radius: 1.4rem; padding: 1.2rem;
			margin-bottom: 1rem;
		}
		.acr-hero h1 { font-size: 2rem; font-weight: 800; letter-spacing: -0.04em; margin: 0.5rem 0 0.2rem; }
		.acr-hero p { font-size: 0.78rem; color: rgba(255,255,255,0.7); margin: 0; }
		.hero-badges { display: flex; gap: 0.4rem; margin-top: 0.8rem; }
		.status-badge { padding: 0.25rem 0.55rem; border-radius: 9999px; background: rgba(125,211,252,0.25); font-size: 0.65rem; font-weight: 700; }
		.status-badge.alt { background: rgba(124,58,237,0.35); }
		.acr-card { display: block; border-radius: 1.2rem; padding: 0.9rem; margin-bottom: 1rem; }
		.acr-mobile-content { min-height: 8rem; }
		.acr-actions { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.5rem; margin-bottom: 1rem; }
		.acr-actions a {
			display: flex; flex-direction: column; align-items: center; gap: 0.25rem;
			padding: 0.6rem 0.2rem; border-radius: 0.9rem;
			color: #fff; text-decoration: none; font-size: 0.62rem;
		}
		.acr-mobile-nav {
			position: fixed; left: 0.5rem; right: 0.5rem; bottom: 0.5rem; z-index: 55;
			display: flex; justify-content: space-around; padding: 0.5rem; border-radius: 1.2rem;
		}
		.acr-mobile-nav a {
			width: 2.6rem; height: 2.3rem; display: grid; place-items: center;
			border-radius: 0.8rem; color: rgba(255,255,255,0.72); text-decoration: none;
		}
		.acr-mobile-nav a.active { background: rgba(125,211,252,0.24); color: #fff; }
	}

	@media (prefers-reduced-motion: reduce) {
		.acr-root * { transition: none !important; animation: none !important; }
	}
</style>
