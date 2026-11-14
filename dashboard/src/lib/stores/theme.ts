import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type ThemeLayout = 'sidebar' | 'topnav' | 'command' | 'bottom';
export type ThemeColor = 'rack-room' | 'noc-console' | 'daylight-ops' | 'saweria' | 'modern-list';

export interface Theme {
	layout: ThemeLayout;
	color: ThemeColor;
}

const THEME_KEY = 'stock-opname-theme';

const defaultTheme: Theme = {
	layout: 'command',
	color: 'saweria'
};

function createThemeStore() {
	// Load theme from localStorage or use default
	const initialTheme = browser
		? (() => {
				try {
					const saved = localStorage.getItem(THEME_KEY);
					return saved ? JSON.parse(saved) : defaultTheme;
				} catch {
					return defaultTheme;
				}
		  })()
		: defaultTheme;

	const { subscribe, set, update } = writable<Theme>(initialTheme);

	return {
		subscribe,
		setLayout: (layout: ThemeLayout) => {
			update((theme) => {
				const newTheme = { ...theme, layout };
				if (browser) {
					localStorage.setItem(THEME_KEY, JSON.stringify(newTheme));
					applyTheme(newTheme);
				}
				return newTheme;
			});
		},
		setColor: (color: ThemeColor) => {
			update((theme) => {
				const newTheme = { ...theme, color };
				if (browser) {
					localStorage.setItem(THEME_KEY, JSON.stringify(newTheme));
					applyTheme(newTheme);
				}
				return newTheme;
			});
		},
		setTheme: (theme: Theme) => {
			set(theme);
			if (browser) {
				localStorage.setItem(THEME_KEY, JSON.stringify(theme));
				applyTheme(theme);
			}
		},
		init: () => {
			if (browser) {
				try {
					const saved = localStorage.getItem(THEME_KEY);
					const theme = saved ? JSON.parse(saved) : defaultTheme;
					applyTheme(theme);
					set(theme);
				} catch {
					applyTheme(defaultTheme);
					set(defaultTheme);
				}
			}
		}
	};
}

function applyTheme(theme: Theme) {
	if (!browser) return;

	// Apply color theme via data-theme attribute
	document.documentElement.setAttribute('data-theme', theme.color);

	// Apply layout class
	document.documentElement.setAttribute('data-layout', theme.layout);
}

export const themeStore = createThemeStore();

// Theme metadata
export const THEME_LAYOUTS = [
	{
		id: 'sidebar' as ThemeLayout,
		name: 'Classic Sidebar',
		description: 'Traditional sidebar navigation',
		icon: '📋'
	},
	{
		id: 'topnav' as ThemeLayout,
		name: 'Modern Top Nav',
		description: 'Horizontal navigation bar',
		icon: '🎯'
	},
	{
		id: 'command' as ThemeLayout,
		name: 'Command Center',
		description: 'Split-panel ops layout',
		icon: '🖥️'
	},
	{
		id: 'bottom' as ThemeLayout,
		name: 'Bottom Bar',
		description: 'Bottom navigation, mobile-first modern',
		icon: '📱'
	}
] as const;

export const THEME_COLORS = [
	{
		id: 'rack-room' as ThemeColor,
		name: 'Rack Room',
		description: 'Dark NOC theme with teal',
		preview: { bg: '#14171C', primary: '#5EEAD4', text: '#C7CDD6' }
	},
	{
		id: 'noc-console' as ThemeColor,
		name: 'NOC Console',
		description: 'Terminal green theme',
		preview: { bg: '#0B0E0C', primary: '#00FF9C', text: '#00FF9C' }
	},
	{
		id: 'daylight-ops' as ThemeColor,
		name: 'Daylight Ops',
		description: 'Light theme for daylight',
		preview: { bg: '#F7F8FA', primary: '#1D4ED8', text: '#1C222B' }
	},
	{
		id: 'saweria' as ThemeColor,
		name: 'Saweria',
		description: 'Warm orange donation style',
		preview: { bg: '#FFFFFF', primary: '#FF6B35', text: '#2D3142' }
	},
	{
		id: 'modern-list' as ThemeColor,
		name: 'Modern List',
		description: 'Clean purple gradient',
		preview: { bg: '#F9FAFB', primary: '#667EEA', text: '#111827' }
	}
] as const;
