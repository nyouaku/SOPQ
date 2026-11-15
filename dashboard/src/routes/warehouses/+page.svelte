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
	let editingId: string | null = null; // untuk edit mode
	let deleteConfirm: { kind: 'warehouse' | 'location' | 'box'; id: string; name: string } | null = null;

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
		editingId = null;
		formError = '';
		form = { code: '', name: '', address: '' };
	}

	function openEdit(kind: 'warehouse' | 'location' | 'box', id: string, item: any) {
		modal = kind;
		editingId = id;
		formError = '';
		form = { code: item.code, name: item.name, address: item.address || '' };
	}

	function openDelete(kind: 'warehouse' | 'location' | 'box', id: string, name: string) {
		deleteConfirm = { kind, id, name };
	}

	async function confirmDelete() {
		if (!deleteConfirm) return;
		const { kind, id } = deleteConfirm;
		isSaving = true;

		try {
			let endpoint = `${API_ENDPOINTS.warehouses}/${id}`;
			if (kind === 'location') endpoint = `${API_ENDPOINTS.warehouseLocations(openWarehouse)}/${id}`;
			if (kind === 'box') endpoint = `${API_ENDPOINTS.locationBoxes(openLocation)}/${id}`;

			const res = await apiClient.delete(endpoint);
			if (!res.success) {
				toastStore.error(res.error?.message || 'Gagal hapus');
				return;
			}

			if (kind === 'warehouse') await loadWarehouses();
			if (kind === 'location') await fetchLocations(openWarehouse);
			if (kind === 'box') await fetchBoxes(openLocation);

			toastStore.success('Dihapus');
			deleteConfirm = null;
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal hapus');
		} finally {
			isSaving = false;
		}
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

			let res;
			if (editingId) {
				// Edit: PUT
				endpoint = `${endpoint}/${editingId}`;
				res = await apiClient.put(endpoint, body);
			} else {
				// Create: POST
				res = await apiClient.post(endpoint, body);
			}

			if (!res.success) {
				formError = res.error?.message || 'Gagal menyimpan';
				return;
			}

			// Segarkan cache tingkat yang berubah saja.
			if (modal === 'warehouse') await loadWarehouses();
			if (modal === 'location') await fetchLocations(openWarehouse);
			if (modal === 'box') await fetchBoxes(openLocation);

			toastStore.success(
				editingId
					? modal === 'warehouse'
						? 'Gudang diperbarui'
						: modal === 'location'
						? 'Lokasi diperbarui'
						: 'Box diperbarui'
					: modal === 'warehouse'
					? 'Gudang ditambahkan'
					: modal === 'location'
					? 'Lokasi ditambahkan'
					: 'Box ditambahkan'
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
							{#if isAdmin}
								<div class="flex shrink-0 gap-1">
									<a
										href="/warehouses/{wh.id}/layout"
										class="btn btn-ghost btn-xs btn-square"
										title="Layout Editor"
										aria-label="Edit layout gudang"
									>
										<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
										</svg>
									</a>
									<button
										class="btn btn-ghost btn-xs btn-square"
										title="Edit"
										aria-label="Edit gudang"
										on:click|stopPropagation={() => openEdit('warehouse', wh.id, wh)}
									>
										<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M4 20h4l10-10a2.5 2.5 0 0 0-3.5-3.5L4.5 16.5 4 20Z" />
										</svg>
									</button>
									<button
										class="btn btn-ghost btn-xs btn-square text-error"
										title="Hapus"
										aria-label="Hapus gudang"
										on:click|stopPropagation={() => openDelete('warehouse', wh.id, wh.name)}
									>
										<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
											<path stroke-linecap="round" stroke-linejoin="round" d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13" />
										</svg>
									</button>
								</div>
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
													{#if isAdmin}
														<div class="ml-auto flex shrink-0 gap-1">
															<button
																class="btn btn-ghost btn-xs btn-square"
																title="Edit"
																aria-label="Edit lokasi"
																on:click|stopPropagation={() => openEdit('location', loc.id, loc)}
															>
																<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
																	<path stroke-linecap="round" stroke-linejoin="round" d="M4 20h4l10-10a2.5 2.5 0 0 0-3.5-3.5L4.5 16.5 4 20Z" />
																</svg>
															</button>
															<button
																class="btn btn-ghost btn-xs btn-square text-error"
																title="Hapus"
																aria-label="Hapus lokasi"
																on:click|stopPropagation={() => openDelete('location', loc.id, loc.name)}
															>
																<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
																	<path stroke-linecap="round" stroke-linejoin="round" d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13" />
																</svg>
															</button>
														</div>
													{/if}
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
															<ul class="space-y-1">
																{#each boxes[loc.id] as box}
																	<li class="flex items-center justify-between gap-2 rounded-lg border border-base-300 px-2 py-1.5 text-xs">
																		<span>
																			<span class="metric text-base-content/50">{box.code}</span>
																			{#if box.name}<span class="ml-1">{box.name}</span>{/if}
																		</span>
																		{#if isAdmin}
																			<div class="flex shrink-0 gap-1">
																				<button
																					class="btn btn-ghost btn-xs btn-square"
																					title="Edit"
																					aria-label="Edit box"
																					on:click={() => openEdit('box', box.id, box)}
																				>
																					<svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
																						<path stroke-linecap="round" stroke-linejoin="round" d="M4 20h4l10-10a2.5 2.5 0 0 0-3.5-3.5L4.5 16.5 4 20Z" />
																					</svg>
																				</button>
																				<button
																					class="btn btn-ghost btn-xs btn-square text-error"
																					title="Hapus"
																					aria-label="Hapus box"
																					on:click={() => openDelete('box', box.id, box.code)}
																				>
																					<svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
																						<path stroke-linecap="round" stroke-linejoin="round" d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13" />
																					</svg>
																				</button>
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
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</LayoutWrapper>

<!-- Modal tambah/edit -->
{#if modal}
	<div class="modal modal-open" role="dialog" aria-modal="true">
		<div class="modal-box max-w-md">
			<h3 class="text-lg font-semibold">
				{editingId ? 'Edit' : 'Tambah'}
				{modal === 'warehouse' ? 'Gudang' : modal === 'location' ? 'Lokasi' : 'Rak / Box'}
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
						class="btn btn-ghost btn-sm"
						title="Batal"
						aria-label="Batal"
						on:click={() => (modal = null)}
					>
						Batal
					</button>
					<button
						type="submit"
						class="btn btn-primary btn-sm"
						title={editingId ? 'Update' : 'Simpan'}
						aria-label={editingId ? 'Update' : 'Simpan'}
						disabled={isSaving}
					>
						{#if isSaving}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						{editingId ? 'Update' : 'Simpan'}
					</button>
				</div>
			</form>
		</div>
	</div>
	<button class="modal-backdrop" aria-label="Tutup" on:click={() => (modal = null)}></button>
{/if}

<!-- Delete confirmation dialog -->
{#if deleteConfirm}
	<div class="modal modal-open" role="dialog" aria-modal="true">
		<div class="modal-box max-w-sm">
			<h3 class="text-lg font-semibold">Hapus?</h3>
			<p class="mt-3 text-sm">
				Yakin ingin menghapus <strong>{deleteConfirm.name}</strong>?
			</p>
			{#if deleteConfirm.kind === 'warehouse'}
				<p class="mt-1 text-xs text-base-content/60">Semua lokasi dan rak di gudang ini juga akan dihapus.</p>
			{:else if deleteConfirm.kind === 'location'}
				<p class="mt-1 text-xs text-base-content/60">Semua rak di lokasi ini juga akan dihapus.</p>
			{/if}

			<div class="modal-action mt-6">
				<button
					class="btn btn-ghost btn-sm"
					on:click={() => (deleteConfirm = null)}
					disabled={isSaving}
				>
					Batal
				</button>
				<button
					class="btn btn-error btn-sm"
					on:click={confirmDelete}
					disabled={isSaving}
				>
					{#if isSaving}
						<span class="loading loading-spinner loading-sm"></span>
					{/if}
					Hapus
				</button>
			</div>
		</div>
	</div>
	<button class="modal-backdrop" aria-label="Tutup" on:click={() => (deleteConfirm = null)}></button>
{/if}
