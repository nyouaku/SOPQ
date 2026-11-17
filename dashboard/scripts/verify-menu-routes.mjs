// Self-check: setiap href di menu.ts harus menunjuk route yang ada.
// Jalankan: node scripts/verify-menu-routes.mjs
import { readFileSync, readdirSync } from 'node:fs';

const menuSrc = readFileSync('src/lib/components/layouts/menu.ts', 'utf8');
const hrefs = [...menuSrc.matchAll(/href: ['"]([^'"]+)['"]/g)].map((m) => m[1]);

const routes = readdirSync('src/routes', { recursive: true })
	.filter((f) => String(f).endsWith('+page.svelte'))
	.map((f) => {
		let p = '/' + String(f).replace(/\\/g, '/').replace(/\/?\+page\.svelte$/, '');
		return p === '//' ? '/' : p;
	});

const dynamic = routes.filter((r) => r.includes('['));
function exists(h) {
	if (h === '/') return routes.includes('/');
	if (routes.includes(h)) return true;
	return dynamic.some((d) => h.startsWith(d.replace(/\[.*\]/, '')));
}

const bad = hrefs.filter((h) => !exists(h));
console.log('href menu:', hrefs.length);
if (bad.length) {
	console.error('TAK ADA ROUTE:', bad.join(', '));
	process.exit(1);
}
console.log('semua href menu punya route ✓');
