<script lang="ts">
	import { onDestroy } from 'svelte';
	import { soundSettingsStore, type SoundSettings } from '$lib/stores/soundSettings';
	import { playSound } from '$lib/stores/soundSettings';

	type SoundType = 'scan' | 'success' | 'error';

	const SOUND_LABELS: Record<SoundType, string> = {
		scan: 'Suara Scan',
		success: 'Suara Berhasil',
		error: 'Suara Error'
	};

	const SOUND_DESCRIPTIONS: Record<SoundType, string> = {
		scan: 'Diputar saat QR code terdeteksi',
		success: 'Diputar saat data berhasil disimpan',
		error: 'Diputar saat terjadi kesalahan'
	};

	let settings: SoundSettings = {
		scanEnabled: true,
		successEnabled: true,
		errorEnabled: true,
		volume: 70
	};

	// Subscribe — selalu tersubscribe selama komponen hidup
	const unsub = soundSettingsStore.subscribe((s) => (settings = s));
	onDestroy(unsub);

	function toggle(key: 'scanEnabled' | 'successEnabled' | 'errorEnabled', checked: boolean) {
		soundSettingsStore.update({ ...settings, [key]: checked });
	}

	function changeVolume(v: number) {
		soundSettingsStore.update({ ...settings, volume: v });
	}

	function test(type: SoundType) {
		playSound(type);
	}
</script>

<div class="space-y-4">
	{#each Object.keys(SOUND_LABELS) as key}
		{@const st = key as SoundType}
		{@const enabledKey = `${st}Enabled` as 'scanEnabled' | 'successEnabled' | 'errorEnabled'}
		<div class="flex items-center justify-between gap-4 rounded-xl border border-base-300 bg-base-100 p-4">
			<div class="flex-1">
				<label class="flex cursor-pointer items-center gap-3">
					<input
						type="checkbox"
						class="toggle toggle-primary"
						checked={settings[enabledKey]}
						on:change={(e) => toggle(enabledKey, e.target.checked)}
					/>
					<span class="flex-1">
						<span class="block font-medium">{SOUND_LABELS[st]}</span>
						<span class="block text-xs text-base-content/60">{SOUND_DESCRIPTIONS[st]}</span>
					</span>
				</label>
			</div>
			<button
				class="btn btn-ghost btn-sm btn-circle shrink-0"
				title="Test {SOUND_LABELS[st]}"
				aria-label="Test {SOUND_LABELS[st]}"
				on:click={() => test(st)}
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
			on:input={(e) => changeVolume(parseInt(e.target.value))}
			class="range range-primary range-sm"
			aria-label="Master volume"
		/>
		<div class="mt-1 flex justify-between text-xs text-base-content/50">
			<span>Diam</span>
			<span>Keras</span>
		</div>
	</div>
</div>
