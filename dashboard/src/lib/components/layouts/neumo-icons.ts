// Ikon Neumorphism — outline stroke 1.8, ruang lapang, muted gray / accent.
export const NEUMO_ICONS: Record<string, string> = {
	dashboard: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
	chart: 'M4 20V10M10 20V4M16 20v-8M22 20V7M2 20h20',
	users: 'M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM22 21v-2a4 4 0 00-3-3.8M16 3a4 4 0 010 7.8',
	package: 'M21 8l-9-5-9 5v8l9 5 9-5zM3 8l9 5 9-5M12 13v8',
	wallet: 'M20 7H5a2 2 0 010-4h13v4M20 7v12a2 2 0 01-2 2H5a2 2 0 01-2-2V7h17zM16 13h2',
	calendar: 'M4 5h16v15H4zM4 9h16M8 3v4M16 3v4M8 13h3v3H8z',
	report: 'M6 2h9l4 4v16H6zM15 2v4h4M9 13h7M9 17h5',
	search: 'M11 3a8 8 0 105 14l6 6-3 3-6-6A8 8 0 0011 3zM11 7a4 4 0 100 8 4 4 0 000-8z',
	settings: 'M12 8a4 4 0 100 8 4 4 0 000-8zM19.4 15a1.7 1.7 0 00.3 1.9l.1.1-2.8 2.8-.1-.1a1.7 1.7 0 00-1.9-.3 1.7 1.7 0 00-1 1.6V21h-4v-.1a1.7 1.7 0 00-1-1.6 1.7 1.7 0 00-1.9.3l-.1.1L4.2 17l.1-.1a1.7 1.7 0 00.3-1.9 1.7 1.7 0 00-1.6-1H3v-4h.1a1.7 1.7 0 001.6-1 1.7 1.7 0 00-.3-1.9L4.2 7 7 4.2l.1.1a1.7 1.7 0 001.9.3A1.7 1.7 0 0010 3V3h4v.1a1.7 1.7 0 001 1.6 1.7 1.7 0 001.9-.3l.1-.1L19.8 7l-.1.1a1.7 1.7 0 00-.3 1.9 1.7 1.7 0 001.6 1H21v4h-.1a1.7 1.7 0 00-1.5 1z',
	bell: 'M18 8a6 6 0 00-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4',
	profile: 'M12 12a4 4 0 100-8 4 4 0 000 8zM4 21a8 8 0 0116 0',
	plus: 'M12 5v14M5 12h14',
	edit: 'M12 20h9M4 20l4.5-1 11-11-3.5-3.5-11 11zM14.5 6.5l3 3',
	trash: 'M4 6h16M9 6V4h6v2M6 6l1 15h10l1-15M10 10v7M14 10v7',
	menu: 'M4 6h16M4 12h16M4 18h16',
	close: 'M6 6l12 12M18 6L6 18'
};

export function neumoIconFor(href: string): string {
	if (href === '/') return NEUMO_ICONS.dashboard;
	if (href.includes('settings') || href.includes('backup')) return NEUMO_ICONS.settings;
	if (href.includes('report') || href.includes('activity')) return NEUMO_ICONS.chart;
	if (href.includes('warehouse') || href.includes('reconcil')) return NEUMO_ICONS.package;
	if (href.includes('session') || href.includes('history') || href.includes('compare')) return NEUMO_ICONS.calendar;
	if (href.includes('scan')) return NEUMO_ICONS.search;
	if (href.includes('ledger') || href.includes('approv') || href.includes('product')) return NEUMO_ICONS.report;
	if (href.includes('transfer')) return NEUMO_ICONS.wallet;
	return NEUMO_ICONS.dashboard;
}
