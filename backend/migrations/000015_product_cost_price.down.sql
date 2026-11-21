-- ============================================================
-- Migration 000015 DOWN: hapus kolom harga modal produk.
-- CATATAN: data harga modal lama akan hilang.
-- ============================================================

ALTER TABLE products
    DROP COLUMN IF EXISTS cost_price;
