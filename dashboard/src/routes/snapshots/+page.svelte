<script lang="ts">
	import { onMount } from 'svelte';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let days: '7' | '30' | '90' = '30';
	let warehouseId = '';
	let warehouses: any[] = [];
	let comparisons: any[] = [];
	let summary = { total_products: 0, total_increased: 0, total_decreased: 0, total_unchanged: 0 };
	let isLoading = true;
	let filterStatus: 'all' | 'increased' | 'decreased' | 'critical' = 'all';

	onMount(async () => {
		try {
			const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
			if (res.success && res.data) warehouses = res.data.warehouses || [];
		} catch {}
		await loadComparison();
		isLoading = false;
	});

	async function loadComparison() {
		try {
			let q = `?days=${days}`;
			if (warehouseId) q += `&warehouse_id=${warehouseId}`;
			const res = await apiClient.get<{
				comparisons: any[];
				total_products: number;
				total_increased: number;
				total_decreased: number;
				total_unchanged: number;
			}>(`/snapshots/compare${q}`);
			if (res.success && res.data) {
				comparisons = res.data.comparisons || [];
				summary = {
					total_products: res.data.total_products,
					total_increased: res.data.total_increased,
					total_decreased: res.data.total_decreased,
					total_unchanged: res.data.total_unchanged
				};
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat perbandingan');
		}
	}

	$: filtered = filterStatus === 'all'
		? comparisons
		: comparisons.filter((c) => c.status === filterStatus);

	function statusBadge(status: string): string {
		switch (status) {
			case 'increased': return 'badge-success';
			case 'low': return 'badge-warning';
			case 'critical': return 'badge-error';
			default: return 'badge-ghost';
		}
	}

	function statusLabel(status: string): string {
		switch (status) {
			case 'increased': return 'Naik';
			case 'low': return 'Rendah';
			case 'critical': return 'Kritis';
			case 'decreased': return 'Turun';
			default: return 'Stabil';
		}
	}
</script>

<svelte:head>
	<title>Perbandingan Stok - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6">
			<h1 class="page-title">Perbandingan Stok Snapshot</h1>
			<p class="mt-1 text-sm text-base-content/60">
				Bandingkan stok saat ini dengan snapshot periode sebelumnya
			</p>
		</div>

		<!-- Filter -->
		<div class="mb-6 rounded-2xl border border-base-300 bg-base-200 p-4">
			<div class="flex flex-wrap items-end gap-3">
				<div>
					<label class="label py-0" for="days">
						<span class="label-text text-xs">Periode</span>
					</label>
					<select id="days" class="select select-bordered select-sm" bind:value={days} on:change={loadComparison}>
						<option value="7">7 Hari Lalu</option>
						<option value="30">30 Hari Lalu</option>
						<option value="90">90 Hari Lalu</option>
					</select>
				</div>
				<div>
					<label class="label py-0" for="wh">
						<span class="label-text text-xs">Gudang</span>
					</label>
					<select id="wh" class="select select-bordered select-sm" bind:value={warehouseId} on:change={loadComparison}>
						<option value="">Semua Gudang</option>
						{#each warehouses as wh}
							<option value={wh.id}>{wh.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="label py-0" for="status">
						<span class="label-text text-xs">Status</span>
					</label>
					<select id="status" class="select select-bordered select-sm" bind:value={filterStatus}>
						<option value="all">Semua</option>
						<option value="increased">Naik</option>
						<option value="decreased">Turun</option>
						<option value="critical">Kritis</option>
					</select>
				</div>
			</div>
		</div>

		{#if isLoading}
			<div class="flex h-64 items-center justify-center">
				<Loader size={48} />
			</div>
		{:else}
			<!-- Summary Cards -->
			<div class="mb-6 grid grid-cols-2 gap-4 md:grid-cols-4">
				<div class="rounded-xl border border-base-300 bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Total Produk</div>
					<div class="metric text-2xl font-bold">{summary.total_products}</div>
				</div>
				<div class="rounded-xl border border-success/40 bg-success/10 p-4">
					<div class="text-xs text-success">Naik</div>
					<div class="metric text-2xl font-bold text-success">{summary.total_increased}</div>
				</div>
				<div class="rounded-xl border border-error/40 bg-error/10 p-4">
					<div class="text-xs text-error">Turun</div>
					<div class="metric text-2xl font-bold text-error">{summary.total_decreased}</div>
				</div>
				<div class="rounded-xl border border-base-300 bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Stabil</div>
					<div class="metric text-2xl font-bold">{summary.total_unchanged}</div>
				</div>
			</div>

			<!-- Table -->
			<div class="rounded-2xl border border-base-300 bg-base-200">
				<div class="border-b border-base-300 p-4">
					<h2 class="font-semibold">Detail Perubahan ({filtered.length})</h2>
				</div>
				{#if filtered.length === 0}
					<p class="p-8 text-center text-sm text-base-content/50">Tidak ada data</p>
				{:else}
					<div class="overflow-x-auto">
						<table class="table table-sm">
							<thead>
								<tr>
									<th>Kode</th>
									<th>Nama Produk</th>
									<th>Gudang</th>
									<th class="text-right">Snapshot</th>
									<th class="text-right">Sekarang</th>
									<th class="text-right">Perubahan</th>
									<th>Status</th>
								</tr>
							</thead>
							<tbody>
								{#each filtered as item}
									<tr>
										<td class="metric">{item.product_code}</td>
										<td class="max-w-[200px] truncate">{item.product_name}</td>
										<td class="text-xs">{item.warehouse_name}</td>
										<td class="metric text-right">{item.snapshot_stock}</td>
										<td class="metric text-right font-semibold">{item.current_stock}</td>
										<td class="metric text-right">
											<span class="{item.change > 0 ? 'text-success' : item.change < 0 ? 'text-error' : ''}">
												{item.change > 0 ? '+' : ''}{item.change}
											</span>
											{#if item.snapshot_stock > 0}
												<span class="text-xs text-base-content/50">({item.change_percent > 0 ? '+' : ''}{item.change_percent.toFixed(0)}%)</span>
											{/if}
										</td>
										<td>
											<span class="badge badge-sm {statusBadge(item.status)}">
												{statusLabel(item.status)}
											</span>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</LayoutWrapper>
