// ===== FILE: dashboard/src/lib/api/client.ts =====
import { goto } from '$app/navigation';
import { API_BASE_URL } from './config';

interface ApiResponse<T> {
	success: boolean;
	data?: T;
	message?: string;
	error?: {
		code: string;
		message: string;
	};
}

class ApiClient {
	private baseUrl: string;

	// Single-flight refresh: saat banyak request paralel kena 401, semuanya berbagi
	// SATU proses refresh. Tanpa ini tiap request me-refresh sendiri; refresh kedua
	// memakai token lama yang sudah dirotasi, backend menganggapnya token reuse dan
	// mencabut semua sesi (force-logout). Inilah sebabnya transfer tiba-tiba 401.
	private refreshPromise: Promise<boolean> | null = null;

	constructor(baseUrl: string) {
		this.baseUrl = baseUrl;
	}

	private async getAccessToken(): Promise<string | null> {
		if (typeof window === 'undefined') return null;
		return localStorage.getItem('access_token');
	}

	private async setAccessToken(token: string): Promise<void> {
		if (typeof window === 'undefined') return;
		localStorage.setItem('access_token', token);
	}

	private async getRefreshToken(): Promise<string | null> {
		if (typeof window === 'undefined') return null;
		return localStorage.getItem('refresh_token');
	}

	private async refreshAccessToken(): Promise<boolean> {
		// Sudah ada refresh berjalan? Gabungkan — jangan trigger ke-dua kali.
		if (this.refreshPromise) return this.refreshPromise;

		this.refreshPromise = this._doRefresh();
		try {
			return await this.refreshPromise;
		} finally {
			this.refreshPromise = null;
		}
	}

	private async _doRefresh(): Promise<boolean> {
		const refreshToken = await this.getRefreshToken();
		if (!refreshToken) return false;

		try {
			const response = await fetch(`${this.baseUrl}/auth/refresh`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ refresh_token: refreshToken })
			});

			if (response.ok) {
				const ct = response.headers.get('content-type') || '';
				if (!ct.includes('application/json')) return false;
				const text = await response.text();
				if (!text) return false;
				const data: ApiResponse<{ access_token: string; refresh_token?: string }> =
					JSON.parse(text);
				if (data.success && data.data) {
					await this.setAccessToken(data.data.access_token);
					if (data.data.refresh_token) {
						localStorage.setItem('refresh_token', data.data.refresh_token);
					}
					return true;
				}
			}
			return false;
		} catch {
			return false;
		}
	}

	async request<T>(
		endpoint: string,
		options: RequestInit = {}
	): Promise<ApiResponse<T>> {
		const token = await this.getAccessToken();
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			...((options.headers as Record<string, string>) || {})
		};

		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}

		// Endpoint auth (login/refresh) tidak lewat logika refresh-token.
		// 401 di sini berarti kredensial salah — biarkan respons asli sampai ke pemanggil
		// supaya pesan errornya (mis. "username/password salah") bisa ditampilkan.
		const isAuthEndpoint = endpoint.startsWith('/auth/');

		try {
			let response = await fetch(`${this.baseUrl}${endpoint}`, {
				...options,
				headers
			});

			// Retry sekali dengan refresh token saat 401 (bukan pada endpoint auth).
			if (response.status === 401 && !isAuthEndpoint) {
				const refreshed = await this.refreshAccessToken();
				if (refreshed) {
					const newToken = await this.getAccessToken();
					if (newToken) {
						headers['Authorization'] = `Bearer ${newToken}`;
						response = await fetch(`${this.baseUrl}${endpoint}`, {
							...options,
							headers
						});
					}
				}

				// Refresh gagal ATAU retry tetap 401: hentikan loop, paksa login.
				// Tanpa ini svelte-query terus refetch dan memicu request berulang.
				if (response.status === 401) {
					if (typeof window !== 'undefined' && !location.pathname.startsWith('/login')) {
						localStorage.clear();
						goto('/login');
					}
					throw new Error('Unauthorized');
				}
			}

			// Parse respons dengan aman: respons non-JSON (204 kosong, error HTML dari
			// reverse-proxy, body rusak) tidak boleh melempar "Unexpected ... after JSON".
			let data: ApiResponse<T>;
			const ct = response.headers.get('content-type') || '';
			if (!ct.includes('application/json')) {
				if (!response.ok) {
					throw new Error(`Server error ${response.status}`);
				}
				data = { success: response.ok } as ApiResponse<T>;
			} else {
				const text = await response.text();
				data = text ? (JSON.parse(text) as ApiResponse<T>) : ({ success: response.ok } as ApiResponse<T>);
			}
			return data;
		} catch (error) {
			throw error;
		}
	}

	async get<T>(endpoint: string): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, { method: 'GET' });
	}

	async post<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'POST',
			body: body ? JSON.stringify(body) : undefined
		});
	}

	async put<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'PUT',
			body: body ? JSON.stringify(body) : undefined
		});
	}

	async patch<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'PATCH',
			body: body ? JSON.stringify(body) : undefined
		});
	}

	async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, { method: 'DELETE' });
	}
}

export const apiClient = new ApiClient(API_BASE_URL);
// ===== END FILE =====
