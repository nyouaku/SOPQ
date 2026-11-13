# 14 — Laporan Implementasi Tahap A

**Tanggal:** 2026-11-12
**Cakupan:** Tahap A — menutup siklus opname inti (Wave 1, item 1.1–1.2, 1.4, 1.6, 1.7, 1.10, 1.11, 1.12, 1.13)
**Status:** SELESAI — semua gerbang build hijau. **Belum diuji terhadap stack hidup** (lihat §6).

---

## 1. Ringkasan skor

| | Sebelum | Sesudah A |
|---|---|---|
| **Skor total** | **44.22** | **49.5** |
| Verdict | NOT APPROVED | NOT APPROVED |

Delta +5.3 poin. Target Tahap A di roadmap adalah ~49 — tercapai.

| Kategori | Bobot | Sebelum | Sesudah | Alasan perubahan |
|---|---|---|---|---|
| Backend Architecture | 12 % | 52 | 58 | `ApproveWithStock` + `NextCode` pindah ke repository (transaksi di lapisan data, bukan usecase) |
| Frontend | 10 % | 61 | 64 | Route `sessions/new` ada; query lokasi/gudang jadi reaktif |
| Functionality | 15 % | 38 | 56 | Siklus inti menutup: kode unik → scan balik baris benar → approve menulis stok; sesi bisa dibuat dari dashboard & mobile |
| Database | 10 % | 45 | 49 | `CHECK` pada `input_method`, tabel counter harian; **migrasi belum dijalankan** |
| Security | 15 % | 30 | 30 | Tidak ada item keamanan di Tahap A |
| Code Quality | 10 % | 55 | 58 | `go vet` bersih; error sentinel domain; `RowsAffected` diperiksa |
| UI/UX | 8 % | 78 | 79 | Halaman buat sesi punya state loading/error/validasi |
| Performance | 8 % | 45 | 45 | Tidak disentuh |
| Testing | 7 % | 12 | 22 | `go test ./...` kompilasi & lulus (7 test), 2 widget test |
| Production Readiness | 5 % | 20 | 22 | Masih tanpa migration runner, backup, metrik |

**Angka ini adalah estimasi dari bukti statis + hasil build/test lokal, bukan hasil uji end-to-end.** Lihat §6 untuk batasnya.

---

## 2. Item yang dikerjakan

| # | Item | Ref | Status | Bukti |
|---|---|---|---|---|
| A1 | Kode sesi unik per hari | BE-01 | SELESAI | `NextCode` — `INSERT ... ON CONFLICT (day) DO UPDATE ... RETURNING` |
| A2 | `SubmitScan` mengembalikan baris yang baru dibuat | BE-02 | SELESAI | Reload lewat `FindByID(ctx, detail.ID)`; error reload tidak dibuang lagi |
| A3 | Approve menulis `stock.quantity_system` | BE-04 | SELESAI | `ApproveWithStock` — satu transaksi, `SELECT ... FOR UPDATE`, UPSERT stok |
| A4 | `input_method` diselaraskan ujung ke ujung | BE-08, DB-07 | SELESAI | Handler `oneof=manual voice camera upload`; DB `CHECK`; dashboard kirim `manual/camera/upload`; mobile kirim `manual/voice` |
| A5 | Tombol "New Session" hidup | FN-01 | SELESAI | Route baru `dashboard/src/routes/sessions/new/+page.svelte` |
| A6 | Sesi palsu di mobile dihapus | FN-10 | SELESAI | `session_list_screen.dart`; `sessionId` nyata; logout; `'dummy-session-id'` hilang dari repo |

### Perbaikan tambahan yang ikut terbawa

