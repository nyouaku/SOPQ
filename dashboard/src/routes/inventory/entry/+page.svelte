<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { localMonth, previousMonth } from '$lib/utils/date';

	// Mode pergerakan & gudang & bulan.
	let mode: 'out' | 'in' = 'out';
	let warehouseId = '';
	let month = localMonth(); // YYYY-MM, waktu lokal
	let warehouses: any[] = [];

	type Row = { product_id: string; name: string; product_code: string };
	let rows: Row[] = [];
	let days: number[] = [];
	let daysInMonth = 31;

	// Isian grid: key = `${product_id}:${day}` -> {in, out}.
	// Sengaja di-key product_id, BUKAN index baris: menghapus/menambah baris tak boleh
	// memindahkan angka ke produk lain.
	let cells: Record<string, { in: number; out: number }> = {};
	// Baseline = nilai yang sudah tersimpan di server. Simpan hanya mengirim selisih,
	// sehingga klik "Simpan" berulang tidak menggandakan stok.
	let saved: Record<string, { in: number; out: number }> = {};
	let unitCosts: Record<string, number> = {}; // key product_id -> harga satuan
	let isLoading = true;
	let isSaving = false;
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
			saved = {};
			rows = [];
			if (res.success && res.data?.cells) {
				for (const c of res.data.cells) {
					let idx = rows.findIndex((r) => r.product_id === c.product_id);
					if (idx === -1) {
						rows.push({ product_id: c.product_id, name: c.name, product_code: c.product_code });
						idx = rows.length - 1;
					}
					const k = `${c.product_id}:${c.day}`;
					if (!cells[k]) cells[k] = { in: 0, out: 0 };
					if (mode === 'in') cells[k].in = c.in_qty;
					else cells[k].out = c.out_qty;
					saved[k] = { ...cells[k] };
					if (c.unit_cost) unitCosts[c.product_id] = c.unit_cost;
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

	function cellKey(productId: string, d: number) {
		return `${productId}:${d}`;
	}

	function onInput(productId: string, d: number, kind: 'in' | 'out', val: string) {
		const n = Math.max(0, parseInt(val) || 0);
		const k = cellKey(productId, d);
		if (!cells[k]) cells[k] = { in: 0, out: 0 };
		cells[k][kind] = n;
	}

	/** Selisih terhadap nilai tersimpan; positif = tambah, negatif = koreksi turun. */
	function deltaOf(productId: string, d: number): number {
		const k = cellKey(productId, d);
		const cur = cells[k]?.[mode] ?? 0;
		const base = saved[k]?.[mode] ?? 0;
		return cur - base;
	}

	/**
	 * Tempel blok dari Excel: Ctrl+V pada satu sel mengisi baris & hari berikutnya.
	 * Menghapus kebiasaan mengetik ulang puluhan angka satu per satu.
	 */
	function onPaste(e: ClipboardEvent, productId: string, startDay: number) {
		const text = e.clipboardData?.getData('text/plain');
		if (!text || !text.includes('\t')) return; // satu angka: biarkan paste biasa
		e.preventDefault();

		const startRow = rows.findIndex((r) => r.product_id === productId);
		if (startRow === -1) return;

		const matrix = text
			.replace(/\r/g, '')
			.split('\n')
			.filter((line) => line.trim() !== '')
			.map((line) => line.split('\t'));

		let filled = 0;
		matrix.forEach((line, rowOffset) => {
			const targetRow = rows[startRow + rowOffset];
			// Baris kosong di grid = produk belum dipilih; angka tak bisa ditempel ke sana.
			if (!targetRow?.product_id) return;
			line.forEach((raw, colOffset) => {
				const day = startDay + colOffset;
				if (day > daysInMonth) return;
				const n = parseInt(raw.trim().replace(/[^\d-]/g, '')) || 0;
				const k = cellKey(targetRow.product_id, day);
				if (!cells[k]) cells[k] = { in: 0, out: 0 };
				cells[k][mode] = Math.max(0, n);
				filled++;
			});
		});

		cells = { ...cells }; // picu re-render
		if (filled > 0) toastStore.info(`${filled} sel terisi dari clipboard`);
	}

	/** Enter / Shift+Enter pindah ke sel di bawah/atasnya, seperti Excel. */
	function onCellKeydown(e: KeyboardEvent, rowIdx: number, d: number) {
		if (e.key !== 'Enter' && e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return;
		e.preventDefault();
		const step = e.key === 'ArrowUp' || e.shiftKey ? -1 : 1;
		const nextRow = rows[rowIdx + step];
		if (!nextRow?.product_id) return;
		document
			.querySelector<HTMLInputElement>(`[data-cell="${nextRow.product_id}:${d}"]`)
			?.focus();
	}

	function hasChanges(): boolean {
		return rows.some((r) => r.product_id && days.some((d) => deltaOf(r.product_id, d) !== 0));
	}

	/** Total kolom per produk (semua tanggal), seperti kolom SUM di Excel. */
	function rowTotal(productId: string): number {
		return days.reduce((sum, d) => sum + (cells[cellKey(productId, d)]?.[mode] ?? 0), 0);
	}

	/** Total baris per tanggal (semua produk). */
	function dayTotal(d: number): number {
		return rows.reduce(
			(sum, r) => sum + (r.product_id ? cells[cellKey(r.product_id, d)]?.[mode] ?? 0 : 0),
			0
		);
	}

	function grandTotal(): number {
		return days.reduce((sum, d) => sum + dayTotal(d), 0);
	}

	async function save() {
		// Kirim SELISIH, bukan seluruh isi grid. Tanpa ini klik Simpan kedua kali
		// memposting ulang semua sel dan stok terlipat dua.
		const additions: any[] = []; // isian baru -> dokumen normal
		const reductions: any[] = []; // koreksi turun -> dokumen lawan berkode 'other'

		for (const row of rows) {
			if (!row.product_id) continue;
			for (const d of days) {
				const delta = deltaOf(row.product_id, d);
				if (delta === 0) continue;
				const item = {
					product_id: row.product_id,
					date: `${month}-${String(d).padStart(2, '0')}`,
					quantity: Math.abs(delta),
					unit_cost: unitCosts[row.product_id] || 0
				};
				if (delta > 0) additions.push(item);
				else reductions.push(item);
			}
		}

		if (additions.length === 0 && reductions.length === 0) {
			toastStore.info('Tidak ada perubahan untuk disimpan');
			return;
		}

		isSaving = true;
		try {
			const post = (entries: any[], movementType: 'in' | 'out', subType: string) =>
				apiClient.post<{ movements_created: number; items_posted: number }>(
					API_ENDPOINTS.inventoryGrid,
					{ warehouse_id: warehouseId, movement_type: movementType, sub_type: subType, entries }
				);

			let posted = 0;
			let docs = 0;

			if (additions.length > 0) {
				const res = await post(additions, mode, mode === 'in' ? 'purchase' : 'sale');
				if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyimpan');
				posted += res.data.items_posted;
				docs += res.data.movements_created;
			}

			if (reductions.length > 0) {
				// Koreksi turun dicatat sebagai pergerakan berlawanan dengan alasan 'other',
				// supaya stok & kartu stok tetap seimbang dan bisa diaudit.
				const opposite = mode === 'in' ? 'out' : 'in';
				const res = await post(reductions, opposite, 'other');
				if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyimpan koreksi');
				posted += res.data.items_posted;
				docs += res.data.movements_created;
			}

			toastStore.success(`${posted} isian tersimpan (${docs} dokumen)`);
			await loadGrid();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal menyimpan');
		} finally {
			isSaving = false;
		}
	}

	// Copy isian grid bulan lalu ke bulan aktif (produk & hari sama).
	let isCopying = false;
	async function copyPrevMonth() {
		if (!warehouseId) return;
		const srcMonth = previousMonth(month);
		if (srcMonth === month) {
			toastStore.error('Bulan sumber dan tujuan sama');
			return;
		}

		isCopying = true;
		try {
			const res = await apiClient.post<{ movements_created: number; items_posted: number }>(
				API_ENDPOINTS.inventoryGridCopy,
				{
					warehouse_id: warehouseId,
					src_month: srcMonth,
					dst_month: month,
					movement_type: mode,
					sub_type: mode === 'in' ? 'purchase' : 'sale'
				}
			);
			if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyalin');
			toastStore.success(`${res.data.items_posted} isian disalin dari ${srcMonth}`);
			await loadGrid();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal menyalin bulan lalu');
		} finally {
			isCopying = false;
		}
	}

	// Ekspor grid bulan aktif ke Excel (unduh file).
	let isExporting = false;
	async function exportGrid() {
		if (!warehouseId) return;
		isExporting = true;
		try {
			const token = localStorage.getItem('access_token');
			const url = `${API_ENDPOINTS.inventoryGrid}/export?warehouse_id=${warehouseId}&month=${month}&movement_type=${mode}`;
			const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } });
			if (!res.ok) throw new Error('Gagal mengekspor grid');
			const blob = await res.blob();
			const dlUrl = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = dlUrl;
			a.download = `inventory-grid-${mode}-${month}.xlsx`;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(dlUrl);
			document.body.removeChild(a);
			toastStore.success('Grid diekspor ke Excel');
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal mengekspor grid');
		} finally {
			isExporting = false;
		}
	}

	// Autocomplete produk. Baris kosong selalu sama-sama tak punya product_id, jadi
	// pelacakannya pakai referensi objek baris, bukan index.
	let searchTimer: ReturnType<typeof setTimeout> | null = null;
	let activeRow: Row | null = null;

	function onNameInput(row: Row) {
		activeRow = row;
		if (searchTimer) clearTimeout(searchTimer);
		const q = row.name ?? '';
		if (!q || q.length < 2) {
			searchResults = [];
			return;
		}
		searchTimer = setTimeout(() => doSearch(q), 300);
	}

	async function doSearch(q: string) {
		try {
			const res = await apiClient.get<{ products: any[] }>(
				`${API_ENDPOINTS.products}?limit=10&search=${encodeURIComponent(q)}`
			);
			searchResults = res.success && res.data ? res.data.products || [] : [];
		} catch {
			searchResults = [];
		}
	}

	function pickProduct(row: Row, p: any) {
		// Isian lama baris ini milik produk sebelumnya — jangan terbawa.
		if (row.product_id) {
			for (const d of days) {
				delete cells[cellKey(row.product_id, d)];
				delete saved[cellKey(row.product_id, d)];
			}
		}
		row.product_id = p.id;
		row.name = p.name;
		row.product_code = p.product_code;
		searchResults = [];
		activeRow = null;
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
	<div class="page page-with-bottombar">
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
				<button class="btn btn-ghost btn-sm" on:click={exportGrid} disabled={isExporting}>
					{#if isExporting}
						<span class="loading loading-spinner loading-xs"></span>
					{/if}
					📥 Ekspor Excel
				</button>
				<button class="btn btn-ghost btn-sm" on:click={copyPrevMonth} disabled={isCopying}>
					{#if isCopying}
						<span class="loading loading-spinner loading-xs"></span>
					{/if}
					📋 Copy bulan lalu
				</button>
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
								<th class="sticky left-0 top-0 z-20 bg-base-200 min-w-[14rem]">Produk</th>
								{#each days as d}
									<th class="sticky top-0 z-10 bg-base-200 text-center text-xs" title="Tanggal {d}">{d}</th>
								{/each}
								<th class="sticky top-0 z-10 bg-base-200"></th>
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
												on:input={() => onNameInput(row)}
											/>
											{#if activeRow === row && searchResults.length > 0}
												<div class="absolute z-30 mt-1 max-h-48 w-full overflow-y-auto rounded-xl border border-base-300 bg-base-100 shadow-lg">
													{#each searchResults as p}
														<button class="block w-full px-2 py-1.5 text-left text-xs hover:bg-base-200"
															on:click={() => pickProduct(row, p)}>
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
												disabled={!row.product_id}
												data-cell={row.product_id ? cellKey(row.product_id, d) : undefined}
												class="w-14 border-0 bg-transparent text-center text-sm focus:bg-primary/10 focus:outline-none disabled:opacity-30"
												class:pending-cell={row.product_id && deltaOf(row.product_id, d) !== 0}
												value={row.product_id ? (cells[cellKey(row.product_id, d)]?.[mode] ?? '') : ''}
												on:input={(e) => row.product_id && onInput(row.product_id, d, mode, e.currentTarget.value)}
												on:paste={(e) => row.product_id && onPaste(e, row.product_id, d)}
												on:keydown={(e) => onCellKeydown(e, i, d)}
											/>
										</td>
									{/each}
									<td class="p-0">
										<div class="w-14 text-center text-xs font-medium text-base-content/60">
											{row.product_id ? rowTotal(row.product_id) || '' : ''}
										</div>
									</td>
									<td>
										{#if rows.length > 1}
											<button class="btn btn-ghost btn-xs text-error" on:click={() => removeRow(i)}>✕</button>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
						<tfoot>
							<tr>
								<th class="sticky left-0 z-10 bg-base-300 text-xs">Total / hari</th>
								{#each days as d}
									<th class="bg-base-300 text-center text-xs font-medium">
										{dayTotal(d) || ''}
									</th>
								{/each}
								<th class="bg-base-300 text-center text-xs">{grandTotal() || ''}</th>
								<th class="bg-base-300"></th>
							</tr>
						</tfoot>
					</table>
				</div>
			</div>

			<!-- Simpan -->
			<div class="mt-4 flex items-center justify-end gap-3">
				<p class="text-xs text-base-content/60">
					Sel kuning = belum tersimpan. Hanya selisih yang dikirim.
				</p>
				<button class="btn btn-primary btn-md" on:click={save} disabled={isSaving || !hasChanges()}>
					{#if isSaving}
						<span class="loading loading-spinner loading-sm"></span>
						Menyimpan...
					{:else}
						Simpan Perubahan
					{/if}
				</button>
			</div>
		{/if}
	</div>
</LayoutWrapper>

<style>
	/* Sel yang diubah tapi belum disimpan. */
	.pending-cell {
		background-color: color-mix(in srgb, var(--fallback-wa, oklch(var(--wa))) 25%, transparent);
	}
</style>
