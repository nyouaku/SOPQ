import { derived, readable, type Readable } from 'svelte/store';
import { createQuery, createMutation } from '@tanstack/svelte-query';
import { apiClient } from '$lib/api/client';
import { API_ENDPOINTS } from '$lib/api/config';

/**
 * Satu tempat untuk semua query read-only.
 * Halaman memanggil helper ini, bukan `apiClient.get` langsung, supaya
 * cache/dedupe/refetch konsisten antar-halaman.
 *
 * Catatan: `createQuery` mengembalikan store Svelte — pakai `$query`.
 */

export const queryKeys = {
	warehouses: ['warehouses'] as const,
	locations: (warehouseId: string) => ['warehouses', warehouseId, 'locations'] as const,
	boxes: (locationId: string) => ['locations', locationId, 'boxes'] as const,
	products: (params: string) => ['products', params] as const,
	suppliers: ['suppliers'] as const,
	notifications: ['notifications'] as const,
	activityLogs: (params: string) => ['activity-logs', params] as const,
	report: (params: string) => ['reports', 'stock-opname', params] as const,
	sessions: (params: string) => ['sessions', params] as const
};

/** Buka `data` dari amplop respons; lempar error agar ditangani query. */
function unwrap<T>(res: {
	success: boolean;
	data?: T;
	error?: { message?: string };
}): T {
	if (!res.success || !res.data) {
		throw new Error(res.error?.message || 'Gagal memuat data');
	}
	return res.data;
}

// --- Fetcher per resource ---

export async function fetchWarehouses() {
	const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
	return unwrap(res).warehouses || [];
}

export async function fetchLocations(warehouseId: string) {
	const res = await apiClient.get<{ locations: any[] }>(
		API_ENDPOINTS.warehouseLocations(warehouseId)
	);
	return unwrap(res).locations || [];
}

export async function fetchBoxes(locationId: string) {
	const res = await apiClient.get<{ boxes: any[] }>(API_ENDPOINTS.locationBoxes(locationId));
	return unwrap(res).boxes || [];
}


export async function fetchSuppliers() {
	const res = await apiClient.get<{ suppliers: any[] }>(API_ENDPOINTS.suppliers);
	return unwrap(res).suppliers || [];
}

export async function fetchNotifications() {
	const res = await apiClient.get<{ notifications: any[]; total: number }>(
		`${API_ENDPOINTS.notifications}?limit=20`
	);
	return unwrap(res);
}

export async function fetchActivityLogs(params: URLSearchParams) {
	const res = await apiClient.get<{ logs: any[]; total: number; page: number; limit: number }>(
		`${API_ENDPOINTS.activityLogs}?${params.toString()}`
	);
	return unwrap(res);
}

export async function fetchStockOpnameReport(params: URLSearchParams) {
	const res = await apiClient.get<{
		data: any[];
		total: number;
		page: number;
		limit: number;
		summary: { total_system: number; total_physical: number; total_difference: number; mismatch_count: number };
	}>(
		`${API_ENDPOINTS.reportsStockOpname}?${params.toString()}`
	);
	return unwrap(res);
}

// --- Query factories ---

/** Cache data master jarang berubah: 5 menit. */
const STATIC = { staleTime: 5 * 60 * 1000 };

export const warehousesQuery = () =>
	createQuery({
		queryKey: queryKeys.warehouses,
		queryFn: fetchWarehouses,
		...STATIC
	});

type StringInput = string | Readable<string>;

/** Bungkus nilai biasa jadi store agar `createQuery` bisa ikut berubah. */
function valStore<T>(value: T | Readable<T>): Readable<T> {
	return typeof (value as Readable<T>).subscribe === 'function'
		? (value as Readable<T>)
		: readable(value as T);
}

export const locationsQuery = (warehouseId: StringInput) =>
	createQuery(
		derived(valStore(warehouseId), ($id) => ({
			queryKey: queryKeys.locations($id),
			queryFn: () => fetchLocations($id),
			enabled: !!$id,
			...STATIC
		}))
	);

export const boxesQuery = (locationId: StringInput) =>
	createQuery(
		derived(valStore(locationId), ($id) => ({
			queryKey: queryKeys.boxes($id),
			queryFn: () => fetchBoxes($id),
			enabled: !!$id,
			...STATIC
		}))
	);


export const suppliersQuery = () =>
	createQuery({ queryKey: queryKeys.suppliers, queryFn: fetchSuppliers, ...STATIC });

/** Notifikasi cepat basi: refetch tiap 60 detik. */
export const notificationsQuery = () =>
	createQuery({
		queryKey: queryKeys.notifications,
		queryFn: fetchNotifications,
		refetchInterval: 60_000
	});

type ParamsInput = URLSearchParams | Readable<URLSearchParams>;

/**
 * `createQuery` membaca options sekali saja. Supaya filter berubah ikut
 * memicu query baru, params harus berupa store — bukan nilai biasa.
 */
function paramsStore(params: ParamsInput): Readable<URLSearchParams> {
	return typeof (params as Readable<URLSearchParams>).subscribe === 'function'
		? (params as Readable<URLSearchParams>)
		: readable(params as URLSearchParams);
}

export const activityLogsQuery = (params: ParamsInput) =>
	createQuery(
		derived(paramsStore(params), ($params) => ({
			queryKey: queryKeys.activityLogs($params.toString()),
			queryFn: () => fetchActivityLogs($params)
		}))
	);

export const stockOpnameReportQuery = (params: ParamsInput) =>
	createQuery(
		derived(paramsStore(params), ($params) => ({
			queryKey: queryKeys.report($params.toString()),
			queryFn: () => fetchStockOpnameReport($params)
		}))
	);

/** Tandai notifikasi sudah dibaca, lalu segarkan daftar. */
export const markNotificationRead = () =>
	createMutation({
		mutationFn: (id: string) => apiClient.patch(`${API_ENDPOINTS.notifications}/${id}/read`, {}),
		onSuccess: () => {
			// Invalidasi ditangani pemanggil lewat queryClient bila perlu.
		}
	});
