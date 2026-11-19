<script lang="ts">
	import { goto } from '$app/navigation';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let sessions: any[] = [];
	let prevId = '';
	let currId = '';
	let result: any = null;
	let isLoading = false;

	async function loadSessions() {
		try {
			const res = await apiClient.get<{ sessions: any[] }>(
				`${API_ENDPOINTS.stockOpnameSessions}?limit=100`
			);
			if (res.success && res.data) sessions = res.data.sessions || [];
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal memuat sesi');
		}
	}
	loadSessions();

	async function compare() {
		if (!prevId || !currId) {
			toastStore.error('Pilih 2 sesi untuk dibandingkan');
			return;
		}
		isLoading = true;
		try {
			const res = await apiClient.get<any>(
				`/stock-opname/compare?prev=${prevId}&curr=${currId}`
			);
			if (res.success && res.data) result = res.data;
			else toastStore.error(res.error?.message || 'Gagal membandingkan');
		} catch (e: any) {
			toastStore.error(e?.message || 'Gagal membandingkan');
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Riwayat SO - Perbandingan</title>
</svelte:head>

<LayoutWrapper>
	<div class="page page-with-bottombar">
		<div class="flex items-center gap-3 mb-6">
			<button class="btn btn-ghost btn-sm" on:click={() => goto('/sessions')}>←</button>
			<h1 class="page-title">Riwayat SO — Perbandingan</h1>
		</div>

		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-4">
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<div>
					<label class="label py-0" for="c-prev"><span class="label-text text-xs">Sesi sebelumnya</span></label>
					<select id="c-prev" class="select select-bordered w-full" bind:value={prevId}>
						<option value="">-- Pilih --</option>
						{#each sessions as s}
							<option value={s.id}>{s.code} — {s.status}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="label py-0" for="c-curr"><span class="label-text text-xs">Sesi sekarang</span></label>
					<select id="c-curr" class="select select-bordered w-full" bind:value={currId}>
						<option value="">-- Pilih --</option>
						{#each sessions as s}
							<option value={s.id}>{s.code} — {s.status}</option>
						{/each}
					</select>
				</div>
			</div>
			<button class="btn btn-primary btn-sm mt-3" on:click={compare} disabled={isLoading || !prevId || !currId}>
				{#if isLoading}<span class="loading loading-spinner loading-sm"></span>{/if}
				Bandingkan
			</button>
		</div>

		{#if result}
			<div class="mb-4 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-base-300 bg-base-300 lg:grid-cols-4">
				<div class="bg-base-200 p-4"><div class="text-xs text-base-content/60">Ada di keduanya</div><div class="metric text-2xl font-semibold">{result.summary.in_both}</div></div>
				<div class="bg-base-200 p-4"><div class="text-xs text-base-content/60">Berubah</div><div class="metric text-2xl font-semibold text-warning">{result.summary.changed}</div></div>
				<div class="bg-base-200 p-4"><div class="text-xs text-base-content/60">Hanya di sebelumnya</div><div class="metric text-2xl font-semibold">{result.summary.only_in_prev}</div></div>
				<div class="bg-base-200 p-4"><div class="text-xs text-base-content/60">Hanya di sekarang</div><div class="metric text-2xl font-semibold">{result.summary.only_in_curr}</div></div>
			</div>

			<div class="overflow-x-auto rounded-2xl border border-base-300 bg-base-200 p-4">
				<table class="table table-sm">
					<thead>
						<tr><th>Produk</th><th class="text-right">Sebelumnya</th><th class="text-right">Sekarang</th><th class="text-right">Perubahan</th></tr>
					</thead>
					<tbody>
						{#each result.rows as r}
							<tr>
								<td><div class="truncate">{r.product_name}</div><div class="metric text-xs text-base-content/60">{r.product_code}</div></td>
								<td class="text-right">{r.prev_physical}</td>
								<td class="text-right">{r.curr_physical}</td>
								<td class="text-right font-medium {r.change > 0 ? 'text-success' : r.change < 0 ? 'text-error' : ''}">
									{r.change > 0 ? '+' : ''}{r.change}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</LayoutWrapper>
