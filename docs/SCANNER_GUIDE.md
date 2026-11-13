# Scanner Guide - Stock Opname App

## Overview

Aplikasi ini menyediakan **2 versi scanner** untuk fleksibilitas maksimal:

1. **Desktop Scanner** (`/scan`) - Untuk digunakan di komputer/laptop dengan webcam atau upload gambar
2. **Mobile Scanner** (`/mobile`) - Untuk digunakan di smartphone dengan camera

---

## 1. Desktop Scanner

**URL:** `http://localhost/scan`

### Features:
- ✅ **Scan dengan Kamera** - Real-time scanning menggunakan webcam
- ✅ **Upload QR Code** - Upload gambar QR code/barcode dari file
- ✅ Native Barcode Detection API (Chrome/Edge)
- ✅ ZXing library fallback (browser lain)
- ✅ Auto product lookup
- ✅ Session selection
- ✅ Desktop-optimized layout

### Cara Menggunakan:

#### A. Scan dengan Kamera (Webcam)
1. Buka `http://localhost/scan`
2. Login dengan akun Anda
3. Pilih **Sesi Stock Opname** yang aktif
4. Pastikan tab **"Scan dengan Kamera"** terpilih
5. Klik tombol **"Aktifkan Kamera"**
6. Browser akan minta izin akses camera → Klik **"Allow"**
7. Arahkan webcam ke barcode/QR code produk
8. Tunggu hingga terdeteksi otomatis (scanner overlay akan berkedip)
9. Produk akan muncul secara otomatis
10. Input **Jumlah Fisik**
11. Klik **"✓ Simpan Data"**
12. Klik **"Reset & Scan Lagi"** untuk scan produk berikutnya

#### B. Upload QR Code (Gambar File)
1. Buka `http://localhost/scan`
2. Login dengan akun Anda
3. Pilih **Sesi Stock Opname** yang aktif
4. Klik tab **"Upload QR Code"**
5. Klik **"Choose File"** atau drag-and-drop gambar
6. Pilih gambar QR code/barcode dari:
   - Foto dari HP yang sudah di-transfer ke komputer
   - Screenshot QR code
   - Gambar dari galeri
7. Upload file (format: JPG, PNG, WEBP, dll)
8. Aplikasi akan auto-decode barcode dari gambar
9. Produk akan muncul secara otomatis
10. Input **Jumlah Fisik**
11. Klik **"✓ Simpan Data"**
12. Klik **"Reset & Scan Lagi"** untuk scan produk berikutnya

### Tips Desktop Scanner:
- **Webcam tidak muncul?**
  - Check browser permissions: Settings → Privacy → Camera
  - Gunakan Chrome/Edge untuk hasil terbaik
  - Pastikan webcam tidak sedang digunakan aplikasi lain

- **Upload QR Code gagal detect?**
  - Pastikan gambar jelas & tidak blur
  - QR code/barcode harus terlihat dengan jelas
  - Gunakan lighting yang cukup saat foto
  - Coba crop gambar agar fokus ke barcode saja

- **Browser support:**
  - ✅ Chrome/Edge: Native Barcode Detection API (fastest)
  - ✅ Firefox/Safari: ZXing library fallback (still works)

---

## 2. Mobile Scanner

**URL:** `http://localhost/mobile`

### Features:
- ✅ **Camera Scanner** - Back camera auto-selected
- ✅ **Manual Input** - Ketik barcode manual jika camera gagal
- ✅ Animated scanner overlay
- ✅ Touch-optimized number pad
- ✅ Offline queue (localStorage)
- ✅ Auto-focus flow
- ✅ PWA-ready (install to home screen)

### Cara Menggunakan:

#### A. Scan dengan Kamera (Mobile)
1. Buka `http://localhost/mobile` di browser HP
2. Login dengan akun Anda
3. Pilih **Sesi Aktif**
4. Klik tombol hijau **"SCAN DENGAN KAMERA"**
5. Browser akan minta izin akses camera → Klik **"Izinkan"**
6. Camera akan aktif (otomatis pakai back camera)
7. Arahkan camera ke barcode/QR code produk
8. Tunggu hingga auto-detect (scan line bergerak)
9. Produk akan muncul otomatis
10. Input **Jumlah Fisik** dengan number pad
11. Klik **"✓ SIMPAN"**
12. Form reset otomatis, langsung scan berikutnya

#### B. Manual Input (Keyboard)
1. Buka `http://localhost/mobile` di browser HP
2. Login dengan akun Anda
3. Pilih **Sesi Aktif**
4. Ketik barcode langsung di input field
5. Klik **"Cari"** atau tekan Enter
6. Produk akan muncul jika ditemukan
7. Input **Jumlah Fisik** dengan number pad
8. Klik **"✓ SIMPAN"**

