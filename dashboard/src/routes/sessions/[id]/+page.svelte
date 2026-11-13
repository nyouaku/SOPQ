<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { WS_BASE_URL } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	let isLoading = true;
	let session: any = null;
	let details: any[] = [];
	let filter = {
		page: 1,
		limit: 50
	};
	let total = 0;
	let ws: WebSocket | null = null;

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
			} else if (data.type === 'session_update') {
				session = { ...session, ...data.session };
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

	async function completeSession() {
		if (!confirm('Yakin ingin menyelesaikan sesi ini?')) return;

		try {
			const response = await apiClient.post(
				`/stock-opname/sessions/${sessionId}/complete`,
				{}
			);
			if (response.success) {
				await loadSession();
				toastStore.success('Sesi berhasil diselesaikan');
			}
		} catch (error: any) {
			toastStore.error('Gagal menyelesaikan sesi: ' + (error.message || 'Unknown error'));
		}
	}

	async function approveSession() {
		if (!confirm('Yakin ingin menyetujui sesi ini?')) return;

		try {
			const response = await apiClient.post(
				`/stock-opname/sessions/${sessionId}/approve`,
				{}
			);
			if (response.success) {
				await loadSession();
				toastStore.success('Sesi berhasil disetujui');
			}
		} catch (error: any) {
			toastStore.error('Gagal menyetujui sesi: ' + (error.message || 'Unknown error'));
		}
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
	});
</script>

<svelte:head>
	<title>Session Details - Stock Opname</title>
</svelte:head>

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
								<button class="btn btn-success btn-sm" on:click={completeSession}>
									Complete Session
								</button>
							{/if}
							{#if session.status === 'completed' && ($authStore.user?.role?.name === 'supervisor' || $authStore.user?.role?.name === 'admin')}
								<button class="btn btn-primary btn-sm" on:click={approveSession}>
									Approve Session
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
				<div class="flex justify-between items-center p-4">
					<h3 class="text-xl font-bold">Scan Details</h3>
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

				{#if isLoading}
					<div class="flex justify-center items-center h-64">
						<Loader size={48} />
					</div>
				{:else if details.length === 0}
					<div class="p-8 text-center text-base-content/50">No scans yet</div>
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
