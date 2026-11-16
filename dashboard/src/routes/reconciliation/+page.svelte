<script lang="ts">
	import { onMount } from 'svelte';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let warehouses: any[] = [];
	let selected = new Set<string>();
	let results: any[] = [];
	let isLoadingWarehouses = true;
	let isReconciling = false;
	let hasRun = false;

	onMount(async () => {
		try {
			const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
			if (res.success && res.data) {
				warehouses = res.data.warehouses || [];
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat gudang');
		} finally {
			isLoadingWarehouses = false;
		}
	});

	function toggle(id: string) {
		if (selected.has(id)) selected.delete(id);
		else selected.add(id);
		selected = new Set(selected);
	}

	async function reconcile() {
		if (selected.size < 2) {
			toastStore.error('Pilih minimal 2 gudang');
			return;
		}
		isReconciling = true;
		try {
			const res = await apiClient.post<{ reconciliation: any[] }>('/reconciliation', {
				warehouse_ids: Array.from(selected)
			});
			if (res.success && res.data) {
				results = res.data.reconciliation || [];
				hasRun = true;
			} else {
				toastStore.error(res.error?.message || 'Gagal rekonsiliasi');
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal rekonsiliasi');
		} finally {
			isReconciling = false;
		}
	}

	// Produk yang hilang dari setidaknya satu gudang
	$: missingOnly = results.filter((r) => r.missing_from?.length > 0);
</script>

<svelte:head>
	<title>Rekonsiliasi Gudang - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6">
			<h1 class="page-title">Rekonsiliasi Antar Gudang</h1>
			<p class="mt-1 text-sm text-base-content/60">
				Bandingkan stok sistem produk di beberapa gudang sekaligus
			</p>
		</div>

		<!-- Pilih Gudang -->
		<div class="mb-6 rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
			<h2 class="mb-3 font-semibold">Pilih Gudang (minimal 2)</h2>
			{#if isLoadingWarehouses}
				<div class="flex h-20 items-center justify-center">
					<Loader size={32} />
				</div>
			{:else if warehouses.length === 0}
				<p class="text-sm text-base-content/60">Belum ada gudang terdaftar</p>
			{:else}
				<div class="flex flex-wrap gap-2">
					{#each warehouses as wh}
						<label class="flex cursor-pointer items-center gap-2 rounded-xl border border-base-300 bg-base-100 px-3 py-2 text-sm">
							<input
								type="checkbox"
								class="checkbox checkbox-primary checkbox-sm"
								checked={selected.has(wh.id)}
								on:change={() => toggle(wh.id)}
							/>
							{wh.name}
						</label>
					{/each}
				</div>
				<button
					class="btn btn-primary mt-4"
					on:click={reconcile}
					disabled={selected.size < 2 || isReconciling}
				>
					{#if isReconciling}
						<Loader size={20} />
						Memproses...
					{:else}
						Rekonsiliasi
					{/if}
				</button>
			{/if}
		</div>

		<!-- Hasil -->
		{#if hasRun}
			<div class="rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
				<div class="mb-3 flex items-center justify-between">
					<h2 class="font-semibold">
						Hasil ({results.length} produk)
						{#if missingOnly.length > 0}
							<span class="badge badge-warning badge-sm ml-2">{missingOnly.length} ada gudang yang tidak memuat</span>
						{/if}
					</h2>
				</div>

				{#if results.length === 0}
					<p class="text-sm text-base-content/60">Tidak ada produk di gudang-gudang terpilih</p>
				{:else}
					<div class="overflow-x-auto">
						<table class="table table-sm">
							<thead>
								<tr>
									<th>Kode</th>
									<th>Nama Produk</th>
									<th class="text-right">Total Sistem</th>
									<th class="text-center">Ada Di</th>
									<th>Hilang Dari</th>
								</tr>
							</thead>
							<tbody>
								{#each results as r}
									<tr class={r.missing_from?.length > 0 ? 'bg-warning/5' : ''}>
										<td class="metric">{r.product_code}</td>
										<td class="max-w-[220px] truncate">{r.product_name}</td>
										<td class="metric text-right font-semibold">{r.total_all}</td>
										<td class="text-center">
											<span
												class="badge badge-sm {r.present_in_count === r.total_warehouses
													? 'badge-success'
													: 'badge-warning'}"
											>
												{r.present_in_count}/{r.total_warehouses}
											</span>
										</td>
										<td>
											{#if r.missing_from?.length > 0}
												<span class="text-xs text-warning">{r.missing_from.join(', ')}</span>
											{:else}
												<span class="text-xs text-base-content/40">—</span>
											{/if}
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
