import { beforeEach, describe, expect, it, vi } from 'vitest';

const get = vi.fn();
vi.mock('$lib/api/client', () => ({ apiClient: { get: (...a: unknown[]) => get(...a) } }));
vi.mock('$lib/api/config', () => ({
	API_ENDPOINTS: { reportsStockOpname: '/reports/stock-opname' }
}));

const { fetchStockOpnameReport } = await import('./queries');

describe('fetchStockOpnameReport', () => {
	beforeEach(() => get.mockReset());

	it('meneruskan query string dan mengembalikan summary dari server', async () => {
		const summary = { total_system: 10, total_physical: 8, total_difference: -2, mismatch_count: 1 };
		get.mockResolvedValue({
			success: true,
			data: { data: [], total: 1, page: 1, limit: 25, summary }
		});

		const params = new URLSearchParams('warehouse_id=w1&page=1&limit=25');
		const out = await fetchStockOpnameReport(params);

		expect(get).toHaveBeenCalledWith('/reports/stock-opname?warehouse_id=w1&page=1&limit=25');
		expect(out.summary).toEqual(summary);
	});

	it('melempar error bila respons gagal', async () => {
		get.mockResolvedValue({ success: false, error: { code: 'X', message: 'gagal' } });
		await expect(fetchStockOpnameReport(new URLSearchParams())).rejects.toBeDefined();
	});
});
