# ===== FILE: mobile/README.md =====
# Stock Opname Mobile App

Flutter mobile application untuk Stock Opname dengan QR Code scanner dan voice input.

## Features

- 🔐 JWT Authentication dengan auto-refresh token
- 📷 QR Code Scanner dengan auto-focus dan TTS feedback
- 🎤 Voice input untuk input jumlah barang
- 📴 Offline mode dengan SQLite queue dan auto-sync
- 🎨 Material Design 3 Dark Theme
- 🔄 Real-time sync status (Online/Offline/Syncing)

## Tech Stack

- Flutter 3.x
- Riverpod (State Management)
- Dio (HTTP Client)
- SQLite (Local Database)
- Mobile Scanner (QR Code)
- Speech to Text & Flutter TTS
- Flutter Secure Storage

## Setup

1. Install dependencies:
```bash
flutter pub get
```

2. Update API base URL di `lib/core/constants/app_constants.dart`:
```dart
static const String apiBaseUrl = 'http://YOUR_IP:8080/api/v1';
```

3. Run app:
```bash
flutter run
```

## Permissions

Android (`android/app/src/main/AndroidManifest.xml`):
```xml
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.RECORD_AUDIO" />
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

iOS (`ios/Runner/Info.plist`):
```xml
<key>NSCameraUsageDescription</key>
<string>Diperlukan untuk scan QR Code produk</string>
<key>NSMicrophoneUsageDescription</key>
<string>Diperlukan untuk input suara</string>
```

## Folder Structure

```
lib/
├── core/
│   ├── constants/      # App constants & API endpoints
│   ├── network/        # Dio client & API response
│   ├── storage/        # Local database
│   ├── theme/          # App theme
│   └── providers/      # Core providers
├── data/
│   ├── datasources/    # Remote & local data sources
│   └── models/         # Data models
└── features/
    ├── auth/           # Login & authentication
    ├── scanner/        # QR scanner & input
    └── offline_sync/   # Offline queue & sync
```

## Offline Mode

App otomatis detect koneksi internet:
- **Online**: Data langsung submit ke server
- **Offline**: Data disimpan ke SQLite queue
- **Auto-sync**: Saat kembali online, queue otomatis di-sync ke server
- Badge menampilkan status koneksi dan jumlah pending scans

## Usage Flow

1. Login dengan username & password
2. Scanner screen terbuka otomatis
3. Arahkan kamera ke QR Code produk
4. TTS membaca info produk
5. Input jumlah fisik via voice atau keypad
6. Tekan Enter atau tombol SIMPAN
7. Scanner auto-restart untuk scan berikutnya

## Build

Android APK:
```bash
flutter build apk --release
```

Android App Bundle:
```bash
flutter build appbundle --release
```

iOS:
```bash
flutter build ios --release
```
# ===== END FILE =====
