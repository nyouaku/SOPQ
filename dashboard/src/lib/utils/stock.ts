/**
 * Sumber tunggal angka stok untuk UI.
 *
 * Backend menyimpan stok di tabel `stock` (per produk + gudang, kolom
 * `quantity_system`). Beberapa endpoint mengirim bentuk berbeda:
 *  - GET /products            → `stock_system` (nilai ringkasan)
 *  - GET /products/:id        → `stock` array bila ada
 *  - detail opname            → `stock_system` = snapshot saat scan
 *
 * Semua pembacaan UI lewat helper ini agar tidak ada lagi empat definisi
 * berbeda. Bila backend menambah field, cukup ubah di sini.
 */

export interface StockLike {
	stock_system?: number | null;
	stock?: { quantity_system?: number | null }[] | null;
}

/** Angka stok sistem produk; 0 bila belum ada data stok. */
export function productStock(product: StockLike | null | undefined): number {
	if (!product) return 0;

	if (typeof product.stock_system === 'number') return product.stock_system;

	if (Array.isArray(product.stock) && product.stock.length) {
		return product.stock.reduce((sum, row) => sum + (row.quantity_system || 0), 0);
	}

	return 0;
}

/** Stok per gudang, bila endpoint mengirim rinciannya. */
export function stockByWarehouse(product: StockLike | null | undefined): Record<string, number> {
	const out: Record<string, number> = {};
	const rows = (product?.stock || []) as any[];
	for (const row of rows) {
		if (row?.warehouse_id) out[row.warehouse_id] = row.quantity_system || 0;
	}
	return out;
}
