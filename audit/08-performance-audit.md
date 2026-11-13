# 08 — Performance Audit

**No load test was run.** No stack was started during this audit. Everything below is static analysis of query shape, allocation, transfer size and unbounded growth. Measured numbers are marked **NOT TESTED**.

## A. Verified facts

| Item | Finding |
|---|---|
| DB pool | `SetMaxIdleConns(10)`, `SetMaxOpenConns(100)`, `SetConnMaxLifetime(1h)` — correctly configured (`postgres.go:45-47`). |
| N+1 in dashboard | **Yes** (BE-09) — worst case ~1 M rows per dashboard load. |
| Unbounded queries | `productRepo.List(Limit: 9999)` and `stockOpnameDetailRepo.FindByStockOpnameID(..., 9999, 0)` per session. |
| Gzip | On (nginx, `gzip_types` includes `application/json`). |
| Static asset caching | Service worker cache-first for hashed assets; nginx sets no `Cache-Control`/`expires` for `/` — **medium**. |
| API response caching | None (deliberate — the SW skips `/api/`). |
| N+1 in session detail | `List` preloads `Category, Supplier, Box` (BE-11 note) — no per-row query. |
| Product search | `ILIKE '%?%'` on 3 columns, no index used. |
| Report path | Returns zero rows (BE-05) — **unmeasurable until implemented**. |

## B. Findings

### PERF-01 — CRITICAL — Dashboard summary is O(sessions × details) round trips
`internal/usecase/dashboard_usecase.go:41-98`

```
1 query  : products LIMIT 9999           → up to 9 999 rows, all columns, fully hydrated
1 query  : sessions LIMIT 100           → up to 100 rows
N queries: details LIMIT 9999 per session → up to 100 × 9 999 = 999 900 rows
```

Plus each `FindByStockOpnameID` preloads related rows. On a seeded 5 000-SKU catalogue and a few dozen sessions this is tens of thousands of rows per page load, transferred over the wire and then **discarded** — only `len()` and two counters are used.

Two frontend consequences ride along: the dashboard is the app's landing page (every login pays this), and `refetchInterval` / remount on navigation repeats it.

**Fix:** three aggregates, no row transfer:
```sql
SELECT count(*) FROM products WHERE deleted_at IS NULL;
SELECT count(*), count(*) FILTER (WHERE d.difference <> 0)
FROM stock_opname_detail d JOIN stock_opname s ON s.id = d.stock_opname_id
WHERE s.status IN ('in_progress','completed');
```
**Expected gain: NOT MEASURED**, but the row-transfer reduction is ~4 orders of magnitude on the seeded dataset.

---

### PERF-02 — HIGH — `products.name ILIKE '%x%'` cannot use an index
`product_repository.go:130`
```go
query.Where("name ILIKE ? OR barcode ILIKE ? OR product_code ILIKE ?", ...)
```
`ILIKE '%q%'` is a non-sargable predicate. `barcode`/`product_code` have unique B-tree indexes (`000001:117-118`) that **are not used** because of the leading wildcard. `performance_indexes.sql` *intends* a `gin_trgm_ops` index on `name` — that file never runs (DB-01, DB-02), so the index does not exist either.

Two problems stacked: a full table scan, and an OR across three columns which also defeats a single partial index.

**Fix:** `CREATE EXTENSION pg_trgm; CREATE INDEX ... ON products USING gin (name gin_trgm_ops);` then rewrite as `WHERE name % ?` (trigram) **or** — cheaper and simpler — split: exact/prefix match on `barcode`/`product_code` (`LIKE 'q%'`, index-usable) and trigram only for `name`. Do not OR them into one predicate.

---

### PERF-03 — HIGH — `/reports` and `/activity` are paginated but the export is not
`ExportStockOpnameReport` forwards the same URL params including `page`/`limit` and ignores them (BE-07). Once BE-05 lands and reports return real rows, the CSV export will serialize the **entire** result set with no row cap — a filter-less export on a 500 k-row detail table means a 500 k-row CSV built in memory.

**Fix:** stream with `csv.Writer` writing directly to `c.Writer`, and cap at a documented maximum (e.g. 100 k) with an explicit error above it.

---

### PERF-04 — HIGH — Excel export builds the whole workbook in memory
`export_usecase.go` uses `excelize` (a zip + XML model). `ExportProductsToExcel` populates all products into one sheet in memory, then `file.Write(c.Writer)`. Memory scales with catalogue size; there is no `StreamWriter` usage. With the 5 000-SKU seed this is fine; at 200 k SKUs it is a memory spike on the API container, not a slow response.

**Fix:** `excelize.NewStreamWriter` + `SetRow` — bounded memory, no API change.

---

### PERF-05 — MEDIUM — DB index set is thin and the "performance" file never ran
Present in `000001_init_schema.up.sql` (15 indexes). Missing for the queries the code actually issues:

