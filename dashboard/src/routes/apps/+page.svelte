<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';

	const MANAGER = ['admin', 'supervisor'];
	const STAFF = 'staff gudang';

	$: user = $authStore.user;
	$: roleName = user?.role?.name ?? '';
	$: isManager = MANAGER.includes(roleName);
	$: initial = (user?.full_name || user?.username || '?')?.[0]?.toUpperCase() ?? '?';
	$: greeting = greetingForHour(new Date().getHours());

	// Aplikasi yang boleh dilihat role tsb. Staff hanya Stock Opname.
	$: apps = isManager
		? [
				{
					href: '/',
					name: 'Stock Opname',
					desc: 'Audit stok fisik berkala, bandingkan dengan stok sistem',
					icon: 'app-opname',
					accent: 'var(--p)'
				},
				{
					href: '/inventory',
					name: 'Inventory Management',
					desc: 'Catat barang masuk, keluar, dan transfer antar gudang',
					icon: 'app-inventory',
					accent: 'var(--s)'
				}
		  ]
		: [
				{
					href: '/',
					name: 'Stock Opname',
					desc: 'Audit stok fisik berkala, bandingkan dengan stok sistem',
					icon: 'app-opname',
					accent: 'var(--p)'
				}
		  ];

	function greetingForHour(h: number): string {
		if (h >= 4 && h < 11) return 'Selamat pagi';
		if (h >= 11 && h < 15) return 'Selamat siang';
		if (h >= 15 && h < 18) return 'Selamat sore';
		return 'Selamat malam';
	}

	function toggleTheme() {
		const next = $themeStore.color === 'apple-dark' ? 'apple-light' : 'apple-dark';
		themeStore.setColor(next);
	}

	$: isDark = $themeStore.color === 'apple-dark';

	// Guard tambahan: staff mencoba akses inventory lewat URL langsung.
	let blocked = false;
	onMount(() => {
		// Tidak melakukan apa pun saat mount; blok hanya dipakai di /inventory.
	});

	function enter(href: string) {
		goto(href);
	}

	async function logout() {
		await authStore.logout();
	}
</script>

<svelte:head>
	<title>Aplikasi - SOPQ</title>
</svelte:head>

