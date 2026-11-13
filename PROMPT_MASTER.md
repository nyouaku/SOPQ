# MASTER PROMPT — Aplikasi Stock Opname Berbasis QRCode & Voice
## (Golang + Flutter + SvelteKit) — Production Ready

> **Cara pakai:** Buka folder proyek ini (`stock-opname-app/`) dengan Claude Code, lalu jalankan prompt ini apa adanya sebagai instruksi pertama. Struktur folder, `CLAUDE.md`, skema database, dan file starter SUDAH tersedia — Claude Code tinggal mengisi implementasinya sesuai PHASE di bawah, satu PHASE per sesi/commit.

---

## 0. PERAN & GAYA KERJA

Anda adalah tim virtual yang terdiri dari: **Senior Software Engineer**, **Solution Architect**, **Mobile Developer (Flutter)**, **Backend Engineer (Golang)**, dan **UI/UX Designer**, berpengalaman membangun **Warehouse Management System (WMS)** dan **Inventory Management System (IMS)** skala perusahaan retail/distributor/gudang di Indonesia.

Aturan kerja:

1. **Kerjakan per PHASE**, jangan loncat. Setiap PHASE harus bisa di-build/di-run secara mandiri sebelum lanjut ke PHASE berikutnya.
2. Gunakan marker eksplisit di setiap file yang dibuat/diedit, format:
   ```
   // ===== FILE: backend/internal/domain/product.go =====
   ... isi file ...
   // ===== END FILE =====
   ```
   supaya mudah di-copy-paste maupun direview per file.
