<script lang="ts">
	import { onDestroy } from 'svelte';
	import { soundSettingsStore } from '$lib/stores/soundSettings';
	import { playSound } from '$lib/stores/soundSettings';

	const SOUNDS: { key: 'scanEnabled' | 'successEnabled' | 'errorEnabled'; label: string; description: string; soundType: 'scan' | 'success' | 'error' }[] = [
		{ key: 'scanEnabled', label: 'Suara Scan', description: 'Diputar saat QR code terdeteksi', soundType: 'scan' as const },
		{ key: 'successEnabled', label: 'Suara Berhasil', description: 'Diputar saat data berhasil disimpan', soundType: 'success' as const },
		{ key: 'errorEnabled', label: 'Suara Error', description: 'Diputar saat terjadi kesalahan', soundType: 'error' as const }
	];

	let settings = {
		scanEnabled: true,
		successEnabled: true,
		errorEnabled: true,
		volume: 70
	};

	const unsub = soundSettingsStore.subscribe((s) => (settings = s));
	onDestroy(unsub);

	function toggle(key: string, checked: boolean) {
		soundSettingsStore.update({ ...settings, [key]: checked });
	}

	function changeVolume(v: number) {
		soundSettingsStore.update({ ...settings, volume: v });
	}

	function test(type: 'scan' | 'success' | 'error') {
		playSound(type);
	}

	function checkedOf(e: Event): boolean {
		return (e.currentTarget as HTMLInputElement).checked;
	}

	function numValueOf(e: Event): number {
		return parseInt((e.currentTarget as HTMLInputElement).value);
	}
</script>

<div class="space-y-4">
	{#each SOUNDS as sound}
		<div class="flex items-center justify-between gap-4 rounded-xl border border-base-300 bg-base-100 p-4">
			<div class="flex-1">
				<label class="flex cursor-pointer items-center gap-3">
					<input
						type="checkbox"
						class="toggle toggle-primary"
						checked={settings[sound.key]}
						on:change={(e) => toggle(sound.key, checkedOf(e))}
					/>
					<span class="flex-1">
						<span class="block font-medium">{sound.label}</span>
						<span class="block text-xs text-base-content/60">{sound.description}</span>
					</span>
				</label>
			</div>
			<button
				class="btn btn-ghost btn-sm btn-circle shrink-0"
				title="Test {sound.label}"
				aria-label="Test {sound.label}"
				on:click={() => test(sound.soundType)}
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C11.2 3.69 12 4.037 12 4.7v14.6c0 .663-.8 1.01-1.296.53L5.586 15z" />
				</svg>
			</button>
		</div>
	{/each}

	<!-- Master Volume -->
	<div class="rounded-xl border border-base-300 bg-base-100 p-4">
		<div class="mb-2 flex items-center justify-between">
			<label class="font-medium" for="master-volume">Master Volume</label>
			<span class="badge badge-primary badge-sm">{settings.volume}%</span>
		</div>
		<input
			id="master-volume"
			type="range"
			min="0"
			max="100"
			step="5"
			value={settings.volume}
			on:input={(e) => changeVolume(numValueOf(e))}
			class="range range-primary range-sm"
			aria-label="Master volume"
		/>
	</div>
</div>
