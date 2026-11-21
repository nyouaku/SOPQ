// Ikon Editorial — Lucide-style outline, stroke 1.6, hitang/charcoal, tanpa warna.
export const EDITORIAL_ICONS: Record<string, string> = {
	dashboard: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
	overview: 'M3 3h18v18H3zM3 9h18M9 21V9',
	chart: 'M3 17l6-6 4 4 8-8M21 7v6h-6',
	report: 'M6 2h9l4 4v16H6zM15 2v4h4M9 13h7M9 17h5',
	article: 'M4 3h16v18H4zM8 7h8M8 11h8M8 15h5',
	database: 'M12 2c4 0 6 1.5 6 3v14c0 1.5-2 3-6 3s-6-1.5-6-3V5c0-1.5 2-3 6-3zM6 5c0 1.5 2 3 6 3s6-1.5 6-3M6 12c0 1.5 2 3 6 3s6-1.5 6-3',
	users: 'M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM22 21v-2a4 4 0 00-3-3.8M16 3a4 4 0 010 7.8',
	calendar: 'M4 5h16v15H4zM4 9h16M8 3v4M16 3v4',
	search: 'M11 3a8 8 0 105 14l6 6-3 3-6-6A8 8 0 0011 3zM11 7a4 4 0 100 8 4 4 0 000-8z',
	settings: 'M12 9a3 3 0 100 6 3 3 0 000-6zM19 12l2-1-2-3-2 1-2-1-1-2h-4l-1 2-2 1-2-1-2 3 2 1v2l-2 1 2 3 2-1 2 1 1 2h4l1-2 2-1 2 1 2-3-2-1z',
	bell: 'M18 8a6 6 0 00-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4',
	profile: 'M12 12a4 4 0 100-8 4 4 0 000 8zM4 21a8 8 0 0116 0',
	print: 'M6 3h12v5H6zM4 8h16v8h-4v5H8v-5H4zM16 11h2v2h-2',
	export: 'M12 3v12m0 0l-4-4m4 4l4-4M4 19h16',
	menu: 'M4 6h16M4 12h16M4 18h16',
	close: 'M6 6l12 12M18 6L6 18'
};

export function editorialIconFor(href: string): string {
	if (href === '/') return EDITORIAL_ICONS.dashboard;
	if (href.includes('settings') || href.includes('backup')) return EDITORIAL_ICONS.settings;
	if (href.includes('report') || href.includes('activity')) return EDITORIAL_ICONS.chart;
	if (href.includes('warehouse') || href.includes('reconcil')) return EDITORIAL_ICONS.database;
	if (href.includes('session') || href.includes('history') || href.includes('compare')) return EDITORIAL_ICONS.calendar;
	if (href.includes('scan')) return EDITORIAL_ICONS.search;
	if (href.includes('ledger') || href.includes('approv') || href.includes('product')) return EDITORIAL_ICONS.article;
	if (href.includes('transfer')) return EDITORIAL_ICONS.export;
	return EDITORIAL_ICONS.dashboard;
}
