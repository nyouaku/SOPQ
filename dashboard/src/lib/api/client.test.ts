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

	it('401 pada login: respons asli diteruskan, tanpa redirect ke login', async () => {
		const fetchMock = vi
			.fn()
			.mockResolvedValueOnce(
				json(401, { success: false, error: { code: 'UNAUTHORIZED', message: 'username atau password salah' } })
			);
		vi.stubGlobal('fetch', fetchMock);

		const res = await apiClient.post('/auth/login', { username: 'admin', password: 'salah' });

		// Tidak melempar; respons asli sampai ke pemanggil.
		expect(res.success).toBe(false);
		expect(res.error?.message).toBe('username atau password salah');
		// Hanya 1 fetch (tidak coba refresh).
		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(goto).not.toHaveBeenCalled();
	});

	it('banyak request 401 paralel: hanya SATU refresh yang dijalankan (single-flight)', async () => {
		localStorage.setItem('access_token', 'old');
		localStorage.setItem('refresh_token', 'r1');

		let refreshCalled = 0;
		const fetchMock = vi.fn((url: string, opts: RequestInit) => {
			if (url.endsWith('/auth/refresh')) {
				refreshCalled++;
				return Promise.resolve(
					json(200, { success: true, data: { access_token: 'new', refresh_token: 'r2' } })
				);
			}
			// Request data: 401 hanya saat masih pakai token lama.
			const auth = (opts.headers as Record<string, string>)['Authorization'];
			if (auth === 'Bearer old') return Promise.resolve(json(401, { success: false }));
			return Promise.resolve(json(200, { success: true, data: { ok: true } }));
		});
		vi.stubGlobal('fetch', fetchMock);

		// Luncurkan 5 request paralel, semua 401 lalu refresh bersama.
		const calls = Array.from({ length: 5 }, () =>
			apiClient.get('/protected').catch(() => null)
		);
		await Promise.all(calls);

		expect(refreshCalled).toBe(1);
	});
});
