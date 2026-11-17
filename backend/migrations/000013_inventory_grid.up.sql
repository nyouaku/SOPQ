-- ============================================================
-- Migration 000013: Inventory Grid Entry (tanggal per item)
-- Grid ala Excel: tiap isian punya tanggal sendiri (per hari dalam sebulan).
-- ============================================================

-- movement_date: tanggal efektif transaksi (diisi saat posting dari grid).
-- NULL = dibuat lewat form biasa (created_at dipakai).
ALTER TABLE stock_movement_item
    ADD COLUMN IF NOT EXISTS movement_date date;

CREATE INDEX IF NOT EXISTS idx_stock_movement_item_date
    ON stock_movement_item(movement_date);

-- ============================================================
-- DOWN
-- ============================================================
