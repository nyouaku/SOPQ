-- Rollback: kolom dikembalikan kosong. Data lama tidak bisa dipulihkan dari sini.
ALTER TABLE products ADD COLUMN IF NOT EXISTS barcode VARCHAR(64);
ALTER TABLE products ADD COLUMN IF NOT EXISTS type VARCHAR(100);
ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES categories(id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
