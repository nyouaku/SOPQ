<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	// Mode pergerakan & gudang & bulan.
	let mode: 'out' | 'in' = 'out';
	let warehouseId = '';
	let month = new Date().toISOString().slice(0, 7); // YYYY-MM
	let warehouses: any[] = [];

	type Row = { product_id: string; name: string; product_code: string };
	let rows: Row[] = [];
	let days: number[] = [];
	let daysInMonth = 31;

	// Isian grid: key = `${rowIdx}:${day}` -> {in, out}
	let cells: Record<string, { in: number; out: number }> = {};
	let unitCosts: Record<string, number> = {}; // key rowIdx -> harga satuan terakhir
	let isLoading = true;
	let isSaving = false;
	let searchIdx = -1;
	let searchQuery = '';
	let searchResults: any[] = [];

	$: daysInMonth = new Date(
		parseInt(month.slice(0, 4)),
		parseInt(month.slice(5, 7)),
		0
	).getDate();
	$: days = Array.from({ length: daysInMonth }, (_, i) => i + 1);

	async function loadWarehouses() {
		const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
		warehouses = res.success && res.data ? res.data.warehouses || [] : [];
		if (warehouses.length > 0) warehouseId = warehouses[0].id;
	}

	async function loadGrid() {
		if (!warehouseId) return;
		isLoading = true;
		try {
			const res = await apiClient.get<{ cells: any[] }>(
				`${API_ENDPOINTS.inventoryGrid}?warehouse_id=${warehouseId}&month=${month}&movement_type=${mode}`
			);
			cells = {};
			rows = [];
			if (res.success && res.data?.cells) {
				for (const c of res.data.cells) {
					let idx = rows.findIndex((r) => r.product_id === c.product_id);
					if (idx === -1) {
						rows.push({ product_id: c.product_id, name: c.name, product_code: c.product_code });
						idx = rows.length - 1;
					}
					const k = `${idx}:${c.day}`;
					if (!cells[k]) cells[k] = { in: 0, out: 0 };
					if (mode === 'in') cells[k].in = c.in_qty;
					else cells[k].out = c.out_qty;
				}
			}
			if (rows.length === 0) addRow();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal memuat grid');
		} finally {
			isLoading = false;
		}
	}

	function addRow() {
		rows = [...rows, { product_id: '', name: '', product_code: '' }];
	}

	function removeRow(i: number) {
		rows = rows.filter((_, idx) => idx !== i);
	}

	function switchMode(m: 'out' | 'in') {
		mode = m;
		loadGrid();
	}

	function cellKey(i: number, d: number) {
		return `${i}:${d}`;
	}

	function onInput(i: number, d: number, kind: 'in' | 'out', val: string) {
		const n = parseInt(val) || 0;
		const k = cellKey(i, d);
		if (!cells[k]) cells[k] = { in: 0, out: 0 };
		cells[k][kind] = n;
	}

	async function save() {
		// Kumpulkan sel terisi -> entries[].
		const entries: any[] = [];
		for (let i = 0; i < rows.length; i++) {
			const row = rows[i];
			if (!row.product_id) continue;
			for (const d of days) {
				const c = cells[cellKey(i, d)];
				if (!c) continue;
				const qty = mode === 'in' ? c.in : c.out;
				if (qty > 0) {
					entries.push({
						product_id: row.product_id,
						date: `${month}-${String(d).padStart(2, '0')}`,
						quantity: qty,
						unit_cost: unitCosts[i] || 0
					});
				}
			}
		}
		if (entries.length === 0) {
			toastStore.error('Tidak ada isian untuk disimpan');
			return;
		}

		isSaving = true;
		try {
			const res = await apiClient.post<{ movements_created: number; items_posted: number }>(
				API_ENDPOINTS.inventoryGrid,
				{
					warehouse_id: warehouseId,
					movement_type: mode,
					sub_type: mode === 'in' ? 'purchase' : 'sale',
					entries
				}
			);
			if (!res.success) throw new Error(res.error?.message || 'Gagal menyimpan');
			toastStore.success(
				`${res.data.items_posted} isian tersimpan (${res.data.movements_created} dokumen)`
			);
			await loadGrid();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal menyimpan');
		} finally {
			isSaving = false;
		}
	}

	// Autocomplete produk per baris.
	let searchTimer: ReturnType<typeof setTimeout> | null = null;
	let activeRow = -1;
	function onNameInput(i: number) {
		activeRow = i;
		if (searchTimer) clearTimeout(searchTimer);
		const q = rows[i]?.name ?? '';
		if (!q || q.length < 2) {
			searchResults = [];
			return;
		}
		searchTimer = setTimeout(() => doSearch(i, q), 300);
	}
	async function doSearch(i: number, q: string) {
		try {
			const res = await apiClient.get<{ products: any[] }>(
				`${API_ENDPOINTS.products}?limit=10&search=${encodeURIComponent(q)}`
			);
			searchResults = res.success && res.data ? res.data.products || [] : [];
		} catch {
			searchResults = [];
		}
	}
	function pickProduct(i: number, p: any) {
		rows[i].product_id = p.id;
		rows[i].name = p.name;
		rows[i].product_code = p.product_code;
		searchResults = [];
		activeRow = -1;
	}

	async function init() {
		await loadWarehouses();
		await loadGrid();
	}
	init();
