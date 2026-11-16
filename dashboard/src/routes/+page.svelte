<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { Chart, registerables } from 'chart.js';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';

	Chart.register(...registerables);

	let isLoading = true;
	let summary: any = null;
	let chart: Chart<'doughnut', number[], string> | null = null;
	let themeObserver: MutationObserver | null = null;

	// Angka & persen diturunkan dari data API (bukan dummy).
	$: checked = summary?.total_checked || 0;
	$: unchecked = summary?.total_unchecked || 0;
	$: total = checked + unchecked;
	$: percent = total > 0 ? Math.round((checked / total) * 100) : 0;

	$: tiles = [
		{ label: 'Total SKU', value: summary?.total_sku || 0, desc: 'Produk di sistem', accent: 'text-base-content' },
		{ label: 'Progress hari ini', value: summary?.today_progress || 0, desc: 'Item di-scan', accent: 'text-primary' },
		{ label: 'Selisih', value: summary?.total_difference || 0, desc: 'Item dengan selisih', accent: 'text-warning' },
		{ label: 'Belum dicek', value: summary?.total_unchecked || 0, desc: 'Item tersisa', accent: 'text-error' }
	];

	// Warna chart dibaca dari token DaisyUI → ikut berubah saat tema diganti.
	function token(name: string, alpha = 1): string {
		if (typeof document === 'undefined') return 'transparent';
		const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
		return `hsl(${v} / ${alpha})`;
	}

	function chartColors() {
		return { done: token('--su'), rest: token('--bc', 0.12), text: token('--bc', 0.7), grid: token('--bc', 0.1) };
	}

	onMount(async () => {
		await loadDashboardData();
		await tick();
		initChart();

		themeObserver = new MutationObserver(() => applyChartTheme());
		themeObserver.observe(document.documentElement, {
			attributes: true,
			attributeFilter: ['data-theme']
		});
	});

	onDestroy(() => {
		chart?.destroy();
		themeObserver?.disconnect();
	});

	async function loadDashboardData() {
		try {
			const response = await apiClient.get(API_ENDPOINTS.dashboardSummary);
			if (response.success && response.data) {
				summary = response.data;
			}
		} catch (error) {
			console.error('Failed to load dashboard data:', error);
		} finally {
			isLoading = false;
		}
	}

	function initChart() {
		const ctx = document.getElementById('progressChart') as HTMLCanvasElement | null;
		if (!ctx) return;

		const c = chartColors();
		chart = new Chart(ctx, {
			type: 'doughnut',
			data: {
				labels: ['Selesai', 'Belum dicek'],
				datasets: [
					{
						data: [0, 100],
						backgroundColor: [c.done, c.rest],
						borderWidth: 0,
						hoverOffset: 3
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				cutout: '74%',
				plugins: {
					legend: { display: false },
					tooltip: {
						backgroundColor: c.text,
						displayColors: false
					}
				}
			}
		});

		applyChartTheme();
	}

	function applyChartTheme() {
		if (!chart) return;
		const c = chartColors();
		chart.data.datasets[0].backgroundColor = [c.done, c.rest];
		if (chart.options.plugins?.tooltip) {
			chart.options.plugins.tooltip.backgroundColor = c.text;
		}
		const data = total > 0 ? [checked, unchecked] : [0, 100];
		chart.data.datasets[0].data = data;
		chart.update();
	}

	// Data datang setelah chart dibuat → sinkronkan.
	$: if (chart) applyChartTheme();
</script>

<svelte:head>
	<title>Dashboard - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<!-- Page header -->
		<div class="page-header">
			<div>
				<h1 class="page-title">Dashboard</h1>
				<p class="page-subtitle">
					Ringkasan stock opname hari ini
				</p>
			</div>
			<a href="/scan" class="btn btn-primary btn-sm md:btn-md">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-4 w-4"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="2"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M3 8V5a2 2 0 0 1 2-2h3M16 3h3a2 2 0 0 1 2 2v3M21 16v3a2 2 0 0 1-2 2h-3M8 21H5a2 2 0 0 1-2-2v-3" />
				</svg>
				Scan
			</a>
		</div>

		{#if isLoading}
			<div class="flex h-64 items-center justify-center">
				<Loader size={48} />
			</div>
		{:else}
			<!-- Metric grid: satu permukaan, garis rambut pemisah (bukan 4 kartu mengambang) -->
			<div class="mb-6 overflow-hidden rounded-2xl border border-base-300">
				<div class="grid grid-cols-2 gap-px bg-base-300 lg:grid-cols-4">
					{#each tiles as tile}
						<div class="bg-base-200 p-4 md:p-5">
							<div class="text-xs font-medium text-base-content/60">{tile.label}</div>
							<div class="metric mt-2 text-3xl font-semibold leading-none md:text-4xl">
								{tile.value}
							</div>
							<div class="mt-2 text-xs {tile.accent}">{tile.desc}</div>
						</div>
					{/each}
				</div>
			</div>

			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				<!-- Progress chart -->
				<section class="rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
					<div class="mb-4 flex items-baseline justify-between gap-3">
						<h2 class="font-semibold">Progress overview</h2>
						<span class="metric text-sm text-base-content/60">{checked}/{total}</span>
					</div>

					<div class="relative h-56">
						<canvas id="progressChart"></canvas>
						<!-- Label tengah: angka utama yang dibaca sekali lihat -->
						<div class="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
							<span class="metric text-4xl font-semibold leading-none">{percent}%</span>
							<span class="mt-1 text-xs text-base-content/60">selesai</span>
						</div>
					</div>

					<div class="mt-4 flex items-center justify-center gap-5 text-xs">
						<span class="flex items-center gap-2">
							<span class="h-2.5 w-2.5 rounded-full bg-success"></span>
							Selesai
							<span class="metric text-base-content/60">{checked}</span>
						</span>
						<span class="flex items-center gap-2">
							<span class="h-2.5 w-2.5 rounded-full bg-base-content/20"></span>
							Belum dicek
							<span class="metric text-base-content/60">{unchecked}</span>
						</span>
					</div>
				</section>

				<!-- Recent activity -->
				<section class="rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
					<div class="mb-4">
						<h2 class="font-semibold">Aktivitas terbaru</h2>
						<p class="text-xs text-base-content/60">Pergerakan scan terakhir</p>
					</div>

					{#if summary?.recent_activities?.length > 0}
						<ul class="max-h-72 divide-y divide-base-300 overflow-y-auto pr-1">
							{#each summary.recent_activities as activity}
								<li class="flex items-start gap-3 py-3">
									<div
										class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-base-300 text-xs font-semibold uppercase"
									>
										{(activity.user?.full_name || '?').charAt(0)}
									</div>
									<div class="min-w-0 flex-1">
										<div class="truncate text-sm">{activity.action}</div>
										<div class="truncate text-xs text-base-content/60">{activity.user?.full_name || activity.user?.username}</div>
									</div>
									<time
										class="metric shrink-0 text-xs text-base-content/50"
										datetime={activity.created_at}
									>
										{new Date(activity.created_at).toLocaleTimeString('id-ID')}
									</time>
								</li>
							{/each}
						</ul>
					{:else}
						<div class="flex h-40 flex-col items-center justify-center gap-1 text-center">
							<span class="text-sm text-base-content/60">Belum ada aktivitas</span>
							<a href="/scan" class="link link-primary text-xs">Mulai scan pertama</a>
						</div>
					{/if}
				</section>
			</div>
		{/if}
	</div>
</LayoutWrapper>
