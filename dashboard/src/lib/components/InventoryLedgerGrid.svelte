<script lang="ts">
	/**
	 * InventoryLedgerGrid — Grid Harian model EXAMPLE_INVENTORY.xlsx.
	 * Baris = produk. Kolom: identitas, SALDO AWAL, HARGA MODAL, rekap
	 * PEMBELIAN/PENGELUARAN/RETUR/SALDO AKHIR, lalu M/K per tanggal (1-31).
	 *
	 * Sumber saldo: GET /stock-ledger/summary (derived dari stock_ledger).
	 * Input M/K disimpan lewat POST /inventory/grid (sudah transaksional).
	 */
	import { onMount } from 'svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	export let warehouseId = '';
	export let month = ''; // YYYY-MM

	type SummaryRow = {
		product_id: string;
		product_code: string;
		product_name: string;
		unit: string;
		cost_price: number;
		opening: number;
		total_in: number;
		total_out: number;
		closing: number;
		value_in: number;
		value_out: number;
	};
	type Cell = { m: number; k: number; r: number };

	let loading = true;
	let saving = false;
	let rows: SummaryRow[] = [];
	// cells[productId][day] = {m,k}; day = 1..31
	let cells: Record<string, Record<number, Cell>> = {};
	let dirty = false;

	function monthRange(m: string): { from: string; to: string; days: number } {
		const [y, mo] = m.split('-').map(Number);
		const days = new Date(y, mo, 0).getDate();
		const from = `${m}-01`;
		const toY = mo === 12 ? y + 1 : y;
		const toMo = mo === 12 ? 1 : mo + 1;
		const to = `${toY}-${String(toMo).padStart(2, '0')}-01`;
		return { from, to, days };
	}

	async function load() {
		if (!warehouseId || !month) return;
		loading = true;
		const { from, to } = monthRange(month);
		try {
			const res = await apiClient.get<{ rows: SummaryRow[] }>(
				`${API_ENDPOINTS.stockLedgerSummary}?warehouse_id=${warehouseId}&from=${from}&to=${to}`
			);
			rows = res.success && res.data ? res.data.rows || [] : [];

			// Pre-fill isian grid dari transaksi yang sudah tersimpan (M & K terpisah).
			const [inRes, outRes] = await Promise.all([
				apiClient.get<{ cells: any[] }>(
					`${API_ENDPOINTS.inventoryGrid}?warehouse_id=${warehouseId}&month=${month}&movement_type=in`
				),
				apiClient.get<{ cells: any[] }>(
					`${API_ENDPOINTS.inventoryGrid}?warehouse_id=${warehouseId}&month=${month}&movement_type=out`
				)
			]);
			const next: Record<string, Record<number, Cell>> = {};
			for (const c of (inRes.data?.cells || [])) {
				(next[c.product_id] ||= {})[c.day] ||= { m: 0, k: 0, r: 0 };
				next[c.product_id][c.day].m = c.in_qty || 0;
			}
			for (const c of (outRes.data?.cells || [])) {
				(next[c.product_id] ||= {})[c.day] ||= { m: 0, k: 0, r: 0 };
				next[c.product_id][c.day].k = c.out_qty || 0;
			}
			cells = next;
			dirty = false;
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal memuat ledger');
			rows = [];
		} finally {
			loading = false;
		}
	}

	function cellFor(pid: string, day: number): Cell {
		return cells[pid]?.[day] || { m: 0, k: 0, r: 0 };
	}

	function setCell(pid: string, day: number, field: 'm' | 'k' | 'r', raw: string) {
		const n = Math.max(0, parseInt(raw, 10) || 0);
		(cells[pid] ||= {})[day] ||= { m: 0, k: 0, r: 0 };
		cells[pid][day][field] = n;
		dirty = true;
		cells = cells; // trigger reactivity
	}

	function fmt(n: number): string {
		return new Intl.NumberFormat('id-ID').format(n || 0);
	}
	function fmtRp(n: number): string {
		return 'Rp ' + new Intl.NumberFormat('id-ID').format(Math.round(n || 0));
	}

	$: days = month ? Array.from({ length: monthRange(month).days }, (_, i) => i + 1) : [];

	// Rekap dari isian grid (M/K/R) — bukan dari saldo live.
	function totalsFor(r: SummaryRow) {
		let m = 0, k = 0, ret = 0;
		for (const d of Object.values(cells[r.product_id] || {})) {
			m += d.m; k += d.k; ret += d.r || 0;
		}
		const cost = r.cost_price || 0;
		return {
			m, k, retur: ret, valueIn: m * cost, valueOut: k * cost,
			closing: r.opening + m - k + ret
		};
	}

	async function save() {
		if (!dirty) {
			toastStore.info('Tidak ada perubahan');
			return;
		}
		saving = true;
		try {
			const inEntries: any[] = [];
			const outEntries: any[] = [];
			const retEntries: any[] = [];
			for (const r of rows) {
				for (const [dayStr, c] of Object.entries(cells[r.product_id] || {})) {
					const day = Number(dayStr);
					const date = `${month}-${String(day).padStart(2, '0')}`;
					const cost = r.cost_price || 0;
					if (c.m > 0) inEntries.push({ product_id: r.product_id, date, quantity: c.m, unit_cost: cost });
					if (c.k > 0) outEntries.push({ product_id: r.product_id, date, quantity: c.k, unit_cost: cost });
					if ((c.r || 0) > 0) retEntries.push({ product_id: r.product_id, date, quantity: c.r, unit_cost: cost });
				}
			}
			const post = (entries: any[], type: 'in' | 'out', sub: string) =>
				apiClient.post(API_ENDPOINTS.inventoryGrid, {
					warehouse_id: warehouseId, movement_type: type, sub_type: sub, entries
				});
			if (inEntries.length) {
				const res = await post(inEntries, 'in', 'purchase');
				if (!res.success) throw new Error(res.error?.message || 'Gagal simpan masuk');
			}
			if (outEntries.length) {
				const res = await post(outEntries, 'out', 'sale');
				if (!res.success) throw new Error(res.error?.message || 'Gagal simpan keluar');
			}
			if (retEntries.length) {
				const res = await post(retEntries, 'in', 'return');
				if (!res.success) throw new Error(res.error?.message || 'Gagal simpan retur');
			}
			toastStore.success('Grid tersimpan');
			await load();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal menyimpan grid');
		} finally {
			saving = false;
		}
	}

	onMount(load);
