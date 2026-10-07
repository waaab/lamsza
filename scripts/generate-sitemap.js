#!/usr/bin/env node
/**
 * Builds `dist/sitemap.xml` after `vite build`.
 *
 * The app ships through adapter-static, so there is no server to answer a
 * `/sitemap.xml` route at runtime. The file is produced here instead, from two
 * sources:
 *   1. the prerendered pages that already exist as `.html` files in `dist`
 *   2. the public backend endpoints, for the slugs that are not prerendered
 *      (listings, counties, settlements, széks, venues, upcoming events)
 *
 * The backend is optional: if it cannot be reached the prerendered pages are
 * still written and the missing groups are reported as warnings. Set
 * SITEMAP_REQUIRE_API=1 to make a missing backend fail the build instead.
 *
 * Env:
 *   SITE_ORIGIN           public origin for <loc> (default: SITE_ORIGIN from
 *                         src/lib/seo.js, the same host canonical tags use)
 *   SITEMAP_API_ORIGIN    backend base URL      (default http://localhost:3001)
 *   SITEMAP_OUT_DIR       build output dir      (default dist)
 *   SITEMAP_REQUIRE_API   1 = fail when a backend group cannot be fetched
 *
 * No <lastmod> is emitted. None of the public endpoints expose an updated-at
 * field, and a build-time timestamp on every URL would just teach crawlers to
 * ignore the hint.
 */

import { readdir, readFile, writeFile, unlink, stat } from 'node:fs/promises';
import path from 'node:path';
import process from 'node:process';
import { pathToFileURL } from 'node:url';
import { SITE_ORIGIN as CANONICAL_ORIGIN } from '../src/lib/seo.js';

export const SITE_ORIGIN = (process.env.SITE_ORIGIN || CANONICAL_ORIGIN).replace(/\/+$/, '');
const API_ORIGIN = (process.env.SITEMAP_API_ORIGIN || 'http://localhost:3001').replace(/\/+$/, '');
const OUT_DIR = path.resolve(process.cwd(), process.env.SITEMAP_OUT_DIR || 'dist');
const REQUIRE_API = process.env.SITEMAP_REQUIRE_API === '1';
const API_TIMEOUT_MS = Number(process.env.SITEMAP_API_TIMEOUT_MS || 20000);

/** sitemaps.org caps one file at 50,000 URLs / 50 MB. Stay under both. */
export const MAX_URLS_PER_FILE = 45000;
const MAX_BYTES_PER_FILE = 45 * 1024 * 1024;

/** `appDir` in svelte.config.js — build assets, never pages. */
const ASSET_DIR = 'app';

/** Signed-in-only pages. Kept out of the sitemap and Disallow-ed in robots.txt. */
export const PRIVATE_PATHS = ['/beallitasok', '/fiok', '/profil'];

/**
 * Legacy paths that only 301 elsewhere; a sitemap should list the target.
 * `/{slug}-szek` is anchored to one segment so a szék whose own slug ends in
 * `-szek` (`/szekek/maros-szek`) is not swept up with it.
 */
