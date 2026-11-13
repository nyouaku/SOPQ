# 00 — Executive Summary

Audit of **Stock Opname Barcode + Voice** (`D:\aldeb\project\project_21\stock-opname-app`).
Backend Go/Gin/GORM · Dashboard SvelteKit · Mobile Flutter · PostgreSQL 16 · Redis · Docker Compose · nginx.

Audit date: 2026-11-12. Phases 01–09 executed. **No source file was modified, no file deleted, no migration run, no production config touched, no external service contacted.** One defect (FE-01) had already been fixed before this audit began and is recorded as such.

## 1. Overall score

# 44 / 100

**Status: NOT APPROVED FOR PRODUCTION.**

| Category | Weight | Score | Weighted |
|---|---|---|---|
| Backend Architecture | 12 % | 52 | 6.24 |
| Frontend | 10 % | 61 | 6.10 |
| Functionality | 15 % | 38 | 5.70 |
| Database | 10 % | 45 | 4.50 |
| Security | 15 % | 30 | 4.50 |
| Code Quality | 10 % | 55 | 5.50 |
| UI/UX | 8 % | 78 | 6.24 |
| Performance | 8 % | 45 | 3.60 |
| Testing | 7 % | 12 | 0.84 |
| Production Readiness | 5 % | 20 | 1.00 |
| **Total** | **100 %** | | **44.22** |

Read the score as: *a well-shaped codebase of half-finished features, with a solid frontend and almost no verification.*

## 2. Bug counts

| Severity | Count | Blocking |
|---|---|---|
| CRITICAL | 16 distinct (17 rows — 2 pairs are the same defect from two axes) | **yes** |
| HIGH | 41 | **yes** |
| MEDIUM | 42 | scheduled |
| LOW | 24 | optional |
| **Total** | **124** | |
| Fixed during audit | 1 (FE-01) | |

## 3. The five findings that decide the verdict

**1. The opname cycle does not close.** `SubmitScan` returns the newest detail row in the session rather than the one just inserted (`BE-02`), so a scanner can be told it recorded someone else's product. `ApproveSession` only flips a status — `stock.quantity_system` is never updated (`BE-04`), so every subsequent opname measures against the same stale baseline. This is the product's core function.

**2. Sessions cannot be created a second time in a day.** `code := fmt.Sprintf("SO-%s-0001", ...)` is a placeholder against a `UNIQUE` column (`BE-01`). One session per warehouse per day, ever.

**3. A revoked token still authenticates.** `RefreshToken` skips the database lookup entirely — the `TODO` says so (`BE-03`). Logout is decorative. `Logout` additionally passes a user id into a parameter named `sessionID`, so the revoke targets a row that does not exist and reports success.

**4. The WebSocket is anonymous.** `routes.go:197` registers it outside every auth group (`SEC-01`). Anyone with network reach subscribes to any session's live scan feed. The client helpfully appends the JWT to the query string, where it lands in nginx logs — and the server never reads it.

**5. It cannot be deployed.** The dashboard image bakes `VITE_API_BASE_URL=http://localhost:8080` at build time (`PR-05`), so it is unreachable from every device except its build host. There is no real migration runner (`PR-08`), no backups (`PR-03`), no metrics (`PR-02`), and no test suite that compiles (`T-01`).

## 4. What is genuinely good

This is not a codebase that needs rewriting. It needs finishing.

- **Clean architecture is really respected.** `domain/*.go` imports only `time`, `uuid`, GORM tags. Interfaces are declared by the consumer. This is the thing most teams get wrong first; here it is right.
- **The database design has real craft.** `difference` as a `GENERATED ALWAYS AS (...) STORED` column makes the invariant a property of the DB rather than of app code. `stock` carries a `version` column for optimistic locking. Partial indexes `WHERE deleted_at IS NULL` are selective and correct. `refresh_token_hash` is stored as a SHA256 hash, not plaintext. Access TTL is 15 minutes. These are the choices of someone who knows what they are doing.
- **The UI is the best-executed part of the product (78/100).** One menu source, three coherent shells, consistent DaisyUI tokens, chart colours read from theme tokens via `MutationObserver`, PWA with a *correct* caching policy (assets cache-first, `/api/` never cached), proper loading/empty/error branches on the two newest pages. It looks and feels like a real product.
- **The Go code reads well.** `%w` wrapping in repositories, `ShouldBindJSON` not `BindJSON`, uniform response envelope, `validate` tags on every DTO, mocks and table-driven tests already written.
- **The tests that exist describe the right design.** They reference `FindActiveByRefreshTokenHash`, `AccessExpiry`, `domain.ProductFilter` — precisely the things the implementation lacks. They were written against the intended system and then the system moved on without them.

## 5. Why the numbers are what they are

