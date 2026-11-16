// Satu sumber menu navigasi untuk semua layout.
// Role filter: hanya menyembunyikan menu; proteksi sebenarnya tetap di backend.
//
// PARTISI APLIKASI: setiap item milik salah satu aplikasi (Stock Opname ATAU
// Inventory Management) atau 'shared' (tampil di keduanya). Sidebar Apple hanya
// menampilkan item milik aplikasi yang sedang aktif — inilah pemisahan fitur.

export interface MenuItem {
	href: string;
	label: string;
	/** Unicode emoji dipertahankan agar tidak ada perubahan tampilan (Batch A). */
	icon: string;
	/** Shortcut keyboard untuk layout command center. */
	shortcut: string;
	/** Role yang boleh melihat menu. Kosong/undefined = semua role terautentikasi. */
	roles?: string[];
	/** Aplikasi pemilik item. Default 'opname'. */
	app?: AppId;
}

export type AppId = 'opname' | 'inventory' | 'shared';

const MANAGER = ['admin', 'supervisor'];

export const MENU_ITEMS: MenuItem[] = [
	// ===== STOCK OPNAME =====
	{ href: '/', label: 'Dashboard', icon: '📊', shortcut: 'D', app: 'opname' },
	{ href: '/scan', label: 'Scan', icon: '📸', shortcut: 'S', app: 'opname' },
	{ href: '/sessions', label: 'Sessions', icon: '📋', shortcut: 'E', app: 'opname' },
	{ href: '/reports', label: 'Reports', icon: '📈', shortcut: 'R', roles: MANAGER, app: 'opname' },
	{ href: '/reconciliation', label: 'Rekonsiliasi', icon: '🔄', shortcut: 'N', roles: MANAGER, app: 'opname' },
	{ href: '/snapshots', label: 'Snapshots', icon: '📸', shortcut: 'H', roles: MANAGER, app: 'opname' },
	{ href: '/stock-ledger', label: 'Stock Ledger', icon: '📒', shortcut: 'L', roles: MANAGER, app: 'opname' },
	{ href: '/opname/products', label: 'Products', icon: '📦', shortcut: 'P', app: 'opname' },
	{ href: '/activity', label: 'Activity', icon: '🧾', shortcut: 'A', roles: MANAGER, app: 'opname' },

	// ===== SHARED (kedua aplikasi) =====
	{ href: '/warehouses', label: 'Gudang', icon: '🏭', shortcut: 'G', roles: MANAGER, app: 'shared' },
	{ href: '/settings', label: 'Settings', icon: '⚙️', shortcut: 'T', app: 'shared' },
	{ href: '/backup', label: 'Backup', icon: '💾', shortcut: 'B', roles: ['admin'], app: 'shared' },

	// ===== INVENTORY MANAGEMENT =====
	// Products: kelola penuh (tambah/edit/hapus/import) milik Inventory.
	// Stock Opname hanya punya versi BACA-SAJA (route /opname/products).
	{ href: '/inventory', label: 'Dashboard', icon: '📊', shortcut: 'I', app: 'inventory' },
	{ href: '/inventory/movements', label: 'Pergerakan', icon: '📥', shortcut: 'M', roles: MANAGER, app: 'inventory' },
	{ href: '/inventory/products', label: 'Products', icon: '📦', shortcut: 'P', app: 'inventory' },
	{ href: '/transfer', label: 'Transfer', icon: '🔁', shortcut: 'F', roles: MANAGER, app: 'inventory' }
];

/** Menu yang boleh dilihat role tertentu. Hanya penyembunyian UI; proteksi nyata ada di backend. */
export function menuForRole(items: MenuItem[], roleName: string | undefined): MenuItem[] {
	return items.filter((item) => !item.roles || (roleName !== undefined && item.roles.includes(roleName)));
}

/**
 * Menu untuk satu aplikasi (sudah difilter role + aplikasi). Dipakai sidebar Apple
 * agar fitur Stock Opname dan Inventory Management benar-benar terpisah.
 */
export function menuForApp(items: MenuItem[], app: AppId, roleName: string | undefined): MenuItem[] {
	return menuForRole(items, roleName).filter((item) => (item.app ?? 'opname') === app || item.app === 'shared');
}
