<script lang="ts">
	export let rows = 5;
	export let columns = 5;
</script>

<div class="overflow-x-auto">
	<table class="table">
		<thead>
			<tr>
				{#each Array(columns) as _}
					<th><div class="skeleton h-4 w-20"></div></th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each Array(rows) as _}
				<tr>
					{#each Array(columns) as _}
						<td><div class="skeleton h-4 w-full"></div></td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
</div>
