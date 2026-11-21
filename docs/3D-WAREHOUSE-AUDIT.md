# SOPQ — 3D Digital Twin Warehouse: PHASE 1 Audit Report

Status: **Draft untuk review** · Tanggal: 2026-11-20
Cakupan: audit stack, dependensi, skema DB, API, auth, realtime, design system. **Tidak ada perubahan kode.**

---

## 1. Ringkasan Eksekutif

Spek 3D Digital Twin (20 bab) layak dibangun, **tetapi** audit menemukan **4 blocker** yang membuat
spek itu tidak bisa diterapkan apa adanya. Semua bisa diselesaikan, tapi dua di antaranya adalah
**keputusan arsitektural yang butuh persetujuanmu** sebelum PHASE 2.

| # | Temuan | Dampak | Status |
|---|--------|--------|--------|
| B1 | Spek diasumsikan **Rack-level** masih ada. Kita baru saja **menghapus** rak/box (migration 000014). | Spek & DB bentrok | 🔴 Butuh keputusan → **Opsi 3 dipilih** |
| B2 | **Threlte tidak kompatibel** dengan Svelte 4.2.7 proyek ini. | Pilihan 3D engine | 🟠 Sudah terjawab: pakai **Three.js murni** |
| B3 | **Tidak ada field `barcode`** pada produk & **tidak ada konsep carton/dus** di backend. | BAB 6 & 7 tak bisa langsung diterapkan | 🔴 Butuh keputusan cakupan |
| B4 | Scan existing = `product_code` (bukan barcode dus individual). WS existing **belum** warehouse-scoped & belum ada auth WS. | BAB 7, 16 perlu kerja tambahan | 🟡 Direncanakan di PHASE 10–11 |

---

## 2. Hasil Audit

### 2.1 Frontend (dashboard)
| Item | Fakta | Implikasi |
|------|-------|-----------|
| Svelte | **4.2.7** (dev) | Threlte 8.x butuh Svelte ≥5 → **tidak bisa dipakai** |
| SvelteKit | 2.0.0 | OK |
| Vite | 5.0.3 | OK, tahu menangani GLB via plugin |
| three | **belum terpasang** | Perlu `npm i three` + `@types/three` |
| svelte-query | @tanstack/svelte-query 5.17 | Data fetching 3D lewat sini |
| Tailwind/DaisyUI | 3.4.1 / 4.6.0 | Design system `card-apple`, `metric`, glass — konsisten |

