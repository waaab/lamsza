import { directoryPrerenderSlugs } from '$lib/entryCategory.js';

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
