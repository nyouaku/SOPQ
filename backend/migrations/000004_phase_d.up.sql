-- ===== PHASE D =====

-- 3.1: laporan & detail sesi selalu mengurutkan scanned_at DESC per sesi.
-- Tanpa indeks ini, Postgres menyortir ulang seluruh baris sesi tiap request.
CREATE INDEX IF NOT EXISTS idx_so_detail_opname_scanned
    ON stock_opname_detail (stock_opname_id, scanned_at DESC);
