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
	// Rak terpilih — bisa lewat mouse, Tab, atau Enter. Panel samping mengedit posisinya.
	let selectedId: string | null = null;
	$: selectedBox = layoutBoxes.find((b) => b.id === selectedId) ?? null;

	const GRID_SIZE = 20;
	const BOX_WIDTH = 80;
	const BOX_HEIGHT = 50;

	const COLORS = [
		'#FF6B35', '#5EEAD4', '#3B82F6', '#10B981', '#F59E0B',
		'#EF4444', '#8B5CF6', '#EC4899', '#14B8A6', '#F97316'
	];

	onMount(async () => {
		try {
			const wid = warehouseId as string;
			const [whRes, layoutRes] = await Promise.all([
				apiClient.get<any>(`${API_ENDPOINTS.warehouses}/${wid}`),
				apiClient.get<{ layout_json: { boxes: any[] } | null }>(
					`${API_ENDPOINTS.warehouses}/${wid}/layout`
				)
			]);

			if (whRes.success && whRes.data) {
				warehouse = whRes.data;
			}

			// Load boxes dari lokasi pertama sebagai default
			if (warehouse) {
				const locRes = await apiClient.get<{ locations: any[] }>(
					API_ENDPOINTS.warehouseLocations(wid)
				);
				if (locRes.success && locRes.data?.locations?.[0]) {
					const boxRes = await apiClient.get<{ boxes: any[] }>(
						API_ENDPOINTS.locationBoxes(locRes.data.locations[0].id)
					);
					if (boxRes.success && boxRes.data) {
						boxes = boxRes.data.boxes || [];
					}
				}
			}

			// Load saved layout
			if (layoutRes.success && layoutRes.data?.layout_json?.boxes) {
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
		selectedId = box.id;
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

	/** Pilih rak (klik atau Enter di SVG). */
	function selectBox(box: LayoutBox) {
		selectedId = box.id;
	}

	/** Geser rak terpilih 1 grid ke arah tsb. Arrow key di panel & di SVG keduanya pakai ini. */
	function nudge(dx: number, dy: number) {
		if (!selectedBox) return;
		const maxX = 800 - selectedBox.width;
		const maxY = 500 - selectedBox.height;
		selectedBox.x = Math.max(0, Math.min(maxX, Math.round((selectedBox.x + dx) / GRID_SIZE) * GRID_SIZE));
		selectedBox.y = Math.max(0, Math.min(maxY, Math.round((selectedBox.y + dy) / GRID_SIZE) * GRID_SIZE));
		// Picu reactivity: layoutBoxes di-assign ulang agar panel ikut ter-render.
		layoutBoxes = [...layoutBoxes];
	}

	/** Set x/y manual dari input angka di panel. */
	function setPos(axis: 'x' | 'y', raw: string) {
		if (!selectedBox) return;
		const maxX = 800 - selectedBox.width;
		const maxY = 500 - selectedBox.height;
		const v = parseInt(raw, 10) || 0;
		const snapped = Math.round(v / GRID_SIZE) * GRID_SIZE;
		selectedBox[axis] = axis === 'x'
			? Math.max(0, Math.min(maxX, snapped))
			: Math.max(0, Math.min(maxY, snapped));
		layoutBoxes = [...layoutBoxes];
	}

	/** Keyboard pada rak di SVG: Enter/Space pilih, panah geser 1 grid, Shift+panah 5 grid. */
	function onBoxKeydown(e: KeyboardEvent, box: LayoutBox) {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			e.stopPropagation();
			selectedId = box.id;
			return;
		}
		const step = e.shiftKey ? GRID_SIZE * 5 : GRID_SIZE;
		const moves: Record<string, [number, number]> = {
			ArrowLeft: [-step, 0],
			ArrowRight: [step, 0],
			ArrowUp: [0, -step],
			ArrowDown: [0, step]
		};
		const mv = moves[e.key];
		if (!mv) return;
		e.preventDefault();
		e.stopPropagation();
		selectedId = box.id;
		nudge(mv[0], mv[1]);
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
					Drag rak, atau pilih lalu atur posisi lewat panel samping.
					Tab untuk memilih, Enter menandai, panah menggeser (Shift = 5 grid).
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
						<!-- role="group": wadah rect + tombol hapus.
						     svelte-ignore di bawah: handler panah di sini memang disengaja —
						     itu jalur keyboard untuk menggeser rak (rect di dalam role=button
						     tak bisa punya arrow key handler sendiri tanpa memakai tabindex ganda). -->
						<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
						<g
							role="group"
							aria-label={`Rak ${box.code}. Panah untuk menggeser, Enter untuk memilih.`}
							on:mousedown={(e) => handleMouseDown(e, box)}
							on:keydown={(e) => onBoxKeydown(e, box)}
							class="cursor-grab active:cursor-grabbing"
						>
							<rect
								x={box.x}
								y={box.y}
								width={box.width}
								height={box.height}
								rx="4"
								fill={box.color}
								fill-opacity={selectedId === box.id ? 0.4 : 0.2}
								stroke={selectedId === box.id ? 'hsl(var(--p))' : box.color}
								stroke-width={selectedId === box.id ? 3 : 2}
							/>
							<text x={box.x + box.width / 2} y={box.y + box.height / 2 - 5} text-anchor="middle" font-size="12" font-weight="bold" fill="currentColor">
								{box.code}
							</text>
							<text x={box.x + box.width / 2} y={box.y + box.height / 2 + 12} text-anchor="middle" font-size="10" fill="currentColor" opacity="0.7">
								{box.name.length > 12 ? box.name.slice(0, 12) + '...' : box.name}
							</text>
							<!-- Delete button -->
							<circle cx={box.x + box.width - 10} cy={box.y + 10} r="8" fill="#EF4444" />
							<!-- Tombol hapus: <g> ber-role button dengan gelar rakit sendiri,
							     supaya <text> di dalamnya tak perlu role terpisah. -->
							<g
								role="button"
								tabindex="0"
								aria-label={`Hapus rak ${box.code}`}
								class="cursor-pointer"
								on:click|stopPropagation={() => removeBox(box.id)}
								on:keydown|stopPropagation={(e) => {
									if (e.key === 'Enter' || e.key === ' ') {
										e.preventDefault();
										removeBox(box.id);
									}
								}}
							>
								<text
									x={box.x + box.width - 10}
									y={box.y + 14}
									text-anchor="middle"
									font-size="10"
									fill="white"
								>
									×
								</text>
							</g>
						</g>
					{/each}
				</svg>
			</div>

			<!-- Panel samping: posisi rak terpilih, editable lewat angka & tombol panah.
			     Jalur keyboard utama — drag mouse hanyalah pintasan. -->
			<div class="mt-4 rounded-2xl border border-base-300 bg-base-200 p-4">
				{#if !selectedBox}
					<p class="text-sm text-base-content/60">
						Pilih rak pada canvas (klik, atau Tab + Enter) untuk mengatur posisinya.
					</p>
				{:else}
					<div class="mb-3 flex items-center gap-2">
						<span
							class="inline-block h-3 w-3 rounded-full"
							style="background:{selectedBox.color}"
						></span>
						<span class="font-semibold">{selectedBox.code}</span>
						<span class="text-sm text-base-content/60">{selectedBox.name}</span>
					</div>

					<div class="flex flex-wrap items-end gap-4">
						<label class="form-control">
							<span class="label-text text-xs text-base-content/60">X (px)</span>
							<input
								type="number" step={GRID_SIZE} min="0"
								class="input input-bordered input-sm w-24"
								value={selectedBox.x}
								on:input={(e) => setPos('x', e.currentTarget.value)}
							/>
						</label>

						<label class="form-control">
							<span class="label-text text-xs text-base-content/60">Y (px)</span>
							<input
								type="number" step={GRID_SIZE} min="0"
								class="input input-bordered input-sm w-24"
								value={selectedBox.y}
								on:input={(e) => setPos('y', e.currentTarget.value)}
							/>
						</label>

						<!-- Tombol panah: ±1 grid, Shift = ±5 grid -->
						<div class="flex items-center gap-1" role="group" aria-label="Geser rak terpilih">
							<button class="btn btn-sm" aria-label="Geser kiri 1 grid"
								on:click={() => nudge(-GRID_SIZE, 0)}>←</button>
							<button class="btn btn-sm" aria-label="Geser kanan 1 grid"
								on:click={() => nudge(GRID_SIZE, 0)}>→</button>
							<button class="btn btn-sm" aria-label="Geser atas 1 grid"
								on:click={() => nudge(0, -GRID_SIZE)}>↑</button>
							<button class="btn btn-sm" aria-label="Geser bawah 1 grid"
								on:click={() => nudge(0, GRID_SIZE)}>↓</button>
							<span class="ml-2 text-xs text-base-content/60">Shift + panah = 5 grid</span>
						</div>

						<button class="btn btn-sm text-error" on:click={() => removeBox(selectedBox.id)}>
							Hapus rak
						</button>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</LayoutWrapper>
