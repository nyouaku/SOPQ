# Build Mobile APK - Stock Opname App

## Prerequisites

### 1. Install Android Studio
1. Download dari https://developer.android.com/studio
2. Install dengan komponen:
   - Android SDK
   - Android SDK Platform
   - Android SDK Build-Tools
   - Android Emulator (optional)

### 2. Setup Flutter Android Toolchain
```powershell
# Accept Android licenses
flutter doctor --android-licenses

# Verify setup
flutter doctor -v
```

### 3. Configure API Base URL
Edit file `mobile/lib/core/constants/app_constants.dart`:
```dart
static const String apiBaseUrl = 'http://192.168.18.145:8080/api/v1';
```
Ganti dengan IP server backend yang accessible dari HP.

## Build APK

### Debug APK (untuk testing)
```powershell
cd mobile
flutter build apk --debug
```
Output: `build/app/outputs/flutter-apk/app-debug.apk`

### Release APK (untuk production)
```powershell
cd mobile
flutter build apk --release
```
Output: `build/app/outputs/flutter-apk/app-release.apk`

### Split APK by ABI (smaller file size)
```powershell
flutter build apk --split-per-abi --release
```
Output:
- `app-armeabi-v7a-release.apk` (32-bit ARM)
- `app-arm64-v8a-release.apk` (64-bit ARM - most modern phones)
- `app-x86_64-release.apk` (x86 - emulator)

## Install ke HP

### Via USB
1. Enable Developer Options & USB Debugging di HP
2. Connect HP via USB
3. Run:
```powershell
flutter install
```

### Via File Transfer
1. Copy APK ke HP (WhatsApp, Email, atau USB)
2. Install manually di HP
3. Enable "Install from Unknown Sources" jika perlu

## Build via Docker (Alternative)

Jika tidak mau install Android Studio lokal:

```dockerfile
# Dockerfile.flutter-builder
FROM cirrusci/flutter:stable

WORKDIR /app
COPY mobile/ .

RUN flutter pub get
RUN flutter build apk --release --split-per-abi

# Output ada di /app/build/app/outputs/flutter-apk/
```

Build:
```bash
docker build -f Dockerfile.flutter-builder -t flutter-builder .
docker create --name flutter-build flutter-builder
docker cp flutter-build:/app/build/app/outputs/flutter-apk/ ./mobile/build/
docker rm flutter-build
```

## Testing di HP

1. Pastikan HP dan server backend dalam network yang sama
2. Cek IP server: `ipconfig` (Windows) atau `ip addr` (Linux)
3. Update `apiBaseUrl` di `app_constants.dart` dengan IP tersebut
4. Rebuild APK
5. Install dan test

## Troubleshooting

### "Unable to locate Android SDK"
```powershell
flutter config --android-sdk "C:\Users\<USER>\AppData\Local\Android\Sdk"
```

### "Android license not accepted"
```powershell
flutter doctor --android-licenses
```
Ketik `y` untuk semua license.

### Build error: "Minimum supported Gradle version"
Edit `mobile/android/gradle/wrapper/gradle-wrapper.properties`:
```properties
distributionUrl=https\://services.gradle.org/distributions/gradle-8.3-all.zip
```

### Network error di HP
- Pastikan backend accessible dari HP (ping/curl test)
- Check firewall Windows: allow port 8080
- Gunakan IP yang benar (bukan 127.0.0.1 atau localhost)

## Current Status

✅ Mobile app code complete:
- Authentication (login/logout)
- QR/Barcode scanner
- Voice input with TTS
- Offline sync ready
- Session management

❌ APK belum di-build (Android SDK belum terinstall)

**API Base URL saat ini:** `http://192.168.18.145:8080/api/v1`

Setelah build APK, transfer ke HP dan test!
