# Stock Opname Dashboard

SvelteKit web dashboard untuk admin & supervisor Stock Opname system.

## Features

- 🔐 JWT Authentication untuk admin/supervisor
- 📊 Dashboard dengan kartu ringkasan & Chart.js
- 📦 Master data management (Products, Users, Warehouses)
- 📈 Report dengan filter lengkap & export CSV
- 🌓 Dark theme dengan TailwindCSS + DaisyUI
- 📱 Responsive design

## Tech Stack

- SvelteKit + TypeScript
- TailwindCSS + DaisyUI
- Chart.js untuk grafik
- TanStack Query (svelte-query) untuk data fetching

## Setup

1. Install dependencies:
```bash
npm install
```

2. Create `.env` file:
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_BASE_URL=ws://localhost:8080/api/v1
```

3. Run development server:
```bash
npm run dev
```

4. Build for production:
```bash
npm run build
npm run preview
```

## Folder Structure

```
src/
├── lib/
│   ├── api/          # API client & config
│   ├── stores/       # Svelte stores (auth)
│   ├── components/   # Reusable components
│   └── utils/        # Utility functions
└── routes/
    ├── +layout.svelte       # Root layout dengan QueryClientProvider
    ├── +page.svelte         # Dashboard page
    ├── login/+page.svelte   # Login page
    ├── products/+page.svelte # Products management
    └── reports/+page.svelte  # Reports & export
```

## Pages

### Dashboard (`/`)
- Summary cards: Total SKU, Progress Hari Ini, Selisih, Belum Dicek
- Progress chart (Doughnut chart dengan Chart.js)
- Aktivitas terbaru table
- **Role access:** Admin, Supervisor

### Products (`/products`)
- Product list dengan search & pagination
- CRUD operations (Admin only)
- **Role access:** Admin (full), Supervisor (read-only)

### Reports (`/reports`)
- Filter: Warehouse, Category, Brand, Supplier, Date Range, Difference Range
- Report table dengan hasil filter
- Export to CSV
- **Role access:** Admin, Supervisor

## Authentication

- Login dengan username & password
- JWT access token disimpan di localStorage
- Auto-refresh token pada 401 response
- Auto-redirect ke `/login` jika unauthorized

## API Integration

API client (`lib/api/client.ts`) sudah handle:
- Auto-inject Authorization header
- Auto-refresh token on 401
- Redirect ke login jika refresh gagal
- Request/response interceptor

## Development

```bash
# Run dev server
npm run dev

# Type check
npm run check

# Lint
npm run lint

# Format
npm run format

# Build
npm run build
```

## Production

Build dengan adapter-auto akan auto-detect platform (Node, Vercel, Netlify, dll).

Untuk deployment manual:
```bash
npm run build
node build
```

Atau dengan Docker (lihat `Dockerfile` di root folder).
