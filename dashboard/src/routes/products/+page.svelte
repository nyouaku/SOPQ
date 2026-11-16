<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import ProductQr from '$lib/components/ProductQr.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import TableSkeleton from '$lib/components/skeletons/TableSkeleton.svelte';
	import { toastStore } from '$lib/stores/toast';
	import { productStock } from '$lib/utils/stock';

	let products: any[] = [];
	let isLoading = true;
	let isExportingExcel = false;
	let isSaving = false;
	let currentPage = 1;
	let totalPages = 1;
	let searchQuery = '';
	let showModal = false;
	let editingProduct: any = null;
	// QR popup (kolom QR) & detail popup (aksi Detail) — data asli, bukan dummy.
	let qrProduct: any = null;
	let detailProduct: any = null;
	let detailQr: any = null;
	let qrQr: any = null;

	// Import produk via template Excel.
	let showImportModal = false;
	let importFile: File | null = null;
	let isImporting = false;
	let importResult: any = null;
	let importError = '';

	// QR meng-encode KODE produk (product_code).
	function qrValue(product: any): string {
		return product?.product_code || '';
	}

	async function downloadImportTemplate() {
		try {
			const token = localStorage.getItem('access_token');
			const res = await fetch(`${API_ENDPOINTS.base}${API_ENDPOINTS.importProductTemplate}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!res.ok) throw new Error('Gagal unduh template');
			const blob = await res.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'template-import-produk.xlsx';
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal unduh template');
		}
	}

	function onImportFile(event: Event) {
		const target = event.target as HTMLInputElement;
		importFile = target.files?.[0] || null;
		importResult = null;
		importError = '';
	}

	async function handleImport() {
		if (!importFile) return;
		isImporting = true;
		importError = '';
		importResult = null;
		toastStore.info('Mengunggah dan memproses file...');
		try {
			const token = localStorage.getItem('access_token');
			const form = new FormData();
			form.append('file', importFile);
			const res = await fetch(`${API_ENDPOINTS.base}${API_ENDPOINTS.importProducts}`, {
				method: 'POST',
				headers: { Authorization: `Bearer ${token}` },
				body: form
			});
			const text = await res.text();
			const json = text ? JSON.parse(text) : { success: res.ok };
			if (!res.ok || !json.success) {
				importError = json.error?.message || 'Import gagal';
				toastStore.error(importError);
			} else {
				importResult = json.data;
				await loadProducts();
				const { created = 0, updated = 0, failed = 0 } = importResult || {};
				if (failed > 0) {
					toastStore.error(`Import selesai: ${created} dibuat, ${updated} diperbarui, ${failed} gagal`);
				} else {
					toastStore.success(`Import berhasil: ${created} dibuat, ${updated} diperbarui`);
				}
			}
		} catch (e: any) {
			importError = e.message || 'Import gagal';
			toastStore.error(importError);
		} finally {
			isImporting = false;
		}
	}

	let formData = {
		product_code: '',
		name: '',
		brand: '',
		unit: '',
		minimum_stock: 0,
		box_id: ''
	};

	// --- Cascading pilihan Gudang → Lokasi → Box ---
	// Produk hanya menyimpan box_id; nama gudang diturunkan dari box.location.warehouse.
	let warehouses: any[] = [];
	let formLocations: any[] = [];
	let formBoxes: any[] = [];
	let formWarehouseId = '';
	let formLocationId = '';

	async function loadWarehouses() {
		try {
			const res = await apiClient.get<{ warehouses: any[] }>(API_ENDPOINTS.warehouses);
			warehouses = res.success && res.data ? res.data.warehouses || [] : [];
		} catch {
			warehouses = [];
		}
	}

	async function onWarehouseChange() {
		formLocationId = '';
		formData.box_id = '';
		formBoxes = [];
		formLocations = [];
		if (!formWarehouseId) return;
		try {
			const res = await apiClient.get<{ locations: any[] }>(
				API_ENDPOINTS.warehouseLocations(formWarehouseId)
			);
			formLocations = res.success && res.data ? res.data.locations || [] : [];
		} catch {
			formLocations = [];
		}
	}

	async function onLocationChange() {
		formData.box_id = '';
		formBoxes = [];
		if (!formLocationId) return;
		try {
			const res = await apiClient.get<{ boxes: any[] }>(API_ENDPOINTS.locationBoxes(formLocationId));
			formBoxes = res.success && res.data ? res.data.boxes || [] : [];
		} catch {
			formBoxes = [];
		}
	}

	// Isi ulang cascading dari produk yang diedit (box_id → lokasi → gudang).
	async function syncWarehouseFrom(product: any) {
		formWarehouseId = product?.box?.location?.warehouse?.id || '';
		formLocationId = '';
		formLocations = [];
		formBoxes = [];
		if (formWarehouseId) {
			await onWarehouseChange();
			formLocationId = product?.box?.location?.id || '';
			if (formLocationId) await onLocationChange();
		}
		// onWarehouseChange/onLocationChange mengosongkan box_id; kembalikan nilai asli.
		formData.box_id = product?.box?.id || '';
	}

	function warehouseName(product: any): string {
		return product?.box?.location?.warehouse?.name || '-';
	}

	onMount(async () => {
		await loadProducts();
		await loadWarehouses();
	});

	async function loadProducts() {
		isLoading = true;
		try {
			const response = await apiClient.get<{ products: any[]; total: number }>(
				`${API_ENDPOINTS.products}?page=${currentPage}&limit=20&search=${searchQuery}`
			);
			if (response.success && response.data) {
				products = response.data.products || [];
				totalPages = Math.ceil((response.data.total || 0) / 20);
			}
		} catch (error) {
			console.error('Failed to load products:', error);
		} finally {
			isLoading = false;
		}
	}

	function handleSearch() {
		currentPage = 1;
		loadProducts();
	}

	function handlePageChange(page: number) {
		currentPage = page;
		loadProducts();
	}

	async function handleExportExcel() {
		isExportingExcel = true;
		try {
			const token = localStorage.getItem('access_token');
			const response = await fetch(`${API_ENDPOINTS.base}/export/products`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});

			if (response.ok) {
				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `products_${new Date().toISOString().split('T')[0]}.xlsx`;
				document.body.appendChild(a);
				a.click();
				window.URL.revokeObjectURL(url);
				document.body.removeChild(a);
			} else {
				toastStore.error('Gagal export data');
			}
		} catch (error) {
			console.error('Failed to export Excel:', error);
			toastStore.error('Gagal export data');
		} finally {
			isExportingExcel = false;
		}
	}

	function openCreateModal() {
		editingProduct = null;
		formData = {
			product_code: '',
			name: '',
			brand: '',
			unit: 'pcs',
			minimum_stock: 0,
			box_id: ''
		};
		formWarehouseId = '';
		formLocationId = '';
		formLocations = [];
		formBoxes = [];
		if (!warehouses.length) loadWarehouses();
		showModal = true;
	}

	function openEditModal(product: any) {
		editingProduct = product;
		formData = {
			product_code: product.product_code || '',
			name: product.name || '',
			brand: product.brand || '',
			unit: product.unit || 'pcs',
			minimum_stock: product.minimum_stock || 0,
			box_id: product.box_id || ''
		};
		if (!warehouses.length) loadWarehouses();
		syncWarehouseFrom(product);
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		editingProduct = null;
		formData = {
			product_code: '',
			name: '',
			brand: '',
			unit: '',
			minimum_stock: 0,
			box_id: ''
		};
		formWarehouseId = '';
		formLocationId = '';
		formLocations = [];
		formBoxes = [];
	}

	async function handleSubmit() {
		if (!formData.name || !formData.product_code) {
			toastStore.error('Nama dan Kode Produk wajib diisi');
			return;
		}

		isSaving = true;
		try {
			// box_id kosong harus null, bukan "" — backend mem-parse UUID.
			const payload = {
				...formData,
				box_id: formData.box_id || null,
			};
			let response;
			if (editingProduct) {
				// Update
				response = await apiClient.put(
					`${API_ENDPOINTS.products}/${editingProduct.id}`,
					payload
				);
			} else {
				// Create
				response = await apiClient.post(API_ENDPOINTS.products, payload);
			}

			if (response.success) {
				toastStore.success(editingProduct ? 'Produk berhasil diupdate' : 'Produk berhasil ditambahkan');
				closeModal();
				await loadProducts();
			} else {
				toastStore.error(response.error?.message || 'Gagal menyimpan produk');
			}
		} catch (error: any) {
			console.error('Failed to save product:', error);
			toastStore.error(error.message || 'Gagal menyimpan produk');
		} finally {
			isSaving = false;
		}
	}

	async function handleDelete(id: string) {
		if (!confirm('Yakin ingin menghapus produk ini?')) return;

		try {
			const response = await apiClient.delete(`${API_ENDPOINTS.products}/${id}`);
			if (response.success) {
				toastStore.success('Produk dihapus');
				await loadProducts();
			}
		} catch (error) {
			toastStore.error('Gagal menghapus produk');
		}
	}
</script>

<svelte:head>
	<title>Products - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6 flex items-center justify-between gap-3">
			<h1 class="page-title min-w-0 truncate">Products</h1>
			{#if $authStore.user?.role?.name === 'admin'}
				<div class="flex shrink-0 gap-2">
					<button
						class="btn btn-outline btn-sm btn-square md:btn-md"
						title="Import dari Excel"
						aria-label="Import produk dari Excel"
						on:click={() => (showImportModal = true)}
					>
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2" />
						</svg>
					</button>
					<button
						class="btn btn-primary btn-sm btn-square md:btn-md"
						title="Tambah produk"
						aria-label="Tambah produk"
						on:click={openCreateModal}
					>
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
							<path stroke-linecap="round" d="M12 4v16M4 12h16" />
						</svg>
					</button>
				</div>
			{/if}
		</div>

		<!-- Search -->
		<div class="mb-6">
			<div class="flex flex-wrap sm:flex-nowrap gap-2 sm:gap-4">
				<div class="join flex-1 min-w-0">
					<input
						class="input input-bordered join-item flex-1 min-w-0"
						placeholder="Cari produk (nama, kode)..."
						bind:value={searchQuery}
						on:keypress={(e) => e.key === 'Enter' && handleSearch()}
					/>
					<button
						class="btn btn-primary join-item btn-square"
						title="Cari"
						aria-label="Cari produk"
						on:click={handleSearch}
					>
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="11" cy="11" r="7" />
							<path d="M16.5 16.5L21 21" stroke-linecap="round" />
						</svg>
					</button>
				</div>
				<button
					class="btn btn-success btn-square shrink-0"
					title="Export Excel"
					aria-label="Export produk ke Excel"
					on:click={handleExportExcel}
					disabled={isExportingExcel}
				>
					{#if isExportingExcel}
						<Loader size={20} />
					{:else}
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 21V9m0 0l-4 4m4-4l4 4M4 5h16" />
						</svg>
					{/if}
				</button>
			</div>
		</div>

		<!-- Products Table -->
		<div class="card bg-base-200 shadow-xl">
			<div class="card-body">
				{#if isLoading}
					<TableSkeleton rows={8} columns={6} />
				{:else}
					<div class="overflow-x-auto">
						<table class="table">
							<thead>
								<tr>
									<th>Kode</th>
									<th>Nama</th>
									<th>QR Code</th>
									<th>Gudang</th>
									<th>Stock Sistem</th>
									<th class="text-right">Aksi</th>
								</tr>
							</thead>
							<tbody>
								{#if products.length > 0}
									{#each products as product}
										<tr>
											<td>{product.product_code}</td>
											<td>{product.name}</td>
											<td>
												{#if qrValue(product)}
													<button
														class="btn btn-ghost btn-sm btn-square"
														title="Lihat QR code"
														aria-label="Lihat QR code produk"
														on:click={() => (qrProduct = product)}
													>
														<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<rect x="3" y="3" width="7" height="7" rx="1"/>
															<rect x="14" y="3" width="7" height="7" rx="1"/>
															<rect x="3" y="14" width="7" height="7" rx="1"/>
															<path d="M14 14h3v3h-3zM20 14h1M14 20h1M18 20h3v-1"/>
														</svg>
													</button>
												{:else}
													<span class="text-base-content/40">-</span>
												{/if}
											</td>
											<td>{warehouseName(product)}</td>
											<td class="metric">{productStock(product)}</td>
											<td>
												<div class="flex justify-end gap-1">
													<a
														href="/products/{product.id}"
														class="btn btn-ghost btn-sm btn-square"
														title="Riwayat stok & grafik"
														aria-label="Riwayat stok produk"
													>
														<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<path stroke-linecap="round" stroke-linejoin="round" d="M3 13h2v5H3zM8 8h2v10H8zM13 4h2v14h-2zM18 10h2v8h-2z" />
														</svg>
													</a>
													<button
														class="btn btn-ghost btn-sm btn-square"
														title="Detail"
														aria-label="Detail produk"
														on:click={() => (detailProduct = product)}
													>
														<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<circle cx="12" cy="12" r="9" />
															<path d="M12 11v5M12 7.5v.5" stroke-linecap="round" />
														</svg>
													</button>
													{#if $authStore.user?.role?.name === 'admin'}
														<button
															class="btn btn-ghost btn-sm btn-square"
															title="Edit"
															aria-label="Edit produk"
															on:click={() => openEditModal(product)}
														>
															<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<path d="M4 20h4l10-10a2.5 2.5 0 0 0-3.5-3.5L4.5 16.5 4 20Z" stroke-linejoin="round" />
															</svg>
														</button>
														<button
															class="btn btn-ghost btn-sm btn-square text-error"
															title="Hapus"
															aria-label="Hapus produk"
															on:click={() => handleDelete(product.id)}
														>
															<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<path d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13" stroke-linecap="round" stroke-linejoin="round" />
															</svg>
														</button>
													{/if}
												</div>
											</td>
										</tr>
									{/each}
								{:else}
									<tr>
										<td colspan="7" class="text-center text-base-content/50">
											Tidak ada data produk
										</td>
									</tr>
								{/if}
							</tbody>
						</table>
					</div>

					<!-- Pagination -->
					{#if totalPages > 1}
						<div class="flex justify-center mt-4">
							<div class="join">
								<button
									class="join-item btn"
									disabled={currentPage === 1}
									on:click={() => handlePageChange(currentPage - 1)}
								>
									«
								</button>
								<button class="join-item btn">Page {currentPage} of {totalPages}</button>
								<button
									class="join-item btn"
									disabled={currentPage === totalPages}
									on:click={() => handlePageChange(currentPage + 1)}
								>
									»
								</button>
							</div>
						</div>
					{/if}
				{/if}
			</div>
		</div>
	</div>
</LayoutWrapper>

<!-- Modal placeholder - simplified for PHASE 7 -->
{#if showModal}
	<div class="modal modal-open">
		<div class="modal-box w-full max-w-sm sm:max-w-lg p-4 sm:p-6">
			<h3 class="font-bold text-lg">{editingProduct ? 'Edit' : 'Tambah'} Produk</h3>

			<form on:submit|preventDefault={handleSubmit} class="mt-6 space-y-3 sm:space-y-4">
				<!-- Kode Produk -->
				<div class="form-control">
					<label class="label" for="p-code">
						<span class="label-text font-medium text-sm sm:text-base">Kode Produk <span class="text-error">*</span></span>
					</label>
					<input
						id="p-code"
						type="text"
						class="input input-bordered input-sm sm:input-md"
						bind:value={formData.product_code}
						placeholder="EL-001"
						required
					/>
				</div>

				<!-- Nama Produk -->
				<div class="form-control">
					<label class="label" for="p-name">
						<span class="label-text font-medium text-sm sm:text-base">Nama Produk <span class="text-error">*</span></span>
					</label>
					<input
						id="p-name"
						type="text"
						class="input input-bordered input-sm sm:input-md"
						bind:value={formData.name}
						placeholder="Kabel NYA 2.5mm"
						required
					/>
				</div>

				<!-- Satuan & Minimum Stock (stack di mobile, side by side di tablet) -->
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-2 sm:gap-3">
					<div class="form-control">
						<label class="label" for="p-unit">
							<span class="label-text font-medium text-sm sm:text-base">Satuan</span>
						</label>
						<select id="p-unit" class="select select-bordered select-xs sm:select-sm" bind:value={formData.unit}>
							<option value="pcs">pcs</option>
							<option value="box">box</option>
							<option value="meter">meter</option>
							<option value="kg">kg</option>
							<option value="liter">liter</option>
							<option value="unit">unit</option>
						</select>
					</div>
					<div class="form-control">
						<label class="label" for="p-min">
							<span class="label-text font-medium text-sm sm:text-base">Min Stok</span>
						</label>
						<input
							id="p-min"
							type="number"
							class="input input-bordered input-xs sm:input-sm"
							bind:value={formData.minimum_stock}
							min="0"
							placeholder="0"
						/>
					</div>
				</div>

				<!-- Divider -->
				<div class="divider my-2"></div>

				<!-- Gudang -->
				<div class="form-control">
					<label class="label" for="p-warehouse">
						<span class="label-text font-medium text-sm sm:text-base">Gudang</span>
					</label>
					<select
						id="p-warehouse"
						class="select select-bordered select-xs sm:select-sm"
						bind:value={formWarehouseId}
						on:change={onWarehouseChange}
					>
						<option value="">-- Pilih gudang --</option>
						{#each warehouses as wh}
							<option value={wh.id}>{wh.name}</option>
						{/each}
					</select>
				</div>

				<!-- Lokasi -->
				<div class="form-control">
					<label class="label" for="p-location">
						<span class="label-text font-medium text-sm sm:text-base">Lokasi / Area</span>
					</label>
					<select
						id="p-location"
						class="select select-bordered select-xs sm:select-sm"
						bind:value={formLocationId}
						on:change={onLocationChange}
						disabled={!formWarehouseId}
					>
						<option value="">-- {formWarehouseId ? 'Pilih lokasi' : 'Pilih gudang dulu'} --</option>
						{#each formLocations as loc}
							<option value={loc.id}>{loc.name}</option>
						{/each}
					</select>
				</div>

				<!-- Rak / Box -->
				<div class="form-control">
					<label class="label" for="p-box">
						<span class="label-text font-medium text-sm sm:text-base">Rak / Box</span>
					</label>
					<select
						id="p-box"
						class="select select-bordered select-xs sm:select-sm"
						bind:value={formData.box_id}
						disabled={!formLocationId}
					>
						<option value="">-- {formLocationId ? 'Pilih rak' : 'Pilih lokasi dulu'} --</option>
						{#each formBoxes as box}
							<option value={box.id}>{box.code}{box.name ? ` — ${box.name}` : ''}</option>
						{/each}
					</select>
				</div>

				<!-- Stok Sistem (read-only) -->
				{#if editingProduct}
					<div class="form-control">
						<label class="label" for="p-stock">
							<span class="label-text font-medium text-sm sm:text-base">Stok Sistem</span>
						</label>
						<input
							id="p-stock"
							type="text"
							class="input input-bordered input-xs sm:input-sm bg-base-300"
							value={productStock(editingProduct)}
							readonly
							tabindex="-1"
						/>
						<label class="label py-1" for="p-stock">
							<span class="label-text-alt text-xs">Otomatis dari stock opname</span>
						</label>
					</div>
				{/if}

				<!-- Action Buttons -->
				<div class="modal-action mt-6 gap-2">
					<button
						type="button"
						class="btn btn-ghost btn-sm"
						on:click={closeModal}
						disabled={isSaving}
					>
						Batal
					</button>
					<button
						type="submit"
						class="btn btn-primary btn-sm"
						disabled={isSaving}
					>
						{#if isSaving}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						{editingProduct ? 'Update' : 'Simpan'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- QR Code Popup (kolom QR di tabel) -->
{#if qrProduct}
	<div class="modal modal-open" role="dialog" aria-modal="true" aria-label="QR code produk">
		<div class="modal-box max-w-sm">
			<h3 class="text-lg font-semibold">QR Code Produk</h3>
			<p class="mt-1 truncate text-sm text-base-content/60">{qrProduct.name}</p>
			<div class="mt-4">
				<ProductQr
					bind:this={qrQr}
					value={qrValue(qrProduct)}
					productCode={qrProduct.product_code}
					productName={qrProduct.name}
					size={220}
				/>
			</div>
			<div class="modal-action">
				<button
					class="btn btn-sm btn-square"
					title="Tutup"
					aria-label="Tutup"
					on:click={() => (qrProduct = null)}
				>
					<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M6 6l12 12M18 6L6 18" stroke-linecap="round" />
					</svg>
				</button>
				<button
					class="btn btn-primary btn-sm btn-square"
					title="Unduh label QR"
					aria-label="Unduh label QR"
					on:click={() => qrQr?.downloadLabel()}
				>
					<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2" />
					</svg>
				</button>
			</div>
		</div>
	</div>
	<button class="modal-backdrop" aria-label="Tutup QR code" on:click={() => (qrProduct = null)}></button>
{/if}

<!-- Detail Popup (aksi Detail di tabel) -->
{#if detailProduct}
	<div class="modal modal-open" role="dialog" aria-modal="true" aria-label="Detail produk">
		<div class="modal-box max-w-2xl">
			<h3 class="text-lg font-semibold">Detail Produk</h3>
			<p class="metric mt-1 text-sm text-base-content/60">{detailProduct.product_code}</p>

			<div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-[auto_1fr]">
				<div>
					<ProductQr
						bind:this={detailQr}
						value={qrValue(detailProduct)}
						productCode={detailProduct.product_code}
						productName={detailProduct.name}
						size={180}
					/>
				</div>

				<dl class="grid grid-cols-1 gap-x-6 gap-y-3 text-sm sm:grid-cols-2">
					<div class="sm:col-span-2">
						<dt class="text-base-content/60">Nama</dt>
						<dd class="font-medium">{detailProduct.name || '-'}</dd>
					</div>
					<div>
						<dt class="text-base-content/60">Gudang</dt>
						<dd>{warehouseName(detailProduct)}</dd>
					</div>
					<div>
						<dt class="text-base-content/60">Satuan</dt>
						<dd>{detailProduct.unit || '-'}</dd>
					</div>
					<div>
						<dt class="text-base-content/60">Stok sistem</dt>
						<dd class="metric">{productStock(detailProduct)}</dd>
					</div>
					<div>
						<dt class="text-base-content/60">Minimum stok</dt>
						<dd class="metric">{detailProduct.minimum_stock ?? 0}</dd>
					</div>
				</dl>
			</div>

			<div class="modal-action">
				<button
					class="btn btn-sm btn-square"
					title="Tutup"
					aria-label="Tutup"
					on:click={() => (detailProduct = null)}
				>
					<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M6 6l12 12M18 6L6 18" stroke-linecap="round" />
					</svg>
				</button>
				<button
					class="btn btn-primary btn-sm btn-square"
					title="Download label QR"
					aria-label="Download label QR"
					disabled={!qrValue(detailProduct)}
					on:click={() => detailQr?.downloadLabel()}
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2"
						/>
					</svg>
				</button>
			</div>
		</div>
	</div>
	<button
		class="modal-backdrop"
		aria-label="Tutup detail produk"
		on:click={() => (detailProduct = null)}
	></button>
{/if}

<!-- Import Produk (template Excel) -->
{#if showImportModal}
	<div class="modal modal-open" role="dialog" aria-modal="true" aria-label="Import produk">
		<div class="modal-box">
			<h3 class="text-lg font-semibold">Import Produk</h3>
			<p class="mt-1 text-sm text-base-content/60">
				Isi template Excel lalu unggah. Baris dengan product_code yang sudah ada akan diperbarui.
			</p>

			<div class="mt-4 flex flex-col items-start gap-3">
				<button
					class="btn btn-outline btn-sm"
					title="Unduh template Excel"
					aria-label="Unduh template Excel"
					on:click={downloadImportTemplate}
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M12 4v12m0 0l-4-4m4 4l4-4"
						/>
					</svg>
				</button>

				<input
					type="file"
					accept=".xlsx,.xls"
					class="file-input file-input-bordered file-input-primary w-full max-w-sm"
					on:change={onImportFile}
				/>
			</div>

			{#if importError}
				<div class="alert alert-error mt-3 text-sm">
					<span>{importError}</span>
				</div>
			{/if}

			{#if importResult}
				<div class="mt-3 rounded-xl border border-base-300 bg-base-100 p-3 text-sm">
					<div class="grid grid-cols-3 gap-2 text-center">
						<div>
							<div class="metric text-lg font-semibold text-success">{importResult.created}</div>
							<div class="text-xs text-base-content/60">Ditambah</div>
						</div>
						<div>
							<div class="metric text-lg font-semibold text-info">{importResult.updated}</div>
							<div class="text-xs text-base-content/60">Diperbarui</div>
						</div>
						<div>
							<div class="metric text-lg font-semibold text-error">{importResult.failed}</div>
							<div class="text-xs text-base-content/60">Gagal</div>
						</div>
					</div>
					{#if importResult.errors?.length > 0}
						<ul class="mt-3 max-h-32 space-y-1 overflow-y-auto border-t border-base-300 pt-2 text-xs">
							{#each importResult.errors as err}
								<li>
									<span class="font-mono text-base-content/50">Baris {err.row}:</span>
									<span class="text-error">{err.message}</span>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			{/if}

			<div class="modal-action">
				<button
					class="btn btn-sm btn-square"
					title="Tutup"
					aria-label="Tutup"
					on:click={() => (showImportModal = false)}
				>
					<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M6 6l12 12M18 6L6 18" stroke-linecap="round" />
					</svg>
				</button>
				<button
					class="btn btn-primary btn-sm btn-square"
					title="Import"
					aria-label="Mulai import"
					disabled={!importFile || isImporting}
					on:click={handleImport}
				>
					{#if isImporting}
						<Loader size={20} />
					{:else}
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round" />
						</svg>
					{/if}
				</button>
			</div>
		</div>
	</div>
	<button class="modal-backdrop" aria-label="Tutup import" on:click={() => (showImportModal = false)}></button>
{/if}
