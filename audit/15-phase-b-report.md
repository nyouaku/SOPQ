# 15 — Laporan Implementasi Tahap B

**Tanggal:** 2026-11-12
**Cakupan:** Tahap B — migration runner, refresh-token verification & rotation, offline queue, idempotency scan
**Status:** SELESAI — semua gerbang build/test hijau. **Tidak diuji terhadap stack hidup** (lihat §6).

---

## 1. Ringkasan skor

| | Sebelum A | Setelah A | Setelah B |
|---|---|---|---|
| **Skor total** | 44.22 | 49.5 | **56.0** |
| Verdict | NOT APPROVED | NOT APPROVED | NOT APPROVED |

Delta Tahap B: +6.5 poin. Target roadmap untuk Wave 1 item 1.3/1.5/1.8/1.9/1.12 adalah ~55 — tercapai.

| Kategori | Bobot | A | B | Alasan perubahan |
|---|---|---|---|---|
| Backend Architecture | 12 % | 58 | 62 | Migration runner nyata; rotasi sesi pindah ke repository (transaksi di lapisan data) |
| Frontend | 10 % | 64 | 64 | Tidak disentuh |
| Functionality | 15 % | 56 | 62 | Retry scan tidak lagi menggandakan baris; antrean offline punya batas percobaan |
| Database | 10 % | 49 | 58 | Migrasi benar-benar diterapkan otomatis + tabel `schema_migrations`; index & unique partial index baru |
| Security | 15 % | 30 | 42 | **BE-03 tertutup**: refresh token diverifikasi ke DB, rotasi sekali pakai, deteksi pemakaian ulang, logout benar-benar mencabut |
| Code Quality | 10 % | 58 | 62 | 12 test Go lulus; helper migrasi punya test unit |
| UI/UX | 8 % | 79 | 79 | Tidak disentuh |
| Performance | 8 % | 45 | 47 | Index `idx_sessions_refresh_hash` + partial unique index scan |
| Testing | 7 % | 22 | 40 | 12 test backend (dari 7); mobile tetap 2 |
| Production Readiness | 5 % | 22 | 30 | Migration runner menghapus blocker deploy paling keras (PR-08) |

Masih estimasi dari bukti statis + build/test lokal. Lihat §6.

---

## 2. Item yang dikerjakan

| # | Item | Ref | Status | Bukti |
|---|---|---|---|---|
| B1 | Migration runner sungguhan | PR-08, CQ-08 | SELESAI | `internal/infrastructure/postgres/migrate.go`; tabel `schema_migrations`; baseline untuk DB lama |
| B2 | Refresh token diverifikasi ke DB + rotasi | BE-03, SEC-04 | SELESAI | `FindByRefreshTokenHash` + `RotateRefreshToken` (transaksi, sekali pakai) |
| B3 | Deteksi pemakaian ulang refresh token | SEC-04 | SELESAI | Token yang sudah dirotasi → `ErrRefreshTokenReused` → semua sesi user dicabut |
| B4 | Logout mencabut sesi nyata | SEC-13 | SELESAI | `RevokeByUserID`; dulu `userID` masuk ke parameter `sessionID` |
| B5 | Antrean offline: re-entrancy + attempt + dead-letter | FN-02…FN-05 | SELESAI | `_isSyncing` guard, `maxAttempts = 5`, status `syncing` diklaim atomik, `resetStuckSyncing` |
| B6 | Idempotency key per scan | FN-04 | SELESAI | Kolom `idempotency_key` + partial unique index; mobile kirim UUID antrean |
| B7 | Test regresi untuk B1–B6 | T-01, T-07 | SELESAI | 12 test, semuanya lulus |

### Perbaikan tambahan yang ikut terbawa

| Item | Kenapa |
|---|---|
| **`GenerateRefreshToken` menghasilkan token identik** | Token hanya berisi `sub` + `iat` (resolusi detik). Dua login dalam detik yang sama menghasilkan string **sama persis**, sehingga rotasi tidak pernah benar-benar mengganti token dan pengujian pemakaian ulang tidak mungkin lulus. Ditambahkan `jti` (UUID acak). **Bug ini tidak ada di daftar audit** — ditemukan karena test menolak lulus. |
| Mount `docker-entrypoint-initdb.d` dihapus | Kalau dibiarkan, Postgres menjalankan `000001`+`000002` saat init, lalu runner menjalankan lagi → `ALTER TABLE ... ADD CONSTRAINT` gagal (constraint sudah ada). Sekarang satu jalur migrasi saja. |
| `splitSQL` menangani `$$` kosong | Tag `$tag$` kosong (`$$`) menjadi kasus yang salah terdeteksi sebagai tag kosong; dipisah ke flag `inDollar`. Ditemukan oleh test unit, bukan oleh review. |

