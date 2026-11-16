# Panduan Fitur Stock Opname App

Dokumen ini menjelaskan semua fitur yang tersedia di aplikasi, fungsinya, dan cara menggunakannya.

---

## Daftar Isi

1. [Autentikasi & Akses](#1-autentikasi--akses)
2. [Dashboard](#2-dashboard)
3. [Sesi Stock Opname](#3-sesi-stock-opname)
4. [Scanner](#4-scanner)
5. [Manajemen Produk](#5-manajemen-produk)
6. [Manajemen Gudang](#6-manajemen-gudang)
7. [Laporan & Ekspor](#7-laporan--ekspor)
8. [Log Aktivitas](#8-log-aktivitas)
9. [Notifikasi & Peringatan](#9-notifikasi--peringatan)
10. [Rekonsiliasi Antar Gudang](#10-rekonsiliasi-antar-gudang)
11. [Perbandingan Stok Snapshot](#11-perbandingan-stok-snapshot)
12. [Pindah Rak (Transfer)](#12-pindah-rak-transfer)
13. [Penyesuaian Stok Manual](#13-penyesuaian-stok-manual)
14. [Grafik Riwayat Stok](#14-grafik-riwayat-stok)
15. [Cetak Label Barcode](#15-cetak-label-barcode)
16. [Cetak Thermal Printer](#16-cetak-thermal-printer)
17. [Import dari Excel/CSV](#17-import-dari-excelcsv)
18. [Layout Editor Gudang](#18-layout-editor-gudang)
19. [Pengaturan Tampilan](#19-pengaturan-tampilan)
20. [Pengaturan Suara](#20-pengaturan-suara)
21. [Manajemen Pengguna (Admin)](#21-manajemen-pengguna-admin)
22. [PWA (Progressive Web App)](#22-pwa-progressive-web-app)
23. [Mode Offline](#23-mode-offline)
24. [Keamanan & RBAC](#24-keamanan--rbac)

---

## 1. Autentikasi & Akses

**Fungsi:** Mengontrol siapa yang bisa masuk ke sistem dan apa yang boleh mereka lakukan.

**Cara pakai:**
1. Buka aplikasi → halaman login muncul otomatis.
2. Masukkan **username** dan **password**.
3. Klik **Login**.

Sistem menggunakan JWT (JSON Web Token) dengan dua token:
- **Access token** — berlaku 15 menit, dipakai untuk setiap request.
- **Refresh token** — berlaku 7 hari, dipakai untuk memperbarui access token otomatis.

**Role yang tersedia:**

| Role | Akses |
|------|-------|
| Admin | Semua fitur: master data, user management, laporan, pengaturan |
| Supervisor | Lihat semua sesi, approve/reject hasil, export laporan |
| Staff Gudang | Hanya scan pada gudang yang di-assign |

**Endpoint terkait:**
```
POST /api/v1/auth/login          — login
POST /api/v1/auth/refresh        — refresh token
POST /api/v1/auth/logout         — logout
GET  /api/v1/auth/me             — profil user saat ini
```

---

## 2. Dashboard

**Fungsi:** Ringkasan real-time semua aktivitas stock opname yang sedang berjalan.

**Cara pakai:**
1. Login → halaman utama (Dashboard) muncul otomatis.
2. Lihat kartu ringkasan: total SKU, progress hari ini, total selisih, belum dicek.
3. Grafik progress ditampilkan per gudang.
4. Data diperbarui otomatis via WebSocket tanpa refresh halaman.

**Endpoint terkait:**
```
GET /api/v1/dashboard/summary    — ringkasan angka dashboard
```

---

## 3. Sesi Stock Opname

**Fungsi:** Membuat, menjalankan, menyelesaikan, dan menyetujui sesi stock opname.

### 3a. Mulai Sesi Baru
1. Buka menu **Sessions** → klik **New Session**.
2. Pilih **Gudang** (wajib).
3. Pilih **Lokasi/Area** (opsional — kosongkan untuk seluruh gudang).
4. Klik **Mulai**.
5. Sistem membuat kode sesi otomatis (format: `SO-20261116-0001`).

### 3b. Jalankan Sesi
1. Buka sesi yang statusnya `in_progress`.
2. Klik **Scan** → arahkan kamera ke QR code produk.
3. Data produk muncul otomatis (nama, stok sistem, lokasi rak).
4. Masukkan **jumlah fisik** yang ada di rak.
5. Klik **Simpan** — kamera otomatis aktif lagi untuk scan berikutnya.

### 3c. Selesaikan Sesi
1. Setelah semua produk discan, klik **Complete Session**.
2. Status berubah menjadi `completed`.
3. Dialog ringkasan muncul: total scan, cocok, selisih, durasi.

### 3d. Setujui Sesi (Supervisor/Admin)
1. Buka sesi yang statusnya `completed`.
2. Review hasil di tab Scan Details.
3. Klik **Approve** — stok sistem diperbarui dengan stok fisik.

**Endpoint terkait:**
```
POST /api/v1/stock-opname/sessions              — mulai sesi baru
POST /api/v1/stock-opname/sessions/:id/scan     — submit hasil scan
POST /api/v1/stock-opname/sessions/:id/complete — selesaikan sesi
POST /api/v1/stock-opname/sessions/:id/approve  — setujui sesi
GET  /api/v1/stock-opname/sessions/:id/summary  — ringkasan sesi
```

---

## 4. Scanner

**Fungsi:** Membaca barcode/QR code produk untuk input jumlah fisik.

**Cara pakai:**
1. Buka halaman **Scan**.
2. Pilih sesi aktif dari dropdown di atas.
3. Pilih mode scan di tab:

| Mode | Cara Kerja |
|------|-----------|
| **Manual** | Ketik kode produk → Enter |
| **Kamera** | Klik "Aktifkan kamera" → arahkan ke QR code → scan otomatis tiap 0.5 detik |
| **Upload** | Upload foto QR code dari galeri/file |

4. Setelah QR terdeteksi: suara berbunyi, produk ditemukan, kamera berhenti.
5. Masukkan jumlah fisik di kolom yang muncul.
6. Klik **Simpan Data** — data tersimpan, kamera aktif lagi untuk scan berikutnya.

**Keunggulan:**
- Suara "beep" saat QR terdeteksi dan saat data tersimpan.
- Kamera otomatis berhenti saat produk ditemukan (fokus input).
- Kamera otomatis aktif lagi setelah simpan (siap scan berikutnya).
- Ukuran kamera proporsional (tidak memenuhi layar).
- Flash toggle & switch camera tersedia.

**Endpoint terkait:**
```
GET /api/v1/products/code/:code    — cari produk by barcode
POST /api/v1/stock-opname/sessions/:id/scan — simpan hasil scan
```

---

## 5. Manajemen Produk

**Fungsi:** CRUD produk, cari produk, lihat detail & riwayat stok.

**Cara pakai:**
1. Buka menu **Products**.
2. Gunakan kolom pencarian untuk cari by nama/kode.
3. Klik **Tambah Produk** (admin only) untuk menambah produk baru.
4. Klik ikon **Edit** untuk ubah data produk.
5. Klik ikon **Hapus** untuk menghapus produk.
6. Klik ikon **Grafik** untuk melihat riwayat stok produk.
7. Klik **Import** untuk tambah produk massal dari Excel.

**Detail Produk** (klik ikon grafik):
- Menampilkan nama, kode, brand, stok sistem.
- Grafik riwayat stok (7/30/90 hari atau semua).
- Data dari hasil opname sebelumnya.

**Endpoint terkait:**
```
GET    /api/v1/products                 — daftar produk (pagination)
GET    /api/v1/products/:id             — detail produk
GET    /api/v1/products/code/:code      — cari by barcode
POST   /api/v1/products                 — tambah produk
PUT    /api/v1/products/:id             — update produk
DELETE /api/v1/products/:id             — hapus produk
GET    /api/v1/products/:id/stock-history — riwayat stok
POST   /api/v1/products/transfer        — pindah rak
```

---

## 6. Manajemen Gudang

**Fungsi:** CRUD gudang, lokasi/area, dan rak/box secara hierarkis.

**Hierarki:**
```
Gudang (Warehouse)
  └── Lokasi/Area (Location)
        └── Rak/Box (Box)
```

**Cara pakai:**
1. Buka menu **Gudang**.
2. Daftar gudang tampil dalam bentuk accordion.
3. Klik gudang untuk membuka daftar lokasinya.
4. Klik lokasi untuk membuka daftar rak di dalamnya.

**Aksi (admin only):**
- **Tambah Gudang** — tombol di kanan atas.
- **Edit** — ikon pensil di samping item.
- **Hapus** — ikon tempat sampah (dengan konfirmasi).
- **Tambah Lokasi** — tombol di dalam gudang yang terbuka.
- **Tambah Rak** — tombol di dalam lokasi yang terbuka.
- **Layout Editor** — ikon grid untuk atur posisi rak secara visual.

**Endpoint terkait:**
```
GET    /api/v1/warehouses                     — daftar gudang
POST   /api/v1/warehouses                     — tambah gudang
PUT    /api/v1/warehouses/:id                 — update gudang
DELETE /api/v1/warehouses/:id                 — hapus gudang
GET    /api/v1/warehouses/:id/locations       — daftar lokasi
POST   /api/v1/warehouses/:id/locations       — tambah lokasi
GET    /api/v1/locations/:id/boxes            — daftar rak
POST   /api/v1/locations/:id/boxes            — tambah rak
GET    /api/v1/warehouses/:id/layout          — ambil layout
PUT    /api/v1/warehouses/:id/layout          — simpan layout
```

---

## 7. Laporan & Ekspor

**Fungsi:** Membuat laporan hasil stock opname dengan filter dan export ke berbagai format.

**Cara pakai:**
1. Buka menu **Reports**.
2. Set filter: gudang, brand, supplier, rentang tanggal, min/max selisih.
3. Klik **Terapkan** untuk melihat hasil.
4. Klik **Export CSV** atau **Export Excel** untuk unduh.
5. Klik **Export PDF** untuk laporan siap cetak.

**Endpoint terkait:**
```
GET /api/v1/reports/stock-opname         — data laporan dengan filter
GET /api/v1/reports/stock-opname/export  — export file
GET /api/v1/export/products              — export daftar produk
```

---

## 8. Log Aktivitas

**Fungsi:** Melihat riwayat semua tindakan pengguna di sistem (audit trail).

**Cara pakai:**
1. Buka menu **Activity**.
2. Filter by: jenis aksi, jenis entitas.
3. Navigasi halaman dengan tombol pagination.
4. Setiap entri menampilkan: user, aksi, waktu, detail perubahan.

**Aksi yang dicatat otomatis:**
- Login / Logout
- Mulai opname / Selesaikan opname / Setujui opname
- Scan produk
- Tambah / Ubah / Hapus produk
- Tambah / Ubah / Hapus gudang, lokasi, rak
- Penyesuaian stok manual
- Reset password user

**Endpoint terkait:**
```
GET /api/v1/activity-logs?page=1&limit=25 — daftar log aktivitas
```

---

## 9. Notifikasi & Peringatan

**Fungsi:** Menerima pemberitahuan real-time tentang peristiwa penting.

**Jenis notifikasi:**
- **Selisih > 10%** — peringatan saat ada produk dengan selisih besar.
- **Sesi selesai** — notifikasi saat sesi opname diselesaikan.
- **Sync offline** — pemberitahuan saat scan offline berhasil dikirim.

**Cara pakai:**
1. Lihat ikon lonceng di navbar — angka merah menunjukkan notifikasi belum dibaca.
2. Klik ikon lonceng → dropdown daftar notifikasi muncul.
3. Klik notifikasi untuk menandai sudah dibaca.
4. Klik "Tandai semua" untuk menandai semua sudah dibaca.

**Real-time alerts:**
- Saat scan dengan selisih > 10%, toast merah muncul di sesi detail.
- Saat sesi diselesaikan, dialog ringkasan muncul otomatis.

**Endpoint terkait:**
```
GET   /api/v1/notifications?limit=20         — daftar notifikasi
PATCH /api/v1/notifications/:id/read         — tandai sudah dibaca
```

---

## 10. Rekonsiliasi Antar Gudang

**Fungsi:** Membandingkan stok produk di beberapa gudang sekaligus. Deteksi produk yang hilang dari gudang tertentu.

**Cara pakai:**
1. Buka menu **Rekonsiliasi**.
2. Pilih minimal 2 gudang yang ingin dibandingkan.
3. Klik **Rekonsiliasi**.
4. Hasil menampilkan:
   - Kode & nama produk
   - Total stok di semua gudang terpilih
   - Berapa gudang yang memiliki produk tersebut
   - Gudang mana yang tidak memiliki produk (indikasi salah pindah)

**Endpoint terkait:**
```
POST /api/v1/reconciliation    — rekonsiliasi multi-gudang
```

---

## 11. Perbandingan Stok Snapshot

**Fungsi:** Membandingkan stok saat ini dengan stok N hari lalu.

**Cara pakai:**
1. Buka menu **Snapshot** (atau akses langsung `/snapshots`).
2. Pilih periode: 7 hari, 30 hari, atau 90 hari lalu.
3. Filter by gudang (opsional).
4. Lihat kartu ringkasan: total naik, turun, stabil.
5. Tabel detail menunjukkan per produk: stok snapshot, stok sekarang, perubahan, status.

**Status produk:**
- **Naik** (hijau) — stok bertambah
- **Turun** (merah) — stok berkurang
- **Kritis** (merah) — stok di bawah minimum
- **Stabil** (abu-abu) — tidak berubah

**Endpoint terkait:**
```
GET /api/v1/snapshots/compare?days=30&warehouse_id=xxx
```

---

## 12. Pindah Rak (Transfer)

**Fungsi:** Memindahkan beberapa produk ke rak/box yang berbeda dalam satu operasi.

**Cara pakai:**
1. Buka menu **Transfer**.
2. Pilih **Rak Tujuan**: Gudang → Lokasi → Rak.
3. Pilih produk yang ingin dipindahkan (centang dari daftar).
4. Klik **Pindahkan**.
5. Produk berpindah ke rak tujuan.

**Catatan:** Stok tidak berubah, hanya lokasi fisik yang di-update.

**Endpoint terkait:**
```
POST /api/v1/products/transfer    — transfer produk ke rak tujuan
```

---

## 13. Penyesuaian Stok Manual

**Fungsi:** Mengubah stok sistem secara manual dengan alasan wajib. Untuk koreksi stok rusak, hilang, atau salah input.

**Cara pakai:**
1. Akses via API: `POST /api/v1/stocks/adjust`
2. Kirim data:
   ```json
   {
     "product_id": "uuid-produk",
     "warehouse_id": "uuid-gudang",
     "delta": -5,
     "reason": "damaged",
     "note": "5 unit rusak karena terjatuh"
   }
   ```
3. Stok sistem diperbarui. Semua perubahan tercatat di activity log.

**Alasan yang tersedia:**
| Kode | Arti |
|------|------|
| `damaged` | Produk rusak |
| `lost` | Produk hilang |
| `expired` | Produk kedaluwarsa |
| `found` | Stok tambahan ditemukan |
| `typo_fix` | Koreksi salah input sistem |
| `other` | Lainnya (wajib isi catatan) |

**Hak akses:** Supervisor & Admin.

**Endpoint terkait:**
```
POST /api/v1/stocks/adjust    — penyesuaian stok manual
```

---

## 14. Grafik Riwayat Stok

**Fungsi:** Menampilkan grafik line chart riwayat stok fisik vs stok sistem per produk.

**Cara pakai:**
1. Buka **Products** → klik ikon **grafik** pada produk yang diinginkan.
2. Halaman detail produk terbuka dengan grafik.
3. Pilih rentang waktu: 7 hari, 30 hari, 90 hari, atau Semua.
4. Grafik menampilkan 2 garis:
   - **Garis oranye** — stok fisik (hasil opname)
   - **Garis putus-putus teal** — stok sistem
5. Hover pada titik data untuk melihat detail (kode sesi, selisih).

**Endpoint terkait:**
```
GET /api/v1/products/:id/stock-history?from=...&to=...&limit=100
```

---

## 15. Cetak Label Barcode

**Fungsi:** Membuat label QR code dalam format PDF untuk produk yang belum ada barcode fisik.

**Cara pakai:**
1. Buka **Products** → klik **Print Labels**.
2. Konfigurasi ukuran label (lebar, tinggi, kolom per baris).
3. Pilih produk yang ingin dicetak (centang dari daftar).
4. Klik **Generate PDF** — file PDF terunduh.
5. Buka PDF → print dengan ukuran label yang sesuai.

**Endpoint terkait:**
Tidak ada — PDF di-generate di sisi frontend menggunakan jsPDF + qrcode library.

---

## 16. Cetak Thermal Printer

**Fungsi:** Cetak label langsung ke printer thermal (58mm/80mm) tanpa PDF.

**Cara pakai:**
1. Buka **Products** → **Print Labels**.
2. Pilih produk yang ingin dicetak.
3. Di bagian "Cetak Langsung ke Printer Thermal", pilih ukuran kertas:
   - 58mm — untuk mini printer (mobile thermal)
   - 80mm — untuk standard thermal printer
4. Klik **Cetak Thermal**.
5. Window baru terbuka dengan layout label → dialog print muncul otomatis.
6. Pilih printer thermal yang terpasang → Print.

**Catatan:** Printer thermal harus terpasang dan dikenali oleh sistem operasi.

---

## 17. Import dari Excel/CSV

**Fungsi:** Menambah atau memperbarui produk massal dari file Excel.

**Cara pakai:**
1. Buka **Products** → klik **Import** (admin only).
2. Download template: klik **Download Template** — file Excel dengan format kolom yang benar.
3. Isi template dengan data produk (kode, nama, brand, dll).
4. Upload file Excel yang sudah diisi.
5. Sistem memproses: produk yang sudah ada di-update, produk baru ditambahkan.
6. Lihat hasil: jumlah berhasil, gagal, dan detail error per baris.

**Format file:** `.xlsx` (Excel 2007+), maksimal 5 MB.

**Endpoint terkait:**
```
GET  /api/v1/import/products/template   — download template
POST /api/v1/import/products            — upload & import
```

---

## 18. Layout Editor Gudang

**Fungsi:** Editor visual untuk mengatur posisi rak di dalam gudang menggunakan drag-and-drop.

**Cara pakai:**
1. Buka menu **Gudang** → klik ikon **grid** pada gudang yang diinginkan.
2. Halaman Layout Editor terbuka dengan grid canvas.
3. Drag rak untuk mengatur posisi (snap ke grid).
4. Klik **Tambah Rak** untuk menambah rak baru.
5. Klik **×** pada rak untuk menghapus dari layout.
6. Klik **Simpan Layout** — posisi tersimpan di database.

**Format penyimpanan:** JSON di kolom `layout_json` pada tabel `warehouses`.

**Endpoint terkait:**
```
GET /api/v1/warehouses/:id/layout    — ambil layout tersimpan
PUT /api/v1/warehouses/:id/layout    — simpan layout
```

---

## 19. Pengaturan Tampilan

**Fungsi:** Mengatur tema warna, layout navigasi, dan mode terang/gelap.

**Cara pakai:**
1. Buka menu **Settings**.
2. **Layout** — pilih gaya navigasi:
   - Classic Sidebar (📋) — sidebar tradisional
   - Modern Top Nav (🎯) — navbar horizontal
   - Command Center (🖥️) — layout split-panel ops (default)
   - Bottom Bar (📱) — navigasi bawah, mobile-first
3. **Mode Tema**:
   - **Manual** — pilih warna tetap dari 5 tema tersedia
   - **Otomatis** — tema berganti otomatis berdasarkan jam (gelap malam, terang siang)
4. Jika mode Otomatis: atur warna siang, warna malam, jam mulai malam & siang.
5. Perubahan langsung diterapkan tanpa refresh.

**Tema warna tersedia:**

| Tema | Deskripsi |
|------|-----------|
| Rack Room | Dark NOC dengan teal |
| NOC Console | Terminal hijau |
| Daylight Ops | Terang untuk siang hari |
| Saweria | Oranye hangat (default) |
| Modern List | Ungu gradient bersih |

**Cara pakai:**
Buka **Settings** → bagian **Tampilan** → pilih layout & warna.

---

## 20. Pengaturan Suara

**Fungsi:** Mengatur efek suara untuk notifikasi scan, simpan, dan error.

**Cara pakai:**
1. Buka **Settings** → bagian **Suara & Notifikasi**.
2. Toggle on/off untuk setiap jenis suara:
   - **Suara Scan** — saat QR code terdeteksi
   - **Suara Berhasil** — saat data berhasil disimpan
   - **Suara Error** — saat terjadi kesalahan
3. Atur **Master Volume** dengan slider (0-100%).
4. Klik tombol **Test** pada setiap suara untuk mendengar preview.

**Catatan:** Suara menggunakan Web Audio API (synthesized tones), tidak perlu file audio eksternal. Pengaturan tersimpan di localStorage browser.

**Cara pakai:**
Buka **Settings** → bagian **Suara & Notifikasi** → atur toggle & volume.

---

## 21. Manajemen Pengguna (Admin)

**Fungsi:** Admin dapat mengelola akun pengguna: tambah staff, aktif/nonaktifkan, reset password.

**Cara pakai:**
1. Login sebagai **admin**.
2. Buka menu **Settings** → bagian **Staff**.
3. **Tambah staff**: klik "Tambah staff" → isi username, nama lengkap, password, role → Simpan.
4. **Aktif/Nonaktifkan**: klik ikon ⊘ pada staff → konfirmasi → status berubah.
5. **Reset password**: klik ikon kunci → konfirmasi → password baru muncul di modal → salin & bagikan ke user.

**Catatan:**
- Admin tidak bisa menonaktifkan akun sendiri.
- Reset password akan mencabut semua sesi aktif user tersebut.
- Password baru di-generate otomatis (12 karakter alfanumerik).

**Endpoint terkait:**
```
GET   /api/v1/auth/users                        — daftar user
GET   /api/v1/auth/roles                        — daftar role
POST  /api/v1/auth/register                     — tambah user
PATCH /api/v1/auth/users/:id/toggle-status       — aktif/nonaktif
POST  /api/v1/auth/users/:id/reset-password      — reset password
```

---

## 22. PWA (Progressive Web App)

**Fungsi:** Dashboard dapat dipasang di HP sebagai aplikasi native (Add to Home Screen).

**Cara pakai:**
1. Buka dashboard di browser HP (Chrome/Edge).
2. Banner "Pasang Stock Opname" muncul di bawah layar.
3. Klik **Pasang** → aplikasi terpasang di home screen.
4. Buka dari home screen — tampilan full-screen tanpa address bar.

**Keunggulan PWA:**
- Akses cepat dari home screen.
- Bekerja offline untuk halaman yang sudah dikunjungi.
- Update otomatis saat online.
- Icon & splash screen seperti aplikasi native.

**Catatan:** PWA hanya tersedia di HTTPS (production). Di localhost, fitur install tidak muncul.

---

## 23. Mode Offline

**Fungsi:** Scan tetap bisa dilakukan tanpa koneksi internet. Data disimpan lokal dan dikirim saat online kembali.

**Cara pakai:**
1. Saat offline, lakukan scan seperti biasa.
2. Data masuk ke **antrian offline** (disimpan di browser).
3. Badge "X pending" muncul di navbar.
4. Saat koneksi kembali, data otomatis dikirim ke server.
5. Badge hilang saat semua data tersinkron.

**Indikator:**
- 🟢 **Online** — koneksi aktif
- ⚪ **Offline** — tidak ada koneksi, scan masuk antrian
- 🟠 **X pending** — X scan menunggu dikirim

**Catatan:** Offline mode tersedia di halaman Scan (dashboard) dan aplikasi mobile Flutter.

---

## 24. Keamanan & RBAC

**Fungsi:** Mengontrol akses berdasarkan role dan melindungi data dari akses tidak sah.

### Rate Limiting
| Endpoint | Limit | Per |
|----------|-------|-----|
| Login | 10x / 15 menit | IP + username |
| Refresh token | 30x / 15 menit | IP |
| Products (QR search) | 60x / menit | IP |
| Import | 5x / 5 menit | IP |
| Stock adjustment | 20x / menit | IP |

### RBAC
| Endpoint Group | Admin | Supervisor | Staff |
|---------------|-------|-----------|-------|
| Products CRUD | ✅ | ❌ | ❌ |
| Products scan | ✅ | ✅ | ✅ |
| Warehouses CRUD | ✅ | ❌ | ❌ |
| Sessions approve | ✅ | ✅ | ❌ |
| Stock adjustment | ✅ | ✅ | ❌ |
| Reconciliation | ✅ | ✅ | ❌ |
| User management | ✅ | ❌ | ❌ |
| Reports | ✅ | ✅ | ❌ |

### Keamanan lainnya
- **Password hashing** — bcrypt cost 12
- **JWT dual token** — access (15 min) + refresh (7 hari)
- **CORS** — whitelist origin + wildcard subdomain support
- **Security headers** — X-Content-Type-Options, X-Frame-Options, CSP
- **HTTPS** — konfigurasi Let's Encrypt tersedia di nginx.conf
- **Input validation** — go-playground/validator di semua endpoint
- **Audit trail** — semua perubahan data tercatat di activity_logs

---

## Ringkasan Endpoint API

| Metode | Endpoint | Fungsi | Role |
|--------|----------|--------|------|
| POST | `/auth/login` | Login | Public |
| POST | `/auth/refresh` | Refresh token | Public |
| POST | `/auth/logout` | Logout | Authenticated |
| GET | `/auth/me` | Profil | Authenticated |
| GET | `/auth/users` | Daftar user | Admin |
| GET | `/auth/roles` | Daftar role | Admin |
| POST | `/auth/register` | Tambah user | Admin |
| PATCH | `/auth/users/:id/toggle-status` | Aktif/nonaktif | Admin |
| POST | `/auth/users/:id/reset-password` | Reset password | Admin |
| GET | `/products` | Daftar produk | Authenticated |
| GET | `/products/:id` | Detail produk | Authenticated |
| GET | `/products/code/:code` | Cari by barcode | Authenticated |
| GET | `/products/:id/stock-history` | Riwayat stok | Authenticated |
| POST | `/products` | Tambah produk | Admin |
| PUT | `/products/:id` | Update produk | Admin |
| DELETE | `/products/:id` | Hapus produk | Admin |
| POST | `/products/transfer` | Pindah rak | Admin |
| GET | `/warehouses` | Daftar gudang | Authenticated |
| POST | `/warehouses` | Tambah gudang | Admin |
| PUT | `/warehouses/:id` | Update gudang | Admin |
| DELETE | `/warehouses/:id` | Hapus gudang | Admin |
| GET | `/warehouses/:id/layout` | Ambil layout | Authenticated |
| PUT | `/warehouses/:id/layout` | Simpan layout | Admin |
| GET | `/stock-opname/sessions` | Daftar sesi | Authenticated |
| POST | `/stock-opname/sessions` | Mulai sesi | Authenticated |
| GET | `/stock-opname/sessions/:id` | Detail sesi | Authenticated |
| GET | `/stock-opname/sessions/:id/details` | Detail scan | Authenticated |
| GET | `/stock-opname/sessions/:id/summary` | Ringkasan sesi | Authenticated |
| POST | `/stock-opname/sessions/:id/scan` | Submit scan | Authenticated |
| POST | `/stock-opname/sessions/:id/complete` | Complete | Authenticated |
| POST | `/stock-opname/sessions/:id/approve` | Approve | Supervisor+ |
| GET | `/stock-opname/sessions/:id/summary` | Summary | Authenticated |
| GET | `/dashboard/summary` | Dashboard | Authenticated |
| GET | `/reports/stock-opname` | Laporan | Authenticated |
| GET | `/reports/stock-opname/export` | Export | Authenticated |
| POST | `/stocks/adjust` | Adjust stok | Supervisor+ |
| POST | `/reconciliation` | Rekonsiliasi | Supervisor+ |
| GET | `/snapshots/compare` | Snapshot | Authenticated |
| GET | `/activity-logs` | Log aktivitas | Authenticated |
| GET | `/notifications` | Notifikasi | Authenticated |
| PATCH | `/notifications/:id/read` | Mark read | Authenticated |
| GET | `/import/products/template` | Template import | Admin |
| POST | `/import/products` | Import produk | Admin |
| GET | `/ws/stock-opname/sessions/:id` | WebSocket | Authenticated |
