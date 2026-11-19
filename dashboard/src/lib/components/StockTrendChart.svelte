<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Chart, registerables, type ChartConfiguration } from 'chart.js';

	Chart.register(...registerables);

	export let data: Array<{
		scanned_at: string;
		stock_system: number;
		stock_physical: number;
		difference: number;
		session_code: string;
	}> = [];
	export let height = 320;

	let canvas: HTMLCanvasElement | undefined;
	let chart: Chart | null = null;

	$: if (canvas && data.length > 0) {
		renderChart();
	}

	function renderChart() {
		if (!canvas) return;

		// Cleanup previous chart
		if (chart) {
			chart.destroy();
			chart = null;
		}

		const labels = data.map((d) => new Date(d.scanned_at).toLocaleDateString('id-ID', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		}));

		const config: ChartConfiguration = {
			type: 'line',
			data: {
				labels,
				datasets: [
					{
						label: 'Stok Fisik',
						data: data.map((d) => d.stock_physical),
						borderColor: '#FF6B35',
						backgroundColor: 'rgba(255, 107, 53, 0.1)',
						tension: 0.3,
						pointRadius: 4,
						pointHoverRadius: 6
					},
					{
						label: 'Stok Awal',
						data: data.map((d) => d.stock_system),
						borderColor: '#5EEAD4',
						backgroundColor: 'rgba(94, 234, 212, 0.1)',
						tension: 0.3,
						pointRadius: 3,
						pointHoverRadius: 5,
						borderDash: [5, 5]
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				interaction: {
					mode: 'index',
					intersect: false
				},
				plugins: {
					tooltip: {
						callbacks: {
							title: (items) => {
								const idx = items[0].dataIndex;
								return `${data[idx].session_code || '-'} — ${labels[idx]}`;
							},
							label: (ctx) => {
								const idx = ctx.dataIndex;
								const diff = data[idx].difference;
								const sign = diff > 0 ? '+' : '';
								return `${ctx.dataset.label}: ${ctx.parsed.y} (selisih ${sign}${diff})`;
							}
						}
					},
					legend: { position: 'top' }
				},
				scales: {
					x: { grid: { display: false } },
					y: { grid: { color: 'rgba(0,0,0,0.05)' } }
				}
			}
		};

		chart = new Chart(canvas, config);
	}

	onDestroy(() => {
		if (chart) chart.destroy();
	});
</script>

<div class="w-full" style="height: {height}px">
	{#if data.length === 0}
		<div class="flex h-full flex-col items-center justify-center gap-2 text-center">
			<svg class="h-12 w-12 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
				<path stroke-linecap="round" stroke-linejoin="round" d="M3 13h2v5H3zM8 8h2v10H8zM13 4h2v14h-2zM18 10h2v8h-2z" />
			</svg>
			<p class="text-sm text-base-content/60">Belum ada data opname untuk produk ini</p>
		</div>
	{:else}
		<canvas bind:this={canvas} aria-label="Grafik riwayat stok"></canvas>
	{/if}
</div>
