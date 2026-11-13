<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { toastStore } from '$lib/stores/toast';

	// Tiga tingkat: gudang -> lokasi/area -> box/rak.
	// Data asli dari API; tidak ada data contoh.

	type Warehouse = { id: string; code: string; name: string; address?: string; is_active?: boolean };

	let warehouses: Warehouse[] = [];
	let locations: Record<string, any[]> = {};
	let boxes: Record<string, any[]> = {};

	let isLoading = true;
	let loadError = '';

	// Gudang yang sedang dibuka (accordion).
	let openWarehouse = '';
	let openLocation = '';

	// Modal: 'warehouse' | 'location' | 'box' | null
	let modal: 'warehouse' | 'location' | 'box' | null = null;
	let isSaving = false;
	let formError = '';
	let form = { code: '', name: '', address: '' };

	// Reaktif: user baru terisi setelah checkAuth selesai (penting di mobile).
	$: isAdmin = $authStore.user?.role?.name === 'admin';

	onMount(async () => {
		await loadWarehouses();
	});

	async function loadWarehouses() {
		isLoading = true;
		loadError = '';
		try {
			const res = await apiClient.get<{ warehouses: Warehouse[] }>(API_ENDPOINTS.warehouses);
			if (res.success && res.data) warehouses = res.data.warehouses || [];
			else loadError = res.error?.message || 'Gagal memuat gudang';
		} catch (e: any) {
			loadError = e.message || 'Gagal memuat gudang';
		} finally {
			isLoading = false;
		}
	}

	// Lokasi dimuat saat gudang pertama kali dibuka, lalu di-cache.
	async function toggleWarehouse(id: string) {
		openWarehouse = openWarehouse === id ? '' : id;
		openLocation = '';
		if (openWarehouse) await fetchLocations(id);
	}

	async function fetchLocations(warehouseId: string) {
		try {
			const res = await apiClient.get<{ locations: any[] }>(
				API_ENDPOINTS.warehouseLocations(warehouseId)
			);
			locations[warehouseId] = res.success && res.data ? res.data.locations || [] : [];
		} catch {
			locations[warehouseId] = [];
		}
	}

	async function toggleLocation(id: string) {
		openLocation = openLocation === id ? '' : id;
		if (openLocation) await fetchBoxes(id);
	}

	async function fetchBoxes(locationId: string) {
		try {
			const res = await apiClient.get<{ boxes: any[] }>(API_ENDPOINTS.locationBoxes(locationId));
			boxes[locationId] = res.success && res.data ? res.data.boxes || [] : [];
		} catch {
			boxes[locationId] = [];
		}
	}

	function openCreate(kind: 'warehouse' | 'location' | 'box') {
		modal = kind;
		formError = '';
		form = { code: '', name: '', address: '' };
	}

	async function submit() {
		if (!modal) return;
		if (!form.code.trim() || !form.name.trim()) {
			formError = 'Kode dan nama wajib diisi';
			return;
		}

		isSaving = true;
		formError = '';
		try {
			let endpoint = API_ENDPOINTS.warehouses;
			if (modal === 'location') endpoint = API_ENDPOINTS.warehouseLocations(openWarehouse);
			if (modal === 'box') endpoint = API_ENDPOINTS.locationBoxes(openLocation);

			const body: any = { code: form.code.trim(), name: form.name.trim() };
			if (modal === 'warehouse') body.address = form.address.trim();

			const res = await apiClient.post(endpoint, body);
			if (!res.success) {
				formError = res.error?.message || 'Gagal menyimpan';
				return;
			}

			// Segarkan cache tingkat yang berubah saja.
			if (modal === 'warehouse') await loadWarehouses();
			if (modal === 'location') await fetchLocations(openWarehouse);
			if (modal === 'box') await fetchBoxes(openLocation);

			toastStore.success(
				modal === 'warehouse' ? 'Gudang ditambahkan' : modal === 'location' ? 'Lokasi ditambahkan' : 'Box ditambahkan'
			);
			modal = null;
		} catch (e: any) {
			formError = e.message || 'Gagal menyimpan';
		} finally {
			isSaving = false;
		}
	}
