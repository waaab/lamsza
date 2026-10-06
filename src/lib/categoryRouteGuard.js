/**
 * Decides whether `/index/<slug>` is a real category page.
 *
 * Deleting the four retired slugs (egeszsegugy, egyeb, vendeglo, bolt) from the
 * prerender list removed the files, not the URLs: production nginx falls back to
 * `app.html` for anything it cannot find on disk, so a retired slug still rendered
 * an empty category instead of an error. The page has to refuse the slug itself.
 *
 * Two lists, in this order, because they fail in opposite directions:
 *  1. `DIRECTORY_CATALOG` is compiled in. It answers instantly and covers every
 *     page we prerender, so a legitimate visit costs no extra request.
 *  2. The live `/api/entry-categories` list covers a category an admin added
 *     after this bundle was built. It is only consulted for a slug the catalog
 *     does not know, which is the rare case.
 *
 * Both checks fail open. An unreachable or empty API must never turn a real
 * category into a 404 - a wrong 404 loses a page that exists, while a wrong 200
 * only keeps a dead URL alive one more deploy.
 */

import { DIRECTORY_CATALOG, directoryPrerenderSlugs } from "./entryCategory.js";

/**
 * @param {unknown} raw
 * @returns {string}
 */
export function normalizeCategorySlug(raw) {
    return String(raw ?? "")
        .trim()
        .toLowerCase();
}

/**
 * Is the slug one of the categories this bundle was built with?
 *
 * @param {unknown} slug
 * @param {Array<{ slug: string }>} [catalog]
 * @returns {boolean}
 */
export function isCatalogCategory(slug, catalog = DIRECTORY_CATALOG) {
    const want = normalizeCategorySlug(slug);
    if (!want) return false;
    return directoryPrerenderSlugs(catalog).some(
        (row) => normalizeCategorySlug(row) === want,
    );
}

/**
 * Is the slug in the list the backend just returned?
 *
 * @param {unknown} slug
 * @param {unknown} rows rows from `/api/entry-categories`, or null when the call failed
 * @returns {boolean} true also when `rows` is unusable, so a dead API cannot 404 a live page
 */
export function isLiveCategory(slug, rows) {
    if (!Array.isArray(rows) || rows.length === 0) return true;
    const want = normalizeCategorySlug(slug);
    if (!want) return false;
    return rows.some((row) => normalizeCategorySlug(row?.slug) === want);
}

/**
 * The verdict for one visit. `"known"` renders the page, `"unknown"` is a 404.
 *
 * @param {unknown} slug
 * @param {unknown} liveRows rows from `/api/entry-categories`, or null when not consulted
 * @param {Array<{ slug: string }>} [catalog]
 * @returns {"known" | "unknown"}
 */
export function categoryVerdict(slug, liveRows = null, catalog = DIRECTORY_CATALOG) {
    if (!normalizeCategorySlug(slug)) return "unknown";
    if (isCatalogCategory(slug, catalog)) return "known";
    return isLiveCategory(slug, liveRows) ? "known" : "unknown";
}
