<script lang="ts">
	import { onMount } from 'svelte';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import { apiClient } from '$lib/api/client';
	import { toastStore } from '$lib/stores/toast';
	import { authStore } from '$lib/stores/auth';

	type Request = {
		id: string;
		change_type: string;
		product_id?: string;
		payload: any;
		reason: string;
		status: string;
		requester?: { full_name?: string; username?: string };
		reviewed_at?: string;
		review_note?: string;
		created_at: string;
	};

	let rows: Request[] = [];
	let total = 0;
	let status = 'pending';
	let loading = true;
	let busy: string | null = null;
	let showSubmit = false;
	let form = { change_type: 'update', product_id: '', product_code: '', name: '', unit: 'PCS', cost_price: 0, minimum_stock: 0, reason: '' };

	async function load() {
		loading = true;
		try {
			const res = await apiClient.get<{ requests: Request[]; total: number }>(
				`/product-change-requests?status=${status}&limit=100`
			);
			rows = res.success && res.data ? res.data.requests || [] : [];
			total = res.data?.total || 0;
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal memuat pengajuan');
		} finally {
			loading = false;
		}
	}

	async function submitChange() {
		busy = 'submit';
		try {
			const payload: any = {
				product_code: form.product_code,
				name: form.name,
				unit: form.unit,
				cost_price: form.cost_price,
				minimum_stock: form.minimum_stock
			};
			const res = await apiClient.post('/product-change-requests', {
				change_type: form.change_type,
				product_id: form.change_type === 'create' ? null : form.product_id || null,
				payload,
				reason: form.reason
			});
			if (!res.success) throw new Error(res.error?.message || 'Gagal mengajukan');
			toastStore.success('Pengajuan terkirim');
			showSubmit = false;
			status = 'pending';
			await load();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal mengajukan');
		} finally {
			busy = null;
		}
	}

	async function review(id: string, approve: boolean) {
		busy = id;
		try {
			let note = '';
			if (!approve) {
				note = window.prompt('Alasan penolakan (opsional):') || '';
			}
			const res = await apiClient.post(`/product-change-requests/${id}/${approve ? 'approve' : 'reject'}`, { note });
			if (!res.success) throw new Error(res.error?.message || 'Gagal memproses');
			toastStore.success(approve ? 'Pengajuan disetujui' : 'Pengajuan ditolak');
			await load();
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal memproses');
		} finally {
			busy = null;
		}
	}

	function label(t: string) {
		return t === 'create' ? 'Tambah produk' : t === 'update' ? 'Ubah produk' : 'Hapus produk';
	}
	function fmtDate(s: string) {
		return s ? new Date(s).toLocaleString('id-ID') : '-';
	}
	function payloadSummary(r: Request) {
		const p = typeof r.payload === 'string' ? JSON.parse(r.payload) : r.payload;
		return `${p.product_code || '-'} — ${p.name || '-'}`;
	}

	$: canReview = ['admin', 'supervisor'].includes($authStore.user?.role?.name ?? '');

	onMount(load);
</script>

<svelte:head>
	<title>Persetujuan Produk - SOPQ</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="page-header">
			<div>
				<h1 class="page-title">Persetujuan Produk</h1>
				<p class="page-subtitle">Tinjau pengajuan perubahan master produk dari staff</p>
			</div>
			<div class="flex gap-2">
				<button class="btn btn-primary btn-sm" on:click={() => (showSubmit = !showSubmit)}>
					{showSubmit ? 'Tutup form' : '+ Ajukan perubahan'}
				</button>
				<select class="select select-bordered select-sm" bind:value={status} on:change={load}>
					<option value="pending">Menunggu</option>
					<option value="approved">Disetujui</option>
					<option value="rejected">Ditolak</option>
					<option value="">Semua</option>
				</select>
			</div>
		</div>

		{#if showSubmit}
			<div class="mb-5 rounded-2xl border border-base-300 bg-base-200 p-4">
				<h2 class="mb-3 font-semibold">Form pengajuan</h2>
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
					<div>
						<label class="label py-0" for="ct"><span class="label-text text-xs">Jenis</span></label>
						<select id="ct" class="select select-bordered select-sm w-full" bind:value={form.change_type}>
							<option value="update">Ubah produk</option>
							<option value="create">Tambah produk</option>
							<option value="delete">Hapus produk</option>
						</select>
					</div>
					{#if form.change_type !== 'create'}
						<div>
							<label class="label py-0" for="pid"><span class="label-text text-xs">Product ID</span></label>
							<input id="pid" class="input input-bordered input-sm w-full" bind:value={form.product_id} placeholder="UUID produk" />
						</div>
					{/if}
					<div>
						<label class="label py-0" for="pcode"><span class="label-text text-xs">Kode produk</span></label>
						<input id="pcode" class="input input-bordered input-sm w-full" bind:value={form.product_code} />
					</div>
					<div>
						<label class="label py-0" for="pname"><span class="label-text text-xs">Nama produk</span></label>
						<input id="pname" class="input input-bordered input-sm w-full" bind:value={form.name} />
					</div>
					<div>
						<label class="label py-0" for="pcost"><span class="label-text text-xs">Harga modal</span></label>
						<input id="pcost" type="number" min="0" class="input input-bordered input-sm w-full" bind:value={form.cost_price} />
					</div>
					<div>
						<label class="label py-0" for="preason"><span class="label-text text-xs">Alasan</span></label>
						<input id="preason" class="input input-bordered input-sm w-full" bind:value={form.reason} />
					</div>
				</div>
				<button class="btn btn-primary btn-sm mt-3" disabled={busy === 'submit' || !form.name || !form.product_code} on:click={submitChange}>
					{busy === 'submit' ? 'Mengirim...' : 'Kirim pengajuan'}
				</button>
			</div>
		{/if}

		{#if loading}
			<div class="flex h-40 items-center justify-center"><span class="loading loading-spinner"></span></div>
		{:else if rows.length === 0}
			<div class="rounded-2xl border border-dashed border-base-300 py-16 text-center text-sm text-base-content/60">
				Tidak ada pengajuan dengan status ini.
			</div>
		{:else}
			<p class="mb-3 text-xs text-base-content/60">{total} pengajuan</p>
			<div class="space-y-3">
				{#each rows as r}
					<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
						<div class="flex flex-wrap items-start justify-between gap-3">
							<div class="min-w-0">
								<div class="flex flex-wrap items-center gap-2">
									<span class="badge badge-primary badge-sm">{label(r.change_type)}</span>
									<span class="badge badge-ghost badge-sm capitalize">{r.status}</span>
								</div>
								<div class="mt-2 font-semibold">{payloadSummary(r)}</div>
								{#if r.reason}
									<p class="mt-1 text-sm text-base-content/70">Alasan: {r.reason}</p>
								{/if}
								<p class="mt-1 text-xs text-base-content/50">
									Diajukan oleh {r.requester?.full_name || r.requester?.username || '-'} · {fmtDate(r.created_at)}
								</p>
								{#if r.reviewed_at}
									<p class="text-xs text-base-content/50">
										Ditinjau {fmtDate(r.reviewed_at)}{r.review_note ? ` — ${r.review_note}` : ''}
									</p>
								{/if}
							</div>
							{#if r.status === 'pending' && canReview}
								<div class="flex shrink-0 gap-2">
									<button class="btn btn-success btn-sm" disabled={busy === r.id} on:click={() => review(r.id, true)}>Setujui</button>
									<button class="btn btn-ghost btn-sm text-error" disabled={busy === r.id} on:click={() => review(r.id, false)}>Tolak</button>
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</LayoutWrapper>
