# Neumorphism — Desktop Audit

File: `NeumoLayout.svelte`, `neumo-icons.ts`.

| Kriteria | Bukti | Status |
|---|---|---|
| Centered header | `.n-top` grid 1fr auto 1fr, brand kiri, pill tengah, user kanan | PASS |
| Compact floating vertical dock | `.n-dock` fixed-ish 4.5rem, icon-only | PASS |
| Large central workspace | `.n-center` panel hero + main | PASS |
| Neumorphic control panels | `.soft-raise` dual shadow, `.n-inset` inset shadow | PASS |
| Right statistics panel | `.n-right` stats + context | PASS |
| Bottom status | `.n-status` fixed | PASS |
| Soft monochromatic palette | `--n-bg #e8e9ee`, ink muted, accent `#6f7fc1`, sage | PASS |
| Rounded components | radius 0.8–1.5rem | PASS |
| Toggle & progress | `.n-toggle`, `.n-progress` | PASS |

Temuan: bukan sidebar besar/bento — sesuai. Kontras dijaga dengan border tipis & label teks (bukan shadow saja). Bell di header dead (P2). Toggle "Kompak" berfungsi menyembunyikan hero (PASS).
