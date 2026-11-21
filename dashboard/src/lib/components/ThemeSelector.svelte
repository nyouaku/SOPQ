<script lang="ts">
	import { themeStore, THEME_LAYOUTS } from '$lib/stores/theme';
	import type { ThemeLayout } from '$lib/stores/theme';
	import { get } from 'svelte/store';

	export let currentLayout: ThemeLayout;

	let currentTheme = get(themeStore);
	themeStore.subscribe((t) => (currentTheme = t));
</script>

<div>
	<h3 class="text-sm font-medium">Layout</h3>
	<p class="mt-1 text-xs text-base-content/60">
		Warna dan gaya konten mengikuti layout yang dipilih — pengaturan warna tidak diperlukan.
	</p>
	<div class="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
		{#each THEME_LAYOUTS as layout}
			<button
				class="rounded-xl border-2 bg-base-100 p-3 text-left transition-colors hover:bg-base-300"
				class:border-primary={currentLayout === layout.id}
				class:border-base-300={currentLayout !== layout.id}
				aria-pressed={currentLayout === layout.id}
				on:click={() => themeStore.setLayout(layout.id)}
			>
				<div class="text-2xl">{layout.icon}</div>
				<div class="mt-1.5 text-sm font-semibold leading-tight">{layout.name}</div>
				<p class="mt-0.5 text-xs text-base-content/60">{layout.description}</p>
				{#if currentLayout === layout.id}
					<span class="badge badge-primary badge-xs mt-2">Aktif</span>
				{/if}
			</button>
		{/each}
	</div>
</div>
