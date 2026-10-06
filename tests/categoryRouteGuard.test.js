import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";

import {
    categoryVerdict,
    isCatalogCategory,
    isLiveCategory,
    normalizeCategorySlug,
} from "../src/lib/categoryRouteGuard.js";

const ROUTE = "src/routes/(public)/index/[category]/+page.js";
const ERROR_PAGE = "src/routes/+error.svelte";

/** Slugs the v1 tree had and the v2 tree does not. */
const RETIRED_SLUGS = ["egeszsegugy", "egyeb", "vendeglo", "bolt"];

/** The live list the backend returns, trimmed to the fields the guard reads. */
const LIVE_ROWS = [
    { slug: "etkezes", name: "Étkezés" },
    { slug: "etterem", name: "Étterem" },
    { slug: "sportegyesulet", name: "Sportegyesület" },
];

test("a catalog slug is known without asking the backend", () => {
    for (const slug of ["etkezes", "etterem", "sportegyesulet", "webfejlesztes"]) {
        assert.equal(isCatalogCategory(slug), true, slug);
        assert.equal(categoryVerdict(slug, null), "known", slug);
    }
});

test("a retired slug is a 404 when the backend list does not have it", () => {
    for (const slug of RETIRED_SLUGS) {
        assert.equal(isCatalogCategory(slug), false, `${slug} left the catalog`);
        assert.equal(categoryVerdict(slug, LIVE_ROWS), "unknown", slug);
    }
});

test("a category an admin added after the build still renders", () => {
    const slug = "uj-kategoria";
    assert.equal(isCatalogCategory(slug), false, "not in the compiled catalog");
    assert.equal(
        categoryVerdict(slug, [...LIVE_ROWS, { slug, name: "Új kategória" }]),
        "known",
    );
});

test("an unreachable or empty backend never 404s a slug", () => {
    for (const rows of [null, undefined, [], "boom", {}]) {
        assert.equal(isLiveCategory("vendeglo", rows), true, String(rows));
        assert.equal(categoryVerdict("vendeglo", rows), "known", String(rows));
    }
});

test("an empty slug is unknown", () => {
    assert.equal(categoryVerdict("", LIVE_ROWS), "unknown");
    assert.equal(categoryVerdict("   ", LIVE_ROWS), "unknown");
    assert.equal(categoryVerdict(null, LIVE_ROWS), "unknown");
});

test("slug matching ignores case and surrounding space", () => {
    assert.equal(normalizeCategorySlug("  ÉTKEZES "), "étkezes");
    assert.equal(isCatalogCategory(" Etterem "), true);
    assert.equal(isLiveCategory("ETTEREM", LIVE_ROWS), true);
});

test("the route checks the guard in the browser only", () => {
    const source = readFileSync(new URL(`../${ROUTE}`, import.meta.url), "utf8");
    assert.match(source, /categoryVerdict/, "route must use the guard");
    assert.match(source, /error\(404/, "route must raise a 404");
    assert.match(source, /if \(!browser\) return \{\};/, "guard must be browser-only");
});

test("the route does not fetch during prerender", () => {
    const source = readFileSync(new URL(`../${ROUTE}`, import.meta.url), "utf8");
    const browserGate = source.indexOf("if (!browser) return {};");
    const firstFetch = source.indexOf("fetch(");
    assert.ok(browserGate > -1, "browser gate is missing");
    assert.ok(
        firstFetch === -1 || firstFetch > browserGate,
        "a fetch before the browser gate puts the backend back on the build path",
    );
});

test("the error page is not indexable", () => {
    const source = readFileSync(new URL(`../${ERROR_PAGE}`, import.meta.url), "utf8");
    assert.match(source, /name="robots"[^>]*content="noindex"/);
});
