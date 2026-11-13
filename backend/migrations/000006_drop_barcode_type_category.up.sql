-- ===== DROP barcode, type, category_id dari products =====
-- Identitas produk sekarang product_code (juga yang di-encode QR).
-- Aman: produk produksi 0 baris saat migrasi ini dibuat.

DROP INDEX IF EXISTS idx_products_barcode;
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_category_status;

ALTER TABLE products DROP COLUMN IF EXISTS barcode;
ALTER TABLE products DROP COLUMN IF EXISTS type;
ALTER TABLE products DROP COLUMN IF EXISTS category_id;
