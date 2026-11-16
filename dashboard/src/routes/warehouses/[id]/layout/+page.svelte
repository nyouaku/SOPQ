<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import { toastStore } from '$lib/stores/toast';

	$: warehouseId = $page.params.id;

	interface LayoutBox {
		id: string;
		code: string;
		name: string;
		x: number;
		y: number;
		width: number;
		height: number;
		color: string;
	}

	let warehouse: any = null;
	let boxes: any[] = [];
	let layoutBoxes: LayoutBox[] = [];
	let isLoading = true;
	let isSaving = false;
	let dragging: LayoutBox | null = null;
	let dragOffset = { x: 0, y: 0 };
	let svgElement: SVGSVGElement;

	const GRID_SIZE = 20;
	const BOX_WIDTH = 80;
	const BOX_HEIGHT = 50;

	const COLORS = [
		'#FF6B35', '#5EEAD4', '#3B82F6', '#10B981', '#F59E0B',
		'#EF4444', '#8B5CF6', '#EC4899', '#14B8A6', '#F97316'
	];

	onMount(async () => {
		try {
			const [whRes, layoutRes] = await Promise.all([
				apiClient.get(`${API_ENDPOINTS.warehouses}/${warehouseId}`),
				apiClient.get(`${API_ENDPOINTS.warehouses}/${warehouseId}/layout`)
			]);

			if (whRes.success && whRes.data) {
				warehouse = whRes.data;
			}

			// Load boxes dari lokasi pertama sebagai default
			if (warehouse) {
				const locRes = await apiClient.get(API_ENDPOINTS.warehouseLocations(warehouseId));
				if (locRes.success && locRes.data?.locations?.[0]) {
					const boxRes = await apiClient.get(API_ENDPOINTS.locationBoxes(locRes.data.locations[0].id));
					if (boxRes.success && boxRes.data) {
						boxes = boxRes.data.boxes || [];
					}
				}
			}

			// Load saved layout
			if (layoutRes.success && layoutRes.data?.layout_json) {
				layoutBoxes = layoutRes.data.layout_json.boxes || [];
			} else {
				// Initialize default layout dari boxes
				layoutBoxes = boxes.map((box, i) => ({
					id: box.id,
					code: box.code,
					name: box.name,
					x: 20 + (i % 5) * (BOX_WIDTH + GRID_SIZE),
					y: 20 + Math.floor(i / 5) * (BOX_HEIGHT + GRID_SIZE),
					width: BOX_WIDTH,
					height: BOX_HEIGHT,
					color: COLORS[i % COLORS.length]
				}));
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal memuat data');
		} finally {
			isLoading = false;
		}
	});

	function handleMouseDown(e: MouseEvent, box: LayoutBox) {
		if (e.button !== 0) return;
		dragging = box;
		dragOffset = {
			x: e.clientX - box.x,
			y: e.clientY - box.y
		};
	}

	function handleMouseMove(e: MouseEvent) {
		if (!dragging) return;
		const rect = svgElement.getBoundingClientRect();
		const x = Math.max(0, Math.min(rect.width - dragging.width, e.clientX - dragOffset.x));
		const y = Math.max(0, Math.min(rect.height - dragging.height, e.clientY - dragOffset.y));

		// Snap to grid
		dragging.x = Math.round(x / GRID_SIZE) * GRID_SIZE;
		dragging.y = Math.round(y / GRID_SIZE) * GRID_SIZE;
	}

	function handleMouseUp() {
		dragging = null;
	}

	function addBox() {
		const newBox: LayoutBox = {
			id: `new-${Date.now()}`,
			code: `NEW-${layoutBoxes.length + 1}`,
			name: 'Rak Baru',
			x: 20,
			y: 20,
			width: BOX_WIDTH,
			height: BOX_HEIGHT,
			color: COLORS[layoutBoxes.length % COLORS.length]
		};
		layoutBoxes = [...layoutBoxes, newBox];
	}

	function removeBox(id: string) {
		layoutBoxes = layoutBoxes.filter((b) => b.id !== id);
	}

	async function saveLayout() {
		isSaving = true;
		try {
			const res = await apiClient.put(`${API_ENDPOINTS.warehouses}/${warehouseId}/layout`, {
				layout_json: {
					version: 1,
					boxes: layoutBoxes
				}
			});
			if (res.success) {
				toastStore.success('Layout tersimpan');
			} else {
				toastStore.error(res.error?.message || 'Gagal menyimpan');
			}
		} catch (e: any) {
			toastStore.error(e.message || 'Gagal menyimpan');
		} finally {
			isSaving = false;
		}
	}
</script>

<svelte:head>
	<title>Layout Editor - {warehouse?.name || 'Gudang'}</title>
</svelte:head>

<svelte:window on:mousemove={handleMouseMove} on:mouseup={handleMouseUp} />

<LayoutWrapper>
	<div class="page">
		<!-- Breadcrumb -->
		<div class="breadcrumbs mb-6 text-sm">
			<ul>
				<li><a href="/">Dashboard</a></li>
				<li><a href="/warehouses">Gudang</a></li>
				<li><a href="/warehouses/{warehouseId}">{warehouse?.name || '...'}</a></li>
				<li>Layout</li>
			</ul>
		</div>

		<div class="mb-6 flex flex-wrap items-center justify-between gap-3">
			<div>
				<h1 class="page-title">Layout Editor</h1>
				<p class="mt-1 text-sm text-base-content/60">
					Drag rak untuk mengatur posisi. Klik "Simpan" untuk menyimpan layout.
				</p>
			</div>
			<div class="flex gap-2">
				<button class="btn btn-outline btn-sm" on:click={addBox}>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" d="M12 4v16m8-8H4" />
					</svg>
					Tambah Rak
				</button>
				<button class="btn btn-primary btn-sm" on:click={saveLayout} disabled={isSaving}>
					{#if isSaving}
						<span class="loading loading-spinner loading-sm"></span>
					{:else}
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
						</svg>
					{/if}
					Simpan Layout
				</button>
			</div>
		</div>

		{#if isLoading}
			<div class="flex h-96 items-center justify-center">
				<Loader size={48} />
			</div>
		{:else}
			<!-- SVG Canvas -->
			<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
				<div class="mb-3 flex items-center justify-between text-sm">
					<span class="text-base-content/60">Grid: {GRID_SIZE}px</span>
					<span class="text-base-content/60">{layoutBoxes.length} rak</span>
				</div>
				<svg
					bind:this={svgElement}
					viewBox="0 0 800 500"
					class="w-full cursor-move rounded-xl bg-base-100"
					style="height: 500px; background-image: linear-gradient(to right, rgba(0,0,0,0.05) 1px, transparent 1px), linear-gradient(to bottom, rgba(0,0,0,0.05) 1px, transparent 1px); background-size: {GRID_SIZE}px {GRID_SIZE}px;"
				>
					{#each layoutBoxes as box (box.id)}
						<g
							role="button"
							tabindex="0"
							on:mousedown={(e) => handleMouseDown(e, box)}
							class="cursor-grab active:cursor-grabbing"
						>
							<rect
								x={box.x}
								y={box.y}
								width={box.width}
								height={box.height}
								rx="4"
								fill={box.color}
								fill-opacity="0.2"
								stroke={box.color}
								stroke-width="2"
							/>
							<text x={box.x + box.width / 2} y={box.y + box.height / 2 - 5} text-anchor="middle" font-size="12" font-weight="bold" fill="currentColor">
								{box.code}
							</text>
							<text x={box.x + box.width / 2} y={box.y + box.height / 2 + 12} text-anchor="middle" font-size="10" fill="currentColor" opacity="0.7">
								{box.name.length > 12 ? box.name.slice(0, 12) + '...' : box.name}
							</text>
							<!-- Delete button -->
							<circle cx={box.x + box.width - 10} cy={box.y + 10} r="8" fill="#EF4444" />
							<text
								x={box.x + box.width - 10}
								y={box.y + 14}
								text-anchor="middle"
								font-size="10"
								fill="white"
								style="cursor: pointer;"
								role="button"
								tabindex="0"
								aria-label="Hapus rak"
								on:click|stopPropagation={() => removeBox(box.id)}
							>
								×
							</text>
						</g>
					{/each}
				</svg>
			</div>
		{/if}
	</div>
</LayoutWrapper>
