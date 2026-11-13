# 16 — Laporan Implementasi Tahap C

**Tanggal:** 2026-11-13
**Cakupan:** Tahap C — keamanan & kontrol akses (Wave 2.1–2.12)
**Status:** SELESAI untuk kode dan build/test lokal. **Belum diuji terhadap stack hidup** (lihat §6). Ada satu blocker konfigurasi lokal (lihat §5).

---

## 1. Ringkasan skor

| | Setelah B | Setelah C (estimasi) |
|---|---|---|
| **Skor total** | 56.0 | **~62** (estimasi statis; belum dihitung ulang oleh audit penuh) |
| Verdict | NOT APPROVED | NOT APPROVED |

Target roadmap Tahap C adalah ~62. Angka ini estimasi dari bukti kode + build/test lokal, bukan hasil audit ulang.

---

## 2. Status per item

| # | Item | Status | Catatan |
|---|---|---|---|
| 2.1 | Autentikasi WebSocket | SELESAI | Token pindah ke `Sec-WebSocket-Protocol` (`access_token, <jwt>`); query string dihapus; cek akses per sesi (`canAccessSession`); origin dari config, origin kosong ditolak. Client `sessions/[id]/+page.svelte` ikut diubah. |
| 2.2 | Role gate export | SELESAI | `/export` sekarang `RequireRole("supervisor", "admin")`. |
| 2.3 | Rate limit login & refresh | SELESAI | `/auth/login`: 10/15 mnt per IP + 5/15 mnt per username. `/auth/refresh`: 30/15 mnt per IP. Penghitung memakai Lua script (INCR + PEXPIRE atomik). |
| 2.4 | CORS | SELESAI | Wildcard `*` tidak lagi digabung `Allow-Credentials`. Default `CORS_ALLOWED_ORIGINS` = `http://localhost:3000`. Origin asing tidak mendapat `Allow-Origin`. nginx: `Access-Control-Allow-Origin *` dihapus. |
| 2.5 | Registrasi | SELESAI | Password min 12 karakter; `role_id` divalidasi ke tabel `roles` (400 bila tidak ada); username duplikat → **409** (termasuk balapan lewat pemetaan `23505`). Endpoint sudah admin-only sebelumnya. |
| 2.6 | Revoke & audit auth | SELESAI | `RevokeByID` cek `RowsAffected`. Peristiwa auth dicatat ke `activity_logs`: `refresh_reuse_detected`, `logout`. Audit gagal tidak menggagalkan request. |
| 2.7 | Perbandingan token konstan-waktu | SELESAI | `VerifyToken` pakai `crypto/subtle.ConstantTimeCompare`. **JWT `iss`/`aud` TIDAK ditambahkan** (belum dikerjakan). |
| 2.8 | Filter menu per role | SELESAI | `menu.ts` punya field `roles` + `menuForRole()`. Sidebar, TopNav, Command, BottomBar memakai filter. Menu hanya disembunyikan di UI; proteksi nyata tetap di backend. |
| 2.9 | Security headers nginx | SELESAI | `nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy`, `Permissions-Policy`, CSP. **CSP mengizinkan `'unsafe-inline'` untuk script & style** karena SvelteKit belum memakai nonce; ini melemahkan CSP. |
| 2.10 | Fail-fast config & port | SELESAI | Secret < 32 karakter, kosong, mengandung nilai contoh, atau access = refresh → gagal start. `DB_PASSWORD` wajib saat `APP_ENV=production`. Port host Postgres & Redis dihapus. Default `changeme` di compose dihapus (`:?` wajib). |
| 2.11 | Batas import | SEBAGIAN | Cap 2000 baris data per file (`ErrImportTooLarge`, 400). Laporan per-baris sudah ada sebelumnya. **Transaksi import belum ada**: baris yang sukses tetap tersimpan meski sebagian gagal. |
| 2.12 | Metadata notifikasi | SELESAI | `fmt.Sprintf` JSON diganti `json.Marshal`; nama produk dengan kutip/backslash tidak lagi menghasilkan jsonb rusak. |

---

## 3. File yang berubah (Tahap C)

Backend:
- `internal/delivery/websocket/stock_opname_ws.go`, `config.go`, `publisher.go` — auth WS (sudah dimulai sebelumnya, dilengkapi)
- `internal/delivery/http/route/routes.go` — role gate export, limiter auth, `canAccessSession`, wiring `activityLogRepo`
- `internal/middleware/rate_limit.go` — ditulis ulang: `allow()`, `LoginLimiter`, `RateLimiterByKey`, Lua script
- `internal/middleware/cors.go` — ditulis ulang (origin allow-list, wildcard tanpa credentials)
- `internal/middleware/cors_test.go` — **baru**
- `internal/config/config.go` — `validateSecret`, default CORS, fail-fast produksi
- `internal/config/config_test.go` — **baru**
- `internal/domain/errors.go` — error baru: `ErrUsernameTaken`, `ErrRoleNotFound`, `ErrWeakPassword`, `ErrImportTooLarge`
- `internal/usecase/auth_usecase.go` — validasi register, audit, `activityLogRepo`
- `internal/usecase/auth_usecase_test.go` — password fixture 12 karakter; konstruktor baru
- `internal/usecase/import_usecase.go` — `maxImportRows = 2000`
- `internal/usecase/notification_usecase.go` — `json.Marshal` metadata
- `internal/repository/user_repository.go` — pemetaan `23505` → `ErrUsernameTaken`
- `internal/repository/session_repository.go` — `RevokeByID` cek `RowsAffected`
- `internal/delivery/http/handler/auth.go` — mapping 409 / 400 / 500 untuk register
- `pkg/utils/password.go` — `VerifyToken` konstan-waktu
- `internal/middleware` & `internal/usecase` lain: penyesuaian signature

