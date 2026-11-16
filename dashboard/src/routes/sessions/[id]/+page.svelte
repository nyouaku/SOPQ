<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import SaveNotificationToast from '$lib/components/SaveNotificationToast.svelte';
	import SessionCompletionDialog from '$lib/components/SessionCompletionDialog.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import ScanTimeline from '$lib/components/ScanTimeline.svelte';
	import { apiClient } from '$lib/api/client';
	import { WS_BASE_URL } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';
	import { playSound } from '$lib/stores/soundSettings';

	let isLoading = true;
	let session: any = null;
	let details: any[] = [];
	let filter = {
		page: 1,
		limit: 50
	};
	let total = 0;
	let ws: WebSocket | null = null;
	let showNotification = false;
	let notificationMessage = 'Data tersimpan';
	let scannedCount = 0;
	let showCompletionDialog = false;
	let summary: any = null;
	let toastTimer: ReturnType<typeof setTimeout> | null = null;
	let detailView: 'table' | 'timeline' = 'table';

	// Flag mencegah double-submit pada aksi state sesi yang irreversible
	// (complete → approve). Tanpa ini, klik ganda = transisi duplikat.
	let isMutating = false;

	// Dialog konfirmasi berbahaya (mengganti window.confirm bawaan browser).
	// action menyimpan aksi tertunda ('complete'|'approve'|'reopen'); reject
	// butuh alasan sehingga memakai dialog textarea terpisah.
	let showActionConfirm = false;
	let pendingAction: 'complete' | 'approve' | 'reopen' | null = null;
	let showRejectDialog = false;
	let rejectReason = '';
	let rejectError = '';

	$: sessionId = $page.params.id;
	$: totalPages = Math.ceil(total / filter.limit);

	onMount(async () => {
		await loadSession();
		await loadDetails();
		connectWebSocket();
	});

	async function loadSession() {
		try {
			const response = await apiClient.get(`/stock-opname/sessions/${sessionId}`);
			if (response.success && response.data) {
				session = response.data;
			}
		} catch (error) {
			console.error('Failed to load session:', error);
		}
	}

	async function loadDetails() {
		try {
			isLoading = true;
			const response = await apiClient.get<{ details: any[]; total: number }>(
				`/stock-opname/sessions/${sessionId}/details?page=${filter.page}&limit=${filter.limit}`
			);
			if (response.success && response.data) {
				details = response.data.details || [];
				total = response.data.total || 0;
			}
		} catch (error) {
			console.error('Failed to load details:', error);
		} finally {
			isLoading = false;
		}
	}

	function connectWebSocket() {
		const token = localStorage.getItem('access_token');
		if (!token) return;

		// Token dikirim lewat sub-protokol handshake, BUKAN query string:
		// query string tercatat di access log nginx dan riwayat browser.
		// Server menjawab dengan sub-protokol yang sama agar handshake sah.
		ws = new WebSocket(`${WS_BASE_URL}/ws/stock-opname/sessions/${sessionId}`, [
			'access_token',
			token
		]);

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data);
			if (data.type === 'new_scan') {
				// Prepend new scan to list
				details = [data.detail, ...details];
				total++;

				// Show notification toast dengan auto-dismiss
				scannedCount = total;
				notificationMessage = `${data.detail?.product?.name || 'Produk'} disimpan`;
				playSound('scan');
				showNotification = true;
				if (toastTimer) clearTimeout(toastTimer);
				toastTimer = setTimeout(() => {
					showNotification = false;
				}, 2000);

				// Semua produk sudah ter-scan? Tampilkan dialog hasil + tombol ke report.
				// Guard total_products > 0: sesi tanpa target tak memicu dialog palsu.
				if (session?.total_products > 0 && total >= session.total_products) {
					notificationMessage = 'Semua produk sudah ter-scan!';
					if (toastTimer) clearTimeout(toastTimer);
					playSound('success');
					loadSummaryAndShowDialog();
				}
			} else if (data.type === 'discrepancy_alert') {
				// Alert selisih besar — toast merah, auto-dismiss 4 detik
				notificationMessage = `⚠️ ${data.product}: selisih ${data.difference} (${data.percent.toFixed(0)}%)`;
				showNotification = true;
				if (toastTimer) clearTimeout(toastTimer);
				toastTimer = setTimeout(() => {
					showNotification = false;
				}, 4000);
			} else if (data.type === 'session_update') {
				const prevStatus = session?.status;
				session = { ...session, ...data.session };
				// Auto-show completion dialog ketika session baru selesai
				if (
					prevStatus === 'in_progress' &&
					(data.session.status === 'completed' || data.session.status === 'approved')
				) {
					playSound('success');
					loadSummaryAndShowDialog();
				}
			}
		};

		ws.onerror = (error) => {
			console.error('WebSocket error:', error);
		};

		ws.onclose = () => {
			console.log('WebSocket closed');
		};
	}

	function getStatusBadge(status: string) {
		const badges: Record<string, string> = {
			in_progress: 'badge-info',
			completed: 'badge-success',
			approved: 'badge-primary',
			cancelled: 'badge-error'
		};
		return badges[status] || 'badge-ghost';
	}

	function getDifferenceBadge(difference: number) {
		if (difference === 0) return 'badge-success';
		if (difference > 0) return 'badge-warning';
		return 'badge-error';
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleString('id-ID', {
			day: '2-digit',
			month: 'short',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	// Buka dialog konfirmasi (tanpa alasan) lalu jalankan aksinya.
	function requestAction(action: 'complete' | 'approve' | 'reopen') {
		if (isMutating) return;
		pendingAction = action;
		showActionConfirm = true;
	}

	async function confirmAction() {
		const action = pendingAction;
		if (!action) return;
		isMutating = true;
		try {
			const response = await apiClient.post(
				`/stock-opname/sessions/${sessionId}/${action}`,
				{}
			);
			if (response.success) {
				await loadSession();
				const msg = action === 'complete'
					? 'Sesi berhasil diselesaikan'
					: action === 'approve'
					? 'Sesi berhasil disetujui'
					: 'Sesi berhasil dibuka kembali';
				toastStore.success(msg);
			} else {
				toastStore.error(response.error?.message || 'Gagal memproses sesi');
			}
		} catch (error: any) {
			toastStore.error('Gagal memproses sesi: ' + (error.message || 'Unknown error'));
		} finally {
			isMutating = false;
			showActionConfirm = false;
			pendingAction = null;
		}
	}

	// Reject butuh alasan — pakai dialog textarea, bukan window.prompt.
	function openRejectDialog() {
		if (isMutating) return;
		rejectReason = '';
		rejectError = '';
		showRejectDialog = true;
	}

	async function confirmReject() {
		const reason = rejectReason.trim();
		if (reason.length < 3) {
			rejectError = 'Alasan penolakan minimal 3 karakter';
			return;
		}
		isMutating = true;
		try {
			const response = await apiClient.post(
				`/stock-opname/sessions/${sessionId}/reject`,
				{ reason }
			);
			if (response.success) {
				await loadSession();
				toastStore.success('Sesi berhasil ditolak');
			} else {
				toastStore.error(response.error?.message || 'Gagal menolak sesi');
			}
		} catch (error: any) {
			toastStore.error('Gagal menolak sesi: ' + (error.message || 'Unknown error'));
		} finally {
			isMutating = false;
			showRejectDialog = false;
		}
	}

	async function loadSummaryAndShowDialog() {
		try {
			const response = await apiClient.get(`/stock-opname/sessions/${sessionId}/summary`);
			if (response.success && response.data) {
				summary = response.data;
				showCompletionDialog = true;
			}
		} catch (error) {
			console.error('Failed to load summary:', error);
		}
	}

	function handleViewReport() {
		showCompletionDialog = false;
		// goto (bukan window.location) supaya SPA tidak reload penuh dan tetap membawa auth.
		goto(`/reports?session_id=${sessionId}`);
	}

	function changePage(newPage: number) {
		if (newPage < 1 || newPage > totalPages) return;
		filter.page = newPage;
		loadDetails();
	}

	onDestroy(() => {
		if (ws) {
			ws.close();
		}
		if (toastTimer) clearTimeout(toastTimer);
	});
</script>

<svelte:head>
	<title>Session Details - Stock Opname</title>
</svelte:head>

{#if showNotification}
	<SaveNotificationToast
		message={notificationMessage}
		scannedCount={scannedCount}
		totalProducts={session?.total_products || 0}
	/>
{/if}

{#if showCompletionDialog && summary}
	<SessionCompletionDialog
		isOpen={showCompletionDialog}
		totalScanned={summary.total_scanned}
		totalProducts={summary.total_products}
		matched={summary.matched_count}
		unmatched={summary.unmatched_count}
		startedAt={new Date(summary.started_at)}
		onClose={() => (showCompletionDialog = false)}
		onReport={handleViewReport}
	/>
{/if}

<!-- Dialog konfirmasi aksi sesi: complete / approve / reopen -->
<ConfirmDialog
	isOpen={showActionConfirm}
	title={pendingAction === 'complete'
		? 'Selesaikan sesi?'
		: pendingAction === 'approve'
		? 'Setujui sesi?'
		: 'Buka kembali sesi?'}
	message={pendingAction === 'complete'
		? 'Sesi akan berubah menjadi completed. Langkah ini menandai hitung fisik selesai.'
		: pendingAction === 'approve'
		? 'Sesi akan disetujui dan hasil hitung fisik menjadi stok sistem resmi. Tindakan ini tidak dapat dibatalkan.'
		: 'Sesi yang ditolak akan dibuka kembali untuk dikerjakan ulang.'}
	confirmLabel={pendingAction === 'complete'
		? 'Selesaikan'
		: pendingAction === 'approve'
		? 'Setujui'
		: 'Buka kembali'}
	danger={pendingAction === 'approve'}
	isLoading={isMutating}
	on:confirm={confirmAction}
	on:cancel={() => { showActionConfirm = false; pendingAction = null; }}
/>

<!-- Dialog tolak sesi: wajib isi alasan -->
<ConfirmDialog
	isOpen={showRejectDialog}
	title="Tolak sesi?"
	message={'Berikan alasan penolakan (minimal 3 karakter).\n' + (rejectError || '')}
	confirmLabel="Tolak Sesi"
	cancelLabel="Batal"
	danger={true}
	isLoading={isMutating}
	on:confirm={confirmReject}
	on:cancel={() => { showRejectDialog = false; rejectReason = ''; rejectError = ''; }}
>
	<textarea
		class="textarea textarea-bordered w-full mt-3"
		placeholder="Alasan penolakan..."
		bind:value={rejectReason}
		rows="3"
	></textarea>
</ConfirmDialog>

<LayoutWrapper>
	<div class="page">
		<!-- Breadcrumb -->
		<div class="text-sm breadcrumbs mb-6">
			<ul>
				<li><a href="/">Dashboard</a></li>
				<li><a href="/sessions">Sessions</a></li>
				<li>{session?.code || sessionId}</li>
			</ul>
		</div>

		{#if session}
			<!-- Session Info Card -->
			<div class="card bg-base-200 shadow-xl mb-6">
				<div class="card-body">
					<div class="flex justify-between items-start">
						<div>
							<h2 class="card-title text-2xl font-mono">{session.code}</h2>
							<div class="mt-2 space-y-1">
								<p class="text-sm">
									<span class="text-base-content/70">Warehouse:</span>
									<span class="font-semibold ml-2">{session.warehouse?.name || '-'}</span>
								</p>
								<p class="text-sm">
									<span class="text-base-content/70">Location:</span>
									<span class="font-semibold ml-2">{session.location?.name || 'All'}</span>
								</p>
								<p class="text-sm">
									<span class="text-base-content/70">Started by:</span>
									<span class="font-semibold ml-2"
										>{session.started_user?.full_name || '-'}</span
									>
								</p>
								<p class="text-sm">
									<span class="text-base-content/70">Started at:</span>
									<span class="font-semibold ml-2">{formatDate(session.started_at)}</span>
								</p>
								{#if session.completed_at}
									<p class="text-sm">
										<span class="text-base-content/70">Completed at:</span>
										<span class="font-semibold ml-2">{formatDate(session.completed_at)}</span>
									</p>
								{/if}
								{#if session.approved_at}
									<p class="text-sm">
										<span class="text-base-content/70">Approved by:</span>
										<span class="font-semibold ml-2"
											>{session.approved_user?.full_name || '-'}</span
										>
									</p>
									<p class="text-sm">
										<span class="text-base-content/70">Approved at:</span>
										<span class="font-semibold ml-2">{formatDate(session.approved_at)}</span>
									</p>
								{/if}
							</div>
						</div>
						<div class="flex flex-col gap-2 items-end">
							<span class="badge {getStatusBadge(session.status)} badge-lg capitalize">
								{session.status.replace('_', ' ')}
							</span>
							{#if session.status === 'in_progress'}
								<button
									class="btn btn-success btn-sm"
									disabled={isMutating}
									on:click={() => requestAction('complete')}
								>
									{#if isMutating}<span class="loading loading-spinner loading-sm"></span>{/if}
									Complete Session
								</button>
							{/if}
							{#if session.status === 'completed' && ($authStore.user?.role?.name === 'supervisor' || $authStore.user?.role?.name === 'admin')}
								<button
									class="btn btn-primary btn-sm"
									disabled={isMutating}
									on:click={() => requestAction('approve')}
								>
									{#if isMutating}<span class="loading loading-spinner loading-sm"></span>{/if}
									Approve Session
								</button>
								<button
									class="btn btn-error btn-sm"
									disabled={isMutating}
									on:click={openRejectDialog}
								>
									{#if isMutating}<span class="loading loading-spinner loading-sm"></span>{/if}
									Reject Session
								</button>
							{/if}
							{#if session.status === 'rejected' && $authStore.user?.role?.name === 'admin'}
								<button
									class="btn btn-warning btn-sm"
									disabled={isMutating}
									on:click={() => requestAction('reopen')}
								>
									{#if isMutating}<span class="loading loading-spinner loading-sm"></span>{/if}
									Reopen Session
								</button>
							{/if}
						</div>
					</div>
				</div>
			</div>

			<!-- Stats -->
			<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
				<div class="stats shadow bg-base-200">
					<div class="stat">
						<div class="stat-title">Total Scanned</div>
						<div class="stat-value text-primary">{total}</div>
					</div>
				</div>
				<div class="stats shadow bg-base-200">
					<div class="stat">
						<div class="stat-title">Match</div>
						<div class="stat-value text-success">
							{details.filter((d) => d.difference === 0).length}
						</div>
					</div>
				</div>
				<div class="stats shadow bg-base-200">
					<div class="stat">
						<div class="stat-title">Over Stock</div>
						<div class="stat-value text-warning">
							{details.filter((d) => d.difference > 0).length}
						</div>
					</div>
				</div>
				<div class="stats shadow bg-base-200">
					<div class="stat">
						<div class="stat-title">Under Stock</div>
						<div class="stat-value text-error">
							{details.filter((d) => d.difference < 0).length}
						</div>
					</div>
				</div>
			</div>
		{/if}

		<!-- Details Table -->
		<div class="card bg-base-200 shadow-xl">
			<div class="card-body p-0">
				<div class="flex flex-wrap items-center justify-between gap-3 p-4">
					<h3 class="text-xl font-bold">Scan Details</h3>
					<div class="flex items-center gap-2">
						<!-- View Toggle -->
						<div class="btn-group">
							<button
								class="btn btn-sm {detailView === 'table' ? 'btn-active' : ''}"
								on:click={() => (detailView = 'table')}
							>
								<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
								</svg>
								Tabel
							</button>
							<button
								class="btn btn-sm {detailView === 'timeline' ? 'btn-active' : ''}"
								on:click={() => (detailView = 'timeline')}
							>
								<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
								Timeline
							</button>
						</div>
						<select
							class="select select-bordered select-sm"
							bind:value={filter.limit}
							on:change={() => {
								filter.page = 1;
								loadDetails();
							}}
						>
							<option value={20}>20 per page</option>
							<option value={50}>50 per page</option>
							<option value={100}>100 per page</option>
						</select>
					</div>
				</div>

				{#if isLoading}
					<div class="flex justify-center items-center h-64">
						<Loader size={48} />
					</div>
				{:else if details.length === 0}
					<div class="p-8 text-center text-base-content/50">No scans yet</div>
				{:else if detailView === 'timeline'}
					<div class="p-4">
						<ScanTimeline details={details} view="timeline" />
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="table table-zebra table-sm">
							<thead>
								<tr>
									<th>Scanned At</th>
									<th>Product</th>
											<th>Box</th>
									<th>System</th>
									<th>Physical</th>
									<th>Difference</th>
									<th>Method</th>
									<th>Scanned By</th>
								</tr>
							</thead>
							<tbody>
								{#each details as detail}
									<tr>
										<td class="text-xs">{formatDate(detail.scanned_at)}</td>
										<td>
											<div class="font-semibold">{detail.product?.name || '-'}</div>
											<div class="text-xs text-base-content/70">
											{detail.product?.product_code || '-'}
											</div>
										</td>
													<td class="text-xs">{detail.box?.code || '-'}</td>
										<td class="text-right">{detail.stock_system}</td>
										<td class="text-right">{detail.stock_physical}</td>
										<td class="text-right">
											<span class="badge {getDifferenceBadge(detail.difference)} badge-sm">
												{detail.difference > 0 ? '+' : ''}{detail.difference}
											</span>
										</td>
										<td>
											<span class="badge badge-ghost badge-sm capitalize">
												{detail.input_method}
											</span>
										</td>
										<td class="text-xs">{detail.scanned_user?.full_name || '-'}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>

					<!-- Pagination -->
					{#if totalPages > 1}
						<div class="flex justify-center p-4">
							<div class="join">
								<button
									class="join-item btn btn-sm"
									disabled={filter.page === 1}
									on:click={() => changePage(filter.page - 1)}
								>
									«
								</button>
								<button class="join-item btn btn-sm">
									Page {filter.page} of {totalPages}
								</button>
								<button
									class="join-item btn btn-sm"
									disabled={filter.page === totalPages}
									on:click={() => changePage(filter.page + 1)}
								>
									»
								</button>
							</div>
						</div>
					{/if}
				{/if}
			</div>
		</div>
	</div>
</LayoutWrapper>
