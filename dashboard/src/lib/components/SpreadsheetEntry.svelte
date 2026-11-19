<script lang="ts">
	/**
	 * SpreadsheetEntry — tabel keluar-masuk produk ala Excel.
	 * - Baris = satu item: jenis (Masuk/Keluar) + produk + tanggal + qty + harga satuan.
	 * - Ctrl+C/Ctrl+V dari Excel tetap jalan (paste banyak baris sekaligus).
	 * - Fullscreen: tombol, atau F11 saat tabel terfokus.
	 * - Simpan = kirim setiap baris ke backend (endpoint /inventory/in dan /inventory/out).
	 *
	 * Belum ada: formula Excel (SUM, referensi sel), multi-sheet, sel merge.
	 */
	import { createEventDispatcher } from 'svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { localMonth, localDate } from '$lib/utils/date';

	type Kind = 'in' | 'out';
	type Row = {
		_id: number; // id lokal baris, untuk :each key
		kind: Kind;
		product_id: string;
		name: string;
		product_code: string;
		date: string; // YYYY-MM-DD
		qty: number;
		unitCost: number;
	};

	export let warehouseId = '';
	export let month = localMonth();

	const dispatch = createEventDispatcher<{ saved: number }>();

	let rows: Row[] = [];
	let isSaving = false;
	let isFullscreen = false;
	let nextId = 1;

	function fmtRp(n: number): string {
		return 'Rp ' + new Intl.NumberFormat('id-ID').format(Math.round(n ?? 0));
	}

	// ---- Autocomplete
	let searchResults: any[] = [];
	let activeSearchRowId: number | null = null;
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	function today(): string {
		const now = new Date();
		if (now.toISOString().slice(0, 7) !== month) return `${month}-01`;
		return localDate();
	}

	function emptyRow(kind: Kind = 'out'): Row {
		return {
			_id: nextId++,
			kind,
			product_id: '',
			name: '',
			product_code: '',
			date: today(),
			qty: 1,
			unitCost: 0
		};
	}

	$: totalRows = rows.length;
	$: totalIn = rows.filter((r) => r.kind === 'in').reduce((s, r) => s + (r.qty || 0), 0);
	$: totalOut = rows.filter((r) => r.kind === 'out').reduce((s, r) => s + (r.qty || 0), 0);
	$: totalValueIn = rows.filter((r) => r.kind === 'in').reduce((s, r) => s + (r.qty || 0) * (r.unitCost || 0), 0);
	$: totalValueOut = rows.filter((r) => r.kind === 'out').reduce((s, r) => s + (r.qty || 0) * (r.unitCost || 0), 0);

	// ---- Baris
	function addRow(kind: Kind = 'out'): Row {
		const r = emptyRow(kind);
		rows = [...rows, r];
		return r;
	}

	function removeRow(id: number) {
		rows = rows.filter((r) => r._id !== id);
	}

	function duplicateRow(id: number) {
		const src = rows.find((r) => r._id === id);
		if (!src) return;
		const copy: Row = { ...src, _id: nextId++ };
		const idx = rows.findIndex((r) => r._id === id);
		rows = [...rows.slice(0, idx + 1), copy, ...rows.slice(idx + 1)];
	}

	function clearAll() {
		rows = [addRow('out')];
	}

	// ---- Autocomplete produk per baris
	function onNameInput(row: Row) {
		activeSearchRowId = row._id;
		if (searchTimer) clearTimeout(searchTimer);
		const q = row.name;
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
		const idx = rows.findIndex((r) => r._id === row._id);
		if (idx === -1) return;
		rows[idx] = {
			...rows[idx],
			product_id: p.id,
			name: p.name,
			product_code: p.product_code
		};
		searchResults = [];
		activeSearchRowId = null;
	}

	function onRowKindChange(row: Row, e: Event) {
		const val = (e.currentTarget as HTMLSelectElement).value as Kind;
		const idx = rows.findIndex((r) => r._id === row._id);
		if (idx !== -1) rows[idx] = { ...rows[idx], kind: val };
	}

	function onRowInput(row: Row, field: 'qty' | 'unitCost' | 'date', e: Event) {
		const idx = rows.findIndex((r) => r._id === row._id);
		if (idx === -1) return;
		const raw = (e.currentTarget as HTMLInputElement).value;
		if (field === 'date') {
			rows[idx] = { ...rows[idx], date: raw };
		} else {
			const n = Math.max(0, parseInt(raw, 10) || 0);
			rows[idx] = { ...rows[idx], [field]: n };
		}
	}

	// ---- Paste dari Excel
	function onPaste(e: ClipboardEvent, targetRow: Row) {
		const text = e.clipboardData?.getData('text/plain');
		if (!text || !text.includes('\t')) return; // satu kolom: biarkan paste biasa
		e.preventDefault();

		const lines = text.replace(/\r/g, '').split('\n').filter((l) => l.trim() !== '');
		const startIdx = rows.findIndex((r) => r._id === targetRow._id);
		if (startIdx === -1) return;

		const newRows: Row[] = [];
		lines.forEach((line, i) => {
			const cols = line.split('\t').map((c) => c.trim());
			// Kolom paste Excel: name | qty | unitCost (opsional)
			const [name = '', qtyStr = '1', costStr = '0'] = cols;
			const src = rows[startIdx + i];
			newRows.push({
				...emptyRow(src?.kind ?? 'out'),
				name,
				qty: Math.max(1, parseInt(qtyStr.replace(/[^\d-]/g, ''), 10) || 1),
				unitCost: parseFloat(costStr.replace(/[^\d.-]/g, '')) || 0
			});
		});

		rows = [...rows.slice(0, startIdx), ...newRows, ...rows.slice(startIdx + 1)];
		toastStore.info(`${newRows.length} baris ditempel dari clipboard`);
	}

	// ---- Fullscreen
	function toggleFullscreen() {
		isFullscreen = !isFullscreen;
	}

	function onTableKeydown(e: KeyboardEvent) {
		if (!isFullscreen) return;
		if (e.key === 'F11' || (e.key === 'Escape' && isFullscreen)) {
			e.preventDefault();
			isFullscreen = false;
		}
	}

	// ---- Simpan
	async function save() {
		const valid = rows.filter((r) => r.product_id && r.qty > 0);
		if (valid.length === 0) {
			toastStore.error('Belum ada baris valid. Isi produk dan qty dulu.');
			return;
		}
		const invalid = rows.length - valid.length;
		if (invalid > 0) {
			toastStore.error(`${invalid} baris dilewati: produk belum dipilih atau qty = 0`);
		}

		// Kelompokkan per kind dan per tanggal (satu dokumen per tanggal, sesuai PostGrid).
		const inEntries: any[] = [];
		const outEntries: any[] = [];
		for (const r of valid) {
			const entry = {
				product_id: r.product_id,
				date: r.date,
				quantity: r.qty,
				unit_cost: r.unitCost
			};
			if (r.kind === 'in') inEntries.push(entry);
			else outEntries.push(entry);
		}

		isSaving = true;
		try {
			let docs = 0;
			let items = 0;

			// PostGrid satu-satunya endpoint yang menerima movement_date per item.
			// Kirim dua panggilan: satu untuk Masuk, satu untuk Keluar.
			const post = (entries: any[], movementType: 'in' | 'out', subType: string) =>
				apiClient.post<{ movements_created: number; items_posted: number }>(
					API_ENDPOINTS.inventoryGrid,
					{ warehouse_id: warehouseId, movement_type: movementType, sub_type: subType, entries }
				);

			if (inEntries.length > 0) {
				const res = await post(inEntries, 'in', 'purchase');
				if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyimpan barang masuk');
				docs += res.data.movements_created;
				items += res.data.items_posted;
			}

			if (outEntries.length > 0) {
				const res = await post(outEntries, 'out', 'sale');
				if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyimpan barang keluar');
				docs += res.data.movements_created;
				items += res.data.items_posted;
			}

			toastStore.success(`${items} item tersimpan (${docs} dokumen)`);
			dispatch('saved', items);
			clearAll();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal menyimpan');
		} finally {
			isSaving = false;
		}
	}
