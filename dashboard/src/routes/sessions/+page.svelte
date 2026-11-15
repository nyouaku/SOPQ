<script lang="ts">
	import { onMount } from 'svelte';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import TableSkeleton from '$lib/components/skeletons/TableSkeleton.svelte';
	import { goto } from '$app/navigation';
	import { apiClient } from '$lib/api/client';

	let isLoading = true;
	let sessions: any[] = [];
	let filter = {
		page: 1,
		limit: 20,
		status: '',
		warehouse_id: ''
	};
	let total = 0;

	// totalPages sebelumnya dipakai tapi tidak pernah dideklarasikan (compile error).
	$: totalPages = Math.max(1, Math.ceil(total / filter.limit));

	onMount(async () => {
		await loadSessions();
	});

	async function loadSessions() {
		try {
			isLoading = true;
			const params = new URLSearchParams({
				page: filter.page.toString(),
				limit: filter.limit.toString()
			});
			if (filter.status) params.append('status', filter.status);
			if (filter.warehouse_id) params.append('warehouse_id', filter.warehouse_id);

			const response = await apiClient.get<{ sessions: any[]; total: number }>(
				`/stock-opname/sessions?${params}`
			);
			if (response.success && response.data) {
				sessions = response.data.sessions || [];
				total = response.data.total || 0;
			}
		} catch (error) {
			console.error('Failed to load sessions:', error);
		} finally {
			isLoading = false;
		}
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

	function changePage(newPage: number) {
		if (newPage < 1 || newPage > totalPages) return;
		filter.page = newPage;
		loadSessions();
	}

	function viewDetails(sessionId: string) {
		goto(`/sessions/${sessionId}`);
	}
</script>

<svelte:head>
	<title>Stock Opname Sessions - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="page-header">
			<h1 class="page-title">Stock Opname Sessions</h1>
			<button class="btn btn-primary" on:click={() => goto('/sessions/new')}>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5"
					viewBox="0 0 20 20"
					fill="currentColor"
				>
					<path
						fill-rule="evenodd"
						d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"
						clip-rule="evenodd"
					/>
				</svg>
				New Session
			</button>
		</div>

		<!-- Filters -->
		<div class="card bg-base-200 shadow-xl mb-6">
			<div class="card-body">
				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					<div class="form-control">
						<label class="label" for="status-filter">
							<span class="label-text">Status</span>
						</label>
						<select
							id="status-filter"
							class="select select-bordered"
							bind:value={filter.status}
							on:change={() => {
								filter.page = 1;
								loadSessions();
							}}
						>
							<option value="">All Status</option>
							<option value="in_progress">In Progress</option>
							<option value="completed">Completed</option>
							<option value="approved">Approved</option>
							<option value="cancelled">Cancelled</option>
						</select>
					</div>

					<div class="form-control">
						<label class="label" for="limit-filter">
							<span class="label-text">Items per page</span>
						</label>
						<select
							id="limit-filter"
							class="select select-bordered"
							bind:value={filter.limit}
							on:change={() => {
								filter.page = 1;
								loadSessions();
							}}
						>
							<option value={10}>10</option>
							<option value={20}>20</option>
							<option value={50}>50</option>
							<option value={100}>100</option>
						</select>
					</div>

					<div class="form-control">
						<span class="label" aria-hidden="true">
							<span class="label-text">&nbsp;</span>
						</span>
						<button class="btn btn-outline" on:click={loadSessions}>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								class="h-5 w-5"
								viewBox="0 0 20 20"
								fill="currentColor"
							>
								<path
									fill-rule="evenodd"
									d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z"
									clip-rule="evenodd"
								/>
							</svg>
							Refresh
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Sessions Table -->
		{#if isLoading}
			<TableSkeleton rows={6} columns={7} />
		{:else if sessions.length === 0}
			<div class="card bg-base-200 shadow-xl">
				<div class="card-body items-center text-center">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-16 w-16 text-base-content/30"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
						/>
					</svg>
					<h2 class="card-title">No sessions found</h2>
					<p>Start a new stock opname session to begin scanning</p>
				</div>
			</div>
		{:else}
			<div class="card bg-base-200 shadow-xl">
				<div class="card-body p-0">
					<div class="overflow-x-auto">
						<table class="table table-zebra">
							<thead>
								<tr>
									<th>Code</th>
									<th>Warehouse</th>
									<th>Status</th>
									<th>Started By</th>
									<th>Started At</th>
									<th>Completed At</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each sessions as session}
									<tr>
										<td>
											<div class="font-mono font-semibold">{session.code}</div>
										</td>
										<td>{session.warehouse?.name || '-'}</td>
										<td>
											<span class="badge {getStatusBadge(session.status)} capitalize">
												{session.status.replace('_', ' ')}
											</span>
										</td>
										<td>{session.started_user?.full_name || '-'}</td>
										<td>{formatDate(session.started_at)}</td>
										<td>{formatDate(session.completed_at)}</td>
										<td>
											<button
												class="btn btn-sm btn-ghost"
												on:click={() => viewDetails(session.id)}
											>
												<svg
													xmlns="http://www.w3.org/2000/svg"
													class="h-5 w-5"
													viewBox="0 0 20 20"
													fill="currentColor"
												>
													<path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
													<path
														fill-rule="evenodd"
														d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z"
														clip-rule="evenodd"
													/>
												</svg>
												View
											</button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="flex justify-center mt-6">
					<div class="join">
						<button
							class="join-item btn"
							disabled={filter.page === 1}
							on:click={() => changePage(filter.page - 1)}
						>
							«
						</button>
						<button class="join-item btn">
							Page {filter.page} of {totalPages}
						</button>
						<button
							class="join-item btn"
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
</LayoutWrapper>
