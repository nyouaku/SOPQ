# Panduan Menjalankan Proyek di Lokal

Langkah uji lokal untuk **dashboard** (SvelteKit) + **backend** (Go) + **PostgreSQL**.
Ikuti berurutan. Perintah dijalankan dari root proyek `stock-opname-app/` kecuali disebut lain.

> Catatan: tutorial ini memakai port **5435** untuk PostgreSQL karena port 5432 di mesin ini
> sudah dipakai container lain. Kalau 5432 kosong di mesin Anda, ganti 5435 → 5432.

---

## 1. Jalankan PostgreSQL

Container `docker compose up` **tidak dipakai** di tutorial ini, karena folder
`backend/migrations` di-mount sebagai initdb dan berisi beberapa file yang error
(lihat bagian "Masalah yang diketahui"). Kita jalankan Postgres bersih lalu isi manual.

```bash
docker run -d --name so-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=stockopname_db \
  -p 5435:5432 \
  postgres:16-alpine
```

Tunggu ±5 detik sampai siap:

```bash
docker exec so-postgres pg_isready -U postgres
```

Harus keluar `accepting connections`. Kalau belum, tunggu dan ulangi.

---

## 2. Isi skema dan data awal

Jalankan 3 file ini **berurutan** (urutannya penting: role & user harus ada
sebelum produk, karena seed produk membaca `warehouses`).

**PowerShell:**

```powershell
Get-Content backend/migrations/000001_init_schema.up.sql -Raw | docker exec -i so-postgres psql -U postgres -d stockopname_db
Get-Content backend/migrations/seed_users.sql -Raw | docker exec -i so-postgres psql -U postgres -d stockopname_db
Get-Content backend/migrations/seed_products.sql -Raw | docker exec -i so-postgres psql -U postgres -d stockopname_db
```

**Git Bash / WSL** (jika memakai shell ini, bentuk `<` juga bisa):

```bash
docker exec -i so-postgres psql -U postgres -d stockopname_db < backend/migrations/000001_init_schema.up.sql
docker exec -i so-postgres psql -U postgres -d stockopname_db < backend/migrations/seed_users.sql
docker exec -i so-postgres psql -U postgres -d stockopname_db < backend/migrations/seed_products.sql
```

Cek hasilnya:

```bash
docker exec -i so-postgres psql -U postgres -d stockopname_db -c "select u.username, r.name as role from users u join roles r on r.id=u.role_id order by u.username;"
docker exec -i so-postgres psql -U postgres -d stockopname_db -c "select count(*) as jumlah_produk from products;"
```

Yang diharapkan: 3 user (`admin`, `staff`, `supervisor`) dan 20 produk.

**Akun login:**

| Username | Password | Role |
|---|---|---|
| `admin` | `staff123` | admin |
| `supervisor` | `staff123` | supervisor |
| `staff` | `staff123` | staff_gudang |

Ketiga user memakai hash bcrypt yang sama, sudah diverifikasi cocok dengan `staff123`.

---

## 3. Konfigurasi backend

File `backend/.env` sudah ada. Ubah **satu baris** port database agar cocok
dengan langkah 1:

```diff
- DB_PORT=5432
+ DB_PORT=5435
```

Sisanya biarkan apa adanya. Jangan gunakan `.env` root proyek (`ENV=production`,
`DB_HOST=postgres`) — itu untuk Docker Compose, bukan untuk lokal.

---

## 4. Jalankan backend

```bash
cd backend
go run cmd/api/main.go
```

Kalau berhasil akan muncul log server listening di `:8080`.
Biarkan terminal ini tetap terbuka.

Uji cepat dari terminal lain:

```bash
curl http://localhost:8080/api/v1/ping
# {"message":"pong"}

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"staff123"}'
# {"success":true,"data":{"access_token":"...","refresh_token":"...","user":{...}}}
```

Kalau login mengembalikan `access_token`, backend siap.

---

## 5. Konfigurasi dashboard

Folder `dashboard/` hanya punya `.env.example`, belum ada `.env`. Buat baru:

```bash
cd dashboard
cp .env.example .env
```

Isinya sudah benar untuk lokal:

```
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_BASE_URL=ws://localhost:8080/api/v1
```

---

## 6. Jalankan dashboard

```bash
cd dashboard
npm install
npm run dev
```

Buka **http://localhost:5173** lalu login dengan `admin` / `staff123`.

---

## 7. Yang bisa diuji

Halaman dan statusnya saat ini:

