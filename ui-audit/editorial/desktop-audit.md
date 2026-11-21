# Editorial UI — Desktop Audit

File: `EditorialLayout.svelte`, `editorial-icons.ts`.

| Kriteria | Bukti | Status |
|---|---|---|
| Publication masthead | `.masthead` border-bottom 3px double, brand serif 3rem, volume/tanggal | PASS |
| Horizontal editorial nav | `.ed-nav` thin rule, title case uppercase, active underline burgundy | PASS |
| Asymmetric two-column | `.ed-columns` grid `1.9fr 0.85fr`, gap 2.5rem | PASS |
| Main feature | `.ed-feature h1` Georgia clamp 2.6–5.2rem, standfirst serif 1.15rem | PASS |
| Right highlights | `.ed-sidebar` numbered 01–06 + pull quote + metadata | PASS |
| Full-width data section | `.ed-data` grid-column penuh, slot `data` | PASS |
| Footer metadata | `.ed-footer` 3 kolom metadata, thin rule | PASS |
| Whitespace & thin rules | padding 2rem, `.ed-rule` 1px, double rule | PASS |
| Serif + sans combination | Georgia display + Inter body | PASS |

Temuan: search strip ada tapi tidak memfilter/navigate (P1); data section default placeholder karena halaman belum memakai slot `data` (P2).
