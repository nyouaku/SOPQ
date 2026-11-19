<script lang="ts">
	import { writable } from 'svelte/store';
	import { page as pageStore } from '$app/stores';
	import { replaceState } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import TableSkeleton from '$lib/components/skeletons/TableSkeleton.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { authStore } from '$lib/stores/auth';
	import { useQueryClient } from '@tanstack/svelte-query';
	import { inventoryQuery, warehousesQuery } from '$lib/api/queries';

	const queryClient = useQueryClient();
	const limit = 25;
	const init = new URLSearchParams($pageStore.url.search);

	let warehouseId = init.get('warehouse_id') ?? '';
	let movementType = init.get('movement_type') ?? '';
	let from = init.get('from') ?? '';
	let to = init.get('to') ?? '';
	let page = Math.max(1, parseInt(init.get('page') ?? '1', 10) || 1);
	let applied = writable(buildParams(page));

	onMount(() =>
		applied.subscribe((p) => replaceState(`${window.location.pathname}?${p.toString()}`, {}))
	);

	const movements = inventoryQuery(applied);
	const warehousesQ = warehousesQuery();

	$: rows = ($movements.data?.data as any[]) || [];
	$: total = $movements.data?.total || 0;
	$: totalPages = Math.max(1, Math.ceil(total / limit));
	$: warehouseList = ($warehousesQ.data as any[]) ?? [];

	function buildParams(nextPage: number) {
		const p = new URLSearchParams();
		if (warehouseId) p.set('warehouse_id', warehouseId);
		if (movementType) p.set('movement_type', movementType);
		if (from) p.set('from', from);
		if (to) p.set('to', to);
		p.set('page', String(nextPage));
		p.set('limit', String(limit));
		return p;
	}

	function apply() {
		page = 1;
		applied.set(buildParams(1));
	}

	function goPage(n: number) {
		page = n;
		const p = new URLSearchParams($applied);
		p.set('page', String(n));
		applied.set(p);
	}

	function clearFilters() {
		warehouseId = '';
		movementType = '';
		from = '';
		to = '';
		apply();
	}

	const typeBadge: Record<string, string> = {
		in: 'badge-success',
		out: 'badge-error',
		transfer: 'badge-info'
	};

	function fmtDate(iso: string) {
		if (!iso) return '-';
		return new Date(iso).toLocaleString('id-ID', {
			day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit'
		});
	}

	// Ringkasan (agregat client-side dari halaman aktif) — tanpa ubah backend/API.
	$: stats = (() => {
		let inCount = 0, outCount = 0, transferCount = 0;
		for (const m of rows) {
			if (m.movement_type === 'in') inCount++;
			else if (m.movement_type === 'out') outCount++;
			else if (m.movement_type === 'transfer') transferCount++;
		}
		return { inCount, outCount, transferCount };
	})();

	// Pembatalan dokumen: balik stok + tandai 'cancelled' (bukan hapus). Alasan wajib.
	let cancelTarget: any = null;
	let cancelReason = '';
	let isCancelling = false;

	function openCancel(m: any) {
		cancelTarget = m;
		cancelReason = '';
	}

	function closeCancel() {
		if (isCancelling) return;
		cancelTarget = null;
		cancelReason = '';
	}

	async function confirmCancel() {
		if (!cancelTarget) return;
		if (cancelReason.trim().length < 3) {
			toastStore.error('Alasan minimal 3 karakter');
			return;
		}
		isCancelling = true;
		try {
			const res = await apiClient.post(API_ENDPOINTS.inventoryCancel(cancelTarget.id), {
				reason: cancelReason.trim()
			});
			if (!res.success) throw new Error(res.error?.message || 'Gagal membatalkan');
			toastStore.success(`Dokumen ${cancelTarget.code} dibatalkan, stok dikembalikan`);
			cancelTarget = null;
			cancelReason = '';
			await queryClient.invalidateQueries({ queryKey: ['inventory'] });
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal membatalkan dokumen');
		} finally {
			isCancelling = false;
		}
	}

	import { onMount } from 'svelte';
</script>