</script>

<div class="spreadsheet-root">
	<div class="mb-3 flex flex-wrap items-center justify-between gap-2">
		<div>
			<h2 class="text-lg font-semibold">Inventory Stock Ledger — {month}</h2>
			<p class="text-xs text-base-content/60">
				Sumber: stock ledger. Saldo awal = saldo akhir bulan sebelumnya.
			</p>
		</div>
		<div class="flex gap-2">
			<button class="btn btn-ghost btn-sm" on:click={load} disabled={saving}>↻ Muat ulang</button>
			<button class="btn btn-primary btn-sm" on:click={save} disabled={saving || !dirty}>
				{saving ? 'Menyimpan...' : 'Simpan Grid'}
			</button>
		</div>
	</div>

	{#if loading}
		<div class="flex h-40 items-center justify-center"><span class="loading loading-spinner"></span></div>
	{:else if rows.length === 0}
		<div class="rounded-2xl border border-dashed border-base-300 py-12 text-center text-sm text-base-content/60">
			Belum ada produk dengan mutasi di gudang ini. Tambah produk / stok awal dulu.
		</div>
	{:else}
		<div class="overflow-auto rounded-2xl border border-base-300 bg-base-200">
			<table class="spreadsheet-table text-xs">
				<thead class="sticky top-0 z-10 bg-base-300">
					<tr>
						<th class="sticky left-0 z-20 bg-base-300">NO</th>
						<th class="sticky left-8 z-20 bg-base-300">KODE</th>
						<th class="sticky left-32 z-20 bg-base-300">BARANG</th>
						<th>SAT</th>
						<th class="text-right">SALDO AWAL</th>
						<th class="text-right">HARGA MODAL</th>
						<th class="text-right">PEMBELIAN</th>
						<th class="text-right">PENGELUARAN</th>
						<th class="text-right">RETUR</th>
						<th class="text-right">SALDO AKHIR</th>
						{#each days as d}
							<th class="text-center" colspan="2">TGL {d}</th>
						{/each}
					</tr>
					<tr class="bg-base-300/70">
						<th class="sticky left-0 z-20 bg-base-300/70"></th>
						<th class="sticky left-8 z-20 bg-base-300/70"></th>
						<th class="sticky left-32 z-20 bg-base-300/70"></th>
						<th></th><th></th><th></th><th></th><th></th><th></th><th></th>
						{#each days as _}
							<th class="text-center">M</th><th class="text-center">K</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					{#each rows as r, i}
						{@const t = totalsFor(r)}
						<tr>
							<td class="sticky left-0 z-10 bg-base-200 text-base-content/50">{i + 1}</td>
							<td class="sticky left-8 z-10 bg-base-200 metric">{r.product_code}</td>
							<td class="sticky left-32 z-10 bg-base-200 min-w-[12rem] truncate">{r.product_name}</td>
							<td>{r.unit || '-'}</td>
							<td class="text-right metric">{fmt(r.opening)}</td>
							<td class="text-right metric text-base-content/60">{fmtRp(r.cost_price || 0)}</td>
							<td class="text-right metric">{fmt(t.m)} <span class="text-base-content/40">({fmtRp(t.valueIn)})</span></td>
							<td class="text-right metric">{fmt(t.k)} <span class="text-base-content/40">({fmtRp(t.valueOut)})</span></td>
							<td class="text-right metric text-base-content/50">{t.retur}</td>
							<td class="text-right metric font-semibold {t.closing < 0 ? 'text-error' : 'text-success'}">{fmt(t.closing)}</td>
							{#each days as d}
								{@const c = cellFor(r.product_id, d)}
								<td class="p-0">
									<input
										class="grid-cell"
										type="number" min="0"
										value={c.m || ''}
										on:input={(e) => setCell(r.product_id, d, 'm', e.currentTarget.value)}
									/>
								</td>
								<td class="p-0">
									<input
										class="grid-cell grid-cell-out"
										type="number" min="0"
										value={c.k || ''}
										on:input={(e) => setCell(r.product_id, d, 'k', e.currentTarget.value)}
									/>
								</td>
							{/each}
							<!-- RETUR per produk (bukan per tanggal), satu input ringkas -->
							<td class="p-0">
								<input
									class="grid-cell grid-cell-ret"
									type="number" min="0"
									value={totalsFor(r).retur || ''}
									title="Retur produk (qty)"
									on:input={(e) => {
										const v = Math.max(0, parseInt(e.currentTarget.value, 10) || 0);
										(cells[r.product_id] ||= {})[1] ||= { m: 0, k: 0, r: 0 };
										cells[r.product_id][1].r = v;
										dirty = true;
										cells = cells;
									}}
								/>
							</td>
						</tr>
					{/each}
				</tbody>
				<tfoot class="sticky bottom-0 bg-base-300 font-semibold">
					<tr>
						<td colspan="6">TOTAL</td>
						<td class="text-right metric">{fmt(rows.reduce((s, r) => s + totalsFor(r).m, 0))}</td>
						<td class="text-right metric">{fmt(rows.reduce((s, r) => s + totalsFor(r).k, 0))}</td>
						<td class="text-right metric">{fmt(rows.reduce((s, r) => s + totalsFor(r).retur, 0))}</td>
						<td class="text-right metric">{fmt(rows.reduce((s, r) => s + totalsFor(r).closing, 0))}</td>
						<td colspan={days.length * 2 + 1}></td>
					</tr>
				</tfoot>
			</table>
		</div>
	{/if}
</div>

<style>
	.spreadsheet-table { border-collapse: separate; border-spacing: 0; }
	.spreadsheet-table th, .spreadsheet-table td {
		white-space: nowrap; padding: 0.3rem 0.5rem; border-bottom: 1px solid oklch(var(--bc) / 0.08);
	}
	.spreadsheet-table tbody tr:hover td { background: oklch(var(--bc) / 0.04); }
	.grid-cell {
		width: 3.5rem; border: 0; background: transparent; color: inherit;
		font-family: 'JetBrains Mono', monospace; text-align: center; outline: none;
	}
	.grid-cell:focus { background: oklch(var(--p) / 0.12); box-shadow: inset 0 0 0 1px oklch(var(--p)); }
	.grid-cell-out:focus { background: oklch(var(--er) / 0.12); box-shadow: inset 0 0 0 1px oklch(var(--er)); }
	.grid-cell-ret:focus { background: oklch(var(--warning) / 0.12); box-shadow: inset 0 0 0 1px oklch(var(--warning)); }
	input::-webkit-outer-spin-button, input::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
</style>
