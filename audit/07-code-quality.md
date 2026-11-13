# 07 — Code Quality

## A. Size inventory

| Stack | Files | LOC |
|---|---|---|
| Backend Go | 51 `.go` | 6 904 |
| Dashboard | 48 files (`src/`) | 7 539 |
| Mobile | 19 `.dart` | 2 336 |
| SQL migrations | 9 files | — |

Total hand-written ≈ 16 800 LOC. Small enough that the debt below is affordable to clear; large enough that the placeholder count is a problem.

## B. Signals

| Signal | Count | Verdict |
|---|---|---|
| `TODO`/`FIXME` in `backend/**/*.go` | 5 | **All 5 sit on a critical or high path** (BE-01 code gen, BE-03 refresh verify, BE-05 report query, BE-06 export data, BE-13 WS broadcast). |
| `TODO`/`FIXME` in `dashboard/src` | 1 | `captureFrame()` no-op (`/mobile`). |
| `TODO`/`FIXME` in `mobile/lib` | 0 | — |
| `console.*` in `dashboard/src` | 24 | Ships to the browser in production. |
| `any` in `.ts`/`.svelte` | measured: see FE-08 | TS strict is on but the API surface is `any[]` at every consumption site — strict mode buys almost nothing. |
| `alert()` | 4 | `sessions/[id]/+page.svelte:120,123,137,140` |
| Duplicated route implementations | 1 | `/scan` vs `/mobile` — 490 vs ~370 lines, same workflow, divergent bugs. |

## C. Findings

### CQ-01 — HIGH — Configuration keys do not match the env files
`.env.example` declares `SERVER_PORT`, `JWT_ACCESS_EXPIRY`, `JWT_REFRESH_EXPIRY`, `ENV`, `CORS_ALLOWED_ORIGINS`.
`internal/config/config.go` reads `APP_PORT`, `JWT_ACCESS_TTL_MINUTES`, `JWT_REFRESH_TTL_DAYS`, `APP_ENV`.

Every one of those falls through to its compiled default. The operator sets `SERVER_PORT=9000`, gets 8080, and has no error to explain it. Silent misconfiguration is the worst class of quality defect: the system appears to honour configuration and does not.

**Fix:** one struct of `env` tags read with a single loader; fail-fast on unknown/required-missing.

---

### CQ-02 — HIGH — Dead code that looks alive
| Symbol | Location | Why it matters |
|---|---|---|
| `middleware.RateLimit` | `rate_limit.go` | Fully implemented, never mounted (SEC-06). Reviewer reads "we have rate limiting". |
| `stockRepo.UpdateWithOptimisticLock`, `stockRepo.CreateOrUpdate` | `stock_repository.go` | Implemented, never called (BE-04). Reads as a working reconciliation loop. |
| `notificationUsecase.CheckAndNotifyDifference` | `notification_usecase.go` | Implemented, never called (BE-18). |
| `markNotificationRead` | `queries.ts:160` | Exported mutation factory, zero importers (FE-08). |
| `sessionRepository.FindActiveByRefreshTokenHash` | `session_repository.go:51` | Implemented correctly, never called by production code (only the failing test references it). |
| `reportUsecase.db interface{}` | `report_usecase.go` | Untyped placeholder field, never assigned. |
| `Mobile` route `pendingCount` | `routes/mobile/+page.svelte` | Computed, never rendered. |
| 4 `StockOpnameRemoteDataSource` methods | `mobile/.../datasources` | Implemented, never called from any screen (FN-10). |

These are worse than absent code: each is evidence a feature was built and then never wired, and each misleads the next reader. Six of the eight are on security or data-integrity paths.

**Fix:** wire or delete. A comment `// unused — see BE-04` is not a substitute.

---

### CQ-03 — HIGH — Test suite does not compile; mobile test references a class that does not exist
```go
// go vet ./...  →
internal\usecase\stock_opname_usecase_test.go:144:73: undefined: domain.ProductFilter

// go test ./...  (does not compile)
auth_usecase_test.go: mock missing List(ctx, string, int, int)
                      missing FindActiveByRefreshTokenHash
                      too many arguments in call to authUsecase.Register
                      unknown field AccessExpiry / RefreshExpiry
stock_opname_usecase_test.go: undefined domain.ProductFilter, domain.ActivityLogFilter
```
```dart
// mobile/test/widget_test.dart:14
await tester.pumpWidget(const MyApp());   // main.dart exports StockOpnameApp
```
`MyApp` does not exist. The test file is the unmodified Flutter template (`Counter increments smoke test`), asserting `find.text('0')` against an app with no counter. `flutter test` cannot pass.

The Go tests also encode the **intended** design: they reference `FindActiveByRefreshTokenHash`, `AccessExpiry`, `RefreshExpiry` — the methods/fields the implementation omitted (BE-03, BE-ARCH-04). The tests were written against the correct design; the code was never finished to match.

**Fix:** this is the single highest-leverage quality action available — the tests already describe the target. Finish the implementation until they compile and pass.

---

