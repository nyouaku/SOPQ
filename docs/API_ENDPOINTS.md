# API Endpoints — Source of Truth

Base URL: `/api/v1`
Auth: `Authorization: Bearer <access_token>` kecuali disebut publik.
Format response standar:
```json
{ "success": true, "data": {}, "message": "" }
{ "success": false, "error": { "code": "STRING_CODE", "message": "human readable" } }
```

## Auth
| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| POST | `/auth/login` | publik | Login, return access + refresh token |
| POST | `/auth/refresh` | publik | Tukar refresh token → access token baru |
| POST | `/auth/logout` | semua | Revoke refresh token aktif |
| GET | `/auth/me` | semua | Profil user login |
| POST | `/auth/users` | admin | Buat user baru |

## Master Data
| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| GET/POST | `/products` | admin | List (paginated, filter) / create produk |
| GET/PUT/DELETE | `/products/:id` | admin | Detail/update/soft-delete produk |
| GET | `/products/barcode/:barcode` | semua | **Kritikal**: lookup cepat by barcode, < 500ms |
| POST | `/products/:id/photo` | admin | Upload foto produk |
| GET/POST | `/categories` | admin | Kategori |
| GET/POST | `/suppliers` | admin | Supplier |
| GET/POST | `/warehouses` | admin | Gudang |
| GET/POST | `/warehouses/:id/locations` | admin | Area dalam gudang |
| GET/POST | `/locations/:id/boxes` | admin | Rak/box dalam area |

## Stock Opname
| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| POST | `/stock-opname/sessions` | supervisor, staff | Mulai sesi opname (warehouse_id, location_id) |
| GET | `/stock-opname/sessions/:id` | semua | Detail + progress sesi |
| POST | `/stock-opname/sessions/:id/scan` | staff | Submit hasil scan 1 item (barcode, stock_physical, input_method) |
| POST | `/stock-opname/sessions/:id/complete` | supervisor, staff | Tandai sesi selesai |
| POST | `/stock-opname/sessions/:id/approve` | supervisor | Approve hasil opname |
| GET | `/stock-opname/sessions/:id/details` | semua | List detail item yang sudah discan |
| WS | `/ws/stock-opname/sessions/:id` | semua | Realtime progress broadcast |

## Dashboard & Reports
| Method | Endpoint | Role | Deskripsi |
|---|---|---|---|
| GET | `/dashboard/summary` | semua | Total SKU, opname hari ini, progress, selisih, belum dicek |
| GET | `/dashboard/user-progress` | supervisor, admin | Progress per staff |
| GET | `/reports/stock-opname` | supervisor, admin | Laporan dengan filter (gudang, kategori, merk, supplier, tanggal, user, selisih) |
| GET | `/reports/stock-opname/export?format=pdf\|xlsx\|csv` | supervisor, admin | Export laporan |
| GET | `/activity-logs` | supervisor, admin | Riwayat aktivitas (paginated, filter) |
| GET | `/notifications` | semua | Notifikasi user (paginated) |
| PATCH | `/notifications/:id/read` | semua | Tandai sudah dibaca |

## Konvensi
- Semua endpoint list mendukung `?page=&limit=&sort=`
- Semua field tanggal filter format ISO 8601, contoh `?date_from=2026-08-01&date_to=2026-08-07`
- Validasi request body wajib pakai `go-playground/validator`, response error `422` untuk validasi gagal
- Rate limit: `/auth/login` maksimal 5 request/menit/IP
