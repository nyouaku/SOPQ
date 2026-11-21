<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { themeStore } from '$lib/stores/theme';
	import type { ThemeLayout } from '$lib/stores/theme';

	import AppleLayout from './AppleLayout.svelte';
	import FocusLayout from './FocusLayout.svelte';
	import WorkspaceLayout from './WorkspaceLayout.svelte';

	let currentLayout: ThemeLayout = 'apple';
	let currentPath = '';

	onMount(() => {
		themeStore.subscribe((theme) => {
			currentLayout = theme.layout;
		});
	});

	$: currentPath = $page.url.pathname;
</script>

{#if currentLayout === 'focus'}
	<FocusLayout><slot /></FocusLayout>
{:else if currentLayout === 'workspace'}
	<WorkspaceLayout><slot /></WorkspaceLayout>
{:else}
	<AppleLayout {currentPath}><slot /></AppleLayout>
{/if}
