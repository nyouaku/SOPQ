<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { themeStore } from '$lib/stores/theme';
	import type { ThemeLayout } from '$lib/stores/theme';

	import SidebarLayout from './SidebarLayout.svelte';
	import TopNavLayout from './TopNavLayout.svelte';
	import CommandLayout from './CommandLayout.svelte';
	import BottomLayout from './BottomLayout.svelte';
	import AppleLayout from './AppleLayout.svelte';

	let currentLayout: ThemeLayout = 'apple';
	let currentPath = '';

	onMount(() => {
		themeStore.subscribe((theme) => {
			currentLayout = theme.layout;
		});
	});

	$: currentPath = $page.url.pathname;

	// Grid Harian (/inventory/entry) butuh ruang penuh ala spreadsheet Excel:
	// sidebar & bottom bar disembunyikan, konten mengisi seluruh layar.
	const FULLSCREEN_ROUTES = ['/inventory/entry'];
	$: isFullscreen = FULLSCREEN_ROUTES.some(
		(r) => currentPath === r || currentPath.startsWith(r + '/')
	);
</script>

{#if isFullscreen}
	<slot />
{:else if currentLayout === 'apple'}
	<AppleLayout {currentPath}>
		<slot />
	</AppleLayout>
{:else if currentLayout === 'sidebar'}
	<SidebarLayout {currentPath}>
		<slot />
	</SidebarLayout>
{:else if currentLayout === 'topnav'}
	<TopNavLayout {currentPath}>
		<slot />
	</TopNavLayout>
{:else if currentLayout === 'command'}
	<CommandLayout {currentPath}>
		<slot />
	</CommandLayout>
{:else if currentLayout === 'bottom'}
	<BottomLayout {currentPath}>
		<slot />
	</BottomLayout>
{/if}