</script>

<svelte:window on:keydown={onTableKeydown} />

<!-- Kontainer spreadsheet: fullscreen = fixed penuh layar -->
<div class:spreadsheet-fullscreen={isFullscreen} class="spreadsheet-root">
	<!-- Header -->
	<div class="mb-3 flex flex-wrap items-center justify-between gap-2">
		<div class="flex items-center gap-3">
			<h2 class="text-lg font-semibold">Keluar-Masuk Produk</h2>
			<span class="text-xs text-base-content/60">{totalRows} baris</span>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<button class="btn btn-ghost btn-sm" on:click={() => addRow('in')}>
				+ Masuk
			</button>
			<button class="btn btn-ghost btn-sm" on:click={() => addRow('out')}>
				+ Keluar
			</button>
			<button class="btn btn-ghost btn-sm" on:click={toggleFullscreen} aria-label="Layar penuh">
				{#if isFullscreen}
					❌ Keluar layar penuh
				{:else}
					⛶ Layar penuh (F11)
				{/if}
			</button>
			<button class="btn btn-ghost btn-sm text-warning" on:click={clearAll} disabled={isSaving}>
				↺ Reset
			</button>
			<button class="btn btn-primary btn-sm" on:click={save} disabled={isSaving}>
				{#if isSaving}
					<span class="loading loading-spinner loading-sm"></span>
					Menyimpan...
				{:else}
					Simpan Semua
				{/if}
			</button>
		</div>
	</div>

	<!-- Ringkasan -->
	<div class="mb-3 grid grid-cols-2 gap-3 md:grid-cols-4">
		<div class="card-apple p-3">
			<div class="text-xs text-base-content/60">Total Masuk</div>
			<div class="metric text-xl font-semibold text-success">{totalIn}</div>
		</div>
		<div class="card-apple p-3">
			<div class="text-xs text-base-content/60">Total Keluar</div>
			<div class="metric text-xl font-semibold text-error">{totalOut}</div>
		</div>
		<div class="card-apple p-3">
			<div class="text-xs text-base-content/60">Nilai Masuk</div>
			<div class="metric text-sm font-medium text-success">{fmtRp(totalValueIn)}</div>
		</div>
		<div class="card-apple p-3">
			<div class="text-xs text-base-content/60">Nilai Keluar</div>
			<div class="metric text-sm font-medium text-error">{fmtRp(totalValueOut)}</div>
		</div>
	</div>

	<!-- Tabel spreadsheet -->
	<div class="overflow-hidden rounded-2xl border border-base-300 bg-base-200">
		<div class="spreadsheet-scroll overflow-auto">
			<table class="table table-sm">
				<thead>
					<tr>
						<th class="bg-base-300 text-center text-xs" style="width: 3rem">#</th>
						<th class="bg-base-300 text-xs" style="width: 8rem">Jenis</th>
						<th class="bg-base-300 text-xs">Produk</th>
						<th class="bg-base-300 text-xs" style="width: 10rem">Tanggal</th>
						<th class="bg-base-300 text-center text-xs" style="width: 6rem">Qty</th>
						<th class="bg-base-300 text-right text-xs" style="width: 8rem">Harga</th>
						<th class="bg-base-300 text-right text-xs" style="width: 10rem">Subtotal</th>
						<th class="bg-base-300 text-center text-xs" style="width: 5rem"></th>
					</tr>
				</thead>
				<tbody>
					{#each rows as row, i (row._id)}
						<tr class="spreadsheet-row" class:bg-base-100={i % 2 === 0}>
							<td class="text-center text-xs text-base-content/40">{i + 1}</td>
							<td>
								<select
									class="select select-bordered select-xs w-full"
									value={row.kind}
									on:change={(e) => onRowKindChange(row, e)}
								>
									<option value="in">Masuk</option>
									<option value="out">Keluar</option>
								</select>
							</td>
							<td class="relative">
								<input
									class="input input-bordered input-xs w-full"
									placeholder="Cari / pilih produk"
									bind:value={row.name}
									on:input={() => onNameInput(row)}
								/>
								{#if activeSearchRowId === row._id && searchResults.length > 0}
									<div class="absolute z-30 mt-1 max-h-48 w-full overflow-y-auto rounded-xl border border-base-300 bg-base-100 shadow-lg">
										{#each searchResults as p}
											<button class="block w-full px-3 py-1.5 text-left text-xs hover:bg-base-200"
												on:click={() => pickProduct(row, p)}>
												<div class="truncate font-medium">{p.name}</div>
												<div class="metric text-[10px] text-base-content/60">{p.product_code}</div>
											</button>
										{/each}
									</div>
								{/if}
							</td>
							<td>
								<input
									type="date"
									class="input input-bordered input-xs w-full"
									value={row.date}
									on:input={(e) => onRowInput(row, 'date', e)}
								/>
							</td>
							<td>
								<input
									type="number" min="1"
									class="input input-bordered input-xs w-full text-center"
									value={row.qty}
									on:input={(e) => onRowInput(row, 'qty', e)}
								/>
							</td>
							<td>
								<input
									type="number" min="0" step="0.01"
									class="input input-bordered input-xs w-full text-right"
									value={row.unitCost}
									on:input={(e) => onRowInput(row, 'unitCost', e)}
								/>
							</td>
							<td class="text-right text-xs font-medium">
								{fmtRp((row.qty || 0) * (row.unitCost || 0))}
							</td>
							<td class="text-center">
								<button class="btn btn-ghost btn-xs" title="Duplikat" aria-label="Duplikat baris"
									on:click={() => duplicateRow(row._id)}>⧉</button>
								<button class="btn btn-ghost btn-xs text-error" title="Hapus" aria-label="Hapus baris"
									on:click={() => removeRow(row._id)}>✕</button>
							</td>
						</tr>
					{/each}
				</tbody>
				<tfoot>
					<tr>
						<td colspan="4" class="bg-base-300 text-xs font-medium">Total</td>
						<td class="bg-base-300 text-center text-xs font-semibold">
							<span class="text-success">{totalIn}</span> / <span class="text-error">{totalOut}</span>
						</td>
						<td class="bg-base-300"></td>
						<td class="bg-base-300 text-right text-xs font-semibold">
							{fmtRp(totalValueIn + totalValueOut)}
						</td>
						<td class="bg-base-300"></td>
					</tr>
				</tfoot>
			</table>
		</div>
	</div>

	<p class="mt-2 text-xs text-base-content/50">
		Tips: Copy beberapa baris di Excel lalu klik di kolom Produk baris pertama di sini dan tekan
		Ctrl+V — kolom name/qty/harga terisi otomatis. Tekan F11 saat tabel aktif untuk keluar dari
		layar penuh.
	</p>
</div>

<style>
	.spreadsheet-root {
		position: relative;
	}

	.spreadsheet-scroll {
		max-height: 70vh;
	}

	.spreadsheet-fullscreen {
		position: fixed;
		inset: 0;
		z-index: 50;
		background: hsl(var(--b2));
		padding: 1rem;
		overflow: auto;
	}

	.spreadsheet-fullscreen .spreadsheet-scroll {
		max-height: calc(100vh - 12rem);
	}

	.spreadsheet-row:nth-child(even) {
		background-color: color-mix(in srgb, hsl(var(--b1)) 70%, transparent);
	}
</style>
