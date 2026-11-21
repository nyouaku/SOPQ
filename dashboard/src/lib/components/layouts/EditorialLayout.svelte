<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import { MENU_ITEMS, menuForApp, activeAppFor, activeMenuHref } from './menu';
	import { EDITORIAL_ICONS, editorialIconFor } from './editorial-icons';

	$: currentPath = $page.url.pathname;
	$: activeApp = activeAppFor(currentPath);
	$: visible = menuForApp(MENU_ITEMS, activeApp, $authStore.user?.role?.name);
	$: activeLabel =
		visible.find((i) => i.href === activeMenuHref(currentPath, visible))
			?.label ?? (activeApp === 'inventory' ? 'Inventory' : 'Stock Opname');
	$: user = $authStore.user;

	let mobileNav = false;
	let searchOpen = false;
	let searchQuery = '';
	let today = '';
	$: if (currentPath) mobileNav = false;

	function searchSubmit(e: Event) {
		e.preventDefault();
		if (!searchQuery.trim()) return;
		goto(`/products?search=${encodeURIComponent(searchQuery.trim())}`);
		searchOpen = false;
	}

	onMount(() => {
		today = new Date().toLocaleDateString('id-ID', {
			weekday: 'long', year: 'numeric', month: 'long', day: 'numeric'
		});
	});
</script>

