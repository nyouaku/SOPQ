<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let isLoading = true;
	let movements: any[] = [];
	let warehouses: any[] = [];
	let selectedWarehouse = '';

	// Agregat dari data termuat (tanpa ubah backend/API).
	$: inCount = movements.filter((m) => m.movement_type === 'in').length;
	$: outCount = movements.filter((m) => m.movement_type === 'out').length;
	$: transferCount = movements.filter((m) => m.movement_type === 'transfer').length;
	$: recent = movements.slice(0, 6);

	async function load() {
		try {
			const q = selectedWarehouse ? `?warehouse_id=${selectedWarehouse}&limit=100` : '?limit=100';
			const [mRes, wRes] = await Promise.all([
				apiClient.get<{ data: any[] }>(`${API_ENDPOINTS.inventory}${q}`),
				apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses)
			]);
			movements = mRes.success && mRes.data ? mRes.data.data || [] : [];
			warehouses = wRes.success && wRes.data ? wRes.data.warehouses || [] : [];
		} catch (e: any) {
			toastError(e);
		} finally {
			isLoading = false;
		}
	}

	function toastError(e: any) {
		toastStore.error(e?.message || 'Gagal memuat');
	}

	function warehouseName(id: string): string {
		return warehouses.find((w) => w.id === id)?.name || '-';
	}

	function fmtDate(iso: string) {
		if (!iso) return '-';
		return new Date(iso).toLocaleString('id-ID', {
			day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit'
		});
	}

	const typeBadge: Record<string, string> = { in: 'success', out: 'error', transfer: 'info' };
	const typeLabel: Record<string, string> = { in: 'Masuk', out: 'Keluar', transfer: 'Transfer' };

	load();
</script>

<svelte:head>
	<title>Inventory - Dashboard</title>
</svelte:head>

<LayoutWrapper>
	<div class="page pb-24 md:pb-28">
		<div class="mb-6 flex flex-wrap items-end justify-between gap-3">
			<div>
				<h1 class="page-title">Inventory Management</h1>
				<p class="page-subtitle">Pantau pergerakan stok: barang masuk, keluar, dan transfer</p>
			</div>
		</div>

		<!-- Filter gudang -->
		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-3">
			<div class="flex flex-wrap items-center gap-2">
				<label class="text-xs text-base-content/60" for="dash-wh">Gudang:</label>
				<select id="dash-wh" class="select select-bordered select-sm" bind:value={selectedWarehouse} on:change={load}>
					<option value="">Semua gudang</option>
					{#each warehouses as wh}
						<option value={wh.id}>{wh.name}</option>
					{/each}
				</select>
				<a href="/inventory/movements" class="btn btn-ghost btn-sm ml-auto">Semua pergerakan →</a>
			</div>
		</div>

		{#if isLoading}
			<div class="flex h-40 items-center justify-center"><Loader size={48} /></div>
		{:else}
			<!-- Stat tiles (hairline grid, Apple-adjacent) -->
			<div class="mb-4 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-base-300 bg-base-300 lg:grid-cols-4">
				<div class="bg-base-200 p-4" in:fly={{ y: 8, duration: 360, delay: 0, easing: cubicOut }}>
					<div class="text-xs text-base-content/60">Total Pergerakan</div>
					<div class="metric text-2xl font-semibold">{movements.length}</div>
				</div>
				<div class="bg-base-200 p-4" in:fly={{ y: 8, duration: 360, delay: 60, easing: cubicOut }}>
					<div class="text-xs text-base-content/60">Masuk (IN)</div>
					<div class="metric text-2xl font-semibold text-success">{inCount}</div>
				</div>
				<div class="bg-base-200 p-4" in:fly={{ y: 8, duration: 360, delay: 120, easing: cubicOut }}>
					<div class="text-xs text-base-content/60">Keluar (OUT)</div>
					<div class="metric text-2xl font-semibold text-error">{outCount}</div>
				</div>
				<div class="bg-base-200 p-4" in:fly={{ y: 8, duration: 360, delay: 180, easing: cubicOut }}>
					<div class="text-xs text-base-content/60">Transfer</div>
					<div class="metric text-2xl font-semibold text-info">{transferCount}</div>
				</div>
			</div>

			<!-- Transaksi terbaru -->
			<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
				<div class="mb-3 flex items-center justify-between">
					<h2 class="font-semibold">Transaksi Terbaru</h2>
					<a href="/inventory/movements" class="text-xs link link-primary">Lihat semua</a>
				</div>
				{#if recent.length === 0}
					<p class="py-6 text-center text-sm text-base-content/60">Belum ada pergerakan stok</p>
				{:else}
					<ul class="divide-y divide-base-300">
						{#each recent as m (m.id)}
							<li class="flex items-center gap-3 py-2.5">
								<span class="badge badge-sm capitalize {typeBadge[m.movement_type] || 'badge-ghost'}">
									{typeLabel[m.movement_type] || m.movement_type}
								</span>
								<div class="min-w-0 flex-1">
									<a href="/inventory/{m.id}" class="link link-hover truncate text-sm font-medium">{m.code}</a>
									<div class="truncate text-xs text-base-content/60">{warehouseName(m.warehouse_id)}</div>
								</div>
								<div class="text-xs text-base-content/50">{fmtDate(m.created_at)}</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<!-- Quick actions -->
			<div class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
				<a href="/inventory/new" class="card-apple card-apple--hover p-4">
					<div class="text-sm font-semibold text-success">⬇️ Barang Masuk</div>
					<div class="text-xs text-base-content/60">Catat pembelian / retur</div>
				</a>
				<a href="/inventory/new" class="card-apple card-apple--hover p-4">
					<div class="text-sm font-semibold text-error">⬆️ Barang Keluar</div>
					<div class="text-xs text-base-content/60">Catat penjualan / rusak</div>
				</a>
				<a href="/inventory/products" class="card-apple card-apple--hover p-4">
					<div class="text-sm font-semibold text-info">📦 Kelola Produk</div>
					<div class="text-xs text-base-content/60">Tambah / ubah produk</div>
				</a>
			</div>
		{/if}
	</div>
</LayoutWrapper>
