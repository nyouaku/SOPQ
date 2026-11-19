/**
 * Tanggal/jam dalam zona waktu perangkat, bukan UTC.
 *
 * `new Date().toISOString()` selalu UTC: di WIB (UTC+7) pukul 00:00–06:59 tanggal 1,
 * string-nya masih bulan/tanggal sebelumnya. Filter bulan jadi salah default dan
 * nama file backup/export berlabel tanggal kemarin.
 */

/** YYYY-MM menurut waktu lokal. */
export function localMonth(d: Date = new Date()): string {
	return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

/** YYYY-MM-DD menurut waktu lokal. */
export function localDate(d: Date = new Date()): string {
	return `${localMonth(d)}-${String(d.getDate()).padStart(2, '0')}`;
}

/** YYYY-MM-DDTHH:mm:ss (lokal, tanpa offset) — untuk query `from`/`to`. */
export function localDateTime(d: Date = new Date()): string {
	const pad = (n: number) => String(n).padStart(2, '0');
	return `${localDate(d)}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

/** Bulan sebelumnya dalam format YYYY-MM. */
export function previousMonth(month: string): string {
	const [y, m] = month.split('-').map(Number);
	const prev = new Date(y, m - 2, 1);
	return localMonth(prev);
}
