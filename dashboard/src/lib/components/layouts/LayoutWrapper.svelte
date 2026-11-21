<script lang="ts">
	import { page } from '$app/stores';
	import AppleLayout from './AppleLayout.svelte';

	let currentPath = '';
	$: currentPath = $page.url.pathname;
</script>

<AppleLayout {currentPath}>
	<slot />
</AppleLayout>
