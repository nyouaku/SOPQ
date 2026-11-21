# SOPQ — Revisi Grid Harian jadi Inventory Stock Ledger
## FASE 1 — Audit & Pemetaan (2026-11-21)

Status: **Draft audit — belum ada perubahan kode.**
Referensi: prompt revisi `EXAMPLE_INVENTORY.xlsx`, kode aktual repo, `docs/PROGRESS.md`,
`docs/DATABASE_SCHEMA.sql`, `docs/API_ENDPOINTS.md`.

> Catatan keterbatasan: file `EXAMPLE_INVENTORY.xlsx` tidak tersedia sebagai file di repo yang
> bisa kubaca langsung. Struktur yang dipakai adalah spek yang kamu tempel di chat
> (sheet NOVEMBER/OKTOBER, kolom SALDO AWAL/MASUK/KELUAR/RETUR/SALDO AKHIR, grid M/K harian).
> Sebelum FASE 2, spreadsheet itu perlu ditempatkan di repo supaya kolom & rumus bisa diverifikasi
> persis (terutama RETUR & HARGA MODAL).

---

## 1. Kondisi Grid Harian sekarang

| Aspek | Fakta |
|---|---|
| Fungsi | Grid Excel-like di `dashboard/src/lib/components/SpreadsheetEntry.svelte`, mode alternatif `QuickEntry.svelte` |
| Model | `GridCell` = agregat per produk × hari (day, in_qty, out_qty, unit_cost). **Belum ada** saldo awal/akhir, retur, harga modal |
| Konsep | Grid Harian adalah **tabel input**, bukan ledger. Menyimpan mutasi IN/OUT per tanggal |
| Posting | `PostGrid` mengelompokkan entri per tanggal → satu dokumen `stock_movement` per tanggal, lalu `PostIn`/`PostOut` |
| Transaksi | `PostMovementTx`: header+item + update `stock.quantity_system` (optimistic lock) + baris `stock_ledger`, **dalam satu DB transaction** |
| Sumber data grid | `GridCells`: agregat `stock_movement_item` dengan `movement_date`, filter warehouse + month + movement_type |
| Copy | `CopyGrid` menyalin cell bulan sumber ke bulan tujuan |
| Export | `ExportGridToExcel` sudah ada |
| RBAC | Tulis grid: supervisor/admin. Baca grid: semua terauth (tapi endpoint berada di bawah auth) |

**Inti:** mekanisme transaksi sudah **benar & aman**. Yang belum adalah **tampilan/konsep ledger**: saldo awal, saldo akhir, harga modal, retur, dan ringkasan bulanan seperti workbook.

---

## 2. Kondisi Master Produk sekarang

| Aspek | Fakta |
|---|---|
| Field produk | `product_code, barcode? (lihat catatan), name, brand, unit, location_id, minimum_stock, supplier_id, photo_url, status` |
| Barcode | Domain `Product` **tidak punya field barcode**; scan pakai `product_code` (endpoint `GET /products/code/:code`) |
| Create/Update/Delete | Hanya **admin** (`RequireRole("admin")`) |
| Approval | **Tidak ada** workflow approval untuk create/update produk |
| Nonaktif | Ada `status` (active/inactive/discontinued) + soft delete `deleted_at` |
| Import | Ada template & import Excel, kolom `warehouse_code` + `stock_system` untuk stok awal; stok existing **tidak ditimpa** saat import ulang |
| Daftar produk | Baru saja diperbaiki agar mengirim `stock_system` (total) & `stock[]` per gudang |

**Kesenjangan:** prompt minta **perubahan master produk lewat approval Admin/Supervisor**. Sekarang update langsung oleh admin, tanpa status pending/approval trail.

---

## 3. Kondisi Stock Opname sekarang

| Aspek | Fakta |
|---|---|
| Alur | start session → submit scan → complete → approve → rejected/reopen tersedia |
| Snapshot | `stock_opname_detail.stock_system` = angka saat scan; `difference` generated column |
| Approve | `ApproveWithStock` **menulis `stock.quantity_system = hasil fisik`** + status approved, dalam satu transaksi |
| Ledger | Approve & adjustment mencatat `stock_ledger` (`opname_correction`) bila ledgerRepo diset |
| Role | Supervisor/admin approve; staff scan |
| Realtime | WebSocket session-scoped (`BroadcastNewScan`, `BroadcastSessionUpdate`, `BroadcastDiscrepancyAlert`) |

