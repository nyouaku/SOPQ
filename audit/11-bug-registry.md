# 11 — Bug Registry

Single index of every finding across the audit. Detail, code excerpts, and reasoning live in `01`–`10`. Here: triage data only. **Status** is `OPEN` unless noted.

Legend — Severity: **C**ritical / **H**igh / **M**edium / **L**ow. Verification: **T** = automated test to be written, **G** = grep/static check, **M** = manual on running stack, **X** = external tool.

## A. Totals

| Severity | Count | Blocking release? |
|---|---|---|
| CRITICAL | 17 | **YES — all** |
| HIGH | 41 | **YES** |
| MEDIUM | 42 | No, but must be scheduled |
| LOW | 24 | No |
| **Total** | **124** | |

Fixed during the audit: **1** (FE-01).

## B. CRITICAL

| ID | Location | Defect | Fix | Verify | Status |
|---|---|---|---|---|---|
| BE-01 | `usecase/stock_opname_usecase.go:68` | `code` hardcoded `SO-<date>-0001`; UNIQUE violation on 2nd session/day | per-day counter or `(warehouse_id, code)` uniqueness | T: create 2 sessions same day | OPEN |
| BE-02 | `usecase/stock_opname_usecase.go:171` | `SubmitScan` returns `details[0]` (newest in session), not the inserted row; discards reload error | re-select by `detail.ID` or preload in create path | T: 2 concurrent scans, assert returned ID matches | OPEN |
| BE-03 | `usecase/auth_usecase.go:202` | refresh token never checked against DB; revoked tokens still mint credentials | `FindActiveByRefreshTokenHash` + rotate in txn | T: revoke → refresh must 401 | OPEN |
| DB-04 | `migrations/seed_user.sql` | invalid SQL: no `email` column, fake hash, `role` name `staff` vs `staff_user`, `password` vs `password_hash` | delete file | M: fresh `docker compose up` | OPEN |
| FE-01 | `api/queries.ts` | `createQuery` options read once → filters never refetch | `derived(paramsStore(...))` | T: change filter → new queryKey | **FIXED** |
| FN-01 | `routes/sessions/+page.svelte:89` | "New Session" → `/sessions/new` has no route → resolves to `[id]` → 400 | build the page or remove button | M: click it | OPEN |
| SEC-01 | `route/routes.go:197` | WS endpoint outside all auth groups — anonymous live opname feed | mount under `authMW`, validate token, check warehouse scope | M: connect without token | OPEN |
| SEC-04 | `usecase/auth_usecase.go:202` | same as BE-03 (security framing) | see BE-03 | see BE-03 | OPEN |
| PERF-01 | `usecase/dashboard_usecase.go:41` | N+1: ≤100 sessions × ≤9999 details per dashboard load; rows discarded after `len()` | 3 SQL aggregates | T: assert query count == 3 | OPEN |
| T-01 | `usecase/*_test.go` | Go test package does not compile; suite never ran | update mocks/interfaces to current signatures | T: `go test ./...` green | OPEN |
| T-02 | `mobile/test/widget_test.dart` | template test references `MyApp`, which does not exist | delete + 3 real widget tests | T: `flutter test` | OPEN |
| T-03 | `dashboard/` | zero tests, no test runner installed | vitest + `@testing-library/svelte` | T: 3 suites | OPEN |
| PR-01 | `cmd/api/main.go:75` | no `http.Server`/`Shutdown` — SIGTERM drops in-flight scans | `srv.Shutdown(ctx)` + hub close + `stop_grace_period` | M: SIGTERM under load | OPEN |
| PR-02 | entire backend | no `/metrics`, no tracing, no `X-Request-ID` | Prometheus middleware + OTel + request ID | M: scrape `/metrics` | OPEN |
| PR-03 | `docker-compose.yml:74` | no backup, no restore drill, no WAL archive | nightly `pg_dump -Fc` + restore doc | M: restore into scratch DB | OPEN |
| BE-ARCH-01 | `cmd/api/main.go:75` | same as PR-01 (architecture framing) | see PR-01 | see PR-01 | OPEN |
| SEC-08 | `migrations/seed_users.sql:9` | three accounts share one committed bcrypt hash, comment states the plaintext | remove from mounted dir; `make create-admin` | G: no hash literals in repo | OPEN |

