// Ikon garis ala SF Symbols untuk navigasi Apple. Setiap item menu dipetakan ke
// path SVG 24x24 (stroke). Dipisah dari menu.ts agar menu.ts tetap sumber navigasi.

export const ICONS: Record<string, string> = {
	// Dashboard
	'/':
		'M3 12l9-9 9 9M5 10v10h14V10M9 20v-6h6v6',
	// Scan
	'/scan':
		'M3 7V5a2 2 0 012-2h2M17 3h2a2 2 0 012 2v2M21 17v2a2 2 0 01-2 2h-2M7 21H5a2 2 0 01-2-2v-2M7 12h10',
	// Sessions
	'/sessions':
		'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4',
	// Products
	'/products':
		'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
	// Warehouses
	'/warehouses':
		'M3 21V8l9-5 9 5v13M9 21v-6h6v6M3 12h18',
	// Reports
	'/reports':
		'M9 17v-6M13 17V7M17 17v-3M4 21h16M4 3v18',
	// Stock ledger
	'/stock-ledger':
		'M4 5a2 2 0 012-2h12a2 2 0 012 2v14a2 2 0 01-2 2H6a2 2 0 01-2-2V5zM8 3v18M12 8h6M12 12h6M12 16h4',
	// Inventory
	'/inventory':
		'M12 3l9 5v8l-9 5-9-5V8l9-5zM3 8l9 5 9-5M12 13v8',
	// Inventory movements list
	'/inventory/movements':
		'M4 6h16M4 12h16M4 18h10',
	// Inventory products (same as products)
	'/inventory/products':
		'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
	// Stock Opname read-only products
	'/opname/products':
		'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4',
	// Stock Opname read-only warehouses
	'/opname/warehouses':
		'M3 21V8l9-5 9 5v13M9 21v-6h6v6M3 12h18',
	// Inventory grid entry (tabel)
	'/inventory/entry':
		'M3 3h18v18H3V3zM3 9h18M3 15h18M9 3v18M15 3v18',
	// Inventory settings (sliders)
	'/inventory/settings':
		'M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6',
	// Reconciliation
	'/reconciliation':
		'M4 4h6v6H4V4zM14 14h6v6h-6v-6zM14 4h6v6h-6V4zM4 14h6v6H4v-6z',
	// Snapshots
	'/snapshots':
		'M4 8V6a2 2 0 012-2h2M16 4h2a2 2 0 012 2v2M20 16v2a2 2 0 01-2 2h-2M8 20H6a2 2 0 01-2-2v-2M12 8v4l3 2',
	// Transfer
	'/transfer':
		'M4 7h13m0 0l-4-4m4 4l-4 4M20 17H7m0 0l4 4m-4-4l4-4',
	// Activity
	'/activity':
		'M12 8v4l3 2m-3-9a9 9 0 100 18 9 9 0 000-18z',
	// Backup
	'/backup':
		'M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2',
	// Settings
	'/settings':
		'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
	// Inventory app icon (selector)
	'app-inventory':
		'M3 7l9-4 9 4v10l-9 4-9-4V7zM3 7l9 4 9-4M12 11v10',
	// Stock opname app icon (selector)
	'app-opname':
		'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4'
};

/** Path SVG untuk href menu/aplikasi; fallback ikon kotak. */
export function iconFor(href: string): string {
	return ICONS[href] ?? ICONS['/products'];
}
