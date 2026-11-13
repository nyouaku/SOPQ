<script lang="ts">
	import { onMount } from 'svelte';
	import QRCode from 'qrcode';

	/** Nilai yang di-encode ke QR — KODE produk. */
	export let value: string = '';
	export let productCode: string = '';
	export let productName: string = '';
	/** Ukuran canvas QR di layar (px). */
	export let size: number = 180;

	let canvas: HTMLCanvasElement;
	let error = '';

	// qrcode hanya jalan di browser.
	onMount(() => {
		if (!canvas || !value) return;
		QRCode.toCanvas(canvas, value, { width: size, margin: 1, errorCorrectionLevel: 'M' }, (err) => {
			error = err ? 'Gagal membuat QR' : '';
		});
	});

	/**
	 * Label stiker siap cetak: QR 2x + kode produk + nama.
	 * Digambar ke canvas 600x760 (skala ~2x untuk cetak tajam), lalu diunduh PNG.
	 */
	// Di-export agar parent bisa memanggilnya lewat bind:this bila perlu.
	export async function downloadLabel() {
		if (!value) return;

		const W = 600;
		const H = 760;
		const qrPx = 440;

		const qrDataUrl = await QRCode.toDataURL(value, {
			width: qrPx,
			margin: 1,
			errorCorrectionLevel: 'M'
		});

		const img = new Image();
		img.src = qrDataUrl;
		await img.decode();

		const c = document.createElement('canvas');
		c.width = W;
		c.height = H;
		const ctx = c.getContext('2d');
		if (!ctx) return;

		// Kertas putih + garis potong tipis (label tercetak jelas di printer apa pun).
		ctx.fillStyle = '#ffffff';
		ctx.fillRect(0, 0, W, H);
		ctx.strokeStyle = '#d4d4d8';
		ctx.lineWidth = 2;
		ctx.strokeRect(12, 12, W - 24, H - 24);

		ctx.drawImage(img, (W - qrPx) / 2, 48, qrPx, qrPx);

		ctx.fillStyle = '#111111';
		ctx.textAlign = 'center';
		ctx.font = 'bold 34px "JetBrains Mono", Consolas, monospace';
		ctx.fillText(value, W / 2, 560);

		if (productCode && productCode !== value) {
			ctx.font = '24px "Inter", system-ui, sans-serif';
			ctx.fillStyle = '#52525b';
			ctx.fillText(productCode, W / 2, 604);
		}

		// Nama produk dipotong agar tetap satu baris rapi.
		if (productName) {
			ctx.font = '600 26px "Inter", system-ui, sans-serif';
			ctx.fillStyle = '#111111';
			ctx.fillText(truncate(ctx, productName, W - 96), W / 2, 660);
		}

		ctx.font = '20px "Inter", system-ui, sans-serif';
		ctx.fillStyle = '#71717a';
		ctx.fillText('STOCK OPNAME', W / 2, 706);

		const a = document.createElement('a');
		a.href = c.toDataURL('image/png');
		a.download = `label-${productCode || value}.png`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
	}

	function truncate(ctx: CanvasRenderingContext2D, text: string, maxWidth: number): string {
		if (ctx.measureText(text).width <= maxWidth) return text;
		let out = text;
		while (out.length > 1 && ctx.measureText(out + '…').width > maxWidth) {
			out = out.slice(0, -1);
		}
		return out + '…';
	}
</script>

<div class="flex flex-col items-center gap-3">
	<canvas
		bind:this={canvas}
		class="rounded-xl bg-white p-2"
		style="width:{size}px;height:{size}px"
	></canvas>
	{#if error}
		<p class="text-xs text-error">{error}</p>
	{/if}
	<p class="metric text-xs text-base-content/60">{value}</p>
</div>
