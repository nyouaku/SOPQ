<script lang="ts">
	import { derived, writable } from 'svelte/store';
	import { get } from 'svelte/store';
	import { goto } from '$app/navigation';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { warehousesQuery, locationsQuery } from '$lib/api/queries';
	import { toastStore } from '$lib/stores/toast';

	const queryClient = useQueryClient();

	const warehouses = warehousesQuery();
	const warehouseId = writable('');
	let selectedWarehouse = '';
	let selectedLocation = '';
	let submitting = false;
	let errorMessage = '';

	// Query lokasi mengikuti store — berubah saat gudang diganti, tanpa remount.
	const locations = locationsQuery(derived(warehouseId, ($id) => $id));

	type CreatePayload = { warehouse_id: string; location_id?: string };

	const createSession = createMutation<{ id: string }, Error, CreatePayload>({
		mutationFn: async (payload) => {
			const res = await apiClient.post<{ id: string }>(
				API_ENDPOINTS.stockOpnameSessions,
				payload
			);
			if (!res.success || !res.data) {
				throw new Error(res.error?.message || 'Gagal membuat sesi');
			}
			return res.data;
		},
		onSuccess: (data) => {
			queryClient.invalidateQueries({ queryKey: ['sessions'] });
			toastStore.success('Sesi stock opname dibuat');
			goto(`/sessions/${data.id}`);
		},
		onError: (err: Error) => {
			errorMessage = err.message;
			toastStore.error(err.message);
		},
		onSettled: () => {
			submitting = false;
		}
	});

	function submit() {
		errorMessage = '';
		if (!selectedWarehouse) {
			errorMessage = 'Gudang wajib dipilih';
			return;
		}
		submitting = true;
		$createSession.mutate({
			warehouse_id: selectedWarehouse,
			...(selectedLocation ? { location_id: selectedLocation } : {})
		});
	}

	$: warehouseList = get(warehouses).data ?? [];
</script>

<svelte:head>
	<title>New Session - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<div class="page">
		<div class="flex items-center gap-3 mb-6">
			<button class="btn btn-ghost btn-sm" on:click={() => goto('/sessions')}>←</button>
			<h1 class="page-title">New Stock Opname Session</h1>
		</div>

		<div class="card bg-base-200 shadow-xl">
			<div class="card-body gap-4">
				{#if $warehouses.isLoading}
					<span class="loading loading-spinner loading-md"></span>
				{:else if $warehouses.isError}
					<div class="alert alert-error">
						<span>Gagal memuat gudang: {$warehouses.error?.message}</span>
					</div>
				{:else}
					<div class="form-control">
						<label class="label" for="warehouse">
							<span class="label-text">Gudang *</span>
						</label>
						<select
							id="warehouse"
							class="select select-bordered"
							bind:value={selectedWarehouse}
							on:change={() => {
								selectedLocation = '';
								warehouseId.set(selectedWarehouse);
							}}
						>
							<option value="">-- Pilih gudang --</option>
							{#each warehouseList as wh (wh.id)}
								<option value={wh.id}>{wh.name}</option>
							{/each}
						</select>
					</div>

					{#if selectedWarehouse}
						<div class="form-control">
							<label class="label" for="location">
								<span class="label-text">Lokasi (opsional)</span>
							</label>
							{#if $locations.isLoading}
								<span class="loading loading-spinner loading-sm"></span>
							{:else}
								<select
									id="location"
									class="select select-bordered"
									bind:value={selectedLocation}
								>
									<option value="">-- Semua lokasi --</option>
									{#each $locations.data ?? [] as loc (loc.id)}
										<option value={loc.id}>{loc.name}</option>
									{/each}
								</select>
							{/if}
						</div>
					{/if}

					{#if errorMessage}
						<div class="alert alert-error">
							<span>{errorMessage}</span>
						</div>
					{/if}

					<div class="card-actions justify-end mt-2">
						<button class="btn btn-ghost" on:click={() => goto('/sessions')}>Batal</button>
						<button
							class="btn btn-primary"
							disabled={submitting || !selectedWarehouse}
							on:click={submit}
						>
							{#if submitting}
								<span class="loading loading-spinner loading-sm"></span>
							{/if}
							Mulai Sesi
						</button>
					</div>
				{/if}
			</div>
		</div>
	</div>
</LayoutWrapper>