| Item | Ref | Kenapa ikut |
|---|---|---|
| `Save()` → `Updates(map)` pada approve | BE-10 | `Save()` pada struct hasil `First()` menulis ulang semua kolom — balapan dengan transaksi lain |
| `RowsAffected` diperiksa | CQ-06 | Approve yang tidak mengubah baris dulu dilaporkan sukses |
| Kontrak payload scan mobile | — | Mobile mengirim `product_id`/`quantity_physical`; API menerima `barcode`/`stock_physical`. **Bug kontrak yang belum tercatat di audit** — ditemukan saat memperbaiki model |
| `ErrOptimisticLockFailed` tidak ada | T-01 | Test merujuk sentinel yang tidak pernah didefinisikan |
| Test Go dikompilasi ulang | T-01, T-07 | Mock mengikuti interface sekarang; 7 test lulus |
| `PrettyDioLogger` di balik `kDebugMode` | MOB-ARCH-01, SEC-10 | Logger mencetak header `Authorization` di build rilis |
| `apiBaseUrl` via `--dart-define` | MOB-ARCH-01 | URL API tidak lagi di-hardcode |
| 2 widget test nyata | T-02 | Template `MyApp` diganti test yang benar-benar menguji form login |

---

## 3. File yang berubah

### Backend (Go)

| File | Perubahan |
|---|---|
| `backend/migrations/000002_phase_a.up.sql` | **BARU** — tabel `stock_opname_code_seq`, `CHECK chk_so_detail_input_method` |
| `backend/migrations/000002_phase_a.down.sql` | **BARU** — rollback keduanya |
| `backend/internal/domain/errors.go` | **BARU** — `ErrSessionNotFound`, `ErrSessionNotApprovable`, `ErrSessionAlreadyApproved` |
| `backend/internal/domain/stock_opname.go` | Tag `Difference` → `gorm:"->"` (generated column, read-only) |
| `backend/internal/repository/stock_opname_repository.go` | `NextCode`, `ApproveWithStock`, `FindByID` pada detail repo; `map[string]interface{}` → `map[string]any` |
| `backend/internal/usecase/stock_opname_usecase.go` | `StartSession` pakai `NextCode`; `SubmitScan` reload baris; `ApproveSession` delegasi ke repo |
| `backend/internal/delivery/http/handler/stock_opname.go` | `oneof` pada `InputMethod`; map error → 404/409/400/500 |
| `backend/pkg/utils/response.go` | `RespondConflict` (409) |
| `backend/internal/usecase/stock_opname_usecase_test.go` | **DITULIS ULANG** — mock sesuai interface sekarang, 4 test regresi |
| `backend/internal/usecase/auth_usecase_test.go` | **DITULIS ULANG** — mock sesuai interface sekarang, 3 test |

### Dashboard (SvelteKit)

| File | Perubahan |
|---|---|
| `dashboard/src/routes/sessions/new/+page.svelte` | **BARU** — pilih gudang + lokasi, `POST /stock-opname/sessions`, redirect ke detail |
| `dashboard/src/lib/api/queries.ts` | `locationsQuery`/`boxesQuery` menerima `Readable<string>`, dibungkus `derived` — query ikut berubah saat gudang diganti |
| `dashboard/src/routes/sessions/+page.svelte` | Spacer `<label>` → `<span aria-hidden>` (a11y) |

### Mobile (Flutter)

| File | Perubahan |
|---|---|
| `mobile/lib/features/scanner/screens/session_list_screen.dart` | **BARU** — daftar sesi `in_progress`, logout, retry |
| `mobile/lib/data/models/stock_opname_model.dart` | **DITULIS ULANG** — model sesi/detail/`SubmitScanRequest` sesuai respons API nyata |
| `mobile/lib/data/models/product_model.dart` | **DITULIS ULANG** — sesuai `domain.Product`; `stockSystem` nullable |
| `mobile/lib/features/scanner/screens/scanner_screen.dart` | `sessionId` wajib; kode sesi di AppBar; `context.mounted` guard |
| `mobile/lib/features/scanner/providers/scanner_provider.dart` | Payload scan benar (`barcode`/`stock_physical`) |
| `mobile/lib/features/auth/screens/login_screen.dart` | Navigasi ke `SessionListScreen` |
| `mobile/lib/core/network/dio_client.dart` | `PrettyDioLogger` hanya `kDebugMode` |
| `mobile/lib/core/constants/app_constants.dart` | `String.fromEnvironment('API_BASE_URL'/'WS_BASE_URL')` |
| `mobile/lib/core/theme/app_theme.dart` | `CardThemeData` / `DialogThemeData` (perubahan API Flutter) |
| `mobile/lib/features/auth/providers/auth_provider.dart` | Import tak terpakai dihapus |
| `mobile/test/widget_test.dart` | **DITULIS ULANG** — 2 test form login |