> BE-ARCH-01/PR-01, BE-03/SEC-04 are the same defect seen from two axes — counted twice in the table above so each report reconciles, but **16 distinct criticals**.

## C. HIGH

| ID | Location | Defect | Fix | Verify |
|---|---|---|---|---|
| BE-ARCH-02 | `middleware/rate_limit.go` | implemented, never mounted | mount on `login`/`refresh` | G: `routes.go` references it |
| BE-04 | `usecase/stock_opname_usecase.go:180,211` | `stock` never written back after approve — reconciliation never closes | write stock in approve txn | T: approve → assert `stock.quantity_system` |
| BE-05 | `usecase/report_usecase.go` | report query is a stub returning zero rows always | implement join + 8 filters + pagination | T: seeded data → non-zero rows |
| BE-06 | `usecase/export_usecase.go` | Excel export = headers, zero rows | populate from detail repo | T: parse xlsx row count |
| BE-07 | `handler/report.go` | CSV export reads 2 of 10 filters; forwards `page`/`limit` meaningless | shared `parseReportFilter` | T: export with 8 filters |
| BE-08 | `handler/stock_opname.go:36` | `oneof=manual voice` rejects `camera`/`upload` the UI sends | align enum + DB CHECK | T: POST camera → 201 |
| DB-01 | `migrations/performance_indexes.sql:5` | `deleted_at` column does not exist on `stock_opname`/`sessions` → file aborts | delete file, add numbered migration | M: run file |
| DB-02 | `migrations/performance_indexes.sql:18` | `gin_trgm_ops` used before `CREATE EXTENSION` | reorder / separate file | M: run file |
| DB-03 | `migrations/performance_indexes.sql:43` | `VACUUM` in txn; index on nonexistent `stock_opname_detail.created_at` | remove | M: run under golang-migrate |
| DB-05 | `migrations/seed_users.sql:9` | shared password hash, hardcoded `role_id` 1/2/3 | delete | G |
| FE-02 | `routes/reports/+page.svelte:44` | `$:` reads `get(applied)` on a `const` → runs once, never true | `$applied` | T: apply filter → badge shows |
| FE-03 | `components/layouts/menu.ts` | no `roles` field; staff sees supervisor/admin pages then 403s | `roles?: string[]` + filter in shells | M: login as staff |
| FE-04 | `routes/products`, `routes/reports` | 3 sites use raw `fetch`, bypassing refresh-on-401 | `apiClient.download/upload` | T: expired token → export succeeds after refresh |
| FN-02 | `routes/scan/+page.svelte:345` | any thrown error, and `/mobile` any `!success`, enqueues permanently rejected scans | enqueue only on transport error | T: 422 → not queued |
| FN-03 | `lib/offline/queue.ts:105` | `flushQueue` re-queues 4xx forever; no attempts/dead-letter | attempts counter + dead-letter | T: 422 item drops after N |
| FN-04 | `lib/offline/queue.ts:10` | comment claims idempotency key; none exists; backend has none → duplicate detail rows | client UUID + `UNIQUE(stock_opname_id, idempotency_key)` | T: replay → 1 row |
| FN-05 | `lib/offline/queue.ts:85` | no re-entrancy guard; concurrent flush duplicates posts | `_isSyncing` flag | T: call twice concurrently |
| FN-06 | `routes/sessions/[id]/+page.svelte:61` | WS never delivers scans; `BroadcastProgress` is a stub; token in query string | emit `new_scan` from `SubmitScan`; move token | M: two browsers, one scans |
| SEC-02 | `route/routes.go` | `/export/*` any-authenticated — staff exfiltrates full catalogue | `RequireRole("supervisor","admin")` | T: staff → 403 |
| SEC-03 | `menu.ts` | same as FE-03 (security framing) | see FE-03 | M |
| SEC-05 | `pkg/utils/password.go:36` | `VerifyToken` uses `==` — non-constant-time | `subtle.ConstantTimeCompare` | G |
| SEC-06 | `middleware/rate_limit.go` | same as BE-ARCH-02 | mount | X: 11 logins in 1 min → 429 |
| SEC-07 | `middleware/cors.go:35` | `Allow-Credentials: true` always; `*` default reflects any origin | default to real origin; branch correctly | M: curl with foreign Origin |
| SEC-09 | `handler/auth.go:26` | 6-char password; client-supplied `role_id` clamped 1..3 | min 12 + role validated server-side | T: role_id=1 as staff → 403 |
| SEC-10 | `mobile/lib/core/network/dio_client.dart:24` | `PrettyDioLogger(requestBody,responseBody)` in release — logs credentials | `if (kDebugMode)` | G |
| MOB-ARCH-01 | `mobile/lib/core/constants/app_constants.dart:4` | hardcoded `http://localhost:8080`; release build reaches no server | `--dart-define` per flavor | M: release APK on device |
| CQ-01 | `config/config.go` | reads `APP_PORT`/`JWT_*_TTL_*`/`APP_ENV`; `.env` defines `SERVER_PORT`/`JWT_*_EXPIRY`/`ENV` — all silently defaulted | one loader, fail-fast | T: set `SERVER_PORT=9000` → binds 9000 |
| CQ-02 | 8 sites | dead code that reads as live: `RateLimit`, `UpdateWithOptimisticLock`, `CreateOrUpdate`, `CheckAndNotifyDifference`, `FindActiveByRefreshTokenHash`, `markNotificationRead`, `reportUsecase.db`, 4 mobile datasource methods | wire or delete | G: no unused exported symbol |
| CQ-03 | `usecase/*_test.go` | same as T-01 | see T-01 | T |
| PERF-02 | `repository/product_repository.go:130` | `ILIKE '%q%'` × 3 columns → full scan; unique btree indexes unusable | trigram on `name`, prefix on `barcode`/`product_code` | X: `EXPLAIN ANALYZE` |
| PERF-03 | `handler/report.go` | export ignores `page`/`limit`, no row cap; whole set in memory | stream + cap | T: export on 100k rows |
| PERF-04 | `usecase/export_usecase.go` | whole workbook built in memory | `excelize.StreamWriter` | X: memory profile |
| T-04 | — | no E2E for the core cycle | 1 Playwright happy path | T |
| T-05 | repo root | no CI at all | 3 GitHub Actions jobs | T: PR blocks on red |
| T-06 | repo root | `go vet` red; no ESLint config; no staticcheck/gosec | add configs + jobs | T |
| PR-04 | `docker/nginx.conf:9` | `listen 80` only — no TLS, no HSTS | 443 + certbot + redirect | M |
| PR-05 | `docker-compose.yml:49` | `VITE_API_BASE_URL: http://localhost:8080` baked into bundle → dashboard unreachable from any other device | relative `/api/v1`, delete build args | M: open from LAN device |
| PR-06 | `docker-compose.yml:37` | no healthcheck on backend/dashboard; redis gated on `service_started` | add healthchecks + `service_healthy` | M |
| PR-07 | `docker-compose.yml` | `container_name:` on every service blocks scaling | remove; use internal network | M: `--scale backend=2` |
| PR-08 | `infrastructure/postgres/postgres.go:70` | `RunMigrations()` is a no-op that logs "run psql manually" | golang-migrate at entrypoint | T: fresh DB gets schema |

