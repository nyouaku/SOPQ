<script>
	export let size = 24;
	export let color = '#ff6b35';
</script>

<div class="loader-2" style="--size: {size}px; --color: {color}">
	<svg viewBox="0 0 50 50">
		<circle class="circle-1" cx="25" cy="25" r="15" />
		<circle class="circle-2" cx="25" cy="25" r="15" />
	</svg>
</div>

<style>
	.loader-2 {
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	svg {
		width: var(--size);
		height: var(--size);
		animation: spin 2s linear infinite;
	}

	circle {
		fill: none;
		stroke: var(--color);
		stroke-width: 2;
		stroke-linecap: round;
	}

	.circle-1 {
		stroke-dasharray: 94.2;
		stroke-dashoffset: 0;
		animation: draw1 1.5s ease-in-out infinite;
	}

	.circle-2 {
		stroke-dasharray: 94.2;
		stroke-dashoffset: 94.2;
		animation: draw2 1.5s ease-in-out infinite;
	}

	@keyframes spin {
		100% {
			transform: rotate(360deg);
		}
	}

	@keyframes draw1 {
		0% {
			stroke-dashoffset: 0;
		}
		50% {
			stroke-dashoffset: 47.1;
		}
		100% {
			stroke-dashoffset: 94.2;
		}
	}

	@keyframes draw2 {
		0% {
			stroke-dashoffset: 94.2;
		}
		50% {
			stroke-dashoffset: 47.1;
		}
		100% {
			stroke-dashoffset: 0;
		}
	}
</style>
