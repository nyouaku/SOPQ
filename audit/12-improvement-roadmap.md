# 12 — Improvement Roadmap

Sequenced by **risk reduced per unit of effort**, not by category. Effort scale: **XS** ≤2 h, **S** ≤1 day, **M** 2–3 days, **L** 4–5 days, **XL** 1–2 weeks. Assumes one full-stack developer. No item requires a stack change.

## Wave 0 — Audit hygiene (do first, 1 day)

Cheap work that makes every later fix verifiable.

| # | Item | Refs | Effort |
|---|---|---|---|
| 0.1 | `git init` at root, commit the tree, `.gitignore` (`node_modules`, `.env`, `build/`, `*.xlsx`) | REPO-01, CQ-11 | XS |
| 0.2 | Delete nested `stock-opname-app/stock-opname-app/` | REPO-01 | XS |
| 0.3 | Delete `performance_indexes.sql`, `seed_user.sql`, `seed_users.sql` from the mounted dir; remove the `docker-entrypoint-initdb.d` mount | DB-01…DB-05, SEC-08, CQ-08 | XS |
| 0.4 | Add `eslint.config.js`; run `prettier --write .` | FE-ARCH-01 | S |
| 0.5 | Bump `golang:1.25-alpine`; `go mod tidy`; pin `nginx:1.27-alpine` | BE-15, BE-16, PR-09 | XS |
| 0.6 | GitHub Actions: 3 jobs (`go build+vet+test`, `npm ci+check+build+lint`, `flutter analyze+test`) — **leave red for now, it is the truth** | T-05, T-06 | S |
| 0.7 | Write `docs/DEPLOYMENT.md` + env-var reference from the actual `config.go` | PR-13 | S |

**Exit:** repo under version control, CI exists, the two broken SQL files are gone.

## Wave 1 — Make the core cycle correct (the release blocker, ~2 weeks)

Nothing else matters until a scan can be recorded, completed, approved, and reflected in stock.

| # | Item | Refs | Effort |
|---|---|---|---|
| 1.1 | `StockOpname.code` — daily counter inside a transaction, or `UNIQUE(warehouse_id, code)` | BE-01 | S |
| 1.2 | `SubmitScan` returns the inserted row (`WHERE id = detail.ID`) and stops discarding the reload error | BE-02 | XS |
| 1.3 | Refresh token verified against `sessions.refresh_token_hash`; rotate in a txn; re-check `users.is_active` | BE-03, SEC-04 | M |
| 1.4 | `ApproveSession` writes `stock.quantity_system` from approved details in one transaction; use the existing `UpdateWithOptimisticLock` | BE-04, CQ-02 | M |
| 1.5 | Real migration runner (golang-migrate binary in the entrypoint), numbered files, `000001` frozen | PR-08, CQ-08 | M |
| 1.6 | Build `sessions/new/+page.svelte` (warehouse + location → `POST /sessions`) | FN-01 | S |
| 1.7 | `input_method` enum aligned end-to-end: handler `oneof`, DB `CHECK`, dashboard, mobile | BE-08, DB-07 | S |
| 1.8 | `flushQueue` — re-entrancy guard + attempt counter + dead-letter; enqueue only on transport failure | FN-02…FN-05 | M |
| 1.9 | Idempotency: client UUID per scan + `UNIQUE(stock_opname_id, idempotency_key)` | FN-04 | M |
| 1.10 | Mobile: session-selection screen, remove `'dummy-session-id'`, add logout, wire the 4 unused datasource methods | FN-10 | M |
| 1.11 | `apiBaseUrl` via `--dart-define`; `PrettyDioLogger` behind `kDebugMode` | MOB-ARCH-01, SEC-10 | XS |
| 1.12 | Go tests compile — update mocks to current interfaces, then add the 4 regression tests (1.1–1.4) | T-01, T-07 | M |
| 1.13 | Delete `mobile/test/widget_test.dart`; write 3 real widget tests | T-02 | S |

**Exit:** login → create session → scan (web + mobile) → complete → approve → `stock` updated. Regression tests for each step. **This is the gate for "works at all".**

## Wave 2 — Security and access control (~1 week)

| # | Item | Refs | Effort |
|---|---|---|---|
| 2.1 | WebSocket under `authMW`; token in `Sec-WebSocket-Protocol` or first frame; warehouse scope check; single unregister path; config-driven `CheckOrigin` | SEC-01, BE-13 | M |
| 2.2 | `RequireRole("supervisor","admin")` on `/export/*` | SEC-02 | XS |
| 2.3 | Mount `RateLimit` on `login` + `refresh`; add a per-username counter | SEC-06, BE-ARCH-02, BE-12 | S |
| 2.4 | CORS: default to the real dashboard origin, fail on mismatch, never `*` + credentials | SEC-07 | XS |
| 2.5 | Registration: min 12 chars, role validated against `roles`, admin-only route, unique → 409 | SEC-09 | S |
| 2.6 | Fix `Logout`'s `userID`→`sessionID` argument; check `RowsAffected` on both revoke methods; audit rows for auth events | SEC-13, CQ-06 | S |
| 2.7 | `VerifyToken` → `subtle.ConstantTimeCompare`; JWT gains `jti`/`iss`/`aud` | SEC-05, SEC-16 | XS |
| 2.8 | Menu role filtering in all shells | SEC-03, FE-03 | S |
| 2.9 | Security headers in nginx (CSP, `X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`, HSTS once TLS lands) | SEC-14 | XS |
| 2.10 | Fail-fast if `JWT_SECRET` is unset/default; drop `5432`/`6379` host port mappings; remove `changeme` default | SEC-15, PR-11 | XS |
| 2.11 | Import: row cap 2000, txn, per-row report, validate before first write | SEC-11, BE-14 | S |
| 2.12 | `json.Marshal` for `notifications.metadata`; call `CheckAndNotifyDifference` from `SubmitScan` | SEC-12, BE-18, DB-11 | XS |