## D. MEDIUM

| ID | Location | Defect | Fix |
|---|---|---|---|
| BE-ARCH-03 | `lib/api/config.ts` | baked `localhost` fallback (same root as PR-05) | relative path |
| BE-ARCH-04 | `config/config.go` | key mismatch (same as CQ-01) | see CQ-01 |
| FE-ARCH-01 | `dashboard/` | no ESLint config; `lint` script broken | add `eslint.config.js` |
| BE-09 | `usecase/dashboard_usecase.go` | `todayProgress` ignores date; `totalDifference` counts items not sums; `recent_activities` field absent | aggregates + field |
| BE-10 | `repository/stock_opname_repository.go:60` | `Save()` full-row overwrite, lost update on concurrent approve | targeted `Updates` + `RowsAffected` |
| BE-11 | `repository/*_repository.go` | one GORM chain mutated for `Count` then `Find` | helper applied twice |
| BE-12 | `middleware/rate_limit.go` | (same as BE-ARCH-02) | mount |
| BE-13 | `websocket/stock_opname_ws.go` | hardcoded `CheckOrigin` internal IP; double `unregister`; stub broadcast | config origin; single path |
| BE-14 | `usecase/import_usecase.go` | non-transactional, no row cap, upserts by barcode | txn + cap + report |
| DB-06 | `migrations/*` | `sessions` has no soft delete though SQL assumes it | align |
| DB-07 | `000001:157` | `input_method` no CHECK constraint | `CHECK (input_method IN (...))` |
| DB-08 | `migrations/*` | no `updated_at` trigger | `set_updated_at()` trigger |
| DB-09 | `000001` | index on `(stock_opname_id, scanned_at DESC)` and composite | numbered migration |
| FE-05 | `routes/reports/+page.svelte:39` | summary tiles sum current page while `total` is full set | server aggregates |
| FE-06 | `routes/sessions/[id]/+page.svelte:120,123,137,140` | 4 blocking `alert()` | `toastStore` |
| FE-07 | `lib/stores/auth.ts` | `checkAuth()` trusts localStorage without token presence/expiry | require token, optional `/auth/me` |
| FE-08 | `lib/api/queries.ts:160` | dead export, `pendingCount` unrendered, 24 `console.*`, `any[]` everywhere | delete / type |
| FN-07 | `repository/product_repository.go:149` | `List` does not preload `Box.Location.Warehouse` → "Gudang" column always `-` | add preload |
| FN-08 | `menu.ts` | privileged pages reachable by staff (same as FE-03, UX framing) | see FE-03 |
| FN-09 | `routes/sessions/[id]` | filter/pagination state not in URL; back/forward loses position | URL-synced state |
| FN-10 | `mobile/.../scanner_screen.dart:26` | `setSessionId('dummy-session-id')` — every scan posts to an invalid UUID; no session picker, no logout, 4 datasource methods unused | session selection screen |
| SEC-11 | `usecase/import_usecase.go` | no row cap, no zip-bomb guard, no schema validation before write | cap + validate + txn |
| SEC-13 | `handler/auth.go:168` | `Logout` passes `userID` into `sessionID` → `RevokeByID` no-ops silently; no audit rows for auth events | fix arg, check `RowsAffected`, audit |
| SEC-14 | `docker/nginx.conf` | no CSP / `X-Content-Type-Options` / `Referrer-Policy` / `X-Frame-Options`; second `Allow-Origin *` | `add_header` block |
| SEC-15 | `.env`, `docker-compose.yml` | secrets hygiene: no fail-fast on `JWT_SECRET`; compose default `changeme`; pg/redis ports published | require secret; drop port maps |
| CQ-04 | `routes/mobile` vs `routes/scan` | two implementations of one workflow, divergent failure policy | delete `/mobile` |
| CQ-05 | `handler/report.go` | two filter parsers, one incomplete | share |
| CQ-06 | `repository/session_repository.go:65` | `RevokeByID`/`RevokeByUserID` ignore `RowsAffected` | check |
| CQ-07 | `handler/export.go:36,62` | raw `err.Error()` returned to client; leaked driver messages | typed sentinels + mapping |
| CQ-08 | `docker-compose.yml:16` | `initdb.d` mount → non-deterministic schema | golang-migrate |
| PERF-05 | `migrations/*` | missing composites; perf file duplicates existing indexes | replace file |
| PERF-06 | `repository/stock_opname_repository.go:60` | `Save()` writes every column | see BE-10 |
| PERF-07 | `routes/+page.svelte:9` | `Chart.register(...registerables)` on landing page (~200 KB) | register used components only |
| PERF-08 | `service-worker.ts:14` | `cache.addAll([...build, ...files])` on install | shell only |
| PERF-09 | `lib/api/queries.ts:128` | 60 s poll, no `staleTime`, polls in background tabs | `staleTime` + `refetchIntervalInBackground: false` |
| T-07 | `usecase/*_test.go` | existing tests target arithmetic, not control flow | 4 regression tests |
| T-08 | — | no fixtures / test DB | testcontainers |
| PR-09 | `Dockerfile`s | `golang:1.23` < `go 1.25`; `nginx:stable-alpine` unpinned | bump / pin |
| PR-10 | `docker-compose.yml` | no resource limits, no log rotation, no `stop_grace_period` | add blocks |
| PR-11 | `docker-compose.yml:9` | `changeme` default password committed | require `.env` |
| PR-12 | `.env.example` | 4 keys documented but not read (same as CQ-01) | one source |
| PR-13 | `docs/` | no deployment/runbook/env reference | write |

