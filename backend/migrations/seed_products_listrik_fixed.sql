-- DEPRECATED: memakai kolom barcode/type/category_id yang sudah di-drop di 000006. Jangan dijalankan.
-- Dummy data produk listrik sesuai struktur tabel yang benar
-- Table structure: barcode, product_code, name, brand, type, category_id, unit, minimum_stock, status

-- Insert kategori jika belum ada
INSERT INTO categories (name, created_at, updated_at)
VALUES ('Elektronik & Listrik', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Insert produk listrik
INSERT INTO products (product_code, barcode, name, brand, type, category_id, unit, minimum_stock, status, created_at, updated_at)
VALUES
-- Kabel & Instalasi
('EL-001', '8991234567001', 'Kabel NYA 1.5mm 100m', 'Supreme', 'Kabel Instalasi', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', 10, 'active', NOW(), NOW()),
('EL-002', '8991234567002', 'Kabel NYM 2x1.5mm 100m', 'Supreme', 'Kabel Instalasi', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', 5, 'active', NOW(), NOW()),
('EL-003', '8991234567003', 'Kabel NYM 3x2.5mm 100m', 'Supreme', 'Kabel Instalasi', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', 5, 'active', NOW(), NOW()),

-- Stop Kontak & Saklar
('EL-004', '8991234567004', 'Stop Kontak Outbow', 'Broco', 'Saklar & Stop Kontak', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 50, 'active', NOW(), NOW()),
('EL-005', '8991234567005', 'Saklar Tunggal Outbow', 'Broco', 'Saklar & Stop Kontak', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 50, 'active', NOW(), NOW()),
('EL-006', '8991234567006', 'Saklar Ganda Outbow', 'Broco', 'Saklar & Stop Kontak', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 30, 'active', NOW(), NOW()),

-- Fitting & Lampu
('EL-007', '8991234567007', 'Fitting Lampu E27', 'Hanyoung', 'Fitting & Aksesoris', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 100, 'active', NOW(), NOW()),
('EL-008', '8991234567008', 'Lampu LED 9W Putih', 'Philips', 'Lampu LED', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 100, 'active', NOW(), NOW()),
('EL-009', '8991234567009', 'Lampu LED 12W Putih', 'Philips', 'Lampu LED', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 80, 'active', NOW(), NOW()),
('EL-010', '8991234567010', 'Lampu TL LED 18W', 'Philips', 'Lampu LED', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 50, 'active', NOW(), NOW()),

-- MCB & Instalasi Panel
('EL-011', '8991234567011', 'MCB 1P 6A', 'Schneider', 'MCB & Proteksi', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 20, 'active', NOW(), NOW()),
('EL-012', '8991234567012', 'MCB 1P 10A', 'Schneider', 'MCB & Proteksi', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 30, 'active', NOW(), NOW()),
('EL-013', '8991234567013', 'MCB 2P 16A', 'Schneider', 'MCB & Proteksi', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 20, 'active', NOW(), NOW()),

-- Pipa & Aksesoris
('EL-014', '8991234567014', 'Pipa PVC 3/4 inch 4m', 'Rucika', 'Pipa & Conduit', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'batang', 50, 'active', NOW(), NOW()),
('EL-015', '8991234567015', 'Pipa PVC 1 inch 4m', 'Rucika', 'Pipa & Conduit', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'batang', 50, 'active', NOW(), NOW()),
('EL-016', '8991234567016', 'Socket Pipa 3/4 inch', 'Rucika', 'Pipa & Conduit', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 100, 'active', NOW(), NOW()),

-- Steker & Colokan
('EL-017', '8991234567017', 'Steker Arde Bulat', 'Broco', 'Steker & Colokan', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 50, 'active', NOW(), NOW()),
('EL-018', '8991234567018', 'Steker Biasa 2 Pin', 'Broco', 'Steker & Colokan', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 80, 'active', NOW(), NOW()),

-- Terminal & Connector
('EL-019', '8991234567019', 'Terminal Block 10A', 'Phoenix', 'Terminal & Connector', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 30, 'active', NOW(), NOW()),
('EL-020', '8991234567020', 'Skun Kabel 1.5mm', 'Hanyoung', 'Terminal & Connector', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'box', 20, 'active', NOW(), NOW()),

-- Exhaust & Kipas
('EL-021', '8991234567021', 'Exhaust Fan 8 inch', 'Panasonic', 'Exhaust & Ventilasi', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 20, 'active', NOW(), NOW()),
('EL-022', '8991234567022', 'Kipas Angin Stand 16 inch', 'Miyako', 'Kipas Angin', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 10, 'active', NOW(), NOW()),

-- Meteran & Alat Ukur
('EL-023', '8991234567023', 'KWH Meter Digital 1P', 'Itron', 'Meteran Listrik', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 10, 'active', NOW(), NOW()),
('EL-024', '8991234567024', 'Tang Ampere Digital', 'Hioki', 'Alat Ukur', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 5, 'active', NOW(), NOW()),

-- Isolasi & Perlengkapan
('EL-025', '8991234567025', 'Isolasi PVC Hitam', '3M', 'Isolasi & Lakban', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', 100, 'active', NOW(), NOW()),
('EL-026', '8991234567026', 'Cable Tie 20cm', 'Kss', 'Cable Management', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pack', 50, 'active', NOW(), NOW()),
('EL-027', '8991234567027', 'Lakban Listrik Merah', '3M', 'Isolasi & Lakban', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', 50, 'active', NOW(), NOW()),

-- Extension & Multi Socket
('EL-028', '8991234567028', 'Stop Kontak 4 Lubang', 'Uticon', 'Extension & Multi Socket', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 30, 'active', NOW(), NOW()),
('EL-029', '8991234567029', 'Kabel Roll 15m + 4 Socket', 'Kenmaster', 'Extension & Multi Socket', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'unit', 10, 'active', NOW(), NOW()),

-- Stabilizer & UPS
('EL-030', '8991234567030', 'Stabilizer 500VA', 'Matsunaga', 'Stabilizer & UPS', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'unit', 10, 'active', NOW(), NOW())

ON CONFLICT (product_code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();

-- Verify
SELECT COUNT(*) as total_produk_listrik FROM products WHERE category_id = (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1);
