<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { QueryClientProvider } from '@tanstack/svelte-query';
	import { getQueryClient } from '$lib/api/queryClient';
	import LoadingBar from '$lib/components/LoadingBar.svelte';
	import Preloader from '$lib/components/Preloader.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import ErrorBoundary from '$lib/components/ErrorBoundary.svelte';
	import PWAInstallPrompt from '$lib/components/PWAInstallPrompt.svelte';

	// Satu instance dipakai semua halaman (lihat lib/api/queryClient.ts).
	const queryClient = getQueryClient();

	// Guard autentikasi terpusat: satu tempat, bukan per halaman.
	// Halaman publik dikecualikan lewat PUBLIC_ROUTES.
	const PUBLIC_ROUTES = ['/login'];
	$: isPublicRoute = PUBLIC_ROUTES.includes($page.url.pathname);
	$: if (!$authStore.isLoading && !$authStore.isAuthenticated && !isPublicRoute) {
		goto('/login');
	}

	// Guard role: Inventory Management tak untuk staff gudang (UI-only; backend
	// tetap proteksi via RequireRole). Staff yang memaksa URL dialihkan ke /apps.
	// Inventory = /inventory*, /products*, /warehouses*, /transfer*.
	const INVENTORY_ROOTS = ['/inventory', '/products', '/warehouses', '/transfer'];
	$: isInventory = INVENTORY_ROOTS.some(
		(r) => $page.url.pathname === r || $page.url.pathname.startsWith(r + '/')
	);
	$: if (
		!$authStore.isLoading &&
		$authStore.isAuthenticated &&
		isInventory &&
		$authStore.user?.role?.name === 'staff gudang'
	) {
		goto('/apps');
	}

	onMount(() => {
		authStore.checkAuth();
		themeStore.init();
	});
</script>

<ErrorBoundary />
<LoadingBar />
<Preloader />
<Toast />
<QueryClientProvider client={queryClient}>
	<slot />
</QueryClientProvider>
<PWAInstallPrompt />
