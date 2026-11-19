<script lang="ts">
	/**
	 * QuickEntry — form per-transaksi: pilih produk, tanggal, qty. Baris dikumpulkan
	 * lalu dikirim sekaligus ke PostGrid. Cocok untuk input 1–5 item per hari.
	 * Untuk input massal gunakan SpreadsheetEntry (clone Excel).
	 */
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { localMonth } from '$lib/utils/date';

	type Row = { product_id: string; name: string; product_code: string };

	/** Gudang, arah pergerakan (in/out), dan bulan dikontrol parent agar toolbar tersatukan. */
	export let warehouseId = '';
	export let direction: 'out' | 'in' = 'out';
	export let month = localMonth();

	let isSaving = false;
	let searchResults: any[] = [];
	let activeRow: Row | null = null;

	type QuickLine = { product_id: string; name: string; product_code: string; day: number; qty: number };
	let quickLines: QuickLine[] = [];
	let quickProduct: Row | null = null;
	let quickProductName = '';
	let quickDay = new Date().getDate();
	let quickQty = 1;
	let quickUnitCost = 0;

	$: daysInMonth = new Date(parseInt(month.slice(0, 4)), parseInt(month.slice(5, 7)), 0).getDate();

	// ---- Autocomplete produk
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	function onQuickNameInput() {
		if (searchTimer) clearTimeout(searchTimer);
		const q = quickProductName;
		if (!q || q.length < 2) {
			searchResults = [];
			return;
		}
		activeRow = { product_id: '', name: q, product_code: '' };
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

	function pickQuickProduct(p: any) {
		quickProduct = { product_id: p.id, name: p.name, product_code: p.product_code };
		quickProductName = p.name;
		searchResults = [];
		activeRow = null;
	}

	function addQuickLine() {
		if (!quickProduct) {
			toastStore.error('Pilih produk terlebih dahulu');
			return;
		}
		if (quickQty <= 0) {
			toastStore.error('Jumlah harus lebih dari 0');
			return;
		}
		if (quickDay < 1 || quickDay > daysInMonth) {
			toastStore.error(`Tanggal harus antara 1 dan ${daysInMonth}`);
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
		quickProduct = null;
		quickProductName = '';
		quickQty = 1;
	}

	function removeQuickLine(i: number) {
		quickLines = quickLines.filter((_, idx) => idx !== i);
	}

	async function save() {
		if (quickLines.length === 0) {
			toastStore.info('Belum ada baris untuk disimpan');
			return;
		}

		const entries = quickLines.map((l) => ({
			product_id: l.product_id,
			date: `${month}-${String(l.day).padStart(2, '0')}`,
			quantity: l.qty,
			unit_cost: quickUnitCost || 0
		}));

		isSaving = true;
		try {
			const res = await apiClient.post<{ movements_created: number; items_posted: number }>(
				API_ENDPOINTS.inventoryGrid,
				{
					warehouse_id: warehouseId,
					movement_type: direction,
					sub_type: direction === 'in' ? 'purchase' : 'sale',
					entries
				}
			);
			if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyimpan');

			toastStore.success(
				`${res.data.items_posted} isian tersimpan (${res.data.movements_created} dokumen)`
			);
			quickLines = [];
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal menyimpan');
		} finally {
			isSaving = false;
		}
	}
</script>

<!-- Form tambah pergerakan -->
<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
	<h2 class="mb-3 font-semibold">Tambah pergerakan {direction === 'in' ? 'Masuk' : 'Keluar'}</h2>

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

		<div>
			<label class="label py-0" for="q-cost">
				<span class="label-text text-xs text-base-content/60">Harga satuan (opsional)</span>
			</label>
			<input id="q-cost" type="number" min="0"
				class="input input-bordered input-sm mt-1 w-full"
				bind:value={quickUnitCost} />
		</div>
	</div>

	<button class="btn btn-primary btn-sm mt-3" on:click={addQuickLine}>
		+ Tambah ke daftar
	</button>

	{#if quickLines.length > 0}
		<div class="mt-4 overflow-x-auto">
			<table class="table table-sm">
				<thead>
					<tr>
						<th>Produk</th>
						<th>Tanggal</th>
						<th>Jumlah</th>
						<th>Harga</th>
						<th>Subtotal</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each quickLines as line, i}
						<tr>
							<td class="text-sm">{line.name}</td>
							<td class="text-xs text-base-content/60">Tgl {line.day}</td>
							<td class="metric font-semibold">{line.qty}</td>
							<td class="text-xs">{quickUnitCost || 0}</td>
							<td class="text-xs">{(quickUnitCost || 0) * line.qty}</td>
							<td>
								<button class="btn btn-ghost btn-xs text-error"
									on:click={() => removeQuickLine(i)}>✕</button>
							</td>
						</tr>
					{/each}
				</tbody>
				<tfoot>
					<tr>
						<td colspan="4" class="text-xs font-medium">Total</td>
						<td class="text-xs font-semibold">
							{quickLines.reduce((s, l) => s + (quickUnitCost || 0) * l.qty, 0)}
						</td>
						<td></td>
					</tr>
				</tfoot>
			</table>
		</div>

		<button class="btn btn-primary btn-md mt-3" on:click={save} disabled={isSaving}>
			{#if isSaving}
				<span class="loading loading-spinner loading-sm"></span>
				Menyimpan...
			{:else}
				Simpan {quickLines.length} baris
			{/if}
		</button>
	{/if}
</div>