</script>

<svelte:head>
	<title>Gudang - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6 flex items-center justify-between gap-3">
			<div class="min-w-0">
				<h1 class="page-title">Gudang</h1>
				<p class="mt-1 text-sm text-base-content/60">
					Gudang, lokasi/area, dan rak penyimpanan
				</p>
			</div>
			{#if isAdmin}
				<button
					class="btn btn-primary btn-sm shrink-0 md:btn-md"
					on:click={() => openCreate('warehouse')}
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
					</svg>
					<span class="hidden sm:inline">Gudang baru</span>
					<span class="sm:hidden">Tambah</span>
				</button>
			{/if}
		</div>

		{#if isLoading}
			<div class="flex h-48 items-center justify-center">
				<Loader size={48} />
			</div>
		{:else if loadError}
			<div class="rounded-2xl border border-error/40 bg-error/10 p-4 text-sm text-error">
				{loadError}
				<button class="btn btn-ghost btn-xs ml-2" on:click={loadWarehouses}>Coba lagi</button>
			</div>
		{:else if warehouses.length === 0}
			<div class="flex flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-base-300 py-16 text-center">
				<p class="text-sm text-base-content/60">Belum ada gudang</p>
				{#if isAdmin}
					<button class="btn btn-primary btn-sm" on:click={() => openCreate('warehouse')}>
						Tambah gudang
					</button>
				{/if}
			</div>
		{:else}
			<ul class="space-y-3">
				{#each warehouses as wh}
					{@const isOpen = openWarehouse === wh.id}
					<li class="overflow-hidden rounded-2xl border border-base-300 bg-base-200">
						<button
							class="flex w-full items-center gap-3 p-4 text-left hover:bg-base-300/50"
							aria-expanded={isOpen}
							on:click={() => toggleWarehouse(wh.id)}
						>
							<svg
								class="h-4 w-4 shrink-0 text-base-content/40 transition-transform"
								class:rotate-90={isOpen}
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								stroke-width="2"
							>
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
							</svg>
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2">
									<span class="metric text-xs text-base-content/60">{wh.code}</span>
									<span class="truncate font-semibold">{wh.name}</span>
								</div>
								{#if wh.address}
									<div class="truncate text-xs text-base-content/60">{wh.address}</div>
								{/if}
							</div>
							{#if wh.is_active === false}
								<span class="badge badge-ghost badge-sm shrink-0">nonaktif</span>
							{/if}
						</button>

						{#if isOpen}
							<div class="border-t border-base-300 p-4">
								<div class="mb-3 flex items-center justify-between gap-2">
									<h3 class="text-xs font-medium text-base-content/60">Lokasi / area</h3>
									{#if isAdmin}
										<button
											class="btn btn-ghost btn-xs btn-square"
											title="Tambah lokasi"
											aria-label="Tambah lokasi"
											on:click={() => openCreate('location')}
										>
											<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
												<path stroke-linecap="round" d="M12 5v14M5 12h14" />
											</svg>
										</button>
									{/if}
								</div>

								{#if !locations[wh.id]}
									<Loader size={20} />
								{:else if locations[wh.id].length === 0}
									<p class="text-xs text-base-content/50">Belum ada lokasi.</p>
								{:else}
									<ul class="space-y-2">
										{#each locations[wh.id] as loc}
											{@const locOpen = openLocation === loc.id}
											<li class="rounded-xl border border-base-300 bg-base-100">
												<button
													class="flex w-full items-center gap-2 p-3 text-left text-sm hover:bg-base-300/40"
													aria-expanded={locOpen}
													on:click={() => toggleLocation(loc.id)}
												>
													<svg
														class="h-3.5 w-3.5 shrink-0 text-base-content/40 transition-transform"
														class:rotate-90={locOpen}
														fill="none"
														viewBox="0 0 24 24"
														stroke="currentColor"
														stroke-width="2"
													>
														<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
													</svg>
													<span class="metric text-xs text-base-content/50">{loc.code}</span>
													<span class="truncate">{loc.name}</span>
												</button>

												{#if locOpen}
													<div class="border-t border-base-300 p-3">
														<div class="mb-2 flex items-center justify-between gap-2">
															<h4 class="text-xs text-base-content/60">Rak / box</h4>
															{#if isAdmin}
																<button
																	class="btn btn-ghost btn-xs btn-square"
																	title="Tambah box"
																	aria-label="Tambah box"
																	on:click={() => openCreate('box')}
																>
																	<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
																		<path stroke-linecap="round" d="M12 5v14M5 12h14" />
																	</svg>
																</button>
															{/if}
														</div>

														{#if !boxes[loc.id]}
															<Loader size={14} />
														{:else if boxes[loc.id].length === 0}
															<p class="text-xs text-base-content/50">Belum ada rak.</p>
														{:else}
															<ul class="flex flex-wrap gap-2">
																{#each boxes[loc.id] as box}
																	<li class="rounded-lg border border-base-300 px-2 py-1 text-xs">
																		<span class="metric text-base-content/50">{box.code}</span>
																		{#if box.name}<span class="ml-1">{box.name}</span>{/if}
																	</li>
																{/each}
															</ul>
														{/if}
													</div>
												{/if}
											</li>
										{/each}
									</ul>
								{/if}
							</div>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</LayoutWrapper>

<!-- Modal tambah -->
{#if modal}
	<div class="modal modal-open" role="dialog" aria-modal="true">
		<div class="modal-box max-w-md">
			<h3 class="text-lg font-semibold">
				{modal === 'warehouse' ? 'Gudang baru' : modal === 'location' ? 'Lokasi baru' : 'Rak / box baru'}
			</h3>

			<form class="mt-4 space-y-3" on:submit|preventDefault={submit}>
				<div>
					<label class="label py-0" for="wh-code">
						<span class="label-text text-xs text-base-content/60">Kode</span>
					</label>
					<input
						id="wh-code"
						type="text"
						class="input input-bordered input-sm w-full"
						placeholder={modal === 'warehouse' ? 'mis. GDG-01' : modal === 'location' ? 'mis. A' : 'mis. A-01'}
						bind:value={form.code}
						required
					/>
				</div>

				<div>
					<label class="label py-0" for="wh-name">
						<span class="label-text text-xs text-base-content/60">Nama</span>
					</label>
					<input
						id="wh-name"
						type="text"
						class="input input-bordered input-sm w-full"
						placeholder={modal === 'warehouse' ? 'mis. Gudang Utama' : modal === 'location' ? 'mis. Area Elektronik' : 'mis. Rak Kabel'}
						bind:value={form.name}
						required
					/>
				</div>

				{#if modal === 'warehouse'}
					<div>
						<label class="label py-0" for="wh-address">
							<span class="label-text text-xs text-base-content/60">Alamat (opsional)</span>
						</label>
						<input
							id="wh-address"
							type="text"
							class="input input-bordered input-sm w-full"
							bind:value={form.address}
						/>
					</div>
				{/if}

				{#if formError}
					<div class="alert alert-error py-2 text-sm"><span>{formError}</span></div>
				{/if}

				<div class="modal-action">
					<button
						type="button"
						class="btn btn-sm btn-square"
						title="Batal"
						aria-label="Batal"
						on:click={() => (modal = null)}
					>
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M6 6l12 12M18 6L6 18" stroke-linecap="round" />
						</svg>
					</button>
					<button
						type="submit"
						class="btn btn-primary btn-sm btn-square"
						title="Simpan"
						aria-label="Simpan"
						disabled={isSaving}
					>
						{#if isSaving}
							<Loader size={18} />
						{:else}
							<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
								<path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round" />
							</svg>
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
	<button class="modal-backdrop" aria-label="Tutup" on:click={() => (modal = null)}></button>
{/if}
