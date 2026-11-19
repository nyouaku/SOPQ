<script lang="ts">
	/**
	 * GridEntry — tiga mode input pergerakan harian dalam satu komponen.
	 * mode 'grid'    : grid 31 kolom ala Excel (utama)
	 * mode 'compact' : hanya menampilkan kolom yang punya isian
	 * mode 'quick'   : form per-transaksi (pilih produk → tanggal → qty)
	 *
	 * Semua mode memakai mekanisme kirim sama: kirim SELISIH terhadap baseline
	 * `saved`, jadi menekan Simpan berulang tak menggandakan stok.
	 */
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { localMonth } from '$lib/utils/date';

	type Mode = 'grid' | 'compact' | 'quick';
	type Row = { product_id: string; name: string; product_code: string };

	export let viewMode: Mode = 'grid';
	export let onViewModeChange: (m: Mode) => void = () => {};
	/** Gudang, arah pergerakan (in/out), dan bulan dikontrol parent agar toolbar tersatukan. */
	export let warehouseId = '';
	export let direction: 'out' | 'in' = 'out';
	export let month = localMonth();

	let rows: Row[] = [];
	/** Isian aktif: key `${product_id}:${day}` → {in,out}. */
	let cells: Record<string, { in: number; out: number }> = {};
	/** Baseline tersimpan di server. Simpan hanya mengirim selisih terhadap ini. */
	let saved: Record<string, { in: number; out: number }> = {};
	let unitCosts: Record<string, number> = {};
	let days: number[] = [];
	let daysInMonth = 31;
	let isLoading = true;
	let isSaving = false;
	let searchResults: any[] = [];
	let activeRow: Row | null = null;

	// Mode quick: daftar baris baru yang siap dikirim.
	type QuickLine = { product_id: string; name: string; product_code: string; day: number; qty: number };
	let quickLines: QuickLine[] = [];
	let quickProduct: Row | null = null;
	let quickProductName = '';
	let quickDay = new Date().getDate();
	let quickQty = 1;
	let quickUnitCost = 0;

	$: daysInMonth = new Date(parseInt(month.slice(0, 4)), parseInt(month.slice(5, 7)), 0).getDate();
	$: days = Array.from({ length: daysInMonth }, (_, i) => i + 1);
	$: filteredRows = filterRows();

	/** Baris produk yang punya setidaknya satu isian (dipakai mode compact). */
	$: populatedDayCols = Array.from(
		new Set(
			Object.entries(cells)
				.filter(([, v]) => (direction === 'in' ? v.in : v.out) > 0)
				.map(([k]) => parseInt(k.split(':')[1], 10))
		)
	).sort((a, b) => a - b);

	function filterRows(): Row[] {
		// Mode compact: sembunyikan baris tanpa isian.
		if (viewMode === 'compact') return rows.filter((r) => r.product_id && rowTotal(r.product_id) > 0);
		return rows;
	}

	async function loadGrid() {
		if (!warehouseId) return;
		isLoading = true;
		try {
			const res = await apiClient.get<{ cells: any[] }>(
				`${API_ENDPOINTS.inventoryGrid}?warehouse_id=${warehouseId}&month=${month}&movement_type=${direction}`
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
					if (direction === 'in') cells[k].in = c.in_qty;
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

	function cellKey(productId: string, d: number) {
		return `${productId}:${d}`;
	}

	function onInput(productId: string, d: number, val: string) {
		const n = Math.max(0, parseInt(val) || 0);
		const k = cellKey(productId, d);
		if (!cells[k]) cells[k] = { in: 0, out: 0 };
		cells[k][direction] = n;
	}

	function deltaOf(productId: string, d: number): number {
		const k = cellKey(productId, d);
		const cur = cells[k]?.[direction] ?? 0;
		const base = saved[k]?.[direction] ?? 0;
		return cur - base;
	}

	function rowTotal(productId: string): number {
		return days.reduce((sum, d) => sum + (cells[cellKey(productId, d)]?.[direction] ?? 0), 0);
	}

	function dayTotal(d: number): number {
		return rows.reduce(
			(sum, r) => sum + (r.product_id ? cells[cellKey(r.product_id, d)]?.[direction] ?? 0 : 0),
			0
		);
	}

	function grandTotal(): number {
		return days.reduce((sum, d) => sum + dayTotal(d), 0);
	}

	function hasChanges(): boolean {
		return rows.some((r) => r.product_id && days.some((d) => deltaOf(r.product_id, d) !== 0));
	}

	function onPaste(e: ClipboardEvent, productId: string, startDay: number) {
		const text = e.clipboardData?.getData('text/plain');
		if (!text || !text.includes('\t')) return;
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
			if (!targetRow?.product_id) return;
			line.forEach((raw, colOffset) => {
				const day = startDay + colOffset;
				if (day > daysInMonth) return;
				const n = Math.max(0, parseInt(raw.trim().replace(/[^\d-]/g, '')) || 0);
				const k = cellKey(targetRow.product_id, day);
				if (!cells[k]) cells[k] = { in: 0, out: 0 };
				cells[k][direction] = n;
				filled++;
			});
		});
		cells = { ...cells };
		if (filled > 0) toastStore.info(`${filled} sel terisi dari clipboard`);
	}

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

	// ---- Autocomplete produk
	let searchTimer: ReturnType<typeof setTimeout> | null = null;
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

	function addRow() {
		rows = [...rows, { product_id: '', name: '', product_code: '' }];
	}

	function removeRow(i: number) {
		rows = rows.filter((_, idx) => idx !== i);
	}

	// ---- Mode quick: tambah baris
	function addQuickLine() {
		if (!quickProduct) {
			toastStore.error('Pilih produk terlebih dahulu');
			return;
		}
		if (quickQty <= 0) {
			toastStore.error('Jumlah harus lebih dari 0');
			return;
		}
		quickLines = [
			...quickLines,
			{
				product_id: quickProduct.product_id,
				name: quickProduct.name,
				product_code: quickProduct.product_code,
				day: quickDay,
				qty: quickQty
			}
		];
		// Setelah simpan ke daftar, reset untuk input berikutnya.
		quickProduct = null;
		quickProductName = '';
		quickQty = 1;
	}

	function removeQuickLine(i: number) {
		quickLines = quickLines.filter((_, idx) => idx !== i);
	}

	function onQuickNameInput() {
		if (searchTimer) clearTimeout(searchTimer);
		const q = quickProductName;
		if (!q || q.length < 2) {
			searchResults = [];
			return;
		}
		searchTimer = setTimeout(() => doSearch(q), 300);
		activeRow = { product_id: '', name: q, product_code: '' };
	}

	function pickQuickProduct(p: any) {
		quickProduct = { product_id: p.id, name: p.name, product_code: p.product_code };
		quickProductName = p.name;
		searchResults = [];
		activeRow = null;
	}

	async function save() {
		let additions: any[] = [];
		let reductions: any[] = [];

		if (viewMode === 'quick') {
			// Mode quick: semua baris adalah tambahan; koreksi tak tersedia di sini.
			if (quickLines.length === 0) {
				toastStore.info('Belum ada baris untuk disimpan');
				return;
			}
			additions = quickLines.map((l) => ({
				product_id: l.product_id,
				date: `${month}-${String(l.day).padStart(2, '0')}`,
				quantity: l.qty,
				unit_cost: quickUnitCost || 0
			}));
		} else {
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
				const res = await post(additions, direction, direction === 'in' ? 'purchase' : 'sale');
				if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyimpan');
				posted += res.data.items_posted;
				docs += res.data.movements_created;
			}

			if (reductions.length > 0) {
				const opposite = direction === 'in' ? 'out' : 'in';
				const res = await post(reductions, opposite, 'other');
				if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyimpan koreksi');
				posted += res.data.items_posted;
				docs += res.data.movements_created;
			}

			toastStore.success(`${posted} isian tersimpan (${docs} dokumen)`);
			quickLines = [];
			await loadGrid();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal menyimpan');
		} finally {
			isSaving = false;
		}
	}

	$: if (warehouseId || month || direction) loadGrid();

	export function reload() {
		loadGrid();
	}
