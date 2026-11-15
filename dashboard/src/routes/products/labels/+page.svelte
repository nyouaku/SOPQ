<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import ThermalPrintPanel from '$lib/components/ThermalPrintPanel.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import QRCode from 'qrcode';
	import { authStore } from '$lib/stores/auth';

	let products: any[] = [];
	let selected = new Set<string>();
	let isLoading = true;
	let isGenerating = false;

	// Derived array dari produk yang dipilih, untuk dikirim ke ThermalPrintPanel
	$: selectedProducts = products.filter((p) => selected.has(p.id));

	// Ukuran label (dalam mm)
	let labelWidth = 50;
	let labelHeight = 30;
	let columns = 2;

	onMount(async () => {
		if (!$authStore.isAuthenticated) {
			goto('/login');
			return;
		}
		try {
			const res = await apiClient.get<{ products: any[] }>(`${API_ENDPOINTS.products}?limit=500`);
			if (res.success && res.data) {
				products = res.data.products || [];
			}
		} catch (e: any) {
			console.error('Failed to load products:', e);
		} finally {
			isLoading = false;
		}
	});

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

	function mmToPx(mm: number) {
		return Math.round(mm * 3.7795); // 1mm = 3.7795px at 96dpi
	}

	async function generatePDF() {
		if (selected.size === 0) return;
		isGenerating = true;
		try {
			const selectedProducts = products.filter((p) => selected.has(p.id));

			// Load jsPDF dynamically
			const { jsPDF } = await import('jspdf');
			const pdf = new jsPDF({
				unit: 'mm',
				format: 'a4',
				orientation: 'portrait'
			});

			const pageWidth = 210;
			const pageHeight = 297;
			const margin = 5;
			const colWidth = (pageWidth - margin * 2) / columns;
			const rowHeight = labelHeight + 2;
			const maxRowsPerPage = Math.floor((pageHeight - margin * 2) / rowHeight);

			let x = margin;
			let y = margin;
			let col = 0;
			let row = 0;

			for (const product of selectedProducts) {
				if (row >= maxRowsPerPage) {
					pdf.addPage();
					x = margin;
					y = margin;
					col = 0;
					row = 0;
				}

				// Generate QR code as data URL
				const qrDataUrl = await QRCode.toDataURL(product.product_code, {
					width: mmToPx(labelHeight * 0.7),
					margin: 0
				});

				// Draw label border
				pdf.setDrawColor(200);
				pdf.setLineWidth(0.2);
				pdf.rect(x, y, colWidth, labelHeight);

				// Draw QR code (left side)
				const qrSize = labelHeight - 6;
				pdf.addImage(qrDataUrl, 'PNG', x + 2, y + 3, qrSize, qrSize);

				// Draw text (right of QR)
				const textX = x + qrSize + 4;
				const textWidth = colWidth - qrSize - 6;

				pdf.setFont('helvetica', 'bold');
				pdf.setFontSize(9);
				pdf.text(product.name.substring(0, 25), textX, y + 8, { maxWidth: textWidth });

				pdf.setFont('helvetica', 'normal');
				pdf.setFontSize(7);
				pdf.text(`Kode: ${product.product_code}`, textX, y + 14, { maxWidth: textWidth });
				if (product.brand) {
					pdf.text(`Brand: ${product.brand}`, textX, y + 18, { maxWidth: textWidth });
				}

				// Move to next position
				col++;
				if (col >= columns) {
					col = 0;
					row++;
					y += rowHeight;
				}
				x = margin + col * colWidth;
			}

			// Download PDF
			pdf.save(`labels-${selected.size}-produk.pdf`);
		} catch (e: any) {
			console.error('PDF generation error:', e);
			alert('Gagal generate PDF: ' + (e.message || 'unknown error'));
		} finally {
			isGenerating = false;
		}
	}
</script>

<svelte:head>
	<title>Cetak Label - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6">
			<h1 class="page-title">Cetak Label Barcode</h1>
			<p class="mt-1 text-sm text-base-content/60">
				Pilih produk untuk generate label QR code dalam format PDF
			</p>
		</div>

		<!-- Konfigurasi -->
		<div class="mb-6 rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
			<h2 class="mb-3 font-semibold">Konfigurasi Label</h2>
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
				<div>
					<label class="label py-0" for="label-width">
						<span class="label-text text-xs">Lebar Label (mm)</span>
					</label>
					<input
						id="label-width"
						type="number"
						class="input input-bordered input-sm"
						bind:value={labelWidth}
						min="30"
						max="100"
					/>
				</div>
				<div>
					<label class="label py-0" for="label-height">
						<span class="label-text text-xs">Tinggi Label (mm)</span>
					</label>
					<input
						id="label-height"
						type="number"
						class="input input-bordered input-sm"
						bind:value={labelHeight}
						min="20"
						max="60"
					/>
				</div>
				<div>
					<label class="label py-0" for="columns">
						<span class="label-text text-xs">Kolom per Baris</span>
					</label>
					<select id="columns" class="select select-bordered select-sm" bind:value={columns}>
						<option value={1}>1</option>
						<option value={2}>2</option>
						<option value={3}>3</option>
						<option value={4}>4</option>
					</select>
				</div>
			</div>
		</div>

		<!-- Produk List -->
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
			{:else if products.length === 0}
				<p class="text-sm text-base-content/60">Belum ada produk</p>
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
								</div>
							</label>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Thermal Print Option -->
			{#if selected.size > 0}
				<div class="mt-4 border-t border-base-300 pt-4">
					<h3 class="mb-2 text-sm font-semibold">Atau Cetak Langsung ke Printer Thermal</h3>
					<ThermalPrintPanel products={selectedProducts} />
				</div>
			{/if}

			<button
				class="btn btn-primary mt-4 w-full"
				on:click={generatePDF}
				disabled={selected.size === 0 || isGenerating}
			>
				{#if isGenerating}
					<span class="loading loading-spinner loading-sm"></span>
					Membuat PDF...
				{:else}
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
					Generate PDF ({selected.size} label)
				{/if}
			</button>
		</div>
	</div>
</LayoutWrapper>
