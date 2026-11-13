-- Performance Optimization Indexes
-- Run this to speed up queries significantly

-- Stock Opname Indexes
CREATE INDEX IF NOT EXISTS idx_stock_opname_status ON stock_opname(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stock_opname_started_by ON stock_opname(started_by);
CREATE INDEX IF NOT EXISTS idx_stock_opname_warehouse ON stock_opname(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_opname_created_at ON stock_opname(created_at DESC);

-- Stock Opname Detail Indexes
CREATE INDEX IF NOT EXISTS idx_stock_opname_detail_session ON stock_opname_detail(stock_opname_id);
CREATE INDEX IF NOT EXISTS idx_stock_opname_detail_product ON stock_opname_detail(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_opname_detail_scanned_by ON stock_opname_detail(scanned_by);
CREATE INDEX IF NOT EXISTS idx_stock_opname_detail_created_at ON stock_opname_detail(created_at DESC);

-- Products Indexes (additional)
CREATE INDEX IF NOT EXISTS idx_products_category_status ON products(category_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_products_name_trgm ON products USING gin(name gin_trgm_ops); -- Full-text search
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand) WHERE deleted_at IS NULL;

-- Users Indexes (additional)
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active) WHERE deleted_at IS NULL AND is_active = true;

-- Activity Logs Indexes
CREATE INDEX IF NOT EXISTS idx_activity_logs_user ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_entity ON activity_logs(entity_type, entity_id);

-- Stock Indexes
CREATE INDEX IF NOT EXISTS idx_stock_product ON stock(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_warehouse ON stock(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_updated_at ON stock(updated_at DESC);

-- Sessions Indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at) WHERE deleted_at IS NULL;

-- Enable pg_trgm extension for full-text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Analyze tables to update statistics
ANALYZE products;
ANALYZE stock_opname;
ANALYZE stock_opname_detail;
ANALYZE users;
ANALYZE stock;

-- Vacuum to reclaim space and update stats
VACUUM ANALYZE products;
VACUUM ANALYZE stock_opname;
VACUUM ANALYZE stock_opname_detail;

-- Show index sizes
SELECT
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;
