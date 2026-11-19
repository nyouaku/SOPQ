-- ============================================================
-- Migration 000014: Hapus Rak/Box, sederhanakan jadi Gudang → Lokasi/Area
--
-- Sebelumnya hierarki 3 level: warehouse → location → box.
-- Sekarang 2 level: warehouse → location (area/rak).
-- Produk & detail opname kini menunjuk langsung ke location.
-- ============================================================

-- 1. Pindahkan produk ke lokasi langsung.
--    box_id → location_id lewat box.location_id.
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS location_id uuid REFERENCES locations(id);

UPDATE products p
SET location_id = b.location_id
FROM boxes b
WHERE p.box_id = b.id AND p.location_id IS NULL;

ALTER TABLE products
    ALTER COLUMN location_id DROP NOT NULL;

-- 2. Pindahkan detail opname ke lokasi langsung.
ALTER TABLE stock_opname_detail
    ADD COLUMN IF NOT EXISTS location_id uuid REFERENCES locations(id);

UPDATE stock_opname_detail d
SET location_id = b.location_id
FROM boxes b
WHERE d.box_id = b.id AND d.location_id IS NULL;

-- 3. Lepaskan foreign key & kolom box_id.
ALTER TABLE products
    DROP CONSTRAINT IF EXISTS products_box_id_fkey;
ALTER TABLE products
    DROP COLUMN IF EXISTS box_id;

ALTER TABLE stock_opname_detail
    DROP CONSTRAINT IF EXISTS stock_opname_detail_box_id_fkey;
ALTER TABLE stock_opname_detail
    DROP COLUMN IF EXISTS box_id;

-- 4. Hapus tabel boxes.
DROP TABLE IF EXISTS boxes;

-- ============================================================
-- DOWN
-- ============================================================
-- Rekonstruksi boxes dari location_id produk (nilai box hilang
-- permanen; ini best-effort untuk struktur skema).