<div class="ed-root">
	<!-- ===== MASTHEAD ===== -->
	<header class="masthead">
		<div class="masthead__meta">
			<span>VOL. 21 · NO. 08</span>
			<span>{today}</span>
		</div>
		<div class="masthead__brand">
			<span class="brand-sopq">SOPQ</span>
			<span class="brand-edition">— EDITION</span>
		</div>
		<div class="masthead__right">
			<button class="ed-icon press" aria-label="Cari" title="Cari" on:click={() => (searchOpen = !searchOpen)}>
				<svg viewBox="0 0 24 24" class="ed-ico"><path d={EDITORIAL_ICONS.search} /></svg>
			</button>
			<a href="/settings" class="ed-icon press" aria-label="Pengaturan" title="Pengaturan">
				<svg viewBox="0 0 24 24" class="ed-ico"><path d={EDITORIAL_ICONS.settings} /></svg>
			</a>
			<span class="masthead__user">{user?.full_name || user?.username || 'READER'}</span>
		</div>
	</header>

	{#if searchOpen}
		<form class="search-strip" on:submit={searchSubmit}>
			<svg viewBox="0 0 24 24" class="ed-ico"><path d={EDITORIAL_ICONS.search} /></svg>
			<input bind:value={searchQuery} placeholder="Cari produk… (Enter membuka Products)" aria-label="Cari" />
			<button type="submit" class="press">Cari</button>
			<button type="button" class="press" aria-label="Tutup pencarian" title="Tutup" on:click={() => (searchOpen = false)}>✕</button>
		</form>
	{/if}

	<!-- ===== EDITORIAL NAVIGATION ===== -->
	<nav class="ed-nav" aria-label="Navigasi editorial">
		{#each visible.slice(0, 8) as item (item.href)}
			<a
				href={item.href}
				class:active={activeMenuHref(currentPath, visible) === item.href}
			>{item.label}</a>
		{/each}
		<span class="ed-nav__end">{activeApp === 'inventory' ? 'INVENTORY' : 'STOCK OPNAME'}</span>
	</nav>

	<!-- ===== DESKTOP: asymmetric two-column publication ===== -->
	<div class="ed-desktop">
		<div class="ed-columns">
			<!-- LEFT: feature story / main insight -->
			<article class="ed-feature">
				<div class="section-label"><span>01</span> FEATURE / {activeApp.toUpperCase()}</div>
				<h1>{activeLabel}</h1>
				<p class="ed-standfirst">
					Laporan kerja dan data operasional disajikan dalam satu edisi — dibaca seperti publikasi,
					dikelola seperti sistem.
				</p>
				<div class="ed-rule"></div>
				<div class="ed-stats">
					<div><strong>{visible.length}</strong><span>MODUL TERSEDIA</span></div>
					<div><strong>100%</strong><span>SISTEM SIAP</span></div>
					<div><strong>{activeApp === 'inventory' ? 'INV' : 'SO'}</strong><span>EDISI AKTIF</span></div>
				</div>
				<div class="ed-rule"></div>
				<!-- Main content printed as the feature body -->
				<div class="ed-body"><slot /></div>
			</article>

			<!-- RIGHT: compact sidebar -->
			<aside class="ed-sidebar">
				<div class="section-label"><span>02</span> HIGHLIGHTS</div>
				<ul class="ed-highlights">
					{#each visible.slice(0, 6) as item, i (item.href)}
						<li>
							<span class="hl-num">{String(i + 1).padStart(2, '0')}</span>
							<a href={item.href}>{item.label}</a>
							<svg viewBox="0 0 24 24" class="ed-ico"><path d={editorialIconFor(item.href)} /></svg>
						</li>
					{/each}
				</ul>
				<div class="ed-rule"></div>
				<blockquote class="ed-quote">
					“Data yang dibaca dengan jelas menghasilkan keputusan yang tepat.”
					<cite>— SOPQ, catatan editorial</cite>
				</blockquote>
				<div class="ed-rule"></div>
				<div class="ed-side-meta">
					<span>PENYUNTING</span>
					<strong>{user?.full_name || user?.username || 'SOPQ'}</strong>
					<span>HALAMAN {activeApp.toUpperCase()}</span>
				</div>
				<slot name="context" />
			</aside>
		</div>

		<!-- BELOW: full-width analytical data section -->
		<section class="ed-data">
			<div class="section-label"><span>03</span> DATA & TABULASI</div>
			<div class="ed-data-inset"><slot name="data"><p class="ed-muted">Bagian data tersedia ketika halaman mengirimkan konten ke slot <code>data</code>.</p></slot></div>
		</section>
	</div>

	<!-- ===== MOBILE: long-form reading ===== -->
	<header class="ed-mobile-top">
		<span class="brand-sopq small">SOPQ</span>
		<span class="mobile-ed">— EDITION</span>
		<button class="ed-icon press" aria-label="Menu" title="Menu" on:click={() => (mobileNav = !mobileNav)}>
			<svg viewBox="0 0 24 24" class="ed-ico"><path d={mobileNav ? EDITORIAL_ICONS.close : EDITORIAL_ICONS.menu} /></svg>
		</button>
	</header>

	{#if mobileNav}
		<nav class="ed-mobile-tabs" aria-label="Kategori">
			{#each visible as item (item.href)}
				<a href={item.href} class:active={activeMenuHref(currentPath, visible) === item.href}>{item.label}</a>
			{/each}
		</nav>
	{/if}

	<main class="ed-mobile-main">
		<div class="section-label"><span>01</span> {activeApp.toUpperCase()}</div>
		<h1 class="ed-mobile-title">{activeLabel}</h1>
		<div class="ed-mobile-stat">
			<strong>{visible.length}</strong>
			<span>MODUL DALAM EDISI INI</span>
		</div>
		<div class="ed-rule"></div>
		<div class="ed-mobile-body"><slot /></div>
		<div class="ed-rule"></div>
		<div class="section-label"><span>02</span> SOROTAN</div>
		<ul class="ed-mobile-highlights">
			{#each visible.slice(0, 6) as item, i (item.href)}
				<li><span class="hl-num">{String(i + 1).padStart(2, '0')}</span><a href={item.href}>{item.label}</a></li>
			{/each}
		</ul>
		<div class="ed-rule"></div>
		<div class="section-label"><span>03</span> DATA &amp; TABULASI</div>
		<div class="ed-mobile-data"><slot name="data"><p class="ed-muted">Data tidak tersedia untuk halaman ini.</p></slot></div>
	</main>

	<nav class="ed-mobile-nav" aria-label="Navigasi bawah">
		{#each visible.slice(0, 5) as item (item.href)}
			<a href={item.href} class:active={activeMenuHref(currentPath, visible) === item.href} aria-label={item.label} title={item.label}>
				<svg viewBox="0 0 24 24" class="ed-ico"><path d={editorialIconFor(item.href)} /></svg>
			</a>
		{/each}
	</nav>

	<footer class="ed-footer">
		<span>SOPQ · EDITORIAL EDITION</span>
		<span>{today}</span>
		<span>{user?.role?.name || 'READER'}</span>
	</footer>
</div>

<style>
	.ed-root {
		--e-bg: #faf8f2;
		--e-ink: #1b1b1b;
		--e-muted: #5e5e5e;
		--e-accent: #7b2d3a;
		--e-line: #cfcbc0;
		min-height: 100dvh;
		background: var(--e-bg);
		color: var(--e-ink);
		font-family: 'Inter', system-ui, sans-serif;
	}
	.ed-ico {
		width: 1.05rem; height: 1.05rem; flex: none;
		fill: none; stroke: currentColor; stroke-width: 1.6;
		stroke-linecap: round; stroke-linejoin: round;
	}
	.press:active { transform: translateY(1px); opacity: 0.75; }
	.ed-rule { border-top: 1px solid var(--e-line); margin: 1.2rem 0; }
	.section-label {
		display: flex; align-items: center; gap: 0.5rem;
		font-size: 0.65rem; font-weight: 800; letter-spacing: 0.14em; text-transform: uppercase;
		color: var(--e-muted); margin-bottom: 0.8rem;
	}
	.section-label span { color: var(--e-accent); font-weight: 900; }
	.ed-muted { color: var(--e-muted); font-size: 0.8rem; }

	/* Masthead */
	.masthead {
		display: grid; grid-template-columns: 1fr auto 1fr; align-items: end;
		gap: 1rem; padding: 1.2rem 2rem 0.8rem; border-bottom: 3px double var(--e-ink);
	}
	.masthead__meta, .masthead__right { display: flex; align-items: center; gap: 0.7rem; font-size: 0.62rem; letter-spacing: 0.08em; color: var(--e-muted); text-transform: uppercase; }
	.masthead__right { justify-content: flex-end; }
	.masthead__brand { text-align: center; }
	.brand-sopq { font-family: Georgia, 'Times New Roman', serif; font-size: 3rem; font-weight: 900; letter-spacing: -0.05em; line-height: 0.9; }
	.brand-sopq.small { font-size: 1.6rem; }
	.brand-edition { font-family: Georgia, serif; font-style: italic; font-size: 0.9rem; color: var(--e-accent); margin-left: 0.4rem; }
	.masthead__user { font-weight: 800; color: var(--e-ink); }
	.ed-icon { border: 0; background: transparent; color: var(--e-ink); cursor: pointer; padding: 0.25rem; text-decoration: none; }
	.ed-icon:hover { color: var(--e-accent); }

	.search-strip {
		display: flex; align-items: center; gap: 0.6rem;
		padding: 0.6rem 2rem; border-bottom: 1px solid var(--e-line);
	}
	.search-strip input { flex: 1; border: 0; border-bottom: 1px solid var(--e-ink); background: transparent; font-family: inherit; font-size: 0.85rem; outline: 0; padding: 0.25rem 0; }
	.search-strip button { border: 0; background: transparent; cursor: pointer; }

	/* Nav */
	.ed-nav {
		display: flex; flex-wrap: wrap; align-items: center; gap: 1.1rem;
		padding: 0.7rem 2rem; border-bottom: 1px solid var(--e-ink);
	}
	.ed-nav a {
		font-size: 0.72rem; font-weight: 700; letter-spacing: 0.04em;
		text-transform: uppercase; color: var(--e-muted); text-decoration: none;
		padding-bottom: 0.15rem; border-bottom: 1px solid transparent;
	}
	.ed-nav a:hover { color: var(--e-ink); border-bottom-color: var(--e-ink); }
	.ed-nav a.active { color: var(--e-accent); border-bottom-color: var(--e-accent); }
	.ed-nav__end { margin-left: auto; font-size: 0.62rem; letter-spacing: 0.1em; color: var(--e-muted); }

	/* Desktop two-column */
	.ed-desktop { max-width: 84rem; margin: 0 auto; padding: 2rem 2rem 4rem; }
	.ed-columns { display: grid; grid-template-columns: minmax(0, 1.9fr) minmax(0, 0.85fr); gap: 2.5rem; }
	.ed-feature h1 {
		font-family: Georgia, 'Times New Roman', serif;
		font-size: clamp(2.6rem, 5.2vw, 5.2rem);
		font-weight: 900; letter-spacing: -0.045em; line-height: 0.98;
		margin: 0 0 1rem;
	}
	.ed-standfirst { font-family: Georgia, serif; font-size: 1.15rem; line-height: 1.6; color: #343434; max-width: 44rem; }
	.ed-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1.5rem; }
	.ed-stats strong { display: block; font-family: Georgia, serif; font-size: 2.2rem; line-height: 1; }
	.ed-stats span { font-size: 0.6rem; letter-spacing: 0.12em; color: var(--e-muted); }
	.ed-body { font-size: 0.88rem; line-height: 1.65; }
	.ed-body :global(h1), .ed-body :global(h2), .ed-body :global(h3) { font-family: Georgia, serif; }

	.ed-sidebar { border-left: 1px solid var(--e-line); padding-left: 1.6rem; }
	.ed-highlights, .ed-mobile-highlights { list-style: none; margin: 0; padding: 0; }
	.ed-highlights li { display: flex; align-items: center; gap: 0.6rem; padding: 0.55rem 0; border-bottom: 1px solid var(--e-line); }
	.ed-highlights li a { flex: 1; font-size: 0.82rem; font-weight: 700; color: var(--e-ink); text-decoration: none; }
	.ed-highlights li a:hover { text-decoration: underline; }
	.hl-num { font-family: Georgia, serif; font-size: 0.72rem; color: var(--e-accent); font-weight: 900; }
	.ed-quote { margin: 0; font-family: Georgia, serif; font-size: 1.05rem; line-height: 1.5; font-style: italic; }
	.ed-quote cite { display: block; margin-top: 0.6rem; font-family: 'Inter', sans-serif; font-size: 0.65rem; font-style: normal; color: var(--e-muted); text-transform: uppercase; letter-spacing: 0.08em; }
	.ed-side-meta { display: flex; flex-direction: column; gap: 0.25rem; font-size: 0.62rem; color: var(--e-muted); letter-spacing: 0.08em; }
	.ed-side-meta strong { font-family: Georgia, serif; font-size: 1rem; color: var(--e-ink); letter-spacing: 0; }

	.ed-data { margin-top: 3rem; border-top: 3px double var(--e-ink); padding-top: 1.5rem; }
	.ed-data-inset { font-size: 0.86rem; }

	.ed-footer {
		display: flex; justify-content: space-between; gap: 1rem;
		padding: 1rem 2rem; border-top: 1px solid var(--e-ink);
		font-size: 0.62rem; letter-spacing: 0.1em; text-transform: uppercase; color: var(--e-muted);
	}

	/* Mobile: long-form single column */
	.ed-mobile-top, .ed-mobile-main, .ed-mobile-nav, .ed-mobile-tabs { display: none; }
	@media (max-width: 900px) {
		.masthead, .ed-nav, .ed-desktop, .ed-footer, .search-strip { display: none; }
		.ed-mobile-top {
			position: sticky; top: 0; z-index: 50;
			display: flex; align-items: center; gap: 0.5rem;
			padding: 0.7rem 1rem; border-bottom: 3px double var(--e-ink); background: var(--e-bg);
		}
		.mobile-ed { font-family: Georgia, serif; font-style: italic; font-size: 0.78rem; color: var(--e-accent); margin-right: auto; }
		.ed-mobile-tabs {
			display: flex; gap: 1rem; overflow-x: auto; padding: 0.6rem 1rem;
			border-bottom: 1px solid var(--e-ink); white-space: nowrap;
		}
		.ed-mobile-tabs a { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; color: var(--e-muted); text-decoration: none; }
		.ed-mobile-tabs a.active { color: var(--e-accent); text-decoration: underline; }
		.ed-mobile-main { display: block; padding: 1.5rem 1.2rem 5rem; }
		.ed-mobile-title { font-family: Georgia, serif; font-size: 3rem; font-weight: 900; letter-spacing: -0.045em; line-height: 1; margin: 0 0 1rem; }
		.ed-mobile-stat { display: flex; align-items: baseline; gap: 0.8rem; }
		.ed-mobile-stat strong { font-family: Georgia, serif; font-size: 4rem; line-height: 1; color: var(--e-accent); }
		.ed-mobile-stat span { font-size: 0.6rem; letter-spacing: 0.12em; color: var(--e-muted); }
		.ed-mobile-body { font-size: 0.9rem; line-height: 1.7; }
		.ed-mobile-highlights li { display: flex; gap: 0.6rem; padding: 0.6rem 0; border-bottom: 1px solid var(--e-line); }
		.ed-mobile-highlights li a { flex: 1; font-size: 0.85rem; font-weight: 700; color: var(--e-ink); text-decoration: none; }
		.ed-mobile-data { overflow-x: auto; -webkit-overflow-scrolling: touch; font-size: 0.84rem; }
		.ed-mobile-data :global(table) { min-width: 36rem; border-collapse: collapse; }
		.ed-mobile-data :global(th), .ed-mobile-data :global(td) { border-bottom: 1px solid var(--e-line); padding: 0.4rem 0.5rem; text-align: left; }
		.ed-mobile-nav {
			position: fixed; left: 0; right: 0; bottom: 0; z-index: 55;
			display: flex; justify-content: space-around;
			padding: 0.5rem 0.8rem; border-top: 1px solid var(--e-ink); background: var(--e-bg);
		}
		.ed-mobile-nav a {
			width: 2.6rem; height: 2.2rem; display: grid; place-items: center;
			color: var(--e-muted); text-decoration: none;
		}
		.ed-mobile-nav a.active { color: var(--e-accent); border-bottom: 2px solid var(--e-accent); }
	}

	@media (prefers-reduced-motion: reduce) {
		.press:active { transform: none; }
	}
</style>
