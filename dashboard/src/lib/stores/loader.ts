import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type PreloaderStyle = 'scan' | 'rings' | 'boxes' | 'orbit' | 'skeleton';
export type LoaderStyle = 'spinner' | 'dots' | 'bars' | 'ring' | 'barcode';

export interface LoaderSettings {
	preloader: PreloaderStyle;
	loader: LoaderStyle;
	/** Preloader hanya tampil sekali per sesi browser. */
	enabled: boolean;
}

export const PRELOADER_STYLES: PreloaderStyle[] = ['scan', 'rings', 'boxes', 'orbit', 'skeleton'];
export const LOADER_STYLES: LoaderStyle[] = ['spinner', 'dots', 'bars', 'ring', 'barcode'];

const KEY = 'stock-opname-loader';

const defaults: LoaderSettings = { preloader: 'scan', loader: 'spinner', enabled: true };

function load(): LoaderSettings {
	if (!browser) return defaults;
	try {
		const raw = localStorage.getItem(KEY);
		return raw ? { ...defaults, ...JSON.parse(raw) } : defaults;
	} catch {
		return defaults;
	}
}

const { subscribe, set, update } = writable<LoaderSettings>(load());

function persist(value: LoaderSettings) {
	if (browser) {
		try {
			localStorage.setItem(KEY, JSON.stringify(value));
		} catch {
			// localStorage bisa diblokir (mode privat) — abaikan.
		}
	}
}

export const loaderStore = {
	subscribe,
	setPreloader: (preloader: PreloaderStyle) =>
		update((s) => {
			const next = { ...s, preloader };
			persist(next);
			return next;
		}),
	setLoader: (loader: LoaderStyle) =>
		update((s) => {
			const next = { ...s, loader };
			persist(next);
			return next;
		}),
	setEnabled: (enabled: boolean) =>
		update((s) => {
			const next = { ...s, enabled };
			persist(next);
			return next;
		}),
	reset: () => {
		persist(defaults);
		set(defaults);
	}
};