| Halaman | URL | Catatan |
|---|---|---|
| Login | `/login` | Migrasi ke token DaisyUI, tanpa `var(--color*)` |
| Dashboard | `/` | Kartu ringkasan + grafik; sebagian metrik dari backend masih 0 |
| Sessions | `/sessions` | Pagination sudah diperbaiki (`totalPages` dulu tidak dideklarasikan) |
| Detail sesi | `/sessions/:id` | Kolom kode produk sudah benar (`product_code`) |
| Products | `/products` | Kode & kategori produk sudah benar (`product_code`, `category.name`) |
| Reports | `/reports` | **Supervisor/admin saja**; hasil masih kosong (stub backend) |
| Scan | `/scan` | Perlu kamera; butuh sesi `in_progress` |
| Mobile | `/mobile` | Sudah dibungkus layout; zoom tidak lagi diblokir |
| Settings | `/settings` | Pemilih layout & tema warna |

Login sebagai `supervisor` atau `admin` untuk membuka **Reports**.

Untuk mencoba Scan: buat sesi dulu lewat menu **Sessions → New Session**
(pilih gudang "Gudang Utama"), baru buka **Scan**.

---

## 8. Ganti layout & tema

**Settings** menyediakan 4 layout dan 5 tema warna. Pilihan disimpan di `localStorage`
browser, bukan di server. Setelah ganti, klik menu apa saja untuk melihat layout baru.

| Layout | Perilaku |
|---|---|
| **Classic Sidebar** | Mobile (<1024px): default **tersembunyi**, buka lewat tombol ☰ di kiri atas; tutup lewat ✕ atau tap area gelap. Menutup otomatis setiap berpindah halaman. Desktop (≥1024px): selalu tampil, **tanpa** tombol hide. |
| **Modern Top Nav** | Navbar horizontal di atas; dropdown di layar sempit. |
| **Command Center** | Layout ops tiga kolom, ala terminal monitoring. |
| **Bottom Bar** | Mobile: bar navigasi di bawah layar (4 menu), navbar ringkas di atas. Desktop: bar pindah ke atas, di bawah navbar. Mobile-first. |

Tema warna mengubah palet seluruh aplikasi via `data-theme` (DaisyUI).

---

## 9. Verifikasi kualitas kode (opsional)

```bash
cd dashboard
npm run check     # 0 errors, 22 warnings (semuanya a11y: label tanpa for/id)
npm run build     # build produksi
```

```bash
cd backend
go build ./...    # harus exit 0
```

`go test ./...` masih gagal — itu **bukan** akibat perubahan UI. Penyebabnya file
test lama yang belum disesuaikan dengan signature usecase terbaru
(`domain.ProductFilter`, `config.JWTConfig.AccessExpiry`, dll). Perlu sesi terpisah.

---

## 10. Membersihkan

```bash
docker stop so-postgres
docker rm so-postgres
```

Data hilang karena tidak memakai volume. Untuk menyimpan data, tambahkan
`-v so-pgdata:/var/lib/postgresql/data` saat `docker run` di langkah 1.

---

## Masalah yang diketahui

Ini kondisi nyata kode saat ini, bukan kesalahan setup Anda:

1. **`seed_user.sql` jangan dijalankan.** File ini memakai kolom `users.password`
   yang tidak ada (skema memakai `password_hash`) dan role `staff` yang tidak ada
   (seed role memakai `staff_gudang`). Gunakan `seed_users.sql`.
2. **`performance_indexes.sql` error.** Membuat index dengan
   `WHERE deleted_at IS NULL` pada tabel `stock_opname` dan `sessions` yang
   tidak punya kolom `deleted_at`.
3. **`seed_products_listrik.sql` error.** Menulis kolom `description` dan
   `updated_at` pada tabel `categories`, yang tidak punya kedua kolom itu.
   Pakai `seed_products.sql`.
4. **WebSocket tidak terhubung dari `localhost:5173`.** Whitelist origin di
   `internal/delivery/websocket/stock_opname_ws.go` belum memuat port 5173.
   Progress real-time di detail sesi tidak akan update otomatis; data tetap
   benar setelah refresh halaman.
5. **Laporan kosong.** `report_usecase.go` masih stub yang mengembalikan slice
   kosong. Filter di halaman Reports terkirim ke backend tapi diabaikan.
6. **Export Excel hanya header.** `export_usecase.go` mendeklarasikan
   `sampleData` tapi tidak pernah mengisinya.
7. **Kolom "Stock Sistem" di Products selalu 0.** `domain.Product` tidak punya
   field stok; stok ada di tabel `stock` terpisah dan belum di-join.
8. **Angka selisih di detail sesi hanya menghitung halaman aktif.** Statistik
   Match/Over/Under dihitung dari array `details` yang sudah dipaginasi,
   bukan dari seluruh sesi.
9. **Role `staff_gudang` vs `staff`.** Middleware hanya memeriksa `admin` dan
   `supervisor`, jadi user `staff` tetap bisa scan. Tapi pengecekan role di
   frontend memakai `admin`, sehingga user `staff` tidak melihat tombol
   tambah/edit produk — sesuai harapan.