**Risiko besar revisi:** approve SO **menimpa stok sistem** (`quantity_system = fisik`). Kalau Grid Harian juga mencatat mutasi, harus dipastikan **tidak double count**. Saat ini SO = koreksi saldo, bukan transaksi `stock_movement`/ledger IN-OUT yang terpisah — keputusan ini harus dipertahankan atau dinormalisasi eksplisit di FASE 2.

---

## 4. Mekanisme approval yang tersedia

- **SO approval**: lengkap (status `pending_approval`/`completed`, `ApproveSession`, role supervisor/admin).
- **Produk approval**: **tidak ada**.
- **Inventory movement approval**: **tidak ada** — semua posting langsung `status='posted'`.
- `stock_movement.status` sudah punya `cancelled`, tapi bukan `pending`/`approved`.

Artinya: untuk memenuhi prompt, perlu menambah workflow approval **khusus master produk** (dan mungkin movement) tanpa merusak alur SO yang sudah jalan.

---

## 5. Struktur database relevan

| Tabel | Peran | Catatan |
|---|---|---|
| `products` | master produk | `location_id`, `minimum_stock`, `status`, soft delete |
| `stock` | saldo per produk × gudang | `quantity_system`, `version` (optimistic lock), `UNIQUE(product_id,warehouse_id)` |
| `stock_movement` | header IN/OUT/transfer | `status` posted/cancelled, `created_by`, `deleted_at` |
| `stock_movement_item` | baris produk | `quantity`, `unit_cost`, `movement_date` |
| `stock_ledger` | kartu stok | `transaction_type`, `quantity_in/out`, `balance_after`, `reference_type/id/code` |
| `stock_opname` + detail | sesi & hasil scan | `stock_system`, `stock_physical`, `difference` |
| `stock_adjustment_journal` | penyesuaian formal | sudah ada |
| `warehouses`, `locations` | gudang & area/lokasi | rak/box sudah dihapus (migration 000014) |

**Tidak ada:** tabel `monthly_balance` / `saldo_awal` eksplisit, `retur` terpisah, `product_change_request`/approval, `product_barcode`.

---

## 6. Bagian yang bisa dipakai kembali

- `stock_movement` + `stock_movement_item` + `movement_date` — sudah mendekati grid M/K harian.
- `PostMovementTx` — transaksi aman + optimistic lock + ledger.
- `stock_ledger` — source of truth histori mutasi; **jangan** buat tabel histori duplikat.
- `CancelMovement` — pembalikan transaksi + audit trail.
- `GridCells`, `CopyGrid`, `ExportGridToExcel`.
- `StockRepository.ListByProductIDs` — stok per gudang untuk daftar produk.
- SO approval yang sudah stabil — jangan diubah tanpa perlu.

---

## 7. Bagian yang harus direvisi

| Prioritas | Bagian | Alasan |
|---|---|---|
| P0 | `Summary()` query `SUM(smi.total_cost)` | **BUG**: kolom `stock_movement_item` hanya punya `unit_cost`, bukan `total_cost`. Query berisiko error/aggregate salah. Harus `SUM(smi.quantity * smi.unit_cost)` |
| P0 | Grid UI | Tambah SALDO AWAL, SALDO AKHIR, HARGA MODAL, RETUR, total M/K harian, sesuai workbook |
| P0 | Saldo awal bulan | Harus berasal dari saldo akhir bulan sebelumnya / ledger, **bukan input bebas yang dobel** |
| P1 | Workflow approval produk | Prompt wajib; belum ada |
| P1 | Retur | `MovementSubReturn` sudah ada, tapi belum jadi kolom tersendiri di grid |
| P1 | Dashboard & laporan | Harus pakai source data yang sama (ledger) |
| P2 | `GridCell` | Perlu field saldo awal/akhir/retur/harga modal bila dipakai UI |
| P2 | Barcode produk | Perlu diputuskan apakan barcode ditambahkan di FASE lanjutan |

---

## 8. Risiko terhadap data & fitur existing

