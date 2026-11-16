import { QueryClient } from '@tanstack/svelte-query';

/**
 * Satu QueryClient untuk seluruh aplikasi.
 * Dipakai di +layout.svelte (provider) dan di komponen yang perlu
 * invalidate cache (mis. NotificationBell setelah mark-read).
 */
let client: QueryClient | null = null;

export function getQueryClient(): QueryClient {
	if (!client) {
		client = new QueryClient({
			defaultOptions: {
				queries: {
					refetchOnWindowFocus: false,
					// 401 tak pernah dipercoba ulang: retry hanya memperbanyak request 401
					// (memburukkan loop "blinking" saat sesi kedaluwarsa). Retry hanya utk
					// error jaringan/500 — 4xx adalah kegagalan permanen.
					retry: (failureCount, error) => {
						const status = (error as { status?: number } | null)?.status;
						if (status === 401 || status === 403 || status === 404) return false;
						return failureCount < 1;
					}
				}
			}
		});
	}
	return client;
}
