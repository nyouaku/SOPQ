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
					retry: 1
				}
			}
		});
	}
	return client;
}
