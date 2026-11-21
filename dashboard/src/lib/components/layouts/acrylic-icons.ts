// Ikon Acrylic — outline modern ala Lucide, stroke 1.7, konsisten.
export const ACRYLIC_ICONS: Record<string, string> = {
	dashboard: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
	overview: 'M3 3h18v18H3zM3 9h18M9 21V9',
	analytics: 'M3 3v18h18M7 15l4-5 3 3 5-7',
	files: 'M6 2h8l4 4v16H6zM14 2v4h4M9 12h6M9 16h6',
	users: 'M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2M9 11a4 4 0 100-8 4 4 0 000 8zM22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75',
	package: 'M21 8l-9-5-9 5v8l9 5 9-5zM3 8l9 5 9-5M12 13v8',
	wallet: 'M20 7H5a2 2 0 010-4h13v4M20 7v12a2 2 0 01-2 2H5a2 2 0 01-2-2V7h17zM16 13h2',
	chart: 'M4 20V10M10 20V4M16 20v-8M22 20V7M2 20h20',
	search: 'M11 3a8 8 0 105 14l6 6-3 3-6-6A8 8 0 0011 3zM11 7a4 4 0 100 8 4 4 0 000-8z',
	settings: 'M12 8a4 4 0 100 8 4 4 0 000-8zM19.4 15a1.7 1.7 0 00.3 1.9l.1.1-2.8 2.8-.1-.1a1.7 1.7 0 00-1.9-.3 1.7 1.7 0 00-1 1.6V21h-4v-.1a1.7 1.7 0 00-1-1.6 1.7 1.7 0 00-1.9.3l-.1.1L4.2 17l.1-.1a1.7 1.7 0 00.3-1.9 1.7 1.7 0 00-1.6-1H3v-4h.1a1.7 1.7 0 001.6-1 1.7 1.7 0 00-.3-1.9L4.2 7 7 4.2l.1.1a1.7 1.7 0 001.9.3A1.7 1.7 0 0010 3V3h4v.1a1.7 1.7 0 001 1.6 1.7 1.7 0 001.9-.3l.1-.1L19.8 7l-.1.1a1.7 1.7 0 00-.3 1.9 1.7 1.7 0 001.6 1H21v4h-.1a1.7 1.7 0 00-1.5 1z',
	bell: 'M18 8a6 6 0 00-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4',
	profile: 'M12 12a4 4 0 100-8 4 4 0 000 8zM4 21a8 8 0 0116 0',
	cloud: 'M18 10h-1.3A7 7 0 004 8.5a5 5 0 001.5 9.5H18a4 4 0 000-8z',
	shield: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10zM9 12l2 2 4-4',
	menu: 'M4 6h16M4 12h16M4 18h16',
	close: 'M6 6l12 12M18 6L6 18'
};

export function acrylicIconFor(href: string): string {
	if (href === '/') return ACRYLIC_ICONS.dashboard;
	if (href.includes('settings') || href.includes('backup')) return ACRYLIC_ICONS.settings;
	if (href.includes('report') || href.includes('activity')) return ACRYLIC_ICONS.chart;
	if (href.includes('warehouse') || href.includes('reconcil')) return ACRYLIC_ICONS.package;
	if (href.includes('session') || href.includes('history') || href.includes('compare')) return ACRYLIC_ICONS.overview;
	if (href.includes('scan') || href.includes('search')) return ACRYLIC_ICONS.search;
	if (href.includes('ledger') || href.includes('approv') || href.includes('product')) return ACRYLIC_ICONS.files;
	if (href.includes('transfer')) return ACRYLIC_ICONS.wallet;
	return ACRYLIC_ICONS.dashboard;
}