| Category | Why not higher | Why not lower |
|---|---|---|
| Backend Arch 52 | No graceful shutdown; config keys mismatch the env files by name in four places; rate limiter implemented and never mounted | Layering is genuinely correct; dependencies current |
| Frontend 61 | Strict TS but `any[]` at every API boundary; lint gate broken; 3 sites bypass the API client; role-less menu | Design system is coherent; a11y near-clean; build green |
| Functionality 38 | Reports return zero rows always; Excel export has headers only; the mobile app posts to a hardcoded fake session; "New Session" is dead | Sessions, products, warehouses, activity, settings all work |
| Database 45 | Two migration files are invalid SQL; no version tracking; a published admin hash ships in `initdb.d` | Schema itself is well-designed; constraints and indexes are thoughtful |
| Security 30 | Anonymous WS; unverified refresh; no rate limiting; any-origin credentialed CORS; no TLS | bcrypt 12; parameterised queries throughout; short access TTL |
| Code Quality 55 | 8 dead symbols that read as live features; tests do not compile; `/scan` and `/mobile` are divergent duplicates | Idiomatic; consistent error wrapping; no abandoned deps |
| UI/UX 78 | 4 blocking `alert()`; summary tiles mix scopes; staff stares at 403s | Best-in-class for an internal tool |
| Performance 45 | N+1 dashboard loading ~1 M rows and discarding them; non-sargable search; unbounded exports | Pool tuned; gzip on; `Preload` used; generated column saves report cost |
| Testing 12 | 5 test functions in 16 800 LOC; Go suite does not compile; dashboard has none; mobile test is the Flutter template; no CI | The Go tests are good tests, once fixed |
| Prod Readiness 20 | No TLS, no backups, no metrics, no real migrations, dashboard unreachable off-host | Health endpoint works; containers build and start |

## 6. Unfinished features — advertised vs. delivered

| Feature | Presented as | Actually |
|---|---|---|
| Reports | Full filterable report with CSV/Excel export | Endpoint returns `[]` unconditionally |
| Excel export (opname) | Downloadable report | Column headers, zero rows |
| CSV export | Respects the filters on screen | Reads 2 of the 10 filters the UI sends |
| Live progress | WebSocket updates on session detail | Broadcast is a TODO stub; `SubmitScan` broadcasts nothing |
| Notifications | Threshold alerts on differences | Generator function implemented, never called |
| Mobile scanner | Scan into an active session | Posts to `'dummy-session-id'`; no session picker exists |
| New Session | Primary action on `/sessions` | Route does not exist; 400 |
| Offline queue | Syncs queued scans | Rejected scans retry forever; claim of idempotency is a comment |
| Rate limiting | — | Fully implemented, unmounted |
| Stock write-back | — | Both methods implemented, never called |

## 7. Path to release

| Wave | Focus | Effort | Outcome |
|---|---|---|---|
| 0 | Version control, CI skeleton, delete the broken SQL, lint config | 1 day | Verifiable |
| 1 | Core cycle correct + regression tests | ~10 days | **Works** |
| 2 | Security + access control | ~5 days | **Safe** |
| 3 | Reports, exports, dashboard aggregates, mobile session picker | ~10 days | **Complete** |
| 4 | Shutdown, TLS, metrics, backups, migrations, E2E | ~10 days | **Operable** |
| 5 | Polish | ~5 days | — |
| | | **≈ 41 days** | |

**11 working days** gets the product to "it does what it claims". **36 working days** gets it to "it can be run in a warehouse by people who are not its author".

## 8. Verdict

**NOT APPROVED FOR PRODUCTION.**

Three independent grounds, each sufficient alone:

1. **The core workflow does not close** — no second session per day, wrong row returned, stock never written back.
2. **The security posture has an open read path** — anonymous WebSocket, revoked-token acceptance, committed admin password, no TLS.
3. **It cannot be deployed or verified** — dashboard unreachable off-host, no migration path, no backups, no metrics, no compiling test suite.

The engineering underneath is sound and the frontend is genuinely good. The debt is **unfinished work, not bad work** — which is why the estimate is weeks rather than months, and why the fix list is mostly small items in a clear order.

## 9. Audit artefacts

| File | Contents |
|---|---|
| `00-executive-summary.md` | this document |
| `01-project-architecture.md` | repo map, layering verification, 7 architecture defects |
| `02-backend-audit.md` | BE-01…BE-18 |
| `03-frontend-audit.md` | FE-01…FE-11, verified build gates |
| `04-database-audit.md` | DB-01…DB-12, migration strategy |
| `05-functionality-audit.md` | route reachability, FN-01…FN-13 |
| `06-security-audit.md` | OWASP Top 10 mapping, SEC-01…SEC-16 |
| `07-code-quality.md` | CQ-01…CQ-12, dead-code inventory |
| `08-performance-audit.md` | PERF-01…PERF-13 |
| `09-testing-audit.md` | T-01…T-09, verified gate output |
| `10-production-readiness.md` | PR-01…PR-14 |
| `11-bug-registry.md` | all 124 findings, one table, with fix + verification |
| `12-improvement-roadmap.md` | Waves 0–5, effort estimates, critical path |
| `13-release-checklist.md` | 33 blockers, 25 manual scenarios, sign-off gate, rollback plan |

## 10. Honest limits of this audit

- **No stack was run.** No `docker compose up`, no live HTTP, no browser, no device. Every behavioural claim is from source reading.
- The following were **not executed** and are marked **NOT TESTED** throughout: `go test`, `flutter analyze`, `flutter test`, any E2E, any load test, any live auth probe, `govulncheck`, `npm audit`, actual response headers, actual HTTP status codes per branch.
- Verified by running: `go build ./...` (pass), `go vet ./...` (fail), `go test ./...` (fail, no compile), `npx svelte-check` (pass, 1 warning), `npm run build` (pass), `npx eslint .` (fail), `npm run lint` (fail).
- **Build success ≠ production ready.** The build passes; the product does not ship.
