-- DEPRECATED: memakai kolom barcode/type/category_id yang sudah di-drop di 000006. Jangan dijalankan.
-- Dummy data untuk produk listrik
-- Category: Elektronik & Listrik

-- Insert kategori jika belum ada
INSERT INTO categories (name, description, created_at, updated_at)
VALUES ('Elektronik & Listrik', 'Produk elektronik dan komponen listrik', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Insert produk listrik
INSERT INTO products (product_code, barcode, name, description, category_id, unit, brand, price, stock_system, min_stock, max_stock, is_active, created_at, updated_at)
VALUES
-- Kabel & Instalasi
('EL-001', '8991234567001', 'Kabel NYA 1.5mm 100m', 'Kabel instalasi rumah NYA 1.5mm roll 100 meter', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', 'Supreme', 450000, 50, 10, 200, true, NOW(), NOW()),
('EL-002', '8991234567002', 'Kabel NYM 2x1.5mm 100m', 'Kabel instalasi NYM 2x1.5mm roll 100 meter', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', 'Supreme', 650000, 30, 5, 100, true, NOW(), NOW()),
('EL-003', '8991234567003', 'Kabel NYM 3x2.5mm 100m', 'Kabel instalasi NYM 3x2.5mm roll 100 meter', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', 'Supreme', 1200000, 20, 5, 80, true, NOW(), NOW()),

-- Stop Kontak & Saklar
('EL-004', '8991234567004', 'Stop Kontak Outbow', 'Stop kontak outbow standar SNI', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Broco', 15000, 200, 50, 500, true, NOW(), NOW()),
('EL-005', '8991234567005', 'Saklar Tunggal Outbow', 'Saklar tunggal outbow standar SNI', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Broco', 12000, 250, 50, 500, true, NOW(), NOW()),
('EL-006', '8991234567006', 'Saklar Ganda Outbow', 'Saklar ganda outbow standar SNI', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Broco', 18000, 150, 30, 300, true, NOW(), NOW()),

-- Fitting & Lampu
('EL-007', '8991234567007', 'Fitting Lampu E27', 'Fitting lampu ulir E27 plastik', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Hanyoung', 5000, 300, 100, 1000, true, NOW(), NOW()),
('EL-008', '8991234567008', 'Lampu LED 9W Putih', 'Lampu LED bulb 9 watt cahaya putih', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Philips', 25000, 500, 100, 2000, true, NOW(), NOW()),
('EL-009', '8991234567009', 'Lampu LED 12W Putih', 'Lampu LED bulb 12 watt cahaya putih', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Philips', 35000, 400, 80, 1500, true, NOW(), NOW()),
('EL-010', '8991234567010', 'Lampu TL LED 18W', 'Lampu TL LED 18 watt 120cm', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Philips', 75000, 200, 50, 500, true, NOW(), NOW()),

-- MCB & Instalasi Panel
('EL-011', '8991234567011', 'MCB 1P 6A', 'MCB 1 phase 6 ampere SNI', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Schneider', 45000, 100, 20, 300, true, NOW(), NOW()),
('EL-012', '8991234567012', 'MCB 1P 10A', 'MCB 1 phase 10 ampere SNI', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Schneider', 48000, 150, 30, 400, true, NOW(), NOW()),
('EL-013', '8991234567013', 'MCB 2P 16A', 'MCB 2 phase 16 ampere SNI', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Schneider', 95000, 80, 20, 200, true, NOW(), NOW()),

-- Pipa & Aksesoris
('EL-014', '8991234567014', 'Pipa PVC 3/4 inch 4m', 'Pipa PVC conduit 3/4 inch panjang 4 meter', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'batang', 'Rucika', 35000, 300, 50, 1000, true, NOW(), NOW()),
('EL-015', '8991234567015', 'Pipa PVC 1 inch 4m', 'Pipa PVC conduit 1 inch panjang 4 meter', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'batang', 'Rucika', 45000, 250, 50, 800, true, NOW(), NOW()),
('EL-016', '8991234567016', 'Socket Pipa 3/4 inch', 'Socket sambungan pipa 3/4 inch', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Rucika', 3000, 500, 100, 2000, true, NOW(), NOW()),

-- Steker & Colokan
('EL-017', '8991234567017', 'Steker Arde Bulat', 'Steker arde bulat 3 pin standar SNI', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Broco', 8000, 200, 50, 500, true, NOW(), NOW()),
('EL-018', '8991234567018', 'Steker Biasa 2 Pin', 'Steker biasa 2 pin standar', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Broco', 5000, 300, 80, 800, true, NOW(), NOW()),

-- Terminal & Connector
('EL-019', '8991234567019', 'Terminal Block 10A', 'Terminal block connector 10 ampere', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Phoenix', 12000, 150, 30, 400, true, NOW(), NOW()),
('EL-020', '8991234567020', 'Skun Kabel 1.5mm', 'Skun kabel ring 1.5mm diameter 4', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'box', 'Hanyoung', 25000, 100, 20, 300, true, NOW(), NOW()),

-- Exhaust & Kipas
('EL-021', '8991234567021', 'Exhaust Fan 8 inch', 'Exhaust fan dinding 8 inch', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Panasonic', 125000, 80, 20, 200, true, NOW(), NOW()),
('EL-022', '8991234567022', 'Kipas Angin Stand 16 inch', 'Kipas angin berdiri 16 inch', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Miyako', 250000, 50, 10, 150, true, NOW(), NOW()),

-- Meteran & Alat Ukur
('EL-023', '8991234567023', 'KWH Meter Digital 1P', 'KWH meter digital 1 phase', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Itron', 350000, 30, 10, 100, true, NOW(), NOW()),
('EL-024', '8991234567024', 'Tang Ampere Digital', 'Tang ampere digital AC/DC', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Hioki', 450000, 20, 5, 50, true, NOW(), NOW()),

-- Isolasi & Perlengkapan
('EL-025', '8991234567025', 'Isolasi PVC Hitam', 'Isolasi listrik PVC hitam roll', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', '3M', 8000, 400, 100, 1000, true, NOW(), NOW()),
('EL-026', '8991234567026', 'Cable Tie 20cm', 'Kabel ties hitam 20cm', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pack', 'Kss', 15000, 200, 50, 600, true, NOW(), NOW()),
('EL-027', '8991234567027', 'Lakban Listrik Merah', 'Lakban isolasi listrik warna merah', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'roll', '3M', 12000, 150, 50, 500, true, NOW(), NOW()),

-- Extension & Multi Socket
('EL-028', '8991234567028', 'Stop Kontak 4 Lubang + Saklar', 'Stop kontak portable 4 lubang dengan saklar', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'pcs', 'Uticon', 55000, 100, 30, 300, true, NOW(), NOW()),
('EL-029', '8991234567029', 'Kabel Roll 15m + 4 Socket', 'Kabel roll extension 15 meter 4 lubang', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'unit', 'Kenmaster', 175000, 50, 10, 150, true, NOW(), NOW()),

-- Stabilizer & UPS
('EL-030', '8991234567030', 'Stabilizer 500VA', 'Stabilizer listrik 500VA automatic', (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1), 'unit', 'Matsunaga', 350000, 30, 10, 80, true, NOW(), NOW())

ON CONFLICT (product_code) DO UPDATE SET
    stock_system = EXCLUDED.stock_system,
    updated_at = NOW();

-- Verify
SELECT COUNT(*) as total_produk_listrik FROM products WHERE category_id = (SELECT id FROM categories WHERE name = 'Elektronik & Listrik' LIMIT 1);
