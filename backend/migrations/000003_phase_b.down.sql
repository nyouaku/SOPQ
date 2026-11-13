DROP INDEX IF EXISTS uq_so_detail_idempotency;
ALTER TABLE stock_opname_detail DROP COLUMN IF EXISTS idempotency_key;
DROP INDEX IF EXISTS idx_sessions_refresh_hash;