**Exit:** no unauthenticated endpoint, no published credential, no unbounded auth surface.

## Wave 3 — The features that are advertised but absent (~2 weeks)

| # | Item | Refs | Effort |
|---|---|---|---|
| 3.1 | Real report query — joins, all 8 filters, count + pagination, index `(stock_opname_id, scanned_at DESC)` | BE-05, DB-09, PERF-05 | L |
| 3.2 | One `parseReportFilter(c)` shared by JSON and CSV export | BE-07, CQ-05 | S |
| 3.3 | CSV export streamed with a documented row cap | PERF-03 | S |
| 3.4 | Excel export populates real rows via `excelize.StreamWriter` | BE-06, PERF-04 | M |
| 3.5 | Dashboard aggregates — 3 SQL queries replace the N+1; add `recent_activities` to the response | PERF-01, BE-09 | M |
| 3.6 | WS broadcast — `SubmitScan` emits `new_scan`; `BroadcastProgress` sends a real payload | FN-06 | M |
| 3.7 | Report tiles read server aggregates, not the current page | FE-05 | XS |
| 3.8 | Product search: trigram on `name`, prefix on `barcode`/`product_code`; preload `Box.Location.Warehouse` | PERF-02, FN-07 | M |
| 3.9 | Delete `/mobile`; move its `pendingScans` display into `/scan` | CQ-04, FN-12, FE-10 | S |
| 3.10 | Replace 4 `alert()` with `toastStore`; `$applied` reactivity; URL-synced filter state | FE-02, FE-06, FN-09 | S |

**Exit:** every menu item returns real data; no advertised feature is a stub.

## Wave 4 — Production readiness (~2 weeks)

| # | Item | Refs | Effort |
|---|---|---|---|
| 4.1 | Graceful shutdown — `http.Server` + `Shutdown(30s)` + hub close; `stop_grace_period` | PR-01, BE-ARCH-01 | S |
| 4.2 | Dashboard served same-origin — delete `VITE_*` build args, relative `/api/v1` | PR-05, BE-ARCH-03 | M |
| 4.3 | One config loader with `env` tags, fail-fast, `.env.example` matches exactly | CQ-01, BE-ARCH-04, PR-12 | S |
| 4.4 | `/metrics` (Prometheus) + OTel traces + `X-Request-ID` in logs and responses | PR-02 | M |
| 4.5 | Nightly `pg_dump -Fc`, 14-day retention, written restore drill, one executed restore into a scratch DB | PR-03 | M |
| 4.6 | TLS: `listen 443 ssl`, certbot, HTTP→HTTPS redirect, HSTS | PR-04 | S |
| 4.7 | Healthchecks on backend + dashboard; nginx gated on `service_healthy`; redis healthcheck | PR-06 | XS |
| 4.8 | Resource limits, log rotation, drop `container_name`, remove backend `ports:` | PR-07, PR-10 | S |
| 4.9 | Dashboard test suite: vitest + `@testing-library/svelte` — `queries.ts`, `queue.ts`, `apiClient` refresh | T-03 | M |
| 4.10 | One Playwright E2E: login → session → scan → complete → approve → report | T-04 | M |
| 4.11 | `govulncheck` + `npm audit` in CI; `gosec`; `staticcheck` | T-06 | S |
| 4.12 | Runbook: start/stop, logs, common failures, on-call contacts | PR-13 | S |

**Exit:** deployable, observable, restorable, testable, from a clean host.

## Wave 5 — Polish (schedule, not blocking)

`FE-04` (unify dashboard download/upload through `apiClient`), `FE-07` (`checkAuth` verification), `FE-08`/`CQ-02` (delete the remaining dead code, type the API rows instead of `any[]`), `FN-13` (surface errors instead of empty states), `PERF-07` (Chart.js tree-shake), `PERF-08` (SW shell-only precache), `PERF-09` (polling `staleTime`), `PERF-11` (mobile batch endpoint), `SEC-14` (remaining CSP tuning), `DB-08` (`updated_at` trigger), `DB-10`, `DB-12` (delete the stale schema doc), `T-09` (acceptance criteria per feature).

## Critical path

```
0.1 → 1.5 → 1.1 → 1.4 → 1.12 → 4.1
```
Version control, then a real migration path, then session creation, then stock write-back, then tests that prove it, then shutdown that does not lose the result.

## Total estimate

| Wave | Effort |
|---|---|
| 0 | 1 day |
| 1 | ~10 working days |
| 2 | ~5 working days |
| 3 | ~10 working days |
| 4 | ~10 working days |
| 5 | ~5 working days |
| **Total** | **≈ 41 working days (8–9 weeks, one developer)** |

Waves 0–1 (~11 days) are the minimum to reach **"the product works"**. Waves 2–4 (~25 days) are the minimum to reach **"the product can be operated"**.
