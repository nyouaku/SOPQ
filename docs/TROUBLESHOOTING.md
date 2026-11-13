# TROUBLESHOOTING - Stock Opname App

## Network Error saat Login

### Penyebab Umum:
1. **Browser cache** - JavaScript lama masih di-cache
2. **Docker container** belum restart
3. **API URL** masih hardcoded ke IP lama

### Solusi:

#### 1. Clear Browser Cache (PENTING!)
**Chrome/Edge:**
- Tekan `Ctrl + Shift + Delete`
- Pilih "Cached images and files"
- Time range: "All time"
- Click "Clear data"
- **ATAU** tekan `Ctrl + F5` untuk hard refresh

**Firefox:**
- Tekan `Ctrl + Shift + Delete`
- Pilih "Cache"
- Click "Clear Now"

#### 2. Force Reload Tanpa Cache
- Chrome: `Ctrl + Shift + R` atau `Ctrl + F5`
- Firefox: `Ctrl + Shift + R`
- Safari: `Cmd + Option + R`

#### 3. Restart Docker Containers
```bash
cd D:\aldeb\project\project_21\stock-opname-app
docker-compose restart
```

#### 4. Rebuild Dashboard (jika masih error)
```bash
docker-compose build dashboard
docker-compose up -d
```

---

## Scanner Features

### Desktop Scanner (Dashboard):
```
http://localhost/scan
```

**Features:**
- ✅ **Camera Scanner** - Scan barcode/QR code dengan webcam
- ✅ **Upload QR Code** - Upload gambar QR code dari file
- ✅ Native Barcode Detection API + ZXing fallback
- ✅ Product lookup otomatis
- ✅ Session selection
- ✅ Desktop-optimized layout

**Cara Pakai:**
1. Buka http://localhost/scan
2. Pilih sesi stock opname aktif
3. **Tab "Scan dengan Kamera":**
   - Klik "Aktifkan Kamera"
   - Allow camera permission
   - Arahkan ke barcode
   - Auto-detect & lookup produk
4. **Tab "Upload QR Code":**
   - Klik "Choose File"
   - Pilih gambar QR code
   - Auto-decode & lookup produk
5. Input jumlah fisik
6. Simpan

### Mobile Web Scanner:
```
http://localhost/mobile
```

**Features:**
- ✅ Camera scanner dengan back camera
- ✅ Manual barcode input
- ✅ Product lookup
- ✅ Quantity input dengan number pad
- ✅ Offline queue (auto-save jika offline)
- ✅ Session selection
- ✅ Touch-optimized UI (large buttons)
- ✅ PWA-ready (bisa di-install ke home screen)

### Install sebagai App (Android/iOS):
1. Buka `http://localhost/mobile` di browser
2. Chrome: Menu → "Install app" atau "Add to Home screen"
3. Safari (iOS): Share → "Add to Home Screen"
4. Icon app akan muncul di home screen

---

## Flutter Mobile App Issues

### Flutter pub get lama
**Normal!** Download dependencies memang bisa 5-15 menit pertama kali.

**Penyebab:**
- Banyak dependencies (mobile_scanner, speech_to_text, dll)
- Android SDK download artifacts
- Network speed

**Status Check:**
```bash
cd mobile
flutter pub get --verbose
```

Tunggu hingga muncul:
```
Got dependencies!
```

### Android SDK Not Found
**Setelah install Android Studio:**
1. Buka Android Studio → Settings → SDK Manager
2. Install:
   - Android SDK Platform (latest)
   - Android SDK Build-Tools
   - Android SDK Platform-Tools
3. Accept licenses:
```bash
flutter doctor --android-licenses
```
(Ketik 'y' untuk semua)

4. Verify:
```bash
flutter doctor -v
```

### Build APK
**Setelah SDK ready:**
```bash
cd mobile
flutter build apk --release --split-per-abi
```

APK location:
```
mobile/build/app/outputs/flutter-apk/
- app-arm64-v8a-release.apk (most phones)
- app-armeabi-v7a-release.apk (older phones)
```

---

## Service Status Check

### All Services:
```bash
docker ps --format "{{.Names}}: {{.Status}}"
```

### Logs:
```bash
# Backend
docker logs stockopname-backend --tail 50

# Dashboard
docker logs stockopname-dashboard --tail 50

# Database
docker logs stockopname-postgres --tail 50
```

### Restart Specific Service:
```bash
docker-compose restart backend
docker-compose restart dashboard
```

---

## API Testing

### Health Check:
```bash
curl http://localhost:8080/health
```

### Login Test:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"staff123"}'
```

### Dashboard Summary:
```bash
TOKEN="<your-token>"
curl http://localhost:8080/api/v1/dashboard/summary \
  -H "Authorization: Bearer $TOKEN"
```

---

## Database Issues

### Reset Database:
```bash
docker-compose down -v
docker-compose up -d postgres
# Wait 10 seconds
docker-compose exec -T postgres psql -U stockopname -d stockopname_db < backend/migrations/seed_users.sql
docker-compose exec -T postgres psql -U stockopname -d stockopname_db < backend/migrations/seed_products.sql
docker-compose up -d
```

---

## Performance Issues

### Dashboard Slow:
1. Clear browser cache
2. Check backend logs for errors
3. Check database connection

### Backend Slow:
1. Check PostgreSQL CPU usage
2. Check Redis connection
3. Restart backend:
```bash
docker-compose restart backend
```

---

## Common Errors

### "Invalid authorization format"
- Token expired atau salah format
- Login ulang untuk dapat token baru

### "CORS policy"
- Backend CORS config salah
- Check `.env` file: `CORS_ALLOWED_ORIGINS`

### "Connection refused"
- Service belum start
- Check: `docker ps`

### "Database connection failed"
- PostgreSQL belum ready
- Wait 10-20 seconds setelah `docker-compose up`

---

## Quick Reset (jika semua error)

```bash
# Stop all
docker-compose down

# Rebuild all
docker-compose build

# Start all
docker-compose up -d

# Wait 20 seconds
sleep 20

# Test
curl http://localhost/
curl http://localhost:8080/health
```

---

## Contacts & Links

- Documentation: `docs/` folder
- API Endpoints: `docs/API_ENDPOINTS.md`
- Database Schema: `docs/DATABASE_SCHEMA.sql`
- Build Mobile: `docs/BUILD_MOBILE.md`

**Access URLs:**
- Dashboard: http://localhost/
- Mobile Scanner: http://localhost/mobile
- API: http://localhost:8080/api/v1
- Health: http://localhost:8080/health
