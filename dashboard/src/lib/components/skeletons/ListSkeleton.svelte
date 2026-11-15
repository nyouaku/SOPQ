<script lang="ts">
	export let count = 5;
</script>

<div class="space-y-3">
	{#each Array(count) as _}
		<div class="flex items-center gap-3 rounded-xl border border-base-300 bg-base-200 p-4">
			<div class="skeleton h-10 w-10 rounded-full"></div>
			<div class="flex-1 space-y-2">
				<div class="skeleton h-4 w-3/4"></div>
				<div class="skeleton h-3 w-1/2"></div>
			</div>
			<div class="skeleton h-4 w-16"></div>
		</div>
	{/each}
</div>