### Tips Mobile Scanner:
- **Camera tidak muncul?**
  - Pastikan browser punya izin camera
  - Chrome mobile recommended
  - Coba reload page (swipe down)

- **Barcode tidak terdeteksi?**
  - Dekatkan/jauhkan camera
  - Pastikan lighting cukup
  - Coba tahan lebih stabil
  - Fallback: gunakan manual input

- **Install sebagai App:**
  1. Chrome: Menu (⋮) → "Install app"
  2. Safari (iOS): Share → "Add to Home Screen"
  3. Icon muncul di home screen
  4. Buka seperti app biasa (no browser UI)

- **Offline Mode:**
  - Data auto-save ke localStorage jika offline
  - Badge "X offline" muncul di navbar
  - Auto-sync saat online kembali

---

## Comparison: Desktop vs Mobile

| Feature | Desktop Scanner (`/scan`) | Mobile Scanner (`/mobile`) |
|---------|---------------------------|----------------------------|
| **Camera** | Webcam | Back camera (smartphone) |
| **Upload File** | ✅ Yes | ❌ No |
| **Manual Input** | ✅ Yes | ✅ Yes |
| **Number Pad** | ❌ Keyboard | ✅ Touch-optimized |
| **Offline Queue** | ❌ No | ✅ Yes (localStorage) |
| **PWA Install** | ❌ No | ✅ Yes |
| **Best For** | Desktop dengan webcam | Smartphone/tablet |
| **Layout** | Wide screen | Mobile-first |

---

## Navigation

Akses scanner dari menu dashboard:

```
Dashboard → Menu (☰) → Scan
```

Atau langsung:
- Desktop: `http://localhost/scan`
- Mobile: `http://localhost/mobile`

---

## Troubleshooting

### Camera Permission Denied
**Desktop:**
```
Chrome: chrome://settings/content/camera
Edge: edge://settings/content/camera
```
1. Cari "localhost" di list
2. Ubah permission menjadi "Allow"
3. Reload page

**Mobile:**
```
Chrome: Settings → Site settings → Camera
Safari: Settings → Privacy → Camera → Safari
```

### Barcode Not Detected
1. **Lighting** - Pastikan cukup cahaya
2. **Distance** - 10-30cm dari camera
3. **Stability** - Tahan stabil 2-3 detik
4. **Focus** - Tunggu camera auto-focus
5. **Fallback** - Gunakan manual input

### Browser Compatibility
- ✅ Chrome (Desktop & Mobile) - **RECOMMENDED**
- ✅ Edge (Desktop) - **RECOMMENDED**
- ⚠️ Firefox - Works dengan ZXing fallback
- ⚠️ Safari - Works dengan ZXing fallback
- ❌ IE11 - Not supported

### Network Error
1. Hard refresh: `Ctrl + F5` (Desktop) atau swipe down (Mobile)
2. Clear cache: `Ctrl + Shift + Delete`
3. Check API: `http://localhost:8080/health`
4. Restart services: `docker-compose restart`

---

## API Endpoints Used

Scanner menggunakan API endpoints berikut:

```typescript
// Get active sessions
GET /api/v1/stock-opname/sessions?status=in_progress

// Lookup product by barcode
GET /api/v1/products/barcode/:barcode

// Submit scan result
POST /api/v1/stock-opname/sessions/:id/scan
Body: {
  session_id: string,
  barcode: string,
  stock_physical: number,
  input_method: 'camera' | 'upload' | 'manual',
  note: string
}
```

---

## Keyboard Shortcuts (Desktop)

- `Enter` pada barcode input → Cari produk
- `Enter` pada quantity input → Simpan data
- `Esc` saat camera aktif → Stop camera

---

## Security Notes

- Camera permission hanya diminta saat button diklik
- Video stream tidak di-record atau di-save
- Semua processing dilakukan di browser (client-side)
- No video/image data dikirim ke server
- Hanya barcode result yang dikirim ke API

---

## Support

**Documentation:**
- API Endpoints: `docs/API_ENDPOINTS.md`
- Database Schema: `docs/DATABASE_SCHEMA.sql`
- Troubleshooting: `docs/TROUBLESHOOTING.md`

**Access URLs:**
- Dashboard: http://localhost/
- Desktop Scanner: http://localhost/scan
- Mobile Scanner: http://localhost/mobile
- API: http://localhost:8080/api/v1
- Health Check: http://localhost:8080/health
