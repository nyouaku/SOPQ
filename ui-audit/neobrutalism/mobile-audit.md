# Neobrutalism — Mobile Audit

File: `NeoLayout.svelte` blok `@media (max-width: 900px)`.

| Kriteria | Bukti | Status |
|---|---|---|
| Compact brand header | `.neo-mobile-top` sticky bg kuning, brand + hamburger | PASS |
| Oversized page title | `.neo-mobile-title` 2.4rem weight 900 uppercase | PASS |
| Highlight KPI card | `.mobile-hero` kuning + border 3px + hard shadow; angka 4rem | PASS |
| Stacked cards | `.mobile-card`, `.neo-window` equivalent | PASS |
| Horizontal swipeable analytics | `.mobile-swipe` overflow-x auto + scroll-snap; `.swipe-card` flex 0 0 8.5rem | PASS |
| Vertical activity list | `.mobile-feed` numbered | PASS |
| Floating bottom navigation | `.neo-mobile-nav` fixed, 5 ikon, active red | PASS |
| Komposisi beda dari desktop | bento grid → stacked block | PASS |

Temuan:
- Breakpoint 900px → tablet 834×1194 masuk mobile layout (sesuai intent "mobile", perlu uji runtime).
- `.neo-mobile-menu` sticky `top:3.3rem` bisa menimpan konten saat header overflow (P3).
