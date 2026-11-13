# 10 — Production Readiness

**No stack was started during this audit.** Items below are from source, `docker-compose.yml`, `Dockerfile`s, and `nginx.conf` inspection. Anything requiring a live environment is **NOT TESTED**.

## A. Readiness scorecard

| Dimension | Status | Notes |
|---|---|---|
| Containerized | PASS | 5 services, healthcheck on `postgres`. |
| Health endpoint | PASS | `/health` pings DB + Redis, returns 503 when unhealthy. |
| Readiness probe / startup ordering | **FAIL** | `backend` waits on postgres health, but there is no readiness probe for `backend` itself and `redis` only waits on `service_started` (redis may reject first connects). |
| Graceful shutdown | **FAIL** | `main.go` `router.Run` + `<-quit` with no `http.Server.Shutdown` — in-flight requests dropped on SIGTERM. |
| Structured logging | PARTIAL | `zap` used, JSON in prod, but logs full `RawQuery` — the WS `?token=` lands in them. |
| Metrics | **FAIL** | No `/metrics`, no Prometheus, no histogram. |
| Tracing | **FAIL** | No OpenTelemetry. |
| Request ID | **FAIL** | `X-Request-ID` not generated, not echoed to client. |
| Centralized logging | **FAIL** | No sink. `docker logs` only. |
| Backup / restore | **FAIL** | `pgdata` named volume with no backup job, no PITR, no restore drill documented. |
| Migration strategy | **FAIL** | `RunMigrations()` is a no-op that returns "skip"; the only migration path is `docker-entrypoint-initdb.d`, which runs once ever. |
| Rollback strategy | **FAIL** | Not documented, no deploy tagging, no image version pinning. |
| Secrets management | **FAIL** | `.env` on the host, values also committed into `docker-compose.yml` defaults. |
| HTTPS/TLS | **FAIL** | nginx listens on `:80` only; no `listen 443 ssl`, no certbot, no HSTS. |
| Autoscaling | **FAIL** | Hardcoded `container_name` on every service prevents scaling (`Error response from daemon: Conflict. The container name "/stockopname-backend" is already in use`). |
| Resource limits | **FAIL** | No `mem_limit` / `cpus` / `deploy.resources`. |
| Log rotation | **FAIL** | No `logging:` block; default json-file driver grows without bound. |
| Image versions | PARTIAL | Backend `golang:1.23-alpine` pinned but wrong (BE-16); `nginx:stable-alpine` unpinned. |
| Rate limiting / WAF | **FAIL** | `rate_limit.go` unmounted. |
| Security headers | **FAIL** | See SEC-14. |
| Dependency audit | **NOT TESTED** | `govulncheck`/`npm audit` not run (no network). |
| Observability dashboard | **FAIL** | None. |
| On-call / runbook | **FAIL** | No runbook in `docs/`. |

## B. Findings

### PR-01 — CRITICAL — No graceful shutdown → dropped requests on every deploy
`cmd/api/main.go:75-87`
```go
go func() {
    if err := router.Run(addr); err != nil { logger.Fatal(...) }
}()
<-quit
logger.Info("Shutting down server...")
logger.Info("Server exited")
```
`router.Run` boots a bare `*gin.Engine`; no `*http.Server`, so there is no `.Shutdown(ctx)` to call. When docker sends SIGTERM the process exits immediately, dropping in-flight requests. During the scan phase of a stock opname this means **lost scans**.

Amplified by the WebSocket endpoint: long-lived WS connections are severed with no draining window and the client has no reconnect-with-backoff (`session_store.ts` opens once).

**Fix:** `srv := &http.Server{...}; go srv.ListenAndServe(); <-quit; ctx, cancel := context.WithTimeout(..., 30*time.Second); srv.Shutdown(ctx)`; add `srv.RegisterOnShutdown(hub.Close)` for the WS hub; add `docker-compose: stop_grace_period: 30s`.

---

### PR-02 — CRITICAL — No metrics, no tracing, no request ID
The system has no observability story:
- No `/metrics` endpoint. Prometheus cannot scrape it. No request rate, no error rate, no latency histogram, no DB pool stats, no GC.
- No OpenTelemetry SDK. No trace propagation between dashboard → backend → DB.
- `middleware/logger.go` logs method/path/status/latency but **no request ID**, so a 500 in the browser cannot be correlated to a backend log line. The API response envelope also omits a request ID field.

For a warehouse-floor app this matters concretely: a "scan not saving" report from a worker is currently undiagnosable — there is no way to find the corresponding request server-side.

