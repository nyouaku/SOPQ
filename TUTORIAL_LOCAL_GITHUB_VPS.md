# Tutorial: Lokal → GitHub → VPS

Panduan lengkap memindahkan proyek dari mesin lokal ke VPS yang **sudah menjalankan aplikasi lain**, dengan Cloudflare sebagai TLS.

Tiga tahap:

1. [Tahap 1 — Rapikan lokal](#tahap-1--rapikan-lokal)
2. [Tahap 2 — Push ke GitHub](#tahap-2--push-ke-github)
3. [Tahap 3 — Deploy ke VPS](#tahap-3--deploy-ke-vps)

---

## Prasyarat

| Di mana | Butuh |
|---|---|
| Lokal | Git, Go 1.25, Node 22+, Docker Desktop |
| VPS | Docker + Docker Compose v2, SSH key, port **8088** (atau port bebas lain) |
| Cloudflare | Domain aktif, akses DNS |

Cek versi VPS:

```bash
ssh user@IP_VPS
docker --version && docker compose version
free -h && df -h /           # minimal 2 GB RAM, 10 GB disk kosong
ss -tlnp | grep -E ':(80|443|8088)\b'   # pastikan 8088 belum dipakai
```

---

## Tahap 1 — Rapikan lokal

### 1.1 Pastikan test hijau

Semua perintah dari root proyek.

```bash
cd backend
go build ./... && go vet ./... && go test ./...

cd ../dashboard
npm ci
npm run check
npm test
npm run build
```

Kalau ada yang merah, berhenti dan perbaiki dulu.

### 1.2 Pastikan `.env` tidak ikut ter-commit

`.gitignore` sudah memuat `.env`. Verifikasi:

```bash
git check-ignore -v .env
# harus keluar: .gitignore:2:.env
```

Kalau kosong, tambahkan `.env` ke `.gitignore` **sebelum** `git add`.

### 1.3 Jangan commit file besar / data asli

File yang **harus** di luar repo:

- `STOCK CMI RB19 2026.xlsx` — data stok asli
- `dashboard/node_modules/`, `dashboard/build/`, `dashboard/.svelte-kit/`
- `mobile/build/`, `mobile/.dart_tool/`
- `backend/uploads/`, `backend/server`, `backend/tmp/`

Tambah yang belum tercakup:

```bash
printf '\n# Data asli\n*.xlsx\n!docs/*.xlsx\n' >> .gitignore
```

### 1.4 Inisialisasi git

```bash
git init -b main
git add -A
git status --short | head -40     # tinjau — tidak boleh ada .env atau .xlsx
git commit -m "Initial commit: stock opname app"
```

---

## Tahap 2 — Push ke GitHub

### 2.1 Buat repository kosong

Di GitHub → **New repository** → nama `stock-opname-app` → **Private** → jangan centang "Add a README".

Jangan pakai HTTPS + password; GitHub sudah tidak menerima password. Pilih salah satu:

**A. SSH (disarankan)**

```bash
ssh-keygen -t ed25519 -C "lokal" -f ~/.ssh/id_ed25519 -N ""
cat ~/.ssh/id_ed25519.pub     # tempel di GitHub → Settings → SSH and GPG keys → New SSH key

ssh -T git@github.com          # harus: Hi <user>! You've successfully authenticated
git remote add origin git@github.com:USER/stock-opname-app.git
```

**B. Personal Access Token**

Buat token di GitHub → Settings → Developer settings → Tokens (classic) → scope `repo`. Lalu:

```bash
git remote add origin https://github.com/USER/stock-opname-app.git
# saat push, username = USER, password = token
```

### 2.2 Push

```bash
git push -u origin main
```

### 2.3 CI otomatis

`.github/workflows/ci.yml` sudah ada. Setiap push, GitHub Actions menjalankan:

- `go build`, `go vet`, `go test`, `govulncheck`, `staticcheck`, `gosec`
- `npm ci`, `npm run check`, `npm test`, `npm run build`, `npm audit`

Cek tab **Actions**. **Jangan lanjut ke VPS sebelum semua hijau.**

---

## Tahap 3 — Deploy ke VPS

### 3.1 Clone di VPS

```bash
ssh user@IP_VPS
sudo mkdir -p /opt/stock-opname && sudo chown $USER:$USER /opt/stock-opname
git clone git@github.com:USER/stock-opname-app.git /opt/stock-opname
cd /opt/stock-opname
```

Untuk private repo, siapkan **deploy key** (read-only) di VPS:

```bash
ssh-keygen -t ed25519 -C "vps-deploy" -f ~/.ssh/id_deploy -N ""
cat ~/.ssh/id_deploy.pub
# GitHub repo → Settings → Deploy keys → Add deploy key (jangan centang write access)
```

Lalu di `~/.ssh/config`:

```
Host github.com
  IdentityFile ~/.ssh/id_deploy
  IdentitiesOnly yes
```

### 3.2 Buat `.env` produksi

`.env` **tidak** ada di repo. Buat langsung di VPS — jangan pernah commit.

```bash
cd /opt/stock-opname
cp .env.example .env
chmod 600 .env
```

Isi nilai asli. Hasilkan secret acak:

```bash
echo "JWT_ACCESS_SECRET=$(openssl rand -base64 48)"
echo "JWT_REFRESH_SECRET=$(openssl rand -base64 48)"
echo "DB_PASSWORD=$(openssl rand -base64 24)"
```

Isi `.env` (contoh lengkap):

```dotenv
DB_HOST=postgres
DB_PORT=5432
DB_USER=stockopname
DB_PASSWORD=<hasil openssl rand>
DB_NAME=stockopname_db

REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

JWT_ACCESS_SECRET=<hasil openssl rand 1>
JWT_REFRESH_SECRET=<hasil openssl rand 2>
JWT_ACCESS_EXPIRY=15
JWT_REFRESH_EXPIRY=10080

SERVER_PORT=8080
CORS_ALLOWED_ORIGINS=https://so.domainmu.com
ENV=production

# VPS ini sudah punya app lain di :80 → ikat ke localhost saja
NGINX_BIND=127.0.0.1
NGINX_PORT=8088
```

**Aturan wajib:**

- `JWT_ACCESS_SECRET` dan `JWT_REFRESH_SECRET` **harus berbeda**, masing-masing ≥ 32 karakter
- Backend menolak start kalau secret masih memuat `change-this` / `changeme` / `secret-key`
- `ENV=production` wajib mengisi `DB_PASSWORD`
- `CORS_ALLOWED_ORIGINS` harus origin persis (skema + host, tanpa slash di akhir)

### 3.3 Nyalakan stack

`NGINX_BIND` + `NGINX_PORT` di `.env` membuat nginx mendengarkan `127.0.0.1:8088`, bukan `0.0.0.0:80`.

```bash
cd /opt/stock-opname
docker compose -p so-prod up -d --build
```

Tunggu sampai healthy:

```bash
until [ "$(docker inspect -f '{{.State.Health.Status}}' so-prod-backend-1)" = "healthy" ]; do sleep 3; done
docker compose -p so-prod ps
```

Keempat service harus `healthy`: `postgres`, `redis`, `backend`, `dashboard`, `nginx`.

### 3.4 Migrasi database

**Otomatis.** Backend menjalankan `RunMigrations` saat start, membaca `backend/migrations/*.up.sql` yang belum tercatat di `schema_migrations`. Cek:

```bash
docker compose -p so-prod logs backend | grep -i migrasi
docker compose -p so-prod exec -T postgres \
  sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT version, name FROM schema_migrations ORDER BY version;"'
```

Harus muncul versi 1–6.

### 3.5 Seed user

Skema hanya membuat **roles** (admin, supervisor, staff_gudang). Tabel `users` kosong — tanpa langkah ini tidak ada yang bisa login.

```bash
docker compose -p so-prod exec -T postgres \
  sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < backend/migrations/seed_users.sql
```

Password ketiga akun: `staff123`. **Ganti segera di 3.10.**

### 3.6 Verifikasi dari dalam VPS

```bash
curl -s http://127.0.0.1:8088/health
# {"success":true,"data":{"status":"healthy","services":{"database":"healthy","redis":"healthy"}}}
```

Uji login:

```bash
curl -s -X POST http://127.0.0.1:8088/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"staff123"}' | head -c 200
```

### 3.7 Reverse proxy host → container

Port 80/443 sudah dipakai app lain. **Jangan** hentikan app itu. Tambahkan blok di nginx host.

Subdomain disarankan (path butuh ubah `paths.base` SvelteKit).

Buat `/etc/nginx/sites-available/stock-opname`:

```nginx
server {
    listen 80;
    server_name so.domainmu.com;

    client_max_body_size 10M;

    location / {
        proxy_pass http://127.0.0.1:8088;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
    }
}
```

Aktifkan:

```bash
sudo ln -s /etc/nginx/sites-available/stock-opname /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Kalau pakai Apache/Caddy/Traefik, isi ekuivalennya: teruskan ke `http://127.0.0.1:8088` dengan header `Host`, `X-Forwarded-For`, `X-Forwarded-Proto`, dan dukungan WebSocket.

### 3.8 Cloudflare

1. **DNS** → Add record: `A`, name `so`, IPv4 = IP VPS, **Proxy status: Proxied (oranye)**
2. **SSL/TLS → Overview** → mode **Flexible**

   Flexible = Cloudflare→VPS tetap HTTP. Cukup untuk mulai. Untuk **Full**, pasang origin certificate di nginx host dan ubah `proxy_pass` ke `https://127.0.0.1:8088` (butuh TLS di nginx container — belum dikonfigurasi).
3. **SSL/TLS → Edge Certificates** → aktifkan **Always Use HTTPS**
4. Tunggu DNS propagasi (± 1 menit), lalu:

```bash
curl -sI https://so.domainmu.com/health | head -3
```

Setelah domain aktif, pastikan `CORS_ALLOWED_ORIGINS` di `.env` sudah `https://so.domainmu.com` lalu restart backend:

```bash
docker compose -p so-prod up -d backend
```

### 3.9 Uji end-to-end

Ganti `BASE=https://so.domainmu.com` di bawah.

```bash
BASE=https://so.domainmu.com

# 1. login supervisor
TOKEN=$(curl -s -X POST $BASE/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"username":"supervisor","password":"staff123"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['access_token'])")

# 2. daftar gudang
curl -s $BASE/api/v1/warehouses -H "Authorization: Bearer $TOKEN"

# 3. buat sesi (pakai id gudang dari langkah 2)
curl -s -X POST $BASE/api/v1/stock-opname/sessions \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"warehouse_id":"<ID_GUDANG>"}'
```

Lalu di browser HP: login sebagai `staff` → `/scan` → pilih sesi → scan QR berisi `product_code`.

**Checklist yang harus lolos:**

| Uji | Harapan |
|---|---|
| Login admin/supervisor/staff | 200, dapat `access_token` |
| Buat sesi | kode `SO-YYYYMMDD-0001` |
| Staff buka `/scan` | sesi muncul di dropdown |
| Scan QR produk | produk tampil, isi qty, tersimpan |
| Kirim ulang scan yang sama (offline queue) | tidak ada baris duplikat |
| Complete + approve | stok `quantity_system` terupdate |

### 3.10 Pengerasan setelah jalan

**a. Ganti password seed.** Tidak ada endpoint ubah-password; buat user baru lewat admin lalu nonaktifkan yang lama.

```bash
# hash password baru (butuh Go di VPS; kalau tidak ada, jalankan di lokal)
cd /opt/stock-opname/backend && go run ./cmd/hash 'PasswordBaruYangKuat123'

# update di DB
docker compose -p so-prod exec -T postgres sh -c \
  'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "UPDATE users SET password_hash='"'"'<HASH>'"'"' WHERE username='"'"'admin'"'"';"'
```

**b. Backup harian.** `/usr/local/bin/so-backup.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
cd /opt/stock-opname
mkdir -p /var/backups/stock-opname
docker compose -p so-prod exec -T postgres \
  sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  | gzip > /var/backups/stock-opname/so-$(date +%F-%H%M).sql.gz
# simpan 14 hari terakhir
find /var/backups/stock-opname -name 'so-*.sql.gz' -mtime +14 -delete
```

```bash
sudo chmod +x /usr/local/bin/so-backup.sh
# cron: tiap hari 02:30
( sudo crontab -l 2>/dev/null; echo "30 2 * * * /usr/local/bin/so-backup.sh" ) | sudo crontab -
```

Uji restore sekali (ke DB sementara) sebelum percaya backup.

**c. Firewall.** Port 8088 cukup di localhost — jangan dibuka ke internet.

```bash
sudo ufw status
# 8088 TIDAK boleh muncul di daftar allow
```

Compose sudah mengikat `127.0.0.1:8088`; verifikasi:

```bash
ss -tlnp | grep 8088     # harus 127.0.0.1:8088, bukan 0.0.0.0:8088
```

**d. Rotasi log.** Sudah diatur di `docker-compose.yml` (10 MB × 5 file per service).

**e. Update aplikasi.**

```bash
cd /opt/stock-opname
git pull
docker compose -p so-prod up -d --build
docker image prune -f
```

---

## Pemecahan masalah

| Gejala | Sebab | Perbaikan |
|---|---|---|
| `curl` ke 8088 gagal | nginx tidak start / port bentrok | `docker compose -p so-prod logs nginx`, cek `NGINX_PORT` |
| 502 dari Cloudflare | reverse proxy host salah target | Pastikan `proxy_pass http://127.0.0.1:8088;` (tanpa trailing slash berlebih) |
| Login 401 semua user | `users` kosong | Jalankan ulang langkah 3.5 |
| CORS error di browser | origin tidak cocok | `CORS_ALLOWED_ORIGINS` harus persis `https://so.domainmu.com`, lalu restart backend |
| Backend restart loop | secret lemah / `.env` salah | `docker compose -p so-prod logs backend` — pesan menyebut variabelnya |
| Kamera tidak jalan di HP | bukan HTTPS | Kamera butuh secure context. Pastikan akses lewat `https://`, bukan `http://IP` |
| Sesi dibuat tapi staff tidak lihat | staff buka URL/IP berbeda (stack lain) | Samakan `BASE` untuk semua orang |

---

## Yang belum tercakup

- **APK Flutter** — build nanti: `flutter build apk --dart-define=API_BASE_URL=https://so.domainmu.com/api/v1 --dart-define=WS_BASE_URL=wss://so.domainmu.com/api/v1`
- **TLS di origin (Full mode)** — perlu sertifikat di nginx container
- **Path-based routing** (`domainmu.com/so`) — perlu ubah `paths.base` di `svelte.config.js`
