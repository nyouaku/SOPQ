# Rencana Layout SOPQ

Status: **Proposal — 2 layout langsung dikerjakan; 5 layout lain mengantre menunggu prompt** · 2026-11-21
Konteks: setelah layout lama dihapus, tersisa **Apple** sebagai satu-satunya layout.

## Daftar antrean layout

| No | Layout | Status | Catatan |
|---|---|---|---|
| 0 | Apple | Aktif | Layout existing, tidak diubah |
| 1 | **Focus** | **Dikerjakan sekarang** | Canvas mode, tanpa chrome |
| 2 | **Workspace** | **Dikerjakan sekarang** | Rail + context panel |
| 3 | Retro UI | **Antre — tunggu prompt user** | prompt khusus menyusul |
| 4 | Neobrutalism | **Antre — tunggu prompt user** | prompt khusus menyusul |
| 5 | Acrylic UI | **Antre — tunggu prompt user** | prompt khusus menyusul |
| 6 | Neumorphism | **Antre — tunggu prompt user** | prompt khusus menyusul |
| 7 | Editorial UI | **Antre — tunggu prompt user** | prompt khusus menyusul |

Aturan antrean: **jangan memulai layout #3–#7 sebelum user mengirim prompt spesifik untuk layout itu.**
Satu prompt = satu layout = satu fase implementasi + test + commit terpisah.

---

## 1. Prinsip desain bersama (berlaku untuk semua layout)

1. **Sumber menu tunggal:** tetap `src/lib/components/layouts/menu.ts` (`MENU_ITEMS`, `menuForApp`, `activeAppFor`, `menuForRole`). Tidak boleh ada daftar menu duplikat.
2. **RBAC UI-only:** layout boleh menyembunyikan menu, tapi keputusan keamanan tetap di backend. Tidak boleh ada aksi yang hanya "dihide" tapi endpoint-nya tak diproteksi.
3. **State navigasi konsisten:** `navMode` (full / icons / hidden) milik Apple. Dua layout baru punya mekanisme own, tapi disarankan memakai satu store bersama (lihat §5).
4. **Design token sama:** `glass`, `card-apple`, `hairline`, `press`, `--ease-apple`, `metric`, tema DaisyUI. Tidak boleh warna/font baru.
5. **Responsif:** mobile tetap mengutamakan DONE satu tangan — bottom bar atau rail.
6. **Aksesibilitas:** semua ikon punya `aria-label`; keyboard navigation; `prefers-reduced-motion` dihormati.
7. **Tidak mengubah logika halaman:** layout hanya pembungkus `<slot />`.

---

## 2. Layout A — `Workspace` (Rail + Context Panel)

**Tujuan:** kerja operator sehari-hari. Menu always-visible tapi minimal, plus panel kontekstual untuk info sesi/gudang.

### 2.1 Konsep visual
```
┌────┬──────────────────────────────┬───────────────┐
│ R  │  Topbar (judul + aksi)       │ Context Panel │
│ a  ├──────────────────────────────┤ (opsional)    │
│ i  │                              │ • Sesi aktif  │
│ l  │         SLOT / HALAMAN       │ • Gudang      │
│    │                              │ • Progress    │
│    │                              │ • Refresh     │
└────┴──────────────────────────────┴───────────────┘
```

### 2.2 Elemen
- **Rail kiri** (lebar 3.5–4rem): ikon saja, tooltip saat hover, active indicator tipis.
- **Topbar tipis**: judul halaman, app switcher (opname/inventory), notifikasi, user.
- **Context panel kanan** (lebar 16–18rem): hanya muncul di halaman yang punya konteks (mis. `/scan`, `/sessions/[id]`, `/inventory/entry`). Bisa di-toggle.
- Area kerja `<main>` mengisi sisa ruang.

### 2.3 Kapan dipakai
- Staff/operator yangButuh fokus + info kontekstual tanpa membuka halaman lain.
- Cocok untuk `/scan`, `/sessions/[id]`, Grid Harian.

### 2.4 Data panel kontekstual (rencana)
Panel tidak mengambil API sendiri — memakai `svelte-query` store yang sudah ada atau props dari halaman.
Kandidat konten:
- Aktif SO: kode sesi, gudang, progress `scanned/target`, status.
- Inventory: gudang terpilih, saldo hari ini, jumlah discrepancy.

### 2.5 File baru
- `dashboard/src/lib/components/layouts/WorkspaceLayout.svelte`
- `dashboard/src/lib/components/layouts/WorkspaceContextPanel.svelte`
- Update `LayoutWrapper.svelte`, `theme.ts` (`ThemeLayout` + `THEME_LAYOUTS`), `ThemeSelector.svelte`.

### 2.6 Risiko
- Context panel butuh kontrak data; kalau diambil dari halaman, perlu mekanisme slot bernama (`<slot name="context">`) — lebih aman daripada layout menebak API.

---

## 3. Layout B — `Focus` (Canvas Mode)

**Tujuan:** layar penuh untuk kerja tanpa gangguan. Ini versi resmi dan konsisten dari fullscreen Spreadsheet/Grid Harian.

