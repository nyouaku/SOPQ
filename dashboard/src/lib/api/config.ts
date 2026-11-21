// ===== FILE: dashboard/src/lib/api/config.ts =====
// Same-origin: dashboard dan API lewat nginx di origin yang sama. Tidak ada
// build arg VITE_*; untuk dev lokal, Vite proxy (vite.config.ts) meneruskan /api.
export const API_BASE_URL = '/api/v1';
export const WS_BASE_URL = `${typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss' : 'ws'}://${typeof window !== 'undefined' ? window.location.host : 'localhost'}/api/v1`;

export const API_ENDPOINTS = {
	base: API_BASE_URL,

	// Auth
	login: '/auth/login',
	logout: '/auth/logout',
	refresh: '/auth/refresh',
	profile: '/auth/me',

	// Products
	products: '/products',
	productByCode: (code: string) => `/products/code/${code}`,

	// Suppliers
	suppliers: '/suppliers',

	// Warehouses
	warehouses: '/warehouses',
	warehouseLocations: (id: string) => `/warehouses/${id}/locations`,

	// Stock Opname
	stockOpnameSessions: '/stock-opname/sessions',
	stockOpnameSession: (id: string) => `/stock-opname/sessions/${id}`,
	stockOpnameSessionDetails: (id: string) => `/stock-opname/sessions/${id}/details`,
	stockOpnameApprove: (id: string) => `/stock-opname/sessions/${id}/approve`,

	// Dashboard
	dashboardSummary: '/dashboard/summary',

	// Reports
	reportsStockOpname: '/reports/stock-opname',
	reportsStockOpnameExport: '/reports/stock-opname/export',

	// Activity Logs
	activityLogs: '/activity-logs',

	// Import (produk via template Excel)
	importProductTemplate: '/import/products/template',
	importProducts: '/import/products',

	// Users / roles (manajemen staff)
	users: '/auth/users',
	roles: '/auth/roles',
	register: '/auth/register',

	// Notifications
	notifications: '/notifications',

	// WebSocket
	stockOpnameWS: (sessionId: string) => `/ws/stock-opname/sessions/${sessionId}`,

	// Stock Ledger (kartu stok)
	stockLedgerProduct: (productId: string, warehouseId: string) =>
		`/stock-ledger/product/${productId}/${warehouseId}`,
	stockLedgerSummary: '/stock-ledger/summary',
	stockLedgerWarehouse: (warehouseId: string) =>
		`/stock-ledger/warehouse/${warehouseId}`,
	stockLedgerRange: '/stock-ledger/range',

	// Backup
	backupStatus: '/backup/status',
	backupDownload: '/backup/download',

	// Session lifecycle
	sessionReject: (id: string) => `/stock-opname/sessions/${id}/reject`,
	sessionReopen: (id: string) => `/stock-opname/sessions/${id}/reopen`,

	// Inventory movement (Stock IN/OUT/transfer antar gudang)
	inventory: '/inventory',
	inventoryById: (id: string) => `/inventory/${id}`,
	inventoryCancel: (id: string) => `/inventory/${id}/cancel`,
	inventoryIn: '/inventory/in',
	inventoryOut: '/inventory/out',
	inventoryTransfer: '/inventory/transfer',
	inventoryGrid: '/inventory/grid',
	inventoryGridCopy: '/inventory/grid/copy',
	inventorySummary: '/inventory/summary',
};
// ===== END FILE =====
