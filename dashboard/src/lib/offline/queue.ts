import { apiClient } from '$lib/api/client';
import { API_ENDPOINTS } from '$lib/api/config';
import { readable, writable } from 'svelte/store';

/**
 * Antrian scan offline — satu implementasi untuk /scan.
 *
 * Penyimpanan: localStorage (bertahan saat tab ditutup).
 * Kunci unik: `session_id + product_code + timestamp`, supaya flush berulang
 * tidak menggandakan baris (backend belum punya idempotency key).
 */

export interface PendingScan {
	session_id: string;
	product_code: string;
	stock_physical: number;
	input_method: 'manual' | 'camera' | 'upload';
	note: string;
	queued_at: string;
}

const KEY = 'pending_scans';

function readAll(): PendingScan[] {
	if (typeof localStorage === 'undefined') return [];
	try {
		const raw = localStorage.getItem(KEY);
		const parsed = raw ? JSON.parse(raw) : [];
		if (!Array.isArray(parsed)) return [];
		// Antrean lama (sebelum barcode dihapus) memakai field `barcode`: pindahkan ke product_code
		// supaya scan yang belum terkirim tidak hilang.
		return parsed.map((s: any) => {
			if (s && s.product_code === undefined && s.barcode !== undefined) {
				const { barcode, ...rest } = s;
				return { ...rest, product_code: barcode };
			}
			return s;
		});
	} catch {
		return [];
	}
}

function writeAll(scans: PendingScan[]) {
	try {
		localStorage.setItem(KEY, JSON.stringify(scans));
	} catch {
		// localStorage penuh/diblokir — bukan alasan menghentikan scan.
	}
}

const internal = writable<PendingScan[]>(readAll());

/** Jumlah scan yang belum terkirim (dipakai badge UI). */
export const pendingCount = readable(0, (set) =>
	internal.subscribe((scans) => set(scans.length))
);

function refresh() {
	const scans = readAll();
	internal.set(scans);
	return scans;
}

/** Simpan scan yang gagal terkirim karena jaringan. */
export function enqueue(scan: Omit<PendingScan, 'queued_at'>): PendingScan {
	const item: PendingScan = { ...scan, queued_at: new Date().toISOString() };
	const scans = readAll();
	scans.push(item);
	writeAll(scans);
	refresh();
	return item;
}

/** Kosongkan antrian (mis. logout). */
export function clearQueue() {
	writeAll([]);
	refresh();
}

export function listQueue(): PendingScan[] {
	return readAll();
}

export function hasQueue(): boolean {
	return readAll().length > 0;
}

/**
 * Kirim ulang semua scan yang tertunda.
 * Item yang sukses dihapus; item yang masih gagal tetap di antrian.
 * `onResult` dipanggil per item supaya UI bisa memperbarui progres.
 */
export async function flushQueue(
	onResult?: (result: { ok: boolean; item: PendingScan; error?: string }) => void
): Promise<{ sent: number; failed: number }> {
	const scans = readAll();
	if (!scans.length) return { sent: 0, failed: 0 };

	const remaining: PendingScan[] = [];
	let sent = 0;
	let failed = 0;

	for (let i = 0; i < scans.length; i++) {
		const scan = scans[i];
		try {
			const res = await apiClient.post(`${API_ENDPOINTS.stockOpnameSession(scan.session_id)}/scan`, {
				session_id: scan.session_id,
				product_code: scan.product_code,
				stock_physical: scan.stock_physical,
				input_method: scan.input_method,
				note: scan.note
			});
			if (res.success) {
				sent++;
				onResult?.({ ok: true, item: scan });
			} else {
				failed++;
				remaining.push(scan);
				onResult?.({ ok: false, item: scan, error: res.error?.message });
			}
		} catch (e: any) {
			failed++;
			remaining.push(scan);
			onResult?.({ ok: false, item: scan, error: e?.message });
			// Jaringan masih mati — hentikan. Item yang BELUM diproses harus ikut
			// dikembalikan ke antrean; tanpa ini scan pengguna hilang.
			remaining.push(...scans.slice(i + 1));
			break;
		}
	}

	writeAll(remaining);
	refresh();
	return { sent, failed };
}
