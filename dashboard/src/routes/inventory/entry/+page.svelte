<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import GridEntry from '$lib/components/GridEntry.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { localMonth, previousMonth } from '$lib/utils/date';

	/** Mode tampilan grid harian. Dipakai lintas sesi lewat localStorage. */
	type ViewMode = 'grid' | 'compact' | 'quick';
	const SETTINGS_KEY = 'stock-opname-inventory-settings';

	let viewMode: ViewMode = 'grid';
	let direction: 'out' | 'in' = 'out';
	let warehouseId = '';
	let month = localMonth(); // YYYY-MM, waktu lokal
	let warehouses: any[] = [];
	let isLoading = true;

	let gridEntry: GridEntry;

	// ---- Toolbar: ekspor Excel & copy bulan lalu (dipindah ke sini agar toolbar satukan semua kontrol)
	let isExporting = false;
	let isCopying = false;

	async function loadWarehouses() {
		const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
		warehouses = res.success && res.data ? res.data.warehouses || [] : [];
		if (warehouses.length > 0) warehouseId = warehouses[0].id;
	}

	async function loadSettings() {
		// Selalu muat gudang: warehouseId wajib ada sebelum GridEntry dirender.
		// Sebelumnya kalau settings ada tapi defaultWarehouseId kosong, loadWarehouses()
		// tak pernah dipanggil → warehouseId tetap '' → loader tak pernah berhenti.
		await loadWarehouses();

		const raw = localStorage.getItem(SETTINGS_KEY);
		if (!raw) return;

		try {
			const s = JSON.parse(raw);
			if (s.gridEntryMode === 'grid' || s.gridEntryMode === 'compact' || s.gridEntryMode === 'quick') {
				viewMode = s.gridEntryMode;
			}
			if (s.defaultMode === 'in' || s.defaultMode === 'out') direction = s.defaultMode;
			// Gudang default dipakai hanya bila masih ada di daftar.
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
		// Simpan pilihan agar sesi berikutnya membuka mode yang sama.
		try {
			const raw = localStorage.getItem(SETTINGS_KEY);
			const s = raw ? JSON.parse(raw) : {};
			localStorage.setItem(SETTINGS_KEY, JSON.stringify({ ...s, gridEntryMode: m }));
		} catch {
			// localStorage penuh / diblokir: abaikan, mode tetap aktif untuk sesi ini.
		}
	}

	async function exportGrid() {
		if (!warehouseId) return;
		isExporting = true;
		try {
			const token = localStorage.getItem('access_token');
			const url = `${API_ENDPOINTS.inventoryGrid}/export?warehouse_id=${warehouseId}&month=${month}&movement_type=${direction}`;
			const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } });
			if (!res.ok) throw new Error('Gagal mengekspor grid');
			const blob = await res.blob();
			const dlUrl = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = dlUrl;
			a.download = `inventory-grid-${direction}-${month}.xlsx`;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(dlUrl);
			document.body.removeChild(a);
			toastStore.success('Grid diekspor ke Excel');
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal mengekspor grid');
		} finally {
			isExporting = false;
		}
	}

	async function copyPrevMonth() {
		if (!warehouseId) return;
		const srcMonth = previousMonth(month);
		if (srcMonth === month) {
			toastStore.error('Bulan sumber dan tujuan sama');
			return;
		}
		isCopying = true;
		try {
			const res = await apiClient.post<{ movements_created: number; items_posted: number }>(
				API_ENDPOINTS.inventoryGridCopy,
				{
					warehouse_id: warehouseId,
					src_month: srcMonth,
					dst_month: month,
					movement_type: direction,
					sub_type: direction === 'in' ? 'purchase' : 'sale'
				}
			);
			if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal menyalin');
			toastStore.success(`${res.data.items_posted} isian disalin dari ${srcMonth}`);
			gridEntry?.reload();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal menyalin bulan lalu');
		} finally {
			isCopying = false;
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
	<div class="page" in:fly={{ y: 8, duration: 360, easing: cubicOut }}>
		<div class="flex items-center gap-3 mb-4">
			<button class="btn btn-ghost btn-sm" on:click={() => goto('/inventory')} aria-label="Kembali">
				←
			</button>
			<h1 class="page-title">Grid Harian</h1>
		</div>
		<p class="page-subtitle mb-4">
			Baris = produk, kolom = tanggal.
			{#if viewMode === 'quick'}
				Mode Cepat: pilih produk, tanggal, jumlah, lalu Simpan.
			{:else if viewMode === 'compact'}
				Mode Ringkas: hanya tanggal yang punya isian ditampilkan.
			{:else}
				Klik sel lalu isi angka {direction === 'out' ? 'keluar' : 'masuk'} hari itu. Paste blok
				dari Excel juga bisa (Ctrl+V di sel pertama).
			{/if}
		</p>

		<!-- Toolbar: arah pergerakan, gudang, bulan, ekspor & copy -->
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

				<button class="btn btn-ghost btn-sm" on:click={exportGrid} disabled={isExporting}>
					{#if isExporting}
						<span class="loading loading-spinner loading-xs"></span>
					{/if}
					📥 Excel
				</button>
				<button class="btn btn-ghost btn-sm" on:click={copyPrevMonth} disabled={isCopying}>
					{#if isCopying}
						<span class="loading loading-spinner loading-xs"></span>
					{/if}
					📋 Copy bulan lalu
				</button>
			</div>
		</div>

		{#if isLoading}
			<div class="flex h-64 items-center justify-center"><Loader size={48} /></div>
		{:else if !warehouseId}
			<div class="py-16 text-center">
				<p class="text-sm text-base-content/60">Belum ada gudang.</p>
				<p class="mt-1 text-xs text-base-content/40">Tambahkan gudang dulu di menu Gudang sebelum mengisi grid.</p>
				<a href="/warehouses" class="btn btn-primary btn-sm mt-4">Ke halaman Gudang</a>
			</div>
		{:else}
			<GridEntry
				bind:this={gridEntry}
				viewMode={viewMode}
				onViewModeChange={onViewModeChange}
				warehouseId={warehouseId}
				direction={direction}
				month={month}
			/>
		{/if}
	</div>
</LayoutWrapper>
