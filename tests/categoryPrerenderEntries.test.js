import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";

import { DIRECTORY_CATALOG, directoryPrerenderSlugs } from "../src/lib/entryCategory.js";

const ROUTE = "src/routes/(public)/index/[category]/+page.js";

/**
 * `/index/[category]` used to hand-write the slugs it prerendered. The list
 * drifted away from the catalog and shipped four pages for categories that no
 * longer exist (egeszsegugy, egyeb, vendeglo, bolt). Those URLs returned 200 with
 * an empty category, so a crawler could index them. The list is now derived.
 */

/** Slugs the v1 tree had and the v2 tree does not. */
const RETIRED_SLUGS = ["egeszsegugy", "egyeb", "vendeglo", "bolt"];

test("prerendered slugs are exactly the catalog slugs", () => {
	const slugs = directoryPrerenderSlugs();
	assert.deepEqual(
		[...slugs].sort(),
		[...new Set(DIRECTORY_CATALOG.map((row) => row.slug))].sort(),
	);
});

test("prerendered slugs include parents and children", () => {
	const slugs = new Set(directoryPrerenderSlugs());
	assert.ok(slugs.has("etkezes"), "parent shelf");
	assert.ok(slugs.has("etterem"), "child category");
	assert.ok(slugs.has("sportegyesulet"), "child of sport-es-szabadido");
});

test("no retired slug is prerendered", () => {
	const slugs = new Set(directoryPrerenderSlugs());
	for (const slug of RETIRED_SLUGS) {
		assert.equal(slugs.has(slug), false, `${slug} is not in the catalog any more`);
	}
});

test("every prerendered slug resolves to a catalog row", () => {
	const known = new Set(DIRECTORY_CATALOG.map((row) => row.slug));
	for (const slug of directoryPrerenderSlugs()) {
		assert.ok(known.has(slug), `${slug} has no catalog row`);
	}
});

test("the route derives its entries instead of listing them", () => {
	const source = readFileSync(new URL(`../${ROUTE}`, import.meta.url), "utf8");
	assert.match(source, /directoryPrerenderSlugs/, "route must call the derived list");
	const hardCoded = source.match(/\{\s*category:\s*['"]/g);
	assert.equal(hardCoded, null, "route must not hard-code category slugs");
});

test("directoryPrerenderSlugs is deduped and falls back to the catalog", () => {
	assert.deepEqual(
		directoryPrerenderSlugs([{ slug: "etkezes" }, { slug: "etkezes" }, { slug: "" }]),
		["etkezes"],
	);
	assert.deepEqual(directoryPrerenderSlugs([]), directoryPrerenderSlugs());
	assert.deepEqual(directoryPrerenderSlugs(null), directoryPrerenderSlugs());
});
