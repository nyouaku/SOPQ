// Tema komponen konten mengikuti layout aktif.
// Layout hanya men-set `data-ui` di <html>; stylesheet ui-components.css
// melakukan sisanya tanpa perlu mengubah 100+ pengguna kelas DaisyUI.
export type UiTheme = 'apple' | 'focus' | 'workspace' | 'retro' | 'neo' | 'acrylic' | 'neumo' | 'editorial';

const KEY = 'sopq-ui-theme';

export function applyUiTheme(theme: UiTheme) {
	if (typeof document === 'undefined') return;
	try {
		document.documentElement.dataset.ui = theme;
		localStorage.setItem(KEY, theme);
	} catch {
		/* localStorage diblokir — tema tetap aktif untuk sesi ini */
	}
}

export function restoreUiTheme(): UiTheme {
	if (typeof document === 'undefined') return 'apple';
	try {
		const saved = localStorage.getItem(KEY) as UiTheme | null;
		const allowed: UiTheme[] = ['apple', 'focus', 'workspace', 'retro', 'neo', 'acrylic', 'neumo', 'editorial'];
		const theme = saved && allowed.includes(saved) ? saved : 'apple';
		document.documentElement.dataset.ui = theme;
		return theme;
	} catch {
		document.documentElement.dataset.ui = 'apple';
		return 'apple';
	}
}
