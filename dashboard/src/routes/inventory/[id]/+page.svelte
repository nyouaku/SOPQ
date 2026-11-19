<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { authStore } from '$lib/stores/auth';

	$: movementId = $page.params.id;

	let movement: any = null;
	let isLoading = true;

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

	async function load() {
		// id undefined → jangan fetch, kembali ke list (mencegah /inventory/undefined 400).
		if (!movementId || movementId === 'undefined') {
			goto('/inventory/movements');
			return;
		}
		try {
			const res = await apiClient.get<any>(`/inventory/${movementId}`);
			if (res.success && res.data) {
				movement = res.data;
			} else {
				toastStore.error(res.error?.message || 'Gagal memuat data');
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat data');
		} finally {
			isLoading = false;
		}
	}

	load();

	// Pembatalan dokumen: balik stok, dokumen tetap tersimpan sebagai riwayat.
	let showCancel = false;
	let cancelReason = '';
	let isCancelling = false;

	async function confirmCancel() {
		if (cancelReason.trim().length < 3) {
			toastStore.error('Alasan minimal 3 karakter');
			return;
		}
		isCancelling = true;
		try {
			const res = await apiClient.post(API_ENDPOINTS.inventoryCancel(movementId ?? ''), {
				reason: cancelReason.trim()
			});
			if (!res.success) throw new Error(res.error?.message || 'Gagal membatalkan');
			toastStore.success('Dokumen dibatalkan, stok dikembalikan');
			showCancel = false;
			cancelReason = '';
			movement = { ...movement, status: 'cancelled' };
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal membatalkan dokumen');
		} finally {
			isCancelling = false;
		}
	}
</script>

<svelte:head>
	<title>{movement?.code || 'Detail'} - Inventory</title>
</svelte:head>

<LayoutWrapper>
	<div class="page page-with-bottombar">
		<div class="flex items-center gap-3 mb-6">
			<button class="btn btn-ghost btn-sm" on:click={() => history.back()}>←</button>
			<h1 class="page-title">{movement?.code || 'Detail Pergerakan'}</h1>
		</div>

		{#if isLoading}
			<div class="flex h-40 items-center justify-center">
				<Loader size={48} />
			</div>
		{:else if movement}
			<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-4">
				<div class="flex flex-wrap items-center gap-3">
					<span class="badge badge-lg capitalize {typeBadge[movement.movement_type] || 'badge-ghost'}">
						{movement.movement_type}
					</span>
					<span class="metric text-sm text-base-content/60">{movement.sub_type}</span>
					<span class="text-xs text-base-content/50">{fmtDate(movement.created_at)}</span>
					{#if movement.status === 'cancelled'}
						<span class="badge badge-ghost badge-sm">dibatalkan</span>
					{/if}
					{#if movement.status !== 'cancelled' && $authStore.user?.role?.name !== 'staff_gudang'}
						<button class="btn btn-ghost btn-sm ml-auto text-error" on:click={() => (showCancel = true)}>
							Batalkan dokumen
						</button>
					{/if}
				</div>
				<dl class="mt-3 grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
					<div>
						<dt class="text-xs text-base-content/60">Gudang asal</dt>
						<dd>{movement.warehouse?.name || '-'}</dd>
					</div>
					{#if movement.movement_type === 'transfer'}
						<div>
							<dt class="text-xs text-base-content/60">Gudang tujuan</dt>
							<dd>{movement.to_warehouse?.name || '-'}</dd>
						</div>
					{/if}
					<div>
						<dt class="text-xs text-base-content/60">No. referensi</dt>
						<dd class="metric">{movement.reference_code || '-'}</dd>
					</div>
					<div>
						<dt class="text-xs text-base-content/60">Dibuat oleh</dt>
						<dd>{movement.created_by_user?.full_name || movement.created_by_user?.username || '-'}</dd>
					</div>
					{#if movement.note}
						<div class="col-span-2">
							<dt class="text-xs text-base-content/60">Catatan</dt>
							<dd>{movement.note}</dd>
						</div>
					{/if}
				</dl>
			</div>

			<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
				<h2 class="mb-3 font-semibold">Item ({movement.items?.length || 0})</h2>
				<div class="overflow-x-auto">
					<table class="table table-sm">
						<thead>
							<tr>
								<th>Produk</th>
								<th>Kode</th>
								<th class="text-right">Jumlah</th>
								<th class="text-right">Harga satuan</th>
							</tr>
						</thead>
						<tbody>
							{#each movement.items || [] as it}
								<tr>
									<td>{it.product?.name || '-'}</td>
									<td class="metric text-xs">{it.product?.product_code || '-'}</td>
									<td class="text-right">{it.quantity}</td>
									<td class="metric text-right">{it.unit_cost}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{:else}
			<p class="py-8 text-center text-sm text-base-content/60">Data tidak ditemukan</p>
		{/if}
		<ConfirmDialog
			isOpen={showCancel}
			title="Batalkan dokumen?"
			message={`Dokumen ${movement?.code ?? ''} akan dibatalkan dan stok dikembalikan ke posisi sebelum dokumen.
Dokumen tetap tersimpan sebagai riwayat.`}
			confirmLabel="Ya, batalkan"
			danger
			isLoading={isCancelling}
			on:confirm={confirmCancel}
			on:cancel={() => !isCancelling && (showCancel = false)}
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
