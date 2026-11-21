// Ikon retro BERBASIS SVG — bukan emoji produksi.
// Semua path memakai grid 16x16 dengan sudut tajam (shape-rendering crispEdges)
// supaya terbaca sebagai pixel/bitmap art pada renderer apa pun.
export const RETRO_ICONS: Record<string, string> = {
	monitor: 'M2 2h12v9H2zM4 13h8M2 11h12M5 5h6M5 7h4', // 🖥️
	floppy: 'M2 2h9l3 3v9H2zM5 2v5h6V2M5 14v-4h6v4M11 2v3', // 💾
	folder: 'M1 4h5l2 2h7v8H1zM1 4V2h5v2', // 📁
	gear: 'M6 1h4l1 2 2 1 2-1 1 2-1 2 1 2 1 2-2 1-2-1-1 2H6l-1-2-2 1-2-1 1-2-1-2-1-2 2-1 2 1zM8 6a2 2 0 100 4 2 2 0 000-4z', // ⚙️
	search: 'M7 1a5 5 0 013 9l4 4-2 2-4-4A5 5 0 117 1zM7 3a3 3 0 100 6 3 3 0 000-6z', // 🔍
	user: 'M6 1h4v2h2v4h-2v2h2v6H4v-6h2V7H4V3h2zM6 3v4h4V3', // 👤
	calendar: 'M2 3h12v11H2zM2 6h12M5 1v4M11 1v4M4 8h2v2H4zM8 8h2v2H8z', // 📅
	chart: 'M2 13V2h2v9h2V6h2v5h2V4h2v9zM1 15h14', // 📊
	database: 'M8 1c4 0 6 1 6 3v8c0 2-2 3-6 3s-6-1-6-3V4c0-2 2-3 6-3zM2 4c0 2 2 3 6 3s6-1 6-3M2 8c0 2 2 3 6 3s6-1 6-3', // 🗄️
	bell: 'M8 1a3 3 0 013 3v4l2 3H3l2-3V4a3 3 0 013-3zM6 12h4v2H6z', // 🔔
	printer: 'M4 1h8v4H4zM2 5h12v7h-3v3H5v-3H2zM11 7h2v2h-2', // 🖨️
	trash: 'M5 2h6M3 4h10l-1 11H4zM6 6v7M10 6v7', // 🗑️
	close: 'M3 3l10 10M13 3L3 13',
	minimize: 'M3 8h10',
	maximize: 'M3 3h10v10H3z',
	network: 'M2 8h3l2-3h4l3 3M8 5V2M5 11h6M2 8v3h3',
	power: 'M8 1v6M4 3a5 5 0 108 0'
};

/** Pilih ikon retro sesuai rute; fallback = folder. */
export function retroIconFor(href: string): string {
	if (href === '/') return RETRO_ICONS.monitor;
	if (href.includes('settings')) return RETRO_ICONS.gear;
	if (href.includes('report') || href.includes('activity')) return RETRO_ICONS.chart;
	if (href.includes('backup') || href.includes('snapshot')) return RETRO_ICONS.floppy;
	if (href.includes('product') || href.includes('inventory')) return RETRO_ICONS.folder;
	if (href.includes('warehouse') || href.includes('reconcil')) return RETRO_ICONS.database;
	if (href.includes('session') || href.includes('history') || href.includes('compare')) return RETRO_ICONS.calendar;
	if (href.includes('scan')) return RETRO_ICONS.search;
	if (href.includes('ledger') || href.includes('approv')) return RETRO_ICONS.printer;
	return RETRO_ICONS.folder;
}
