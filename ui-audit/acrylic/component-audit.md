# Acrylic UI — Component Audit

| Komponen | Status | Catatan |
|---|---|---|
| Acrylic navigation | PASS | rail + mobile nav |
| Translucent cards | PASS | panel & inspector |
| Frosted search bar | VISUAL PASS / FUNGSIONAL FAIL | input ada, searchSubmit TODO |
| Layered modal | GAP P2 | tidak ada modal khusus |
| Floating dropdown | GAP P3 | tidak ada |
| Acrylic bottom sheet | FAIL P1 | wajib mobile, belum ada |
| Transparent icon buttons | PASS | ghost-btn, quick-btn |
| Data table readable surface | PARTIAL | slot data generic (DaisyUI table) |
| Status badges | PASS | status-badge mobile |

Arah perbaikan: baut `AcrylicSheet.svelte` + `AcrylicSearch` yang benar-benar memfilter/ navigate.