## E. LOW

| ID | Location | Defect |
|---|---|---|
| REPO-01 | root | nested `stock-opname-app/` duplicate tree with own `.git`; no root VCS |
| BE-15 | `go.mod` | `excelize` marked `// indirect` though directly imported |
| BE-16 | `backend/Dockerfile:2` | `golang:1.23-alpine` vs `go 1.25.0` |
| BE-17 | `usecase/stock_opname_usecase.go` | stock lookup scoped to session warehouse → silent `stockSystem = 0` |
| BE-18 | `usecase/notification_usecase.go` | `CheckAndNotifyDifference` never called; `Metadata` built via `Sprintf` |
| FE-09 | `routes/activity/+page.svelte:135` | retry uses last-applied params, not current input |
| FE-10 | `routes/mobile/+page.svelte` | `pendingScans` never rendered — stuck queue invisible |
| FE-11 | `lib/api/queries.ts` | notifications polled per page with no dedupe |
| DB-10 | `000001` | `stock` has no FK constraining product↔warehouse scope |
| DB-11 | `usecase/notification_usecase.go:64` | JSONB written via `fmt.Sprintf` → invalid JSON lost |
| DB-12 | `docs/DATABASE_SCHEMA.sql` | two divergent schema documents |
| FN-12 | `routes/mobile/+page.svelte:104` | `captureFrame()` is a TODO no-op; `/mobile` orphaned, not in menu |
| FN-13 | `routes/**` | pages swallow errors into `console.error`; 500 looks like empty state |
| SEC-12 | `usecase/notification_usecase.go` | JSON injection shape (same as DB-11) |
| SEC-16 | `pkg/utils/jwt.go` | no `jti`, no `iss`/`aud` |
| CQ-09 | `go.mod` | `go mod tidy` never run |
| CQ-10 | `backend/Dockerfile` | toolchain below `go.mod` (same as BE-16) |
| CQ-11 | root | nested duplicate tree (same as REPO-01) |
| CQ-12 | — | naming drift, noted so it is not "fixed" blindly |
| PERF-10 | `mobile/.../dio_client.dart` | release-build logger overhead |
| PERF-11 | `mobile/.../sync_service.dart` | one-by-one flush, no batch endpoint |
| PERF-12 | `handler/*` | no ETag/Last-Modified on GETs |
| T-09 | `docs/` | no test plan / acceptance criteria |
| PR-14 | `handler/health.go` | `/health` discloses internal service topology unauthenticated |

## F. Fix-before-ship shortlist (the 16 distinct criticals)

1. BE-01 code uniqueness
2. BE-02 wrong-row return
3. BE-03/SEC-04 refresh verification
4. DB-04 + SEC-08 seed files removed
5. FN-01 new-session route
6. SEC-01 WS auth
7. PERF-01 dashboard aggregates
8. PR-01 graceful shutdown
9. PR-02 observability
10. PR-03 backups
11. T-01 Go tests compile
12. T-02 mobile test
13. T-03 dashboard test runner
14. PR-05 dashboard API URL
15. PR-08 real migration runner
16. CQ-02/CQ-03 dead code on the auth path

Everything else is HIGH or below and can be sequenced behind these.