</script>

<!-- Toolbar mode -->
<div class="mb-4 flex flex-wrap items-center gap-2">
	<div class="grid grid-cols-3 gap-1 rounded-xl bg-base-300 p-1" role="tablist">
		<button role="tab" aria-selected={viewMode === 'grid'}
			class="rounded-lg px-3 py-1.5 text-sm font-medium {viewMode === 'grid' ? 'bg-base-100 shadow-sm' : 'opacity-50'}"
			on:click={() => onViewModeChange('grid')}>📊 Grid</button>
		<button role="tab" aria-selected={viewMode === 'compact'}
			class="rounded-lg px-3 py-1.5 text-sm font-medium {viewMode === 'compact' ? 'bg-base-100 shadow-sm' : 'opacity-50'}"
			on:click={() => onViewModeChange('compact')}>📋 Ringkas</button>
		<button role="tab" aria-selected={viewMode === 'quick'}
			class="rounded-lg px-3 py-1.5 text-sm font-medium {viewMode === 'quick' ? 'bg-base-100 shadow-sm' : 'opacity-50'}"
			on:click={() => onViewModeChange('quick')}>⚡ Cepat</button>
	</div>
</div>

{#if isLoading}
	<div class="flex h-64 items-center justify-center">
		<span class="loading loading-spinner loading-lg"></span>
	</div>
{:else if viewMode === 'quick'}
	<!-- Mode cepat: form per-transaksi -->
	<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
		<h2 class="mb-3 font-semibold">Tambah pergerakan</h2>
		<div class="grid grid-cols-1 gap-3 md:grid-cols-5">
			<div class="relative md:col-span-2">
				<label class="label py-0" for="q-prod">
					<span class="label-text text-xs text-base-content/60">Produk</span>
				</label>
				<input
					id="q-prod"
					class="input input-bordered input-sm mt-1 w-full"
					placeholder="Cari nama produk..."
					bind:value={quickProductName}
					on:input={onQuickNameInput}
				/>
				{#if activeRow?.name === quickProductName && searchResults.length > 0}
					<div class="absolute z-30 mt-1 max-h-48 w-full overflow-y-auto rounded-xl border border-base-300 bg-base-100 shadow-lg">
						{#each searchResults as p}
							<button class="block w-full px-3 py-2 text-left text-sm hover:bg-base-200"
								on:click={() => pickQuickProduct(p)}>
								<div class="truncate font-medium">{p.name}</div>
								<div class="metric text-xs text-base-content/60">{p.product_code}</div>
							</button>
						{/each}
					</div>
				{/if}
				{#if quickProduct}
					<p class="mt-1 text-xs text-success">
						✓ {quickProduct.name} ({quickProduct.product_code})
					</p>
				{/if}
			</div>

			<div>
				<label class="label py-0" for="q-day">
					<span class="label-text text-xs text-base-content/60">Tanggal</span>
				</label>
				<input id="q-day" type="number" min="1" max={daysInMonth}
					class="input input-bordered input-sm mt-1 w-full"
					bind:value={quickDay} />
			</div>

			<div>
				<label class="label py-0" for="q-qty">
					<span class="label-text text-xs text-base-content/60">Jumlah</span>
				</label>
				<input id="q-qty" type="number" min="1"
					class="input input-bordered input-sm mt-1 w-full"
					bind:value={quickQty} />
			</div>

			<div class="flex items-end">
				<button class="btn btn-primary btn-sm w-full" on:click={addQuickLine}>
					+ Tambah
				</button>
			</div>
		</div>

		{#if quickLines.length > 0}
			<div class="mt-4 overflow-x-auto">
				<table class="table table-sm">
					<thead>
						<tr>
							<th>Produk</th>
							<th>Tanggal</th>
							<th>Jumlah</th>
							<th></th>
						</tr>
					</thead>
					<tbody>
						{#each quickLines as line, i}
							<tr>
								<td class="text-sm">{line.name}</td>
								<td class="text-xs text-base-content/60">Tgl {line.day}</td>
								<td class="metric font-semibold">{line.qty}</td>
								<td>
									<button class="btn btn-ghost btn-xs text-error"
										on:click={() => removeQuickLine(i)}>✕</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
{:else}
	<!-- Mode grid & compact -->
	<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
		{#if viewMode === 'compact' && populatedDayCols.length === 0}
			<p class="py-8 text-center text-sm text-base-content/60">
				Belum ada isian di bulan ini. Pindah ke mode "Grid" untuk mulai mengisi.
			</p>
		{:else}
			<div class="overflow-x-auto">
				<table class="table table-sm">
					<thead>
						<tr>
							<th class="sticky left-0 top-0 z-20 bg-base-200 min-w-[14rem]">Produk</th>
							{#if viewMode === 'compact'}
								{#each populatedDayCols as d}
									<th class="sticky top-0 z-10 bg-base-200 text-center text-xs" title="Tanggal {d}">{d}</th>
								{/each}
							{:else}
								{#each days as d}
									<th class="sticky top-0 z-10 bg-base-200 text-center text-xs" title="Tanggal {d}">{d}</th>
								{/each}
							{/if}
							<th class="sticky top-0 z-10 bg-base-200 text-center text-xs">Total</th>
							<th class="sticky top-0 z-10 bg-base-200"></th>
						</tr>
					</thead>
					<tbody>
						{#each filteredRows as row, i}
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

								{#if viewMode === 'compact'}
									{#each populatedDayCols as d}
										<td class="p-0">
											<input
												type="number"
												min="0"
												disabled={!row.product_id}
												data-cell={row.product_id ? cellKey(row.product_id, d) : undefined}
												class="w-14 border-0 bg-transparent text-center text-sm focus:bg-primary/10 focus:outline-none disabled:opacity-30"
												class:pending-cell={row.product_id && deltaOf(row.product_id, d) !== 0}
												value={row.product_id ? (cells[cellKey(row.product_id, d)]?.[direction] ?? '') : ''}
												on:input={(e) => row.product_id && onInput(row.product_id, d, e.currentTarget.value)}
												on:paste={(e) => row.product_id && onPaste(e, row.product_id, d)}
												on:keydown={(e) => onCellKeydown(e, i, d)}
											/>
										</td>
									{/each}
								{:else}
									{#each days as d}
										<td class="p-0">
											<input
												type="number"
												min="0"
												disabled={!row.product_id}
												data-cell={row.product_id ? cellKey(row.product_id, d) : undefined}
												class="w-14 border-0 bg-transparent text-center text-sm focus:bg-primary/10 focus:outline-none disabled:opacity-30"
												class:pending-cell={row.product_id && deltaOf(row.product_id, d) !== 0}
												value={row.product_id ? (cells[cellKey(row.product_id, d)]?.[direction] ?? '') : ''}
												on:input={(e) => row.product_id && onInput(row.product_id, d, e.currentTarget.value)}
												on:paste={(e) => row.product_id && onPaste(e, row.product_id, d)}
												on:keydown={(e) => onCellKeydown(e, i, d)}
											/>
										</td>
									{/each}
								{/if}

								<td class="text-center text-xs font-medium text-base-content/60">
									{row.product_id ? rowTotal(row.product_id) || '' : ''}
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
							{#if viewMode === 'compact'}
								{#each populatedDayCols as d}
									<th class="bg-base-300 text-center text-xs font-medium">
										{dayTotal(d) || ''}
									</th>
								{/each}
							{:else}
								{#each days as d}
									<th class="bg-base-300 text-center text-xs font-medium">
										{dayTotal(d) || ''}
									</th>
								{/each}
							{/if}
							<th class="bg-base-300 text-center text-xs">{grandTotal() || ''}</th>
							<th class="bg-base-300"></th>
						</tr>
					</tfoot>
				</table>
			</div>

			<div class="mt-4 flex items-center justify-between gap-3">
				<button class="btn btn-ghost btn-sm" on:click={addRow}>+ Tambah produk</button>
				<button class="btn btn-primary btn-md" on:click={save} disabled={isSaving}>
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
{/if}

<style>
	.pending-cell {
		background-color: color-mix(in srgb, var(--fallback-wa, oklch(var(--wa))) 25%, transparent);
	}
</style>
