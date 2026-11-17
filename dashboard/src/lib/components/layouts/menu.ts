// Satu sumber menu navigasi untuk semua layout.
//
// DUA PARTISI:
// 1. APP — setiap item milik Stock Opname ATAU Inventory Management (atau 'shared').
//    Sidebar Apple hanya menampilkan item milik aplikasi aktif.
// 2. ACCESS — 'view' = lihat-saja (data dibaca, tak ada tombol ubah);
//    'full' = kendali penuh (tambah/ubah/hapus).
//
//   Stock Opname = audit fisik berkala -> mayoritas 'view' (cukup melihat, tak
//   mengubah master produk/gudang).
//   Inventory Management = operasional harian -> 'full' (mengubah produk, gudang, stok).

export interface MenuItem {
	href: string;
	label: string;
	/** Unicode emoji dipertahankan agar tidak ada perubahan tampilan. */
	icon: string;
	/** Shortcut keyboard untuk layout command center. */
	shortcut: string;
	/** Role yang boleh melihat menu. Kosong/undefined = semua role terautentikasi. */
	roles?: string[];
	/** Aplikasi pemilik item. Default 'opname'. */
	app?: AppId;
	/** Tingkat akses di aplikasi pemiliknya. Default 'full'. */
	access?: AccessLevel;
}

export type AppId = 'opname' | 'inventory' | 'shared';

/** Prefix URL milik aplikasi Inventory. Dipakai semua layout utk deteksi konteks. */
export const INVENTORY_ROOTS = ['/inventory', '/products', '/warehouses', '/transfer'];

/** Apakah path termasuk aplikasi Inventory Management. */
export function isInventoryPath(path: string): boolean {
	return INVENTORY_ROOTS.some((r) => path === r || path.startsWith(r + '/'));
}

/** Aplikasi aktif untuk sebuah path ('inventory' | 'opname'). */
export function activeAppFor(path: string): AppId {
	return isInventoryPath(path) ? 'inventory' : 'opname';
}
export type AccessLevel = 'view' | 'full';

const MANAGER = ['admin', 'supervisor'];

export const MENU_ITEMS: MenuItem[] = [
	// ===== STOCK OPNAME (audit fisik — mayoritas lihat-saja) =====
	{ href: '/', label: 'Dashboard', icon: '📊', shortcut: 'D', app: 'opname', access: 'full' },
	{ href: '/scan', label: 'Scan', icon: '📸', shortcut: 'S', app: 'opname', access: 'full' },
	{ href: '/sessions', label: 'Sessions', icon: '📋', shortcut: 'E', app: 'opname', access: 'full' },
	{ href: '/reports', label: 'Reports', icon: '📈', shortcut: 'R', roles: MANAGER, app: 'opname', access: 'view' },
	{ href: '/reconciliation', label: 'Rekonsiliasi', icon: '🔄', shortcut: 'N', roles: MANAGER, app: 'opname', access: 'view' },
	{ href: '/snapshots', label: 'Snapshots', icon: '📸', shortcut: 'H', roles: MANAGER, app: 'opname', access: 'view' },
	{ href: '/stock-ledger', label: 'Stock Ledger', icon: '📒', shortcut: 'L', roles: MANAGER, app: 'opname', access: 'view' },
	// Products & Gudang di SO = LIHAT-SAJA (tak bisa tambah/ubah/hapus).
	{ href: '/opname/products', label: 'Products', icon: '📦', shortcut: 'P', app: 'opname', access: 'view' },
	{ href: '/opname/warehouses', label: 'Gudang', icon: '🏭', shortcut: 'G', roles: MANAGER, app: 'opname', access: 'view' },
	{ href: '/activity', label: 'Activity', icon: '🧾', shortcut: 'A', roles: MANAGER, app: 'opname', access: 'view' },

	// ===== SHARED (kedua aplikasi) =====
	{ href: '/settings', label: 'Settings', icon: '⚙️', shortcut: 'T', app: 'shared', access: 'full' },
	{ href: '/backup', label: 'Backup', icon: '💾', shortcut: 'B', roles: ['admin'], app: 'shared', access: 'full' },

	// ===== INVENTORY MANAGEMENT (operasional — kendali penuh) =====
	{ href: '/inventory', label: 'Dashboard', icon: '📊', shortcut: 'I', app: 'inventory', access: 'full' },
	{ href: '/inventory/movements', label: 'Pergerakan', icon: '📥', shortcut: 'M', roles: MANAGER, app: 'inventory', access: 'full' },
	// Products & Gudang: KENDALI PENUH di Inventory (tambah/ubah/hapus).
	// SO hanya punya versi lihat-saja di atas.
	{ href: '/products', label: 'Products', icon: '📦', shortcut: 'P', app: 'inventory', access: 'full' },
	{ href: '/transfer', label: 'Transfer', icon: '🔁', shortcut: 'F', roles: MANAGER, app: 'inventory', access: 'full' },
	{ href: '/warehouses', label: 'Gudang', icon: '🏭', shortcut: 'G', roles: MANAGER, app: 'inventory', access: 'full' }
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
