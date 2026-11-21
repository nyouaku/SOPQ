# Acrylic UI — Mobile Audit

| Kriteria | Bukti | Status |
|---|---|---|
| Compact transparent header | `.acr-mobile-top` acrylic sticky | PASS |
| Large acrylic hero panel | `.acr-hero` padding 1.2rem, judul 2rem | PASS |
| Stacked translucent cards | `.acr-card` + hero + content | PASS |
| Horizontal floating icon dock | `.acr-actions` 4 ikon | PASS |
| Acrylic bottom navigation | `.acr-mobile-nav` fixed | PASS |
| Bottom-sheet modal | **tidak ada** bottom-sheet khusus acrylic | FAIL P1 |
| Backdrop blur secukupnya | blur 18px | PASS |
| Fallback | @supports global | PASS |

Temuan: bottom-sheet modal wajib prompt tapi belum dibuat (P1). Search desktop tidak tersedia di mobile (P2).
