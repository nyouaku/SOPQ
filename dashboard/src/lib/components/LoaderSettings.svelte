<script lang="ts">
	import Loader from '$lib/components/Loader.svelte';
	import {
		loaderStore,
		PRELOADER_STYLES,
		LOADER_STYLES,
		type PreloaderStyle,
		type LoaderStyle
	} from '$lib/stores/loader';

	const PRELOADER_LABELS: Record<PreloaderStyle, string> = {
		scan: 'Scan',
		rings: 'Rings',
		boxes: 'Boxes',
		orbit: 'Orbit',
		skeleton: 'Skeleton'
	};

	const LOADER_LABELS: Record<LoaderStyle, string> = {
		spinner: 'Spinner',
		dots: 'Dots',
		bars: 'Bars',
		ring: 'Ring',
		barcode: 'Barcode'
	};
</script>

<div class="space-y-5">
	<div>
		<h3 class="text-xs font-medium uppercase tracking-wide text-base-content/50">Preloader</h3>
		<p class="mt-1 text-xs text-base-content/50">
			Tampil sekali saat aplikasi dibuka per sesi tab.
		</p>
		<div class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-3">
			{#each PRELOADER_STYLES as style}
				<button
					class="rounded-xl border p-3 text-center transition-colors"
					class:border-primary={$loaderStore.preloader === style}
					class:border-base-300={$loaderStore.preloader !== style}
					aria-pressed={$loaderStore.preloader === style}
					on:click={() => loaderStore.setPreloader(style)}
				>
					<span class="block text-sm font-medium">{PRELOADER_LABELS[style]}</span>
				</button>
			{/each}
		</div>
	</div>

	<div>
		<h3 class="text-xs font-medium uppercase tracking-wide text-base-content/50">
			Loader (saat memuat data)
		</h3>
		<div class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-3">
			{#each LOADER_STYLES as style}
				<button
					class="flex flex-col items-center gap-2 rounded-xl border p-3 transition-colors"
					class:border-primary={$loaderStore.loader === style}
					class:border-base-300={$loaderStore.loader !== style}
					aria-pressed={$loaderStore.loader === style}
					on:click={() => loaderStore.setLoader(style)}
				>
					<Loader size={26} style={style} />
					<span class="text-xs font-medium">{LOADER_LABELS[style]}</span>
				</button>
			{/each}
		</div>
	</div>

	<label class="flex cursor-pointer items-center justify-between gap-3 rounded-xl border border-base-300 p-3">
		<span>
			<span class="block text-sm font-medium">Tampilkan preloader</span>
			<span class="block text-xs text-base-content/60">Matikan bila ingin aplikasi langsung tampil.</span>
		</span>
		<input
			type="checkbox"
			class="toggle toggle-primary toggle-sm"
			checked={$loaderStore.enabled}
			on:change={(e) => loaderStore.setEnabled(e.currentTarget.checked)}
		/>
	</label>

	<p class="text-xs text-base-content/50">
		Preloader yang dipilih aktif pada pembukaan berikutnya. Loader langsung berubah.
	</p>
</div>
