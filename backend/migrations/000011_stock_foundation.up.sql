-- ============================================================
-- Migration 000011: Stock Foundation
-- Stock Ledger, Opname Snapshot, Blind Count, Session Lifecycle
-- ============================================================

-- 1. STOCK LEDGER — sumber kebenaran riwayat pergerakan stok
CREATE TABLE IF NOT EXISTS stock_ledger (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id uuid NOT NULL REFERENCES products(id),
    warehouse_id uuid NOT NULL REFERENCES warehouses(id),
    transaction_type varchar(30) NOT NULL, -- initial, purchase_in, sale_out, transfer_in, transfer_out, adjustment, opname_correction, return_in, return_out
    reference_type varchar(50),             -- stock_opname, manual, import, dll
    reference_id uuid,                      -- ID dokumen sumber
    reference_code varchar(50),             -- kode sesi/dokumen (mis. SO-20261116-0001)
    quantity_in int NOT NULL DEFAULT 0,
    quantity_out int NOT NULL DEFAULT 0,
    balance_after int NOT NULL,             -- saldo setelah transaksi
    unit_cost numeric(14,2) DEFAULT 0,     -- harga satuan saat transaksi
    total_cost numeric(14,2) DEFAULT 0,    -- qty × unit_cost
    note text,
    created_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_stock_ledger_product ON stock_ledger(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_warehouse ON stock_ledger(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_created ON stock_ledger(created_at);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_product_warehouse ON stock_ledger(product_id, warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_reference ON stock_ledger(reference_type, reference_id);

-- 2. OPNAME SNAPSHOT + CUT-OFF
-- Simpan snapshot stok sistem saat sesi dimulai, dan waktu cut-off.
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    snapshot_taken_at timestamptz;
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    cut_off_at timestamptz;
-- Status is_frozen: saat true, scan baru tidak diterima (stok beku)
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    is_frozen boolean NOT NULL DEFAULT false;

-- 3. BLIND COUNT MODE
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    is_blind_count boolean NOT NULL DEFAULT false;

-- 4. SESSION STATUS LIFECYCLE
-- Tambah kolom untuk lifecycle baru
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    rejection_reason text;
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    recount_count int NOT NULL DEFAULT 0;
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    max_recount int NOT NULL DEFAULT 2;
-- approved_at yang sudah ada tetap; tambahkan rejected_at
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    rejected_at timestamptz;
ALTER TABLE stock_opname ADD COLUMN IF NOT EXISTS
    rejected_by uuid REFERENCES users(id);

-- 5. STOCK ADJUSTMENT JOURNAL (formal, terpisah dari activity_log)
CREATE TABLE IF NOT EXISTS stock_adjustment_journal (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id uuid NOT NULL REFERENCES products(id),
    warehouse_id uuid NOT NULL REFERENCES warehouses(id),
    session_id uuid REFERENCES stock_opname(id),
    reason varchar(30) NOT NULL, -- damaged, lost, expired, found, typo_fix, other
    note text,
    quantity_before int NOT NULL,
    quantity_after int NOT NULL,
    delta int NOT NULL,
    adjusted_by uuid NOT NULL REFERENCES users(id),
    approved_by uuid REFERENCES users(id),
    approved_at timestamptz,
    ip_address varchar(45),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_adj_journal_product ON stock_adjustment_journal(product_id);
CREATE INDEX IF NOT EXISTS idx_adj_journal_warehouse ON stock_adjustment_journal(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_adj_journal_created ON stock_adjustment_journal(created_at);

-- ============================================================
-- DOWN MIGRATION
-- ============================================================