**Fix:** Prometheus middleware (`/metrics` on a non-public port or `metrics_network`), OpenTelemetry with `gin` + `gorm` plugins, `X-Request-ID` generated in `middleware/logger.go` and echoed in the response.

---

### PR-03 — CRITICAL — No backup or restore story
`docker-compose.yml:74` declares `pgdata:` with no backup job, no scheduled `pg_dump`, no WAL archiving, no restore runbook. Stock-opname data is the operational record of inventory accuracy for a warehouse — losing it is a business-level incident.

**Fix:** minimum viable: nightly `pg_dump -Fc` to an off-container volume with 14-day retention + one documented restore drill. Target state: `wal-g`/`pgBackRest` + PITR.

---

### PR-04 — HIGH — No HTTPS at the edge
`docker/nginx.conf:9` `listen 80;` — no TLS block. Every login POST sends the password in cleartext over the internal network, and every JWT is transmitted in the clear. For any deployment reachable outside localhost this is a credential-interception risk (SEC-07 rides along).

**Fix:** `listen 443 ssl http2;` with certbot, HSTS, and an HTTP→HTTPS redirect. Behind a corporate LB, terminate TLS there and require `X-Forwarded-Proto: https` at nginx.

---

### PR-05 — HIGH — Baked-in `localhost` in the dashboard container
`docker-compose.yml:49-51`
```yaml
args:
  VITE_API_BASE_URL: http://localhost:8080/api/v1
  VITE_WS_BASE_URL: ws://localhost:8080/api/v1
```
`dashboard/Dockerfile:6-9` `ARG` → `ENV` → `npm run build`. Vite inlines these at build time into the client bundle. The container therefore **ships with URLs that only work from a browser on the host that built the image**. Any device — a warehouse tablet, a phone on the same LAN, another workstation — opens the dashboard and calls `http://localhost:8080` on *itself*.

This is a deployment-blocker. The dashboard is unreachable except from the build host.

**Fix:** make the dashboard a same-origin client (`/api/v1` relative path, nginx proxies it — it already does for `/api/`). Delete the `VITE_*` build args entirely; the frontend should never know the API origin.

---

### PR-06 — HIGH — No readiness probe; redis not health-gated
`docker-compose.yml:37-41` gates `backend` on postgres *healthy* but only `service_started` on redis. `backend` starts before redis accepts connections. `RedisClient` init in `main.go` is guarded (`if err != nil → nil`) and `rate_limit.go` fails-open when nil, so today the app limps — but the failure mode is silent and undetected.

There is also no `healthcheck` on `backend` itself and no `depends_on: dashboard: condition: service_healthy` in nginx — nginx can start proxying to a container that is still booting.

**Fix:** add a `HEALTHCHECK`/`healthcheck` (curl `/health` or the Go binary itself) to `backend` and `dashboard`; gate `nginx` on both; set redis healthcheck (it is a one-liner).

---

### PR-07 — HIGH — Compose file prevents horizontal scaling
Every service sets `container_name:`. Docker will refuse a second instance of any of them (`name already in use`). The system is architecturally pinned to exactly one backend, one dashboard, one nginx instance. For the current user count that is acceptable; for growth it is a rewrite.

**Fix:** drop `container_name` (compose v2 project-naming is sufficient), remove `ports:` exposure on `backend` (reach it via the internal network only), use `deploy.replicas` + a reverse proxy or swarm/k8s.

---

### PR-08 — HIGH — `RunMigrations()` is a no-op masquerading as a migration runner
`postgres.go:70-82`
```go
func (d *Database) RunMigrations() error {
    if d.DB.Migrator().HasTable("users") {
        logger.Info("Tables already exist, skipping migration")
        return nil
    }
    logger.Warn("Migration harus dijalankan manual: psql -U user -d dbname -f migrations/000001_init_schema.up.sql")
    return nil
}
```
The function is called on every boot. It either does nothing (tables exist) or emits a log line telling a human to run SQL by hand. In a containerised deployment the human path is impossible — there is no `psql` on the API image and no interactive shell.

Net effect: a **fresh** deployment is seeded by `docker-entrypoint-initdb.d` (which fails on two of the mounted files — DB-01…DB-04); an **existing** deployment never receives a schema change. There is no mechanism by which a schema edit reaches production.

**Fix:** golang-migrate as a binary, run as an entrypoint step before the API starts, with a single source of truth in `migrations/` and no `initdb.d` mount.

---