| Query | Needs |
|---|---|
| session list ordered by `created_at` desc, filtered `status` | `(status, created_at DESC)` composite — currently two single-column indexes, so Postgres can use one and sort. |
| session detail ordered by `scanned_at desc` | `(stock_opname_id, scanned_at DESC)` — currently `idx_so_detail_opname(stock_opname_id)` only, then sort. |
| dashboard aggregate | partial index `WHERE status IN ('in_progress','completed')`. |
| `activity_logs` filtered by date range | has `created_at`; ok. |
| `notifications` unread count | `(user_id, is_read)` present; ok. |

`performance_indexes.sql` would have supplied some of these — it is broken and also creates three indexes that duplicate the ones already in `000001` (e.g. `idx_stock_opname_status`, `idx_sessions_user`, `idx_activity_logs_user`). So even if fixed, it adds write cost for no read benefit.

**Fix:** delete `performance_indexes.sql`; add only the three composites above as a properly numbered migration.

---

### PERF-06 — MEDIUM — `Save()` writes the full session row on approve
`stock_opname_repository.go:60` — `Save()` issues `UPDATE` with every column. Two effects: unnecessary WAL for unchanged fields, and the lost-update race (BE-10). Same row, both costs.

---

### PERF-07 — MEDIUM — Frontend bundle ships Chart.js unconditionally on `/`
`routes/+page.svelte:5` imports `chart.js` and `Chart.register(...registerables)` in the landing page's module scope. `registerables` pulls **every** controller, scale, and plugin (~200 KB min). The page uses one chart type.

**Fix:** register only what is used (`BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend`) — typical saving 120–160 KB gzipped-time. Or `await import('chart.js')` inside `onMount`.

**Bundle size: NOT MEASURED** — no `vite build --report` run this session.

---

### PERF-08 — MEDIUM — PWA precaches the whole asset manifest on install
`service-worker.ts:14` `const ASSETS = [...build, ...files]` then `cache.addAll(ASSETS)` on `install`. Every build asset — including any large chunks — is downloaded on first visit, before the app is interactive. With Chart.js and the QR library in `build`, first load is heavier than a normal SPA.

**Fix:** precache the shell only; runtime-cache the rest on first use. Or accept it — but know the trade.

---

### PERF-09 — MEDIUM — Notification polling with no `staleTime`
`queries.ts:128` `refetchInterval: 60_000` on `notificationsQuery`, no `staleTime`, mounted by `NotificationBell` in every shell. Every page navigation + every 60 s = a full round trip. At 10 concurrent supervisors that is 600 req/hour for a bell icon.

**Fix:** `staleTime: 30_000` and `refetchIntervalInBackground: false` (stop polling when the tab is hidden).

---

### PERF-10 — LOW — Mobile: `PrettyDioLogger(requestBody: true, responseBody: true)` in release
Beyond the security problem (SEC-10), string-formatting every request and response body is measurable overhead on a low-end scanner device, and it keeps large JSON blobs alive in memory longer.

### PERF-11 — LOW — Mobile offline flush has no batch endpoint
`SyncService` posts queued scans one at a time. A 500-scan backlog after a warehouse loses signal is 500 sequential round trips over a possibly-weak connection. No `POST /scan/batch`.

**Fix:** add a batch endpoint (also the natural home for an idempotency key, FN-04).

### PERF-12 — LOW — No `ETag`/`Last-Modified` on any GET
Every list query re-transfers full JSON. Fine at current scale; trivially added later via `gin` middleware if needed.

### PERF-13 — NOT TESTED — Response-time and throughput baselines
No `k6`/`hey`/`ab` run. No p50/p95/p99 numbers exist for any endpoint. **Any performance claim about this system beyond "the query shapes are wrong" is unverified.**

## C. Positive performance choices
- Connection pool sized correctly for a single API instance.
- `GORM` logger set to `Silent` unless `Host == "localhost"` — no SQL logging in production.
- Service worker deliberately excludes `/api/` from caching — correct call for stock data.
- Hashed assets served cache-first.
- Pagination capped at `limit <= 100` on every list handler; `MasterData` handlers share one clamp.
- `Preload` used instead of lazy per-row association loads in product/session lists.
- Partial indexes `WHERE deleted_at IS NULL` on the soft-deleted tables — correct and selective.
- `difference` is a **generated stored column**, so report aggregation reads a plain integer instead of computing per row.

## D. Assessment
Nothing here is a scalability wall at the target scale (a handful of warehouses, thousands of SKUs). What is here is one genuinely bad query pattern (PERF-01) on the app's landing page, one non-sargable search (PERF-02) that will degrade linearly with catalogue size, and a set of unbounded exports that become the first real problem once reports start returning data. The infrastructure around them (pool, gzip, partial indexes, SW policy) is correct.
