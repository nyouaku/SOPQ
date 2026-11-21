-- ============================================================
-- Migration 000015: Harga modal produk (cost_price)
-- Dipakai Grid Harian inventory ledger sebagai HARGA MODAL
-- dan untuk nilai pembelian/pengeluaran/retur (qty * cost_price).
-- ============================================================

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS cost_price numeric(14,2) NOT NULL DEFAULT 0;

-- ============================================================
-- DOWN
-- ============================================================
-- ALTER TABLE products DROP COLUMN IF EXISTS cost_price;