<svelte:head>
	<title>Inventory - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
			<div>
				<h1 class="page-title">Inventory</h1>
				<p class="page-subtitle">Pergerakan stok: barang masuk, keluar, dan transfer antar gudang</p>
			</div>
			<a href="/inventory/new" class="btn btn-primary btn-sm shrink-0 md:btn-md">
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" d="M12 4v16m8-8H4" />
				</svg>
				<span class="hidden sm:inline">Catat pergerakan</span>
				<span class="sm:hidden">Baru</span>
			</a>
		</div>

		<!-- Stat ringkasan (glass, dari halaman aktif) -->
		<div class="mb-4 grid grid-cols-3 gap-3">
			<div class="card-apple p-4">
				<div class="text-xs text-base-content/60">Masuk (IN)</div>
				<div class="metric text-2xl font-semibold text-success">{stats.inCount}</div>
			</div>
			<div class="card-apple p-4">
				<div class="text-xs text-base-content/60">Keluar (OUT)</div>
				<div class="metric text-2xl font-semibold text-error">{stats.outCount}</div>
			</div>
			<div class="card-apple p-4">
				<div class="text-xs text-base-content/60">Transfer</div>
				<div class="metric text-2xl font-semibold text-info">{stats.transferCount}</div>
			</div>
		</div>

		<!-- Filters -->
		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-4">
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
				<div>
					<label class="label py-0" for="f-wh"><span class="label-text text-xs text-base-content/60">Gudang</span></label>
					<select id="f-wh" class="select select-bordered select-sm w-full" bind:value={warehouseId}>
						<option value="">Semua gudang</option>
						{#each warehouseList as wh}
							<option value={wh.id}>{wh.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="label py-0" for="f-type"><span class="label-text text-xs text-base-content/60">Jenis</span></label>
					<select id="f-type" class="select select-bordered select-sm w-full" bind:value={movementType}>
						<option value="">Semua jenis</option>
						<option value="in">Masuk (IN)</option>
						<option value="out">Keluar (OUT)</option>
						<option value="transfer">Transfer</option>
					</select>
				</div>
				<div>
					<label class="label py-0" for="f-from"><span class="label-text text-xs text-base-content/60">Dari</span></label>
					<input id="f-from" type="date" class="input input-bordered input-sm w-full" bind:value={from} />
				</div>
				<div>
					<label class="label py-0" for="f-to"><span class="label-text text-xs text-base-content/60">Sampai</span></label>
					<input id="f-to" type="date" class="input input-bordered input-sm w-full" bind:value={to} />
				</div>
			</div>
			<div class="mt-3 flex gap-2">
				<button class="btn btn-primary btn-sm" on:click={apply}>Terapkan</button>
				<button class="btn btn-ghost btn-sm" on:click={clearFilters}>Bersihkan</button>
			</div>
		</div>

		<!-- Table -->
		<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
			{#if $movements.isLoading}
				<TableSkeleton rows={6} columns={6} />
			{:else if rows.length === 0}
				<p class="py-8 text-center text-sm text-base-content/60">Belum ada pergerakan stok</p>
			{:else}
				<div class="overflow-x-auto">
					<table class="table table-sm">
						<thead>
							<tr>
								<th>Kode</th>
								<th>Jenis</th>
								<th>Gudang</th>
								<th>Tujuan</th>
								<th>Ref</th>
								<th>Waktu</th>
								<th>Status</th>
								<th></th>
							</tr>
						</thead>
						<tbody>
							{#each rows as m}
								<tr>
									<td><a class="link link-primary metric text-xs" href="/inventory/{m.id}">{m.code}</a></td>
									<td>
										<span class="badge badge-sm capitalize {typeBadge[m.movement_type] || 'badge-ghost'}">
											{m.movement_type}
										</span>
									</td>
									<td class="text-xs">{m.warehouse?.name || '-'}</td>
									<td class="text-xs">{m.to_warehouse?.name || '-'}</td>
									<td class="metric text-xs">{m.reference_code || '-'}</td>
									<td class="text-xs text-base-content/60">{fmtDate(m.created_at)}</td>
									<td>
										{#if m.status === 'cancelled'}
											<span class="badge badge-ghost badge-sm">dibatalkan</span>
										{:else}
											<span class="badge badge-outline badge-sm">posted</span>
										{/if}
									</td>
									<td class="text-right">
										{#if m.status !== 'cancelled' && $authStore.user?.role?.name !== 'staff_gudang'}
											<button class="btn btn-ghost btn-xs text-error" on:click={() => openCancel(m)}>
												Batalkan
											</button>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				{#if totalPages > 1}
					<div class="mt-4 flex items-center justify-center gap-2">
						<button class="btn btn-sm" disabled={page <= 1} on:click={() => goPage(page - 1)}>Sebelumnya</button>
						<span class="metric text-xs text-base-content/60">{page} / {totalPages}</span>
						<button class="btn btn-sm" disabled={page >= totalPages} on:click={() => goPage(page + 1)}>Berikutnya</button>
					</div>
				{/if}
			{/if}
		</div>

		<ConfirmDialog
			isOpen={!!cancelTarget}
			title="Batalkan dokumen?"
			message={`Dokumen ${cancelTarget?.code ?? ''} akan dibatalkan dan stok dikembalikan ke posisi sebelum dokumen.\nDokumen tetap tersimpan sebagai riwayat.`}
			confirmLabel="Ya, batalkan"
			danger
			isLoading={isCancelling}
			on:confirm={confirmCancel}
			on:cancel={closeCancel}
		>
			<label class="mt-4 block">
				<span class="label-text text-xs text-base-content/60">Alasan pembatalan (wajib)</span>
				<textarea
					class="textarea textarea-bordered mt-1 w-full"
					rows="2"
					placeholder="mis. salah input qty"
					bind:value={cancelReason}
				></textarea>
			</label>
		</ConfirmDialog>
	</div>
</LayoutWrapper>
