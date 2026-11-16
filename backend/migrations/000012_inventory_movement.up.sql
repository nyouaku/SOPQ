-- ============================================================
-- Migration 000012: Inventory Movement (Stock IN / OUT / Transfer)
-- Pencatatan pergerakan stok harian, terpisah dari Stock Opname.
-- stock_ledger & stock_adjustment_journal sudah ada (000011); modul ini
-- menjadi PENULIS PERTAMA stock_ledger.
-- ============================================================

-- Header dokumen pergerakan (bukti fisik tiap IN/OUT/transfer).
CREATE TABLE IF NOT EXISTS stock_movement (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    code varchar(40) UNIQUE NOT NULL,               -- MV-20261116-0001 (urut per hari)
    movement_type varchar(20) NOT NULL,             -- in | out | transfer
    sub_type varchar(30) NOT NULL,                  -- purchase, sale, return, damaged, lost, expired, other
    warehouse_id uuid NOT NULL REFERENCES warehouses(id),
    to_warehouse_id uuid REFERENCES warehouses(id), -- hanya untuk transfer antar gudang
    reference_code varchar(50),
    note text,
    status varchar(20) NOT NULL DEFAULT 'posted',   -- fase 1: langsung posted (tanpa approval)
    created_by uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_stock_movement_warehouse ON stock_movement(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_movement_created ON stock_movement(created_at);
CREATE INDEX IF NOT EXISTS idx_stock_movement_type ON stock_movement(movement_type);

-- Baris item per dokumen pergerakan.
CREATE TABLE IF NOT EXISTS stock_movement_item (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    movement_id uuid NOT NULL REFERENCES stock_movement(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products(id),
    quantity int NOT NULL CHECK (quantity > 0),
    unit_cost numeric(14,2) NOT NULL DEFAULT 0,
    note text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_stock_movement_item_movement ON stock_movement_item(movement_id);
CREATE INDEX IF NOT EXISTS idx_stock_movement_item_product ON stock_movement_item(product_id);

-- Counter kode dokumen per hari (pola sama dengan stock_opname_code_seq).
CREATE TABLE IF NOT EXISTS stock_movement_code_seq (
    day date PRIMARY KEY,
    last_value int NOT NULL DEFAULT 0
);

-- ============================================================
-- DOWN
-- ============================================================
