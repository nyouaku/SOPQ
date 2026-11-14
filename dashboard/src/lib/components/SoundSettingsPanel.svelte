<script lang="ts">
	import { soundSettingsStore, type SoundSettings } from '$lib/stores/soundSettings';
	import { playSound } from '$lib/stores/soundSettings';

	let settings: SoundSettings;
	soundSettingsStore.subscribe(s => (settings = s))();

	function updateSettings(key: keyof SoundSettings, value: any) {
		soundSettingsStore.update({ ...settings, [key]: value });
	}

	function testSound(type: 'scan' | 'success' | 'error') {
		playSound(type);
	}
</script>

<div class="card bg-base-200 shadow-xl">
	<div class="card-body">
		<h2 class="card-title text-lg">🔊 Sound Settings</h2>
		<p class="text-sm text-base-content/60 mb-4">Customize notification sounds</p>

		<!-- Scan Sound -->
		<div class="form-control mb-4">
			<label class="label cursor-pointer">
				<span class="label-text">Scan Sound</span>
				<input
					type="checkbox"
					class="checkbox"
					checked={settings.scanEnabled}
					on:change={(e) => updateSettings('scanEnabled', e.target.checked)}
				/>
			</label>
			<button
				class="btn btn-sm btn-outline mt-2"
				on:click={() => testSound('scan')}
			>
				Test Scan Sound
			</button>
		</div>

		<!-- Success Sound -->
		<div class="form-control mb-4">
			<label class="label cursor-pointer">
				<span class="label-text">Success Sound</span>
				<input
					type="checkbox"
					class="checkbox"
					checked={settings.successEnabled}
					on:change={(e) => updateSettings('successEnabled', e.target.checked)}
				/>
			</label>
			<button
				class="btn btn-sm btn-outline mt-2"
				on:click={() => testSound('success')}
			>
				Test Success Sound
			</button>
		</div>

		<!-- Error Sound -->
		<div class="form-control mb-4">
			<label class="label cursor-pointer">
				<span class="label-text">Error Sound</span>
				<input
					type="checkbox"
					class="checkbox"
					checked={settings.errorEnabled}
					on:change={(e) => updateSettings('errorEnabled', e.target.checked)}
				/>
			</label>
			<button
				class="btn btn-sm btn-outline mt-2"
				on:click={() => testSound('error')}
			>
				Test Error Sound
			</button>
		</div>

		<!-- Volume Control -->
		<div class="form-control">
			<label class="label" for="volume">
				<span class="label-text">Volume</span>
				<span class="label-text-alt">{settings.volume}%</span>
			</label>
			<input
				id="volume"
				type="range"
				min="0"
				max="100"
				value={settings.volume}
				on:change={(e) => updateSettings('volume', parseInt(e.target.value))}
				class="range range-sm"
			/>
		</div>
	</div>
</div>

<style>
	.card {
		border-radius: 12px;
	}
</style>