### 3.1 Konsep visual
```
┌──────────────────────────────────────────────┐
│  Floating mini-bar (judul + exit + actions)  │
│                                              │
│              SLOT / HALAMAN                  │
│                                              │
└──────────────────────────────────────────────┘
```

### 3.2 Elemen
- **Tanpa sidebar, tanpa bottom bar, tanpa topbar penuh.**
- `mini-bar` mengambang di kanan atas: judul halaman, tombol-aksi kontekstual, tombol **Keluar Focus**.
- Konten memakai 100% viewport (`.focus-canvas`).
- Tombol **Keluar Focus** selalu terlihat & jalan via keyboard (`Esc`).
- Bisa diaktifkan dua cara: (a) ganti layout via ThemeSelector, atau (b) tombol "Focus" di halaman tertentu yang men-set layout.

### 3.3 Kapan dipakai
- Grid Harian / Spreadsheet ala Excel.
- Scan mode di layar besar.
- Presentasi/monitoring realtime.

### 3.4 Hubungan dengan fullscreen Spreadsheet saat ini
- Fullscreen Spreadsheet **existing** tetap dipertahankan sebagai mode lokal per-komponen.
- Layout Focus adalah **wrapper global**: kalau user memilih layout Focus, semua halaman tampil tanpa chrome.
- Konflik dicegah: SpreadsheetEntry tak perlu lagi mengurus z-index chrome — layout already clean.

### 3.5 File baru
- `dashboard/src/lib/components/layouts/FocusLayout.svelte`
- Update `LayoutWrapper.svelte`, `theme.ts`, `ThemeSelector.svelte`.

### 3.6 Risiko
- User bisa "terjebak" tanpa navigasi. Mitigasi: tombol Exit jelas + keyboard shortcut + preferensi tersimpan.
- Halaman dengan tombol kembali sendiri (mis. Grid Harian `←`) tetap jalan.

---

## 4. Perbandingan cepat

| Aspek | Apple (existing) | Workspace | Focus |
|---|---|---|---|
| Navigasi | Sidebar + topbar + bottombar | Rail ikon + topbar | Floating mini-bar |
| Panel konteks | Tidak | Ya (opsional) | Tidak |
| Cocok untuk | Umum | Operasional harian | Grid/scan/monitoring |
| Mobile | Drawer + bottombar | Rail + bottom mini | Mini-bar |
| Privasi layar | Menengah | Menengah | Maksimal |

---

## 5. State & preferensi (usulan teknis)

Saat ini `navMode` disimpan lokal di `AppleLayout`. Kalau 2 layout baru ditambahkan, sebaiknya dipusatkan:

```ts
// src/lib/stores/navigation.ts (BARU, rencana)
type NavMode = 'full' | 'icons' | 'hidden';
type RailMode = 'icons' | 'hidden';
```

- Apple Layout memakai `NavMode`.
- Workspace Layout memakai `RailMode` + `contextOpen`.
- Focus Layout tidak punya chrome state (kecuali `focusActive`).
- Semua disimpan di `localStorage`, dibaca di `onMount`, dijaga `try/catch`.

Ini **opsional** untuk fase pertama; bisa shader-nanti kalau dua layout sudah menyala.

---

## 6. Urutan implementasi yang aman

| Fase | Isi | Dampak |
|---|---|---|
| P1 | Ekstensi `ThemeLayout` + `THEME_LAYOUTS` (3 opsi) | Dokumen & type saja |
| P2 | `FocusLayout` (paling sederhana) + wiring LayoutWrapper | Risiko rendah |
| P3 | `WorkspaceLayout` + rail + topbar | Risiko sedang |
| P4 | Context panel via named slot | Risiko sedang |
| P5 | Pusatkan state navigasi (store bersama) | Refactor ringan |
| P6 | Test manual di desktop + mobile + preferensi tersimpan | — |

Kenapa Focus dulu: paling kecil, langsung memberi manfaat nyata (menggantikan fullscreen ad-hoc), dan tak menyentuh menu.

---

## 7. Kriteria penerimaan

- Pilihan layout di Settings berfungsi & bertahan setelah refresh.
- Apple tetap berfungsi persis seperti sekarang.
- Focus: sidebar & bottom bar benar-benar hilang; tombol keluar selalu tersedia.
- Workspace: rail menampilkan menu sesuai role & app aktif.
- Tidak ada API/endpoint baru.
- Tidak ada perubahan logika bisnis.
- `svelte-check`, build, dan test dashboard lulus.

---

## 8. Keputusan yang diminta sebelum implementasi

1. **Nama layout:** `Workspace` + `Focus` — OK, atau kamu punya nama lain?
2. **Context panel Workspace:** pakai named slot (aman) atau panel yang mengambil data sendiri (lebih praktis tapi耦合)?
3. **Focus vs fullscreen Spreadsheet existing:** biarkan keduanya, atau Focus menggantikan fullscreen komponen?
4. **Urutan:** setuju Focus dulu, baru Workspace?
