-- ============================================================
-- Migration 000016 DOWN: hapus tabel permintaan perubahan produk.
-- CATATAN: data pengajuan akan hilang. Produk yang sudah di-approve
-- tidak terpengaruh (tersimpan di tabel products).
-- ============================================================

DROP TABLE IF EXISTS product_change_requests;
