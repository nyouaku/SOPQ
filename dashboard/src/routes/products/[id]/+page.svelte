<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import StockTrendChart from '$lib/components/StockTrendChart.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { localDateTime } from '$lib/utils/date';

	$: productId = $page.params.id;

	// Konteks asal: halaman ini bisa dibuka dari SO (baca-saja) maupun Inventory
	// (kelola). Breadcrumb "Products" kembali ke daftar yang sesuai. Karena /products
	// sekarang milik Inventory, dari SO arahkan ke daftar baca-saja.
	const fromOpname =
		typeof window !== 'undefined' && document.referrer.includes('/opname/products');
	$: productsListHref = fromOpname ? '/opname/products' : '/products';

	let product: any = null;
	let trend: any[] = [];
	let isLoading = true;
	let range: '7' | '30' | '90' | 'all' = '30';
	let firstLoad = true;

	async function loadProduct() {
		try {
			const res = await apiClient.get(`${API_ENDPOINTS.products}/${productId}`);
			if (res.success && res.data) product = res.data;
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat produk');
		}
	}

	async function loadTrend() {
		try {
			let query = '?limit=200';
			if (range !== 'all') {
				const days = parseInt(range);
				const from = new Date();
				from.setDate(from.getDate() - days);
				query += `&from=${localDateTime(from)}`;
			}
			const res = await apiClient.get<{ trend: any[] }>(
				`${API_ENDPOINTS.products}/${productId}/stock-history${query}`
			);
			if (res.success && res.data) {
				trend = res.data.trend || [];
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat riwayat');
		}
	}

	// Trigger reload ketika range berubah (hanya setelah load pertama)
	$: if (!firstLoad && range) loadTrend();

	onMount(async () => {
		await loadProduct();
		await loadTrend();
		isLoading = false;
		firstLoad = false;
	});
</script>

<svelte:head>
	<title>{product?.name || 'Produk'} - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<!-- Breadcrumb -->
		<div class="breadcrumbs mb-6 text-sm">
			<ul>
				<li><a href="/">Dashboard</a></li>
				<li><a href={productsListHref}>Products</a></li>
				<li>{product?.name || 'Detail'}</li>
			</ul>
		</div>

		{#if isLoading}
			<div class="flex h-64 items-center justify-center">
				<Loader size={48} />
			</div>
		{:else if product}
			<!-- Product Info Card -->
			<div class="mb-6 rounded-2xl border border-base-300 bg-base-200 p-6">
				<div class="flex flex-wrap items-start justify-between gap-4">
					<div>
						<h1 class="text-2xl font-bold">{product.name}</h1>
						<p class="metric mt-1 text-sm text-base-content/60">{product.product_code}</p>
						{#if product.brand}
							<p class="mt-1 text-sm text-base-content/60">Brand: {product.brand}</p>
						{/if}
					</div>
					<div class="text-right">
						<div class="text-xs text-base-content/60">Stok Awal</div>
						<div class="metric text-3xl font-bold text-primary">{product.stock?.quantity_system ?? '-'}</div>
					</div>
				</div>
			</div>

			<!-- Chart Card -->
			<div class="rounded-2xl border border-base-300 bg-base-200 p-4 md:p-6">
				<div class="mb-4 flex flex-wrap items-center justify-between gap-3">
					<h2 class="text-lg font-semibold">Riwayat Stok Opname</h2>
					<div class="join">
						<button
							class="btn btn-sm join-item"
							class:btn-primary={range === '7'}
							on:click={() => (range = '7')}
						>7 Hari</button>
						<button
							class="btn btn-sm join-item"
							class:btn-primary={range === '30'}
							on:click={() => (range = '30')}
						>30 Hari</button>
						<button
							class="btn btn-sm join-item"
							class:btn-primary={range === '90'}
							on:click={() => (range = '90')}
						>90 Hari</button>
						<button
							class="btn btn-sm join-item"
							class:btn-primary={range === 'all'}
							on:click={() => (range = 'all')}
						>Semua</button>
					</div>
				</div>

				<StockTrendChart data={trend} height={380} />
			</div>
		{:else}
			<div class="alert alert-error">Produk tidak ditemukan</div>
		{/if}
	</div>
</LayoutWrapper>
