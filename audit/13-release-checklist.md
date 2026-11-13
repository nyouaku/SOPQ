# 13 — Release Checklist

Decision framework for going live. **Every box in §A must be ticked before release.** Unticked = not started, and the current state of the tree is listed honestly.

## A. Release blockers (must all be ✅)

### A1. Core business logic
| # | Requirement | State |
|---|---|---|
| A1.1 | Two stock-opname sessions can be created on the same day | ❌ BE-01 |
| A1.2 | A submitted scan returns that scan, not another user's | ❌ BE-02 |
| A1.3 | Approving a session updates `stock.quantity_system` | ❌ BE-04 |
| A1.4 | Logout actually ends the session | ❌ BE-03 / SEC-04 |
| A1.5 | "New Session" button reaches a create form | ❌ FN-01 |
| A1.6 | The mobile app can record a real scan against a real session | ❌ FN-10 |
| A1.7 | A permanently rejected scan leaves the offline queue | ❌ FN-02, FN-03 |
| A1.8 | Replaying a scan does not duplicate a detail row | ❌ FN-04 |

### A2. Security
| # | Requirement | State |
|---|---|---|
| A2.1 | No unauthenticated endpoint exposes opname data | ❌ SEC-01 (WS) |
| A2.2 | Reports/export endpoints enforce roles | ❌ SEC-02 |
| A2.3 | No default or committed credential ships | ❌ SEC-08, SEC-15, PR-11 |
| A2.4 | Login is rate-limited | ❌ SEC-06 |
| A2.5 | CORS does not reflect arbitrary origins with credentials | ❌ SEC-07 |
| A2.6 | No secrets in client logs in release builds | ❌ SEC-10 |
| A2.7 | TLS terminates at the edge | ❌ PR-04 |
| A2.8 | Security headers present (CSP, nosniff, frame-options, referrer) | ❌ SEC-14 |

### A3. Data
| # | Requirement | State |
|---|---|---|
| A3.1 | Fresh deployment creates the full schema deterministically | ❌ PR-08, DB-04 |
| A3.2 | Schema changes reach an existing deployment | ❌ PR-08 |
| A3.3 | Backup runs and a restore has been rehearsed | ❌ PR-03 |
| A3.4 | `input_method` cannot be written with an unmodelled value | ❌ DB-07 |

### A4. Operability
| # | Requirement | State |
|---|---|---|
| A4.1 | SIGTERM drains in-flight requests | ❌ PR-01 |
| A4.2 | `/metrics` + request IDs exist | ❌ PR-02 |
| A4.3 | Healthchecks gate startup order | ❌ PR-06 |
| A4.4 | The dashboard works from a device that is not the build host | ❌ PR-05 |
| A4.5 | Configuration keys are actually read | ❌ CQ-01 |
| A4.6 | Logs rotate; containers have memory limits | ❌ PR-10 |
| A4.7 | Deployment runbook exists | ❌ PR-13 |

### A5. Verification
| # | Requirement | State |
|---|---|---|
| A5.1 | `go build ./...` | ✅ PASS |
| A5.2 | `go vet ./...` | ❌ FAIL |
| A5.3 | `go test ./... -race` | ❌ FAIL (no compile) |
| A5.4 | `npm run check` | ✅ PASS (1 a11y warning) |
| A5.5 | `npm run build` | ✅ PASS |
| A5.6 | `npm run lint` | ❌ FAIL |
| A5.7 | Dashboard tests | ❌ none exist |
| A5.8 | `flutter analyze` | ⚠️ NOT TESTED |
| A5.9 | `flutter test` | ❌ FAIL (undefined class) |
| A5.10 | E2E happy path green | ❌ none exists |
| A5.11 | CI enforces A5.1–A5.10 | ❌ none exists |

**Blockers open: 27 of 33. Blockers closed: 3. Untested: 1.**

## B. Functional acceptance (manual, on a running stack)

None of these has been executed. All **NOT TESTED**.

| # | Scenario | Expected |
|---|---|---|
| B1 | Admin logs in | dashboard loads, tiles populated within 3 s |
| B2 | Dashboard tiles | "Aktivitas terbaru" shows rows (requires BE-09) |
| B3 | Create session for warehouse A | 201, session appears in list |
| B4 | Create a second session, same day, warehouse A | **must succeed** (currently 400) |
| B5 | Staff opens `/scan`, types a barcode + qty | 201, detail appears in session |
| B6 | Staff scans via camera | accepted (currently 400 — BE-08) |
| B7 | Staff scans with no network | queued, badge increments |
| B8 | Network returns | queue flushes, no duplicates |
| B9 | Staff scans a product not stocked in that warehouse | explicit error, not silently `stockSystem = 0` (BE-17) |
| B10 | Session detail with two browsers open | new scan appears without reload (currently not — FN-06) |
| B11 | Complete session | status `completed` |
| B12 | Approve session | status `approved`, `stock` reflects physical count (currently not — BE-04) |
| B13 | Approve the same session twice, concurrently | second rejected, first `approved_by` preserved (currently last-write-wins — BE-10) |
| B14 | `/reports` with 8 filters applied | non-zero rows, correct count (currently zero always — BE-05) |
| B15 | CSV export with the same 8 filters | file matches the table (currently ignores 6 — BE-07) |
| B16 | Excel export | contains data rows (currently headers only — BE-06) |
| B17 | Staff opens `/reports` | menu hidden or 403 with a clear message (currently a bare 403) |
| B18 | Staff calls `GET /export/products` | 403 (currently 200 with the full catalogue — SEC-02) |
| B19 | Unauthenticated client opens `/ws/stock-opname/sessions/<uuid>` | connection refused (currently accepted — SEC-01) |
| B20 | Logout, then reuse the refresh token | 401 (currently mints new tokens — SEC-04) |
| B21 | Expired access token, then an export | transparent refresh (currently hard failure — FE-04) |
| B22 | Mobile app on a phone on the LAN | reaches the backend (currently `localhost` — MOB-ARCH-01) |
| B23 | Dashboard opened from a phone on the LAN | reaches the backend (currently `localhost` — PR-05) |
| B24 | 11 login attempts in 60 s | 429 after 10 (currently unlimited — SEC-06) |
| B25 | `docker compose up` on a clean host, fresh volume | schema + seed complete, dashboard reachable (currently 2 SQL files abort — DB-01…DB-04) |

