<script lang="ts">
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import WarehousesView from '$lib/components/WarehousesView.svelte';
</script>

<svelte:head>
	<title>Gudang - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<WarehousesView readOnly={true} />
</LayoutWrapper>
