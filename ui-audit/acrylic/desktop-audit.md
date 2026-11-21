# Acrylic UI — Desktop Audit

File: `AcrylicLayout.svelte`, `acrylic-icons.ts`.

| Kriteria | Bukti kode | Status |
|---|---|---|
| Atmospheric background | gradient + 3 orb blur(60px) + grid overlay | PASS |
| Floating nav rail | `.acr-rail` fixed, kiri tengah, rounded, acrylic | PASS |
| Floating command bar | `.acr-command` pill, search input, hint, bell, profile | PASS |
| Layered central workspace | hero layer + main layer + inspector (grid) | PASS |
| Right inspector | `.acr-inspector` status rows, activity, slot context | PASS |
| Floating quick toolbar | `.acr-quickbar` fixed bawah | PASS |
| Backdrop blur | `.acrylic` blur(18px) saturate(150%) | PASS |
| Thin translucent border | `1px solid rgba(255,255,255,.28)` | PASS |
| Fallback tanpa blur | `@supports not (backdrop-filter)` background solid | PASS |
| Clean outline icons | acrylic-icons stroke 1.7 | PASS |

Temuan: search field, bell, profile **dead** (tidak ada handler, TODO di kode) → P1. Inspector bisa ditutup tapi tidak tersimpan preferensinya (P3). Data table di slot generic (P2).
