<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import SpreadsheetEntry from '$lib/components/SpreadsheetEntry.svelte';
	import QuickEntry from '$lib/components/QuickEntry.svelte';
	import InventoryLedgerGrid from '$lib/components/InventoryLedgerGrid.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { localMonth } from '$lib/utils/date';

	/** Tampilan input harian: Ledger (model Excel), Spreadsheet, atau Cepat. */
	type ViewMode = 'ledger' | 'spreadsheet' | 'quick';
	const SETTINGS_KEY = 'stock-opname-inventory-settings';

	let viewMode: ViewMode = 'ledger';
	let direction: 'out' | 'in' = 'out';
	let warehouseId = '';
	let month = localMonth();
	let warehouses: any[] = [];
	let isLoading = true;

	async function loadWarehouses() {
		const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
		warehouses = res.success && res.data ? res.data.warehouses || [] : [];
		if (warehouses.length > 0) warehouseId = warehouses[0].id;
	}

	async function loadSettings() {
		// Daftar gudang selalu dimuat dulu: warehouseId wajib ada sebelum komponen dirender.
		await loadWarehouses();

		const raw = localStorage.getItem(SETTINGS_KEY);
		if (!raw) return;
		try {
			const s = JSON.parse(raw);
			if (s.gridEntryMode === 'quick' || s.gridEntryMode === 'spreadsheet' || s.gridEntryMode === 'ledger') {
				viewMode = s.gridEntryMode;
			}
			if (s.defaultMode === 'in' || s.defaultMode === 'out') direction = s.defaultMode;
			if (
				typeof s.defaultWarehouseId === 'string' &&
				s.defaultWarehouseId &&
				warehouses.some((w) => w.id === s.defaultWarehouseId)
			) {
				warehouseId = s.defaultWarehouseId;
			}
		} catch {
			// JSON rusak: biarkan nilai default.
		}
	}

	function onViewModeChange(m: ViewMode) {
		viewMode = m;
		try {
			const raw = localStorage.getItem(SETTINGS_KEY);
			const s = raw ? JSON.parse(raw) : {};
			localStorage.setItem(SETTINGS_KEY, JSON.stringify({ ...s, gridEntryMode: m }));
		} catch {
			// localStorage penuh / diblokir: abaikan, mode tetap aktif untuk sesi ini.
		}
	}

	onMount(async () => {
		try {
			await loadSettings();
		} finally {
			isLoading = false;
		}
	});
</script>

<svelte:head>
	<title>Grid Harian - Inventory</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="flex items-center gap-3 mb-4">
			<button class="btn btn-ghost btn-sm" on:click={() => goto('/inventory')} aria-label="Kembali">
				←
			</button>
			<h1 class="page-title">Grid Harian</h1>
		</div>

		<!-- Toolbar: arah, gudang, bulan -->
		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-3">
			<div class="flex flex-wrap items-center gap-2">
				<div class="grid grid-cols-2 gap-1 rounded-xl bg-base-300 p-1" role="group" aria-label="Arah pergerakan">
					<button
						class="rounded-lg px-3 py-1.5 text-sm font-medium {direction === 'out' ? 'bg-base-100 text-error shadow-sm' : 'opacity-50'}"
						on:click={() => (direction = 'out')}>⬆️ Keluar</button>
					<button
						class="rounded-lg px-3 py-1.5 text-sm font-medium {direction === 'in' ? 'bg-base-100 text-success shadow-sm' : 'opacity-50'}"
						on:click={() => (direction = 'in')}>⬇️ Masuk</button>
				</div>

				<select class="select select-bordered select-sm" bind:value={warehouseId} aria-label="Gudang">
					{#each warehouses as wh}
						<option value={wh.id}>{wh.name}</option>
					{/each}
				</select>

				<input type="month" class="input input-bordered input-sm" bind:value={month} aria-label="Bulan" />

				<!-- Tab tampilan -->
				<div class="ml-auto grid grid-cols-3 gap-1 rounded-xl bg-base-300 p-1" role="tablist">
					<button role="tab" aria-selected={viewMode === 'ledger'}
						class="rounded-lg px-3 py-1.5 text-sm font-medium {viewMode === 'ledger' ? 'bg-base-100 shadow-sm' : 'opacity-60'}"
						on:click={() => onViewModeChange('ledger')}>📒 Ledger</button>
					<button role="tab" aria-selected={viewMode === 'spreadsheet'}
						class="rounded-lg px-3 py-1.5 text-sm font-medium {viewMode === 'spreadsheet' ? 'bg-base-100 shadow-sm' : 'opacity-60'}"
						on:click={() => onViewModeChange('spreadsheet')}>📊 Spreadsheet</button>
					<button role="tab" aria-selected={viewMode === 'quick'}
						class="rounded-lg px-3 py-1.5 text-sm font-medium {viewMode === 'quick' ? 'bg-base-100 shadow-sm' : 'opacity-60'}"
						on:click={() => onViewModeChange('quick')}>⚡ Cepat</button>
				</div>
			</div>
		</div>

		{#if isLoading}
			<div class="flex h-64 items-center justify-center"><Loader size={48} /></div>
		{:else if !warehouseId}
			<div class="py-16 text-center">
				<p class="text-sm text-base-content/60">Belum ada gudang.</p>
				<p class="mt-1 text-xs text-base-content/40">Tambahkan gudang dulu sebelum mengisi grid harian.</p>
				<a href="/warehouses" class="btn btn-primary btn-sm mt-4">Ke halaman Gudang</a>
			</div>
		{:else if viewMode === 'quick'}
			<QuickEntry {warehouseId} {direction} {month} />
		{:else if viewMode === 'spreadsheet'}
			<SpreadsheetEntry {warehouseId} {month} />
		{:else}
			<InventoryLedgerGrid {warehouseId} {month} />
		{/if}
	</div>
</LayoutWrapper>