1. **Double count SO + Grid** — SO approve mengubah `quantity_system`. Kalau SO juga dicatat sebagai mutasi grid, saldo bisa dobel. **Harus diputuskan eksplisit.**
2. **Import ulang** — sudah ada perlindungan "jangan timpa stok existing". Pakai strategi sama untuk saldo awal.
3. **Saldo awal bulan** — kalau dihitung dari `stock` saat ini, akan berubah saat ada transaksi baru; histori jadi tidak stabil. Harus dari **snapshot/ledger per tanggal**, bukan `stock` live.
4. **Approval produk** — menambah status pending bisa mengganggu create/update admin jika tidak dibuat backward-compatible.
5. **Query Summary bug** — bisa membuat dashboard nilai keluar/masuk salah; ini harus diperbaiki lebih awal.
6. **Migrasi destruktif** — hindari. Semua perubahan sebaiknya additive: tabel baru/kolom nullable + backfill.

---

## 9. Rekomendasi arsitektur final

### Prinsip
- **`stock_ledger` adalah satu-satunya source of truth mutasi.**
- `stock.quantity_system` = saldo berjalan (materialized), selalu diubah **hanya** lewat transaksi yang juga menulis ledger.
- Grid Harian = **view + input** atas ledger, bukan penyimpan angka saldo tersendiri.
- Saldo awal/akhir bulan = **derived** dari ledger (`SUM(quantity_in - quantity_out)` sebelum/tanggal tertentu), bukan kolom free-form.

### Model yang disarankan
- `StockMovement` tetap bukti transaksi IN/OUT/RETUR/transfer.
- Tambah `sub_type = 'retur'` sudah ada (`MovementSubReturn`) — bedakan arah:
  - IN + return = retur pelanggan
  - OUT + return = retur ke supplier
- `GridCell` diperluas dengan `opening`, `closing`, `return_qty`, `unit_cost` (derived, bukan table baru).
- Endpoint baru: `GET /inventory/ledger?warehouse_id&from&to&product_id` — derive saldo & mutasi.
- Approval produk: tabel baru `product_change_request` (additive) + endpoint submit/approve/reject. Produk lama tetap jalan via endpoint admin untuk backward compatibility, atau langsung diarahkan ke request dengan opsi migrasi.

### Hindari
- Menyimpan saldo awal/akhir per produk-bulan sebagai angka statis yang bisa bertentangan dengan ledger.
- Menambah tabel `inventory_history` duplikat `stock_ledger`.
- Mengubah approve SO menjadi transaksi grid tanpa kejelasan — risiko double count.

---

## 10. Urutan implementasi paling aman

| Fase | Isi | Risiko |
|---|---|---|
| **1** | Audit ini (selesai) | rendah |
| **2** | Fix bug `Summary()` (`total_cost` → `quantity*unit_cost`) + test | rendah, langsung bermanfaat |
| **3** | Tambah endpoint `ledger` & derived saldo awal/akhir dari `stock_ledger` + test | sedang |
| **4** | Revisi UI Grid Harian: tampilkan SALDO AWAL, M/K harian, RETUR, SALDO AKHIR, HARGA MODAL; tetap pakai endpoint existing `PostGrid` | sedang |
| **5** | Workflow approval master produk (additive table + endpoint + UI) | tinggi |
| **6** | Integrasi dashboard/laporan ke source ledger yang sama | sedang |
| **7** | Regression test SO (pastikan approve tak double count) + build production | sedang |

---

## Keputusan yang kubutuhkan sebelum FASE 2

1. **File `EXAMPLE_INVENTORY.xlsx`** — tolong letakkan di root project (atau `docs/`) supaya aku bisa baca kolom & rumus persisnya. Spek di chat terpotong; aku tak ingin menebak format RETUR & HARGA MODAL.
2. **SO vs Grid** — apakah hasil approve SO harus **juga tercatat sebagai baris `stock_ledger`** (sudah terjadi bila ledgerRepo diset) dan **jangan** dibuatkan `stock_movement`? Rekomendasiku: **ya, biarkan apa adanya** untuk hindari double count.
3. **Approval produk** — apakah implementasi awal cukup untuk **update** produk saja, atau create juga harus lewat approval?

Tanpa jawaban ini, aku bisa lanjut FASE 2 (fix bug Summary) karena itu aman & jelas. Fase 3+ menunggu keputusan.
