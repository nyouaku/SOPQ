<script>
	export let label = 'Memuat...';
</script>

<div class="preloader-1">
	<div class="container">
		<div class="spinner">
			<div class="bounce bounce-1"></div>
			<div class="bounce bounce-2"></div>
			<div class="bounce bounce-3"></div>
		</div>
		<p class="label">{label}</p>
	</div>
</div>

<style>
	.preloader-1 {
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
		display: flex;
		justify-content: center;
		gap: 12px;
		margin-bottom: 24px;
	}

	.bounce {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background: #ff6b35;
		animation: bounce 1.4s infinite ease-in-out both;
	}

	.bounce-1 {
		animation-delay: -0.32s;
	}

	.bounce-2 {
		animation-delay: -0.16s;
	}

	@keyframes bounce {
		0%,
		80%,
		100% {
			transform: scale(0);
			opacity: 0.5;
		}
		40% {
			transform: scale(1);
			opacity: 1;
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
