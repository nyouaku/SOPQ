-- ===== PHASE D =====

-- 3.8: prefix index untuk pencarian barcode/product_code.
-- PostgreSQL B-tree menangani LIKE 'foo%' secara efisien (anchored left).
CREATE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_products_code ON products(product_code) WHERE deleted_at IS NULL;

-- idx_so_detail_opname_scanned sudah dibuat di 000004_phase_d.up.sql.
