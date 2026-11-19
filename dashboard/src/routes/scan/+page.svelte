<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { authStore } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { apiClient } from '$lib/api/client';
	import { API_ENDPOINTS } from '$lib/api/config';
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import { enqueue, flushQueue, hasQueue, listQueue } from '$lib/offline/queue';
	import { productStock } from '$lib/utils/stock';
	import { MultiFormatReader, RGBLuminanceSource, HybridBinarizer, BinaryBitmap, NotFoundException } from '@zxing/library';
	import { soundSettingsStore } from '$lib/stores/soundSettings';
	import { playSound } from '$lib/stores/soundSettings';
	import SaveNotificationToast from '$lib/components/SaveNotificationToast.svelte';

	/** Decode piksel canvas dengan ZXing core (0.21.3 tidak punya decodeFromImageData). Kosong jika tidak ada kode. */
	function zxingDecode(ctx: CanvasRenderingContext2D, w: number, h: number): string {
		const { data } = ctx.getImageData(0, 0, w, h);
		const lum = new Uint8ClampedArray(w * h);
		for (let i = 0, p = 0; i < lum.length; i++, p += 4) {
			lum[i] = (data[p] * 306 + data[p + 1] * 601 + data[p + 2] * 117) >> 10; // luminance BT.601
		}
		try {
			const bitmap = new BinaryBitmap(new HybridBinarizer(new RGBLuminanceSource(lum, w, h)));
			return new MultiFormatReader().decode(bitmap).getText();
		} catch (e) {
			if (e instanceof NotFoundException) return '';
			throw e;
		}
	}

	type ScanMode = 'camera' | 'upload' | 'manual';

	let scanMode: ScanMode = 'camera';
	let isCameraActive = false;
	let scannedCode = '';
	let product: any = null;
	let quantity = '';
	let isLoading = false;
	let isSaving = false;
	let message = '';
	let messageType: 'success' | 'error' | 'info' = 'info';
	let sessions: any[] = [];
	let selectedSession = '';
	let pendingScans = 0;
	let scanSuccessPlayed = false; // prevent multiple sounds per scan
	let showSaveNotification = false;
	let selectedSessionData: any = null;

	// Ringkasan scan untuk sesi aktif (total_scanned / total_products) — ditampilkan
	// sebagai panel "hasil scan" di bawah tombol simpan.
	let sessionSummary: { total_scanned: number; total_products: number } | null = null;

	// Update selectedSessionData saat session berubah
	$: selectedSessionData = sessions.find((s: any) => s.id === selectedSession);

	let videoElement: HTMLVideoElement;
	let canvasElement: HTMLCanvasElement;
	let stream: MediaStream | null = null;
	let fileInput: HTMLInputElement;

	// Barcode detector — native (BarcodeDetector API) atau @zxing/library.
	let barcodeDetector: any = null;
	let scanInterval: any = null;
		let scanBusy = false;

	onMount(async () => {
		// Check auth first
		if (!$authStore.isAuthenticated && !$authStore.isLoading) {
			goto('/login');
			return;
		}

		await loadSessions();
		await loadSessionSummary();
		initBarcodeDetector();
		pendingScans = listQueue().length;
		// Kirim ulang scan yang tertunda begitu jaringan kembali.
		window.addEventListener('online', syncPending);
	});

	onDestroy(() => {
		window.removeEventListener('online', syncPending);
		stopCamera();
		if (scanInterval) clearInterval(scanInterval);
	});

	/** Kirim ulang antrian offline. */
	async function syncPending() {
		if (!hasQueue()) return;
		const { sent, failed } = await flushQueue();
		pendingScans = listQueue().length;
		if (sent) showMessage(`${sent} scan terkirim`, 'success');
		else if (failed) showMessage('Masih offline, scan menunggu', 'info');
	}

	// loadZXing tidak lagi dipakai — @zxing/library di-import langsung.

	function initBarcodeDetector() {
		if ('BarcodeDetector' in window) {
			try {
				barcodeDetector = new (window as any).BarcodeDetector({
					formats: ['qr_code', 'ean_13', 'ean_8', 'code_128', 'code_39']
				});
			} catch (error) {
				console.warn(' BarcodeDetector failed:', error);
			}
		} else {
			console.warn(' BarcodeDetector API not available');
		}
	}

	async function loadSessions() {
		try {
			const response = await apiClient.get<{ sessions: any[] }>(
				`${API_ENDPOINTS.stockOpnameSessions}?status=in_progress`
			);
			if (response.success && response.data) {
				sessions = response.data.sessions || [];
				if (sessions.length > 0 && !selectedSession) {
					selectedSession = sessions[0].id;
				}
			}
		} catch (error) {
			console.error('Failed to load sessions:', error);
		}
	}

	// Ganti sesi -> muat ulang ringkasan scan sesi tsb.
	async function onSessionChange() {
		sessionSummary = null;
		await loadSessionSummary();
	}

	function switchMode(mode: ScanMode) {
		stopCamera();
		scanMode = mode;
		resetScanData();
	}

	async function startCamera() {
		try {
			if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
				showMessage('Browser tidak mendukung akses kamera. Gunakan Chrome/Edge.', 'error');
				return;
			}

			isCameraActive = true;
			showMessage('Mengaktifkan kamera...', 'info');

			stream = await navigator.mediaDevices.getUserMedia({
				video: {
					facingMode: 'environment',
					width: { ideal: 1280 },
					height: { ideal: 720 }
				},
				audio: false
			});

			if (videoElement) {
				videoElement.srcObject = stream;
				await videoElement.play();
				showMessage('Kamera aktif! Arahkan ke QR code...', 'success');

				// Start continuous scanning
				scanInterval = setInterval(() => {
					scanFromCamera();
				}, 500); // Scan every 500ms
			}
		} catch (error: any) {
			console.error('Camera error:', error);
			isCameraActive = false;

			if (error.name === 'NotAllowedError' || error.name === 'PermissionDeniedError') {
				showMessage('Izin kamera ditolak. Klik ikon kamera di address bar untuk mengizinkan.', 'error');
			} else if (error.name === 'NotFoundError' || error.name === 'DevicesNotFoundError') {
				showMessage('Kamera tidak ditemukan. Pastikan webcam terhubung.', 'error');
			} else if (error.name === 'NotReadableError' || error.name === 'TrackStartError') {
				showMessage('Kamera sedang digunakan aplikasi lain.', 'error');
			} else {
				showMessage(`Tidak dapat mengakses kamera: ${error.message}`, 'error');
			}
		}
	}

	function stopCamera() {
		if (stream) {
			stream.getTracks().forEach((track) => track.stop());
			stream = null;
		}
		if (scanInterval) {
			clearInterval(scanInterval);
			scanInterval = null;
		}
		isCameraActive = false;
	}

		async function scanFromCamera() {
			if (!isCameraActive || !videoElement || scanBusy) return;
			if (videoElement.readyState !== videoElement.HAVE_ENOUGH_DATA) return;

			scanBusy = true;
			try {
				let code = '';
				// Native detector (Chromium) dulu; ZXing sebagai fallback supaya QR jalan di semua browser.
				if (barcodeDetector) {
					try {
						const barcodes = await barcodeDetector.detect(videoElement);
						if (barcodes.length > 0) code = barcodes[0].rawValue;
					} catch {
						// lanjut ke ZXing
					}
				}
				if (!code) code = decodeVideoFrame();
				if (!code) return;

				scannedCode = code;
				stopCamera();
				showMessage('Kode terdeteksi!', 'success');
				playSound('success');
				scanSuccessPlayed = true;
				await handleCodeInput();
			} finally {
				scanBusy = false;
			}
		}

		/** Decode satu frame video dengan ZXing (QR + barcode 1D). Kosong jika tidak ada kode. */
		function decodeVideoFrame(): string {
			const w = videoElement.videoWidth;
			const h = videoElement.videoHeight;
			const ctx = canvasElement.getContext('2d');
			if (!w || !h || !ctx) return '';
			canvasElement.width = w;
			canvasElement.height = h;
			ctx.drawImage(videoElement, 0, 0, w, h);
			try {
				return zxingDecode(ctx, w, h);
			} catch {
				return ''; // frame tanpa kode / gagal decode: coba frame berikutnya
			}
		}

	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;

		if (!file.type.startsWith('image/')) {
			showMessage('File harus berupa gambar (JPG, PNG, dll)', 'error');
			return;
		}

		decodeImageFile(file);
	}

	async function decodeImageFile(file: File) {
		isLoading = true;
		showMessage('Membaca QR code dari gambar...', 'info');

		try {
			const img = new Image();
			const reader = new FileReader();

			reader.onload = async (e) => {
				img.src = e.target?.result as string;

				img.onload = async () => {
					const canvas = canvasElement;
					const ctx = canvas.getContext('2d');
					if (!ctx) {
						showMessage('Canvas error', 'error');
						isLoading = false;
						return;
					}

					// Downscale: foto HP besar bikin decode lambat/gagal.
					const MAX_DIM = 1000;
					const scale = Math.min(1, MAX_DIM / Math.max(img.width, img.height));
					canvas.width = Math.round(img.width * scale);
					canvas.height = Math.round(img.height * scale);
					ctx.drawImage(img, 0, 0, canvas.width, canvas.height);

					// Try native detector first
					if (barcodeDetector) {
						try {
							const barcodes = await barcodeDetector.detect(canvas);
							if (barcodes.length > 0) {
								scannedCode = barcodes[0].rawValue;
								showMessage('Kode terdeteksi!', 'success');
								playSound('success');
								scanSuccessPlayed = true;
								await handleCodeInput();
								isLoading = false;
								return;
							}
						} catch (error) {
							console.warn('Native detector failed, trying ZXing...');
						}
					}

					// Fallback: ZXing core (sync).
					try {
						const text = zxingDecode(ctx, canvas.width, canvas.height);
							if (text) {
								scannedCode = text;
								showMessage('Kode terdeteksi dengan ZXing!', 'success');
								playSound('success');
								scanSuccessPlayed = true;
								await handleCodeInput();
								isLoading = false;
								return;
							}
					} catch (error) {
						console.warn('ZXing failed:', error);
					}

					showMessage('Tidak dapat membaca QR code. Pastikan gambar jelas dan QR code terlihat.', 'error');
					isLoading = false;
				};

				img.onerror = () => {
					showMessage('Gagal load gambar', 'error');
					isLoading = false;
				};
			};

			reader.readAsDataURL(file);
		} catch (error) {
			console.error('Image decode error:', error);
			showMessage('Gagal memproses gambar', 'error');
			isLoading = false;
		}
	}

	async function handleCodeInput() {
		if (!scannedCode.trim()) {
			showMessage('Kode produk kosong', 'error');
			return;
		}

		isLoading = true;
		message = '';

		try {
			const response = await apiClient.get(`${API_ENDPOINTS.productByCode(scannedCode)}`);
			if (response.success && response.data) {
				product = response.data;
				showMessage(` Produk: ${product.name}`, 'success');
			} else {
				showMessage('Produk tidak ditemukan', 'error');
				product = null;
			}
		} catch (error: any) {
			showMessage(error.message || 'Gagal mencari produk', 'error');
			product = null;
		} finally {
			isLoading = false;
		}
	}

	async function handleSubmit() {
		if (!product || !quantity || !selectedSession) {
			showMessage('Lengkapi semua data', 'error');
			return;
		}

		isSaving = true;

		try {
			const payload = {
				session_id: selectedSession,
				product_code: scannedCode,
				stock_physical: parseInt(quantity),
				input_method: scanMode as 'manual' | 'camera' | 'upload',
				note: ''
			};

			const response = await apiClient.post(
				`${API_ENDPOINTS.stockOpnameSession(selectedSession)}/scan`,
				payload
			);

			if (response.success) {
				showMessage('Data berhasil disimpan', 'success');
				playSound('success');
				showSaveToast();
				// Ambil ringkasan terbaru untuk panel hasil (dan deteksi "scan selesai").
				await loadSessionSummary();
				resetForm();
			} else {
				showMessage(response.error?.message || 'Gagal menyimpan data', 'error');
			}
		} catch {
			// Jaringan mati — masukkan antrian.
			enqueue({
				session_id: selectedSession,
				product_code: scannedCode,
				stock_physical: parseInt(quantity),
				input_method: scanMode as 'manual' | 'camera' | 'upload',
				note: ''
			});
			showMessage('Offline: scan masuk antrian', 'info');
			resetForm();
		} finally {
			isSaving = false;
			pendingScans = listQueue().length;
		}
	}

	/** Ambil ringkasan sesi aktif: total ter-scan vs total produk. */
	async function loadSessionSummary() {
		if (!selectedSession) {
			sessionSummary = null;
			return;
		}
		try {
			const res = await apiClient.get<{ total_scanned: number; total_products: number }>(
				`${API_ENDPOINTS.stockOpnameSession(selectedSession)}/summary`
			);
			if (res.success && res.data) {
				sessionSummary = {
					total_scanned: res.data.total_scanned ?? 0,
					total_products: res.data.total_products ?? 0
				};
			}
		} catch {
			// Summary gagal dimuat bukan fatal — panel hasil saja yang tak muncul.
			sessionSummary = null;
		}
	}

	function resetScanData() {
		scannedCode = '';
		product = null;
		quantity = '';
		if (fileInput) fileInput.value = '';
	}

	function resetForm() {
		const wasCameraMode = scanMode === 'camera';
		resetScanData();
		// Auto-restart kamera agar siap scan berikutnya tanpa klik apa pun.
		if (wasCameraMode) startCamera();
	}

	function showMessage(msg: string, type: 'success' | 'error' | 'info') {
		message = msg;
		messageType = type;
		setTimeout(() => {
			message = '';
		}, 5000);
	}

	function showSaveToast() {
		showSaveNotification = false;
		setTimeout(() => (showSaveNotification = true), 0);
		setTimeout(() => (showSaveNotification = false), 2500);
	}
