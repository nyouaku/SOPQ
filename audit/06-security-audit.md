# 06 — Security Audit (OWASP Top 10 2021 framing)

Scope: backend Go HTTP surface, JWT/session lifecycle, RBAC, CORS/headers, file upload/import, WebSocket, dashboard client token handling, mobile client, secrets in repo, container/nginx config.

**No live exploit was attempted. No request was sent to a running server.** All findings are from source inspection. Items requiring a live stack are marked **NOT TESTED**.

## A. OWASP Top 10 mapping

| OWASP | Applicable | Status |
|---|---|---|
| A01 Broken Access Control | **YES** | **FAIL** — see SEC-01, SEC-02, SEC-03 |
| A02 Cryptographic Failures | **YES** | **FAIL** — see SEC-04, SEC-05 |
| A03 Injection | YES | PASS (GORM parameterised; no raw SQL from input) — SEC-12 minor |
| A04 Insecure Design | YES | **FAIL** — see SEC-06 |
| A05 Security Misconfiguration | **YES** | **FAIL** — see SEC-07, SEC-08 |
| A06 Vulnerable & Outdated Components | YES | **NOT VERIFIED** — no `govulncheck` / `npm audit` run (offline). Versions look current. |
| A07 Identification & Auth Failures | **YES** | **FAIL** — see SEC-09, SEC-10 |
| A08 Software & Data Integrity Failures | YES | **FAIL** — see SEC-11 |
| A09 Security Logging & Monitoring | YES | **FAIL** — see SEC-13 |
| A10 SSRF | NO | No server-side URL fetch on any code path. |

## B. Findings

### SEC-01 — CRITICAL — WebSocket endpoint is completely unauthenticated
`backend/internal/delivery/http/route/routes.go:197`
```go
v1.GET("/ws/stock-opname/sessions/:id", websocket.HandleStockOpnameWS)
```
Registered on the bare `v1` group, outside every `AuthMiddleware` sub-group. No token read, no `ParseAccessToken`, no session ownership check. Anyone with network reach to the host can subscribe to any session UUID and receive live scan progress (product names, quantities, counts).

Amplifier: the dashboard client *does* append `?token=<JWT>` (`sessions/[id]/+page.svelte:61`) — the token is unused by the server, but it lands in nginx access logs, cluster ingress logs, and browser history (SEC-05).

**Impact:** information disclosure of the entire inventory reconciliation in real time; the leaked JWT in URLs is separately replayable for 15 minutes.
**Fix:** move under `authMW`, validate `token` via `ParseAccessToken`, verify the session's `warehouse_id` is one the user may see, then move the token to `Sec-WebSocket-Protocol` or an initial auth frame.

---

### SEC-02 — HIGH — Export endpoints are any-authenticated, no role check
`routes.go` groups `/export/products` and `/export/stock-opname` under plain `authMW` with no `RequireRole`. A `staff_gudang` account can download the complete product catalogue (barcodes, prices, suppliers, stock) in one request. Compare `/reports/stock-opname`, which is correctly gated to `supervisor|admin`.

**Fix:** apply `RequireRole("supervisor","admin")` to `/export/*`, or scope the product export to the caller's warehouse.

---

### SEC-03 — HIGH — Frontend menu shows privileged pages to unprivileged users
`dashboard/src/lib/components/layouts/menu.ts` — `MENU_ITEMS` has no `roles` field; no shell filters it. Staff sees Reports/Activity/Settings, clicks, gets 403. Backend is the real gate (so this is not a privilege escalation), but it is broken access-control UX and it leaks the existence and shape of privileged features.

**Fix:** `roles?: string[]` on `MenuItem`, filtered against `$authStore.user.role_name` in `LayoutWrapper`, `Sidebar`, `CommandNav`. Keep backend enforcement.

---

### SEC-04 — CRITICAL — Refresh token is never verified against the database
`backend/internal/usecase/auth_usecase.go:202-203`
```go
// TODO: Di production, tambahkan verifikasi refresh token hash di database
// Untuk PHASE 2, kita skip verifikasi hash dan langsung generate token baru
```
`RefreshToken` decodes the JWT and mints a new access+refresh pair **without** looking up `sessions.refresh_token_hash`. Consequences:

1. **Logout does not work.** `Logout` → `sessionRepo.RevokeByID(...)`, but nothing ever reads `revoked_at`. A stolen refresh token stays usable until its 7-day `exp`.
2. **No rotation detection.** A replayed old refresh token is indistinguishable from the current one.
3. **User deactivation does nothing.** `users.is_active = false` is not re-checked on refresh — a disabled account keeps minting access tokens for a week.

**Fix:** `sessionRepo.FindActiveByRefreshTokenHash(ctx, utils.HashToken(rt))`; nil ⇒ 401 `INVALID_REFRESH_TOKEN`; re-check `users.is_active`; rotate (revoke old, insert new) in one transaction. The repository method already exists (`session_repository.go:51`) and the failing unit test already references it.

