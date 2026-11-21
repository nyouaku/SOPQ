# Neobrutalism — Component Audit

| Komponen | Bukti | Status |
|---|---|---|
| Thick-bordered buttons | `.neo-nav__item`, `.neo-ico-btn`, `.press` (3px + 3px shadow) | PASS |
| Hard-shadow cards | hero/window/feed/table | PASS |
| Bold form inputs | **tidak ada** input neo-specific; global DaisyUI | GAP P2 |
| Geometric badges | `.badge-solid`, `.badge-outline` | PASS |
| Oversized KPI | `.oversized` clamp + hero-shape | PASS |
| Strong data table | `.table-zone` border+shadow, isi dari slot (default placeholder) | PARTIAL |
| High-contrast alerts | tidak ada alert neo-specific | GAP P2 |
| Distinctive pagination | tidak ada | GAP P3 |
| Custom icon buttons | `.neo-ico-btn`, `.wc`-style nav | PASS |
| Toast | Toast global (DaisyUI) belum neo-styled | GAP P2 |

Temuan: identitas kuat di struktur & tombol; komponen konten (form/table/toast/modal) masih generic.
