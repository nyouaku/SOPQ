import { writable, get } from 'svelte/store';
import { browser } from '$app/environment';

export interface SoundSettings {
	scanEnabled: boolean;
	successEnabled: boolean;
	errorEnabled: boolean;
	volume: number; // 0-100
}

const SOUND_KEY = 'stock-opname-sounds';

const defaultSettings: SoundSettings = {
	scanEnabled: true,
	successEnabled: true,
	errorEnabled: true,
	volume: 70
};

function createSoundStore() {
	const initialSettings = browser
		? (() => {
				try {
					const saved = localStorage.getItem(SOUND_KEY);
					return saved ? JSON.parse(saved) : defaultSettings;
				} catch {
					return defaultSettings;
				}
		  })()
		: defaultSettings;

	const { subscribe, set, update } = writable<SoundSettings>(initialSettings);

	return {
		subscribe,
		update: (settings: SoundSettings) => {
			set(settings);
			if (browser) {
				localStorage.setItem(SOUND_KEY, JSON.stringify(settings));
			}
		},
		reset: () => {
			set(defaultSettings);
			if (browser) {
				localStorage.removeItem(SOUND_KEY);
			}
		}
	};
}

export const soundSettingsStore = createSoundStore();

// Reuse one AudioContext across all calls. Browsers cap live contexts (~6 in
// Chrome), after which every new one fails and all sound silently stops. Lazy
// init so SSR/import never touches `window`.
let audioContext: AudioContext | null = null;

function getAudioContext(): AudioContext {
	if (!audioContext) {
		audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
	}
	if (audioContext.state === 'suspended') {
		audioContext.resume();
	}
	return audioContext;
}

// Play sound utility
export function playSound(type: 'scan' | 'success' | 'error') {
	const soundSettings = get(soundSettingsStore);

	// Check if sound enabled
	const enabledKey = `${type}Enabled` as keyof SoundSettings;
	if (!soundSettings[enabledKey]) return;

	// Web Audio API beep
	const audioContext = getAudioContext();
	const oscillator = audioContext.createOscillator();
	const gain = audioContext.createGain();

	oscillator.connect(gain);
	gain.connect(audioContext.destination);

	gain.gain.setValueAtTime(soundSettings.volume / 100, audioContext.currentTime);
	gain.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.3);

	switch (type) {
		case 'scan':
			oscillator.frequency.value = 800;
			oscillator.start(audioContext.currentTime);
			oscillator.stop(audioContext.currentTime + 0.1);
			break;
		case 'success':
			oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
			oscillator.frequency.setValueAtTime(1000, audioContext.currentTime + 0.1);
			oscillator.start(audioContext.currentTime);
			oscillator.stop(audioContext.currentTime + 0.2);
			break;
		case 'error':
			oscillator.frequency.setValueAtTime(400, audioContext.currentTime);
			oscillator.frequency.setValueAtTime(300, audioContext.currentTime + 0.1);
			oscillator.start(audioContext.currentTime);
			oscillator.stop(audioContext.currentTime + 0.3);
			break;
	}
}
