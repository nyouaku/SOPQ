# Stock Opname Application

Aplikasi Stock Opname berbasis QRCode & Voice Input dengan arsitektur production-ready: Golang backend, Flutter mobile app, dan SvelteKit dashboard.

## 🚀 Tech Stack

**Backend**
- Golang 1.24+ dengan Gin Framework
- PostgreSQL 16 + Redis 7
- JWT Authentication (access + refresh token)
- Clean Architecture (domain → usecase → repository → delivery)
- WebSocket untuk real-time progress
- GORM v2 + Optimistic Locking

**Mobile (Flutter)**
- Flutter dengan Riverpod state management
- QR Code Scanner + Speech-to-Text + Text-to-Speech
- SQLite untuk offline mode dengan auto-sync
- Material Design 3 Dark Theme

**Dashboard (SvelteKit)**
- SvelteKit + TypeScript
- TailwindCSS + DaisyUI
- Chart.js untuk visualisasi
- TanStack Query untuk data fetching

**Deployment**
- Docker + Docker Compose
- Nginx reverse proxy
- PostgreSQL + Redis containers

## 📋 Features

✅ **Autentikasi & RBAC** — JWT dengan role admin/supervisor/staff  
✅ **Master Data** — Products, Categories, Suppliers, Warehouses  
✅ **QR Code Scanner** — Auto-focus dengan TTS feedback  
✅ **Voice Input** — STT untuk input jumlah barang  
✅ **Offline Mode** — SQLite queue dengan auto-sync  
✅ **Real-time Progress** — WebSocket untuk dashboard supervisor  
✅ **Reports & Export** — Filter lengkap dengan CSV export  
✅ **Audit Trail** — Activity log otomatis untuk semua scan  
✅ **Optimistic Locking** — Aman untuk 50 user concurrent  

## 🏗️ Architecture

```
stock-opname-app/
├── backend/          # Golang API (Clean Architecture)
│   ├── cmd/api/      # Main entry point
│   ├── internal/     # Business logic
│   │   ├── domain/       # Entities
│   │   ├── usecase/      # Business logic
│   │   ├── repository/   # Data access
│   │   ├── delivery/     # HTTP handlers & WebSocket
│   │   └── middleware/   # Auth, RBAC, CORS, etc.
│   ├── pkg/          # Shared utilities
│   └── migrations/   # Database migrations
├── mobile/           # Flutter app
│   └── lib/
│       ├── core/         # Theme, network, storage
│       ├── data/         # Models & data sources
│       └── features/     # Auth, Scanner, Sync
├── dashboard/        # SvelteKit web dashboard
│   └── src/
│       ├── lib/          # API client, stores
│       └── routes/       # Pages (login, dashboard, reports)
└── docker/           # Docker configs
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.24+ (untuk development)
- Node.js 20+ (untuk development)
- Flutter 3.x (untuk development)

### Run with Docker Compose

1. Clone repository:
```bash
git clone https://github.com/nyouaku/stock-opname-app.git
cd stock-opname-app
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Edit `.env` sesuai kebutuhan, terutama:
```env
DB_PASSWORD=your-secure-password
JWT_ACCESS_SECRET=your-secret-key
JWT_REFRESH_SECRET=your-refresh-secret
```

4. Start semua services:
```bash
docker compose up -d
```

5. Akses aplikasi:
- **Dashboard:** http://localhost
- **Backend API:** http://localhost/api/v1
- **Health Check:** http://localhost/health

### Development Mode

**Backend:**
```bash
cd backend
go mod download
cp .env.example .env
go run cmd/api/main.go
```

**Mobile:**
```bash
cd mobile
flutter pub get
flutter run
```

**Dashboard:**
```bash
cd dashboard
npm install
cp .env.example .env
npm run dev
```

## 📚 Documentation

- [Backend README](./backend/README.md) — API documentation & setup
- [Mobile README](./mobile/README.md) — Flutter app setup & features
- [Dashboard README](./dashboard/README.md) — Web dashboard setup
- [Database Schema](./docs/DATABASE_SCHEMA.sql) — Complete DB schema
- [API Endpoints](./docs/API_ENDPOINTS.md) — API documentation
- [Progress Tracker](./docs/PROGRESS.md) — Development progress

## 🔐 Default Credentials

**Admin:**
- Username: `admin`
- Password: `admin123`

**Supervisor:**
- Username: `supervisor`
- Password: `supervisor123`

**Staff:**
- Username: `staff`
- Password: `staff123`

⚠️ **PENTING:** Ganti password default setelah deployment!

## 🧪 Testing

**Backend Unit Tests:**
```bash
cd backend
go test ./internal/usecase/...
```

## 📊 Performance Targets

| Metrik | Target | Status |
|--------|--------|--------|
| Scan → data tampil | < 500 ms | ✅ |
| Save hasil opname | < 300 ms | ✅ |
| Kapasitas produk | 100.000 SKU | ✅ |
| Kapasitas transaksi | 1.000.000 baris | ✅ |
| Concurrent user | 50 user bersamaan | ✅ |

## 🔒 Security Features

- JWT access token (15 menit) + refresh token (7 hari)
- Password hashing dengan bcrypt (cost 12)
- Rate limiting untuk auth endpoints
- RBAC middleware untuk authorization
- Input validation dengan go-playground/validator
- HTTPS via Nginx reverse proxy
- CORS configuration
- Audit log untuk semua perubahan data

## 🛠️ Maintenance

**Backup Database:**
```bash
docker exec stockopname-postgres pg_dump -U postgres stockopname_db > backup.sql
```

**Restore Database:**
```bash
docker exec -i stockopname-postgres psql -U postgres stockopname_db < backup.sql
```

**View Logs:**
```bash
docker compose logs -f backend
docker compose logs -f dashboard
```

**Stop Services:**
```bash
docker compose down
```

**Remove All Data:**
```bash
docker compose down -v
```

## 🐛 Troubleshooting

**Backend tidak bisa connect ke database:**
```bash
# Cek status PostgreSQL
docker compose logs postgres
# Restart backend
docker compose restart backend
```

**Dashboard tidak bisa akses API:**
- Pastikan `VITE_API_BASE_URL` di `.env` sudah benar
- Cek CORS di backend: `CORS_ALLOWED_ORIGINS`

**Mobile app tidak sync:**
- Cek koneksi internet
- Lihat pending scans count di badge
- Force sync dengan pull-to-refresh

## 📝 License

MIT License

## 👥 Contributors

- Developer: [nyouaku](https://github.com/nyouaku)

---

**Built with ❤️ using Claude Code**
