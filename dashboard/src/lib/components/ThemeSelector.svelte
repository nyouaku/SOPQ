<script lang="ts">
	import { themeStore, THEME_LAYOUTS, THEME_COLORS } from '$lib/stores/theme';
	import type { ThemeLayout, ThemeColor } from '$lib/stores/theme';

	export let currentLayout: ThemeLayout;
	export let currentColor: ThemeColor;
</script>

<!-- Layout Selector -->
<div>
	<h3 class="text-sm font-medium">Layout</h3>
	<div class="mt-3 grid grid-cols-2 gap-3 md:grid-cols-4">
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

<!-- Color Theme Selector -->
<div class="mt-6">
	<h3 class="text-sm font-medium">Warna</h3>
	<div class="mt-3 grid grid-cols-2 gap-3 md:grid-cols-3">
		{#each THEME_COLORS as color}
			<button
				class="rounded-xl border-2 bg-base-100 p-3 text-left transition-colors hover:bg-base-300"
				class:border-primary={currentColor === color.id}
				class:border-base-300={currentColor !== color.id}
				aria-pressed={currentColor === color.id}
				on:click={() => themeStore.setColor(color.id)}
			>
				<!-- Swatch: bg / primary / teks -->
				<div class="flex gap-1.5">
					<span
						class="h-7 w-7 rounded-md border"
						style="background-color: {color.preview.bg}; border-color: {color.preview.primary};"
					></span>
					<span class="h-7 w-7 rounded-md" style="background-color: {color.preview.primary};"></span>
					<span
						class="flex h-7 flex-1 items-center justify-center rounded-md border text-[11px]"
						style="background-color: {color.preview.bg}; color: {color.preview.text}; border-color: {color.preview
							.primary};"
					>
						Aa
					</span>
				</div>
				<div class="mt-1.5 text-sm font-semibold leading-tight">{color.name}</div>
				<p class="mt-0.5 text-xs text-base-content/60">{color.description}</p>
				{#if currentColor === color.id}
					<span class="badge badge-primary badge-xs mt-2">Aktif</span>
				{/if}
			</button>
		{/each}
	</div>
</div>
