<script lang="ts">
	import { themeStore, THEME_LAYOUTS, THEME_COLORS } from '$lib/stores/theme';
	import type { ThemeLayout, ThemeColor, ThemeMode } from '$lib/stores/theme';
	import { get } from 'svelte/store';

	export let currentLayout: ThemeLayout;
	export let currentColor: ThemeColor;

	let currentTheme = get(themeStore);
	themeStore.subscribe((t) => (currentTheme = t));
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

<!-- Mode Selector (Manual vs Auto) -->
<div class="mt-6">
	<h3 class="text-sm font-medium">Mode Tema</h3>
	<div class="mt-3 flex gap-2">
		<button
			class="btn btn-sm flex-1 {currentTheme.mode === 'manual' ? 'btn-primary' : 'btn-outline'}"
			on:click={() => themeStore.setMode('manual')}
		>
			🔧 Manual
		</button>
		<button
			class="btn btn-sm flex-1 {currentTheme.mode === 'auto' ? 'btn-primary' : 'btn-outline'}"
			on:click={() => themeStore.setMode('auto')}
		>
			🕐 Otomatis (Siang/Malam)
		</button>
	</div>
	<p class="mt-2 text-xs text-base-content/50">
		{#if currentTheme.mode === 'auto'}
			Tema otomatis berganti: gelap pada malam hari, terang pada siang hari.
		{:else}
			Pilih warna manual di bawah.
		{/if}
	</p>
</div>

<!-- Auto Mode Configuration (only shown if auto mode) -->
{#if currentTheme.mode === 'auto'}
	<div class="mt-4 rounded-xl border border-base-300 bg-base-200 p-4">
		<h4 class="text-sm font-semibold">Konfigurasi Otomatis</h4>

		<!-- Day/Night Colors -->
		<div class="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
			<div>
				<label class="label py-1">
					<span class="label-text text-xs">Warna Siang Hari</span>
				</label>
				<select
					class="select select-bordered select-sm w-full"
					value={currentTheme.dayColor || 'daylight-ops'}
					on:change={(e) => {
						const val = e.target.value;
						themeStore.setDayNightColors(val, currentTheme.nightColor || 'rack-room');
					}}
				>
					{#each THEME_COLORS as color}
						<option value={color.id}>{color.name}</option>
					{/each}
				</select>
			</div>
			<div>
				<label class="label py-1">
					<span class="label-text text-xs">Warna Malam Hari</span>
				</label>
				<select
					class="select select-bordered select-sm w-full"
					value={currentTheme.nightColor || 'rack-room'}
					on:change={(e) => {
						const val = e.target.value;
						themeStore.setDayNightColors(currentTheme.dayColor || 'daylight-ops', val);
					}}
				>
					{#each THEME_COLORS as color}
						<option value={color.id}>{color.name}</option>
					{/each}
				</select>
			</div>
		</div>

		<!-- Night Hours -->
		<div class="mt-3 grid grid-cols-2 gap-3">
			<div>
				<label class="label py-1" for="night-start">
					<span class="label-text text-xs">Jam Mulai Malam</span>
				</label>
				<input
					id="night-start"
					type="number"
					min="0"
					max="23"
					class="input input-bordered input-sm"
					value={currentTheme.nightStartHour ?? 18}
					on:change={(e) => themeStore.setNightHours(parseInt(e.target.value), currentTheme.dayStartHour ?? 6)}
				/>
			</div>
			<div>
				<label class="label py-1" for="day-start">
					<span class="label-text text-xs">Jam Mulai Siang</span>
				</label>
				<input
					id="day-start"
					type="number"
					min="0"
					max="23"
					class="input input-bordered input-sm"
					value={currentTheme.dayStartHour ?? 6}
					on:change={(e) => themeStore.setNightHours(currentTheme.nightStartHour ?? 18, parseInt(e.target.value))}
				/>
			</div>
		</div>

		<div class="mt-2 text-xs text-base-content/50">
			Contoh: Malam 18:00 - 06:00, Siang 06:00 - 18:00
		</div>
	</div>
{/if}

<!-- Color Theme Selector (only if manual mode) -->
{#if currentTheme.mode === 'manual'}
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
{/if}
