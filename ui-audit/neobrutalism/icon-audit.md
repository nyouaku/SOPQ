# Neobrutalism — Icon Audit

Sumber: `neo-icons.ts` — 13 SVG stroke 2.2, square linecap, warna hitam/contrast.

| Fungsi | Ikon | Status |
|---|---|---|
| Dashboard | grid 4 kuadrat | PASS |
| Analytics | bar chart | PASS |
| Users | users | PASS |
| Inventory | boxes/package | PASS |
| Transactions | receipt | PASS |
| Settings | gear | PASS |
| Search | magnifier | PASS |
| Notification | bell | PASS |
| Add | plus-circle (path plus; tidak benar-benar circle) | PARTIAL |
| Edit | pencil | PASS |
| Delete | trash | PASS |
| Filter | sliders horizontal | PASS |
| Export | download | PASS |

Temuan:
- `plus` berupa silang tanpa lingkaran — prompt meminta plus-circle (P3).
- Stroke 2.2 konsisten di seluruh ikon → PASS.
- Ikon navigasi memakai `neoIconFor`; tombol aksi plus/edit/trash/filter/export **didefinisikan tapi belum dipasang di komponen layout** (P2).
