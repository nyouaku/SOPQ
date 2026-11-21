# Theme Audit — Penyebab Masalah

Tanggal: 2026-11-21 · Audit statis (browser automation tidak tersedia).

## Temuan utama

Sistem theme **tidak satu arah**. Ada 3 lapisan yang saling menimpa:

1. **Token DaisyUI OKLCH** di `ui-components.css` (`--b1`, `--b2`, `--bc`, `--p`) → dipakai utility `bg-base-*`, `text-base-*`.
2. **Override komponen hard-coded** di `ui-components.css` → `background: #e8e2cf !important`, `#0d0d0d`, `#fff`, dll. Ini **mengabaikan token** dan menciptakan "panel generik" yang tidak berubah saat tema berganti.
3. **Inline warna di layout** (Retro `#d8d0bb`, Neo `#f7f3dc`, Acrylic gradient gelap, dll) → hanya berlaku di chrome layout, tidak di konten.

Akibatnya:背景 utama berubah, tapi **nested container & komponen internal** memakai warna statis → kontras buruk & "background hitam generik".

## Bukti

| Lokasi | Masalah |
|---|---|
| `ui-components.css` 40+ aturan `!important` dengan hex statis | Menimpa token tema |
| `settings/+page.svelte` `bg-base-200` | Bergantung token, tapi token di-override statis di beberapa komponen |
| `AcrylicLayout.svelte` `[data-ui='acrylic'] --b2: 19%` (gelap) + override `.table` gelap | Teks/container gelap di luar konteks acrylic |
| `ScanTimeline.svelte` `text-white` | Tetap putih di tema terang |
| `scan/+page.svelte` `bg-black` | Disengaja (area scan), tapi tidakdijelaskan/pengecualian |
| `ThemeSelector.svelte` kartu layout | Generik, tidak preview karakter tema |

## Rencana refactoring

Bukan menghapus semua warna, tapi:

1. **Sentralisasi token tema** di `ui-components.css`: setiap `data-ui` mendefinisikan token semantik:
   `--ui-app`, `--ui-workspace`, `--ui-panel`, `--ui-card`, `--ui-card-hover`, `--ui-input`,
   `--ui-modal`, `--ui-dropdown`, `--ui-nav`, `--ui-floating`,
   `--ui-text`, `--ui-text-muted`, `--ui-text-inverse`, `--ui-accent`,
   `--ui-border`, `--ui-border-strong`, `--ui-border-focus`,
   `--ui-shadow-card`, `--ui-shadow-panel`, `--ui-shadow-floating`, `--ui-shadow-inset`,
   `--ui-font`, `--ui-radius`, `--ui-icon-stroke`.
2. **Ubah override komponen** dari hex statis → `var(--ui-*)`.
3. **Hapus/kurangi `!important`** — cukup di utility DaisyUI yang memang perlu.
4. **ThemeSelector kartu** = mini preview per tema (retro bevel, neo border tebal, acrylic blur, neumo dual shadow, editorial serif rule).
5. **Settings** pakai token & kartu preview baru.
6. `bg-black` di scan page → ganti token + catatan pengecualian yang disengaja.

Catatan: screenshot sebelum/sesudah **tidak tersedia** (no browser automation). Aku akan tandai NOT TESTED di laporan akhir.

## Refactoring applied

- `ui-components.css`: 8 scope `data-ui` kini punya token semantik lengkap (`--ui-app` … `--ui-icon-stroke`). Semua override komponen memakai `var(--ui-*)`, termasuk edge/shadow/toast warna. Hex statis hanya ada di deklarasi token.
- `ThemeSelector.svelte`: kartu layout diganti mini-preview yang memakai scope `data-ui` dan token yang sama dengan runtime.
- `ScanTimeline.svelte`: `text-white` → `text-base-100`.
- `scan/+page.svelte`: `bg-black` pada viewfinder kamera didokumentasikan sebagai pengecualian kontras decoder.
- `npm run check`: 0 error, 0 warning. `npm run build`: sukses.
- Visual/browser regression: **NOT TESTED** (tidak ada browser automation di environment ini).
