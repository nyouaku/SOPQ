<script lang="ts">
	import { goto } from '$app/navigation';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { warehousesQuery, locationsQuery, boxesQuery } from '$lib/api/queries';
	import { queryKeys } from '$lib/api/queries';
	import { toastStore } from '$lib/stores/toast';

	const queryClient = useQueryClient();

	// Mode pergerakan: in / out / transfer.
	let mode: 'in' | 'out' | 'transfer' = 'in';

	const warehousesQ = warehousesQuery();
	$: warehouseList = ($warehousesQ.data as any[]) ?? [];

	// Gudang (asal) + lokasi + box untuk filter produk (opsional di mode in/out).
	let fromWarehouseId = '';
	let toWarehouseId = '';
	let locationId = '';
	let boxId = '';

	const locationsQ = locationsQuery(fromWarehouseId);
	$: locationList = ($locationsQ.data as any[]) ?? [];
	const boxesQ = boxesQuery(locationId);
	$: boxList = ($boxesQ.data as any[]) ?? [];

	// Baris item dinamis.
	type ItemRow = { product_id: string; product_name: string; quantity: number; unit_cost: number };
	let items: ItemRow[] = [{ product_id: '', product_name: '', quantity: 1, unit_cost: 0 }];

	let referenceCode = '';
	let note = '';
	let submitting = false;
	let errorMessage = '';
	let search = '';
	let searchResults: any[] = [];
	let searching = false;

	$: fromLabel = warehouseList.find((w) => w.id === fromWarehouseId)?.name || '(pilih gudang)';
	$: toLabel = warehouseList.find((w) => w.id === toWarehouseId)?.name || '(pilih gudang)';

	function switchMode(m: 'in' | 'out' | 'transfer') {
		mode = m;
		errorMessage = '';
	}

	function addItem() {
		items = [...items, { product_id: '', product_name: '', quantity: 1, unit_cost: 0 }];
	}

	function removeItem(i: number) {
		items = items.filter((_, idx) => idx !== i);
	}

	let searchTimer: ReturnType<typeof setTimeout> | null = null;
	let activeRow = -1;

	function onSearchInput(i: number) {
		activeRow = i;
		if (searchTimer) clearTimeout(searchTimer);
		const q = items[i]?.product_name ?? '';
		if (!q || q.length < 2) {
			searchResults = [];
			return;
		}
		searchTimer = setTimeout(() => doSearch(q), 300);
	}

	async function doSearch(q: string) {
		searching = true;
		try {
			const res = await apiClient.get<{ products: any[] }>(
				`${API_ENDPOINTS.products}?limit=10&search=${encodeURIComponent(q)}`
			);
			searchResults = res.success && res.data ? res.data.products || [] : [];
		} catch {
			searchResults = [];
		} finally {
			searching = false;
		}
	}

	function pickProduct(row: number, p: any) {
		items[row].product_id = p.id;
		items[row].product_name = p.name;
		searchResults = [];
	}

	function submit() {
		errorMessage = '';
		const validItems = items.filter((it) => it.product_id && it.quantity > 0);
		if (validItems.length === 0) {
			errorMessage = 'Isi minimal satu produk dengan jumlah > 0';
			return;
		}
		if (mode === 'transfer' && (!fromWarehouseId || !toWarehouseId)) {
			errorMessage = 'Pilih gudang asal dan tujuan';
			return;
		}
		if (mode !== 'transfer' && !fromWarehouseId) {
			errorMessage = 'Pilih gudang';
			return;
		}

		submitting = true;
		const payload = {
			sub_type: mode === 'in' ? 'purchase' : mode === 'out' ? 'sale' : 'transfer',
			reference_code: referenceCode,
			note,
			items: validItems.map((it) => ({
				product_id: it.product_id,
				quantity: it.quantity,
				unit_cost: it.unit_cost
			}))
		};

		let promise: Promise<any>;
		if (mode === 'transfer') {
			promise = apiClient.post<{ id: string; code: string }>(API_ENDPOINTS.inventoryTransfer, {
				...payload,
				from_warehouse_id: fromWarehouseId,
				to_warehouse_id: toWarehouseId
			});
		} else {
			const endpoint = mode === 'in' ? API_ENDPOINTS.inventoryIn : API_ENDPOINTS.inventoryOut;
			promise = apiClient.post<{ id: string; code: string }>(endpoint, {
				...payload,
				warehouse_id: fromWarehouseId
			});
		}

		promise
			.then((res) => {
				if (!res.success || !res.data) {
					throw new Error(res.error?.message || 'Gagal menyimpan pergerakan');
				}
				queryClient.invalidateQueries({ queryKey: ['inventory'] });
				toastStore.success(`Pergerakan ${res.data.code} tersimpan`);
				goto('/inventory');
			})
			.catch((e: any) => {
				errorMessage = e.message || 'Gagal menyimpan';
				toastStore.error(errorMessage);
			})
			.finally(() => {
				submitting = false;
			});
	}
