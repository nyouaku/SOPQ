<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import { toastStore } from '$lib/stores/toast';

	// Preferensi khusus Inventory, disimpan di localStorage (pola themeStore).
	const KEY = 'stock-opname-inventory-settings';

	interface InvSettings {
		defaultWarehouseId: string;
		defaultMode: 'out' | 'in';
		defaultSubTypeOut: string;
		showInOutColumns: boolean;
		lowStockAlert: boolean;
	}
	let settings: InvSettings = {
		defaultWarehouseId: '',
		defaultMode: 'out',
		defaultSubTypeOut: 'sale',
		showInOutColumns: true,
		lowStockAlert: true
	};

	let warehouses: any[] = [];

	onMount(async () => {
		try {
			// Muat tersimpan
			const raw = localStorage.getItem(KEY);
			if (raw) settings = { ...settings, ...JSON.parse(raw) };

			// Muat daftar gudang untuk pilihan default
			const { apiClient } = await import('$lib/api/client');
			const { API_ENDPOINTS } = await import('$lib/api/config');
			const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
			warehouses = res.success && res.data ? res.data.warehouses || [] : [];
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal memuat pengaturan');
		}
	});

	function save() {
		localStorage.setItem(KEY, JSON.stringify(settings));
		toastStore.success('Pengaturan inventory tersimpan');
	}
</script>

<svelte:head>
	<title>Settings - Inventory</title>
</svelte:head>

<LayoutWrapper>
	<div class="page page-with-bottombar">
		<div class="flex items-center gap-3 mb-6">
			<button class="btn btn-ghost btn-sm" on:click={() => goto('/inventory')}>←</button>
			<h1 class="page-title">Pengaturan Inventory</h1>
		</div>

		<div class="card-apple p-5">
			<h2 class="mb-4 font-semibold">Preferensi Inventory</h2>

			<!-- Gudang default -->
			<div class="mb-4">
				<label class="label py-0" for="inv-wh">
					<span class="label-text text-sm">Gudang default</span>
				</label>
				<select id="inv-wh" class="select select-bordered w-full max-w-sm" bind:value={settings.defaultWarehouseId}>
					<option value="">-- Pilih gudang --</option>
					{#each warehouses as wh}
						<option value={wh.id}>{wh.name}</option>
					{/each}
				</select>
				<p class="text-xs text-base-content/50 mt-1">Dipakai otomatis saat membuka Grid Harian & form pergerakan.</p>
			</div>

			<!-- Mode default -->
			<div class="mb-4">
				<label class="label py-0" for="inv-mode">
					<span class="label-text text-sm">Mode default Grid Harian</span>
				</label>
				<select id="inv-mode" class="select select-bordered w-full max-w-sm" bind:value={settings.defaultMode}>
					<option value="out">Keluar (OUT)</option>
					<option value="in">Masuk (IN)</option>
				</select>
			</div>

			<!-- Sub-tipe keluar default -->
			<div class="mb-4">
				<label class="label py-0" for="inv-sub">
					<span class="label-text text-sm">Alasan default barang keluar</span>
				</label>
				<select id="inv-sub" class="select select-bordered w-full max-w-sm" bind:value={settings.defaultSubTypeOut}>
					<option value="sale">Penjualan</option>
					<option value="damaged">Rusak</option>
					<option value="lost">Hilang</option>
					<option value="expired">Kedaluwarsa</option>
					<option value="return">Retur ke supplier</option>
					<option value="other">Lainnya</option>
				</select>
			</div>

			<!-- Toggle: tampilkan kolom IN/OUT -->
			<div class="mb-4">
				<label class="label cursor-pointer justify-between">
					<div>
						<span class="label-text text-sm">Tampilkan pasangan kolom IN/OUT</span>
						<p class="text-xs text-base-content/50">Tiap tanggal punya 2 kolom (masuk & keluar).</p>
					</div>
					<input type="checkbox" class="toggle toggle-primary" bind:checked={settings.showInOutColumns} />
				</label>
			</div>

			<!-- Toggle: notifikasi stok menipis -->
			<div class="mb-4">
				<label class="label cursor-pointer justify-between">
					<div>
						<span class="label-text text-sm">Peringatan stok menipis</span>
						<p class="text-xs text-base-content/50">Tampilkan widget stok di bawah minimum.</p>
					</div>
					<input type="checkbox" class="toggle toggle-primary" bind:checked={settings.lowStockAlert} />
				</label>
			</div>

			<div class="mt-6">
				<button class="btn btn-primary" on:click={save}>Simpan Pengaturan</button>
			</div>
		</div>
	</div>
</LayoutWrapper>
