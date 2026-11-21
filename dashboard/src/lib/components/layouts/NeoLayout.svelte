<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { MENU_ITEMS, menuForApp, activeAppFor } from './menu';
	import { NEO_ICONS, neoIconFor } from './neo-icons';

	$: currentPath = $page.url.pathname;
	$: activeApp = activeAppFor(currentPath);
	$: visible = menuForApp(MENU_ITEMS, activeApp, $authStore.user?.role?.name);
	$: activeLabel =
		visible.find((i) => currentPath === i.href || (i.href !== '/' && currentPath.startsWith(i.href)))
			?.label ?? (activeApp === 'inventory' ? 'INVENTORY' : 'STOCK OPNAME');
	$: user = $authStore.user;
	$: initial = (user?.full_name || user?.username || '?')?.[0]?.toUpperCase() ?? '?';

	let menuOpen = false;
	let now = '';
	$: if (currentPath) menuOpen = false;

	onMount(() => {
		const tick = () => (now = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }));
		tick();
		const t = setInterval(tick, 15000);
		return () => clearInterval(t);
	});

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') menuOpen = false;
	}
</script>

<svelte:window on:keydown={onKeydown} />

<div class="neo-root">
	<!-- ===== TOP: oversized brand header + horizontal nav ===== -->
	<header class="neo-top">
		<div class="neo-brand">
			<span class="brand-mark">SOPQ<span class="brand-dot">.</span></span>
			<span class="brand-tag">STOCK OPNAME / INVENTORY SYSTEM</span>
		</div>
		<nav class="neo-nav" aria-label="Navigasi utama">
			{#each visible.slice(0, 6) as item (item.href)}
				<a
					href={item.href}
					class="neo-nav__item press"
					class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}
				>
					<svg viewBox="0 0 24 24" class="neo-ico"><path d={neoIconFor(item.href)} /></svg>
					<span>{item.label}</span>
				</a>
			{/each}
		</nav>
		<div class="neo-top__right">
			<span class="neo-clock">{now}</span>
			<a href="/settings" class="neo-ico-btn press" title="Settings" aria-label="Settings">
				<svg viewBox="0 0 24 24" class="neo-ico"><path d={NEO_ICONS.settings} /></svg>
			</a>
			<span class="neo-avatar">{initial}</span>
		</div>
	</header>

	<!-- ===== DESKTOP: asymmetric bento workspace ===== -->
	<div class="bento">
		<!-- upper-left large highlighted KPI card -->
		<section class="bento__hero">
			<span class="kicker">01 / FOKUS UTAMA</span>
			<h1 class="oversized">{activeLabel}</h1>
			<div class="hero-shape" aria-hidden="true"></div>
			<p class="hero-desc">Modul {activeApp === 'inventory' ? 'Inventory Management' : 'Stock Opname'} — data langsung dari sistem.</p>
			<div class="hero-foot">
				<span class="badge-solid">{user?.role?.name || 'USER'}</span>
				<span class="badge-outline">{visible.length} MENU</span>
			</div>
		</section>

		<!-- center asymmetric bento grid: slot halaman -->
		<section class="bento__main">
			<div class="bento__main-head">
				<span class="kicker">02 / WORKSPACE</span>
				<span class="section-title">{activeLabel.toUpperCase()}</span>
			</div>
			<div class="neo-window">
				<div class="neo-window__bar">
					<span class=" neo-window__dots"><i></i><i></i><i></i></span>
					<span>SOPQ / {activeApp.toUpperCase()} / {activeLabel}</span>
					<svg viewBox="0 0 24 24" class="neo-ico"><path d={neoIconFor(currentPath)} /></svg>
				</div>
				<div class="neo-window__body"><slot /></div>
			</div>
		</section>

		<!-- right vertical activity feed -->
		<aside class="bento__feed">
			<div class="feed-head">
				<span class="kicker">03 / AKTIVITAS</span>
				<span class="feed-dot"></span>
			</div>
			<ul class="feed-list">
				{#each visible.slice(0, 7) as item, i (item.href)}
					<li class="feed-item">
						<span class="feed-num">{String(i + 1).padStart(2, '0')}</span>
						<div>
							<a href={item.href} class="feed-link">{item.label}</a>
							<span class="feed-meta">{activeApp.toUpperCase()} · MENU</span>
						</div>
					</li>
				{/each}
			</ul>
			<div class="feed-user">
				<span class="feed-user__avatar">{initial}</span>
				<span>{user?.full_name || user?.username || 'USER'}</span>
			</div>
			<slot name="context" />
		</aside>

		<!-- bottom wide data table zone (slot again by page) -->
		<section class="bento__table">
			<div class="table-head">
				<span class="kicker">04 / DATA</span>
				<span class="section-title">TABEL & RINCIAN</span>
				<span class="table-triangle" aria-hidden="true"></span>
			</div>
			<div class="table-zone">
				<slot name="data">
					<p class="table-hint">Halaman dapat mengirimkan tabel ke zona ini lewat slot <code>data</code>.</p>
				</slot>
			</div>
		</section>
	</div>

	<!-- floating action button -->
	<button class="neo-fab press" title="Kembali ke atas" aria-label="Kembali ke atas" on:click={() => window.scrollTo({ top: 0, behavior: 'smooth' })}>
		<svg viewBox="0 0 24 24" class="neo-ico"><path d="M12 5l-7 7h4v7h6v-7h4z" /></svg>
	</button>

	<!-- ===== MOBILE: stacked editorial blocks ===== -->
	<header class="neo-mobile-top">
		<span class="brand-mark">SOPQ<span class="brand-dot">.</span></span>
		<button class="neo-ico-btn press" aria-label="Buka menu" title="Menu" on:click={() => (menuOpen = !menuOpen)}>
			<svg viewBox="0 0 24 24" class="neo-ico"><path d={menuOpen ? NEO_ICONS.close : NEO_ICONS.menu} /></svg>
		</button>
	</header>

	{#if menuOpen}
		<nav class="neo-mobile-menu" aria-label="Menu">
			{#each visible as item (item.href)}
				<a href={item.href} class="press" class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))}>
					<svg viewBox="0 0 24 24" class="neo-ico"><path d={neoIconFor(item.href)} /></svg>
					{item.label}
				</a>
			{/each}
		</nav>
	{/if}

	<main class="neo-mobile-main">
		<h1 class="neo-mobile-title">{activeLabel}</h1>
		<section class="mobile-hero">
			<span class="kicker">FOKUS UTAMA</span>
			<div class="mobile-hero__number">{visible.length}</div>
			<p>Menu tersedia di modul {activeApp.toUpperCase()}</p>
		</section>

		<section class="mobile-card">
			<div class="mobile-card__head">
				<span class="kicker">WORKSPACE</span>
				<svg viewBox="0 0 24 24" class="neo-ico"><path d={neoIconFor(currentPath)} /></svg>
			</div>
			<slot />
		</section>

		<section class="mobile-swipe" aria-label="Pintasan">
			{#each visible.slice(0, 6) as item (item.href)}
				<a href={item.href} class="swipe-card press">
					<svg viewBox="0 0 24 24" class="neo-ico"><path d={neoIconFor(item.href)} /></svg>
					<span>{item.label}</span>
				</a>
			{/each}
		</section>

		<section class="mobile-card">
			<div class="mobile-card__head"><span class="kicker">AKTIVITAS</span></div>
			<ul class="mobile-feed">
				{#each visible.slice(0, 5) as item, i (item.href)}
					<li><span class="feed-num">{String(i + 1).padStart(2, '0')}</span><a href={item.href}>{item.label}</a></li>
				{/each}
			</ul>
		</section>
	</main>

	<nav class="neo-mobile-nav" aria-label="Navigasi bawah">
		{#each visible.slice(0, 5) as item (item.href)}
			<a href={item.href} class="press" class:active={currentPath === item.href || (item.href !== '/' && currentPath.startsWith(item.href))} aria-label={item.label} title={item.label}>
				<svg viewBox="0 0 24 24" class="neo-ico"><path d={neoIconFor(item.href)} /></svg>
			</a>
		{/each}
	</nav>
</div>

<style>
	.neo-root {
		--n-bg: #f7f3dc;
		--n-ink: #0d0d0d;
		--n-yellow: #ffe23d;
		--n-blue: #2f5bff;
		--n-red: #ff4d3d;
		--n-lime: #b7f35b;
		--n-line: 3px solid var(--n-ink);
		--n-shadow: 6px 6px 0 var(--n-ink);
		min-height: 100dvh;
		background: var(--n-bg);
		color: var(--n-ink);
		font-family: 'Inter', 'Arial Black', system-ui, sans-serif;
		overflow-x: hidden;
	}
	.neo-ico {
		width: 1.15rem; height: 1.15rem; fill: none; stroke: currentColor;
		stroke-width: 2.2; stroke-linecap: square; stroke-linejoin: miter;
	}
	.kicker { font-size: 0.68rem; font-weight: 900; letter-spacing: 0.12em; text-transform: uppercase; }
	.section-title { font-weight: 900; text-transform: uppercase; letter-spacing: -0.02em; }
	.press:active { transform: translate(3px, 3px); box-shadow: 2px 2px 0 var(--n-ink); }

	/* Top */
	.neo-top {
		display: grid;
		grid-template-columns: 1fr auto 1fr;
		align-items: center;
		gap: 1rem;
		padding: 0.7rem 1.2rem;
		border-bottom: var(--n-line);
		background: var(--n-yellow);
	}
	.brand-mark { font-size: 1.7rem; font-weight: 900; letter-spacing: -0.06em; line-height: 1; }
	.brand-dot { color: var(--n-red); }
	.brand-tag { font-size: 0.62rem; font-weight: 800; letter-spacing: 0.08em; }
	.neo-nav { display: flex; flex-wrap: wrap; justify-content: center; gap: 0.35rem; }
	.neo-nav__item {
		display: inline-flex; align-items: center; gap: 0.3rem;
		padding: 0.35rem 0.6rem;
		border: var(--n-line); box-shadow: 3px 3px 0 var(--n-ink);
		background: #fff; color: var(--n-ink);
		font-size: 0.72rem; font-weight: 900; text-decoration: none;
	}
	.neo-nav__item.active { background: var(--n-blue); color: #fff; }
	.neo-top__right { display: flex; align-items: center; justify-content: flex-end; gap: 0.6rem; }
	.neo-clock { font-weight: 900; font-variant-numeric: tabular-nums; font-size: 0.75rem; }
	.neo-ico-btn {
		display: grid; place-items: center; width: 2rem; height: 2rem;
		border: var(--n-line); box-shadow: 3px 3px 0 var(--n-ink);
		background: #fff; color: var(--n-ink); cursor: pointer;
	}
	.neo-avatar {
		display: grid; place-items: center; width: 2rem; height: 2rem;
		border: var(--n-line); background: var(--n-red); color: #fff; font-weight: 900;
	}

	/* Bento desktop */
	.bento {
		display: grid;
		grid-template-columns: 1.15fr 1.7fr 0.85fr;
		grid-template-rows: auto auto;
		gap: 1rem;
		padding: 1rem 1.2rem 5rem;
	}
	.bento__hero {
		grid-column: 1; grid-row: 1;
		position: relative; overflow: hidden;
		min-height: 14rem; padding: 1.2rem;
		border: var(--n-line); box-shadow: var(--n-shadow);
		background: var(--n-lime);
	}
	.oversized { font-size: clamp(1.9rem, 3.4vw, 3.4rem); font-weight: 900; letter-spacing: -0.06em; line-height: 0.95; margin: 0.6rem 0; text-transform: uppercase; }
	.hero-shape { position: absolute; right: -2rem; bottom: -3rem; width: 10rem; height: 10rem; border: var(--n-line); background: var(--n-yellow); transform: rotate(18deg); }
	.hero-desc { position: relative; z-index: 1; font-size: 0.8rem; font-weight: 700; max-width: 18rem; }
	.hero-foot { position: relative; z-index: 1; display: flex; gap: 0.5rem; margin-top: 1rem; }
	.badge-solid, .badge-outline { padding: 0.25rem 0.5rem; font-size: 0.65rem; font-weight: 900; border: var(--n-line); }
	.badge-solid { background: var(--n-ink); color: #fff; }
	.badge-outline { background: #fff; }

	.bento__main { grid-column: 2; grid-row: 1 / span 2; min-width: 0; }
	.bento__main-head { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; margin-bottom: 0.6rem; }
	.neo-window { border: var(--n-line); box-shadow: var(--n-shadow); background: #fff; }
	.neo-window__bar {
		display: flex; align-items: center; gap: 0.6rem;
		padding: 0.45rem 0.7rem; border-bottom: var(--n-line);
		background: var(--n-blue); color: #fff;
		font-size: 0.72rem; font-weight: 900; letter-spacing: 0.04em;
	}
	.neo-window__bar .neo-ico { margin-left: auto; }
	.neo-window__dots { display: flex; gap: 0.3rem; }
	.neo-window__dots i { width: 0.65rem; height: 0.65rem; border: 2px solid #fff; display: block; }
	.neo-window__body { max-height: calc(100dvh - 15rem); overflow: auto; padding: 0.8rem; }

	.bento__feed {
		grid-column: 3; grid-row: 1;
		border: var(--n-line); box-shadow: var(--n-shadow); background: #fff;
		display: flex; flex-direction: column; min-height: 14rem;
	}
	.feed-head { display: flex; align-items: center; justify-content: space-between; padding: 0.6rem; border-bottom: var(--n-line); background: var(--n-yellow); }
	.feed-dot { width: 0.7rem; height: 0.7rem; border: 2px solid var(--n-ink); border-radius: 50%; background: var(--n-red); }
	.feed-list { list-style: none; margin: 0; padding: 0.3rem; overflow: auto; flex: 1; }
	.feed-item { display: flex; gap: 0.5rem; padding: 0.45rem; border-bottom: 2px dotted #b9b5a2; }
	.feed-num { font-weight: 900; font-size: 0.7rem; color: var(--n-blue); }
	.feed-link { display: block; font-size: 0.75rem; font-weight: 800; color: var(--n-ink); text-decoration: none; }
	.feed-link:hover { text-decoration: underline; }
	.feed-meta { font-size: 0.6rem; font-weight: 700; color: #666; }
	.feed-user { display: flex; align-items: center; gap: 0.4rem; padding: 0.5rem; border-top: var(--n-line); font-size: 0.7rem; font-weight: 900; }
	.feed-user__avatar { display: grid; place-items: center; width: 1.6rem; height: 1.6rem; background: var(--n-lime); border: 2px solid var(--n-ink); }

	.bento__table { grid-column: 1 / span 3; grid-row: 2; align-self: end; }
	.table-head { display: flex; align-items: center; gap: 0.8rem; margin-bottom: 0.5rem; }
	.table-triangle { width: 0; height: 0; border-left: 0.7rem solid transparent; border-right: 0.7rem solid transparent; border-bottom: 1.1rem solid var(--n-red); }
	.table-zone { border: var(--n-line); box-shadow: var(--n-shadow); background: #fff; padding: 0.8rem; overflow: auto; }
	.table-hint { font-size: 0.75rem; font-weight: 700; color: #555; }

	.neo-fab {
		position: fixed; right: 1.2rem; bottom: 1.2rem; z-index: 60;
		width: 3.2rem; height: 3.2rem;
		display: grid; place-items: center;
		border: var(--n-line); box-shadow: 5px 5px 0 var(--n-ink);
		background: var(--n-red); color: #fff; cursor: pointer;
	}

	/* ===== MOBILE: stacked editorial blocks ===== */
	.neo-mobile-top, .neo-mobile-main, .neo-mobile-nav, .neo-mobile-menu { display: none; }
	@media (max-width: 900px) {
		.neo-top, .bento, .neo-fab { display: none; }
		.neo-mobile-top {
			position: sticky; top: 0; z-index: 60;
			display: flex; align-items: center; justify-content: space-between;
			padding: 0.6rem 0.8rem; border-bottom: var(--n-line); background: var(--n-yellow);
		}
		.neo-mobile-menu {
			position: sticky; top: 3.3rem; z-index: 59;
			display: grid; grid-template-columns: 1fr 1fr; gap: 0.4rem;
			padding: 0.6rem; border-bottom: var(--n-line); background: #fff;
		}
		.neo-mobile-menu a {
			display: flex; align-items: center; gap: 0.4rem;
			padding: 0.5rem; border: 2px solid var(--n-ink);
			font-size: 0.72rem; font-weight: 900; color: var(--n-ink); text-decoration: none;
		}
		.neo-mobile-menu a.active { background: var(--n-blue); color: #fff; }
		.neo-mobile-main { display: block; padding: 0.8rem 0.8rem 5.5rem; }
		.neo-mobile-title { font-size: 2.4rem; font-weight: 900; letter-spacing: -0.06em; line-height: 0.95; text-transform: uppercase; margin: 0.8rem 0 1rem; }
		.mobile-hero {
			position: relative; padding: 1rem; margin-bottom: 1rem;
			border: var(--n-line); box-shadow: 5px 5px 0 var(--n-ink); background: var(--n-yellow);
		}
		.mobile-hero__number { font-size: 4rem; font-weight: 900; line-height: 1; }
		.mobile-hero p { font-size: 0.75rem; font-weight: 800; margin: 0.4rem 0 0; }
		.mobile-card { border: var(--n-line); box-shadow: 5px 5px 0 var(--n-ink); background: #fff; margin-bottom: 1rem; }
		.mobile-card__head { display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0.7rem; border-bottom: var(--n-line); background: var(--n-blue); color: #fff; }
		.mobile-card > :not(.mobile-card__head) { padding: 0.7rem; }
		.mobile-swipe { display: flex; gap: 0.6rem; overflow-x: auto; padding-bottom: 0.5rem; margin-bottom: 1rem; scroll-snap-type: x mandatory; }
		.swipe-card {
			flex: 0 0 8.5rem; scroll-snap-align: start;
			display: flex; flex-direction: column; gap: 0.4rem;
			padding: 0.7rem; border: var(--n-line); box-shadow: 4px 4px 0 var(--n-ink);
			background: var(--n-lime); color: var(--n-ink); text-decoration: none;
			font-size: 0.75rem; font-weight: 900;
		}
		.mobile-feed { list-style: none; margin: 0; padding: 0; }
		.mobile-feed li { display: flex; gap: 0.6rem; padding: 0.5rem 0; border-bottom: 2px dotted #b9b5a2; font-weight: 800; font-size: 0.78rem; }
		.mobile-feed a { color: var(--n-ink); text-decoration: none; }
		.neo-mobile-nav {
			position: fixed; left: 0.5rem; right: 0.5rem; bottom: 0.5rem; z-index: 70;
			display: flex; justify-content: space-around;
			padding: 0.4rem; border: var(--n-line); box-shadow: 4px 4px 0 var(--n-ink); background: var(--n-yellow);
		}
		.neo-mobile-nav a {
			display: grid; place-items: center; width: 2.8rem; height: 2.4rem;
			border: 2px solid var(--n-ink); background: #fff; color: var(--n-ink);
		}
		.neo-mobile-nav a.active { background: var(--n-red); color: #fff; }
	}

	@media (prefers-reduced-motion: reduce) {
		.press:active { transform: none; box-shadow: 3px 3px 0 var(--n-ink); }
	}
</style>