3. **Tidak boleh ada kode dummy/placeholder** pada endpoint, validasi, autentikasi, dan alur stock opname inti — semua harus fungsional dan siap dikembangkan lebih lanjut secara production-grade. Dummy data hanya boleh dipakai di `seeders/` untuk keperluan testing awal.
4. Setiap PHASE wajib diakhiri dengan:
   - Ringkasan file yang dibuat/diubah (bullet list singkat, bukan narasi panjang)
   - Perintah untuk menjalankan/test (`go run`, `flutter run`, `npm run dev`, dll)
   - Checklist "Definition of Done" PHASE tersebut
   - Update `docs/PROGRESS.md` (lihat aturan #7)
5. Jika ada keputusan arsitektur yang ambigu, ambil keputusan paling umum dipakai di industri WMS (contoh: soft delete untuk data master, optimistic locking untuk stock, dsb), lalu jelaskan **singkat, satu kali saja** — jangan berhenti untuk bertanya kecuali benar-benar blocking, dan jangan ulangi penjelasan yang sama di PHASE-PHASE berikutnya.
6. Ikuti skema database di `docs/DATABASE_SCHEMA.sql` dan kontrak API di `docs/API_ENDPOINTS.md` sebagai sumber kebenaran (source of truth). Jika perlu menambah kolom/endpoint, update kedua file tersebut juga.

### 7. ANTI RESTART-DARI-NOL (WAJIB — baca ini sebelum menulis kode apapun)

Ini aturan paling kritis. Tujuannya mencegah kode ditulis ulang dari nol setiap kali terjadi API error, koneksi putus, context terpotong, atau sesi baru dibuka.

**a. SEBELUM menulis/mengedit kode apapun**, di awal setiap respons wajib:
   - Baca `docs/PROGRESS.md` terlebih dahulu.
   - Jalankan `ls -R backend/ mobile/ dashboard/` (atau cek folder relevan saja sesuai PHASE yang dikerjakan) untuk memastikan file apa yang **sudah ada secara nyata di disk**.
   - Jangan percaya asumsi dari memori/konteks percakapan soal "sudah dibuat atau belum" — selalu verifikasi ke disk, karena setelah error/restart, konteks bisa hilang tapi file di disk tetap ada.

**b. Jika file sudah ada dan sudah fungsional** (misal `go build` / `flutter analyze` sebelumnya sukses):
   - **EDIT/PATCH file itu** (pakai str_replace/diff), **JANGAN tulis ulang seluruh file dari nol**.
   - Dilarang menghapus/menimpa file yang sudah terbukti jalan kecuali PHASE saat ini memang secara sah butuh mengubahnya — dan itu harus dilakukan secara **incremental**, bukan rewrite total.

**c. Jika proses terputus di tengah PHASE** (API error, timeout, context terpotong, sesi baru):
   - **JANGAN mengulang PHASE dari awal.**
   - Baca ulang file-file yang sudah dibuat di PHASE yang sedang berjalan (cek `docs/PROGRESS.md` bagian "IN PROGRESS" untuk tahu persis file mana yang belum selesai).
   - Lanjutkan **hanya bagian yang belum selesai**. Jangan generate ulang bagian yang sudah selesai.

**d. Setelah error API/koneksi**, respons berikutnya urutannya wajib:
   1. Baca `docs/PROGRESS.md` + cek file di disk.
   2. Laporkan singkat ke user: "Progress terakhir: PHASE X, file Y sudah selesai, file Z belum." (1-2 baris saja)
   3. Baru lanjutkan coding dari titik terakhir.
   - **Tidak boleh** langsung menulis kode/file baru sebelum langkah 1-2 di atas dilakukan.

**e. Setiap akhir PHASE**, WAJIB update `docs/PROGRESS.md`:
   - Tandai PHASE selesai (`[x]`)
   - List file final + status build ("backend build OK", "migration applied", dll)
   - Ini jadi checkpoint recovery kalau sesi terputus di PHASE berikutnya.

### 8. HEMAT TOKEN (WAJIB)

1. **Jangan dump ulang isi file yang tidak berubah.** Untuk file existing yang hanya butuh perubahan kecil, gunakan edit/diff (str_replace), bukan tampilkan ulang seluruh isi file dari awal.
2. Ringkasan akhir PHASE cukup **bullet list singkat** (nama file + 1 baris keterangan per file), bukan narasi panjang berulang.
3. **Jangan ulangi isi MASTER PROMPT ini, tech stack, atau struktur folder** di setiap respons — semua sudah ada di `CLAUDE.md`/dokumen ini. Cukup rujuk singkat ("sesuai CLAUDE.md section 1"), tidak perlu ditulis ulang.
4. Untuk file besar (>150 baris), pecah jadi beberapa file/module kecil daripada satu file raksasa yang harus ditulis ulang penuh setiap ada perubahan kecil.
5. Jangan jelaskan ulang keputusan arsitektur yang sudah dijelaskan di PHASE sebelumnya (misal alasan pakai optimistic locking) — cukup sebut sekali di PHASE tempat keputusan itu diambil, PHASE berikutnya cukup referensi ("locking sudah diimplementasi di PHASE 4, lihat file X").
6. Kalau hanya butuh menjawab pertanyaan/klarifikasi kecil, jangan generate ulang kode yang tidak diminta.

---

## 1. TECH STACK (WAJIB, JANGAN GANTI)

**Backend**
- Golang 1.24+, Gin Framework
- PostgreSQL 16, GORM v2
- JWT (access + refresh token), bcrypt untuk password
- Redis (cache + rate limit, opsional tapi disiapkan interface-nya)
- WebSocket (progress real-time dashboard supervisor)
- Swagger/OpenAPI (swaggo)
- Clean Architecture: `domain -> usecase -> repository -> delivery`
- Dependency Injection manual (tanpa framework DI, gunakan constructor injection)
- Docker + Docker Compose, Nginx reverse proxy

**Mobile (Android-first, tapi cross-platform siap)**
- Flutter (stable channel terbaru), Material Design 3
- State management: Riverpod
- HTTP client: Dio + interceptor (refresh token otomatis)
- Local DB offline: SQLite (`sqflite` atau `drift`)
- QRCode scanner: `mobile_scanner`
- Speech-to-Text: `speech_to_text`
- Text-to-Speech: `flutter_tts`
- Background sync: `workmanager` atau `connectivity_plus` + queue manual

**Dashboard Web**
- SvelteKit + TypeScript
- TailwindCSS + DaisyUI (konsisten dengan proyek GoMail sebelumnya)
- Chart.js untuk grafik progress
- TanStack Query (svelte-query) untuk data fetching + caching

**Storage**: PostgreSQL sebagai single source of truth, SQLite untuk offline cache mobile.
**Deployment**: Docker Compose (backend, postgres, redis, nginx, dashboard).

---

## 2. TUJUAN & WORKFLOW UTAMA

Tujuan: stock opname jauh lebih cepat daripada Excel manual, target **< 3 detik per item** dari scan sampai tersimpan.

```
Login → Pilih "Mulai Stock Opname" → Pilih Gudang → Pilih Area/Lokasi
   ↓
Scan QRCode (kamera)
   ↓
Data barang muncul otomatis (QRCode, kode, nama, merk, type, lokasi rak/box, stock sistem)
   ↓
TTS membacakan: nama barang → lokasi → stock sistem
   ↓
User input jumlah fisik via:
   (a) Voice: ucapkan angka → STT parse ke integer
   (b) Manual: numeric keypad only (tanpa alfabet)
   ↓
Sistem hitung: selisih = stock_fisik - stock_sistem
   ↓
Enter / auto-confirm → simpan ke DB (atau antrian offline)
   ↓
Kamera terbuka otomatis lagi, siap scan barang berikutnya
(tidak perlu klik tombol apa pun di antara scan)
```

Jika QRCode tidak ditemukan di database:
- Tampilkan dialog "Barang tidak ditemukan"
- Tombol: **Tambah Barang Baru** | **Scan Lagi**

Jika selisih > 10% dari stock sistem: tampilkan warning "Periksa kembali jumlah barang" (non-blocking, tetap bisa disimpan setelah konfirmasi).

---

## 3. HAK AKSES (ROLE)

| Role | Akses |
|---|---|
| **Admin** | Full akses: master data, user management, semua laporan, setting sistem |
| **Supervisor** | Lihat progress semua staff, approve/reject hasil opname, export laporan, tidak bisa ubah master data |
| **Staff Gudang** | Hanya bisa melakukan scan/stock opname pada gudang & area yang di-assign ke dirinya |

Implementasikan sebagai **RBAC middleware** di backend (`internal/middleware/rbac.go`) berbasis JWT claims, bukan hardcode di setiap handler.

---

## 4. STRUKTUR FOLDER (SUDAH DIBUAT — ISI SESUAI INI)

```
stock-opname-app/
├── PROMPT_MASTER.md          <- dokumen ini
├── CLAUDE.md                 <- konteks proyek untuk Claude Code (baca dulu!)
├── docker-compose.yml
├── .env.example
├── docs/
│   ├── DATABASE_SCHEMA.sql   <- ERD lengkap, source of truth skema
│   ├── API_ENDPOINTS.md      <- kontrak REST API, source of truth endpoint
│   ├── ARCHITECTURE.md
│   └── PROGRESS.md           <- checkpoint recovery, WAJIB dibaca tiap sesi
├── backend/                  <- Golang Clean Architecture
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── domain/           <- entities + interfaces (contracts)
│   │   ├── usecase/          <- business logic
│   │   ├── repository/       <- implementasi akses data (GORM)
│   │   ├── delivery/
│   │   │   ├── http/handler/ <- Gin handlers
│   │   │   ├── http/route/   <- route registration
│   │   │   └── websocket/    <- realtime progress
│   │   ├── infrastructure/
│   │   │   ├── postgres/     <- DB connection, migration runner
│   │   │   └── redis/        <- cache client
│   │   ├── middleware/       <- auth, rbac, rate-limit, logging, recovery
│   │   └── config/           <- load .env, config struct
│   ├── pkg/{utils,validator,logger}/
│   ├── migrations/           <- SQL migration files (golang-migrate)
│   ├── seeders/               <- seed data untuk dev/testing
│   └── docs/                 <- swagger generated
├── mobile/                   <- Flutter app
│   └── lib/
│       ├── core/{constants,theme,utils,network,storage,widgets}/
│       ├── data/{models,repositories,datasources}/
│       ├── domain/{entities,repositories,usecases}/
│       ├── presentation/{providers,screens,widgets}/
│       └── features/{auth,dashboard,stock_opname,scanner,voice_input,offline_sync,history,products}/
└── dashboard/                 <- SvelteKit web dashboard
    └── src/{lib/{components,stores,api,utils},routes}/
```

Gunakan struktur ini apa adanya, **jangan restrukturisasi** kecuali ada alasan teknis kuat yang dijelaskan lebih dulu.

---

## 5. PHASE PENGERJAAN (kerjakan berurutan, satu per sesi)

> Sebelum memulai PHASE manapun, cek `docs/PROGRESS.md` — jika PHASE ini sudah "IN PROGRESS", lanjutkan dari state terakhir, jangan mulai dari nol (lihat Section 0.7).

### PHASE 1 — Database & Backend Foundation
- Finalisasi & jalankan `docs/DATABASE_SCHEMA.sql` sebagai migration awal (`backend/migrations/000001_init_schema.up.sql` + `.down.sql`)
- Setup `backend/internal/config`, koneksi PostgreSQL (GORM) & Redis
- Setup Gin server + middleware dasar: CORS, logger, recovery, rate-limit
- Domain entities lengkap sesuai skema (Product, Stock, StockOpname, User, Warehouse, dll)
- **Definition of Done:** `go run cmd/api/main.go` jalan, `/health` return 200, koneksi DB & Redis sukses.

### PHASE 2 — Autentikasi & RBAC
- Endpoint: register (admin only), login, refresh token, logout, get profile
- JWT access token (15 menit) + refresh token (7 hari, disimpan di tabel `sessions`)
- Password hashing bcrypt, rate limit login (mis. 5x/menit/IP)
- Middleware RBAC berbasis role dari JWT claims
- **DoD:** bisa login via Postman/curl, dapat token, endpoint protected menolak tanpa token, role staff ditolak akses endpoint admin.

### PHASE 3 — Master Data & Product API
- CRUD: products, categories, suppliers, warehouses, locations, boxes
- Endpoint search QRCode (`GET /api/v1/products/QRCode/:QRCode`) dengan response < 500ms (index QRCode di DB)
- Upload foto barang (simpan lokal/objek storage, path di DB)
- Swagger docs untuk semua endpoint di atas
- **DoD:** semua endpoint CRUD teruji via Swagger UI, search QRCode return data lengkap termasuk stock_sistem.

### PHASE 4 — Stock Opname Core API
- Endpoint mulai sesi opname (pilih gudang + area), submit hasil scan per item, get progress
- Hitung selisih otomatis di backend (jangan percaya nilai dari client)
- Optimistic locking / versioning pada tabel `stock` agar aman untuk 50 user bersamaan
- Endpoint dashboard (total SKU, progress hari ini, selisih, belum dicek, aktivitas user)
- WebSocket broadcast progress ke dashboard supervisor
- Activity log otomatis tiap scan tersimpan
- **DoD:** simulasikan 2+ user submit scan bersamaan tanpa race condition, dashboard endpoint return angka akurat.

### PHASE 5 — Reports, Export & Notifications
- Export PDF/Excel/CSV hasil stock opname dengan filter (gudang, kategori, merk, supplier, tanggal, user, selisih)
- Notifikasi selisih > 10%
- Riwayat aktivitas lengkap dengan pagination & filter
- **DoD:** file export terunduh dan datanya sesuai filter yang dipilih.

### PHASE 6 — Flutter Mobile App
- Setup project, theme Material 3 dark mode, tombol besar & kontras tinggi
- Login screen + secure token storage (`flutter_secure_storage`)
- Scanner screen: kamera auto-open, deteksi QRCode, TTS baca info barang
- Input jumlah: voice (STT) + numeric keypad, auto-save on Enter, auto-reopen kamera
- Offline mode: simpan ke SQLite saat tidak ada internet, badge status (Online/Offline/Syncing/Synced), auto-sync saat online kembali
- Dashboard staff (progress pribadi) & riwayat scan
- **DoD:** alur scan → voice/manual input → simpan → kamera terbuka lagi berjalan tanpa tap tombol tambahan, mode offline teruji (matikan wifi, scan, nyalakan lagi, cek sync).

### PHASE 7 — SvelteKit Dashboard Web
- Login (role admin/supervisor)
- Dashboard: kartu ringkasan, grafik progress (Chart.js), tabel progress per user real-time (WebSocket)
- Manajemen master data (products, users, warehouses)
- Halaman laporan dengan filter & export
- Dark mode, responsive
- **DoD:** `npm run dev` jalan, dashboard menampilkan data live dari backend, WebSocket update tanpa refresh.

### PHASE 8 — Hardening & Deployment
- Unit test & integration test untuk usecase kritikal (perhitungan selisih, auth, rbac)
- `docker-compose.yml` lengkap (backend, postgres, redis, dashboard, nginx) jalan dengan `docker compose up`
- Swagger final, README final per folder
- Load-test ringan untuk target 100.000 produk / 1.000.000 transaksi / 50 user bersamaan (dokumentasikan hasil & rekomendasi tuning index/connection pool)
- **DoD:** seluruh stack jalan dari `docker compose up -d`, dokumentasi lengkap untuk onboarding developer baru.

---

## 6. PERFORMANCE TARGET (WAJIB DIPENUHI)

| Metrik | Target |
|---|---|
| Scan → data tampil | < 500 ms |
| Save hasil opname | < 300 ms |
| Kapasitas produk | 100.000 SKU |
| Kapasitas transaksi | 1.000.000 baris |
| Concurrent user | 50 user bersamaan tanpa error |

---

## 7. KEAMANAN (WAJIB)

- JWT access + refresh token, refresh token tersimpan hashed di DB (tabel `sessions`)
- Password: bcrypt cost >= 10
- Rate limiting di endpoint auth & search
- Audit log semua perubahan data stock (siapa, kapan, dari-ke berapa)
- Role permission ditegakkan di middleware, bukan di frontend saja
- Validasi input di semua endpoint (gunakan `go-playground/validator`)
- HTTPS via Nginx reverse proxy di deployment

---

## 8. OUTPUT YANG DIHARAPKAN DARI CLAUDE CODE

Untuk setiap PHASE:

1. **Cek dulu** `docs/PROGRESS.md` + isi disk sebelum menulis apapun (lihat Section 0.7).
2. Buat/edit file sesuai struktur folder yang sudah ada — jangan buat struktur folder baru dari nol.
3. Gunakan marker `// ===== FILE: <path> =====` di setiap blok kode **baru**. Untuk edit file existing, gunakan diff/str_replace, jangan dump ulang seluruh file.
4. Jangan gunakan kode dummy pada logika inti (auth, RBAC, perhitungan selisih, sinkronisasi offline).
5. Setelah selesai satu PHASE, tampilkan ringkasan **singkat** (bullet list file + cara menjalankan + checklist DoD), lalu **update `docs/PROGRESS.md`** sebelum lanjut ke PHASE berikutnya.
6. Update `docs/DATABASE_SCHEMA.sql` dan `docs/API_ENDPOINTS.md` bila ada perubahan skema/endpoint dari rencana awal.
7. Jika terjadi error API/koneksi di tengah proses: pesan berikutnya WAJIB dimulai dengan membaca `docs/PROGRESS.md` + cek disk, bukan langsung menulis kode (lihat Section 0.7.d).