**Keputusan engine (B2):** pakai **Three.js langsung** di dalam komponen Svelte (`onMount` + `onDestroy`),
bukan Threlte. Alasan: proyek di Svelte 4, upgrade ke Svelte 5 = refactor besar seluruh dashboard
(melanggar BAB 19 aturan #7 & #14). Three.js murni memberi kontrol penuh atas InstancedMesh & LOD.

### 2.2 Backend (Go)
| Item | Fakta |
|------|-------|
| Framework | Gin + GORM v2, PostgreSQL 16 |
| Auth | JWT (access+refresh), RBAC middleware `RequireRole` |
| Realtime | `websocket/` dengan `BroadcastNewScan`, `BroadcastSessionUpdate`, `BroadcastDiscrepancyAlert`, `BroadcastProgress` |
| Layout | `warehouses.layout_json` (JSONB) **sudah ada** — bisa dipakai untuk layout 3D awal |
| Barcode | Hanya `product_code`; endpoint `GET /products/code/:code` ("Kritikal untuk scan, QR = product_code") |
| Carton/dus | **Tidak ada** di domain |

### 2.3 Database
| Item | Fakta |
|------|-------|
| Migrasi terakhir | `000014_remove_boxes` — DROP `boxes`, pindah ke `location_id` |
| Hierarki sekarang | `warehouses` → `locations` (leaf). **Tidak ada Zone/Rack/Level/Bin** |
| `locations` kolom | `id, warehouse_id, code, name, created_at` — **belum ada** posisi/dimensi/kapasitas/zone |
| `products.location_id` | nullable, FK ke `locations` |
| `stock_opname_detail.location_id` | snapshot lokasi per scan |
| `schema_migrations` | kolom `(version, name)` — **bukan** format golang-migrate standar |
| Soft delete | hanya `products` (`deleted_at`); `locations`/`warehouses` belum |

### 2.4 Design system & UX
`AppleLayout` + token `card-apple`, `glass`, `metric` (tabular-nums), `press`, hairline,
`--ease-apple`. Sudah premium & minimalis → cocok untuk tampilan digital twin.
Animasi hormat `prefers-reduced-motion`. Layout layout-wrapper punya 5 varian (apple/sidebar/topnav/command/bottom).

---

## 3. Analisis Dampak DB (keputusan Opsi 3)

Opsi 3 = **revisi spek**: konsep "Rack" dipetakan ke `locations`, bukan tabel baru `racks`.
Ini menghindari undo migration 000014 dan tidak membuat duplikat entity (BAB 19 #4).

### Yang perlu ditambah (migrasi ADDITIVE, aman & bisa di-rollback):
`locations` diperluas menjadi objek 3D yang bisa ditempatkan:

```
locations (+ kolom baru)
  warehouse_id   FK (sudah ada)
  code, name     (sudah ada)
  zone_id        FK → zones (BARU, opsional/nullable)   ← memenuhi hierarki Zone
  length_m       NUMERIC                                ← dimensi meter
  width_m        NUMERIC
  height_m       NUMERIC
  pos_x, pos_y, pos_z  NUMERIC                          ← posisi 3D (meter)
  rot_y          NUMERIC (derajat)                      ← rotasi (cukup Y untuk rak)
  capacity       INT                                    ← kapasitas dus
  rack_type      VARCHAR  (light/medium/heavy/pallet/shelving)
  status         VARCHAR  (active/inactive)
  created_at, updated_at, deleted_at                    ← soft delete
```

Tabel **BARU** (hanya yang benar-benar belum ada):
- `warehouse_zones` — Zone dalam gudang (id, warehouse_id, code, name)
- `layout_versions` — versioning layout (BAB 4.5): version_id, warehouse_id, version_number, created_by, change_description, snapshot jsonb
- `cartons` — **HANYA jika kamu menyetujui B3** (lihat §4)
- Event history → bisa memakai `activity_logs` existing (jangan duplikat)

**Tidak dibuat:** `racks`, `bins`, `levels` — dipetakan ke `locations` + `zones`.

---

## 4. Keputusan yang masih dibutuhkan

### B3 — Konsep Carton/Dus (kritis)
Spek BAB 6 & 7 inti-nya memvisualisasikan **dus** (carton) dengan barcode individual, dan
**mengubah warna per dus** setelah scan tervalidasi. Kenyataannya:
- Produk sekarang **tanpa barcode**, tanpa identitas dus.
- Scan = `product_code` produk, bukan barcode dus.
- Tidak ada tabel carton.

Artinya BAB 7 ("dus C001 berubah hijau") **tidak punya sumber data**. Ada 2 arah:

**A. Carton sebagai objek baru** — buat tabel `cartons` (barcode unik per dus, product_id, qty,
location_id). Scan mengarah ke carton. Lebih besar, ubah alur scan mobile+web.
**B. Sederhanakan ke level lokasi** — status warna diberikan **per lokasi/rak** (hijau = lokasi
sudah tervalidasi penuh), bukan per dus. Lebih cepat, tidak butuh barcode dus, selaras data existing.

> Tanpa data barcode dus nyata, opsi A hanya akan menghasilkan data yang dikarang — melanggar
> BAB 19 #8/#9. **Rekomendasiku: mulai dari B, rancang skema agar A bisa ditambah nanti.**

### B1 — Konfirmasi Opsi 3
Kamu sudah pilih Opsi 3 (revisi spek, rak → berbasis `locations`). Migrasi kolom di §3 akan
ditambahkan di PHASE 2, bersifat additive & reversible.

---

## 5. Rencana Fase yang Direvisi (Opsi 3)

| Fase | Isi | Output |
|------|-----|--------|
| 1 | Audit (dokumen ini) | laporan + keputusan |
| 2 | Fondasi DB: `zones`, perluas `locations` (spasial), `layout_versions`, soft-delete | migration + model |
| 3 | Multi-warehouse & zone CRUD (API existing diperluas, bukan duplikat) | API + UI |
| 4 | Rack/Lokasi management + penempatan produk | API + UI |
| 5 | Fondasi Three.js (scene, kamera, lighting, InstancedMesh) | komponen 3D dasar |
| 6 | 3D Warehouse Builder (editor manual + simpan layout + versioning) | editor |
| 7 | Import denah (SVG/Image background + tracing; DXF dinilai kelayakan lib) | workflow import |
| 8 | 3D Visualization (render lokasi, interaksi, panel detail) | halaman 3D |
| 9 | Integrasi Inventory (lihat stok/cari SKU/sorot lokasi) | integrasi |
| 10 | Integrasi Stock Opname (status warna — sesuai keputusan B3) | integrasi |
| 11 | Realtime (warehouse-scoped subscribe, auth WS, reconnect, initial-sync) | realtime |
| 12 | Monitoring + Historical Playback | dashboard + playback |
| 13 | Security & optimasi (InstancedMesh, LOD, fallback 2D/daftar) | hardening |
| 14 | Testing + build + deployment + dokumentasi | lulus produksi |

---

## 6. Yang kubutuhkan darimu sebelum PHASE 2

1. **B3 — pilih A atau B** untuk konsep carton/dus (rekomendasi: **B**).
2. **Setujui** rencana fase yang direvisi di §5.
3. Setelah itu, PHASE 2 (migrasi DB fondasi) baru kumulai — masih terkontrol, additive, reversible.

Aku **tidak** akan menulis kode 3D apa pun sebelum butir 1 & 2 kamu setujui — karena B3
menentukan bentuk skema & alur scan di 5 fase berikutnya.
