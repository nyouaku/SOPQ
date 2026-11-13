-- Seed default users untuk testing
-- Password untuk semua user: staff123

-- Delete existing test users
DELETE FROM users WHERE username IN ('admin', 'supervisor', 'staff');

-- Insert users dengan bcrypt hash untuk password "staff123"
INSERT INTO users (username, password_hash, full_name, role_id, is_active) VALUES
    ('admin', '$2a$12$VJTnRspFRCcxP6hotR7t4uwgwaPxrMbful7Ay0jF8/9YuwYCqWp8q', 'Administrator', 1, true),
    ('supervisor', '$2a$12$VJTnRspFRCcxP6hotR7t4uwgwaPxrMbful7Ay0jF8/9YuwYCqWp8q', 'Supervisor Gudang', 2, true),
    ('staff', '$2a$12$VJTnRspFRCcxP6hotR7t4uwgwaPxrMbful7Ay0jF8/9YuwYCqWp8q', 'Staff Gudang', 3, true);

-- Insert sample warehouse
INSERT INTO warehouses (code, name, address, is_active) VALUES
    ('WH01', 'Gudang Utama', 'Jl. Raya No. 123, Jakarta', true)
ON CONFLICT (code) DO NOTHING;
