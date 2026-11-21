# Remediation Roadmap — UI Audit

Status: **R1 SELESAI. R2–R4 masih draft.**

## Ringkasan masalah berulang

1. **Komponen konten global belum ter-theme** (RET/NEO/ACR/NEU/EDI) — DaisyUI generic dipakai bersama 5 layout.
2. **Kontrol dekoratif / dead** — search Acrylic & Editorial, bell/profile Acrylic, bell Neumo, window dots Neo.
3. **Named slot belum dipakai halaman** — `data` (Neo/Editorial), `context` (Retro/Acrylic/Neumo).
4. **Responsive device test belum dilakukan** — semua layout baru diverifikasi statis (svelte-check/build), belum di viewport nyata.
5. **Emoji di ThemeSelector** — bukan ikon produksi layout, tapi mengurangi konsistensi sistem ikon.

## Daftar prioritas (P1)

| ID | Desain | Perbaikan | File |
|---|---|---|---|
| ACR-001 | Acrylic | Hubungkan search ke navigasi/filter nyata | `AcrylicLayout.svelte` |
| ACR-002 | Acrylic | Bell & profile ke halaman notifikasi/profil | `AcrylicLayout.svelte` |
| ACR-003 | Acrylic | Bottom-sheet modal acrylic | komponen baru + layout |
| EDI-001 | Editorial | Search strip fungsional | `EditorialLayout.svelte` |
| — | Semua | Uji responsif di 10 viewport (BAB 6) | manual/Playwright |

## Urutan perbaikan disarankan

### Phase R1 — Fungsional dead controls ✅ SELESAI
- Search Acrylic & Editorial → arahkan ke halaman pencarian yang sudah ada / filter.
- Bell Acrylic & Neumo → `goto('/notifications')` atau buka panel notifikasi.
- Profile Acrylic → `goto('/settings')`.
- Acceptance: semua tombol-layout punya aksi nyata; tidak ada `TODO` di layout.

### Phase R2 — Mobile component gaps (risiko sedang)
- Acrylic bottom-sheet (dipakaiMobile menu & context).
- Editorial mobile data slot + horizontal scroll table.
- Retro mobile START menu full-width.
- Acceptance: sesuai BAB 3 mobile tiap desain; uji 360×800 & 430×932.

### Phase R3 — Theme-aware content components (risiko tinggi, dampak besar)
- Buat wrapper komponen per desain: button, input, table, modal, toast, badge, pagination.
- Scoped styling: `.retro-scope`, `.neo-scope`, `.acr-scope`, `.neumo-scope`, `.ed-scope`.
- Hindari duplication: gunakan satu komponen `UiButton` dengan `variant` prop.
- Acceptance: komponen konten mengikuti karakter desain aktif, bukan DaisyUI generic.

### Phase R4 — Verification
- Playwright screenshots desktop (1280/1366/1440/1920) & mobile (360/375/390/430).
- Re-run `svelte-check`, build, test.
- Update laporan audit dengan bukti screenshot.

## Kompleksitas (perkiraan)

| Phase | Kompleksitas | Estimasi file |
|---|---|---|
| R1 | Rendah | 3–4 file layout |
| R2 | Sedang | 3 file + 1 komponen baru |
| R3 | Tinggi | 8–12 file komponen |
| R4 | Sedang | test + docs |

## Tunggu persetujuan
R1 dikerjakan & disetujui. R2–R4 menunggu persetujuan berikutnya.
