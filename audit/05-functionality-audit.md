# 05 — Functionality Audit (button → form → modal → filter → pagination → navigation, front→back→DB)

Method: static trace of every interactive element to its handler, to the HTTP call, to the handler, to the usecase, to the repository, to the table. **No live server was run** during this audit; items that require a running stack are explicitly marked **NOT TESTED**.

## A. Route inventory and reachability

| Route | In menu | Reachable | Backend | Works end-to-end? |
|---|---|---|---|---|
| `/login` | public | yes | `POST /auth/login` | **YES** |
| `/` dashboard | yes | yes | `GET /dashboard/summary` | **PARTIAL** — tiles populate; "Aktivitas terbaru" never shows data (field absent, BE-09) |
| `/scan` | yes | yes | sessions list + `POST .../scan` | **PARTIAL** — manual + camera work; saved scans **rejected or queued wrongly** (FN-02, BE-08) |
| `/sessions` | yes | yes | `GET /stock-opname/sessions` | **YES** (list + filter + paginate) |
| `/sessions/[id]` | via list | yes | `GET .../sessions/:id`, `/details`, `POST .../complete`, `.../approve`, WS | **PARTIAL** — approve/complete work; WS never delivers scans (BE-13, FN-06) |
| `/products` | yes | yes | CRUD + import/export | **PARTIAL** — CRUD works; import/export bypass refresh; "Gudang" column always `-` (FN-07) |
| `/warehouses` | yes | yes | warehouses/locations/boxes | **YES** (read + create) |
| `/reports` | yes | only supervisor/admin | `GET /reports/stock-opname` + export | **NO** — endpoint returns zero rows always (BE-05); exports empty (BE-06) |
| `/activity` | yes | only supervisor/admin | `GET /activity-logs` | **YES** |
| `/settings` | yes | admin only | `GET /auth/users`, `/auth/roles`, `POST /auth/register` | **YES** for admin; staff sees menu then 403 (FE-03) |
| `/mobile` | **no (not in menu)** | by URL only | same as `/scan` | **PARTIAL** — orphan page, duplicate of `/scan` |

## B. Findings

### FN-01 — CRITICAL — "New Session" opens a broken detail page
`dashboard/src/routes/sessions/+page.svelte:89`
```svelte
<button class="btn btn-primary" on:click={() => goto('/sessions/new')}>
```
There is **no** `sessions/new/+page.svelte`. SvelteKit resolves `/sessions/new` against `sessions/[id]`, so it loads the detail page with `id="new"` and fires `GET /stock-opname/sessions/new` → handler `uuid.Parse` fails → 400. The user sees an error/empty detail view, not a create form.

**Impact:** the only visible "create a session" affordance in the whole dashboard is dead. Sessions can only be created by calling `POST /api/v1/stock-opname/sessions` directly.

**Fix:** build `sessions/new/+page.svelte` (warehouse + optional location picker → `POST /sessions`), or remove the button.

---

### FN-02 — HIGH — Every non-network save failure is silently queued (`/scan`)
`dashboard/src/routes/scan/+page.svelte:334-354`
```ts
const response = await apiClient.post(...);
if (response.success) { ... } else { showMessage(response.error?.message || 'Gagal menyimpan data', 'error'); }
} catch {
    enqueue({ ... });           // ← ANY thrown error
    showMessage('Offline: scan masuk antrian', 'info');
}
```
`apiClient.request()` **throws** for 401-after-failed-refresh and for JSON parse failures. A 400/422/403 returns `{success:false}` and does *not* throw — but `/mobile` does the opposite and enqueues on `!response.success` too:
```ts
} else {
    // Server menolak (bukan masalah jaringan) — simpan untuk dikirim ulang.
    enqueue(payload);
```
So: **permanently rejected scans (bad input_method, session already completed, product not found) are pushed into the offline queue and retried forever.** Combined with FN-03 they never leave.

**Fix:** distinguish transport failure from server rejection. Enqueue only on `TypeError`/network error; on 4xx show the message and discard.

---

### FN-03 — HIGH — `flushQueue` never discards permanently rejected items
`dashboard/src/lib/offline/queue.ts:105-125`
```ts
if (res.success) { sent++; }
else { failed++; remaining.push(scan); }   // ← no rejection classification
```
A scan rejected with 422 is re-queued on every flush, forever. `pendingScans` grows monotonically; the badge counts dead items; the "online" listener re-flushes on every reconnect. No attempt counter, no dead-letter, no user-visible list of stuck scans.

**Fix:** keep a `attempts` field; after N attempts surface the item as failed with a manual retry/discard action.

