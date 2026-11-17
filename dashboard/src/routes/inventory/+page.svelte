<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let isLoading = true;
	let summary: any = null;
	let warehouses: any[] = [];
	let selectedWarehouse = '';
	let month = new Date().toISOString().slice(0, 7); // YYYY-MM

	async function load() {
		isLoading = true;
		try {
			const wh = selectedWarehouse;
			if (!wh) {
				summary = null;
				const wRes = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
				warehouses = wRes.success && wRes.data ? wRes.data.warehouses || [] : [];
				if (warehouses.length > 0) {
					selectedWarehouse = warehouses[0].id;
					await load();
				}
				return;
			}
			const [sRes, wRes] = await Promise.all([
				apiClient.get<any>(
					`${API_ENDPOINTS.inventorySummary}?warehouse_id=${wh}&month=${month}`
				),
				apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses)
			]);
			summary = sRes.success ? sRes.data : null;
			warehouses = wRes.success && wRes.data ? wRes.data.warehouses || [] : [];
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal memuat ringkasan');
		} finally {
			isLoading = false;
		}
	}

	function fmtNum(n: number): string {
		return new Intl.NumberFormat('id-ID').format(n ?? 0);
	}
	function fmtRp(n: number): string {
		return 'Rp ' + new Intl.NumberFormat('id-ID').format(Math.round(n ?? 0));
	}

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
			<!-- Grid Entry = aksi utama (kebiasaan Excel) -->
			<a href="/inventory/entry" class="btn btn-primary btn-sm shrink-0 md:btn-md">
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h10" />
				</svg>
				Isi Grid Harian
			</a>
		</div>

		<!-- Filter gudang + bulan -->
		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-3">
			<div class="flex flex-wrap items-center gap-2">
				<label class="text-xs text-base-content/60" for="dash-wh">Gudang:</label>
				<select id="dash-wh" class="select select-bordered select-sm" bind:value={selectedWarehouse} on:change={load}>
					{#each warehouses as wh}
						<option value={wh.id}>{wh.name}</option>
					{/each}
				</select>
				<label class="text-xs text-base-content/60 ml-2" for="dash-month">Bulan:</label>
				<input id="dash-month" type="month" class="input input-bordered input-sm" bind:value={month} on:change={load} />
			</div>
		</div>

		{#if isLoading}
			<div class="flex h-40 items-center justify-center"><Loader size={48} /></div>
		{:else if !summary}
			<p class="py-8 text-center text-sm text-base-content/60">Pilih gudang untuk melihat ringkasan</p>
		{:else}
			<!-- KPI (dari server) -->
			<div class="mb-4 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-base-300 bg-base-300 lg:grid-cols-4">
				<div class="bg-base-200 p-4" in:fly={{ y: 8, duration: 360, delay: 0, easing: cubicOut }}>
					<div class="text-xs text-base-content/60">Masuk (unit)</div>
					<div class="metric text-2xl font-semibold text-success">{fmtNum(summary.total_in)}</div>
				</div>
				<div class="bg-base-200 p-4" in:fly={{ y: 8, duration: 360, delay: 60, easing: cubicOut }}>
					<div class="text-xs text-base-content/60">Keluar (unit)</div>
					<div class="metric text-2xl font-semibold text-error">{fmtNum(summary.total_out)}</div>
				</div>
				<div class="bg-base-200 p-4" in:fly={{ y: 8, duration: 360, delay: 120, easing: cubicOut }}>
					<div class="text-xs text-base-content/60">Transfer</div>
					<div class="metric text-2xl font-semibold text-info">{fmtNum(summary.total_transfer)}</div>
				</div>
				<div class="bg-base-200 p-4" in:fly={{ y: 8, duration: 360, delay: 180, easing: cubicOut }}>
					<div class="text-xs text-base-content/60">Nilai Keluar</div>
					<div class="metric text-2xl font-semibold">{fmtRp(summary.total_value_out)}</div>
				</div>
			</div>

			<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
				<!-- Top keluar -->
				<div class="card-apple p-4">
					<h2 class="mb-3 font-semibold">Produk Terbanyak Keluar</h2>
					{#if !summary.top_out?.length}
						<p class="text-sm text-base-content/50">Belum ada data</p>
					{:else}
						<ul class="space-y-2">
							{#each summary.top_out as t}
								<li class="flex items-center justify-between gap-2 text-sm">
									<span class="truncate">{t.name}</span>
									<span class="metric shrink-0 font-semibold text-error">{t.quantity}</span>
								</li>
							{/each}
						</ul>
					{/if}
				</div>

				<!-- Stok menipis -->
				<div class="card-apple p-4">
					<h2 class="mb-3 font-semibold">Stok Menipis</h2>
					{#if !summary.low_stock?.length}
						<p class="text-sm text-base-content/50">Semua stok aman</p>
					{:else}
						<ul class="space-y-2">
							{#each summary.low_stock as l}
								<li class="flex items-center justify-between gap-2 text-sm">
									<span class="truncate">{l.name}</span>
									<span class="metric shrink-0 font-semibold text-warning">
										{l.quantity_system} / {l.minimum_stock}
									</span>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</div>

			<!-- Quick actions -->
			<div class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
				<a href="/inventory/entry" class="card-apple card-apple--hover p-4">
					<div class="text-sm font-semibold">📋 Isi Grid Harian</div>
					<div class="text-xs text-base-content/60">Masuk/keluar per tanggal (ala Excel)</div>
				</a>
				<a href="/inventory/movements" class="card-apple card-apple--hover p-4">
					<div class="text-sm font-semibold text-info">📥 Semua Pergerakan</div>
					<div class="text-xs text-base-content/60">Daftar lengkap dokumen</div>
				</a>
				<a href="/products" class="card-apple card-apple--hover p-4">
					<div class="text-sm font-semibold text-info">📦 Kelola Produk</div>
					<div class="text-xs text-base-content/60">Tambah / ubah produk</div>
				</a>
			</div>
		{/if}
	</div>
</LayoutWrapper>
