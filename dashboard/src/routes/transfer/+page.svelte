<script lang="ts">
	import { onMount } from 'svelte';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let products: any[] = [];
	let selected = new Set<string>();
	let warehouses: any[] = [];
	let locations: any[] = [];
	let boxes: any[] = [];

	let selectedWarehouse = '';
	let selectedLocation = '';
	let selectedBox = '';
	let isLoading = true;
	let isTransferring = false;
	let loadingBoxes = false;
	let loadingLocations = false;

	onMount(async () => {
		try {
			const [prodRes, whRes] = await Promise.all([
				apiClient.get<{ products: any[] }>(`${API_ENDPOINTS.products}?limit=500`),
				apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses)
			]);
			if (prodRes.success && prodRes.data) {
				products = prodRes.data.products || [];
			}
			if (whRes.success && whRes.data) {
				warehouses = whRes.data.warehouses || [];
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat data');
		} finally {
			isLoading = false;
		}
	});

	async function onWarehouseChange() {
		selectedLocation = '';
		selectedBox = '';
		boxes = [];
		if (!selectedWarehouse) {
			locations = [];
			return;
		}
		loadingLocations = true;
		try {
			const res = await apiClient.get<{ locations: any[] }>(
				API_ENDPOINTS.warehouseLocations(selectedWarehouse)
			);
			if (res.success && res.data) {
				locations = res.data.locations || [];
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat lokasi');
		} finally {
			loadingLocations = false;
		}
	}

	async function onLocationChange() {
		selectedBox = '';
		if (!selectedLocation) {
			boxes = [];
			return;
		}
		loadingBoxes = true;
		try {
			const res = await apiClient.get<{ boxes: any[] }>(API_ENDPOINTS.locationBoxes(selectedLocation));
			if (res.success && res.data) {
				boxes = res.data.boxes || [];
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat rak');
		} finally {
			loadingBoxes = false;
		}
	}

	function toggle(id: string) {
		if (selected.has(id)) selected.delete(id);
		else selected.add(id);
		selected = new Set(selected);
	}

	function selectAll() {
		products.forEach((p) => selected.add(p.id));
		selected = new Set(selected);
	}

	function clearSelection() {
		selected = new Set();
	}

	async function transfer() {
		if (selected.size === 0 || !selectedBox) {
			toastStore.error('Pilih produk dan rak tujuan');
			return;
		}
		isTransferring = true;
		try {
			const res = await apiClient.post<{ transferred: number }>(`${API_ENDPOINTS.products}/transfer`, {
				product_ids: Array.from(selected),
				to_box_id: selectedBox
			});
			if (res.success) {
				toastStore.success(`${res.data?.transferred || 0} produk berhasil dipindahkan`);
				selected = new Set();
				// Reload products
				const prodRes = await apiClient.get<{ products: any[] }>(`${API_ENDPOINTS.products}?limit=500`);
				if (prodRes.success && prodRes.data) {
					products = prodRes.data.products || [];
				}
			} else {
				toastStore.error(res.error?.message || 'Gagal memindahkan');
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memindahkan');
		} finally {
			isTransferring = false;
		}
	}
</script>

<svelte:head>
	<title>Pindah Rak - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6">
			<h1 class="page-title">Pindah Rak (Transfer)</h1>
			<p class="mt-1 text-sm text-base-content/60">
				Pindahkan beberapa produk ke rak/box tujuan dalam satu operasi
			</p>
		</div>

		<!-- Pilih Rak Tujuan -->
		<div class="mb-6 rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
			<h2 class="mb-3 font-semibold">Rak Tujuan</h2>
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
				<div>
					<label class="label py-0" for="warehouse">
						<span class="label-text text-xs">Gudang</span>
					</label>
					<select
						id="warehouse"
						class="select select-bordered select-sm w-full"
						bind:value={selectedWarehouse}
						on:change={onWarehouseChange}
					>
						<option value="">Pilih gudang...</option>
						{#each warehouses as wh}
							<option value={wh.id}>{wh.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="label py-0" for="location">
						<span class="label-text text-xs">Lokasi</span>
					</label>
					<select
						id="location"
						class="select select-bordered select-sm w-full"
						bind:value={selectedLocation}
						on:change={onLocationChange}
						disabled={!selectedWarehouse || loadingLocations}
					>
						<option value="">{loadingLocations ? 'Memuat...' : 'Pilih lokasi...'}</option>
						{#each locations as loc}
							<option value={loc.id}>{loc.code} — {loc.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="label py-0" for="box">
						<span class="label-text text-xs">Rak/Box</span>
					</label>
					<select
						id="box"
						class="select select-bordered select-sm w-full"
						bind:value={selectedBox}
						disabled={!selectedLocation || loadingBoxes}
					>
						<option value="">{loadingBoxes ? 'Memuat...' : 'Pilih rak...'}</option>
						{#each boxes as box}
							<option value={box.id}>{box.code} — {box.name}</option>
						{/each}
					</select>
				</div>
			</div>
		</div>

		<!-- Pilih Produk -->
		<div class="mb-6 rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
			<div class="mb-3 flex items-center justify-between">
				<h2 class="font-semibold">Pilih Produk ({selected.size} dipilih)</h2>
				<div class="flex gap-2">
					<button class="btn btn-ghost btn-sm" on:click={selectAll}>Pilih Semua</button>
					<button class="btn btn-ghost btn-sm" on:click={clearSelection}>Kosongkan</button>
				</div>
			</div>

			{#if isLoading}
				<div class="flex h-40 items-center justify-center">
					<Loader size={48} />
				</div>
			{:else}
				<div class="max-h-96 overflow-y-auto rounded-xl bg-base-100 p-2">
					<div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
						{#each products as product}
							<label
								class="flex cursor-pointer items-center gap-3 rounded-lg border border-base-300 p-3
								{selected.has(product.id) ? 'border-primary bg-primary/5' : ''}"
							>
								<input
									type="checkbox"
									class="checkbox checkbox-primary checkbox-sm"
									checked={selected.has(product.id)}
									on:change={() => toggle(product.id)}
								/>
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium">{product.name}</p>
									<p class="metric text-xs text-base-content/60">{product.product_code}</p>
									{#if product.box?.code}
										<p class="text-xs text-base-content/40">Rak: {product.box.code}</p>
									{/if}
								</div>
							</label>
						{/each}
					</div>
				</div>
			{/if}

			<button
				class="btn btn-primary mt-4 w-full"
				on:click={transfer}
				disabled={selected.size === 0 || !selectedBox || isTransferring}
			>
				{#if isTransferring}
					<span class="loading loading-spinner loading-sm"></span>
					Memproses...
				{:else}
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
					</svg>
					Pindahkan ({selected.size} produk)
				{/if}
			</button>
		</div>
	</div>
</LayoutWrapper>
