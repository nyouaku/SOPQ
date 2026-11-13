-- DEPRECATED: memakai kolom barcode/type/category_id yang sudah di-drop di 000006. Jangan dijalankan.
-- Seed sample products untuk testing dashboard

-- Insert sample categories
INSERT INTO categories (id, name) VALUES
    (gen_random_uuid(), 'Elektronik'),
    (gen_random_uuid(), 'Pakaian'),
    (gen_random_uuid(), 'Makanan'),
    (gen_random_uuid(), 'Alat Tulis')
ON CONFLICT DO NOTHING;

-- Insert sample suppliers
INSERT INTO suppliers (code, name, contact_person, phone, email) VALUES
    ('SUP001', 'PT Maju Jaya', 'Budi Santoso', '08123456789', 'budi@majujaya.com'),
    ('SUP002', 'CV Cahaya Abadi', 'Siti Rahayu', '08234567890', 'siti@cahaya.com'),
    ('SUP003', 'UD Berkah Selalu', 'Ahmad Zaini', '08345678901', 'ahmad@berkah.com')
ON CONFLICT (code) DO NOTHING;

-- Get warehouse and category IDs for reference
DO $$
DECLARE
    wh_id UUID;
    cat_elec UUID;
    cat_pak UUID;
    cat_mak UUID;
    cat_atk UUID;
    sup1 UUID;
    sup2 UUID;
    sup3 UUID;
BEGIN
    SELECT id INTO wh_id FROM warehouses LIMIT 1;
    SELECT id INTO cat_elec FROM categories WHERE name = 'Elektronik' LIMIT 1;
    SELECT id INTO cat_pak FROM categories WHERE name = 'Pakaian' LIMIT 1;
    SELECT id INTO cat_mak FROM categories WHERE name = 'Makanan' LIMIT 1;
    SELECT id INTO cat_atk FROM categories WHERE name = 'Alat Tulis' LIMIT 1;
    SELECT id INTO sup1 FROM suppliers WHERE code = 'SUP001' LIMIT 1;
    SELECT id INTO sup2 FROM suppliers WHERE code = 'SUP002' LIMIT 1;
    SELECT id INTO sup3 FROM suppliers WHERE code = 'SUP003' LIMIT 1;

    -- Insert sample products
    INSERT INTO products (barcode, product_code, name, brand, type, category_id, unit, minimum_stock, supplier_id, status) VALUES
        ('8991002001015', 'PRD001', 'Mouse Wireless Logitech M185', 'Logitech', 'Wireless', cat_elec, 'PCS', 10, sup1, 'active'),
        ('8991002001022', 'PRD002', 'Keyboard Mechanical RGB', 'Corsair', 'Gaming', cat_elec, 'PCS', 5, sup1, 'active'),
        ('8991002001039', 'PRD003', 'Webcam HD 1080p', 'Logitech', 'HD', cat_elec, 'PCS', 8, sup1, 'active'),
        ('8991002001046', 'PRD004', 'Headset Gaming RGB', 'Razer', 'Gaming', cat_elec, 'PCS', 6, sup1, 'active'),
        ('8991002001053', 'PRD005', 'USB Hub 4 Port', 'Anker', 'USB 3.0', cat_elec, 'PCS', 15, sup1, 'active'),

        ('8991003001012', 'PRD006', 'Kaos Polos Cotton Combed', 'Uniqlo', 'T-Shirt', cat_pak, 'PCS', 50, sup2, 'active'),
        ('8991003001029', 'PRD007', 'Celana Jeans Slim Fit', 'Levi''s', 'Jeans', cat_pak, 'PCS', 30, sup2, 'active'),
        ('8991003001036', 'PRD008', 'Jaket Hoodie Zipper', 'Nike', 'Hoodie', cat_pak, 'PCS', 20, sup2, 'active'),
        ('8991003001043', 'PRD009', 'Topi Baseball Cap', 'Adidas', 'Cap', cat_pak, 'PCS', 40, sup2, 'active'),
        ('8991003001050', 'PRD010', 'Sepatu Sneakers', 'Converse', 'Sneakers', cat_pak, 'PCS', 25, sup2, 'active'),

        ('8991004001019', 'PRD011', 'Mie Instan Goreng', 'Indomie', 'Instant', cat_mak, 'PCS', 100, sup3, 'active'),
        ('8991004001026', 'PRD012', 'Kopi Sachet 3in1', 'Kapal Api', 'Instant', cat_mak, 'BOX', 50, sup3, 'active'),
        ('8991004001033', 'PRD013', 'Biskuit Kaleng', 'Khong Guan', 'Snack', cat_mak, 'PCS', 30, sup3, 'active'),
        ('8991004001040', 'PRD014', 'Air Mineral Botol 600ml', 'Aqua', 'Beverage', cat_mak, 'PCS', 200, sup3, 'active'),
        ('8991004001057', 'PRD015', 'Teh Celup Kotak', 'Sariwangi', 'Tea', cat_mak, 'BOX', 40, sup3, 'active'),

        ('8991005001016', 'PRD016', 'Pulpen Gel Hitam', 'Pilot', 'Pen', cat_atk, 'PCS', 100, sup3, 'active'),
        ('8991005001023', 'PRD017', 'Pensil 2B', 'Faber Castell', 'Pencil', cat_atk, 'PCS', 150, sup3, 'active'),
        ('8991005001030', 'PRD018', 'Penghapus Putih', 'Steadtler', 'Eraser', cat_atk, 'PCS', 80, sup3, 'active'),
        ('8991005001047', 'PRD019', 'Buku Tulis 38 Lembar', 'Sinar Dunia', 'Book', cat_atk, 'PCS', 200, sup3, 'active'),
        ('8991005001054', 'PRD020', 'Spidol Whiteboard Hitam', 'Snowman', 'Marker', cat_atk, 'PCS', 60, sup3, 'active')
    ON CONFLICT (barcode) DO NOTHING;

    -- Insert stock data (random stock untuk testing)
    IF wh_id IS NOT NULL THEN
        INSERT INTO stock (product_id, warehouse_id, quantity_system, version)
        SELECT id, wh_id, floor(random() * 100 + 10)::int, 0
        FROM products
        ON CONFLICT (product_id, warehouse_id) DO NOTHING;
    END IF;
END $$;
