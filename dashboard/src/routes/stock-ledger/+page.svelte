<script lang="ts">
	import { onMount } from 'svelte';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import TableSkeleton from '$lib/components/skeletons/TableSkeleton.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let warehouses: any[] = [];
	let selectedWarehouse = '';
	let productId = '';
	let entries: any[] = [];
	let total = 0;
	let page = 1;
	const limit = 50;
	let totalPages = 1;
	let isLoading = true;
	let filterMode: 'product' | 'warehouse' | 'range' = 'warehouse';
	let fromDate = '';
	let toDate = '';

	onMount(async () => {
		try {
			const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
			if (res.success && res.data) warehouses = res.data.warehouses || [];
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat gudang');
		} finally {
			isLoading = false;
		}
	});

	async function loadLedger() {
		if (filterMode === 'product' && (!productId || !selectedWarehouse)) {
			toastStore.error('Masukkan Product ID dan pilih Gudang');
			return;
		}
		isLoading = true;
		try {
			let url = '';
			if (filterMode === 'product') {
				url = API_ENDPOINTS.stockLedgerProduct(productId, selectedWarehouse);
			} else if (filterMode === 'warehouse') {
				url = API_ENDPOINTS.stockLedgerWarehouse(selectedWarehouse);
			} else {
				const from = fromDate || new Date(Date.now() - 30 * 864e5).toISOString();
				const to = toDate || new Date().toISOString();
				url = `${API_ENDPOINTS.stockLedgerRange}?from=${from}&to=${to}`;
				if (selectedWarehouse) url += `&warehouse_id=${selectedWarehouse}`;
			}
			url += `&page=${page}&limit=${limit}`;
			const res = await apiClient.get<{ ledger: any[]; total: number }>(url);
			if (res.success && res.data) {
				entries = res.data.ledger || [];
				total = res.data.total || 0;
				totalPages = Math.max(1, Math.ceil(total / limit));
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat ledger');
		} finally {
			isLoading = false;
		}
	}

	function changePage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		loadLedger();
	}

	function formatDate(iso: string): string {
		if (!iso) return '-';
		return new Date(iso).toLocaleString('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getTxLabel(type: string): string {
		const map: Record<string, string> = {
			initial: 'Saldo Awal',
			purchase_in: 'Barang Masuk',
			sale_out: 'Barang Keluar',
			transfer_in: 'Transfer Masuk',
			transfer_out: 'Transfer Keluar',
			adjustment: 'Penyesuaian',
			opname_correction: 'Koreksi Opname',
			return_in: 'Retur Masuk',
			return_out: 'Retur Keluar'
		};
		return map[type] || type;
	}
</script>

<svelte:head>
	<title>Stock Ledger - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6">
			<h1 class="page-title">Stock Ledger (Kartu Stok)</h1>
			<p class="mt-1 text-sm text-base-content/60">
				Riwayat lengkap pergerakan stok dengan saldo berjalan
			</p>
		</div>

		<!-- Filter -->
		<div class="mb-6 rounded-2xl border border-base-300 bg-base-200 p-4">
			<div class="flex flex-wrap items-end gap-3">
				<div>
					<label class="label py-0" for="filter-mode">
						<span class="label-text text-xs">Mode</span>
					</label>
					<select id="filter-mode" class="select select-bordered select-sm" bind:value={filterMode}>
						<option value="warehouse">Per Gudang</option>
						<option value="product">Per Produk</option>
						<option value="range">Antar Tanggal</option>
					</select>
				</div>

				{#if filterMode === 'warehouse' || filterMode === 'product'}
					<div>
						<label class="label py-0" for="warehouse">
							<span class="label-text text-xs">Gudang</span>
						</label>
						<select id="warehouse" class="select select-bordered select-sm" bind:value={selectedWarehouse}>
							<option value="">Pilih gudang...</option>
							{#each warehouses as wh}
								<option value={wh.id}>{wh.name}</option>
							{/each}
						</select>
					</div>
				{/if}

				{#if filterMode === 'product'}
					<div>
						<label class="label py-0" for="product-id">
							<span class="label-text text-xs">Product ID</span>
						</label>
						<input
							id="product-id"
							type="text"
							class="input input-bordered input-sm metric"
							bind:value={productId}
							placeholder="UUID produk"
						/>
					</div>
				{/if}

				{#if filterMode === 'range'}
					<div>
						<label class="label py-0" for="from-date">
							<span class="label-text text-xs">Dari</span>
						</label>
						<input id="from-date" type="date" class="input input-bordered input-sm" bind:value={fromDate} />
					</div>
					<div>
						<label class="label py-0" for="to-date">
							<span class="label-text text-xs">Sampai</span>
						</label>
						<input id="to-date" type="date" class="input input-bordered input-sm" bind:value={toDate} />
					</div>
				{/if}

				<button class="btn btn-primary btn-sm" on:click={loadLedger} disabled={isLoading}>
					{#if isLoading}
						<span class="loading loading-spinner loading-sm"></span>
					{:else}
						Muat Data
					{/if}
				</button>
			</div>
		</div>

		<!-- Ledger Table -->
		{#if isLoading}
			<TableSkeleton rows={8} columns={7} />
		{:else if entries.length === 0}
			<div class="rounded-2xl border border-base-300 bg-base-200 p-8 text-center text-sm text-base-content/50">
				Tidak ada data ledger
			</div>
		{:else}
			<div class="rounded-2xl border border-base-300 bg-base-200">
				<div class="overflow-x-auto">
					<table class="table table-sm">
						<thead>
							<tr>
								<th>Waktu</th>
								<th>Produk</th>
								<th>Jenis Transaksi</th>
								<th>Ref</th>
								<th class="text-right">Masuk</th>
								<th class="text-right">Keluar</th>
								<th class="text-right">Saldo</th>
								<th>Catatan</th>
							</tr>
						</thead>
						<tbody>
							{#each entries as entry}
								<tr>
									<td class="text-xs">{formatDate(entry.created_at)}</td>
									<td class="text-xs">{entry.product?.name || '-'}</td>
									<td>
										<span class="badge badge-ghost badge-sm">{getTxLabel(entry.transaction_type)}</span>
									</td>
									<td class="metric text-xs">{entry.reference_code || '-'}</td>
									<td class="metric text-right text-success">
										{entry.quantity_in > 0 ? `+${entry.quantity_in}` : '-'}
									</td>
									<td class="metric text-right text-error">
										{entry.quantity_out > 0 ? `-${entry.quantity_out}` : '-'}
									</td>
									<td class="metric text-right font-semibold">{entry.balance_after}</td>
									<td class="text-xs text-base-content/60">{entry.note || '-'}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				{#if totalPages > 1}
					<div class="flex justify-center p-4">
						<div class="join">
							<button class="join-item btn btn-sm" disabled={page === 1} on:click={() => changePage(page - 1)}>
								«
							</button>
							<button class="join-item btn btn-sm">Page {page} of {totalPages}</button>
							<button class="join-item btn btn-sm" disabled={page === totalPages} on:click={() => changePage(page + 1)}>
								»
							</button>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</LayoutWrapper>
