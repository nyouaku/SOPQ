<script>
	export let size = 24;
	export let color = '#ff6b35';
</script>

<div class="loader-3" style="--size: {size}px; --color: {color}">
	<div class="pulse pulse-1"></div>
	<div class="pulse pulse-2"></div>
	<div class="pulse pulse-3"></div>
</div>

<style>
	.loader-3 {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		position: relative;
		width: var(--size);
		height: var(--size);
	}

	.pulse {
		position: absolute;
		width: 100%;
		height: 100%;
		border: 2px solid var(--color);
		border-radius: 50%;
		animation: pulse 2s ease-out infinite;
	}

	.pulse-1 {
		animation-delay: 0s;
	}

	.pulse-2 {
		animation-delay: 0.4s;
	}

	.pulse-3 {
		animation-delay: 0.8s;
	}

	@keyframes pulse {
		0% {
			transform: scale(0);
			opacity: 1;
		}
		100% {
			transform: scale(1);
			opacity: 0;
		}
	}
</style>