### PR-09 — MEDIUM — Container images pin toolchains inconsistently
| Image | Pinned | Issue |
|---|---|---|
| `golang:1.23-alpine` | yes | Below `go 1.25.0` in `go.mod` (BE-16). Relies on `GOTOOLCHAIN=auto` to download at build time — fails in air-gapped CI. |
| `node:22-alpine` | yes | OK. |
| `postgres:16-alpine` | yes, with `pg_isready` healthcheck | OK. |
| `redis:7-alpine` | yes, **no** healthcheck | OK-but-unhealthy (PR-06). |
| `nginx:stable-alpine` | no | `stable` is a moving tag; rebuilds are not reproducible. |

Also `docker-compose.yml:1` `version: "3.9"` — deprecated in Compose v2, ignored. Cosmetic.

---

### PR-10 — MEDIUM — No resource limits, no log rotation, no restart policy on redis
- `restart: unless-stopped` on all — good.
- No `mem_limit` / `cpus`. A runaway `excelize` import (BE-14) can OOM the host and take every other service with it.
- No `logging:` block. Default `json-file` grows unbounded. `docker system prune` is the only brake.
- No `stop_grace_period` — see PR-01.

**Fix:** `deploy.resources.limits`, `logging.options.max-size: 10m / max-file: 3`, `stop_grace_period: 30s`.

---

### PR-11 — MEDIUM — Compose leaks credentials into the repo
`docker-compose.yml:9-10` `POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme}` — the default is a literal weak password committed into source. `.env` on the host is the intended override; the compose file does not require it to exist. In any deployment where the operator does not set `DB_PASSWORD`, Postgres is reachable on the LAN with `stockopname / changeme` (and its `5432:5432` port is published to the host).

Same class as DB-05 (committed seed password). The port publish is the more severe half: `5432:5432` and `6379:6379` are exposed to the host network. Redis has no auth at all.

**Fix:** remove port mappings for `postgres`/`redis` (they are only used by the internal network), require `.env` with `env_file: .env` at the project level, fail-fast if missing.

---

### PR-12 — MEDIUM — Config drift between `.env.example`, `config.go`, and the compose file
| Key | `.env.example` | `config.go` reads | compose passes |
|---|---|---|---|
| Port | `SERVER_PORT` | `APP_PORT` | — |
| JWT access TTL | `JWT_ACCESS_EXPIRY` | `JWT_ACCESS_TTL_MINUTES` | — |
| JWT refresh TTL | `JWT_REFRESH_EXPIRY` | `JWT_REFRESH_TTL_DAYS` | — |
| Env name | `ENV` | `APP_ENV` | — |
| CORS | `CORS_ALLOWED_ORIGINS` | reads it, defaults `*` | not passed |

`backend` uses `env_file: .env` so the host's `.env` reaches the container — but four of the operator-tunable keys are simply not read. **CORS_ALLOWED_ORIGINS is the operator's most important production key and it is read from a name that the `.env.example` documents but no deployment guide mentions.**

---

### PR-13 — MEDIUM — No documentation of the deployment procedure
`docs/` has 20 markdown files (schema, API, design). There is no `DEPLOYMENT.md`, no runbook, no env-var reference, no incident-response section, no backup procedure. A new operator has `docker compose up -d` and that is the entire instruction.

---

### PR-14 — LOW — Health endpoint is unauthenticated but information-disclosing
`/health` returns `{database: healthy|unhealthy, redis: healthy|unhealthy|not_configured}`. Useful for the LB, also useful to an attacker mapping the internal topology. Behind the reverse proxy on port 80 with no auth. Acceptable for most deployments; restrict to a private network or an `X-Health-Token` header if exposed.

---

## C. NOT TESTED
- `docker compose up` end-to-end (no docker daemon access during the audit).
- Real HTTPS termination, cert issuance, HSTS.
- Actual image sizes and startup times.
- Whether `docker compose down/up` after a schema edit picks up the change (it will not — see PR-08) — verified by reading the code, not by running it.
- Load behaviour under concurrent scanners.
- Backup/restore.

## D. Assessment
Every container is built and starts. That is the ceiling of what works. Above that ceiling the deployment story is missing everything an operator needs: TLS, observability, backups, a real migration path, resource limits, log rotation, a runbook. The single worst operational fact is PR-05 — the dashboard container ships unreachable from any device other than its build host — because it means `docker compose up` **does not actually work** as a deployment, and the only way to use the product today is `npm run dev` on a laptop pointed at a local backend.

The good news: every one of PR-01…PR-14 is a configuration or infrastructure fix, not a rewrite.
