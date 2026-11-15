<script lang="ts">
	import { writable } from 'svelte/store';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import ListSkeleton from '$lib/components/skeletons/ListSkeleton.svelte';
	import { activityLogsQuery } from '$lib/api/queries';

	let action = '';
	let entityType = '';
	let page = 1;
	const limit = 25;

	// Params berupa store: setiap perubahan filter membuat key query baru.
	const params = writable(new URLSearchParams(`page=1&limit=${limit}`));
	const query = activityLogsQuery(params);

	function applyFilters() {
		const p = new URLSearchParams();
		if (action) p.set('action', action);
		if (entityType) p.set('entity_type', entityType);
		p.set('page', String(page));
		p.set('limit', String(limit));
		params.set(p);
	}

	$: logs = ($query.data?.logs as any[]) || [];
	$: total = $query.data?.total || 0;
	$: totalPages = Math.max(1, Math.ceil(total / limit));

	const ACTION_LABELS: Record<string, string> = {
		login: 'Masuk',
		logout: 'Keluar',
		start_stock_opname: 'Mulai opname',
		submit_scan: 'Scan',
		complete_stock_opname: 'Selesaikan opname',
		approve_stock_opname: 'Setujui opname',
		create_product: 'Tambah produk',
		update_product: 'Ubah produk',
		delete_product: 'Hapus produk'
	};

	function labelFor(a: string) {
		return ACTION_LABELS[a] || a || '-';
	}

	function fmt(iso: string) {
		if (!iso) return '-';
		const d = new Date(iso);
		return d.toLocaleString('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function goPage(next: number) {
		page = next;
		applyFilters();
	}

	function clearFilters() {
		action = '';
		entityType = '';
		page = 1;
		applyFilters();
	}
</script>

<svelte:head>
	<title>Activity Log - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="page-header">
			<div class="min-w-0">
				<h1 class="page-title">Activity Log</h1>
				<p class="page-subtitle">Riwayat tindakan pengguna</p>
			</div>
			<button class="btn btn-ghost btn-sm btn-square" title="Muat ulang" aria-label="Muat ulang" on:click={applyFilters}>
				<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M20 12a8 8 0 1 1-2.34-5.66" stroke-linecap="round" />
					<path d="M20 4v4h-4" stroke-linecap="round" stroke-linejoin="round" />
				</svg>
			</button>
		</div>

		<!-- Filter -->
		<section class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-4">
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
				<div>
					<label class="label py-0" for="log-action">
						<span class="label-text text-xs text-base-content/60">Aksi</span>
					</label>
					<input
						id="log-action"
						class="input input-bordered input-sm mt-1 w-full"
						placeholder="mis. submit_scan"
						bind:value={action}
						on:keypress={(e) => e.key === 'Enter' && goPage(1)}
					/>
				</div>
				<div>
					<label class="label py-0" for="log-entity">
						<span class="label-text text-xs text-base-content/60">Entitas</span>
					</label>
					<input
						id="log-entity"
						class="input input-bordered input-sm mt-1 w-full"
						placeholder="mis. stock_opname"
						bind:value={entityType}
						on:keypress={(e) => e.key === 'Enter' && goPage(1)}
					/>
				</div>
				<div class="flex items-end gap-2">
					<button class="btn btn-primary btn-sm btn-square" title="Terapkan filter" aria-label="Terapkan filter" on:click={() => goPage(1)}>
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="11" cy="11" r="7" />
							<path d="M16.5 16.5L21 21" stroke-linecap="round" />
						</svg>
					</button>
					<button class="btn btn-ghost btn-sm" on:click={clearFilters}>Reset</button>
				</div>
			</div>
		</section>

		<!-- Daftar -->
		<section class="overflow-hidden rounded-2xl border border-base-300 bg-base-200">
			{#if $query.isLoading}
				<ListSkeleton count={6} />
			{:else if $query.isError}
				<div class="p-6 text-center">
					<p class="text-sm text-error">{$query.error?.message || 'Gagal memuat log'}</p>
					<button class="btn btn-ghost btn-sm mt-2" on:click={() => $query.refetch()}>Coba lagi</button>
				</div>
			{:else if logs.length === 0}
				<div class="px-4 py-16 text-center">
					<p class="text-sm text-base-content/60">Tidak ada aktivitas</p>
					<p class="mt-1 text-xs text-base-content/40">Coba ubah atau reset filter</p>
				</div>
			{:else}
				<ul class="divide-y divide-base-300">
					{#each logs as log}
						<li class="flex flex-col gap-1 px-4 py-3 sm:flex-row sm:items-center sm:gap-3">
							<span class="badge badge-ghost badge-sm shrink-0 font-mono text-[11px]">
								{labelFor(log.action)}
							</span>
							<span class="min-w-0 flex-1 text-sm">
								{#if log.entity_type}
									<span class="text-base-content/60">{log.entity_type}</span>
								{/if}
								{#if log.entity_id}
									<span class="metric ml-1 text-xs text-base-content/40">
										{String(log.entity_id).slice(0, 8)}
									</span>
								{/if}
							</span>
							<span class="shrink-0 text-xs text-base-content/50">
								{log.user?.full_name || log.user?.username || '-'}
							</span>
							<span class="metric shrink-0 text-xs text-base-content/40">{fmt(log.created_at)}</span>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="mt-4 flex items-center justify-center gap-2">
				<button class="btn btn-sm" disabled={page <= 1} on:click={() => goPage(page - 1)}>Sebelumnya</button>
				<span class="metric text-xs text-base-content/60">{page} / {totalPages}</span>
				<button class="btn btn-sm" disabled={page >= totalPages} on:click={() => goPage(page + 1)}>Berikutnya</button>
			</div>
		{/if}
	</div>
</LayoutWrapper>
