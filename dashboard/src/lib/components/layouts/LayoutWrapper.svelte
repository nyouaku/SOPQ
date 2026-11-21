<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { themeStore } from '$lib/stores/theme';
	import type { ThemeLayout } from '$lib/stores/theme';
	import { applyUiTheme } from '$lib/stores/ui-theme';

	import AppleLayout from './AppleLayout.svelte';
	import FocusLayout from './FocusLayout.svelte';
	import WorkspaceLayout from './WorkspaceLayout.svelte';
	import RetroLayout from './RetroLayout.svelte';
	import NeoLayout from './NeoLayout.svelte';
	import AcrylicLayout from './AcrylicLayout.svelte';
	import NeumoLayout from './NeumoLayout.svelte';
	import EditorialLayout from './EditorialLayout.svelte';

	let currentLayout: ThemeLayout = 'apple';
	let currentPath = '';

	onMount(() => {
		themeStore.subscribe((theme) => {
			currentLayout = theme.layout;
		});
	});

	$: currentPath = $page.url.pathname;
	$: if (typeof document !== 'undefined') applyUiTheme(currentLayout);
</script>

{#if currentLayout === 'focus'}
	<FocusLayout><slot /></FocusLayout>
{:else if currentLayout === 'workspace'}
	<WorkspaceLayout><slot /></WorkspaceLayout>
{:else if currentLayout === 'retro'}
	<RetroLayout><slot /></RetroLayout>
{:else if currentLayout === 'neo'}
	<NeoLayout><slot /></NeoLayout>
{:else if currentLayout === 'acrylic'}
	<AcrylicLayout><slot /></AcrylicLayout>
{:else if currentLayout === 'neumo'}
	<NeumoLayout><slot /></NeumoLayout>
{:else if currentLayout === 'editorial'}
	<EditorialLayout><slot /></EditorialLayout>
{:else}
	<AppleLayout {currentPath}><slot /></AppleLayout>
{/if}