</script>

<svelte:head>
	<title>Grid Harian - Inventory</title>
</svelte:head>

<LayoutWrapper>
	<div class="page pb-24 md:pb-28">
		<div class="flex items-center gap-3 mb-4">
			<button class="btn btn-ghost btn-sm" on:click={() => goto('/inventory')}>←</button>
			<h1 class="page-title">Isi Grid Harian</h1>
		</div>
		<p class="page-subtitle mb-4">
			Baris = produk, kolom = tanggal. Klik sel lalu isi angka {mode === 'out' ? 'keluar' : 'masuk'} hari itu.
		</p>

		<!-- Toolbar -->
		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-3">
			<div class="flex flex-wrap items-center gap-2">
				<div class="grid grid-cols-2 gap-1 rounded-xl bg-base-300 p-1" role="tablist">
					<button role="tab" aria-selected={mode === 'out'}
						class="rounded-lg px-3 py-1.5 text-sm font-medium {mode === 'out' ? 'bg-base-100 text-error shadow-sm' : 'opacity-50'}"
						on:click={() => switchMode('out')}>⬆️ Keluar</button>
					<button role="tab" aria-selected={mode === 'in'}
						class="rounded-lg px-3 py-1.5 text-sm font-medium {mode === 'in' ? 'bg-base-100 text-success shadow-sm' : 'opacity-50'}"
						on:click={() => switchMode('in')}>⬇️ Masuk</button>
				</div>
				<select class="select select-bordered select-sm" bind:value={warehouseId} on:change={loadGrid}>
					{#each warehouses as wh}
						<option value={wh.id}>{wh.name}</option>
					{/each}
				</select>
				<input type="month" class="input input-bordered input-sm" bind:value={month} on:change={loadGrid} />
				<button class="btn btn-ghost btn-sm ml-auto" on:click={addRow}>+ Tambah produk</button>
			</div>
		</div>

		{#if isLoading}
			<div class="flex h-64 items-center justify-center"><Loader size={48} /></div>
		{:else}
			<!-- Baris produk + grid -->
			<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
				<div class="overflow-x-auto">
					<table class="table table-sm">
						<thead>
							<tr>
								<th class="sticky left-0 z-10 bg-base-200 min-w-[14rem]">Produk</th>
								{#each days as d}
									<th class="text-center text-xs" title="Tanggal {d}">{d}</th>
								{/each}
								<th></th>
							</tr>
						</thead>
						<tbody>
							{#each rows as row, i}
								<tr>
									<td class="sticky left-0 z-10 bg-base-200">
										<div class="relative">
											<input
												class="input input-bordered input-xs w-full"
												placeholder="Cari produk..."
												bind:value={row.name}
												on:input={() => onNameInput(i)}
											/>
											{#if activeRow === i && searchResults.length > 0}
												<div class="absolute z-30 mt-1 max-h-48 w-full overflow-y-auto rounded-xl border border-base-300 bg-base-100 shadow-lg">
													{#each searchResults as p}
														<button class="block w-full px-2 py-1.5 text-left text-xs hover:bg-base-200"
															on:click={() => pickProduct(i, p)}>
															<div class="truncate">{p.name}</div>
															<div class="metric text-[10px] text-base-content/60">{p.product_code}</div>
														</button>
													{/each}
												</div>
											{/if}
										</div>
									</td>
									{#each days as d}
										<td class="p-0">
											<input
												type="number"
												min="0"
												class="w-14 border-0 bg-transparent text-center text-sm focus:bg-primary/10 focus:outline-none"
												value={cells[cellKey(i, d)]?.[mode] ?? ''}
												on:input={(e) => onInput(i, d, mode, e.currentTarget.value)}
											/>
										</td>
									{/each}
									<td>
										{#if rows.length > 1}
											<button class="btn btn-ghost btn-xs text-error" on:click={() => removeRow(i)}>✕</button>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>

			<!-- Simpan -->
			<div class="mt-4 flex justify-end gap-2">
				<button class="btn btn-primary btn-md" on:click={save} disabled={isSaving}>
					{#if isSaving}
						<span class="loading loading-spinner loading-sm"></span>
						Menyimpan...
					{:else}
						Simpan Semua
					{/if}
				</button>
			</div>
		{/if}
	</div>
</LayoutWrapper>
