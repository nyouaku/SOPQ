/// <reference types="@sveltejs/kit" />
import { build, files, version } from '$service-worker';

/**
 * Service worker minimal tanpa dependensi tambahan.
 * - Shell aplikasi di-precache saat install.
 * - Request halaman: network-first, fallback ke cache (mode offline).
 * - Request API: tidak pernah di-cache (data stok harus selalu segar).
 */

const CACHE = `stock-opname-${version}`;
const ASSETS = [...build, ...files];

self.addEventListener('install', (event) => {
	event.waitUntil(
		caches.open(CACHE).then((cache) => cache.addAll(ASSETS)).then(() => self.skipWaiting())
	);
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
			.then(() => self.clients.claim())
	);
});

self.addEventListener('fetch', (event) => {
	const { request } = event;
	if (request.method !== 'GET') return;

	const url = new URL(request.url);
	if (url.origin !== self.location.origin) return;

	// API dan WebSocket selalu lewat jaringan.
	if (url.pathname.startsWith('/api/')) return;

	// Aset ber-hash: cache-first.
	if (ASSETS.includes(url.pathname)) {
		event.respondWith(caches.match(request).then((hit) => hit || fetch(request)));
		return;
	}

	// Navigasi halaman: network-first, fallback cache.
	event.respondWith(
		fetch(request)
			.then((response) => {
				const copy = response.clone();
				caches.open(CACHE).then((cache) => cache.put(request, copy));
				return response;
			})
			.catch(() => caches.match(request).then((hit) => hit || caches.match('/')))
	);
});
