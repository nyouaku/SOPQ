# 17 — Laporan Implementasi Tahap D

**Tanggal:** 2026-11-13
**Cakupan:** Tahap D — fitur yang diiklankan tapi belum ada (Wave 3.1–3.10)
**Status:** SELESAI untuk kode dan build/test lokal. Belum diuji terhadap stack hidup dengan data nyata.

---

## 1. Status per item

| # | Item | Status | Catatan |
|---|---|---|---|
| 3.1 | Query laporan nyata | SELESAI | `repository/report_repository.go`: join detail→sesi→gudang→produk→kategori→supplier→user; 8 filter; count + pagination. Migrasi `000004` memberi index `(stock_opname_id, scanned_at DESC)`. |
| 3.2 | `parseReportFilter` satu sumber | SELESAI | `handler/report_filter.go` dipakai JSON dan CSV/Excel. Filter tidak valid → 400 `INVALID_FILTER`, tidak diabaikan diam-diam. |
| 3.3 | CSV stream dengan batas baris | SELESAI | `ExportStockOpnameCSV` lewat `reportRepo.Stream()`, batas `maxExportRows = 100000` (`ErrExportTooLarge`). |
| 3.4 | Excel berisi baris nyata | SELESAI | `export_usecase.go` memakai `excelize.StreamWriter`, batas `maxExcelExportRows = 50000`. |
| 3.5 | Agregat dashboard | SELESAI | `repository/dashboard_repository.go`: 2 query agregat menggantikan N+1. `recent_activities` ditambah ke response. |
| 3.6 | WS broadcast | SELESAI | `SubmitScan` memanggil `BroadcastNewScan` (`stock_opname_usecase.go:220`); `BroadcastProgress` mengirim payload nyata ke room sesi. |
| 3.7 | Tile ringkasan dari server | SELESAI | `GetStockOpnameReport` mengembalikan `Summary` (total sistem, fisik, selisih, jumlah mismatch) dihitung di SQL atas seluruh hasil filter. Frontend `reports/+page.svelte` membaca `summary` dari response, tidak lagi menjumlah halaman aktif. Tipe di `lib/api/queries.ts` diperbarui. |
| 3.8 | Index pencarian produk | SELESAI | Migrasi `000005`: prefix B-tree pada `barcode` dan `product_code` (partial, `deleted_at IS NULL`). Trigram `name` sudah ada di `performance_indexes.sql`. Preload `Box.Location.Warehouse` ditambahkan di `product_repository.go`. |
| 3.9 | Hapus `/mobile` | SELESAI | Route `dashboard/src/routes/mobile/` dihapus. `pendingScans` tampil di `/scan` (`scan/+page.svelte`). |
| 3.10 | Toast, `$applied`, URL-synced filter | SELESAI | 4 `alert()` di `sessions/[id]/+page.svelte` diganti `toastStore`. Filter reports dibaca dari URL saat load dan ditulis balik lewat `replaceState` setiap kali `applied` berubah. `activeFilters` memakai `$applied` (reaktif). |

---

## 2. Verifikasi

| Pemeriksaan | Hasil |
|---|---|
| `go build ./...` (backend) | lulus |
| `go vet ./...` (backend) | lulus |
| `go test ./...` (backend) | lulus (`config`, `middleware`, `usecase`, `postgres` ok) |
| `npm run check` (dashboard) | 0 error, 0 warning |
| `npm run build` (dashboard) | lulus |
| `flutter analyze` (mobile) | No issues found |
| `flutter test` (mobile) | 3 test lulus |

Pemeriksaan ulang di akhir: `alert(` tidak ada lagi di `dashboard/src`; `/mobile` tidak ada lagi.

---

## 3. Perubahan kontrak API

`GET /api/v1/reports/stock-opname` sekarang menyertakan field baru `summary`:

```json
"summary": {
  "total_system": 0,
  "total_physical": 0,
  "total_difference": 0,
  "mismatch_count": 0
}
```

Field `data`, `total`, `page`, `limit` tidak berubah. Klien lain yang membaca response ini tetap kompatibel (penambahan field saja).

---

## 4. Batasan yang tersisa

- Tile ringkasan pada reports membaca agregat server, tetapi belum ada uji otomatis yang memverifikasi angka `summary` sama dengan penjumlahan manual pada dataset nyata.
- Batas ekspor (100000 CSV / 50000 Excel) adalah pilihan konservatif tanpa pengukuran memori di server produksi.
- Belum ada uji integrasi terhadap PostgreSQL hidup untuk query laporan dan agregat dashboard. Build dan unit test lulus; query SQL belum dijalankan terhadap DB.
- Index trigram dan B-tree baru belum diukur efeknya terhadap `EXPLAIN ANALYZE`.
- Tahap C masih menyisakan 2.11 sebagian (transaksi import belum ada).

---

## 5. File yang berubah (Tahap D)

Backend:
- `internal/repository/report_repository.go` — baru; `Query` kini juga mengembalikan `ReportSummary`
- `internal/repository/dashboard_repository.go` — baru
- `internal/repository/product_repository.go` — preload `Box.Location.Warehouse`
- `internal/repository/session_repository.go` — `RevokeByID` cek `RowsAffected`
- `internal/usecase/report_usecase.go` — `StockOpnameReportResult` + `StockOpnameReportSummary`
- `internal/usecase/export_usecase.go` — ditulis ulang (Excel streaming)
- `internal/usecase/dashboard_usecase.go` — 2 query agregat
- `internal/delivery/http/handler/report.go` — response menyertakan `summary`
- `internal/delivery/http/handler/report_filter.go` — baru
- `internal/delivery/http/handler/export.go` — pakai `parseReportFilter`
- `migrations/000004_phase_d.up.sql`, `000005_phase_d.up.sql` (+ `.down.sql`) — baru

Dashboard:
- `src/routes/sessions/[id]/+page.svelte` — toast menggantikan `alert()`
- `src/routes/reports/+page.svelte` — filter URL-synced, tile dari `summary`
- `src/lib/api/queries.ts` — tipe `summary`
- `src/routes/+page.svelte` — `activity.user?.full_name`
- `src/routes/mobile/` — dihapus
- `src/lib/offline/queue.ts` — komentar diperbarui

---

## 6. Catatan jujur

- Laporan ini ditulis setelah pemeriksaan build/test lokal. Tidak ada klaim performa yang diukur.
- Dua kali edit gagal karena file CRLF dan harus diulang per baris. Tidak mengubah isi kode di luar yang diminta.
