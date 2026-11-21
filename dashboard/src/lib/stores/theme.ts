import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type ThemeLayout = 'apple' | 'focus' | 'workspace' | 'retro';
export type ThemeColor =
	| 'apple-light'
	| 'apple-dark'
	| 'rack-room'
	| 'noc-console'
	| 'daylight-ops'
	| 'saweria'
	| 'modern-list';
export type ThemeMode = 'manual' | 'auto';

export interface Theme {
	layout: ThemeLayout;
	color: ThemeColor;
	mode: ThemeMode;
	// Untuk auto mode: warna saat siang & malam
	dayColor?: ThemeColor;
	nightColor?: ThemeColor;
	// Jam mulai malam (0-23), default 18
	nightStartHour?: number;
	// Jam mulai siang (0-23), default 6
	dayStartHour?: number;
}

const THEME_KEY = 'stock-opname-theme';

const defaultTheme: Theme = {
	layout: 'apple',
	color: 'apple-dark',
	mode: 'manual',
	nightStartHour: 18,
	dayStartHour: 6
};

function getHourBasedColor(theme: Theme): ThemeColor {
	const hour = new Date().getHours();
	const nightStart = theme.nightStartHour ?? 18;
	const dayStart = theme.dayStartHour ?? 6;

	// Jika nightStart > dayStart (misal 18 > 6), malam = hour >= nightStart || hour < dayStart
	if (nightStart > dayStart) {
		return hour >= nightStart || hour < dayStart ? (theme.nightColor ?? 'apple-dark') : (theme.dayColor ?? 'apple-light');
	}
	// Kasus edge: nightStart < dayStart (misal 0 > 12 tidak masuk akal, tapi handle)
	return hour >= nightStart && hour < dayStart ? (theme.nightColor ?? 'apple-dark') : (theme.dayColor ?? 'apple-light');
}

function createThemeStore() {
	// Load theme from localStorage or use default
	const initialTheme = browser
		? (() => {
				try {
					const saved = localStorage.getItem(THEME_KEY);
					const theme = saved ? { ...defaultTheme, ...JSON.parse(saved) } : defaultTheme;
					// Jika auto mode, resolve warna berdasarkan jam
					if (theme.mode === 'auto') {
						theme.color = getHourBasedColor(theme);
					}
					return theme;
				} catch {
					return defaultTheme;
				}
		  })()
		: defaultTheme;

	const { subscribe, set, update } = writable<Theme>(initialTheme);

	// Interval untuk auto mode (check tiap menit)
	let autoInterval: ReturnType<typeof setInterval> | null = null;

	function startAutoInterval() {
		if (!browser || autoInterval) return;
		autoInterval = setInterval(() => {
			update((theme) => {
				if (theme.mode !== 'auto') return theme;
				const newColor = getHourBasedColor(theme);
				if (newColor !== theme.color) {
					const newTheme = { ...theme, color: newColor };
					applyTheme(newTheme);
					return newTheme;
				}
				return theme;
			});
		}, 60000); // Check tiap 1 menit
	}

	function stopAutoInterval() {
		if (autoInterval) {
			clearInterval(autoInterval);
			autoInterval = null;
		}
	}

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
		setMode: (mode: ThemeMode) => {
			update((theme) => {
				const newTheme = { ...theme, mode };
				if (mode === 'auto') {
					newTheme.color = getHourBasedColor(newTheme);
					startAutoInterval();
				} else {
					stopAutoInterval();
				}
				if (browser) {
					localStorage.setItem(THEME_KEY, JSON.stringify(newTheme));
					applyTheme(newTheme);
				}
				return newTheme;
			});
		},
		setDayNightColors: (dayColor: ThemeColor, nightColor: ThemeColor) => {
			update((theme) => {
				const newTheme = { ...theme, dayColor, nightColor };
				if (theme.mode === 'auto') {
					newTheme.color = getHourBasedColor(newTheme);
				}
				if (browser) {
					localStorage.setItem(THEME_KEY, JSON.stringify(newTheme));
					applyTheme(newTheme);
				}
				return newTheme;
			});
		},
		setNightHours: (nightStartHour: number, dayStartHour: number) => {
			update((theme) => {
				const newTheme = { ...theme, nightStartHour, dayStartHour };
				if (theme.mode === 'auto') {
					newTheme.color = getHourBasedColor(newTheme);
				}
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
					const theme = saved ? { ...defaultTheme, ...JSON.parse(saved) } : defaultTheme;
					if (theme.mode === 'auto') {
						theme.color = getHourBasedColor(theme);
						startAutoInterval();
					}
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

// Layout yang tersedia. 5 layout berikutnya (Retro, Neobrutalism, Acrylic,
// Neumorphism, Editorial) mengantre — mulai hanya setelah prompt khusus user.
export const THEME_LAYOUTS = [
	{
		id: 'apple' as ThemeLayout,
		name: 'Apple',
		description: 'Glass sidebar, Apple-inspired premium',
		icon: '🍎'
	},
	{
		id: 'focus' as ThemeLayout,
		name: 'Focus',
		description: 'Canvas mode tanpa sidebar — fokus penuh pada kerja',
		icon: '🎯'
	},
	{
		id: 'workspace' as ThemeLayout,
		name: 'Workspace',
		description: 'Rail ikon + panel konteks untuk operasi harian',
		icon: '🗂️'
	},
	{
		id: 'retro' as ThemeLayout,
		name: 'Retro UI',
		description: 'Vintage computer UI — beige, pixel icon, window panel',
		icon: '🖥️'
	}
] as const;

export const THEME_COLORS = [
	{
		id: 'apple-light' as ThemeColor,
		name: 'Apple Light',
		description: 'Apple-inspired light (#F5F5F7)',
		preview: { bg: '#F5F5F7', primary: '#007AFF', text: '#1D1D1F' }
	},
	{
		id: 'apple-dark' as ThemeColor,
		name: 'Apple Dark',
		description: 'Apple-inspired dark (#151516)',
		preview: { bg: '#151516', primary: '#0A84FF', text: '#F5F5F7' }
	},
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
