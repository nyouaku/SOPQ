// Ikon Neobrutalism — SVG stroke tebal (2.2), square cap, geometris.
// Bukan emoji. Semua path 24x24, konsisten satu ketebalan.
export const NEO_ICONS: Record<string, string> = {
	dashboard: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
	chart: 'M4 20V10M10 20V4M16 20v-8M22 20V7M2 20h20',
	users: 'M9 11a4 4 0 100-8 4 4 0 000 8zM2 21v-2a6 6 0 0112 0v2M17 8a3 3 0 010-6M17 21v-2a5 5 0 00-2-4',
	boxes: 'M3 7l9-4 9 4v10l-9 4-9-4zM3 7l9 4 9-4M12 11v10M7 4l9 4',
	receipt: 'M5 2h14v20l-3-2-4 2-4-2-3 2zM8 7h8M8 11h8M8 15h5',
	settings: 'M12 8a4 4 0 100 8 4 4 0 000-8zM12 2l2 3 3-1 2 3-2 2 1 3 3 2-2 3-3-1-2 3-2-1-2 1-2-3-3 1-2-3 3-2-1-3 3-2-2-2 2-3 3 1z',
	search: 'M10 2a8 8 0 105 14l6 6-3 3-6-6A8 8 0 0010 2zM10 6a4 4 0 100 8 4 4 0 000-8z',
	bell: 'M12 2a5 5 0 015 5v5l3 4H4l3-4V7a5 5 0 015-5zM9 19h6v2H9z',
	plus: 'M12 2v20M2 12h20',
	edit: 'M4 20l4-1 11-11-3-3L5 16zM14 6l3 3M3 21h18',
	trash: 'M4 5h16M8 5V2h8v3M6 5l1 16h10l1-16M10 9v8M14 9v8',
	filter: 'M3 6h18M6 12h12M10 18h4',
	download: 'M12 3v12m0 0l-5-5m5 5l5-5M4 20h16',
	menu: 'M3 6h18M3 12h18M3 18h18',
	close: 'M5 5l14 14M19 5L5 19'
};

export function neoIconFor(href: string): string {
	if (href === '/') return NEO_ICONS.dashboard;
	if (href.includes('settings')) return NEO_ICONS.settings;
	if (href.includes('report') || href.includes('activity')) return NEO_ICONS.chart;
	if (href.includes('backup') || href.includes('snapshot')) return NEO_ICONS.download;
	if (href.includes('warehouse') || href.includes('reconcil')) return NEO_ICONS.boxes;
	if (href.includes('session') || href.includes('history') || href.includes('compare')) return NEO_ICONS.receipt;
	if (href.includes('scan') || href.includes('search')) return NEO_ICONS.search;
	if (href.includes('ledger') || href.includes('approv') || href.includes('product')) return NEO_ICONS.boxes;
	if (href.includes('notification')) return NEO_ICONS.bell;
	return NEO_ICONS.dashboard;
}
