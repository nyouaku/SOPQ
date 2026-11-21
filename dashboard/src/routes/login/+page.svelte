<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	let username = '';
	let password = '';
	let error = '';
	let isLoading = false;

	onMount(() => {
		// Redirect if already authenticated — ke application selector.
		const unsubscribe = authStore.subscribe((state) => {
			if (state.isAuthenticated) {
				goto('/apps');
			}
		});
		return unsubscribe;
	});

	async function handleSubmit() {
		if (!username || !password) {
			error = 'Username dan password tidak boleh kosong';
			return;
		}

		isLoading = true;
		error = '';

		const result = await authStore.login(username, password);

		if (!result.success) {
			error = result.error || 'Login gagal';
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Masuk - SOPQ</title>
</svelte:head>

<div class="login">
	<!-- Ambient animated gradient -->
	<div class="ambient" aria-hidden="true">
		<div class="ambient__blob ambient__blob--a"></div>
		<div class="ambient__blob ambient__blob--b"></div>
		<div class="ambient__blob ambient__blob--c"></div>
	</div>

	<!-- Auth panel -->
	<div class="panel glass" in:fly={{ y: 18, duration: 480, easing: cubicOut }}>
		<div class="brand">
			<div class="brand__mark">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"
					stroke-linecap="round" stroke-linejoin="round">
					<path d="M3 7l9-4 9 4v10l-9 4-9-4V7zM3 7l9 4 9-4M12 11v10" />
				</svg>
			</div>
			<h1 class="brand__name">SOPQ</h1>
			<p class="brand__sub">Warehouse Suite</p>
		</div>

		<form on:submit|preventDefault={handleSubmit}>
			{#if error}
				<div class="err" transition:fade={{ duration: 180 }}>
					<span class="err__dot"></span>
					<span>{error}</span>
				</div>
			{/if}

			<div class="field">
				<label for="username">Username</label>
				<input
					id="username"
					type="text"
					autocomplete="username"
					placeholder="Masukkan username"
					bind:value={username}
					disabled={isLoading}
				/>
			</div>

			<div class="field">
				<label for="password">Password</label>
				<input
					id="password"
					type="password"
					autocomplete="current-password"
					placeholder="Masukkan password"
					bind:value={password}
					disabled={isLoading}
				/>
			</div>

			<button type="submit" class="submit press" disabled={isLoading}>
				{#if isLoading}
					<span class="spinner"></span>
					Memproses...
				{:else}
					Masuk
				{/if}
			</button>
		</form>
	</div>

	<p class="footnote">SOPQ · Warehouse Suite</p>
</div>

<style>
	.login {
		min-height: 100dvh;
		position: relative;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		background: oklch(var(--b2));
		color: oklch(var(--bc));
	}
	.ambient {
		position: absolute;
		inset: 0;
		pointer-events: none;
	}
	.ambient__blob {
		position: absolute;
		border-radius: 50%;
		filter: blur(90px);
		opacity: 0.5;
	}
	.ambient__blob--a {
		width: 38rem;
		height: 38rem;
		top: -12rem;
		left: -10rem;
		background: oklch(var(--p) / 0.35);
		animation: floatA 22s var(--ease-apple) infinite alternate;
	}
	.ambient__blob--b {
		width: 32rem;
		height: 32rem;
		bottom: -10rem;
		right: -8rem;
		background: oklch(var(--s) / 0.3);
		animation: floatB 28s var(--ease-apple) infinite alternate;
	}
	.ambient__blob--c {
		width: 22rem;
		height: 22rem;
		top: 40%;
		left: 45%;
		background: oklch(var(--a) / 0.18);
		animation: floatC 26s var(--ease-apple) infinite alternate;
	}
	@keyframes floatA {
		from { transform: translate3d(0, 0, 0) scale(1); }
		to { transform: translate3d(6rem, 4rem, 0) scale(1.12); }
	}
	@keyframes floatB {
		from { transform: translate3d(0, 0, 0) scale(1); }
		to { transform: translate3d(-5rem, -5rem, 0) scale(1.1); }
	}
	@keyframes floatC {
		from { transform: translate3d(0, 0, 0) scale(1); }
		to { transform: translate3d(-4rem, 3rem, 0) scale(1.15); }
	}

	.panel {
		position: relative;
		z-index: 1;
		width: 100%;
		max-width: 24rem;
		border-radius: 1.5rem;
		padding: 2.25rem 2rem 2rem;
		box-shadow:
			0 2px 4px oklch(var(--bc) / 0.05),
			0 24px 64px oklch(var(--bc) / 0.12);
	}
	.brand {
		text-align: center;
		margin-bottom: 2rem;
	}
	.brand__mark {
		width: 3.75rem;
		height: 3.75rem;
		margin: 0 auto 1rem;
		border-radius: 1.1rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: oklch(var(--p) / 0.14);
		color: oklch(var(--p));
	}
	.brand__mark svg {
		width: 2rem;
		height: 2rem;
	}
	.brand__name {
		font-size: 1.6rem;
		font-weight: 700;
		letter-spacing: -0.03em;
	}
	.brand__sub {
		margin-top: 0.2rem;
		font-size: 0.82rem;
		color: oklch(var(--bc) / 0.55);
	}

	.field {
		margin-bottom: 1rem;
	}
	.field label {
		display: block;
		font-size: 0.8rem;
		font-weight: 500;
		margin-bottom: 0.4rem;
		color: oklch(var(--bc) / 0.7);
	}
	.field input {
		width: 100%;
		padding: 0.7rem 0.85rem;
		border-radius: 0.75rem;
		background: oklch(var(--b2) / 0.7);
		border: 1px solid color-mix(in srgb, oklch(var(--bc)) 10%, transparent);
		color: oklch(var(--bc));
		font-size: 0.95rem;
		transition: border-color var(--dur-hover) var(--ease-apple), box-shadow var(--dur-hover) var(--ease-apple);
	}
	.field input:focus {
		outline: none;
		border-color: oklch(var(--p) / 0.6);
		box-shadow: 0 0 0 4px oklch(var(--p) / 0.12);
	}
	.field input:disabled {
		opacity: 0.6;
	}

	.err {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 1rem;
		padding: 0.65rem 0.85rem;
		border-radius: 0.7rem;
		background: oklch(var(--er) / 0.1);
		color: oklch(var(--er));
		font-size: 0.85rem;
	}
	.err__dot {
		width: 0.45rem;
		height: 0.45rem;
		border-radius: 50%;
		background: oklch(var(--er));
		flex-shrink: 0;
	}

	.submit {
		width: 100%;
		margin-top: 0.5rem;
		padding: 0.8rem;
		border-radius: 0.85rem;
		font-size: 0.95rem;
		font-weight: 600;
		background: oklch(var(--p));
		color: oklch(var(--pc, 1 0 0));
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		transition: opacity var(--dur-micro) var(--ease-apple);
	}
	.submit:disabled {
		opacity: 0.6;
	}
	.spinner {
		width: 1rem;
		height: 1rem;
		border-radius: 50%;
		border: 2px solid currentColor;
		border-right-color: transparent;
		animation: spin 0.7s linear infinite;
	}
	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.footnote {
		position: absolute;
		bottom: 1.5rem;
		left: 0;
		right: 0;
		text-align: center;
		font-size: 0.72rem;
		color: oklch(var(--bc) / 0.45);
		letter-spacing: 0.04em;
	}
</style>