<div class="selector">
	<div class="ambient" aria-hidden="true">
		<div class="ambient__blob ambient__blob--a"></div>
		<div class="ambient__blob ambient__blob--b"></div>
	</div>

	<!-- Header -->
	<header class="sel-header">
		<div class="userchip glass">
			<div class="avatar">{initial}</div>
			<div class="userchip__meta">
				<div class="userchip__greet">
					{greeting}, <b>{user?.full_name || user?.username}</b>
				</div>
				<div class="userchip__role">{roleName}</div>
			</div>
		</div>
		<div class="sel-actions">
			<button class="iconbtn press" aria-label="Ganti tema" on:click={toggleTheme}>
				{#if isDark}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round">
						<circle cx="12" cy="12" r="4" />
						<path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M6.3 17.7l-1.4 1.4M19.1 4.9l-1.4 1.4" />
					</svg>
				{:else}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round">
						<path d="M21 12.8A9 9 0 1111.2 3a7 7 0 009.8 9.8z" />
					</svg>
				{/if}
			</button>
			<button class="iconbtn press" aria-label="Keluar" on:click={logout}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round">
					<path d="M15 12H4m0 0l3-3m-3 3l3 3M10 4h8a2 2 0 012 2v12a2 2 0 01-2 2h-8" />
				</svg>
			</button>
		</div>
	</header>

	<!-- App cards -->
	<main class="sel-main">
		<h1 class="sel-title">Pilih Aplikasi</h1>
		<p class="sel-sub">Kelola gudang Anda dengan SOPQ</p>

		<div class="cards" class:cards--single={apps.length === 1}>
			{#each apps as app, i (app.href)}
				<a
					href={app.href}
					class="appcard glass press"
					style="--accent: {app.accent}"
					in:fly={{ y: 16, duration: 420, delay: i * 90, easing: cubicOut }}
					on:click={() => enter(app.href)}
				>
					<div class="appcard__icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"
							stroke-linecap="round" stroke-linejoin="round">
							{#if app.icon === 'app-inventory'}
								<path d="M3 7l9-4 9 4v10l-9 4-9-4V7zM3 7l9 4 9-4M12 11v10" />
							{:else}
								<path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
							{/if}
						</svg>
					</div>
					<div class="appcard__name">{app.name}</div>
					<div class="appcard__desc">{app.desc}</div>
				</a>
			{/each}
		</div>

		{#if roleName === STAFF}
			<p class="sel-note">Inventory Management hanya tersedia untuk Admin & Supervisor.</p>
		{/if}
	</main>
</div>

<style>
	.selector {
		min-height: 100dvh;
		position: relative;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		background: hsl(var(--b2));
		color: hsl(var(--bc));
	}
	.ambient {
		position: fixed;
		inset: 0;
		pointer-events: none;
		z-index: 0;
	}
	.ambient__blob {
		position: absolute;
		border-radius: 50%;
		filter: blur(90px);
		opacity: 0.4;
	}
	.ambient__blob--a {
		width: 40rem;
		height: 40rem;
		top: -14rem;
		left: -8rem;
		background: hsl(var(--p) / 0.3);
		animation: floatA 26s var(--ease-apple) infinite alternate;
	}
	.ambient__blob--b {
		width: 36rem;
		height: 36rem;
		bottom: -12rem;
		right: -10rem;
		background: hsl(var(--s) / 0.22);
		animation: floatB 32s var(--ease-apple) infinite alternate;
	}
	@keyframes floatA {
		from { transform: translate3d(0, 0, 0) scale(1); }
		to { transform: translate3d(5rem, 5rem, 0) scale(1.12); }
	}
	@keyframes floatB {
		from { transform: translate3d(0, 0, 0) scale(1); }
		to { transform: translate3d(-5rem, -4rem, 0) scale(1.1); }
	}

	.sel-header {
		position: relative;
		z-index: 1;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1.25rem 1.5rem;
	}
	.userchip {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		border-radius: 1rem;
		padding: 0.5rem 0.9rem 0.5rem 0.5rem;
	}
	.avatar {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: hsl(var(--p) / 0.16);
		color: hsl(var(--p));
		font-weight: 700;
		flex-shrink: 0;
	}
	.userchip__greet {
		font-size: 0.9rem;
	}
	.userchip__greet b {
		font-weight: 650;
	}
	.userchip__role {
		font-size: 0.74rem;
		color: hsl(var(--bc) / 0.55);
		text-transform: capitalize;
	}
	.sel-actions {
		display: flex;
		align-items: center;
		gap: 0.4rem;
	}
	.iconbtn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2.4rem;
		height: 2.4rem;
		border-radius: 0.75rem;
		color: hsl(var(--bc));
	}
	.iconbtn svg {
		width: 1.3rem;
		height: 1.3rem;
	}

	.sel-main {
		position: relative;
		z-index: 1;
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 1rem 1.5rem 3rem;
		text-align: center;
	}
	.sel-title {
		font-size: clamp(1.75rem, 4vw, 2.6rem);
		font-weight: 700;
		letter-spacing: -0.03em;
	}
	.sel-sub {
		margin-top: 0.4rem;
		font-size: 1rem;
		color: hsl(var(--bc) / 0.6);
	}

	.cards {
		margin-top: 2.5rem;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1.5rem;
		width: 100%;
		max-width: 46rem;
	}
	.cards--single {
		grid-template-columns: minmax(0, 24rem);
		justify-content: center;
	}
	.appcard {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		padding: 2rem 1.5rem;
		border-radius: 1.5rem;
		text-decoration: none;
		color: hsl(var(--bc));
		transition:
			transform 280ms var(--ease-apple),
			box-shadow 280ms var(--ease-apple);
	}
	.appcard:hover {
		transform: translateY(-6px) scale(1.02);
		box-shadow:
			0 4px 8px hsl(var(--bc) / 0.06),
			0 24px 56px hsl(var(--bc) / 0.14);
	}
	.appcard__icon {
		width: 4.5rem;
		height: 4.5rem;
		border-radius: 1.15rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: color-mix(in srgb, var(--accent) 16%, transparent);
		color: color-mix(in srgb, var(--accent) 90%, hsl(var(--bc)));
		transition: transform 280ms var(--ease-apple);
	}
	.appcard:hover .appcard__icon {
		transform: scale(1.06) translateY(-2px);
	}
	.appcard__icon svg {
		width: 2.4rem;
		height: 2.4rem;
	}
	.appcard__name {
		margin-top: 0.35rem;
		font-size: 1.2rem;
		font-weight: 650;
		letter-spacing: -0.01em;
	}
	.appcard__desc {
		font-size: 0.85rem;
		line-height: 1.45;
		color: hsl(var(--bc) / 0.6);
		max-width: 18rem;
	}
	.sel-note {
		margin-top: 2rem;
		font-size: 0.8rem;
		color: hsl(var(--bc) / 0.5);
	}

	@media (max-width: 640px) {
		.cards {
			grid-template-columns: minmax(0, 1fr);
			gap: 1rem;
			max-width: 24rem;
		}
	}
</style>
