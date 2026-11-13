# CLAUDE.md — Konteks Proyek untuk Claude Code

Proyek: **Stock Opname Barcode + Voice** (WMS/IMS mini) — Backend Golang, Mobile Flutter, Dashboard SvelteKit.

## Sebelum mulai apa pun
1. Baca `PROMPT_MASTER.md` — ini adalah spesifikasi lengkap & urutan PHASE pengerjaan.
2. Baca `docs/DATABASE_SCHEMA.sql` — source of truth skema database.
3. Baca `docs/API_ENDPOINTS.md` — source of truth kontrak API.
4. Kerjakan **satu PHASE per sesi**, sesuai urutan di `PROMPT_MASTER.md` bagian 5.

## Konvensi kode

**Backend (Go)**
- Clean Architecture, arah dependency: `delivery -> usecase -> domain <- repository`
- Domain layer TIDAK boleh import package dari `delivery` atau `infrastructure`
- Semua error dibungkus dengan context (`fmt.Errorf("...: %w", err)`)
- Response API konsisten:
  ```json
  { "success": true, "data": {}, "message": "" }
  { "success": false, "error": { "code": "", "message": "" } }
  ```
- Nama file: `snake_case.go`, nama package: singular lowercase (`product`, bukan `products`)

**Mobile (Flutter)**
- Riverpod untuk state management, hindari `setState` untuk state lintas-widget
- Feature-first folder structure di dalam `features/`, tiap fitur punya `screens/`, `widgets/`, `providers/` sendiri
- Semua panggilan API lewat `core/network/` (Dio + interceptor refresh token)
- Nama file: `snake_case.dart`

**Dashboard (SvelteKit)**
- TypeScript strict mode
- TailwindCSS + DaisyUI, ikuti dark theme konsisten dengan mobile
- Data fetching lewat `src/lib/api/` dengan svelte-query, jangan fetch langsung di komponen

## Yang TIDAK boleh dilakukan
- Jangan ubah tech stack yang sudah ditentukan di `PROMPT_MASTER.md` bagian 1
- Jangan buat struktur folder baru — ikuti yang sudah ada
- Jangan gunakan kode dummy/placeholder untuk logika inti (auth, RBAC, hitung selisih, offline sync)
- Jangan skip PHASE — setiap PHASE harus jalan (`go run` / `flutter run` / `npm run dev`) sebelum lanjut

## Perintah cepat
```bash
# Backend
cd backend && go run cmd/api/main.go

# Mobile
cd mobile && flutter run

# Dashboard
cd dashboard && npm install && npm run dev

# Semua service (setelah PHASE 8)
docker compose up -d
```
