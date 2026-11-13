-- ===== PHASE B =====

-- 1.3: lookup refresh token per request; hanya sesi aktif yang diindeks
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_hash
    ON sessions (refresh_token_hash)
    WHERE revoked_at IS NULL;

-- 1.9: idempotency scan. Klien mengirim UUID per scan; retry jaringan
-- mengembalikan baris yang sama, bukan baris baru.
ALTER TABLE stock_opname_detail
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(64);

CREATE UNIQUE INDEX IF NOT EXISTS uq_so_detail_idempotency
    ON stock_opname_detail (stock_opname_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
