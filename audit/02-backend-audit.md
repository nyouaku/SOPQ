# 02 — Backend Audit (Go / Gin / GORM)

Module: `github.com/nyouaku/SOPQ/backend`, `go 1.25.0`.
Dependencies look healthy and current for the stack (gin 1.10, gorm 1.30, jwt/v5 5.2.1, excelize 2.11.0). No obviously abandoned package.

## Build / static analysis
| Command | Result |
|---|---|
| `go build ./...` | **PASS** (exit 0) — verified this session |
| `go vet ./...` | **FAIL** — `internal/usecase/stock_opname_usecase_test.go:144:73: undefined: domain.ProductFilter` |
| `go test ./...` | **FAIL — does not compile** — see 09-testing-audit.md |

`go vet` failing means the repo can't run a standard CI gate as-is.

## Findings

### BE-01 — CRITICAL — `StockOpname.code` collides on 2nd session of any day
`internal/usecase/stock_opname_usecase.go:68-71`
```go
code := fmt.Sprintf("SO-%s-0001", now.Format("20060102"))
// TODO: implement auto-increment per day
```
`stock_opname.code` is `UNIQUE` (`000001_init_schema.up.sql`). Session #2 on the same date returns a DB unique-violation, surfaced as HTTP 400 `CREATE_FAILED`. **Blocks normal operation:** every warehouse/day combination gets exactly one session, ever.

**Root cause:** placeholder not replaced.
**Fix:** per-day counter — either `SELECT count(*)+1 ... WHERE code LIKE 'SO-YYYYMMDD-%'` inside a transaction with retry, or a Postgres sequence keyed by day, or include warehouse short-code + nanoseconds and relax uniqueness to `(warehouse_id, code)`.

---

### BE-02 — CRITICAL — `SubmitScan` returns the wrong row
`internal/usecase/stock_opname_usecase.go:171-177`
```go
details, _, _ := u.stockOpnameDetailRepo.FindByStockOpnameID(ctx, input.SessionID, 1, 0)
if len(details) > 0 {
    return &details[0], nil
}
```
`FindByStockOpnameID` orders `scanned_at DESC` (`stock_opname_repository.go:140`). With concurrent scanners, `details[0]` is the newest scan in the session — **not** the row just inserted. Client receives another user's product/qty. Also ignores the error from the reload (`_, _`), silently falling back to the un-preloaded `detail`.

**Fix:** `SELECT ... WHERE id = <detail.ID>` after create, or `Preload` inside the create path.

---

### BE-03 — CRITICAL — Refresh token never validated against DB
`internal/usecase/auth_usecase.go:202-203`
```go
// TODO: Di production, tambahkan verifikasi refresh token hash di database
// Untuk PHASE 2, kita skip verifikasi hash dan langsung generate token baru
```
A **revoked** or unknown refresh token still mints a fresh access + refresh pair until the JWT `exp` passes. `Logout` → `RevokeByID` therefore does not end the session; `sessions.refresh_token_hash` is written but never read for verification.

**Fix:** `repo.FindActiveByRefreshTokenHash(ctx, utils.HashToken(rt))` → nil ⇒ 401; rotate (revoke old, insert new) inside one transaction. The repository interface already has no such method; the failing unit test references `FindActiveByRefreshTokenHash`, so the intended design exists.

---

### BE-04 — HIGH — Stock table never written after an opname
`stockRepo.UpdateWithOptimisticLock` and `stockRepo.CreateOrUpdate` are **defined but never called** — grep across `internal/` returns only the definitions. `CompleteSession` (line 180) and `ApproveSession` (line 211) only flip `status`/`completed_at`/`approved_by`.

Consequence: `stock.quantity_system` never reflects the physical count. Every subsequent opname compares against the same stale baseline; the whole reconciliation loop is one-directional and never closes.

**Fix:** in `ApproveSession` (approval = the accounting event), wrap detail iteration + `stockRepo.CreateOrUpdate` in one transaction.

---

### BE-05 — HIGH — Reports endpoint is a hard-coded stub
`internal/usecase/report_usecase.go`
```go
// TODO: Implement real query with joins
// Untuk PHASE 5, return dummy data dulu
return []StockOpnameReportRow{}, 0, nil
```
`GetStockOpnameReport` returns **zero rows, always**. Therefore `/reports/stock-opname`, the Reports page, summary tiles, pagination, and the CSV export are all permanently empty. `reportUsecase.db interface{}` is an unused field (`interface{}` typed — a placeholder that was never wired).

