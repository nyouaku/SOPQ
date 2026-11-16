<script lang="ts">
	import { onMount } from 'svelte';
	import { QueryClient } from '@tanstack/svelte-query';
	import { notificationsQuery, queryKeys } from '$lib/api/queries';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import Loader from '$lib/components/Loader.svelte';
	import { getQueryClient } from '$lib/api/queryClient';
	import { toastStore } from '$lib/stores/toast';

	const query = notificationsQuery();
	let open = false;
	let marking: string | null = null;
	let isMarkingAll = false;

	const client: QueryClient = getQueryClient();

	$: notifications = ($query.data?.notifications as any[]) || [];
	$: unread = notifications.filter((n) => !n.is_read).length;

	async function markRead(id: string) {
		marking = id;
		try {
			await apiClient.patch(`${API_ENDPOINTS.notifications}/${id}/read`, {});
			client.invalidateQueries({ queryKey: queryKeys.notifications });
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal menandai notifikasi');
		} finally {
			marking = null;
		}
	}

	async function markAll() {
		const pending = notifications.filter((n) => !n.is_read);
		if (pending.length === 0) return;
		isMarkingAll = true;
		try {
			await Promise.all(
				pending.map((n) => apiClient.patch(`${API_ENDPOINTS.notifications}/${n.id}/read`, {}))
			);
			client.invalidateQueries({ queryKey: queryKeys.notifications });
			toastStore.success(`${pending.length} notifikasi ditandai sudah dibaca`);
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal menandai semua notifikasi');
		} finally {
			isMarkingAll = false;
		}
	}

	function relativeTime(iso: string): string {
		if (!iso) return '';
		const diff = Date.now() - new Date(iso).getTime();
		const min = Math.floor(diff / 60000);
		if (min < 1) return 'baru saja';
		if (min < 60) return `${min} menit lalu`;
		const hour = Math.floor(min / 60);
		if (hour < 24) return `${hour} jam lalu`;
		const day = Math.floor(hour / 24);
		if (day < 7) return `${day} hari lalu`;
		return new Date(iso).toLocaleDateString('id-ID');
	}

	onMount(() => {
		const onKey = (e: KeyboardEvent) => {
			if (e.key === 'Escape') open = false;
		};
		window.addEventListener('keydown', onKey);
		return () => window.removeEventListener('keydown', onKey);
	});
</script>

<div class="relative">
	<button
		class="btn btn-ghost btn-sm btn-square relative"
		title="Notifikasi"
		aria-label={unread > 0 ? `Notifikasi, ${unread} belum dibaca` : 'Notifikasi'}
		aria-expanded={open}
		on:click={() => (open = !open)}
	>
		<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
			<path d="M18 8a6 6 0 1 0-12 0c0 6-2 7-2 7h16s-2-1-2-7" stroke-linecap="round" stroke-linejoin="round" />
			<path d="M10.5 20a2 2 0 0 0 3 0" stroke-linecap="round" />
		</svg>
		{#if unread > 0}
			<span class="badge badge-primary badge-xs absolute right-1 top-1">
				{unread > 9 ? '9+' : unread}
			</span>
		{/if}
	</button>

	{#if open}
		<button
			class="fixed inset-0 z-40 cursor-default"
			aria-label="Tutup notifikasi"
			on:click={() => (open = false)}
		></button>

		<div
			class="absolute right-0 z-50 mt-2 w-[min(20rem,calc(100vw-2rem))] overflow-hidden rounded-2xl border border-base-300 bg-base-100 shadow-xl"
			role="dialog"
			aria-label="Daftar notifikasi"
		>
			<div class="flex items-center justify-between border-b border-base-300 px-4 py-2.5">
				<span class="text-sm font-semibold">Notifikasi</span>
				{#if unread > 0}
					<button class="btn btn-ghost btn-xs" on:click={markAll} disabled={isMarkingAll}>
						{#if isMarkingAll}
							<span class="loading loading-spinner loading-xs"></span>
						{/if}
						Tandai semua
					</button>
				{/if}
			</div>

			<div class="max-h-80 overflow-y-auto">
				{#if $query.isLoading}
					<div class="flex justify-center py-8"><Loader size={24} /></div>
				{:else if $query.isError}
					<p class="px-4 py-6 text-center text-xs text-error">
						Gagal memuat notifikasi
					</p>
				{:else if notifications.length === 0}
					<p class="px-4 py-8 text-center text-xs text-base-content/50">
						Tidak ada notifikasi
					</p>
				{:else}
					<ul class="divide-y divide-base-300">
						{#each notifications as n (n.id)}
							<li>
								<button
									class="flex w-full gap-3 px-4 py-3 text-left hover:bg-base-200"
									on:click={() => !n.is_read && markRead(n.id)}
								>
									<span
										class="mt-1.5 h-2 w-2 shrink-0 rounded-full"
										class:bg-primary={!n.is_read}
										class:bg-base-300={n.is_read}
									></span>
									<span class="min-w-0 flex-1">
										<span class="block truncate text-sm font-medium">
											{n.title || n.type || 'Notifikasi'}
										</span>
										{#if n.message}
											<span class="mt-0.5 block line-clamp-2 text-xs text-base-content/60">
												{n.message}
											</span>
										{/if}
										<span class="mt-1 block text-[11px] text-base-content/40">
											{relativeTime(n.created_at)}
										</span>
									</span>
									{#if marking === n.id}
										<Loader size={14} />
									{/if}
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</div>
	{/if}
</div>
