# Neobrutalism — Desktop Audit

File: `dashboard/src/lib/components/layouts/NeoLayout.svelte`, `neo-icons.ts`

| Kriteria | Bukti kode | Status |
|---|---|---|
| Oversized brand header | `.neo-top` bg `--n-yellow`; `.brand-mark` 1.7rem font 900; horiz nav | PASS |
| Horizontal navigation | `.neo-nav` flex wrap, 6 menu pertama, active blue solid | PASS |
| Asymmetric bento workspace | `.bento` grid `1.15fr 1.7fr 0.85fr`; hero kolom 1, main kolom 2 spanning 2 row, feed kolom 3 | PASS |
| Highlight KPI card | `.bento__hero` bg lime, min-height 14rem, `.oversized` clamp 1.9–3.4rem, bentuk geometris `.hero-shape` rotate(18deg) | PASS |
| Right activity feed | `.bento__feed` daftar numbered 01–07 + `.feed-user` | PASS |
| Bottom data section | `.bento__table` grid-column 1/span 3, slot `data` | PASS |
| Floating action button | `.neo-fab` fixed kanan bawah, bg red, hard shadow | PASS |
| Thick black border 2–4px | `--n-line: 3px solid #0d0d0d` | PASS |
| Hard offset shadow no blur | `--n-shadow: 6px 6px 0` | PASS |
| Bright accents | yellow `#ffe23d`, blue `#2f5bff`, red `#ff4d3d`, lime `#b7f35b`, bg `#f7f3dc` | PASS |
| Numbered labels & geometric shapes | `.kicker` "01/02/03/04", `.hero-shape`, `.table-triangle` | PASS |

Temuan:
- `neo-window__dots` tiga `<i>` putih di header biru — dekoratif, bukan kontrol window fungsi (P3).
- `slot name="data"` hanya dipakai jika halaman mengirim — secara default ada placeholder (P2).
- Interaksi drag tidak ada di bento (prompt tidak mewajibkan drag).