</script>

<svelte:head>
	<title>Scan - Stock Opname</title>
</svelte:head>

{#if showSaveNotification}
	<SaveNotificationToast message="Scan berhasil disimpan" scannedCount={1} totalProducts={0} />
{/if}

<LayoutWrapper>
	<div class="page">
		<div class="page-header">
			<div>
				<h1 class="page-title">Scan</h1>
				<p class="page-subtitle">Cari produk, lalu isi jumlah fisik</p>
			</div>
		</div>

		<!-- Session Selection — disembunyikan saat produk sudah ter-scan agar
		     tombol simpan tak terdorong ke bawah & tertutup bottom navbar. -->
		{#if !product}
		<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-4">
			<label class="label py-0" for="session-select">
				<span class="label-text text-xs font-medium text-base-content/60">Sesi stock opname</span>
			</label>
			<select
				id="session-select"
				class="select select-bordered mt-1 w-full"
				bind:value={selectedSession}
				on:change={onSessionChange}
			>
				{#if sessions.length === 0}
					<option value="">Tidak ada sesi aktif</option>
				{:else}
					{#each sessions as session}
						<option value={session.id}
							>{session.code} - {session.warehouse_name || 'Warehouse'}</option
						>
					{/each}
				{/if}
			</select>

			{#if pendingScans > 0}
				<div class="mt-3 flex items-center justify-between gap-2 rounded-xl border border-warning/40 bg-warning/10 px-3 py-2">
					<span class="text-xs text-warning">
						{pendingScans} scan menunggu dikirim (offline)
					</span>
					<button class="btn btn-ghost btn-xs" on:click={syncPending}>Kirim</button>
				</div>
			{/if}

			<!-- Blind Count Indicator -->
			{#if selectedSessionData?.is_blind_count}
				<div class="mt-2 flex items-center gap-2 rounded-lg border border-info/40 bg-info/10 px-3 py-1.5">
					<svg class="h-4 w-4 text-info" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
						<path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
					</svg>
					<span class="text-xs font-medium text-info">Blind Count Mode — stok sistem disembunyikan</span>
				</div>
			{/if}
		</div>
		{/if}
		<!-- akhir blok sesi (disembunyikan saat produk ter-scan) -->

		<!-- Scan Mode: segmented control -->
		<div
			class="mb-4 grid grid-cols-3 gap-1 rounded-xl bg-base-200 p-1"
			role="tablist"
			aria-label="Mode scan"
		>
			<button
				role="tab"
				aria-selected={scanMode === 'manual'}
				class="flex items-center justify-center gap-1.5 rounded-lg px-2 py-2 text-sm font-medium transition-colors"
				class:bg-base-100={scanMode === 'manual'}
				class:shadow-sm={scanMode === 'manual'}
				class:text-primary={scanMode === 'manual'}
				class:opacity-50={scanMode !== 'manual'}
				on:click={() => switchMode('manual')}
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<rect x="2" y="6" width="20" height="12" rx="2" />
					<path d="M6 10h1M10 10h1M14 10h1M18 10h1M6 14h12" stroke-linecap="round" />
				</svg>
				Manual
			</button>
			<button
				role="tab"
				aria-selected={scanMode === 'camera'}
				class="flex items-center justify-center gap-1.5 rounded-lg px-2 py-2 text-sm font-medium transition-colors"
				class:bg-base-100={scanMode === 'camera'}
				class:shadow-sm={scanMode === 'camera'}
				class:text-primary={scanMode === 'camera'}
				class:opacity-50={scanMode !== 'camera'}
				on:click={() => switchMode('camera')}
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
					<circle cx="12" cy="13" r="3" />
				</svg>
				Kamera
			</button>
			<button
				role="tab"
				aria-selected={scanMode === 'upload'}
				class="flex items-center justify-center gap-1.5 rounded-lg px-2 py-2 text-sm font-medium transition-colors"
				class:bg-base-100={scanMode === 'upload'}
				class:shadow-sm={scanMode === 'upload'}
				class:text-primary={scanMode === 'upload'}
				class:opacity-50={scanMode !== 'upload'}
				on:click={() => switchMode('upload')}
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
				</svg>
				Upload
			</button>
		</div>

		<!-- Scanner Area (hidden when product found, to keep form in front) -->
		{#if !product}
			<div class="mb-4 rounded-2xl border border-base-300 bg-base-200 p-4">
				{#if scanMode === 'manual'}
					<!-- Manual Input -->
					<div class="space-y-3">
						<label class="label py-0" for="code-input">
							<span class="label-text text-xs font-medium text-base-content/60">Ketik / paste kode produk</span>
						</label>
						<div class="join w-full">
							<input
								id="code-input"
								type="text"
								inputmode="text"
								autocomplete="off"
								placeholder="Contoh: 8991234567001"
								class="input input-bordered join-item min-w-0 flex-1"
								bind:value={scannedCode}
								disabled={isLoading}
								on:keydown={(e) => e.key === 'Enter' && handleCodeInput()}
							/>
							<button
								class="btn btn-primary join-item shrink-0"
								on:click={handleCodeInput}
								disabled={isLoading || !scannedCode.trim()}
							>
								{#if isLoading}
									<Loader size={20} />
								{:else}
									Cari
								{/if}
							</button>
						</div>
						<p class="text-xs text-base-content/50">
							Bisa pakai scanner fisik — data langsung terbaca saat Enter.
						</p>
					</div>

				{:else if scanMode === 'camera'}
					<!-- Camera Scanner -->
					{#if isCameraActive}
						<div class="relative mx-auto w-full max-w-sm overflow-hidden rounded-xl bg-black">
							<video
								bind:this={videoElement}
								class="aspect-[3/4] h-auto w-full object-cover"
								autoplay
								playsinline
								muted
							></video>

							<!-- Viewfinder overlay -->
							<div class="pointer-events-none absolute inset-0 flex items-center justify-center">
								<div class="relative h-[55%] w-[82%] max-w-sm rounded-2xl border-[3px] border-primary/85">
									<div class="animate-scan absolute left-0 top-0 h-0.5 w-full bg-primary"></div>
								</div>
							</div>

							<button
								class="btn btn-error btn-circle absolute bottom-4 left-1/2 -translate-x-1/2 shadow-lg"
								aria-label="Hentikan kamera"
								on:click={stopCamera}
							>
								<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" d="M6 6l12 12M18 6L6 18" />
								</svg>
							</button>
						</div>
					{:else}
						<div class="flex flex-col items-center justify-center gap-3 py-10 text-center">
							<div class="rounded-full bg-base-300 p-4">
								<svg
									class="h-8 w-8 text-base-content/60"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									stroke-width="1.6"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
									/>
									<circle cx="12" cy="13" r="3" />
								</svg>
							</div>
							<button
								class="btn btn-primary"
								on:click={startCamera}
								disabled={!selectedSession || isLoading}
							>
								Aktifkan kamera
							</button>
							<p class="text-xs text-base-content/50">Scan otomatis tiap 0.5 detik</p>
						</div>
					{/if}

				{:else if scanMode === 'upload'}
					<!-- Upload gambar -->
					<div class="flex flex-col items-center justify-center gap-3 py-10 text-center">
						<div class="rounded-full bg-base-300 p-4">
							<svg
								class="h-8 w-8 text-base-content/60"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								stroke-width="1.6"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
								/>
							</svg>
						</div>
						<input
							type="file"
							accept="image/*"
							class="file-input file-input-bordered file-input-primary w-full max-w-xs"
							on:change={handleFileSelect}
							bind:this={fileInput}
							disabled={!selectedSession || isLoading}
						/>
						<p class="text-xs text-base-content/50">JPG, PNG — foto QR code</p>
					</div>
				{/if}

				<canvas bind:this={canvasElement} class="hidden"></canvas>
		</div>
		{/if}

		<!-- Message Alert -->
		{#if message}
			<div
				role="status"
				class="alert mb-4 py-2 text-sm alert-{messageType === 'success'
					? 'success'
					: messageType === 'error'
						? 'error'
						: 'info'}"
			>
				<span>{message}</span>
			</div>
		{/if}

		<!-- Product Info & Quantity -->
		{#if product}
			<div class="mb-4 rounded-2xl border-2 border-primary bg-base-200 p-4">
				<div class="flex items-start justify-between gap-3">
					<div class="min-w-0">
						<div class="text-xs font-medium text-primary">Produk ditemukan</div>
						<h2 class="mt-1 truncate font-semibold">{product.name}</h2>
					</div>
					<button
						class="btn btn-ghost btn-xs shrink-0"
						aria-label="Tutup produk"
						on:click={resetScanData}
					>
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" d="M6 6l12 12M18 6L6 18" />
						</svg>
					</button>
				</div>

				{#if !selectedSessionData?.is_blind_count}
					<dl class="mt-3 grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
						<div>
							<dt class="text-xs text-base-content/60">Kode</dt>
							<dd class="metric">{product.product_code || '-'}</dd>
						</div>
						{#if product.brand}
							<div>
								<dt class="text-xs text-base-content/60">Brand</dt>
								<dd class="truncate">{product.brand}</dd>
							</div>
						{/if}
						<div>
							<dt class="text-xs text-base-content/60">Stok sistem</dt>
							<dd class="metric">{productStock(product)}</dd>
						</div>
					</dl>
				{:else}
					<p class="text-xs text-base-content/50">
						Mode blind count: masukkan jumlah fisik tanpa melihat stok sistem.
					</p>
				{/if}
			</div>

			<!-- Quantity Input -->
			<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
				<label class="label py-0" for="qty-input">
					<span class="label-text text-xs font-medium text-base-content/60">Jumlah fisik</span>
				</label>
				<input
					id="qty-input"
					type="number"
					inputmode="numeric"
					min="0"
					placeholder="0"
					class="input input-bordered metric mt-1 h-14 w-full text-center text-3xl"
					bind:value={quantity}
				/>

				<button
					class="btn btn-primary mt-3 w-full"
					on:click={handleSubmit}
					disabled={!quantity || isSaving || !selectedSession}
				>
					{#if isSaving}
						<Loader size={20} />
						Menyimpan...
					{:else}
						Simpan data
					{/if}
				</button>

				<button class="btn btn-ghost mt-2 w-full" on:click={resetForm}>
					Reset &amp; scan lagi
				</button>
			</div>

			<!-- Hasil scan sesi: progress ter-scan + indikator selesai -->
			{#if sessionSummary && sessionSummary.total_products > 0}
				{@const scanned = sessionSummary.total_scanned}
				{@const total = sessionSummary.total_products}
				{@const done = scanned >= total}
				<div
					class="mt-4 rounded-2xl border p-4 {done
						? 'border-success/50 bg-success/10'
						: 'border-base-300 bg-base-200'}"
					role="status"
				>
					<div class="flex items-center justify-between gap-3">
						<div>
							<div class="text-xs text-base-content/60">Progress sesi ini</div>
							<div class="metric text-lg font-semibold">
								{scanned} / {total}
								{#if done}
									<span class="ml-1 text-sm font-normal text-success">— selesai ✓</span>
								{/if}
							</div>
						</div>
						{#if done}
							<span class="badge badge-success badge-lg">Semua produk ter-scan</span>
						{/if}
					</div>
					{#if !done}
						<progress
							class="progress progress-primary mt-2"
							value={scanned}
							max={total}
						></progress>
					{:else}
						<a
							href="/sessions/{selectedSession}"
							class="btn btn-outline btn-sm mt-2 w-full"
						>
							Lihat ringkasan sesi
						</a>
					{/if}
				</div>
			{/if}
		{/if}
	</div>
</LayoutWrapper>

<style>
	/* Scanner animation */
	@keyframes scan {
		0% {
			top: 0;
		}
		50% {
			top: 100%;
		}
		100% {
			top: 0;
		}
	}

	.animate-scan {
		animation: scan 2s ease-in-out infinite;
		box-shadow: 0 0 10px rgba(94, 234, 212, 0.8);
	}
</style>