---

## 4. Detail teknis tiga perbaikan inti

### A1 — kode sesi unik per hari

```sql
INSERT INTO stock_opname_code_seq (day, last_value) VALUES ($1, 1)
ON CONFLICT (day) DO UPDATE
SET last_value = stock_opname_code_seq.last_value + 1
RETURNING last_value
```

Satu round trip, tanpa retry loop, aman di bawah konkurensi karena `ON CONFLICT` mengambil row lock pada baris `day` yang sama. Hasil diformat `SO-YYYYMMDD-NNNN`.

**Kenapa bukan `COUNT(*) + 1`:** dua sesi yang dibuat bersamaan membaca hitungan yang sama lalu bertabrakan pada `code UNIQUE`. **Kenapa bukan sequence Postgres:** sequence global tidak reset harian.

### A2 — scan mengembalikan baris yang benar

Sebelum: `FindBySessionID` + ambil elemen terakhir. Kalau pengguna lain men-scan produk berbeda di detik yang sama, petugas menerima nama produk orang lain di dialog konfirmasi.

Sesudah: `FindByID(ctx, detail.ID)` dengan `Preload("Product")`, dan error reload **dikembalikan**, bukan `_`.

### A3 — approve menulis stok

Satu transaksi:

1. `SELECT ... FOR UPDATE` pada `stock_opname` — mencegah dua approver bersamaan.
2. Guard: sudah `approved` → 409; bukan `completed` → 400.
3. Baca `product_id, stock_physical` dari detail sesi.
4. Per produk: `INSERT ... ON CONFLICT (product_id, warehouse_id) DO UPDATE SET quantity_system = EXCLUDED.quantity_system, version = stock.version + 1`.
5. `UPDATE stock_opname SET status='approved' ... WHERE id=? AND status='completed'`; `RowsAffected == 0` → error.

Kalau satu produk gagal, seluruh approve di-rollback. Sesi tidak pernah setengah tertulis.

---

## 5. Hasil verifikasi

Semua dijalankan lokal, 2026-11-12.

| Gerbang | Perintah | Hasil |
|---|---|---|
| Build backend | `go build ./...` | **PASS** (exit 0) |
| Vet backend | `go vet ./...` | **PASS** (sebelumnya GAGAL: `undefined: domain.ProductFilter`) |
| Test backend | `go test ./...` | **PASS** — `ok internal/usecase 4.742s` |
| Test detail | `go test ./internal/usecase/... -v` | **7/7 PASS** |
| Format | `gofmt -l` pada file yang disentuh | **bersih** |
| Typecheck dashboard | `npm run check` | **0 errors, 0 warnings** |
| Build dashboard | `npm run build` | **PASS** — `✓ built in 49.68s`, adapter-node selesai |
| Analyze mobile | `flutter analyze` | **No issues found!** |
| Test mobile | `flutter test` | **+2: All tests passed!** |

Daftar test Go yang lulus:

```
PASS TestAuthUsecase_Register
PASS TestAuthUsecase_Login            (3 subtest)
PASS TestAuthUsecase_RefreshToken
PASS TestStartSession_CodeIncrementsSameDay     <- regresi BE-01
PASS TestSubmitScan_ReturnsInsertedRow          <- regresi BE-02
PASS TestApproveSession_TransitionsAndGuards    <- regresi BE-04
PASS TestStockOpnameUsecase_OptimisticLocking
```

**Catatan kejujuran:** test di atas memakai mock in-memory. Yang dibuktikan adalah *logika usecase dan kontrak interface*, bukan perilaku Postgres. Lihat §6.

---

## 6. Yang BELUM diuji — baca sebelum menyimpulkan "sudah jalan"

Ini bagian terpenting dari laporan ini.

