-- ============================================================
-- Migration 000016: Product Change Request (approval master produk)
-- Model ADDITIVE: endpoint lama tetap jalan. Permintaan perubahan
-- baru masuk ke tabel ini; approve/reject menulis ke products.
-- ============================================================

CREATE TABLE IF NOT EXISTS product_change_requests (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- create | update | delete
    change_type varchar(20) NOT NULL,
    -- NULL untuk create; product yang diubah/dihapus untuk update/delete
    product_id uuid REFERENCES products(id),
    -- Payload produk (JSON) — lengkap untuk create, patch untuk update.
    payload jsonb NOT NULL,
    reason text,
    status varchar(20) NOT NULL DEFAULT 'pending', -- pending | approved | rejected
    requested_by uuid NOT NULL REFERENCES users(id),
    reviewed_by uuid REFERENCES users(id),
    reviewed_at timestamptz,
    review_note text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pcr_status ON product_change_requests(status);
CREATE INDEX IF NOT EXISTS idx_pcr_requested ON product_change_requests(requested_by);
CREATE INDEX IF NOT EXISTS idx_pcr_created ON product_change_requests(created_at DESC);

-- ============================================================
-- DOWN
-- ============================================================
-- DROP TABLE IF EXISTS product_change_requests;