---

## 3. File yang berubah

### Backend

| File | Perubahan |
|---|---|
| `backend/internal/infrastructure/postgres/migrate.go` | **BARU** — `RunMigrations(dir)`, `splitSQL`, `listUpMigrations`, baseline, `schema_migrations` |
| `backend/internal/infrastructure/postgres/migrate_test.go` | **BARU** — 10 subtest pemecah SQL + 2 test daftar migrasi |
| `backend/internal/infrastructure/postgres/postgres.go` | Stub `RunMigrations` dihapus |
| `backend/cmd/api/main.go` | `RunMigrations(os.Getenv("MIGRATIONS_DIR") ?? "migrations")` |
| `backend/migrations/000003_phase_b.up.sql` | **BARU** — index refresh hash, `idempotency_key`, partial unique index |
| `backend/migrations/000003_phase_b.down.sql` | **BARU** |
| `backend/internal/domain/errors.go` | `ErrRefreshTokenInvalid`, `ErrRefreshTokenReused`, `ErrUserInactive`, `ErrScanDuplicate` |
| `backend/internal/domain/stock_opname.go` | Field `IdempotencyKey *string` |
| `backend/internal/repository/session_repository.go` | `FindByRefreshTokenHash`, `RotateRefreshToken` |
| `backend/internal/repository/stock_opname_repository.go` | `FindByIdempotencyKey` |
| `backend/internal/usecase/auth_usecase.go` | `RefreshToken` verifikasi+rotasi; `Logout(userID)` pakai `RevokeByUserID` |
| `backend/internal/usecase/stock_opname_usecase.go` | Idempotency lookup sebelum insert; fallback saat balapan |
| `backend/internal/delivery/http/handler/auth.go` | Map error refresh → 401/500, bukan bocor `err.Error()` |
| `backend/internal/delivery/http/handler/stock_opname.go` | Field `idempotency_key` (`omitempty,uuid`) |
| `backend/pkg/utils/jwt.go` | `jti` acak pada refresh token |
| `backend/internal/usecase/auth_usecase_test.go` | Mock `RotateRefreshToken`/`FindByRefreshTokenHash`; test reuse + logout |
| `backend/internal/usecase/stock_opname_usecase_test.go` | Mock `FindByIdempotencyKey`; test retry idempoten |

### Mobile

| File | Perubahan |
|---|---|
| `mobile/lib/data/models/stock_opname_model.dart` | `idempotencyKey` opsional pada `SubmitScanRequest` |
| `mobile/lib/data/datasources/scan_local_datasource.dart` | `claimForSync`, `resetStuckSyncing`, batas attempt 5, `idempotencyKey` getter |
| `mobile/lib/features/scanner/providers/scanner_provider.dart` | `SyncService`: disposability, klaim atomik, dead-letter; scan online pakai UUID |
| `docker-compose.yml` | Mount `docker-entrypoint-initdb.d` dihapus |

---

## 4. Detail teknis

### B1 — migration runner

```
RunMigrations(dir)
  ├─ CREATE TABLE IF NOT EXISTS schema_migrations(version, name, applied_at)
  ├─ baseline: kalau tabel `users` sudah ada → catat versi 1
  ├─ list *.up.sql, urut versi, tolak versi duplikat
  └─ untuk tiap versi yang belum tercatat:
       BEGIN
         jalankan tiap statement (dipecah splitSQL)
         INSERT INTO schema_migrations
       COMMIT
```

`splitSQL` menghormati: komentar `--` dan `/* */`, literal `'...'` (termasuk `''`), identifier `"..."`, dan blok dollar-quoted `$$...$$` / `$tag$...$tag$`. Tanpa ini, `BEGIN ... DELETE FROM t; ... END;` di dalam fungsi plpgsql akan terpotong di tengah dan migrasi gagal.