---

### FN-04 — HIGH — No idempotency; comment claims one exists
`queue.ts:10-12` claims:
> Kunci unik: `session_id + barcode + timestamp`, supaya flush berulang tidak menggandakan baris (backend belum punya idempotency key).

The implementation (`enqueue`) is a bare `scans.push(item)` — no key computation, no dedupe, no comparison. And the backend's `POST /sessions/:id/scan` has no idempotency key either, so a replay after a timeout that *did* reach the server creates a duplicate `stock_opname_detail` row. Duplicate scans double-count in every report aggregation.

**Fix:** send a client-generated `idempotency_key` (UUID per scan) and `UNIQUE(stock_opname_id, idempotency_key)` in the DB — a real constraint, not a comment.

---

### FN-05 — HIGH — No flush re-entrancy guard on web
`flushQueue()` has no `_isSyncing` equivalent (the Flutter `SyncService` does). It can be invoked concurrently by the `online` event on `/scan` **and** `/mobile`, each reading the same localStorage snapshot and POSTing the same items → guaranteed duplicates.

---

### FN-06 — HIGH — WebSocket never delivers scan progress
`dashboard/src/routes/sessions/[id]/+page.svelte:61` opens the socket; the client expects `{type:'new_scan', detail}` and `{type:'session_update', session}`. Backend `BroadcastProgress` is a TODO stub emitting only `{"session_id":..., "type":"progress_update"}` (BE-13). `SubmitScan` never calls any broadcast function at all. Live detail table therefore only updates on manual reload/pagination.

Client also sends the **JWT in the query string** (`?token=`), which lands in nginx access logs and browser history; backend doesn't read it anyway.

---

### FN-07 — MEDIUM — Products "Gudang" column always `-`
`product_repository.go:149-155` `List` preloads `Category, Supplier, Box` — not `Box.Location.Warehouse`. The page reads `warehouseName(product)` from that deep path. `FindByID` does preload it, so the same product shows a warehouse in the edit modal and `-` in the table.

---

### FN-08 — MEDIUM — Reports and Activity pages shown to roles the API rejects
See FE-03. Staff clicks Reports → `403`; the empty-state branch shows "Gagal memuat log" with a Retry that also 403s.

---

### FN-09 — MEDIUM — Session detail pagination resets nothing; filter state diverges
`sessions/[id]/+page.svelte` keeps `filter.page` in a mutable object and calls `loadDetails()` directly — no URL state. Browser back/forward loses position. Same pattern on `/products` and `/sessions`. Only `/reports` and `/activity` use store-driven params.

---

### FN-10 — MEDIUM — Mobile scanner posts to a hardcoded fake session
`mobile/lib/features/scanner/screens/scanner_screen.dart:26`
```dart
// Set dummy session ID for testing - in production this should come from session selection
ref.read(scannerProvider.notifier).setSessionId('dummy-session-id');
```
There is no session-selection screen anywhere in `mobile/lib`. Every scan from the Flutter app posts to `POST /stock-opname/sessions/dummy-session-id/scan` → UUID parse failure → 400. **The mobile app cannot record a single real scan.**

Also: `scanner_screen` never calls `completeSession()`, has no logout, no product search, no session list. `StockOpnameRemoteDataSource.getSessions()`/`getSession()`/`getSessionDetails()`/`completeSession()` are all implemented but **never called from any screen**.

---

### FN-11 — MEDIUM — Filters applied only on explicit button; no debounce/auto-apply
`/products` requires Enter or the search button; `/reports` requires "Terapkan". Consistent with the design, but `/products` search input has no debounce and re-issues a request per Enter press. Acceptable.

### FN-12 — LOW — `/mobile` is orphaned and duplicative
490 lines duplicating `/scan`'s camera + queue logic, with its own bug (`captureFrame()` is a no-op with a TODO, so its "fallback" path silently scans nothing). Not linked from `menu.ts`. Two divergent implementations of the same workflow.

### FN-13 — LOW — `+error.svelte` exists but pages swallow errors
Every page catches into `console.error` and renders an empty state. A 500 looks identical to "no data". `/reports` and `/activity` are the only pages with a distinct error branch.

---

## C. NOT TESTED
- Live login → session create → scan → complete → approve cycle against a running stack.
- Actual HTTP status codes for every handler branch.
- Postgres constraint behavior on insert (no DB available).
- `flutter run` on a device/emulator.
- Camera/barcode paths in a real browser (permissions, `BarcodeDetector` availability).
- CSV/Excel download content.

All conclusions above are from source inspection; no result is fabricated.