Not "dummy data" per the CLAUDE.md prohibition (it's empty), but the feature does not exist.

**Fix:** join `stock_opname_details` → `stock_opname` → `products` (+category/supplier/warehouse), apply the 8 filters the handler already parses, count + paginate.

---

### BE-06 — HIGH — Excel export writes headers + zero rows
`internal/usecase/export_usecase.go`
```go
var sampleData []domain.StockOpnameDetail
// Populate with actual data when repository method is ready
```
`ExportStockOpnameToExcel` produces a file with column headers and no data. `ExportProductsToExcel` does populate. A downloadable, apparently-valid but empty spreadsheet is worse than a 501.

---

### BE-07 — HIGH — CSV export ignores 8 of 10 filters the UI sends
`internal/delivery/http/handler/report.go` → `ExportStockOpnameReport` reads only `warehouse_id` and `brand`. The Reports page sends `category_id`, `supplier_id`, `date_from`, `date_to`, `min_diff`, `max_diff` too (`dashboard/src/routes/reports/+page.svelte:46-59`). User filters, exports unfiltered file, no warning. `page`/`limit` are also forwarded to the export URL and ignored.

**Fix:** share one filter-parsing function between the JSON and export handlers.

---

### BE-08 — HIGH — `input_method` enum rejects two values the UI sends
`internal/delivery/http/handler/stock_opname.go:36`
```go
InputMethod string `json:"input_method" validate:"required,oneof=manual voice"`
```
Dashboard `/scan` and mobile offer `camera` / `upload` paths. Those requests fail validation → HTTP 400 `SUBMIT_SCAN_FAILED`. Any capture that isn't typed or spoken is rejected at the trust boundary.

**Fix:** decide the canonical set; if camera/upload are real, `oneof=manual voice camera upload` **and** carry through to `stock_opname_details.input_method` (which has its own CHECK constraint — verify it matches).

---

### BE-09 — MEDIUM — Dashboard summary is N+1 and semantics are off
`internal/usecase/dashboard_usecase.go`: loads ≤9999 products, ≤100 sessions, then **for each session** loads ≤9999 details. Worst case ~1M rows for one dashboard load.

- `todayProgress` counts every detail in `in_progress` sessions regardless of `scanned_at` date (comment says "simplified").
- `totalDifference` counts *mismatched items*, but the field name reads as a **sum** of differences. The UI labels it accordingly.
- `DashboardSummary` has **no `recent_activities` field**, yet `src/routes/+page.svelte` renders `summary?.recent_activities` — a permanently empty block.

**Fix:** 3 aggregate queries (`COUNT`, `SUM(difference)`, `GROUP BY`) + one activity-log query, all with `WHERE` on the requested window.

---

### BE-10 — MEDIUM — `Save()` used for session update
`internal/repository/stock_opname_repository.go:60-65` — `Save()` writes the full row from an in-memory struct read earlier. Two approvers racing → last write wins, first approver's `approved_by` silently lost. `stock` has a `version` column for optimistic locking; `stock_opname` does not.

**Fix:** targeted `Updates(map[string]any{...})` with `WHERE status = 'completed'`, check `RowsAffected`.

---

### BE-11 — MEDIUM — `List` reuses one `*gorm.DB` chain for Count then Find
`stock_opname_repository.go:71-96` and `product_repository.go:124-158`. GORM's `Count` on a chain mutates/clones it inconsistently when `Preload`/`Order`/`Limit` are appended afterwards. Currently works, but any future `Select()` added before `Count` will silently leak into the `Find`. Latent.

**Fix:** build filters into a helper applied twice, or `.Session(&gorm.Session{})`.

---

### BE-12 — MEDIUM — `rate_limit.go` never mounted
`internal/middleware/rate_limit.go` implements Redis `INCR`-per-IP, fail-open, no-op when redis is nil. `routes.go` never references it. `Login` and `Refresh` are therefore unthrottled — unlimited password attempts.

**Fix:** `auth.POST("/login", middleware.RateLimit(cfg.RedisClient, 10, time.Minute), authHandler.Login)`.

---

### BE-13 — MEDIUM — WebSocket endpoint has no authentication
`routes.go:197` `v1.GET("/ws/stock-opname/sessions/:id", websocket.HandleStockOpnameWS)` sits outside every `AuthMiddleware` group. Anyone who can reach the host subscribes to any session's progress stream by guessing/knowing a session UUID. `stock_opname_ws.go` `CheckOrigin` is a **hardcoded allowlist** including `http://10.250.235.79` — an environment-specific literal in source.

Also in that file: `hub.unregister` is closed twice (`defer` + a goroutine) → potential panic on double-close/send, and `BroadcastProgress` is a TODO stub emitting only `{"session_id":..., "type":"progress_update"}` with no payload.

**Fix:** token via query param or `Sec-WebSocket-Protocol`, validated with `ParseAccessToken`; `CheckOrigin` from config; single unregister path.

---

### BE-14 — MEDIUM — Import is non-transactional and matches by barcode
`internal/usecase/import_usecase.go` upserts products by **barcode** (overwriting existing rows), matches category by lowercased name (creates nothing if missing), defaults `unit` to `PCS`, and runs each row independently. A row-50 failure leaves rows 1-49 committed. No row cap beyond the 5 MB body limit (~60k rows). No audit log per row.

**Fix:** wrap in `db.Transaction`, cap rows (e.g. 2000), report `{inserted, updated, failed:[{row, reason}]}`.

---

### BE-15 — LOW — `excelize` marked `// indirect` though imported directly
`go.mod` lists `github.com/xuri/excelize/v2 v2.11.0 // indirect`; `export_usecase.go` and `import_usecase.go` import it. Cosmetic, but it means `go mod tidy` has not been run and dependency attribution is wrong.

---

### BE-16 — LOW — `Dockerfile` base image below `go.mod` directive
`golang:1.23-alpine` vs `go 1.25.0`. Relies on `GOTOOLCHAIN=auto` downloading a toolchain at build time — slow, and fails in air-gapped CI.

---

### BE-17 — LOW — `SubmitScan` stock lookup scoped to session warehouse
`FindByProductAndWarehouse(product.ID, session.WarehouseID)`. If the product is stocked in a different warehouse, `stock` is nil → `stockSystem = 0` → `difference = physical - 0`. Silent, wrong-but-plausible number. Should surface "product not stocked in this warehouse".

### BE-18 — LOW — `notification_usecase.CheckAndNotifyDifference` is never called
Defined with a proper threshold check, but `SubmitScan` never invokes it — the notification feature produces only whatever manually-inserted rows exist. Its `Metadata` JSON is also built with `fmt.Sprintf`, injecting `productName` verbatim → broken JSON if the name contains `"` or `\`. Should be `json.Marshal`.

---

## Backend summary
Builds. The skeleton is complete and idiomatic. **Three criticals** (code uniqueness, wrong-row return, unvalidated refresh token) plus **stock never written back** mean the core opname cycle is not closed. Reports and Excel export are stubs. Rate limiting is dead code.
