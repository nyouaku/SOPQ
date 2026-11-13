<script lang="ts">
	import { onMount } from 'svelte';
	import { navigating } from '$app/stores';

	let progress = 0;
	let isLoading = false;
	let interval: any;

	$: if ($navigating) {
		startLoading();
	} else {
		completeLoading();
	}

	function startLoading() {
		isLoading = true;
		progress = 0;

		// Simulate progressive loading
		interval = setInterval(() => {
			if (progress < 90) {
				progress += Math.random() * 15;
			}
		}, 200);
	}

	function completeLoading() {
		if (interval) clearInterval(interval);
		progress = 100;

		setTimeout(() => {
			isLoading = false;
			progress = 0;
		}, 300);
	}

	onMount(() => {
		return () => {
			if (interval) clearInterval(interval);
		};
	});
</script>

{#if isLoading}
	<div class="loading-bar-container">
		<div class="loading-bar" style="width: {progress}%"></div>
	</div>
{/if}

<style>
	.loading-bar-container {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 3px;
		z-index: 9999;
		background: transparent;
	}

	.loading-bar {
		height: 100%;
		background: linear-gradient(90deg, #5eead4 0%, #f4b740 50%, #5eead4 100%);
		transition: width 0.2s ease-out;
		box-shadow: 0 0 10px rgba(94, 234, 212, 0.8);
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.7;
		}
	}

	.loading-bar {
		animation: pulse 1.5s ease-in-out infinite;
	}
</style>