</script>

<svelte:head>
	<title>Catat Pergerakan - Inventory</title>
</svelte:head>

<LayoutWrapper>
	<div class="page pb-24 md:pb-28">
		<div class="flex items-center gap-3 mb-6">
			<button class="btn btn-ghost btn-sm" on:click={() => goto('/inventory')}>←</button>
			<h1 class="page-title">Catat Pergerakan Stok</h1>
		</div>

		<!-- Mode segmented control -->
		<div class="mb-4 grid grid-cols-3 gap-1 rounded-xl bg-base-200 p-1" role="tablist" aria-label="Jenis pergerakan">
			<button
				role="tab"
				aria-selected={mode === 'in'}
				class="rounded-lg px-2 py-2 text-sm font-medium transition-colors {mode === 'in' ? 'bg-base-100 text-success shadow-sm' : 'opacity-50'}"
				on:click={() => switchMode('in')}
			>
				⬇️ Masuk (IN)
			</button>
			<button
				role="tab"
				aria-selected={mode === 'out'}
				class="rounded-lg px-2 py-2 text-sm font-medium transition-colors {mode === 'out' ? 'bg-base-100 text-error shadow-sm' : 'opacity-50'}"
				on:click={() => switchMode('out')}
			>
				⬆️ Keluar (OUT)
			</button>
			<button
				role="tab"
				aria-selected={mode === 'transfer'}
				class="rounded-lg px-2 py-2 text-sm font-medium transition-colors {mode === 'transfer' ? 'bg-base-100 text-info shadow-sm' : 'opacity-50'}"
				on:click={() => switchMode('transfer')}
			>
				🔄 Transfer
			</button>
		</div>

		<div class="card bg-base-200 shadow-xl">
			<div class="card-body gap-4">
				<!-- Gudang asal -->
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
					<div>
						<label class="label py-0" for="wh-from"><span class="label-text text-xs text-base-content/60">
							{mode === 'transfer' ? 'Gudang asal *' : 'Gudang *'}
						</span></label>
						<select id="wh-from" class="select select-bordered w-full" bind:value={fromWarehouseId}>
							<option value="">-- Pilih gudang --</option>
							{#each warehouseList as wh}
								<option value={wh.id}>{wh.name}</option>
							{/each}
						</select>
					</div>
					{#if mode === 'transfer'}
						<div>
							<label class="label py-0" for="wh-to"><span class="label-text text-xs text-base-content/60">Gudang tujuan *</span></label>
							<select id="wh-to" class="select select-bordered w-full" bind:value={toWarehouseId}>
								<option value="">-- Pilih gudang tujuan --</option>
								{#each warehouseList as wh}
									{#if wh.id !== fromWarehouseId}
										<option value={wh.id}>{wh.name}</option>
									{/if}
								{/each}
							</select>
						</div>
					{/if}
				</div>

				<!-- Arah pergerakan ringkas -->
				<div class="rounded-xl border border-base-300 bg-base-100 p-3 text-sm">
					{#if mode === 'in'}
						<span class="text-success">⬇️ Stok <b>naik</b> di gudang <b>{fromLabel}</b></span>
					{:else if mode === 'out'}
						<span class="text-error">⬆️ Stok <b>turun</b> di gudang <b>{fromLabel}</b></span>
					{:else}
						<span class="text-info">🔄 Stok <b>turun</b> di <b>{fromLabel}</b> → <b>naik</b> di <b>{toLabel}</b></span>
					{/if}
				</div>

				<!-- Items -->
				<div>
					<div class="mb-2 flex items-center justify-between">
						<span class="text-sm font-medium">Produk</span>
						<button class="btn btn-ghost btn-xs" on:click={addItem}>
							+ Tambah baris
						</button>
					</div>

					{#each items as it, i}
						<div class="mb-3 rounded-xl border border-base-300 bg-base-100 p-3">
							<div class="relative">
								<label class="label py-0" for="prod-{i}"><span class="label-text text-xs text-base-content/60">Produk</span></label>
								<input
									id="prod-{i}"
									class="input input-bordered input-sm w-full"
									placeholder="Ketik nama / kode produk..."
									bind:value={it.product_name}
									on:input={() => onSearchInput(i)}
								/>
								{#if activeRow === i && searchResults.length > 0}
									<div class="absolute z-30 mt-1 max-h-56 w-full overflow-y-auto rounded-xl border border-base-300 bg-base-100 shadow-lg">
										{#each searchResults as p}
											<button
												class="block w-full px-3 py-2 text-left text-sm hover:bg-base-200"
												on:click={() => pickProduct(i, p)}
											>
												<div class="truncate font-medium">{p.name}</div>
												<div class="metric text-xs text-base-content/60">{p.product_code}</div>
											</button>
										{/each}
									</div>
								{/if}
							</div>

							<div class="mt-2 grid grid-cols-2 gap-2">
								<div>
									<label class="label py-0" for="qty-{i}"><span class="label-text text-xs text-base-content/60">Jumlah</span></label>
									<input id="qty-{i}" type="number" min="1" class="input input-bordered input-sm w-full" bind:value={it.quantity} />
								</div>
								<div>
									<label class="label py-0" for="cost-{i}"><span class="label-text text-xs text-base-content/60">Harga satuan</span></label>
									<input id="cost-{i}" type="number" min="0" step="0.01" class="input input-bordered input-sm w-full" bind:value={it.unit_cost} />
								</div>
							</div>

							{#if items.length > 1}
								<button class="btn btn-ghost btn-xs mt-2 text-error" on:click={() => removeItem(i)}>Hapus baris</button>
							{/if}
						</div>
					{/each}
				</div>

				<!-- Reference + note -->
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
					<div>
						<label class="label py-0" for="ref"><span class="label-text text-xs text-base-content/60">No. referensi (opsional)</span></label>
						<input id="ref" class="input input-bordered w-full" placeholder="mis. PO-2026-001" bind:value={referenceCode} />
					</div>
					<div>
						<label class="label py-0" for="note"><span class="label-text text-xs text-base-content/60">Catatan (opsional)</span></label>
						<input id="note" class="input input-bordered w-full" bind:value={note} />
					</div>
				</div>

				{#if errorMessage}
					<div class="alert alert-error"><span>{errorMessage}</span></div>
				{/if}

				<div class="card-actions justify-end mt-2">
					<button class="btn btn-ghost" on:click={() => goto('/inventory')}>Batal</button>
					<button class="btn btn-primary" disabled={submitting} on:click={submit}>
						{#if submitting}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						Simpan pergerakan
					</button>
				</div>
			</div>
		</div>
	</div>
</LayoutWrapper>