### CQ-04 — MEDIUM — Two frontend implementations of one workflow
`/scan` (370 lines) and `/mobile` (490 lines) both: pick a session, drive a camera, fall back to manual entry, enqueue on failure, flush on reconnect. They disagree on the failure policy (FN-02: `/scan` enqueues on throw, `/mobile` enqueues on `!success`), and `/mobile` alone has the `captureFrame()` no-op TODO.

Neither is more correct. Both are reachable; only `/scan` is in the menu.

**Fix:** delete `/mobile` after porting its `pendingScans` display idea into `/scan` (FE-10). Deletion over addition.

---

### CQ-05 — MEDIUM — Duplicated filter parsing between JSON and export handlers
`handler/report.go` — `GetStockOpnameReport` parses 8 filters; `ExportStockOpnameReport` parses 2. Same URL contract, two parsers, one of them incomplete (BE-07). This is the classic "copy the handler, forget half the fields" defect and it cannot be caught by a compiler.

**Fix:** one `parseReportFilter(c) (repository.ReportFilter, error)` used by both.

---

### CQ-06 — MEDIUM — `Save()` and `Updates()` mixed without a policy
`stock_opname_repository.go:60` uses `Save()` (full-row overwrite, BE-10); `session_repository.go:70` uses `Update("revoked_at", now)` (targeted). `RevokeByID`/`RevokeByUserID` do not check `RowsAffected` — a revoke of a nonexistent session reports success (SEC-13). Two idioms, one of which is wrong under concurrency and one of which is silently no-op.

**Fix:** targeted `Updates` + `RowsAffected` check everywhere; `Save()` nowhere.

---

### CQ-07 — MEDIUM — Error handling is inconsistent across the request path
- `utils.RespondError(c, 500, "EXPORT_FAILED", err.Error())` — raw internal error string returned to the client in `export.go:36,62` and `import.go`. Leaks filesystem paths and SQL errors.
- `RespondBadRequest(c, "START_SESSION_FAILED", err.Error())` — a DB unique violation (BE-01) is reported as a 400 with the raw driver message.
- `_, _` discards an error in `stock_opname_usecase.go:171` (BE-02).
- `if err == gorm.ErrRecordNotFound` instead of `errors.Is` in `session_repository.go:43,57` — works today because GORM returns the sentinel directly, breaks the moment it is wrapped.
- `fmt.Errorf("...: %w", err)` convention is followed in repositories — good — but `%w` wrapping is discarded at the handler boundary.

**Fix:** typed sentinel errors (`domain.ErrNotFound`, `ErrConflict`) mapped to status codes in one place; never return `err.Error()` to the client; `errors.Is` throughout.

---

### CQ-08 — MEDIUM — `docker-entrypoint-initdb.d` mount makes the schema non-deterministic
`docker-compose.yml` mounts `./backend/migrations` into the init directory. Consequences: files run alphabetically and unreliably (two of them are broken SQL — DB-01…DB-04); editing any file has zero effect on an existing volume; there is no version record. The DB state of two developers' machines can differ while both claim "docker compose up works".

**Fix:** see 04-database-audit.md migration strategy.

---

### CQ-09 — LOW — `go mod tidy` never run
`excelize` is `// indirect` while directly imported (BE-15). Suggests `go mod tidy` is not part of any workflow.

### CQ-10 — LOW — `Dockerfile` toolchain below `go.mod`
`golang:1.23-alpine` vs `go 1.25.0` — relies on `GOTOOLCHAIN=auto` downloading at build time (BE-16).

### CQ-11 — LOW — Nested duplicate repository tree
`stock-opname-app/stock-opname-app/` with its own `.git`, 18 files. Doubles every search, produces false positives in review, and is where the root repo's absence of version control becomes visible (REPO-01).

### CQ-12 — LOW — Naming drift
`CheckAndNotifyDifference` vs `CheckAndNotifyDifference` call sites, `stock_opname` (table) vs `StockOpname` (struct) vs `sessions` (table) vs `Session` (struct) — acceptable Go/GORM convention. Real inconsistency: the table is `stock_opname`, the detail table `stock_opname_detail`, the plural route `/stock-opname/sessions`. Not a defect, noted so the next auditor does not "fix" it.

---

## D. Positive quality signals
- Clean-architecture boundary is **actually respected** — verified `domain/*.go` imports only `time`, `uuid`, gorm tags. Most projects of this size violate it.
- Interfaces defined at the consumer (`usecase`) with concrete implementations in `repository` — correct Go idiom.
- `errors` wrapped with `%w` in the repository layer — correct.
- TypeScript `strict: true`; `svelte-check` passes with 0 errors.
- `excelize`, `gin`, `gorm`, `jwt/v5`, `go-redis/v9` are all current majors; no abandoned dependency found.
- DaisyUI tokens drive the chart theme via `MutationObserver` — thoughtful, theme-aware detail (`+page.svelte:36`).
- Response envelope is uniform (`utils.Respond*`) — one shape for the whole API.
- The Go tests, though non-compiling, are **genuinely good tests** that document intent.

## E. Assessment
The code reads as written by someone competent working under time pressure: structure correct, idioms right, and then a large number of features stopped at 80 % with a `TODO` marking the gap. The debt is not "messy" — it is **unfinished**. That is the more fixable kind, and CQ-03 is the proof: the tests already encode the target design.
