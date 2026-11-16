<script lang="ts">
	import QRCode from 'qrcode';
	import { toastStore } from '$lib/stores/toast';

	export let products: any[] = [];
	export let paperSize: '58mm' | '80mm' = '80mm';

	let isGenerating = false;

	$: paperWidth = paperSize === '58mm' ? 58 : 80;
	$: qrSize = paperSize === '58mm' ? 20 : 28;

	// CSS untuk label thermal (dipisahkan agar tidak di-parse oleh Svelte)
	const thermalCSS = `
		@page {
			size: PAPER_WIDTHmm auto;
			margin: 0;
		}
		body {
			margin: 0;
			padding: 2mm;
			font-family: 'Arial', sans-serif;
			background: white;
		}
		.label {
			display: flex;
			align-items: center;
			gap: 2mm;
			padding: 2mm;
			border-bottom: 1px dashed #ccc;
			page-break-inside: avoid;
		}
		.qr img {
			width: QR_SIZEmm;
			height: QR_SIZEmm;
		}
		.info {
			flex: 1;
			min-width: 0;
		}
		.name {
			font-size: NAME_SIZEpx;
			font-weight: bold;
			margin-bottom: 1mm;
			word-break: break-word;
		}
		.code {
			font-family: 'Courier New', monospace;
			font-size: CODE_SIZEpx;
			margin-bottom: 1mm;
		}
		.brand {
			font-size: BRAND_SIZEpx;
			color: #666;
		}
		@media print {
			body { -webkit-print-color-adjust: exact; }
		}
	`;

	function getCSS(): string {
		const nameSize = paperSize === '58mm' ? '10' : '12';
		const codeSize = paperSize === '58mm' ? '8' : '10';
		const brandSize = paperSize === '58mm' ? '7' : '9';
		return thermalCSS
			.replace(/PAPER_WIDTH/g, String(paperWidth))
			.replace(/QR_SIZE/g, String(qrSize))
			.replace(/NAME_SIZE/g, nameSize)
			.replace(/CODE_SIZE/g, codeSize)
			.replace(/BRAND_SIZE/g, brandSize);
	}

	async function printThermal() {
		if (products.length === 0) return;
		isGenerating = true;

		try {
			const labels = await Promise.all(
				products.map(async (product) => {
					const qrDataUrl = await QRCode.toDataURL(product.product_code, {
						width: 200,
						margin: 1
					});

					return `
						<div class="label">
							<div class="qr">
								<img src="${qrDataUrl}" alt="QR" />
							</div>
							<div class="info">
								<div class="name">${escapeHtml(product.name)}</div>
								<div class="code">${escapeHtml(product.product_code)}</div>
								${product.brand ? `<div class="brand">${escapeHtml(product.brand)}</div>` : ''}
							</div>
						</div>
					`;
				})
			);

			const html = `
				<!DOCTYPE html>
				<html>
				<head>
					<meta charset="utf-8" />
					<title>Label Thermal</title>
					${getCSS()}
				</head>
				<body>
					${labels.join('')}
				</body>
				</html>
			`;

			const printWindow = window.open('', '_blank', 'width=400,height=600');
			if (!printWindow) {
				toastStore.error('Popup diblokir. Izinkan popup untuk mencetak.');
				return;
			}

			printWindow.document.write(html);
			printWindow.document.close();

			printWindow.onload = () => {
				setTimeout(() => {
					printWindow.print();
					printWindow.close();
				}, 500);
			};
		} catch (e: any) {
			console.error('Print error:', e);
			toastStore.error('Gagal mencetak: ' + (e.message || 'unknown error'));
		} finally {
			isGenerating = false;
		}
	}

	function escapeHtml(text: string): string {
		const div = document.createElement('div');
		div.textContent = text;
		return div.innerHTML;
	}
</script>

<div class="space-y-3">
	<div class="flex flex-wrap items-center gap-3">
		<div>
			<label class="label py-0" for="paper-size">
				<span class="label-text text-xs">Ukuran Kertas</span>
			</label>
			<select id="paper-size" class="select select-bordered select-sm" bind:value={paperSize}>
				<option value="58mm">58mm (Mini Printer)</option>
				<option value="80mm">80mm (Standard Thermal)</option>
			</select>
		</div>
		<div class="flex-1" />
		<button
			class="btn btn-primary"
			on:click={printThermal}
			disabled={products.length === 0 || isGenerating}
		>
			{#if isGenerating}
				<span class="loading loading-spinner loading-sm"></span>
				Menyiapkan...
			{:else}
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
				</svg>
				Cetak Thermal ({products.length} label)
			{/if}
		</button>
	</div>
	<p class="text-xs text-base-content/50">
		Labels akan dibuka di window baru dan otomatis memanggil dialog print. Pilih printer thermal yang terpasang.
	</p>
</div>
