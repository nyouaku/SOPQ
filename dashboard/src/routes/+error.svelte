<script lang="ts">
	import { page } from '$app/stores';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';

	// Halaman error global — dipakai SvelteKit untuk 404 dan error tak terduga.
	$: status = $page.status;
	$: message =
		status === 404
			? 'Halaman tidak ditemukan.'
			: $page.error?.message || 'Terjadi kesalahan saat memuat halaman.';

	const TITLES: Record<number, string> = {
		404: 'Tidak ditemukan',
		500: 'Kesalahan server'
	};
</script>

<svelte:head>
	<title>{status} - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="mx-auto flex w-full max-w-lg flex-col items-center justify-center px-4 py-20 text-center">
		<p class="metric text-5xl font-semibold text-primary">{status}</p>
		<h1 class="mt-3 text-xl font-semibold">{TITLES[status] || 'Terjadi kesalahan'}</h1>
		<p class="mt-2 text-sm text-base-content/60">{message}</p>

		<div class="mt-6 flex flex-wrap items-center justify-center gap-2">
			<a href="/" class="btn btn-primary btn-sm">Ke dashboard</a>
			<button class="btn btn-ghost btn-sm" on:click={() => history.back()}>Kembali</button>
		</div>
	</div>
</LayoutWrapper>
