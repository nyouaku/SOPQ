# 03 — Frontend Audit (SvelteKit Dashboard)

Stack: SvelteKit 2, Svelte 4, TypeScript strict, Tailwind 3 + DaisyUI 4, `@tanstack/svelte-query` 5.90.2, Chart.js 4, `adapter-node`.
9 routes, 1 shared menu source, 3 layout shells, PWA service worker.

## Build / quality gates — verified this session
| Command | Result |
|---|---|
| `npx svelte-check --tsconfig ./tsconfig.json` | **PASS** — 0 errors, 1 warning (a11y, `sessions/+page.svelte:152`) |
| `npm run build` | **PASS** — adapter-node output, service-worker + manifest emitted |
| `npx eslint .` | **FAIL** — `ESLint couldn't find a configuration file.` |
| `npm run lint` | **FAIL** — prettier reports 35 files needing format, then eslint aborts |

The `lint` script is referenced in `package.json` but no `.eslintrc*`/`eslint.config.*` exists. `eslint` is installed and `@types/eslint` declared — someone wired the script and never added the config.

## Findings

### FE-01 — CRITICAL — `createQuery` options read once ⇒ filters never refetch
**Status: FIXED this session.**

`@tanstack/svelte-query`'s `createBaseQuery` does `const optionsStore = isSvelteStore(options) ? options : readable(options)` — a plain object is wrapped and read **once**, so `queryKey`/`queryFn` were frozen at mount. Changing a filter updated the URLSearchParams object but never the cache key.

Fix in `src/lib/api/queries.ts`: param factories accept `URLSearchParams | Readable<URLSearchParams>` and build options inside `derived(...)`:
```ts
function paramsStore(params: ParamsInput): Readable<URLSearchParams> { ... }
export const activityLogsQuery = (params: ParamsInput) =>
	createQuery(derived(paramsStore(params), ($params) => ({
		queryKey: queryKeys.activityLogs($params.toString()),
		queryFn: () => fetchActivityLogs($params)
	})));
```
Callers (`/reports`, `/activity`) now hold `writable(new URLSearchParams(...))` and call `.set(...)`.

**Residual caveat:** the same pattern applies to `locationsQuery(warehouseId)` / `boxesQuery(locationId)` — those still take a plain string captured once. They are only invoked after a warehouse/location selection exists in current call sites, so they work today; any future "change warehouse without remount" flow reintroduces the bug.

---

### FE-02 — HIGH — Non-reactive `activeFilters` in Reports
`src/routes/reports/+page.svelte:44`
```ts
$: activeFilters = [...get(applied).keys()].some((k) => k !== 'page' && k !== 'limit');
```
`applied` is a `const` writable. Svelte's compiler cannot see through `get()` to `applied`, so this reactive statement runs once. `activeFilters` stays at its initial value (`false`) forever → the "active filter" badge/hint never appears.

**Fix:** `$: activeFilters = [...$applied.keys()].some(...)` — `$applied` is auto-subscribed and tracked.

---

### FE-03 — HIGH — `/reports` visible to roles the backend rejects
`src/lib/components/layouts/menu.ts` exports a single `MENU_ITEMS` array with **no role field**. The docstring even says "Role filter: hanya menyembunyikan menu" — but no filtering code exists. A `staff_gudang` user sees Reports and Activity, clicks, and gets `403` from `RequireRole("supervisor","admin")`.

**Fix:** add `roles?: string[]` to `MenuItem`, filter in each shell against `$authStore.user.role`. Backend stays the real gate.

---

### FE-04 — HIGH — Three call sites bypass `apiClient` → no refresh-on-401
`apiClient.request()` refreshes once on 401, clears storage, and redirects. These do raw `fetch` with a manual token read:
- `src/routes/products/+page.svelte` — `handleExportExcel()`
- `src/routes/products/+page.svelte` — import POST
- `src/routes/reports/+page.svelte:97-110` — `exportFile()` (both CSV and Excel)

A user whose access token expired at 15 min gets a hard failure on every export/import instead of a transparent refresh. Two different error surfaces for the same class of failure.

**Fix:** add `apiClient.download(endpoint): Promise<Blob>` and `apiClient.upload(endpoint, form)` to the same class, reusing `request()`'s refresh logic.

---

### FE-05 — MEDIUM — Report summary tiles mix scopes
`reports/+page.svelte:39-42` sums `rows` (the current page, ≤25 items) while `total` in the paginator is the full result set. "Total Selisih" changes when you paginate. Non-obvious to the user, and impossible to reconcile with the label.

**Fix:** either move the aggregates into the API response (`{data,total,summary}`) or label them explicitly "halaman ini".

---

### FE-06 — MEDIUM — 4 `alert()` calls remain in session detail
`src/routes/sessions/[id]/+page.svelte:120,123,137,140` — blocking native dialogs in an otherwise polished DaisyUI app with a `toastStore` already available. Inconsistent UI, unstyleable, blocks the event loop.

---

### FE-07 — MEDIUM — `checkAuth()` trusts localStorage blindly
`src/lib/stores/auth.ts` — reads `localStorage['user']`, sets `isAuthenticated = true` without verifying the token is present or unexpired. A cleared/expired token leaves the UI in a logged-in shell until the first API call 401s. Low security impact (backend is the gate) but produces a flash of authenticated chrome.

**Fix:** require `access_token` presence; optionally hit `/auth/me` once on mount (mobile already does this).

---

### FE-08 — MEDIUM — Dead code and unused exports
- `queries.ts:160` `markNotificationRead()` — `createMutation` factory, exported, never imported.
- `src/routes/mobile` — `pendingCount` computed, never rendered.
- 24 `console.log/error/warn` remain across `src/`.
- `types` for rows are `any[]` throughout (`($report.data?.data as any[])`) despite TS strict — the API contract is untyped at every consumption site, so a backend field rename is a runtime bug, not a compile error.

---

### FE-09 — LOW — `$query.refetch()` in Activity ignores pending filter edits
`activity/+page.svelte:135` — error-state "Coba lagi" calls `$query.refetch()`, which re-runs `queryFn` with the **last applied** params rather than the current input values. Acceptable, but inconsistent with the apply-button mental model.

### FE-10 — LOW — Mobile route (`/mobile`) renders nothing for queued scans
The offline queue state exists (localStorage) but `pendingScans` is never rendered — user cannot see or resolve stuck items.

### FE-11 — LOW — `notificationsQuery` polls every 60 s with no dedupe of read-state
`refetchInterval: 60_000` on every page that mounts it; no `staleTime` — each poll is a full round trip. Fine at this scale.

---

## Frontend summary
The UI is genuinely well-built: consistent design tokens, three coherent layouts, single menu source, proper loading/error/empty branches on `/reports` and `/activity`, PWA, keyboard shortcuts. The blocking defects are one structural reactivity bug (fixed), one remaining non-reactive statement, a role-less menu, and a bypassed API client. Typing is nominal — strict mode on, `any` everywhere.
