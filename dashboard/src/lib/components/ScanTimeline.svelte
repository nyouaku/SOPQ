<script lang="ts">
	export let details: any[] = [];
	export let view: 'table' | 'timeline' = 'table';

	$: sortedDetails = [...details].sort(
		(a, b) => new Date(b.scanned_at).getTime() - new Date(a.scanned_at).getTime()
	);

	function getDotColor(difference: number): string {
		if (difference === 0) return 'bg-success';
		if (Math.abs(difference) <= 5) return 'bg-warning';
		return 'bg-error';
	}

	function formatTime(iso: string): string {
		return new Date(iso).toLocaleTimeString('id-ID', {
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric'
		});
	}
</script>

{#if view === 'timeline'}
	<!-- Timeline View -->
	<div class="max-h-96 overflow-y-auto rounded-xl bg-base-100 p-4">
		{#if sortedDetails.length === 0}
			<p class="py-8 text-center text-sm text-base-content/50">Belum ada scan</p>
		{:else}
			<ul class="relative space-y-4 before:absolute before:left-4 before:top-0 before:h-full before:w-0.5 before:bg-base-300">
				{#each sortedDetails as detail}
					<li class="relative flex gap-4 pl-10">
						<!-- Timeline dot -->
						<span
							class="absolute left-0 top-1 h-8 w-8 rounded-full border-2 border-base-100 {getDotColor(detail.difference)} flex items-center justify-center"
						>
							<svg class="h-4 w-4 text-base-100" fill="currentColor" viewBox="0 0 20 20">
								<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
							</svg>
						</span>

						<!-- Content -->
						<div class="flex-1 rounded-lg border border-base-300 bg-base-200 p-3">
							<div class="flex items-start justify-between gap-2">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-semibold">
										{detail.product?.name || 'Produk'}
									</p>
									<p class="metric text-xs text-base-content/60">
										{detail.product?.product_code || '-'}
									</p>
								</div>
								<span
									class="badge badge-sm {detail.difference === 0 ? 'badge-success' : detail.difference > 0 ? 'badge-warning' : 'badge-error'}"
								>
									{detail.difference > 0 ? '+' : ''}{detail.difference}
								</span>
							</div>
							<div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-base-content/60">
								<span>🕐 {formatTime(detail.scanned_at)}</span>
								<span>📦 Stok Awal: {detail.stock_system}</span>
								<span>📊 Fisik: {detail.stock_physical}</span>
								<span>👤 {detail.scanned_user?.full_name || '-'}</span>
							</div>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
{:else}
	<!-- Table View (fallback/default) -->
	<slot />
{/if}

<style>
	.before\:absolute::before {
		content: '';
	}
</style>