Dashboard:
- `src/routes/sessions/[id]/+page.svelte` — WS memakai sub-protocol; path `/ws/stock-opname/sessions/:id` (path lama `/stock-opname/sessions/:id` tidak cocok dengan route backend sejak awal)
- `src/lib/components/layouts/menu.ts` — field `roles`, `menuForRole()`
- `src/lib/components/layouts/{Sidebar,TopNav,Command}Layout.svelte`, `BottomBar.svelte` — filter menu

Infrastruktur:
- `docker/nginx.conf` — security headers, CORS wildcard dihapus
- `docker-compose.yml` — port 5432/6379 dihapus; `DB_USER`/`DB_PASSWORD` wajib (`:?`)

Tidak diubah: `.env` (lihat §5), skema database, migrasi, alur bisnis, layout UI.

---

## 4. Verifikasi yang dijalankan

| Perintah | Hasil |
|---|---|
| `go build ./...` (backend) | PASS |
| `go vet ./...` (backend) | PASS |
| `go test ./...` (backend) | PASS — usecase, middleware (CORS), config (validasi secret), postgres (migrate) |
| `npm run check` (dashboard) | PASS — svelte-check 0 errors, 0 warnings |
| `npm run build` (dashboard) | PASS |
| `flutter analyze` (mobile) | PASS — no issues |
| `flutter test` (mobile) | PASS — 2/2 |

Catatan: test Go yang baru (`cors_test.go`, `config_test.go`) dan test usecase yang diperbarui sudah lulus. Test rate limiter dan WebSocket **tidak ditulis** (butuh Redis/server hidup).

---

## 5. Blocker konfigurasi lokal (BUTUH KEPUTUSAN ANDA)

`.env` lokal saat ini memakai `JWT_ACCESS_SECRET` dan `JWT_REFRESH_SECRET` yang **mengandung nilai contoh** (`change-this-...`). Fail-fast 2.10 akan menolak start backend dengan `.env` ini.

- Tidak saya ubah: `.env` berisi secret dan brief melarang mengubah konfigurasi produksi.
- Tindakan yang diperlukan: ganti kedua secret dengan nilai acak ≥ 32 karakter yang berbeda satu sama lain. Contoh cara membuat (jalankan sendiri, nilai jangan dibagikan): `openssl rand -hex 32`.
- Catatan: `.env` memakai `ENV=production`, sedangkan kode membaca `APP_ENV`. Akibatnya pengecekan `DB_PASSWORD` untuk produksi **tidak aktif**. Jika ingin aktif, ganti nama variabel menjadi `APP_ENV` (keputusan Anda).

Saya tidak menampilkan nilai secret; hanya status (ada, panjang, ada penanda contoh).

---

## 6. NOT TESTED

Area berikut **belum diuji** karena butuh stack hidup atau browser:

- `docker compose up` end-to-end (postgres, redis, backend, dashboard, nginx)
- Handshake WebSocket nyata dari browser: sub-protocol, penolakan origin asing, penolakan user tanpa akses sesi
- Rate limiter terhadap Redis nyata (termasuk perilaku TTL dan fail-open saat Redis mati)
- CORS dari browser sungguhan (preflight, credentials)
- CSP di browser: apakah halaman SvelteKit tetap jalan dengan `'unsafe-inline'` dan `connect-src` yang dibatasi
- `nginx -t` (sintaks config nginx belum divalidasi)
- Role gate HTTP end-to-end (export 403 untuk staff; menu tersembunyi di browser)
- Registrasi HTTP: 409 untuk username duplikat, 400 untuk role tidak valid
- Import Excel nyata dengan >2000 baris dan pemetaan error 400
- Audit row di `activity_logs` pada database nyata
- Notifikasi dengan nama produk berkarakter khusus (kutip, backslash) pada database nyata

Belum dikerjakan dari daftar Tahap C: JWT `iss`/`aud` (2.7), transaksi import (2.11), nonce/hash CSP (2.9), `PUBLIC_API_BASE_URL`/`VITE_*` yang masih `localhost` di compose (PR-05, bukan item C).

---

## 7. Temuan baru selama Tahap C

- **Path WebSocket di client tidak pernah cocok dengan route backend** (`/stock-opname/...` vs `/ws/stock-opname/...`). Diperbaiki sebagai bagian 2.1.
- **Dua secret di `.env` lokal memakai nilai contoh.** Sudah dilaporkan di §5.
- **`APP_ENV` vs `ENV`** tidak konsisten antara `.env` dan kode (§5).

---

## 8. Langkah berikutnya

Tahap D tidak dimulai sampai Anda mengonfirmasi Tahap C. Sebelum itu, disarankan:
1. Ganti secret di `.env` lokal (§5) agar backend bisa start.
2. Jalankan `docker compose up` dan uji bagian §6 secara manual.
