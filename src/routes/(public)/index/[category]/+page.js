import { browser } from "$app/environment";
import { error } from "@sveltejs/kit";

import { directoryPrerenderSlugs } from "$lib/entryCategory.js";
import {
    categoryVerdict,
    isCatalogCategory,
    normalizeCategorySlug,
} from "$lib/categoryRouteGuard.js";

export const prerender = true;

/**
 * One prerendered page per catalog category, parents and children alike.
 * Never hand-write this list: the old hand-written one drifted and kept
 * shipping dead slugs (egeszsegugy, egyeb, vendeglo, bolt) for months.
 * @type {import('./$types').EntryGenerator}
 */
export function entries() {
    return directoryPrerenderSlugs().map((slug) => ({ category: slug }));
}

/**
 * Refuse a slug that is not a category.
 *
 * Dropping a slug from `entries()` deletes the prerendered file, not the URL:
 * nginx serves `app.html` for anything missing on disk, so `/index/vendeglo`
 * still rendered an empty category. This turns it into the error page instead.
 *
 * Runs in the browser only. During prerender every slug comes from `entries()`,
 * so there is nothing to check, and a fetch here would put the backend back on
 * the build's critical path.
 *
 * @type {import('./$types').PageLoad}
 */
export async function load({ params, fetch }) {
    if (!browser) return {};

    const slug = normalizeCategorySlug(params.category);
    /** Compiled-in catalog first: covers every prerendered page at no cost. */
    if (isCatalogCategory(slug)) return {};

    /** Unknown to this bundle - it may still be a category an admin added since. */
    const liveRows = await fetchCategoryRows(fetch);
    if (categoryVerdict(slug, liveRows) === "known") return {};

    throw error(404, `Nincs ilyen kategória: ${slug}`);
}

/**
 * @param {typeof globalThis.fetch} fetch
 * @returns {Promise<unknown>} rows, or null when the backend did not answer
 */
async function fetchCategoryRows(fetch) {
    try {
        const res = await fetch("/api/entry-categories");
        if (!res.ok) return null;
        return await res.json();
    } catch {
        return null;
    }
}