## C. Non-functional acceptance

| # | Requirement | Target | Measured |
|---|---|---|---|
| C1 | Dashboard load p95 | < 2 s | **NOT MEASURED** |
| C2 | Scan submit p95 | < 500 ms | **NOT MEASURED** |
| C3 | Session detail with 5 000 details | < 1 s, paginated | **NOT MEASURED** |
| C4 | Report query with 8 filters, 100 k rows | < 3 s | **NOT MEASURED** (feature absent) |
| C5 | Concurrent scanners | 20 simultaneous, no duplicate rows | **NOT MEASURED** |
| C6 | Offline queue recovery | 500 queued scans flush without loss | **NOT MEASURED** |
| C7 | Container image size | backend < 50 MB | **NOT MEASURED** |
| C8 | Memory under normal load | < 512 MB per container | **NOT MEASURED** |
| C9 | Restore from backup | < 30 min, RPO ≤ 24 h | **NOT MEASURED** |
| C10 | Graceful shutdown | zero dropped requests on SIGTERM | **NOT MEASURED** |

No performance or reliability claim about this system can be made, because no measurement exists.

## D. Documentation acceptance

| # | Document | State |
|---|---|---|
| D1 | README with quickstart | ✅ exists |
| D2 | API reference matching the code | ⚠️ `docs/API_ENDPOINTS.md` exists; drift unverified against handlers |
| D3 | Schema document matching the migrations | ❌ two divergent files (DB-12) |
| D4 | Deployment guide + env-var reference | ❌ absent (PR-13) |
| D5 | Backup / restore runbook | ❌ absent |
| D6 | Incident response / on-call | ❌ absent |
| D7 | Test plan / acceptance criteria | ❌ absent (T-09) |
| D8 | Changelog maintained | ⚠️ `CHANGELOG.md` exists; currency unverified |

## E. Sign-off gate

A release may proceed only when:

1. **All §A boxes are ✅** — no exceptions, no "we'll fix it after".
2. **§B has been executed on a staging deployment** and every row matches its expected column, with results recorded (not "looks fine").
3. **At least C1, C2, C5 and C10 are measured** and within target.
4. **A restore from backup has been performed** by someone other than the author of the backup script.
5. **CI is green on the release commit**, and the commit is tagged.
6. Sign-off recorded from: technical lead, the warehouse operations owner, and whoever will be on call.

Current state of that gate: **0 of 6 satisfied.**

## F. Rollback plan (currently impossible)

A release today has no rollback path:
- No image tagging strategy — `docker compose build` produces untagged local images.
- No migration down-test — the `.down.sql` for `000001` has never been executed.
- No previous-release image to redeploy.
- No database snapshot taken immediately before deploy.

Minimum viable rollback, to be in place before the first deploy:
```bash
docker compose exec -T postgres pg_dump -U stockopname -Fc stockopname_db > pre-deploy-$(date +%F-%H%M).dump
docker tag stockopname-backend:latest stockopname-backend:$(git rev-parse --short HEAD)
```
and a documented `docker compose up -d --no-build` against the tagged image.

## G. Verdict

**NOT APPROVED FOR PRODUCTION.**

Three independent grounds, any one of which is sufficient:

1. **The core business cycle does not close.** A session cannot be created twice in a day (BE-01), the scan response can name another user's row (BE-02), and approving never updates `stock` (BE-04). The product's entire reason to exist is not functioning.
2. **The security posture has a public read path.** An unauthenticated WebSocket (SEC-01) and a revoked-token-accepting refresh flow (SEC-04) are exploitable by anyone with network access, and a published admin password ships in the migrations directory (SEC-08).
3. **It cannot be deployed or verified.** The dashboard container is unreachable from any device but its build host (PR-05), schema changes never reach an existing database (PR-08), there are no backups (PR-03), no metrics (PR-02), and no test suite that runs (T-01…T-05).

Every one of these is fixable, and most are small. See `12-improvement-roadmap.md` — Waves 0–1 (≈11 working days) close the first ground and most of the third; Waves 2–4 (≈25 working days) close the second.