**Baseline** penting: database yang dibuat lewat `docker-entrypoint-initdb.d` sudah punya `users` tapi tidak punya `schema_migrations`. Tanpa baseline, runner akan mencoba `CREATE TABLE roles (...)` dan gagal.

### B2/B3 — refresh token

Alur baru:

1. Parse JWT → `userID`; gagal → `ErrRefreshTokenInvalid`.
2. `SELECT * FROM sessions WHERE refresh_token_hash = $1 ORDER BY created_at DESC LIMIT 1`.
   - tidak ada → `ErrRefreshTokenInvalid`
   - `revoked_at IS NOT NULL` → **token dipakai ulang** → `RevokeByUserID` → `ErrRefreshTokenReused`
   - `expires_at` lewat → `ErrRefreshTokenInvalid`
3. Cek `users.is_active`.
4. Generate access + refresh baru (dengan `jti` acak).
5. `RotateRefreshToken` dalam transaksi: `UPDATE sessions SET revoked_at=now() WHERE id=$1 AND revoked_at IS NULL`; `RowsAffected == 0` → `ErrRefreshTokenReused`.

Langkah 2 dan 5 sama-sama memeriksa; yang kedua menangani balapan dua request paralel dengan token yang sama.

### B5 — antrean offline

| Sebelum | Sesudah |
|---|---|
| `_isSyncing` tanpa `try/finally` yang aman → panic meninggalkan flag menyala selamanya | `try/finally` + `_disposed`; provider punya `onDispose` |
| Baris langsung ditandai `synced`/`failed`, dua putaran bisa memproses baris yang sama | `claimForSync` — UPDATE bersyarat `sync_status IN (pending, failed)`; klaim gagal → lewati |
| Gagal → selalu `failed`, tidak pernah dicoba lagi setelah 3× | Gagal → `pending` (coba lagi) sampai `maxAttempts = 5`, baru `failed` (dead-letter) |
| App ditutup saat `syncing` → baris hilang | `resetStuckSyncing()` di awal setiap putaran |
| `getPendingScans` tidak melihat `retry_count` | Keduanya filter `retry_count < 5` |

### B6 — idempotency

Klien mengirim `idempotency_key` (UUID). Server:

1. Kalau kunci kosong → perilaku lama (scan web).
2. Kalau ada dan baris dengan kunci sama sudah ada di sesi itu → kembalikan baris itu, **tanpa insert**.
3. Kalau tidak ada → insert. Kalau insert gagal (unique violation karena request kembar menang duluan) → cari lagi, kembalikan baris pemenang.

Mobile memakai `scan_queue.id` sebagai kunci — UUID yang dibuat **sekali** saat scan masuk antrean, jadi 5 kali retry mengirim kunci yang sama.

```sql
CREATE UNIQUE INDEX uq_so_detail_idempotency
    ON stock_opname_detail (stock_opname_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
```

---

## 5. Hasil verifikasi

Dijalankan lokal, 2026-11-12.

| Gerbang | Perintah | Hasil |
|---|---|---|
| Build backend | `go build ./...` | **PASS** |
| Vet backend | `go vet ./...` | **PASS** |
| Test backend | `go test ./...` | **PASS** — 12 test |
| Format | `gofmt -l` pada file tersentuh | **bersih** |
| Analyze mobile | `flutter analyze` | **No issues found!** |
| Test mobile | `flutter test` | **+2: All tests passed!** |
| Typecheck dashboard | `npm run check` | **0 errors, 0 warnings** (tidak berubah, tidak disentuh) |

Test backend yang lulus:

```
PASS TestSplitSQL                                     (10 subtest)
PASS TestListUpMigrations
PASS TestListUpMigrations_RejectsDuplicateVersion
PASS TestAuthUsecase_Register
PASS TestAuthUsecase_Login                            (3 subtest)
PASS TestAuthUsecase_RefreshToken                     <- rotasi + deteksi reuse
PASS TestAuthUsecase_Logout_RevokesAllSessions        <- regresi SEC-13
PASS TestStartSession_CodeIncrementsSameDay           <- regresi BE-01
PASS TestSubmitScan_ReturnsInsertedRow                <- regresi BE-02
PASS TestApproveSession_TransitionsAndGuards          <- regresi BE-04
PASS TestSubmitScan_IdempotentRetry                   <- regresi FN-04
PASS TestStockOpnameUsecase_OptimisticLocking
```