---

### SEC-05 — HIGH — `VerifyToken` uses non-constant-time string comparison
`backend/pkg/utils/password.go:36-38`
```go
func VerifyToken(hashedToken, token string) bool {
	return hashedToken == HashToken(token)
}
```
Byte-wise `==` on a secret-derived value is timing-distinguishable. Low practical exploitability over a network (SHA256 hex, 64 bytes, attacker cannot submit arbitrary candidates without an endpoint that uses it — which none currently does), but it is the wrong primitive and will become a real oracle the moment SEC-04 is fixed and this function is wired in.

**Fix:** `subtle.ConstantTimeCompare([]byte(hashedToken), []byte(HashToken(token))) == 1`.

---

### SEC-06 — HIGH — No rate limiting on `login` or `refresh`
`backend/internal/middleware/rate_limit.go` implements Redis `INCR`-per-IP with a fail-open path. It is **never referenced** in `routes.go`. Login is therefore an unlimited password-oracle: bcrypt cost 12 slows but does not stop a distributed guess, and there is no lockout, no CAPTCHA, no per-username counter.

**Fix:** `auth.POST("/login", middleware.RateLimit(cfg.RedisClient, 10, time.Minute), handler.Login)`; add a per-username counter too (IP rotation defeats per-IP alone).

---

### SEC-07 — HIGH — CORS reflects any origin with credentials allowed
`backend/internal/middleware/cors.go`
```go
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
```
`allowedOrigins` comes from `CORS_ALLOWED_ORIGINS`; `config.go` defaults it to `*`, and `config.go` never reads the key the `.env.example` defines (BE-ARCH-04) — so in practice the default `*` wins. With `*` in the list, `o == "*"` matches and `Access-Control-Allow-Origin` is set to the **reflected** request origin, with credentials `true` on every response.

Today the API authenticates with `Authorization: Bearer` from `localStorage`, not cookies, so classic credential-CORS abuse is limited. It is still wrong, and it becomes critical the moment any cookie or `withCredentials` flow is introduced. Also note the header is set even when `allowed == false` — `Allow-Credentials: true` is emitted for disallowed origins, which is meaningless and misleading.

**Fix:** default `CORS_ALLOWED_ORIGINS` to the real dashboard origin; reject on mismatch (no header); never combine `*` with credentials; mount `gin-contrib/cors` or hand-roll the correct branch.

---

### SEC-08 — HIGH — `seed_users.sql` ships a published admin password
`backend/migrations/seed_users.sql:9-11` — `admin`, `supervisor`, `staff` all share one committed bcrypt hash, comment reads "Password untuk semua user: staff123". The file sits in the directory `docker-compose.yml` mounts to `docker-entrypoint-initdb.d`, so a fresh `docker compose up` creates three known-credential accounts.

`seed_user.sql` (singular) is additionally broken SQL (DB-04) and hardcodes `role_id` values.

**Impact:** trivially exploitable default credentials on any deployment that ran the seeded compose file.
**Fix:** remove both seed files from the mounted directory; ship a `make create-admin` one-off that generates a random password and prints it once, or seed from an env-provided hash.

---

### SEC-09 — HIGH — Weak registration validation with hardcoded role bounds
`backend/internal/delivery/http/handler/auth.go:26-31`
```go
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required,min=3"`
	RoleID   int    `json:"role_id" validate:"required,min=1,max=3"`
}
```
- 6-character minimum, no complexity, no breached-password check.
- `role_id` is a client-supplied integer clamped to `1..3` — **the role is chosen by the caller**, not by an admin workflow. Whether `POST /auth/register` is role-gated depends on the route group; `settings` calls it as admin, but the handler itself imposes no authorization. **NOT TESTED** — route placement not confirmed for every method.
- Bounds are `1..3` instead of resolving `roles.name`, so a `roles` reseed silently changes who gets admin (matches DB-05).
- Duplicate `username` surfaces as a 500 from the DB unique violation rather than a 409 (NOT TESTED against live stack).

**Fix:** min 12 chars + zxcvbn/breach list; `role_id` validated against `roles` table and restricted to the roles the caller may grant; admin-only route; map unique-violation to 409.

---

### SEC-10 — HIGH — Mobile app logs full request/response bodies in every build
`mobile/lib/core/network/dio_client.dart:24-31`
```dart
_dio.interceptors.add(PrettyDioLogger(
  requestHeader: true, requestBody: true, responseBody: true, ...
));
```
No `kDebugMode` / `kReleaseMode` guard. On a release build, every login (username + password in `requestBody`) and every Authorization header is written to logcat. A level-3 `adb logcat` on a rooted/lab device, or any crash-reporting sink that captures stdout, exfiltrates credentials.

**Fix:** wrap in `if (kDebugMode)`.

---

### SEC-11 — MEDIUM — Import accepts arbitrary uploaded files with no integrity check
`backend/internal/usecase/import_usecase.go` + the import handler: 5 MB body cap, extension/MIME sniff, then `excelize.OpenReader`. No row cap (~60k rows possible), no zip-bomb guard (excelize will expand), no schema validation before writes, non-transactional (BE-14). A malformed or hostile workbook can produce a partially-committed product table.

**Fix:** cap `rows <= 2000`, cap unzipped size, run inside `db.Transaction`, validate every row before the first write.

---

### SEC-12 — LOW — `notifications.metadata` built by string concatenation
`notification_usecase.go:~64` — `fmt.Sprintf(`{"product": "%s"}`, productName)`. A product name containing `"` or `\` produces invalid JSON (fails JSONB insert) — a data-integrity bug, and a textbook JSON-injection shape if this value ever reaches a consumer that does not re-parse. Use `json.Marshal`.

