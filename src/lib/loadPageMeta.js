import { apiFetch } from "$lib/api.js";
import { PAGE_HEADER_FALLBACK } from "$lib/pageHeaderDefaults.js";

/** @param {unknown} s @param {string} fallback */
function nonEmptyOrFallback(s, fallback) {
    if (s == null) return fallback;
    const t = String(s).trim();
    return t !== "" ? t : fallback;
}

/**
 * Instant defaults for the hero (same keys as `pages.slug` in admin).
 * Use as initial state so the title/greeting are not "…" before the API responds.
 *
 * @param {string} slug — `pages.slug` (e.g. `index`, `home`, `iranyelvek/sutik`)
 * @returns {{ title: string, greeting: string }}
 */
export function initialPageHeader(slug) {
    const fb = PAGE_HEADER_FALLBACK[slug];
    return {
        title: fb?.title ?? "",
        greeting: fb?.greeting ?? "",
    };
}

/**
 * @param {string} slug — `pages.slug` (pl. `szekek`, `home`, `iranyelvek/sutik`)
 * @returns {Promise<{ title: string, greeting: string, content?: string, slug?: string }>}
 */
export async function loadPageMeta(slug) {
    const fb = PAGE_HEADER_FALLBACK[slug] ?? { title: "", greeting: "" };
    try {
        const page = await apiFetch(`/api/pages?slug=${encodeURIComponent(slug)}`);
        return {
            title: nonEmptyOrFallback(page?.title, fb.title),
            greeting: nonEmptyOrFallback(page?.greeting, fb.greeting),
            content: page?.content ?? "",
            slug: page?.slug,
        };
    } catch {
        return {
            title: fb.title,
            greeting: fb.greeting,
            content: "",
        };
    }
}
