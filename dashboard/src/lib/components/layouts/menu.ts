// Satu sumber menu navigasi untuk ketiga layout (sidebar, topnav, command).
// Role filter: hanya menyembunyikan menu; proteksi sebenarnya tetap di backend.

export interface MenuItem {
	href: string;
	label: string;
	/** Unicode emoji dipertahankan agar tidak ada perubahan tampilan (Batch A). */
	icon: string;
	/** Shortcut keyboard untuk layout command center. */
	shortcut: string;
	/** Role yang boleh melihat menu. Kosong/undefined = semua role terautentikasi. */
	roles?: string[];
}

const MANAGER = ['admin', 'supervisor'];

export const MENU_ITEMS: MenuItem[] = [
	{ href: '/', label: 'Dashboard', icon: '📊', shortcut: 'D' },
	{ href: '/scan', label: 'Scan', icon: '📸', shortcut: 'S' },
	{ href: '/sessions', label: 'Sessions', icon: '📋', shortcut: 'E' },
	{ href: '/products', label: 'Products', icon: '📦', shortcut: 'P' },
	{ href: '/warehouses', label: 'Gudang', icon: '🏭', shortcut: 'G', roles: MANAGER },
	{ href: '/reports', label: 'Reports', icon: '📈', shortcut: 'R', roles: MANAGER },
	{ href: '/stock-ledger', label: 'Stock Ledger', icon: '📒', shortcut: 'L', roles: MANAGER },
	{ href: '/reconciliation', label: 'Rekonsiliasi', icon: '🔄', shortcut: 'N', roles: MANAGER },
	{ href: '/snapshots', label: 'Snapshots', icon: '📸', shortcut: 'H', roles: MANAGER },
	{ href: '/transfer', label: 'Transfer', icon: '🔁', shortcut: 'F', roles: MANAGER },
	{ href: '/activity', label: 'Activity', icon: '🧾', shortcut: 'A', roles: MANAGER },
	{ href: '/backup', label: 'Backup', icon: '💾', shortcut: 'B', roles: ['admin'] },
	{ href: '/settings', label: 'Settings', icon: '⚙️', shortcut: 'T' }
];

/** Menu yang boleh dilihat role tertentu. Hanya penyembunyian UI; proteksi nyata ada di backend. */
export function menuForRole(items: MenuItem[], roleName: string | undefined): MenuItem[] {
	return items.filter((item) => !item.roles || (roleName !== undefined && item.roles.includes(roleName)));
}