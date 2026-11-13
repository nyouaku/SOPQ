# 04 — Database Audit (PostgreSQL 16)

Schema source of truth: `backend/migrations/000001_init_schema.up.sql` (242 lines). `docs/DATABASE_SCHEMA.sql` is a separate, stale copy.

## Schema inventory
16 tables: `roles`, `users`, `sessions`, `warehouses`, `locations`, `boxes`, `categories`, `suppliers`, `products`, `stock`, `stock_opname`, `stock_opname_detail`, `devices`, `activity_logs`, `notifications`, `settings`.

Good design choices present:
- `stock_opname_detail.difference INT GENERATED ALWAYS AS (stock_physical - stock_system) STORED` — invariant enforced by the DB, not app code. Correct call.
- `stock UNIQUE(product_id, warehouse_id)` + `version INT` for optimistic locking.
- `products.barcode` / `products.product_code` UNIQUE + partial indexes `WHERE deleted_at IS NULL`.
- `activity_logs.entity_type/entity_id` indexed for audit traversal.
- `settings` table with seeded thresholds.

## Findings

### DB-01 — HIGH — `performance_indexes.sql` fails on first statement
`backend/migrations/performance_indexes.sql:5`
```sql
CREATE INDEX IF NOT EXISTS idx_stock_opname_status ON stock_opname(status) WHERE deleted_at IS NULL;
```
`stock_opname` has **no `deleted_at` column** (schema lines 138-150). Error: `column "deleted_at" does not exist`. Same class of bug at line 37:
```sql
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at) WHERE deleted_at IS NULL;
```
`sessions` also has no `deleted_at`. Under `docker-entrypoint-initdb.d` (which `docker-compose.yml` mounts `./backend/migrations` into) a mid-file failure aborts the remaining statements in that file depending on how psql is invoked.

### DB-02 — HIGH — `pg_trgm` used before it is created
Line 18 `CREATE INDEX ... USING gin(name gin_trgm_ops)` precedes line 40 `CREATE EXTENSION IF NOT EXISTS pg_trgm`. `gin_trgm_ops` requires the extension at *statement parse* time → `operator class "gin_trgm_ops" does not exist`.

### DB-03 — HIGH — File ends with statements that cannot run in a transaction
Lines 43-62: `ANALYZE`, `VACUUM ANALYZE`, and a `SELECT`. `VACUUM` cannot run inside a transaction block. Any migration runner that wraps each file in BEGIN/COMMIT (golang-migrate, most CI scripts) fails. Also `idx_stock_opname_detail_created_at ON stock_opname_detail(created_at DESC)` — `stock_opname_detail` has **no `created_at`**, only `scanned_at`.

Net effect: `performance_indexes.sql` is **broken end-to-end**; none of the intended performance indexes exist in any deployment that ran it as a migration.

### DB-04 — CRITICAL — `seed_user.sql` is invalid SQL against this schema
`backend/migrations/seed_user.sql`
| Line | Problem |
|---|---|
| 5 | `INSERT INTO users (username, password, full_name, email, ...)` — `users` has **no `email`** column |
| 8 | `'$2a$10$rZJ5WxV8FvH8W7iN5vY9JuKkH7nF3Qf3W5F0F0F0F0F0F0F0F0F0'` — placeholder/fake bcrypt hash, not a valid 60-char bcrypt digest for any password |
| 11 | `WHERE name = 'staff'` — seeded role is `staff_gudang` → `role_id` NULL → NOT NULL violation |
| 17 | `ON CONFLICT (username) DO UPDATE SET password = EXCLUDED.password` — column is `password_hash` |

The whole file aborts. Combined with the `docker-entrypoint-initdb.d` mount, a fresh `docker compose up` gets a partially-seeded DB.

### DB-05 — HIGH — `seed_users.sql` ships one shared password hash, committed
`backend/migrations/seed_users.sql:9-11` — three accounts (`admin`, `supervisor`, `staff`) with the **identical** bcrypt hash and the comment "Password untuk semua user: staff123". Role IDs are hardcoded `1/2/3` rather than subselected by name, so any `roles` insert-order change silently grants admin to the wrong account.

If this file runs in production (it is inside the mounted migrations dir), the system ships with a published admin credential.

### DB-06 — MEDIUM — `sessions` has no soft-delete yet callers assume one
`Session` domain struct has no `DeletedAt`; the schema too. Revocation is `revoked_at` — which is correct and sufficient. The problem is only that `performance_indexes.sql` and possibly other SQL assume `deleted_at`, signaling the two were written against different schema versions.

### DB-07 — MEDIUM — `stock_opname_detail.input_method` has no CHECK constraint
`VARCHAR(10) DEFAULT 'manual'` — the column comment says `manual, voice`, the handler validates `oneof=manual voice`, but the DB accepts anything ≤10 chars. A third writer (import, migration, direct SQL) can insert `camera`/`upload` and break report grouping. Note `VARCHAR(10)` also truncates—not errors—on some drivers.

**Fix:** `CHECK (input_method IN ('manual','voice',...))` matching the decided enum, widened to a sane length.

### DB-08 — MEDIUM — `users.updated_at` never maintained
No trigger, and the app's `Save()`-based `Update` includes it, but `UPDATE users SET ...` from any other path leaves it stale. Same for `warehouses`, `suppliers`, `products.updated_at` outside GORM. No `set_updated_at()` trigger anywhere.

### DB-09 — MEDIUM — Missing index on the hot report join path
Reports will join `stock_opname_detail(stock_opname_id)` → `stock_opname(warehouse_id, created_at)` → `products(category_id, supplier_id, brand)`. Current indexes cover the first hop and category; **`stock_opname_detail.scanned_at`** and **`stock_opname.status + created_at`** composite are absent. Not urgent while reports return zero rows, becomes urgent when BE-05 is implemented.

### DB-10 — LOW — `stock` has no FK to a warehouse-specific product scope
`stock(product_id, warehouse_id)` allows rows for products that logically live in one warehouse only. No constraint; application-level assumption (BE-17).

### DB-11 — LOW — `notifications.metadata` JSONB written via `fmt.Sprintf`
`notification_usecase.go:64` builds JSON by string formatting with an unescaped product name. Invalid JSON → insert error (JSONB validates) → notification silently lost. Use `json.Marshal`.

### DB-12 — LOW — Two divergent schema documents
`docs/DATABASE_SCHEMA.sql` (8.8 KB) vs `backend/migrations/000001_init_schema.up.sql` (8.0 KB). CLAUDE.md calls the former the "source of truth"; the code follows the latter. Whichever is authoritative, the other is a trap for the next developer.

---

## Migration strategy assessment
There is **no migration runner with version tracking**. `postgres.go::RunMigrations()` (per CLAUDE.md/boot flow) plus the `docker-entrypoint-initdb.d` mount means schema state depends on when the volume was first created. Consequences:
- Changing `000001_init_schema.up.sql` after first boot has **no effect** on existing volumes.
- No rollback path beyond manually running the `.down.sql`.
- No way to know which version a deployed DB is at.

**Recommendation:** adopt `golang-migrate` or `goose`, keep `000001` immutable, add monotonically numbered files, drop the `docker-entrypoint-initdb.d` mount, and run migrations as an explicit deploy step.
