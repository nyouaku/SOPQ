-- ===== PHASE A: core opname cycle =====

-- A1: per-day session code counter (SO-YYYYMMDD-NNNN)
CREATE TABLE IF NOT EXISTS stock_opname_code_seq (
    day        DATE PRIMARY KEY,
    last_value INT  NOT NULL DEFAULT 0
);

-- A4: constrain input_method to the set the API and both clients actually send
ALTER TABLE stock_opname_detail
    ADD CONSTRAINT chk_so_detail_input_method
    CHECK (input_method IN ('manual', 'voice', 'camera', 'upload'));