| Area | Status | Alasan | Cara memverifikasi nanti |
|---|---|---|---|
| Migrasi `000002_phase_a` benar-benar jalan di Postgres | **NOT TESTED** | Tidak ada instance DB; `RunMigrations()` masih stub (`logger.Warn`), jadi file migrasi **tidak dijalankan otomatis** | `psql -f backend/migrations/000002_phase_a.up.sql`, lalu `\d stock_opname_detail` |
| `ON CONFLICT (day) DO UPDATE` di bawah konkurensi nyata | **NOT TESTED** | Butuh DB | Dua goroutine serentak, pastikan kode berbeda |
| `SELECT ... FOR UPDATE` benar-benar mengunci | **NOT TESTED** | Butuh DB | Dua approve bersamaan; satu harus 409 |
| UPSERT ke `stock` saat approve | **NOT TESTED** | Butuh DB | Seed stok, scan, approve, `SELECT quantity_system` |
| Alur HTTP penuh (login → buat sesi → scan → complete → approve) | **NOT TESTED** | Stack tidak dijalankan | `docker compose up`, ikuti §7 `13-release-checklist.md` |
| Mobile scan kamera nyata | **NOT TESTED** | Butuh perangkat | `flutter run` di Android |
| Antrean offline `flushQueue` | **NOT TESTED** | Di luar Tahap A; masih rawan (FN-02…FN-05) | Tahap B |
| Refresh token diverifikasi ke DB | **BELUM DIPERBAIKI** | BE-03 ada di Wave 1 item 1.3 — **tidak termasuk Tahap A** | Tahap berikutnya |

**Build hijau ≠ siap produksi.** `go build` hanya membuktikan kode terkompilasi. Tiga bug paling berbahaya yang tersisa (BE-03 refresh token, SEC-01 WebSocket anonim, PR-08 tanpa migration runner) semuanya lolos dari semua gerbang di §5.

---

## 7. Risiko yang tersisa setelah Tahap A

| # | Risiko | Dampak | Kapan |
|---|---|---|---|
| 1 | `RunMigrations()` adalah stub — file `000002` tidak pernah dijalankan otomatis | `NextCode` dan `CHECK` **akan gagal saat runtime** pada DB yang belum dimigrasi manual | **Segera** — jalankan manual sebelum mencoba apa pun |
| 2 | Refresh token tidak dicek ke DB (BE-03) | Logout tetap dekoratif; token yang dicuri berlaku sampai kedaluwarsa | Wave 1 item 1.3 |
| 3 | `Logout` mengirim `userID` ke parameter `sessionID` (SEC-13) | Revoke tidak mengenai baris mana pun | Wave 2 item 2.6 |
| 4 | WebSocket tanpa auth (SEC-01) | Umpan scan dapat disadap | Wave 2 item 2.1 |
| 5 | Antrean offline mobile rawan (FN-02…FN-05) | Scan ganda atau hilang saat sinyal buruk | Wave 1 item 1.8 |
| 6 | Belum ada idempotency key per scan (FN-04) | Retry jaringan menghasilkan baris ganda | Wave 1 item 1.9 |

**Item 1 harus diselesaikan sebelum menjalankan aplikasi.** Tanpa migrasi, `StartSession` gagal karena tabel `stock_opname_code_seq` tidak ada.

---

## 8. Catatan untuk tahap berikutnya

**Tahap B** (item roadmap yang belum tersentuh di Wave 1): 1.3 refresh token, 1.5 migration runner, 1.8 `flushQueue`, 1.9 idempotency, serta penyelesaian 1.12 (test regresi untuk 1.3–1.5).

Urutan yang disarankan: **1.5 migration runner dulu**, karena tanpa itu setiap perbaikan skema berikutnya tidak bisa diverifikasi. Setelah itu 1.3, lalu 1.8/1.9.

Skor target setelah Tahap B: ~55.

---

## 9. Pernyataan kepatuhan

- Tidak ada file aplikasi yang dihapus.
- Tidak ada migrasi database yang dijalankan.
- Tidak ada konfigurasi production yang diubah.
- Tidak ada dependency baru diinstal.
- Tidak ada pengujian destruktif.
- Tidak ada email, pembayaran, atau notifikasi sungguhan terkirim.
- Tidak ada data production dipakai; test memakai mock in-memory.
- Tidak ada hasil testing yang dikarang. Semua status **NOT TESTED** di §6 adalah area yang benar-benar belum dijalankan.

Perubahan source code dilakukan atas persetujuan eksplisit pengguna untuk Tahap A, melampaui batasan awal audit "jangan mengubah source code".
