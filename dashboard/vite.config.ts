import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	// Dev: API diteruskan ke nginx (yang forward ke backend di docker).
	server: {
		proxy: {
			'/api': { target: 'http://localhost:8088', changeOrigin: true, ws: true }
		}
	}
});