const REDIRECT_ONLY = [/^\/szek$/, /^\/szek\//, /^\/[^/]+-szek$/];

/** @param {string} p */
export function isExcluded(p) {
	if (PRIVATE_PATHS.some((priv) => p === priv || p.startsWith(`${priv}/`))) return true;
	return REDIRECT_ONLY.some((re) => re.test(p));
}

/**
 * Prerendered `.html` files under `dir` → URL paths.
 * @param {string} dir
 * @param {string} prefix
 */
export async function readPrerenderedPaths(dir, prefix = '') {
	/** @type {string[]} */
	const found = [];
	let entries;
	try {
		entries = await readdir(dir, { withFileTypes: true });
	} catch {
		return found;
	}
	for (const entry of entries) {
		if (entry.isDirectory()) {
			if (prefix === '' && entry.name === ASSET_DIR) continue;
			found.push(...(await readPrerenderedPaths(path.join(dir, entry.name), `${prefix}/${entry.name}`)));
			continue;
		}
		if (!entry.isFile() || !entry.name.endsWith('.html')) continue;
		// `app.html` is the SPA fallback, not a page of its own.
		if (prefix === '' && entry.name === 'app.html') continue;
		const base = entry.name.slice(0, -'.html'.length);
		found.push(base === 'index' ? prefix || '/' : `${prefix}/${base}`);
	}
	return found;
}

/** @param {string} pathname */
async function apiJson(pathname) {
	const res = await fetch(`${API_ORIGIN}${pathname}`, {
		headers: { accept: 'application/json' },
		signal: AbortSignal.timeout(API_TIMEOUT_MS)
	});
	if (!res.ok) throw new Error(`${pathname} → HTTP ${res.status}`);
	return res.json();
}

/** @param {unknown} value */
function asArray(value) {
	return Array.isArray(value) ? value : [];
}

/** @param {unknown} value */
function slug(value) {
	return String(value ?? '').trim();
}

function today() {
	return new Date().toISOString().slice(0, 10);
}

/**
 * One group per backend endpoint, so a single failing endpoint only costs its
 * own URLs. `/api/locations` carries counties and settlements together, which
 * is also what the `/megyek`, `/varosok` and `/falvak` pages link from.
 * @type {{ name: string, load: () => Promise<string[]> }[]}
 */
const API_GROUPS = [
	{
		name: 'counties + settlements',
		async load() {
			const rows = asArray(await apiJson('/api/locations'));
			/** @type {string[]} */
			const paths = [];
			for (const row of rows) {
				const own = slug(row?.slug);
				if (!own) continue;
				if (String(row?.type ?? '').toLowerCase() === 'megye') {
					paths.push(`/${own}-megye`);
					continue;
				}
				const county = slug(row?.county_slug);
				if (county) paths.push(`/${county}-megye/${own}`);
			}
			return paths;
		}
	},
	{
		name: 'széks',
		async load() {
			return asArray(await apiJson('/api/historical_seats'))
				.map((seat) => slug(seat?.slug))
				.filter(Boolean)
				.map((s) => `/szekek/${s}`);
		}
	},
	{
		name: 'listings',
		async load() {
			// `/api/entries` already filters to published entries.
			return asArray(await apiJson('/api/entries'))
				.map((entry) => slug(entry?.slug))
				.filter(Boolean)
				.map((s) => `/bejegyzes/${s}`);
		}
	},
	{
		name: 'venues',
		async load() {
			/** @type {string[]} */
			const paths = [];
			for (const venue of asArray(await apiJson('/api/venues'))) {
				const county = slug(venue?.county_slug);
				const settlement = slug(venue?.settlement_slug);
				const own = slug(venue?.slug);
				if (county && settlement && own) {
					paths.push(`/${county}-megye/${settlement}/helyszin/${own}`);
				}
			}
			return paths;
		}
	},
	{
		name: 'upcoming events',
		async load() {
			// Past events stay out: their detail pages have no ongoing search value.
			const body = await apiJson(`/api/events?date_from=${today()}`);
			return asArray(body?.events)
				.map((event) => slug(event?.id))
				.filter(Boolean)
				.map((id) => `/esemenyek/${id}`);
		}
	}
];

/** @param {string} value */
function xmlEscape(value) {
	return value
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&apos;');
}

/** @param {string} pathname */
export function toLoc(pathname) {
	return xmlEscape(`${SITE_ORIGIN}${encodeURI(pathname === '/' ? '/' : pathname)}`);
}

/** @param {string[]} locs */
export function renderUrlset(locs) {
	const body = locs.map((loc) => `  <url>\n    <loc>${loc}</loc>\n  </url>`).join('\n');
	return `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${body}\n</urlset>\n`;
}

/** @param {string[]} names */
export function renderIndex(names) {
	const body = names
		.map((name) => `  <sitemap>\n    <loc>${toLoc(`/${name}`)}</loc>\n  </sitemap>`)
		.join('\n');
	return `<?xml version="1.0" encoding="UTF-8"?>\n<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${body}\n</sitemapindex>\n`;
}

/**
 * Splits on whichever limit is hit first: URL count or serialised byte size.
 * @param {string[]} locs
 */
export function chunkLocs(locs) {
	/** @type {string[][]} */
	const chunks = [];
	let current = [];
	let bytes = 0;
	// Per-entry cost of `  <url>\n    <loc></loc>\n  </url>\n` plus the wrapper.
	const perEntryOverhead = 34;
	const wrapperBytes = 200;
	for (const loc of locs) {
		const size = Buffer.byteLength(loc) + perEntryOverhead;
		if (
			current.length >= MAX_URLS_PER_FILE ||
			(current.length && wrapperBytes + bytes + size > MAX_BYTES_PER_FILE)
		) {
			chunks.push(current);
			current = [];
			bytes = 0;
		}
		current.push(loc);
		bytes += size;
	}
	if (current.length) chunks.push(current);
	return chunks;
}

/** Drops `sitemap-N.xml` files left behind by an earlier, larger build. */
async function removeStaleParts(keep) {
	let entries;
	try {
		entries = await readdir(OUT_DIR);
	} catch {
		return;
	}
	for (const name of entries) {
		if (!/^sitemap-\d+\.xml$/.test(name) || keep.has(name)) continue;
		await unlink(path.join(OUT_DIR, name));
	}
}

/** robots.txt is copied from `static/`; check the built copy agrees with it. */
async function checkRobots(indexName) {
	const robotsPath = path.join(OUT_DIR, 'robots.txt');
	const expected = `Sitemap: ${SITE_ORIGIN}/${indexName}`;
	try {
		const robots = await readFile(robotsPath, 'utf8');
		if (!robots.includes(expected)) {
			console.warn(`[sitemap] warning: ${robotsPath} has no "${expected}" line`);
		}
	} catch {
		console.warn(`[sitemap] warning: ${robotsPath} is missing`);
	}
}

async function main() {
	try {
		const info = await stat(OUT_DIR);
		if (!info.isDirectory()) throw new Error('not a directory');
	} catch {
		console.error(`[sitemap] ${OUT_DIR} does not exist — run the build first.`);
		process.exit(1);
	}

	const paths = new Set();
	/** @param {string[]} list */
	const add = (list) => {
		let kept = 0;
		for (const raw of list) {
			const p = raw === '/' ? '/' : `/${String(raw).replace(/^\/+|\/+$/g, '')}`;
			if (p !== '/' && isExcluded(p)) continue;
			paths.add(p);
			kept += 1;
		}
		return kept;
	};

	console.log(`[sitemap] prerendered pages: ${add(await readPrerenderedPaths(OUT_DIR))}`);

	let failed = 0;
	for (const group of API_GROUPS) {
		try {
			const count = add(await group.load());
			console.log(`[sitemap] ${group.name}: ${count}`);
		} catch (err) {
			failed += 1;
			console.warn(`[sitemap] ${group.name}: skipped (${err instanceof Error ? err.message : err})`);
		}
	}
	if (failed && REQUIRE_API) {
		console.error(`[sitemap] ${failed} backend group(s) failed and SITEMAP_REQUIRE_API=1.`);
		process.exit(1);
	}
	if (failed) {
		console.warn(`[sitemap] ${failed} backend group(s) unavailable at ${API_ORIGIN} — sitemap is incomplete.`);
	}

	const locs = [...paths].sort().map(toLoc);
	const chunks = chunkLocs(locs);
	const outLabel = path.relative(process.cwd(), OUT_DIR) || OUT_DIR;

	if (chunks.length <= 1) {
		await removeStaleParts(new Set());
		await writeFile(path.join(OUT_DIR, 'sitemap.xml'), renderUrlset(chunks[0] ?? []), 'utf8');
		console.log(`[sitemap] wrote ${outLabel}/sitemap.xml (${locs.length} URLs)`);
	} else {
		const names = chunks.map((_, i) => `sitemap-${i + 1}.xml`);
		await removeStaleParts(new Set(names));
		await Promise.all(
			chunks.map((part, i) => writeFile(path.join(OUT_DIR, names[i]), renderUrlset(part), 'utf8'))
		);
		await writeFile(path.join(OUT_DIR, 'sitemap.xml'), renderIndex(names), 'utf8');
		console.log(`[sitemap] wrote ${outLabel}/sitemap.xml index + ${names.length} parts (${locs.length} URLs)`);
	}

	await checkRobots('sitemap.xml');
}

const invokedDirectly =
	process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href;
if (invokedDirectly) {
	await main();
}