Dua bug ditemukan **karena** test ini ditulis, bukan oleh review: `$$` pada `splitSQL`, dan refresh token deterministik.

---

## 6. Yang BELUM diuji

| Area | Status | Alasan | Cara verifikasi |
|---|---|---|---|
| `RunMigrations` terhadap Postgres nyata | **NOT TESTED** | Tidak ada instance DB. Yang diuji hanya pemecah SQL dan daftar file | `docker compose up postgres`, jalankan backend, cek `SELECT * FROM schema_migrations` |
| Baseline pada DB lama | **NOT TESTED** | Butuh DB yang sudah terisi | DB hasil `000001` manual → runner harus mencatat versi 1 dan hanya menjalankan 2 dan 3 |
| `FOR UPDATE` / unique index idempotency di bawah konkurensi | **NOT TESTED** | Butuh DB | Dua request serentak dengan kunci sama → satu baris |
| Rotasi refresh token terhadap DB | **NOT TESTED** | Mock in-memory | Login, refresh, refresh lagi dengan token lama → 401 |
| Antrean offline di perangkat | **NOT TESTED** | Butuh Android | Matikan jaringan, 6 scan, nyalakan → 5 terkirim, 1 dead-letter |
| Alur HTTP penuh | **NOT TESTED** | Stack tidak dijalankan | `13-release-checklist.md` |

---

## 7. Risiko yang tersisa

| # | Risiko | Dampak | Kapan |
|---|---|---|---|
| 1 | WebSocket anonim (SEC-01) | Umpan scan dapat disadap | Tahap C / Wave 2.1 |
| 2 | `/export/*` tanpa `RequireRole` (SEC-02) | Staff bisa unduh seluruh data | Wave 2.2 |
| 3 | Rate limit belum dipasang di login/refresh (SEC-06) | Brute force password & token | Wave 2.3 |
| 4 | CORS bisa `*` + credentials (SEC-07) | Origin mana pun bisa kirim kredensial | Wave 2.4 |
| 5 | Registrasi: password min 6, role 1–3 hardcoded (SEC-09) | Password lemah, role tak tervalidasi ke tabel | Wave 2.5 |
| 6 | `VerifyToken` tidak constant-time (SEC-05) | Timing attack teoretis pada hash token | Wave 2.7 |
| 7 | Idempotency belum dipakai klien web | Retry dashboard masih bisa ganda | Opsional; dashboard tidak punya antrean offline |
| 8 | Dead-letter tidak punya UI | Baris gagal permanen tak terlihat petugas | Tahap D |

---

## 8. Catatan untuk tahap berikutnya

**Tahap C** = keamanan dan kontrol akses (Wave 2): 2.1 WebSocket auth, 2.2 role export, 2.3 rate limit, 2.4 CORS, 2.5 registrasi, 2.6 audit auth, 2.7 constant-time, 2.8 filter menu per role.

Skor target: ~62.

**Prasyarat sebelum menjalankan aplikasi:** jalankan `docker compose up` dengan volume `pgdata` lama akan membuat runner mencoba migrasi ke DB yang sudah punya tabel — baseline menanganinya. DB benar-benar baru akan menjalankan 000001 lengkap. Keduanya belum diuji terhadap Postgres sungguhan; jangan menandai lolos sebelum §6 diperiksa.

---

## 9. Pernyataan kepatuhan

- Tidak ada file aplikasi dihapus.
- Tidak ada migrasi database dijalankan terhadap instance nyata.
- Tidak ada konfigurasi production diubah. Satu perubahan `docker-compose.yml`: mount `docker-entrypoint-initdb.d` dihapus agar migrasi punya satu jalur saja — **ini perubahan perilaku deploy dan perlu disadari**.
- Tidak ada dependency baru diinstal. (`jti` memakai `uuid.NewRandom` yang sudah ada.)
- Tidak ada pengujian destruktif, email/pembayaran/notifikasi, atau data production.
- Tidak ada hasil testing dikarang; semua NOT TESTED di §6 benar-benar belum dijalankan.

Perubahan source code dilakukan atas persetujuan eksplisit pengguna untuk Tahap B.
