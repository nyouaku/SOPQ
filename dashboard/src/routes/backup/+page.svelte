<script lang="ts">
	import { onMount } from 'svelte';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { authStore } from '$lib/stores/auth';

	let status: any = null;
	let isLoading = true;
	let isDownloading = false;

	$: isAdmin = $authStore.user?.role?.name === 'admin';

	onMount(async () => {
		if (!isAdmin) return;
		await loadStatus();
		isLoading = false;
	});

	async function loadStatus() {
		try {
			const res = await apiClient.get(API_ENDPOINTS.backupStatus);
			if (res.success && res.data) status = res.data;
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat status backup');
		}
	}

	async function downloadBackup() {
		isDownloading = true;
		try {
			const token = localStorage.getItem('access_token');
			const res = await fetch(`${API_ENDPOINTS.base}${API_ENDPOINTS.backupDownload}`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!res.ok) {
				throw new Error('Gagal mengunduh backup');
			}
			const blob = await res.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `stock_opname_backup_${new Date().toISOString().slice(0, 10)}.sql`;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);
			toastStore.success('Backup berhasil diunduh');
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal mengunduh backup');
		} finally {
			isDownloading = false;
		}
	}

	function formatSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function formatDate(iso: string): string {
		if (!iso) return '-';
		return new Date(iso).toLocaleString('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<svelte:head>
	<title>Backup & Restore - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="mb-6">
			<h1 class="page-title">Backup & Restore</h1>
			<p class="mt-1 text-sm text-base-content/60">
				Kelola backup database dan pantau status penyimpanan
			</p>
		</div>

		{#if !isAdmin}
			<div class="alert alert-error">
				<span>Hanya admin yang dapat mengakses halaman ini.</span>
			</div>
		{:else if isLoading}
			<div class="flex h-48 items-center justify-center">
				<Loader size={48} />
			</div>
		{:else}
			<!-- Status Cards -->
			<div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-3">
				<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Ukuran Database</div>
					<div class="metric mt-1 text-2xl font-bold">{status?.database_size || '-'}</div>
				</div>
				<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Backup Terakhir</div>
					<div class="mt-1 text-lg font-semibold">
						{status?.last_backup_time ? formatDate(status.last_backup_time) : 'Belum ada'}
					</div>
					{#if status?.last_backup_size}
						<div class="text-xs text-base-content/50">{formatSize(status.last_backup_size)}</div>
					{/if}
				</div>
				<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
					<div class="text-xs text-base-content/60">Status</div>
					<div class="mt-1">
						{#if status?.last_backup_time}
							<span class="badge badge-success badge-lg">Siap</span>
						{:else}
							<span class="badge badge-warning badge-lg">Belum di-backup</span>
						{/if}
					</div>
				</div>
			</div>

			<!-- Actions -->
			<div class="rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
				<h2 class="mb-3 font-semibold">Aksi Backup</h2>
				<p class="mb-4 text-sm text-base-content/60">
					Download backup database dalam format SQL. File dapat digunakan untuk restore di server lain.
				</p>
				<button
					class="btn btn-primary"
					on:click={downloadBackup}
					disabled={isDownloading}
				>
					{#if isDownloading}
						<span class="loading loading-spinner loading-sm"></span>
						Membuat backup...
					{:else}
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 10v6m0 0l-3-3m3 3l3-3M3 17v2a2 2 0 002 2h14a2 2 0 002-2v-2" />
						</svg>
						Download Backup (.sql)
					{/if}
				</button>

				<div class="mt-4 rounded-xl border border-warning/40 bg-warning/10 p-3">
					<p class="text-xs text-warning">
						⚠️ Simpan file backup di lokasi aman. Disarankan backup berkala minimal 1x seminggu.
					</p>
				</div>
			</div>

			<!-- Restore Instructions -->
			<div class="mt-6 rounded-2xl border border-base-300 bg-base-200 p-4 md:p-5">
				<h2 class="mb-3 font-semibold">Cara Restore</h2>
				<ol class="list-inside list-decimal space-y-2 text-sm text-base-content/70">
					<li>Download file backup (.sql) dari tombol di atas.</li>
					<li>Buka terminal di server VPS.</li>
					<li>Jalankan perintah berikut:
						<code class="ml-2 rounded bg-base-300 px-2 py-0.5 font-mono text-xs">
							docker exec -i stock-opname-app-postgres-1 psql -U stockopname stockopname_db < backup.sql
						</code>
					</li>
					<li>Restart backend: <code class="rounded bg-base-300 px-2 py-0.5 font-mono text-xs">docker compose restart backend</code></li>
				</ol>
				<div class="mt-3 rounded-xl border border-error/40 bg-error/10 p-3">
					<p class="text-xs text-error">
						⚠️ Restore akan menimpa semua data yang ada. Pastikan sudah backup terbaru sebelum restore.
					</p>
				</div>
			</div>
		{/if}
	</div>
</LayoutWrapper>
