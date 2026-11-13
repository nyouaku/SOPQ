# 09 — Testing Audit

**No test was executed against a live stack this session.** The commands below were run earlier in the session and their output is reproduced verbatim; they were not re-run.

## A. Test inventory

| Stack | Test files | Test cases | Framework |
|---|---|---|---|
| Backend Go | 2 (`auth_usecase_test.go` 251 L, `stock_opname_usecase_test.go` 275 L) | 4 `Test*` funcs, 6 subtests | stdlib `testing` + hand-written mocks |
| Dashboard | **0** | 0 | none installed |
| Mobile | 1 (`test/widget_test.dart`, 29 L) | 1 | `flutter_test` |
| Integration / E2E | **0** | 0 | none |
| Load / perf | **0** | 0 | none |

Total 5 test functions across ~16 800 LOC. Coverage: **NOT MEASURED** (`go test -cover` cannot run — the package does not compile).

## B. Verified gate results

```
$ go build ./...
(exit 0)                                          ← PASS

$ go vet ./...
internal\usecase\stock_opname_usecase_test.go:144:73: undefined: domain.ProductFilter
                                                  ← FAIL

$ go test ./...
# ... [build failed]  ← does not compile
  auth_usecase_test.go     : mock missing List(ctx, string, int, int)
                             missing FindActiveByRefreshTokenHash
                             too many arguments in call to authUsecase.Register
                             unknown field AccessExpiry / RefreshExpiry
  stock_opname_usecase_test.go : undefined domain.ProductFilter
                             undefined domain.ActivityLogFilter

$ npx svelte-check --tsconfig ./tsconfig.json
0 errors, 1 warning (a11y, sessions/+page.svelte:152)   ← PASS

$ npm run build
PASS (adapter-node)

$ npx eslint .
ESLint couldn't find a configuration file.        ← FAIL

$ npm run lint
prettier: 35 files need formatting → aborts     ← FAIL

$ flutter test
NOT RUN (no device/toolchain confirmed)           ← NOT TESTED
```

Zero of the five CI-worthy gates pass coherently: `build` and `svelte-check` pass, but `vet`, `test`, `lint` fail.

## C. Findings

### T-01 — CRITICAL — Go tests do not compile; the suite has never run
The whole `internal/usecase` package fails to build in test mode. Causes, all interface drift:

| Test expects | Code provides |
|---|---|
| `domain.ProductFilter` | `repository.ProductFilter` |
| `domain.ActivityLogFilter` | `repository.ActivityLogFilter` |
| `userRepo.List(ctx, role string, limit, offset int)` | `List(ctx, filter repository.UserFilter)` |
| `sessionRepo.FindByUserID`, `DeleteByUserID` | `FindByID`, `FindActiveByRefreshTokenHash`, `RevokeByID`, `RevokeByUserID` |
| `sessionRepo.FindActiveByRefreshTokenHash` | **absent from the mock** — and never called by production code (SEC-04). |
| `authUsecase.Register(ctx, user, pass, name, role, nil)` — 6 args | 5 args |
| `config.JWTConfig{AccessExpiry, RefreshExpiry}` | `AccessTTLMinutes`, `RefreshTTLDays` |

The mocks were written against an earlier schema and never updated. Nobody ran `go test` after the last refactor — this is not a subtle edge-case failure, it is a `cannot compile`.

**Consequence beyond the missing safety net:** the tests are the only artefact that describes the *intended* design (`FindActiveByRefreshTokenHash` on the session repo, TTL-in-minutes config, `Filter` in the repository layer). They are worth fixing before writing new tests.

---

### T-02 — CRITICAL — Mobile test asserts on an app that does not exist
`mobile/test/widget_test.dart`
```dart
import 'package:stock_opname_mobile/main.dart';
...
await tester.pumpWidget(const MyApp());          // main.dart exports StockOpnameApp
expect(find.text('0'), findsOneWidget);          // no counter in the app
await tester.tap(find.byIcon(Icons.add));        // no add button
```
`MyApp` is undefined; the app is `StockOpnameApp` and its `home` is `LoginScreen`. This is the untouched `flutter create` template. `flutter test` cannot pass.

**Fix:** delete it and write three real widget tests (login form validation, scanner screen renders, offline banner).

---

### T-03 — CRITICAL — Dashboard has zero tests and no test runner
`dashboard/package.json` has no `test` script, no `vitest`, no `@testing-library/svelte`, no Playwright. The largest and most interactive surface in the product (~7 500 LOC, 9 routes, 5 modals, 3 filter sets) has no automated verification of any kind. The `createQuery` reactivity bug (FE-01) is exactly the class of defect a single component test with two filter values would have caught.

