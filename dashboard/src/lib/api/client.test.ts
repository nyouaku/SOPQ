import { beforeEach, describe, expect, it, vi } from 'vitest';

const goto = vi.fn();
vi.mock('$app/navigation', () => ({ goto: (...a: unknown[]) => goto(...a) }));
vi.mock('./config', () => ({ API_BASE_URL: '/api/v1' }));

const { apiClient } = await import('./client');

const json = (status: number, body: unknown) =>
	new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

describe('apiClient refresh', () => {
	beforeEach(() => {
		localStorage.clear();
		goto.mockReset();
		vi.restoreAllMocks();
	});

	it('401 lalu refresh sukses: retry memakai access token baru', async () => {
		localStorage.setItem('access_token', 'old');
		localStorage.setItem('refresh_token', 'r1');

		const fetchMock = vi
			.fn()
			// request awal ditolak
			.mockResolvedValueOnce(json(401, { success: false }))
			// refresh berhasil
			.mockResolvedValueOnce(
				json(200, { success: true, data: { access_token: 'new', refresh_token: 'r2' } })
			)
			// retry dengan token baru
			.mockResolvedValueOnce(json(200, { success: true, data: { ok: true } }));
		vi.stubGlobal('fetch', fetchMock);

		const res = await apiClient.get<{ ok: boolean }>('/products');

		expect(res.data).toEqual({ ok: true });
		expect(localStorage.getItem('access_token')).toBe('new');
		expect(localStorage.getItem('refresh_token')).toBe('r2');

		const retryHeaders = fetchMock.mock.calls[2][1].headers as Record<string, string>;
		expect(retryHeaders['Authorization']).toBe('Bearer new');
	});

	it('401 lalu refresh gagal: sesi dihapus dan diarahkan ke login', async () => {
		localStorage.setItem('access_token', 'old');
		localStorage.setItem('refresh_token', 'r1');

		vi.stubGlobal(
			'fetch',
			vi
				.fn()
				.mockResolvedValueOnce(json(401, { success: false }))
				.mockResolvedValueOnce(json(401, { success: false }))
		);

		await expect(apiClient.get('/products')).rejects.toThrow('Unauthorized');
		expect(localStorage.getItem('access_token')).toBeNull();
		expect(goto).toHaveBeenCalledWith('/login');
	});
});
