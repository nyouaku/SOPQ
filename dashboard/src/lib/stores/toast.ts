import { writable } from 'svelte/store';

export interface Toast {
	id: number;
	message: string;
	type: 'success' | 'error' | 'info';
}

let seq = 0;
const { subscribe, update } = writable<Toast[]>([]);

/** Durasi default (ms) — notifikasi hilang otomatis. */
export const TOAST_DURATION = 3000;

function push(message: string, type: Toast['type'], duration = TOAST_DURATION) {
	const id = ++seq;
	update((all) => [...all, { id, message, type }]);
	setTimeout(() => dismiss(id), duration);
	return id;
}

function dismiss(id: number) {
	update((all) => all.filter((t) => t.id !== id));
}

export const toastStore = {
	subscribe,
	dismiss,
	success: (message: string) => push(message, 'success'),
	error: (message: string) => push(message, 'error'),
	info: (message: string) => push(message, 'info')
};
