<script lang="ts">
	import LayoutWrapper from '$lib/components/layouts/LayoutWrapper.svelte';
	import ProductsView from '$lib/components/ProductsView.svelte';

	// Read-only mode (Stock Opname app).
</script>

<svelte:head>
	<title>Products - Stock Opname</title>
</svelte:head>

<LayoutWrapper>
	<ProductsView readOnly={true} subtitle="Lihat daftar produk (baca-saja)" />
</LayoutWrapper>
