<script>
	export let label = 'Memuat...';
</script>

<div class="preloader-3">
	<div class="container">
		<div class="spinner">
			<div class="bar"></div>
			<div class="bar"></div>
			<div class="bar"></div>
			<div class="bar"></div>
			<div class="bar"></div>
		</div>
		<p class="label">{label}</p>
	</div>
</div>

<style>
	.preloader-3 {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, #ffffff 0%, #f9fafb 100%);
		z-index: 9999;
	}

	.container { text-align: center; }

	.spinner {
		display: flex;
		justify-content: center;
		align-items: flex-end;
		gap: 6px;
		height: 40px;
		margin-bottom: 24px;
	}

	.bar {
		width: 4px;
		background: #ff6b35;
		border-radius: 2px;
		animation: grow 1.2s ease-in-out infinite;
	}
	.bar:nth-child(1) { height: 10px; animation-delay: 0s; }
	.bar:nth-child(2) { height: 20px; animation-delay: 0.15s; }
	.bar:nth-child(3) { height: 30px; animation-delay: 0.3s; }
	.bar:nth-child(4) { height: 20px; animation-delay: 0.45s; }
	.bar:nth-child(5) { height: 10px; animation-delay: 0.6s; }

	@keyframes grow {
		0%, 100% { opacity: 0.5; }
		50% { opacity: 1; }
	}

	.label { font-size: 14px; color: #ff6b35; font-weight: 500; margin: 0; }
</style>
