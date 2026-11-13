import { beforeEach, describe, expect, it, vi } from 'vitest';

// apiClient di-mock: test ini memverifikasi perilaku antrean, bukan jaringan.
const post = vi.fn();
vi.mock('$lib/api/client', () => ({ apiClient: { post: (...a: unknown[]) => post(...a) } }));
vi.mock('$lib/api/config', () => ({
	API_ENDPOINTS: { stockOpnameSession: (id: string) => `/stock-opname/sessions/${id}` }
}));

const { enqueue, flushQueue, listQueue, clearQueue, hasQueue } = await import('./queue');

const scan = (product_code: string) => ({
	session_id: 's1',
	product_code,
	stock_physical: 5,
	input_method: 'manual' as const,
	note: ''
});

describe('offline queue', () => {
	beforeEach(() => {
		localStorage.clear();
		clearQueue();
		post.mockReset();
	});

	it('menyimpan scan ke localStorage', () => {
		enqueue(scan('A1'));
		enqueue(scan('A2'));
		expect(listQueue()).toHaveLength(2);
		expect(hasQueue()).toBe(true);
	});

	it('flush menghapus item yang sukses saja', async () => {
		enqueue(scan('OK'));
		enqueue(scan('FAIL'));
		post.mockImplementation(async (_url: string, body: { product_code: string }) =>
			body.product_code === 'OK' ? { success: true } : { success: false, error: { message: 'x' } }
		);

		const res = await flushQueue();
		expect(res).toEqual({ sent: 1, failed: 1 });
		expect(listQueue().map((s) => s.product_code)).toEqual(['FAIL']);
	});

	it('jaringan mati: sisa item tetap di antrean', async () => {
		enqueue(scan('A1'));
		enqueue(scan('A2'));
		post.mockRejectedValue(new Error('offline'));

		const res = await flushQueue();
		expect(res).toEqual({ sent: 0, failed: 1 });
		expect(listQueue()).toHaveLength(2);
	});
});
