<script>
	export let label = 'Memuat...';
</script>

<div class="preloader-2">
	<div class="container">
		<div class="spinner">
			<svg viewBox="0 0 50 50">
				<circle class="path" cx="25" cy="25" r="20" fill="none" stroke="#ff6b35" stroke-width="2" />
			</svg>
		</div>
		<p class="label">{label}</p>
	</div>
</div>

<style>
	.preloader-2 {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, #ffffff 0%, #f9fafb 100%);
		z-index: 9999;
	}

	.container {
		text-align: center;
	}

	.spinner {
		width: 60px;
		height: 60px;
		margin: 0 auto 24px;
	}

	svg {
		width: 100%;
		height: 100%;
		animation: rotate 2s linear infinite;
	}

	.path {
		stroke-dasharray: 125.6;
		stroke-dashoffset: 0;
		animation: dash 1.5s ease-in-out infinite;
	}

	@keyframes rotate {
		100% {
			transform: rotate(360deg);
		}
	}

	@keyframes dash {
		0% {
			stroke-dashoffset: 125.6;
		}
		50% {
			stroke-dashoffset: 31.4;
		}
		100% {
			stroke-dashoffset: 125.6;
		}
	}

	.label {
		font-size: 14px;
		color: #ff6b35;
		font-weight: 500;
		margin: 0;
		letter-spacing: 0.5px;
	}
</style>
