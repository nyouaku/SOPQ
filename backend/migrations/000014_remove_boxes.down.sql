-- ============================================================
-- Migration 000014 DOWN: kembalikan hierarki warehouse → location → box.
--
-- CATATAN: data box asli sudah hilang saat up. Down migration ini hanya
-- memulihkan struktur skema. Box baru dibuat kosong; produk & detail
-- kembali memakai box_id = NULL.
-- ============================================================

CREATE TABLE IF NOT EXISTS boxes (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id uuid NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    code varchar(30) NOT NULL,
    name varchar(150),
    created_at timestamptz DEFAULT now(),
    UNIQUE(location_id, code)
);

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS box_id uuid REFERENCES boxes(id);

ALTER TABLE stock_opname_detail
    ADD COLUMN IF NOT EXISTS box_id uuid REFERENCES boxes(id);

ALTER TABLE products
    DROP COLUMN IF EXISTS location_id;

ALTER TABLE stock_opname_detail
    DROP COLUMN IF EXISTS location_id;
