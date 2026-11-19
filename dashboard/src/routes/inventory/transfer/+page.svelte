<script lang="ts">
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let fromWarehouseId = '';
	let toWarehouseId = '';
	let referenceCode = '';
	let note = '';

	let products: any[] = [];
	let warehouses: any[] = [];
	let selected = new Set<string>();
	let quantities: Record<string, number> = {}; // product_id -> qty
	let isLoading = true;
	let isTransferring = false;

	async function load() {
		try {
			const [prodRes, whRes] = await Promise.all([
				apiClient.get<{ products: any[] }>(`${API_ENDPOINTS.products}?limit=500`),
				apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses)
			]);
			if (prodRes.success && prodRes.data) products = prodRes.data.products || [];
			if (whRes.success && whRes.data) warehouses = whRes.data.warehouses || [];
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat data');
		} finally {
			isLoading = false;
		}
	}
	load();

	function toggle(id: string) {
		if (selected.has(id)) {
			selected.delete(id);
		} else {
			selected.add(id);
			if (!quantities[id]) quantities[id] = 1; // default jumlah 1
		}
		selected = new Set(selected);
	}
	function selectAll() {
		products.forEach((p) => {
			selected.add(p.id);
			if (!quantities[p.id]) quantities[p.id] = 1;
		});
		selected = new Set(selected);
	}
	function clearSelection() {
		selected = new Set();
	}

	async function transfer() {
		if (!fromWarehouseId || !toWarehouseId) {
			toastStore.error('Pilih gudang asal dan tujuan');
			return;
		}
		if (fromWarehouseId === toWarehouseId) {
			toastStore.error('Gudang asal dan tujuan tidak boleh sama');
			return;
		}
		if (selected.size === 0) {
			toastStore.error('Pilih minimal satu produk');
			return;
		}
		// Validasi qty tiap produk terpilih.
		const items: any[] = [];
		for (const pid of selected) {
			const qty = quantities[pid] || 1;
			if (qty < 1) {
				toastStore.error('Jumlah setiap produk minimal 1');
				return;
			}
			items.push({ product_id: pid, quantity: qty });
		}
		isTransferring = true;
		try {
			const res = await apiClient.post<{ movements_created: number; items_posted: number }>(
				API_ENDPOINTS.inventoryTransfer,
				{
					from_warehouse_id: fromWarehouseId,
					to_warehouse_id: toWarehouseId,
					reference_code: referenceCode,
					note,
					items
				}
			);
			if (!res.success || !res.data) throw new Error(res.error?.message || 'Gagal transfer');
			toastStore.success(`${res.data.items_posted} produk dipindahkan antar gudang`);
			selected = new Set();
			quantities = {};
			goto('/inventory/movements');
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal transfer');
		} finally {
			isTransferring = false;
		}
	}
</script>

<svelte:head>
	<title>Transfer Antar Gudang - Inventory</title>
</svelte:head>

<LayoutWrapper>
	<div class="page page-with-bottombar">
		<div class="flex items-center gap-3 mb-6">
			<button class="btn btn-ghost btn-sm" on:click={() => goto('/inventory')}>←</button>
			<h1 class="page-title">Transfer Antar Gudang</h1>
		</div>
		<p class="page-subtitle mb-4">
			Pindahkan stok dari satu gudang ke gudang lain. Stok asal berkurang, tujuan bertambah.
		</p>

		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-4">
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<div>
					<label class="label py-0" for="t-from"><span class="label-text text-xs text-base-content/60">Gudang asal *</span></label>
					<select id="t-from" class="select select-bordered w-full" bind:value={fromWarehouseId}>
						<option value="">-- Pilih gudang asal --</option>
						{#each warehouses as wh}
							<option value={wh.id}>{wh.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="label py-0" for="t-to"><span class="label-text text-xs text-base-content/60">Gudang tujuan *</span></label>
					<select id="t-to" class="select select-bordered w-full" bind:value={toWarehouseId}>
						<option value="">-- Pilih gudang tujuan --</option>
						{#each warehouses as wh}
							{#if wh.id !== fromWarehouseId}
								<option value={wh.id}>{wh.name}</option>
							{/if}
						{/each}
					</select>
				</div>
			</div>
		</div>

		<!-- Pilih Produk -->
		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-4">
			<div class="mb-3 flex items-center justify-between">
				<h2 class="font-semibold">Pilih Produk ({selected.size} dipilih)</h2>
				<div class="flex gap-2">
					<button class="btn btn-ghost btn-sm" on:click={selectAll}>Pilih Semua</button>
					<button class="btn btn-ghost btn-sm" on:click={clearSelection}>Kosongkan</button>
				</div>
			</div>

			{#if isLoading}
				<div class="flex h-40 items-center justify-center"><Loader size={48} /></div>
			{:else}
				<div class="max-h-96 overflow-y-auto rounded-xl bg-base-100 p-2">
					<div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
						{#each products as p}
							<div
								class="rounded-lg border border-base-300 p-3
								{selected.has(p.id) ? 'border-primary bg-primary/5' : ''}"
							>
								<label class="flex cursor-pointer items-center gap-3">
									<input
										type="checkbox"
										class="checkbox checkbox-primary checkbox-sm"
										checked={selected.has(p.id)}
										on:change={() => toggle(p.id)}
									/>
									<div class="min-w-0 flex-1">
										<p class="truncate text-sm font-medium">{p.name}</p>
										<p class="metric text-xs text-base-content/60">{p.product_code}</p>
									</div>
								</label>
								{#if selected.has(p.id)}
									<div class="mt-2 flex items-center gap-2">
										<span class="text-xs text-base-content/60">Jumlah:</span>
										<input
											type="number"
											min="1"
											class="input input-bordered input-xs w-20"
											bind:value={quantities[p.id]}
										/>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<button
				class="btn btn-primary mt-4 w-full"
				on:click={transfer}
				disabled={selected.size === 0 || !fromWarehouseId || !toWarehouseId || isTransferring}
			>
				{#if isTransferring}
					<span class="loading loading-spinner loading-sm"></span>
					Memproses...
				{:else}
					Transfer ({selected.size} produk)
				{/if}
			</button>
		</div>
	</div>
</LayoutWrapper>
