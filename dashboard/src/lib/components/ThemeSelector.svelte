<script lang="ts">
	import { themeStore, THEME_LAYOUTS } from '$lib/stores/theme';
	import type { ThemeLayout } from '$lib/stores/theme';
	import { get } from 'svelte/store';

	export let currentLayout: ThemeLayout;

	let currentTheme = get(themeStore);
	themeStore.subscribe((t) => (currentTheme = t));

	$: previewLayouts = THEME_LAYOUTS.map((layout) => ({
		...layout,
		previewDataUi: layout.id === 'focus' || layout.id === 'workspace' ? 'apple' : layout.id
	}));
</script>

<div>
	<h3 class="text-sm font-medium">Layout</h3>
	<p class="mt-1 text-xs text-base-content/60">
		Warna dan gaya konten mengikuti layout yang dipilih — pengaturan warna tidak diperlukan.
	</p>
	<div class="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
		{#each previewLayouts as layout (layout.id)}
			<button
				class="theme-card rounded-xl border-2 bg-base-100 p-3 text-left transition-colors hover:bg-base-300"
				class:border-primary={currentLayout === layout.id}
				class:border-base-300={currentLayout !== layout.id}
				aria-pressed={currentLayout === layout.id}
				on:click={() => themeStore.setLayout(layout.id)}
			>
				<div class="theme-card__preview" data-ui={layout.previewDataUi}>
					<div class="theme-card__panel">
						<span class="theme-card__bar"></span>
						<span class="theme-card__bar theme-card__bar--short"></span>
						<span class="theme-card__button"></span>
					</div>
				</div>
				<div class="mt-2 flex items-center justify-between gap-2">
					<div class="text-sm font-semibold leading-tight">{layout.name}</div>
					{#if currentLayout === layout.id}
						<span class="badge badge-primary badge-xs">Aktif</span>
					{/if}
				</div>
				<p class="mt-0.5 text-xs text-base-content/60">{layout.description}</p>
			</button>
		{/each}
	</div>
</div>

<style>
	/* Mini preview uses same data-ui token scopes as the real layouts. */
	.theme-card__preview {
		height: 3.75rem;
		border-radius: 0.6rem;
		padding: 0.5rem;
		overflow: hidden;
		background: var(--ui-app);
		color: var(--ui-text);
		font-family: var(--ui-font);
		box-shadow: var(--ui-shadow-inset);
	}
	.theme-card__panel {
		height: 100%;
		padding: 0.45rem 0.55rem;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		background: var(--ui-card);
		border: 1px solid var(--ui-border);
		border-radius: var(--ui-radius);
		box-shadow: var(--ui-shadow-card);
	}
	.theme-card__bar {
		height: 0.22rem;
		width: 82%;
		background: var(--ui-text-muted);
		border-radius: 999px;
	}
	.theme-card__bar--short {
		width: 55%;
	}
	.theme-card__button {
		margin-top: auto;
		height: 0.35rem;
		width: 2rem;
		background: var(--ui-accent);
		border-radius: var(--ui-radius);
	}
	.theme-card__preview[data-ui='neo'] .theme-card__panel,
	.theme-card__preview[data-ui='retro'] .theme-card__panel,
	.theme-card__preview[data-ui='editorial'] .theme-card__panel {
		border-radius: 0;
	}
	.theme-card__preview[data-ui='neo'] .theme-card__panel {
		border-width: 2px;
		border-color: var(--ui-border-strong);
		box-shadow: 3px 3px 0 var(--ui-border-strong);
	}
	.theme-card__preview[data-ui='retro'] .theme-card__panel {
		border-width: 2px;
		border-color: var(--ui-border-strong) #fff #fff var(--ui-border-strong);
		box-shadow: 2px 2px 0 var(--ui-border-strong);
	}
	.theme-card__preview[data-ui='neumo'] .theme-card__panel {
		border-color: transparent;
		box-shadow: 3px 3px 8px var(--ui-shadow-dark), -3px -3px 8px var(--ui-shadow-light);
	}
	.theme-card__preview[data-ui='acrylic'] .theme-card__panel {
		backdrop-filter: blur(8px);
	}
</style>
