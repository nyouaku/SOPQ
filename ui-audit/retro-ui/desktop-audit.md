# Retro UI — Desktop Audit

File: `RetroLayout.svelte`, `retro-icons.ts`.

| Kriteria | Bukti kode | Status |
|---|---|---|
| Application title bar | `.titlebar` dengan START, judul, versi, jam, power | PASS |
| Left navigation | `.nav-panel` compact, ikon + label, active olive | PASS |
| Central floating window | `.retro-window` draggable, min/max/reset, bevel shadow | PASS |
| Right contextual panel | `.context-panel` SYSTEM INFO, aktivitas, shortcut + slot | PASS |
| Bottom taskbar | `.taskbar` START, task, tray network/clock/user | PASS |
| Monospace + pixel texture | `Courier New`, `.retro-noise` scanline | PASS |
| Beige/olive/orange/brown | token `--r-bg`, `--r-olive`, `--r-orange`, `--r-ink` | PASS |
| Bevel + hard/inset shadow | border color trik `#fff/var(--r-shadow)`, `.press:active` inset | PASS |

Temuan: window drag & minimize/maximize **tidak diuji runtime** (NOT TESTED). Scanline z-index 90 menutupi konten (tipis, P3).