---

### SEC-13 — MEDIUM — Audit logging exists but is not comprehensive
`activity_logs` is populated for some operations (per `activity` page working end-to-end) but not for: login success/failure, logout, refresh, failed authorization, product import per row, export downloads, session creation failure. `middleware/logger.go` logs method + path + `RawQuery` — which is where the WS `?token=` lands (SEC-01). No request-ID correlation header is emitted to the client.

The `Logout` handler compounds this: `handler/auth.go:168-180` passes `userID` into a parameter named `sessionID` that flows to `RevokeByID` — so revoke targets a session row whose `id` equals a user id, i.e. usually none. It returns success regardless (`UPDATE ... RowsAffected == 0` is not checked). **Logout silently no-ops and logs nothing.**

**Fix:** audit rows for auth events (success + failure, with IP/UA); log user id not token; check `RowsAffected`; add `X-Request-ID`.

---

### SEC-14 — MEDIUM — No security headers at the edge
`docker/nginx.conf` sets `Access-Control-Allow-Origin *` (a second, independent wildcard — it overrides nothing but signals intent) and no `Content-Security-Policy`, `X-Content-Type-Options`, `Referrer-Policy`, `Strict-Transport-Security`, or `X-Frame-Options`. The SvelteKit app is a SPA with `localStorage`-held JWTs — a CSP is the main mitigation for any future XSS.

**NOT TESTED:** actual response headers from a running nginx (no stack running).

**Fix:** add a `add_header` block; CSP with `default-src 'self'`, `connect-src 'self' <api>`, no `unsafe-inline` for scripts.

---

### SEC-15 — MEDIUM — Secrets and configuration hygiene
- `.env` present in the working tree (values not reproduced here). Confirm it is git-ignored; the repo has no root `.git` so ignore rules are unverifiable — see REPO-01.
- `JWT_SECRET` default: **NOT VERIFIED** whether `config.go` hard-fails on an empty/default secret. If it falls back to a literal, every token is forgeable.
- `docker-compose.yml` supplies Postgres/Redis credentials as literal values in the compose file rather than an env file.
- No secret-management story (vault/SSM/k8s secret) documented anywhere.

**Fix:** fail-fast on missing `JWT_SECRET`; move compose credentials to `.env`; document the production secret source.

---

### SEC-16 — LOW — JWT has no `jti`, no issuer/audience
`pkg/utils/jwt.go` signs `{user_id, role_id, role_name, exp}`. No `jti` means the only revocation lever is `sessions.revoked_at` — which (SEC-04) is never read. No `iss`/`aud` means a token minted by any other service sharing the secret is accepted.

**Fix:** add `jti`, `iss`, `aud`; validate them on parse.

---

## C. Positive controls observed
- Passwords hashed with **bcrypt cost 12**, never logged.
- GORM used throughout with parameter binding; no string-concatenated SQL found in `internal/`.
- `RequireRole` middleware correctly reads `claims.RoleName` and aborts 403.
- Refresh token stored as **SHA256 hash**, not plaintext (`sessions.refresh_token_hash`) — correct choice.
- Access token TTL 15 minutes — appropriately short.
- Response envelope never echoes a password field in any DTO reviewed.
- `validator` tags on all request DTOs; `ShouldBindJSON` (not `BindJSON`) so bind errors are handled.
- Upload body capped at 5 MB.
- No `eval`, no `dangerouslySetInnerHTML`, no `{@html}` in the dashboard (grep clean).

## D. NOT TESTED
- Any live auth bypass, IDOR probe, or rate-limit measurement.
- JWT forgery (needs the running secret).
- Whether `POST /auth/register` is reachable by non-admin (route group not confirmed for all verbs).
- Actual response headers via `curl -I` against nginx.
- `govulncheck ./...` and `npm audit` (would require network).
- Mobile binary inspection (no release APK built).
- Whether `.env` is committed (no root git index).
