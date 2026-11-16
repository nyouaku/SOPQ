<script lang="ts">
	import { onMount } from 'svelte';
	import { writable, get } from 'svelte/store';
	import { page as pageStore } from '$app/stores';
	import { replaceState } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import TableSkeleton from '$lib/components/skeletons/TableSkeleton.svelte';
	import { toastStore } from '$lib/stores/toast';
	import { stockOpnameReportQuery, warehousesQuery, suppliersQuery } from '$lib/api/queries';
	import { API_ENDPOINTS } from '$lib/api/config';

	const limit = 25;
		const init = new URLSearchParams($pageStore.url.search);

	// Nilai input filter (baru berlaku setelah tombol terapkan ditekan).
	let warehouseId = init.get('warehouse_id') ?? '';
	let brand = init.get('brand') ?? '';
	let supplierId = init.get('supplier_id') ?? '';
	let dateFrom = init.get('date_from') ?? '';
	let dateTo = init.get('date_to') ?? '';
	let minDiff = init.get('min_diff') ?? '';
	let maxDiff = init.get('max_diff') ?? '';
	// session_id: diisi saat dibuka dari dialog "Sesi Selesai" agar report
	// langsung menampilkan hasil sesi tersebut. Disembunyikan dari filter UI.
	let sessionId = init.get('session_id') ?? '';

	// Params yang sudah diterapkan — store, supaya perubahan memicu query baru.
	let page = Math.max(1, parseInt(init.get('page') ?? '1', 10) || 1);
		const applied = writable(buildParams(page));

		// Sinkron filter ke URL: tautan bisa dibagikan, refresh tidak mereset filter.
		onMount(() =>
			applied.subscribe((p) => replaceState(`${window.location.pathname}?${p.toString()}`, {}))
		);
	
	const report = stockOpnameReportQuery(applied);
	const warehousesQ = warehousesQuery();
	const suppliersQ = suppliersQuery();

	$: warehouses = ($warehousesQ.data as any[]) || [];
	$: suppliers = ($suppliersQ.data as any[]) || [];

	$: rows = ($report.data?.data as any[]) || [];
	$: total = $report.data?.total || 0;
	$: totalPages = Math.max(1, Math.ceil(total / limit));

	// Ringkasan dihitung dari baris yang tampil — data asli, bukan angka karangan.
	$: totalSystem = summary.total_system;
	$: totalPhysical = summary.total_physical;
	$: totalDiff = summary.total_difference;
	$: mismatchCount = summary.mismatch_count;
		$: summary = $report.data?.summary || { total_system: 0, total_physical: 0, total_difference: 0, mismatch_count: 0 };

	$: activeFilters = [...$applied.keys()].some((k) => k !== 'page' && k !== 'limit');

	function buildParams(nextPage: number): URLSearchParams {
		const p = new URLSearchParams();
		if (sessionId) p.set('session_id', sessionId);
		if (warehouseId) p.set('warehouse_id', warehouseId);
			if (brand) p.set('brand', brand);
		if (supplierId) p.set('supplier_id', supplierId);
		if (dateFrom) p.set('date_from', dateFrom);
		if (dateTo) p.set('date_to', dateTo);
		if (minDiff) p.set('min_diff', minDiff);
		if (maxDiff) p.set('max_diff', maxDiff);
		p.set('page', String(nextPage));
		p.set('limit', String(limit));
		return p;
	}

	function apply() {
		page = 1;
		applied.set(buildParams(1));
	}

	function goPage(next: number) {
		page = next;
		const p = new URLSearchParams(get(applied));
		p.set('page', String(next));
		applied.set(p);
	}

	function clearFilters() {
		warehouseId = '';
			brand = '';
		supplierId = '';
		dateFrom = '';
		dateTo = '';
		minDiff = '';
		maxDiff = '';
		page = 1;
		applied.set(new URLSearchParams(`page=1&limit=${limit}`));
	}

	function download(blob: Blob, filename: string) {
		const url = window.URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = filename;
		document.body.appendChild(a);
		a.click();
		window.URL.revokeObjectURL(url);
		document.body.removeChild(a);
	}

	async function exportFile(endpoint: string, ext: string) {
		try {
			const token = localStorage.getItem('access_token');
			const res = await fetch(`${API_ENDPOINTS.base}${endpoint}?${get(applied).toString()}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!res.ok) throw new Error('Gagal export');
			const blob = await res.blob();
			download(blob, `stock_opname_${new Date().toISOString().split('T')[0]}.${ext}`);
			toastStore.success('File berhasil diunduh');
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal export data');
		}
	}

	let isExportingCsv = false;
	let isExportingExcel = false;

	async function handleExportCsv() {
		isExportingCsv = true;
		await exportFile(API_ENDPOINTS.reportsStockOpnameExport, 'csv');
		isExportingCsv = false;
	}

	async function handleExportExcel() {
		isExportingExcel = true;
		await exportFile('/export/stock-opname', 'xlsx');
		isExportingExcel = false;
	}
</script>

<svelte:head>
	<title>Reports - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6">
			<h1 class="page-title">Laporan Stock Opname</h1>
			<p class="mt-1 text-sm text-base-content/60">Selisih stok sistem vs fisik per scan</p>
		</div>

		<!-- Filter -->
		<section class="mb-4 rounded-2xl border border-base-300 bg-base-200">
			<header class="border-b border-base-300 px-4 py-3 md:px-5">
				<h2 class="font-semibold">Filter</h2>
			</header>
			<div class="p-4 md:p-5">
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
					<div>
						<label class="label py-0" for="f-warehouse">
							<span class="label-text text-xs text-base-content/60">Gudang</span>
						</label>
						<select id="f-warehouse" class="select select-bordered select-sm mt-1 w-full" bind:value={warehouseId}>
							<option value="">Semua</option>
							{#each warehouses as w}<option value={w.id}>{w.name}</option>{/each}
						</select>
					</div>

					<div>
						<label class="label py-0" for="f-supplier">
							<span class="label-text text-xs text-base-content/60">Supplier</span>
						</label>
						<select id="f-supplier" class="select select-bordered select-sm mt-1 w-full" bind:value={supplierId}>
							<option value="">Semua</option>
							{#each suppliers as s}<option value={s.id}>{s.name}</option>{/each}
						</select>
					</div>

					<div>
						<label class="label py-0" for="f-brand">
							<span class="label-text text-xs text-base-content/60">Brand</span>
						</label>
						<input
							id="f-brand"
							class="input input-bordered input-sm mt-1 w-full"
							placeholder="Semua brand"
							bind:value={brand}
							on:keypress={(e) => e.key === 'Enter' && apply()}
						/>
					</div>

					<div>
						<label class="label py-0" for="f-from">
							<span class="label-text text-xs text-base-content/60">Dari tanggal</span>
						</label>
						<input id="f-from" type="date" class="input input-bordered input-sm mt-1 w-full" bind:value={dateFrom} />
					</div>

					<div>
						<label class="label py-0" for="f-to">
							<span class="label-text text-xs text-base-content/60">Sampai tanggal</span>
						</label>
						<input id="f-to" type="date" class="input input-bordered input-sm mt-1 w-full" bind:value={dateTo} />
					</div>

					<div>
						<label class="label py-0" for="f-min">
							<span class="label-text text-xs text-base-content/60">Selisih min</span>
						</label>
						<input id="f-min" type="number" class="input input-bordered input-sm mt-1 w-full" placeholder="-" bind:value={minDiff} />
					</div>

					<div>
						<label class="label py-0" for="f-max">
							<span class="label-text text-xs text-base-content/60">Selisih max</span>
						</label>
						<input id="f-max" type="number" class="input input-bordered input-sm mt-1 w-full" placeholder="-" bind:value={maxDiff} />
					</div>
				</div>

				<div class="mt-4 flex flex-wrap items-center gap-2">
					<button
						class="btn btn-primary btn-sm btn-square"
						title="Terapkan filter"
						aria-label="Terapkan filter"
						on:click={apply}
					>
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="11" cy="11" r="7" />
							<path d="M16.5 16.5L21 21" stroke-linecap="round" />
						</svg>
					</button>
					<button class="btn btn-ghost btn-sm" on:click={clearFilters}>Reset</button>

					<span class="flex-1"></span>

					<button
						class="btn btn-success btn-sm btn-square"
						title="Export Excel"
						aria-label="Export Excel"
						disabled={isExportingExcel}
						on:click={handleExportExcel}
					>
						{#if isExportingExcel}
							<Loader size={18} />
						{:else}
							<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M12 21V9m0 0l-4 4m4-4l4 4M4 5h16" stroke-linecap="round" stroke-linejoin="round" />
							</svg>
						{/if}
					</button>
					<button
						class="btn btn-outline btn-sm btn-square"
						title="Export CSV"
						aria-label="Export CSV"
						disabled={isExportingCsv}
						on:click={handleExportCsv}
					>
						{#if isExportingCsv}
							<Loader size={18} />
						{:else}
							<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z" stroke-linejoin="round" />
								<path d="M14 3v5h5" stroke-linejoin="round" />
							</svg>
						{/if}
					</button>
				</div>
			</div>
		</section>

		<!-- Ringkasan -->
		{#if rows.length > 0}
			<section class="mb-4 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-base-300 bg-base-300 sm:grid-cols-4">
				<div class="bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Baris</div>
					<div class="metric mt-1 text-xl font-semibold">{total}</div>
				</div>
				<div class="bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Stok sistem</div>
					<div class="metric mt-1 text-xl font-semibold">{totalSystem}</div>
				</div>
				<div class="bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Stok fisik</div>
					<div class="metric mt-1 text-xl font-semibold">{totalPhysical}</div>
				</div>
				<div class="bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Selisih (baris)</div>
					<div
						class="metric mt-1 text-xl font-semibold"
						class:text-error={totalDiff !== 0}
						class:text-success={totalDiff === 0}
					>
						{totalDiff > 0 ? '+' : ''}{totalDiff}
						<span class="ml-1 text-xs font-normal text-base-content/50">/ {mismatchCount} baris</span>
					</div>
				</div>
			</section>
		{/if}

		<!-- Hasil -->
		<section class="overflow-hidden rounded-2xl border border-base-300 bg-base-200">
			{#if $report.isLoading}
				<TableSkeleton rows={6} columns={6} />
			{:else if $report.isError}
				<div class="p-6 text-center">
					<p class="text-sm text-error">{$report.error?.message || 'Gagal memuat laporan'}</p>
					<button class="btn btn-ghost btn-sm mt-2" on:click={() => $report.refetch()}>Coba lagi</button>
				</div>
			{:else if rows.length === 0}
				<div class="px-4 py-16 text-center">
					<p class="text-sm text-base-content/60">Belum ada data laporan</p>
					<p class="mt-1 text-xs text-base-content/40">
						{activeFilters
							? 'Tidak ada baris yang cocok dengan filter ini.'
							: 'Data muncul setelah ada scan pada sesi stock opname.'}
					</p>
					{#if activeFilters}
						<button class="btn btn-ghost btn-sm mt-3" on:click={clearFilters}>Reset filter</button>
					{/if}
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table table-sm">
						<thead>
							<tr>
								<th>Sesi</th>
								<th>Kode</th>
								<th>Produk</th>
								<th class="text-right">Sistem</th>
								<th class="text-right">Fisik</th>
								<th class="text-right">Selisih</th>
								<th>Metode</th>
								<th>Waktu</th>
							</tr>
						</thead>
						<tbody>
							{#each rows as row}
								<tr>
									<td class="metric text-xs">{row.session_code || '-'}</td>
									<td class="metric text-xs">{row.product_code}</td>
									<td>
										<div class="truncate">{row.product_name}</div>
										<div class="text-xs text-base-content/50">
											{row.brand || '-'}
										</div>
									</td>
									<td class="metric text-right">{row.stock_system}</td>
									<td class="metric text-right">{row.stock_physical}</td>
									<td
										class="metric text-right font-medium"
										class:text-error={row.difference < 0}
										class:text-success={row.difference > 0}
									>
										{row.difference > 0 ? '+' : ''}{row.difference}
									</td>
									<td>
										<span class="badge badge-ghost badge-sm">{row.input_method || '-'}</span>
									</td>
									<td class="text-xs text-base-content/60">
										{row.scanned_at ? new Date(row.scanned_at).toLocaleString('id-ID') : '-'}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</section>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="mt-4 flex items-center justify-center gap-2">
				<button class="btn btn-sm" disabled={page <= 1} on:click={() => goPage(page - 1)}>Sebelumnya</button>
				<span class="metric text-xs text-base-content/60">{page} / {totalPages}</span>
				<button class="btn btn-sm" disabled={page >= totalPages} on:click={() => goPage(page + 1)}>Berikutnya</button>
			</div>
		{/if}
	</div>
</LayoutWrapper>
