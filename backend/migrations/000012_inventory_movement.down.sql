-- ============================================================
-- Migration 000012 DOWN: Inventory Movement
-- ============================================================

DROP INDEX IF EXISTS idx_stock_movement_item_product;
DROP INDEX IF EXISTS idx_stock_movement_item_movement;
DROP TABLE IF EXISTS stock_movement_item;

DROP INDEX IF EXISTS idx_stock_movement_type;
DROP INDEX IF EXISTS idx_stock_movement_created;
DROP INDEX IF EXISTS idx_stock_movement_warehouse;
DROP TABLE IF EXISTS stock_movement;

DROP TABLE IF EXISTS stock_movement_code_seq;