**Fix:** add `vitest` + `@testing-library/svelte` (already have Vite). Priority tests: `queries.ts` (params → queryKey; the FE-01 regression), `offline/queue.ts` (enqueue/flush/reject), `apiClient` refresh-on-401.

---

### T-04 — HIGH — No integration or E2E test exists for the only workflow that matters
The core cycle — login → create session → scan → complete → approve → report — has **zero** automated coverage at any level. Every defect in this audit on that path (BE-01 code collision, BE-02 wrong row, BE-04 no write-back, BE-05 empty reports, FN-01 new-session button dead, FN-10 mobile dummy session) is a defect a single happy-path E2E would have surfaced.

**Fix:** one Playwright test against `docker compose up` — the cheapest high-value test in the repo. It exercises all four services.

---

### T-05 — HIGH — No CI pipeline at all
Root contains no `.github/`, no `gitlab-ci.yml`, no `Jenkinsfile`, no `Makefile`. Every gate listed in section B is a manual step that a human may skip — and demonstrably did (`go vet` and `go test` are red on the committed tree).

Combined with the missing root git repo (REPO-01), there is no automated enforcement of anything.

**Fix:** GitHub Actions with three jobs: `go build + vet + test`, `npm ci + check + build + lint`, `flutter analyze + test`. Fail on any red. This alone would have blocked the current tree from merging.

---

### T-06 — HIGH — `go vet` failing means static analysis is not wired
`go vet` is the cheapest correctness check in Go and it is red. `staticcheck` and `golangci-lint` are absent. `gosec` absent (would flag SEC-05, SEC-13). `govulncheck` absent.

Dashboard: `eslint` installed and scripted but **no config file** (FE-ARCH-01) — so the lint gate is decorative. Prettier fails on 35 files.

Mobile: `analysis_options.yaml` exists with `flutter_lints` and `riverpod_lint` — the only stack with a working static-analysis config. Result **NOT TESTED** (`flutter analyze` not run this session).

---

### T-07 — MEDIUM — The two existing Go tests target the wrong layer
`CalculateDifference` and `OptimisticLocking` test pure arithmetic and a repository contract that the usecase does not own. Meanwhile the usecases' actual control flow — the code-collision path (BE-01), the wrong-row reload (BE-02), the never-called stock write-back (BE-04), the transaction boundary in approve — has no test.

Once compiling (T-01), the highest-value additions are:
1. `StartSession` twice on the same day → second succeeds (regression for BE-01).
2. `SubmitScan` with a concurrent insert → returns *this* detail (BE-02).
3. `ApproveSession` → `stock.quantity_system` updated in the same transaction (BE-04).
4. `RefreshToken` with a revoked token → error (SEC-04).

Each is currently a **known-broken behaviour**, which makes them the ideal first tests: they must fail before the fix and pass after.

---

### T-08 — MEDIUM — No test data strategy
No fixtures, no factories, no seed data for tests, no test DB. `seed_users.sql` is broken and contains a real-looking credential (DB-04/DB-05). Any future test needs its own data story — `testcontainers-go` or a `docker compose -f docker-compose.test.yml` with a throwaway pg volume.

---

### T-09 — LOW — Manual acceptance criteria are undocumented
The `docs/` folder has 20 markdown files but no test plan, no acceptance criteria per feature, and no QA checklist. "Build success ≠ production ready" (stated in the audit brief) is exactly the gap: there is a build, and no definition of correct.

## D. What good looks like here

| Gate | Now | Target |
|---|---|---|
| `go build ./...` | PASS | PASS |
| `go vet ./...` | **FAIL** | PASS |
| `go test ./... -race -cover` | **FAIL (no compile)** | PASS, ≥60 % on `usecase`/`repository` |
| `golangci-lint run` | absent | PASS |
| `npm run check` | PASS (1 a11y warn) | PASS (0 warnings) |
| `npm run build` | PASS | PASS |
| `npm run lint` | **FAIL** | PASS |
| `npm run test` | **missing** | PASS |
| `flutter analyze` | NOT TESTED | PASS |
| `flutter test` | **FAIL (undefined class)** | PASS |
| E2E happy path | **missing** | 1 passing |
| CI running all of the above | **missing** | on every push |

## E. Assessment
This is the weakest category in the product. The backend has the *shape* of a tested codebase — mocks, table-driven subtests, a repository interface — but the tests have not compiled since an interface refactor, so they have provided zero protection for some time. The dashboard has none. The mobile test is the Flutter template, unmodified. There is no CI to notice any of it.

The mitigating fact: the mocks and the existing test file encode the correct target design. Fixing the compilation errors (T-01) is a few hours of work and simultaneously unblocks the regression suite for four of this audit's criticals.
