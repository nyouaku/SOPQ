import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	// Dev: API dan WS diteruskan ke backend lokal, supaya kode tetap same-origin.
	server: {
		proxy: {
			'/api': { target: 'http://localhost:8080', changeOrigin: true, ws: true }
		}
	}
});
