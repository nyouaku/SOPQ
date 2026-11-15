-- Add layout_json column to warehouses for visual editor
ALTER TABLE warehouses ADD COLUMN IF NOT EXISTS layout_json JSONB DEFAULT NULL;
