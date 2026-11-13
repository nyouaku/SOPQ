# Arsitektur — Stock Opname App

## Diagram Alur Sistem

```
┌─────────────┐      REST + WS       ┌──────────────────┐      GORM       ┌────────────┐
│ Flutter App │ ───────────────────▶ │  Golang API (Gin) │ ──────────────▶ │ PostgreSQL │
│ (Android)   │ ◀─────────────────── │  Clean Architecture│ ◀────────────── │            │
│ + SQLite    │                      └──────────┬────────┘                 └────────────┘
└─────────────┘                                 │ Cache/RateLimit
      ▲                                          ▼
      │ Offline queue                    ┌──────────────┐
      │ auto-sync                        │    Redis     │
      ▼                                  └──────────────┘
┌─────────────┐      REST + WS
│  SvelteKit  │ ───────────────────▶  (endpoint sama dengan mobile)
│  Dashboard  │
└─────────────┘
```

## Backend — Clean Architecture

```
delivery (Gin handler) → usecase (business logic) → domain (interfaces & entities)
                                                            ▲
repository (implementasi GORM) ────────────────────────────┘
```

- **domain/**: hanya struct entity + interface repository/usecase. Tidak ada dependency ke library luar (kecuali `time`, `uuid`).
- **usecase/**: implementasi logic, menerima interface repository via constructor injection, tidak tahu soal Gin/HTTP.
- **repository/**: implementasi interface domain, satu-satunya layer yang tahu soal GORM/SQL.
- **delivery/http**: handler Gin, hanya parsing request, panggil usecase, format response. Tidak ada business logic di sini.

## Alur Optimistic Locking Stock

1. Saat scan, backend baca `stock.quantity_system` + `stock.version` saat ini sebagai snapshot ke `stock_opname_detail`.
2. Update stock (bila diperlukan) menggunakan `WHERE version = :old_version`, lalu increment version.
3. Jika `RowsAffected == 0` → conflict, retry atau return error ke client agar re-scan.

## Alur Offline Sync (Mobile)

1. Setiap hasil scan disimpan dulu ke SQLite lokal dengan `sync_status = pending`.
2. Background worker cek konektivitas tiap interval / event `connectivity_plus`.
3. Saat online: kirim batch `pending` items ke `/stock-opname/sessions/:id/scan` (bulk atau one-by-one dengan idempotency key = local UUID).
4. Sukses → update `sync_status = synced`. Gagal → tetap `pending`, retry dengan backoff.
5. UI menampilkan badge status: **Online / Offline / Syncing / Synced**.

## Keputusan Desain Default (agar tidak perlu tanya berulang)

- Soft delete untuk seluruh tabel master (`deleted_at`), hard delete tidak digunakan.
- UUID sebagai primary key seluruh tabel transaksional & master (memudahkan sinkronisasi offline-first tanpa konflik ID).
- Snapshot `stock_system` disimpan di `stock_opname_detail` saat scan (bukan dihitung ulang saat laporan), agar laporan historis akurat meski stock sistem berubah setelahnya.
- Idempotency key (UUID lokal dari mobile) dipakai untuk mencegah duplikasi saat retry sync offline.
