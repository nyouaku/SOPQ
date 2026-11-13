/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			fontFamily: {
				sans: ['Inter', 'IBM Plex Sans', 'system-ui', 'sans-serif'],
				mono: ['JetBrains Mono', 'IBM Plex Mono', 'Consolas', 'monospace']
			},
			colors: {
				// Rack Room theme
				'rack-bg': '#14171C',
				'rack-surface': '#1D2229',
				'rack-border': '#2A2F38',
				'rack-teal': '#5EEAD4',
				'rack-amber': '#F4B740',
				'rack-red': '#EF5D5D',
				'rack-text': '#C7CDD6'
			}
		}
	},
	plugins: [require('daisyui')],
	daisyui: {
		themes: [
			{
				'rack-room': {
					primary: '#5EEAD4',
					secondary: '#F4B740',
					accent: '#5EEAD4',
					neutral: '#2A2F38',
					'base-100': '#14171C',
					'base-200': '#1D2229',
					'base-300': '#252A33',
					'base-content': '#C7CDD6',
					info: '#5EEAD4',
					success: '#5EEAD4',
					warning: '#F4B740',
					error: '#EF5D5D'
				},
				'noc-console': {
					primary: '#00FF9C',
					secondary: '#FFB020',
					accent: '#00FF9C',
					neutral: '#1F2620',
					'base-100': '#0B0E0C',
					'base-200': '#121712',
					'base-300': '#1A1F1A',
					'base-content': '#00FF9C',
					info: '#00FF9C',
					success: '#00FF9C',
					warning: '#FFB020',
					error: '#FF4D4D'
				},
				'daylight-ops': {
					primary: '#1D4ED8',
					secondary: '#B45309',
					accent: '#1D4ED8',
					neutral: '#D8DEE6',
					'base-100': '#F7F8FA',
					'base-200': '#FFFFFF',
					'base-300': '#F0F2F5',
					'base-content': '#1C222B',
					info: '#1D4ED8',
					success: '#1D4ED8',
					warning: '#B45309',
					error: '#B91C1C'
				},
				'saweria': {
					primary: '#FF6B35',
					secondary: '#F7931E',
					accent: '#FF6B35',
					neutral: '#2D3142',
					'base-100': '#FFFFFF',
					'base-200': '#F8F9FA',
					'base-300': '#E9ECEF',
					'base-content': '#2D3142',
					info: '#4ECDC4',
					success: '#95E1D3',
					warning: '#F7931E',
					error: '#E63946'
				},
				'modern-list': {
					primary: '#667EEA',
					secondary: '#764BA2',
					accent: '#667EEA',
					neutral: '#1F2937',
					'base-100': '#F9FAFB',
					'base-200': '#FFFFFF',
					'base-300': '#F3F4F6',
					'base-content': '#111827',
					info: '#3B82F6',
					success: '#10B981',
					warning: '#F59E0B',
					error: '#EF4444'
				}
			}
		],
		darkTheme: 'rack-room',
		base: true,
		styled: true,
		utils: true
	}
};
