-- =====================================================================
-- STOCK OPNAME APP — MIGRATION UP
-- =====================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================
-- ROLES & USERS
-- ============================
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(150) NOT NULL,
    role_id INT NOT NULL REFERENCES roles(id),
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    device_id UUID,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_sessions_user ON sessions(user_id);

-- ============================
-- WAREHOUSE & LOCATION
-- ============================
CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(30) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    address TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(warehouse_id, code)
);

CREATE TABLE boxes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(150),
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(location_id, code)
);

-- ============================
-- MASTER DATA
-- ============================
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    parent_id UUID REFERENCES categories(id),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(30) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    contact_person VARCHAR(100),
    phone VARCHAR(30),
    email VARCHAR(150),
    address TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    barcode VARCHAR(64) UNIQUE NOT NULL,
    product_code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    brand VARCHAR(100),
    type VARCHAR(100),
    category_id UUID REFERENCES categories(id),
    color VARCHAR(50),
    size VARCHAR(50),
    unit VARCHAR(20) NOT NULL DEFAULT 'PCS',
    box_id UUID REFERENCES boxes(id),
    minimum_stock INT DEFAULT 0,
    supplier_id UUID REFERENCES suppliers(id),
    photo_url TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_products_barcode ON products(barcode) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_code ON products(product_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category ON products(category_id);

-- ============================
-- STOCK
-- ============================
CREATE TABLE stock (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    quantity_system INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(product_id, warehouse_id)
);
CREATE INDEX idx_stock_product_warehouse ON stock(product_id, warehouse_id);

-- ============================
-- STOCK OPNAME
-- ============================
CREATE TABLE stock_opname (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    location_id UUID REFERENCES locations(id),
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    started_by UUID NOT NULL REFERENCES users(id),
    started_at TIMESTAMPTZ DEFAULT now(),
    completed_at TIMESTAMPTZ,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_stock_opname_warehouse ON stock_opname(warehouse_id);
CREATE INDEX idx_stock_opname_status ON stock_opname(status);

CREATE TABLE stock_opname_detail (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stock_opname_id UUID NOT NULL REFERENCES stock_opname(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    box_id UUID REFERENCES boxes(id),
    stock_system INT NOT NULL,
    stock_physical INT NOT NULL,
    difference INT GENERATED ALWAYS AS (stock_physical - stock_system) STORED,
    input_method VARCHAR(10) NOT NULL DEFAULT 'manual',
    note TEXT,
    photo_url TEXT,
    scanned_by UUID NOT NULL REFERENCES users(id),
    device_id UUID,
    scanned_at TIMESTAMPTZ DEFAULT now(),
    sync_status VARCHAR(20) DEFAULT 'synced'
);
CREATE INDEX idx_so_detail_opname ON stock_opname_detail(stock_opname_id);
CREATE INDEX idx_so_detail_product ON stock_opname_detail(product_id);
CREATE INDEX idx_so_detail_scanned_by ON stock_opname_detail(scanned_by);

-- ============================
-- DEVICES
-- ============================
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    device_name VARCHAR(150),
    platform VARCHAR(30),
    app_version VARCHAR(20),
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- ============================
-- ACTIVITY LOG
-- ============================
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_activity_logs_user ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_entity ON activity_logs(entity_type, entity_id);
CREATE INDEX idx_activity_logs_created ON activity_logs(created_at);

-- ============================
-- NOTIFICATIONS
-- ============================
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(150),
    message TEXT,
    metadata JSONB,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);

-- ============================
-- SETTINGS
-- ============================
CREATE TABLE settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO settings (key, value, description) VALUES
    ('diff_warning_threshold_percent', '10', 'Persentase selisih untuk memicu warning'),
    ('jwt_access_token_ttl_minutes', '15', 'Masa berlaku access token'),
    ('jwt_refresh_token_ttl_days', '7', 'Masa berlaku refresh token');

-- ============================
-- SEED ROLES
-- ============================
INSERT INTO roles (name, description) VALUES
    ('admin', 'Akses penuh ke seluruh sistem'),
    ('supervisor', 'Memantau progress & approve hasil opname'),
    ('staff_gudang', 'Melakukan scan stock opname di lapangan');
