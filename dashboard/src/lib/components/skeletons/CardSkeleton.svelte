<script lang="ts">
	export let count = 4;
	export let cols = 4;
</script>

<div class="grid grid-cols-1 gap-4 md:grid-cols-{cols}">
	{#each Array(count) as _}
		<div class="rounded-2xl border border-base-300 bg-base-200 p-4">
			<div class="skeleton h-6 w-3/4"></div>
			<div class="skeleton mt-3 h-4 w-1/2"></div>
			<div class="skeleton mt-2 h-4 w-2/3"></div>
		</div>
	{/each}
</div>
