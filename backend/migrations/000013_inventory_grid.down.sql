-- ============================================================
-- Migration 000013 DOWN: Inventory Grid Entry
-- ============================================================

DROP INDEX IF EXISTS idx_stock_movement_item_date;
ALTER TABLE stock_movement_item DROP COLUMN IF EXISTS movement_date;
