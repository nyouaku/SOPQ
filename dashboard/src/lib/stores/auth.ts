// ===== FILE: dashboard/src/lib/stores/auth.ts =====
import { writable } from 'svelte/store';
import { apiClient } from '$lib/api/client';
import { API_ENDPOINTS } from '$lib/api/config';
import { goto } from '$app/navigation';

interface User {
	id: string;
	username: string;
	full_name: string;
	name?: string;
	role: {
		id: number;
		name: string;
		description?: string;
	};
	role_id?: number;
	email?: string;
}

interface AuthState {
	user: User | null;
	isAuthenticated: boolean;
	isLoading: boolean;
}

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>({
		user: null,
		isAuthenticated: false,
		isLoading: true
	});

	return {
		subscribe,
		login: async (username: string, password: string) => {
			update((state) => ({ ...state, isLoading: true }));

			try {
				const response = await apiClient.post<{
					access_token: string;
					refresh_token: string;
					user: User;
				}>(API_ENDPOINTS.login, { username, password });

				if (response.success && response.data) {
					localStorage.setItem('access_token', response.data.access_token);
					localStorage.setItem('refresh_token', response.data.refresh_token);
					localStorage.setItem('user', JSON.stringify(response.data.user));

					set({
						user: response.data.user,
						isAuthenticated: true,
						isLoading: false
					});

					goto('/');
					return { success: true };
				} else {
					update((state) => ({ ...state, isLoading: false }));
					return { success: false, error: response.error?.message || 'Login failed' };
				}
			} catch (error) {
				update((state) => ({ ...state, isLoading: false }));
				return { success: false, error: 'Network error' };
			}
		},

		logout: async () => {
			try {
				await apiClient.post(API_ENDPOINTS.logout);
			} catch {
				// Ignore error
			}

			localStorage.clear();
			set({ user: null, isAuthenticated: false, isLoading: false });
			goto('/login');
		},

		checkAuth: async () => {
			const token = localStorage.getItem('access_token');
			const userStr = localStorage.getItem('user');

			if (!token || !userStr) {
				set({ user: null, isAuthenticated: false, isLoading: false });
				return;
			}

			try {
				const user = JSON.parse(userStr);
				set({ user, isAuthenticated: true, isLoading: false });
			} catch {
				localStorage.clear();
				set({ user: null, isAuthenticated: false, isLoading: false });
			}
		}
	};
}

export const authStore = createAuthStore();
// ===== END FILE =====
