import assert from "node:assert/strict";
import { test } from "node:test";
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

const DIST = "dist";
/** The SPA shell is not a page: it has no prerendered head. */
const NOT_A_PAGE = new Set(["app.html"]);

/** @param {string} dir @returns {string[]} */
function htmlFiles(dir) {
    /** @type {string[]} */
    const out = [];
    for (const name of readdirSync(dir)) {
        const p = join(dir, name);
        if (statSync(p).isDirectory()) out.push(...htmlFiles(p));
        else if (name.endsWith(".html") && !NOT_A_PAGE.has(name)) out.push(p);
    }
    return out;
}

/** @param {string} html @param {RegExp} re */
function count(html, re) {
    return (html.match(re) ?? []).length;
}

test("every prerendered page carries canonical and Open Graph tags", (t) => {
    if (!existsSync(DIST)) {
        t.skip("dist/ is not built - run `npm run build` first");
        return;
    }
    const pages = htmlFiles(DIST);
    assert.ok(pages.length > 0, "no prerendered pages found in dist/");

    /** @type {string[]} */
    const problems = [];
    let checked = 0;
    for (const file of pages) {
        const html = readFileSync(file, "utf8");
        // Prerendered redirects (e.g. /profil → /beallitasok) are a meta refresh,
        // not a page: a canonical on them would compete with the target page.
        if (/http-equiv="refresh"/.test(html)) continue;
        checked++;
        const canonicals = count(html, /<link[^>]+rel="canonical"/g);

        if (canonicals === 0) problems.push(`${file}: no <link rel="canonical">`);
        if (canonicals > 1) problems.push(`${file}: ${canonicals} canonical tags, expected 1`);
        if (!/<link[^>]+rel="canonical"[^>]+href="https:\/\/lamsza\.com/.test(html)) {
            problems.push(`${file}: canonical is not an absolute https://lamsza.com URL`);
        }
        for (const tag of [
            'property="og:title"',
            'property="og:description"',
            'property="og:url"',
            'property="og:image"',
            'property="og:type"',
            'property="og:site_name"',
            'property="og:locale"',
            'name="twitter:card"',
        ]) {
            if (!html.includes(tag)) problems.push(`${file}: missing ${tag}`);
        }
        if (!html.includes('content="summary_large_image"')) {
            problems.push(`${file}: twitter:card is not summary_large_image`);
        }
    }

    assert.equal(problems.join("\n"), "", `\n${problems.join("\n")}`);
    assert.ok(checked > 20, `only ${checked} real pages checked, expected the whole site`);
});

test("the default share image is shipped", (t) => {
    if (!existsSync(DIST)) {
        t.skip("dist/ is not built - run `npm run build` first");
        return;
    }
    assert.ok(existsSync(join(DIST, "og-image.png")), "dist/og-image.png is missing");
});

test("the page title is not shadowed by the SPA shell fallback title", () => {
    const shell = readFileSync("src/app.html", "utf8");
    const headStart = shell.indexOf("%sveltekit.head%");
    assert.ok(headStart > -1, "%sveltekit.head% placeholder is gone from src/app.html");
    assert.ok(
        shell.indexOf("<title>") > headStart,
        "the fallback <title> must come after %sveltekit.head%, the first title wins",
    );
    assert.ok(
        shell.indexOf('name="description"') > headStart,
        "the fallback description must come after %sveltekit.head%, the first one wins",
    );
});
