DROP TABLE IF EXISTS stock_adjustment_journal;
DROP TABLE IF EXISTS stock_ledger;

ALTER TABLE stock_opname DROP COLUMN IF EXISTS rejected_by;
ALTER TABLE stock_opname DROP COLUMN IF EXISTS rejected_at;
ALTER TABLE stock_opname DROP COLUMN IF EXISTS max_recount;
ALTER TABLE stock_opname DROP COLUMN IF EXISTS recount_count;
ALTER TABLE stock_opname DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE stock_opname DROP COLUMN IF EXISTS is_blind_count;
ALTER TABLE stock_opname DROP COLUMN IF EXISTS is_frozen;
ALTER TABLE stock_opname DROP COLUMN IF EXISTS cut_off_at;
ALTER TABLE stock_opname DROP COLUMN IF EXISTS snapshot_taken_at;
