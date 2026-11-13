<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let username = '';
	let password = '';
	let error = '';
	let isLoading = false;

	onMount(() => {
		// Redirect if already authenticated
		const unsubscribe = authStore.subscribe((state) => {
			if (state.isAuthenticated) {
				goto('/');
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
	<title>Login - Stock Opname Dashboard</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-base-300">
	<div class="card w-96 bg-base-100 shadow-xl border border-base-300">
		<div class="card-body p-8">
			<div class="text-center mb-6">
				<div class="flex items-center justify-center mb-4">
					<svg xmlns="http://www.w3.org/2000/svg" class="w-16 h-16 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
					</svg>
				</div>
				<h1 class="text-2xl font-bold font-mono tracking-wide text-base-content">STOCK OPNAME</h1>
				<p class="text-sm mt-2 font-mono text-base-content/70">System Dashboard v1.0</p>
			</div>

			<form on:submit|preventDefault={handleSubmit}>
				{#if error}
					<div class="mb-4 p-3 rounded bg-error/10 border border-error">
						<div class="flex items-center gap-2">
							<span class="w-2 h-2 rounded-full bg-error animate-pulse"></span>
							<span class="text-sm text-error">{error}</span>
						</div>
					</div>
				{/if}

				<div class="mb-4">
					<label class="block text-sm font-mono mb-2 text-base-content/70" for="username">
						USERNAME
					</label>
					<input
						id="username"
						type="text"
						placeholder="Enter username"
						class="w-full px-3 py-2 font-mono bg-base-200 border border-base-300 rounded focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/10"
						bind:value={username}
						disabled={isLoading}
					/>
				</div>

				<div class="mb-6">
					<label class="block text-sm font-mono mb-2 text-base-content/70" for="password">
						PASSWORD
					</label>
					<input
						id="password"
						type="password"
						placeholder="Enter password"
						class="w-full px-3 py-2 font-mono bg-base-200 border border-base-300 rounded focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/10"
						bind:value={password}
						disabled={isLoading}
					/>
				</div>

				<button
					type="submit"
					class="w-full py-2.5 font-mono font-semibold tracking-wide bg-primary text-primary-content rounded hover:opacity-90 transition-all disabled:opacity-60 disabled:cursor-not-allowed"
					disabled={isLoading}
				>
					{#if isLoading}
						<span class="inline-block animate-spin mr-2">⟳</span>
						AUTHENTICATING...
					{:else}
						SYSTEM LOGIN →
					{/if}
				</button>
			</form>

			<div class="mt-6 pt-6 border-t border-base-300">
				<div class="flex items-center justify-between text-xs font-mono text-base-content/50">
					<span>NODE: PRIMARY</span>
					<span class="flex items-center gap-1">
						<span class="w-2 h-2 rounded-full bg-success animate-pulse"></span>
						ONLINE
					</span>
				</div>
			</div>
		</div>
	</div>
</div>
