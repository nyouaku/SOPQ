# 01 — Project Architecture

Repo root: `stock-opname-app/`. Monorepo, **not** a git repo at root (nested `stock-opname-app/` contains a stale `.git`, 18 files total — likely an accidental nested clone; recommend removal after audit approval).

## Top-level layout
```
backend/   Go 1.25 / Gin 1.10 / GORM 1.30  — REST + WebSocket API
dashboard/ SvelteKit 2 / Svelte 4 / TS strict / Tailwind 3 + DaisyUI 4 — admin web
mobile/    Flutter 3 / Riverpod 2 / Dio 5 — scanner app
docker/    nginx.conf
docs/      20 .md + DATABASE_SCHEMA.sql (design intent, not always current)
migrations live under backend/migrations (9 files)
```

## Backend layering
Clean architecture, dependency direction `delivery -> usecase -> domain <- repository` is **respected**. Verified: `domain/*.go` imports only `time`, `uuid`, `gorm` tags — no `delivery`/`repository`/`usecase` imports.

- `cmd/api/main.go` — config → logger → postgres (`RunMigrations`) → redis (optional) → `gin.New()` → `route.SetupRoutes` → `go router.Run(addr)` → `<-quit`.
- 8 domain files, 11 repositories, 10 usecases, 9 handlers, 6 middlewares.
- `pkg/utils` (jwt, password, response), `pkg/validator`, `pkg/logger`.

## Frontend (dashboard) layering
- `src/lib/api/{config,client,queries}.ts` — endpoint map, Bearer client with refresh-once-on-401, svelte-query factories.
- `src/lib/stores/{auth,toast,theme,...}` — auth persisted in `localStorage`.
- `src/routes` — 8 pages: `/` dashboard, `/scan`, `/sessions`, `/sessions/[id]`, `/products`, `/warehouses`, `/reports`, `/activity`, `/settings`.
- `src/lib/components/layouts/{menu.ts, LayoutWrapper, Sidebar, Topnav, CommandCenter}` — single `MENU_ITEMS` source for 3 shells.
- PWA: `src/service-worker.ts`, `static/manifest.json`.

## Mobile layering
Feature-first per CLAUDE.md. `core/{constants,network,providers,storage,theme}`, `data/{datasources,models}`, `features/{auth,scanner}`. Riverpod `StateNotifier` + `Provider`. Offline queue in sqflite (`scan_queue`, `cached_sessions`, `cached_products`).

## Architecture-level defects (detail in category files)

| ID | Severity | Issue |
|----|----------|-------|
| BE-ARCH-01 | CRITICAL | `main.go` has no `http.Server`/`Shutdown()` — in-flight requests dropped on SIGTERM. |
| BE-ARCH-02 | HIGH | `middleware/rate_limit.go` is fully implemented but **never mounted** in `routes.go` — dead code, no rate limiting. |
| BE-ARCH-03 | MEDIUM | `dashboard` baked-time `VITE_API_BASE_URL` defaults to `http://localhost:8080/api/v1` inside the client bundle; any non-localhost deploy calls the visitor's own localhost. |
| BE-ARCH-04 | MEDIUM | Config keys mismatch: `.env`/`.env.example` define `SERVER_PORT`/`JWT_ACCESS_EXPIRY`/`JWT_REFRESH_EXPIRY`/`ENV`; `config.go` reads `APP_PORT`/`JWT_ACCESS_TTL_MINUTES`/`JWT_REFRESH_TTL_DAYS`/`APP_ENV`. Defaults silently win. |
| FE-ARCH-01 | MEDIUM | No ESLint config file exists although `eslint` is a devDependency and `"lint"` calls it — lint gate is broken. |
| MOB-ARCH-01 | HIGH | `AppConstants.apiBaseUrl` hardcoded `http://localhost:8080/api/v1`, no flavor/env switch — release build cannot reach any server. |
| REPO-01 | LOW | Nested `stock-opname-app/` duplicate tree with its own `.git` — confusion risk, doubles grep surface. |

## Assessment
Architecture **intent** is sound and modern (clean layers, svelte-query, Riverpod, sqflite queue). Execution has systemic gaps: process-level shutdown, config drift, broken lint gate, and an un-releasable mobile base URL. Score contribution for Backend Architecture and Frontend Architecture is capped by these.
