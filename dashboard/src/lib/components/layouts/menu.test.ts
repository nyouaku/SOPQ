import { readFileSync, readdirSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

/**
 * Regresi pemisahan aplikasi: setiap href di menu.ts harus menunjuk route yang
 * benar-benar ada. Mencegah menu menunjuk route yang sudah dipindah/dihapus
 * (pernah terjadi: /inventory/products bentrok /inventory/[id]).
 */
function readMenu(): string {
	return readFileSync('src/lib/components/layouts/menu.ts', 'utf8');
}

function allRoutes(): string[] {
	return readdirSync('src/routes', { recursive: true })
		.filter((f) => String(f).endsWith('+page.svelte'))
		.map((f) => {
			const p = '/' + String(f).replace(/\\/g, '/').replace(/\/?\+page\.svelte$/, '');
			return p === '//' ? '/' : p;
		});
}

describe('pemisahan aplikasi SO vs Inventory', () => {
	it('setiap href menu menunjuk route yang ada', () => {
		const hrefs = [...readMenu().matchAll(/href: ['"]([^'"]+)['"]/g)].map((m) => m[1]);
		expect(hrefs.length).toBeGreaterThan(0);

		const routes = allRoutes();
		const dynamic = routes.filter((r) => r.includes('['));
		const missing = hrefs.filter(
			(h) => !(routes.includes(h) || dynamic.some((d) => h.startsWith(d.replace(/\[.*\]/, ''))))
		);
		expect(missing).toEqual([]);
	});

	it('Products & Gudang milik Inventory (bukan SO), SO punya versi baca-saja', () => {
		const src = readMenu();
		expect(src).toMatch(/href: '\/products'.*app: 'inventory'/);
		expect(src).toMatch(/href: '\/warehouses'.*app: 'inventory'/);
		expect(src).toMatch(/href: '\/opname\/products'.*app: 'opname'/);
		expect(src).toMatch(/href: '\/opname\/warehouses'.*app: 'opname'/);
	});

	it('activeAppFor memetakan path ke aplikasi benar', async () => {
		const { activeAppFor } = await import('./menu');
		expect(activeAppFor('/inventory')).toBe('inventory');
		expect(activeAppFor('/products')).toBe('inventory');
		expect(activeAppFor('/warehouses')).toBe('inventory');
		expect(activeAppFor('/inventory/movements')).toBe('inventory');
		expect(activeAppFor('/inventory/entry')).toBe('inventory');
		expect(activeAppFor('/inventory/transfer')).toBe('inventory');
		expect(activeAppFor('/inventory/settings')).toBe('inventory');
		// Stock Opname (termasuk /transfer = pindah lokasi box, bukan inventory)
		expect(activeAppFor('/')).toBe('opname');
		expect(activeAppFor('/scan')).toBe('opname');
		expect(activeAppFor('/sessions')).toBe('opname');
		expect(activeAppFor('/opname/products')).toBe('opname');
		expect(activeAppFor('/opname/warehouses')).toBe('opname');
		expect(activeAppFor('/transfer')).toBe('opname');
	});
});
