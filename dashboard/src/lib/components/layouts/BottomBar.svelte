<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { MENU_ITEMS, menuForApp, activeAppFor } from './menu';

	export let currentPath: string = '';

	// Tampilan bottom bar hanya muat 5 slot di layar sempit.
	// 4 menu utama tampil langsung; sisanya masuk panel "Lainnya".
	// Dihitung ulang tiap auth berubah, supaya menu sesuai role user.
	$: visible = menuForApp(MENU_ITEMS, activeAppFor(currentPath), $authStore.user?.role?.name);
	$: primary = visible.slice(0, 4);
	$: overflow = visible.slice(4);
	$: overflowActive = overflow.some((i) => i.href === currentPath);

	let showMore = false;

	// Tutup panel saat pindah halaman.
	let lastPath = '';
	$: if (currentPath !== lastPath) {
		lastPath = currentPath;
		showMore = false;
	}
</script>

<!-- Panel "Lainnya": daftar menu yang tak muat di bar -->
{#if showMore}
	<div class="more-sheet" role="dialog" aria-label="Menu lainnya">
		<ul class="more-sheet__list">
			{#each overflow as item}
				<li>
					<a
						href={item.href}
						class="more-sheet__item"
						class:more-sheet__item--active={currentPath === item.href}
						aria-current={currentPath === item.href ? 'page' : undefined}
					>
						<span class="more-sheet__emoji">{item.icon}</span>
						<span>{item.label}</span>
					</a>
				</li>
			{/each}
		</ul>
	</div>
	<button
		class="more-sheet__backdrop"
		aria-label="Tutup menu"
		on:click={() => (showMore = false)}
	></button>
{/if}

<!--
  BottomBar — bar navigasi ala Apple/MacBook Dock.
  - Mobile: menempel di bawah, frosted glass, safe-area aware.
  - Desktop: mengambang di bawah layar, rounded "pill" dengan shadow.
  Icon memakai SVG stroke (konsisten & tajam di semua DPI).
-->
<nav class="bottombar" aria-label="Navigasi utama">
	<ul class="bottombar__list">
		{#each primary as item}
			{@const active = currentPath === item.href}
			<li>
				<a
					href={item.href}
					class="bottombar__item"
					class:bottombar__item--active={active}
					aria-current={active ? 'page' : undefined}
					title={item.label}
				>
					<span class="bottombar__icon">
						{#if item.href === '/'}
							<!-- Grid / Dashboard -->
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
								<rect x="3" y="3" width="7" height="7" rx="1.5" />
								<rect x="14" y="3" width="7" height="7" rx="1.5" />
								<rect x="3" y="14" width="7" height="7" rx="1.5" />
								<rect x="14" y="14" width="7" height="7" rx="1.5" />
							</svg>
						{:else if item.href === '/scan'}
							<!-- Scan / Viewfinder -->
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
								<path d="M3 8V5a2 2 0 0 1 2-2h3M16 3h3a2 2 0 0 1 2 2v3M21 16v3a2 2 0 0 1-2 2h-3M8 21H5a2 2 0 0 1-2-2v-3" />
								<path d="M7 12h10" stroke-linecap="round" />
							</svg>
						{:else if item.href === '/sessions'}
							<!-- Clipboard / Sessions -->
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
								<rect x="5" y="4" width="14" height="17" rx="2" />
								<path d="M9 4V3a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v1" />
								<path d="M9 11h6M9 15h6" stroke-linecap="round" />
							</svg>
						{:else}
							<!-- Box / Products -->
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
								<path d="M12 2 3 7v10l9 5 9-5V7l-9-5Z" />
								<path d="M3 7l9 5 9-5M12 12v10" />
							</svg>
						{/if}
					</span>
					<span class="bottombar__label">{item.label}</span>
				</a>
			</li>
		{/each}

		<!-- Slot kelima: panel menu lainnya -->
		<li>
			<button
				class="bottombar__item w-full"
				class:bottombar__item--active={overflowActive}
				aria-expanded={showMore}
				aria-label="Menu lainnya"
				on:click={() => (showMore = !showMore)}
			>
				<span class="bottombar__icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
						<circle cx="5" cy="12" r="1.6" fill="currentColor" stroke="none" />
						<circle cx="12" cy="12" r="1.6" fill="currentColor" stroke="none" />
						<circle cx="19" cy="12" r="1.6" fill="currentColor" stroke="none" />
					</svg>
				</span>
				<span class="bottombar__label">Lainnya</span>
			</button>
		</li>
	</ul>
</nav>

<style>
	.bottombar {
		position: fixed;
		inset-inline: 0;
		bottom: 0;
		z-index: 40;
		display: flex;
		justify-content: center;
		pointer-events: none;
		padding: 0.5rem 0.75rem;
		/* Ruang aman untuk notch / home indicator iPhone */
		padding-bottom: calc(0.5rem + env(safe-area-inset-bottom, 0px));
		/* Paksa layer sendiri: di Chrome Android elemen fixed ber-backdrop-filter
		   kadang baru ter-paint setelah halaman digulir. */
		transform: translateZ(0);
		will-change: transform;
	}
	.bottombar__list {
		pointer-events: auto;
		display: grid;
		grid-auto-flow: column;
		grid-auto-columns: 1fr; /* setiap item sama lebar → icon & label presisi di tengah */
		align-items: center;
		width: 100%;
		max-width: 100%; /* mobile: isi hingga tepi layar, tak ada margin asimetris */
		list-style: none;
		margin: 0;
		padding: 0.35rem;
		background: color-mix(in srgb, hsl(var(--b2)) 82%, transparent);
		backdrop-filter: saturate(180%) blur(20px);
		-webkit-backdrop-filter: saturate(180%) blur(20px);
		border: 1px solid hsl(var(--b3));
		border-radius: 1.1rem;
		box-shadow: 0 8px 32px hsl(var(--bc) / 0.18);
	}

	.bottombar__item {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.2rem;
		padding: 0.4rem 0.15rem;
		min-width: 0; /* izinkan label menyusut & di-ellipsis */
		border-radius: 0.85rem;
		border: 0;
		background: transparent;
		color: hsl(var(--bc) / 0.55);
		text-decoration: none;
		cursor: pointer;
		transition:
			color 180ms cubic-bezier(0.32, 0.72, 0, 1),
			background-color 180ms cubic-bezier(0.32, 0.72, 0, 1),
			transform 180ms cubic-bezier(0.32, 0.72, 0, 1);
		-webkit-tap-highlight-color: transparent;
	}

	.bottombar__item:hover,
	.bottombar__item:focus-visible {
		color: hsl(var(--bc));
		background: hsl(var(--bc) / 0.06);
	}

	.bottombar__item:focus-visible {
		outline: 2px solid hsl(var(--p));
		outline-offset: 2px;
	}

	/* Rebound ala Dock: sedikit "melompat" saat disentuh */
	.bottombar__item:active {
		transform: scale(0.92);
	}

	.bottombar__item--active {
		color: hsl(var(--p));
	}

	.bottombar__item--active::after {
		content: '';
		position: absolute;
		top: -0.25rem;
		left: 50%;
		transform: translateX(-50%);
		width: 0.9rem;
		height: 3px;
		border-radius: 9999px;
		background: hsl(var(--p));
	}

	.bottombar__icon {
		display: block;
		width: 1.5rem;
		height: 1.5rem;
	}
	.bottombar__icon svg {
		width: 100%;
		height: 100%;
	}

	.bottombar__label {
		font-size: 0.625rem;
		font-weight: 500;
		letter-spacing: -0.01em;
		line-height: 1;
		max-width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-align: center;
	}

	/* Panel "Lainnya" */
	.more-sheet {
		position: fixed;
		z-index: 45;
		inset-inline: 0.75rem;
		bottom: calc(4.75rem + env(safe-area-inset-bottom, 0px));
		display: flex;
		justify-content: center;
	}

	.more-sheet__list {
		list-style: none;
		margin: 0;
		padding: 0.35rem;
		width: 100%;
		max-width: 22rem;
		border-radius: 1.1rem;
		border: 1px solid hsl(var(--b3));
		background: color-mix(in srgb, hsl(var(--b2)) 92%, transparent);
		backdrop-filter: saturate(180%) blur(20px);
		-webkit-backdrop-filter: saturate(180%) blur(20px);
		box-shadow: 0 12px 40px hsl(var(--bc) / 0.22);
	}

	.more-sheet__item {
		display: flex;
		align-items: center;
		gap: 0.65rem;
		padding: 0.7rem 0.85rem;
		border-radius: 0.8rem;
		color: hsl(var(--bc) / 0.8);
		text-decoration: none;
		font-size: 0.875rem;
	}
	.more-sheet__item:hover,
	.more-sheet__item:focus-visible {
		background: hsl(var(--bc) / 0.06);
		color: hsl(var(--bc));
	}
	.more-sheet__item--active {
		color: hsl(var(--p));
		font-weight: 600;
	}
	.more-sheet__emoji {
		font-size: 1rem;
		line-height: 1;
	}

	.more-sheet__backdrop {
		position: fixed;
		inset: 0;
		z-index: 44;
		border: 0;
		padding: 0;
		background: hsl(var(--bc) / 0.25);
		cursor: default;
	}

	/* Tablet ke atas: bar mengambang di tengah bawah, tidak penuh */
	@media (min-width: 768px) {
		.bottombar {
			padding: 1rem;
			padding-bottom: calc(1rem + env(safe-area-inset-bottom, 0px));
		}
		.bottombar__list {
			width: auto;
			max-width: none;
			grid-auto-columns: auto;
			gap: 0.5rem;
			padding: 0.5rem 0.75rem;
			border-radius: 1.5rem;
		}
		.bottombar__item {
			min-width: 4.5rem;
			padding: 0.5rem 0.75rem;
			gap: 0.3rem;
		}
		.bottombar__icon {
			width: 1.75rem;
			height: 1.75rem;
		}
		.bottombar__label {
			font-size: 0.7rem;
		}
		.bottombar__item--active::after {
			top: -0.35rem;
		}
		.more-sheet {
			bottom: calc(6rem + env(safe-area-inset-bottom, 0px));
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.bottombar__item {
			transition: none;
		}
	}
</style>
