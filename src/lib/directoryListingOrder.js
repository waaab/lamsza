/**
 * Order for the public directory index.
 *
 * Name order ignores surrounding spaces and uses Hungarian collation, so a
 * stored name like " Imperial" sorts with I.
 *
 * With no place selected, Név (A→Z) orders the whole list.
 * With a place selected, listings are ordered by distance from that place,
 * and Név (A→Z) or Legújabb still orders the listings inside each band:
 *   0 the settlement itself
 *   1 its outer range (child villages, or the parent town and its other villages)
 *   2 the rest of that county
 *   3 every other county
 */

/** @param {Record<string, any> | null | undefined} entry */
export function entrySortName(entry) {
    return String(entry?.name ?? "").trim();
}

/**
 * @param {Record<string, any>} a
 * @param {Record<string, any>} b
 */
export function compareEntryNames(a, b) {
    const byName = entrySortName(a).localeCompare(entrySortName(b), "hu");
    if (byName !== 0) return byName;
    return Number(a?.id) - Number(b?.id);
}

/**
 * @param {Record<string, any>} a
 * @param {Record<string, any>} b
 */
export function compareEntryNewest(a, b) {
    const byId = Number(b?.id) - Number(a?.id);
    if (byId !== 0) return byId;
    return compareEntryNames(a, b);
}

/**
 * The signed-in user's saved settlement, or null.
 * The homepage search field does not select it. The location menu lists it first.
 * Search and index filters do not write this.
 *
 * @param {{ slug?: string, name?: string, county_slug?: string, id?: number | null } | null | undefined} preferred
 * @param {boolean} loggedIn
 */
export function searchPreferredLocation(preferred, loggedIn) {
    if (!loggedIn) return null;
    const slug = String(preferred?.slug ?? "").trim();
    if (!slug) return null;
    return {
        slug,
        name: String(preferred?.name ?? "").trim() || slug,
        county_slug: String(preferred?.county_slug ?? "").trim(),
        id: preferred?.id ?? null,
    };
}

/**
 * Towns for the location menu, A–Z, without counties and without the saved place.
 * The saved place is rendered first by the menu itself and is not selected.
 *
 * @param {Array<Record<string, any>> | null | undefined} locations
 * @param {{ slug?: string } | null | undefined} preferred
 */
export function locationMenuTowns(locations, preferred) {
    const preferredSlug = String(preferred?.slug ?? "").trim();
    return (locations || [])
        .filter((loc) => String(loc?.type || "") !== "megye" && String(loc?.slug || "").trim())
        .filter((loc) => loc.slug !== preferredSlug)
        .slice()
        .sort((a, b) => String(a.name || "").localeCompare(String(b.name || ""), "hu"));
}

/**
 * The place the index location button names.
 * A signed-in user's saved place wins; otherwise the site location.
 *
 * @param {{ slug?: string, name?: string, county_slug?: string, id?: number | null } | null | undefined} siteLocation
 * @param {{ slug?: string, name?: string, county_slug?: string, id?: number | null } | null | undefined} preferred
 * @param {boolean} loggedIn
 */
export function listingAnchor(siteLocation, preferred, loggedIn) {
    const chosen = searchPreferredLocation(preferred, loggedIn);
    if (chosen) return chosen;
    const slug = String(siteLocation?.slug ?? "").trim();
    if (!slug) return null;
    return {
        slug,
        name: String(siteLocation?.name ?? "").trim() || slug,
        county_slug: String(siteLocation?.county_slug ?? "").trim(),
        id: siteLocation?.id ?? null,
    };
}

/**
 * @param {Array<{ id?: number, slug?: string, type?: string, parent_id?: number | null, county_slug?: string }> | null | undefined} locations
 */
function settlementsBySlug(locations) {
    /** @type {Map<string, { id: number, slug: string, parent_id: number | null, county_slug: string }>} */
    const map = new Map();
    for (const loc of locations || []) {
        if (String(loc?.type ?? "") === "megye") continue;
        const slug = String(loc?.slug ?? "").trim();
        if (!slug || map.has(slug)) continue;
        const parent = loc?.parent_id;
        map.set(slug, {
            id: Number(loc?.id),
            slug,
            parent_id: parent == null || parent === "" ? null : Number(parent),
            county_slug: String(loc?.county_slug ?? "").trim(),
        });
    }
    return map;
}

/**
 * @param {Record<string, any>} entry
 * @param {{ slug: string, county_slug?: string, id?: number | null }} anchor
 * @param {Array<{ id?: number, slug?: string, type?: string, parent_id?: number | null, county_slug?: string }>} locations
 */
export function locationRing(entry, anchor, locations) {
    const slug = String(entry?.location_slug ?? "").trim();
    const anchorSlug = String(anchor?.slug ?? "").trim();
    if (!anchorSlug || slug === anchorSlug) return 0;

    const bySlug = settlementsBySlug(locations);
    const anchorLoc = bySlug.get(anchorSlug);
    const entryLoc = bySlug.get(slug);
    const anchorId = anchorLoc?.id ?? (anchor?.id == null ? null : Number(anchor.id));
    const entryParent = entryLoc?.parent_id ?? null;
    const anchorParent = anchorLoc?.parent_id ?? null;

    if (anchorId != null && entryParent != null && entryParent === anchorId) return 1;
    if (anchorParent != null && entryLoc && entryLoc.id === anchorParent) return 1;
    if (anchorParent != null && entryParent != null && entryParent === anchorParent) return 1;

    const entryCounty = String(entry?.county_slug ?? entryLoc?.county_slug ?? "").trim();
    const anchorCounty = String(anchor?.county_slug ?? anchorLoc?.county_slug ?? "").trim();
    if (entryCounty && anchorCounty && entryCounty === anchorCounty) return 2;
    return 3;
}

/**
 * @param {Array<Record<string, any>>} entries
 * @param {{
 *   sortMode?: string,
 *   location?: { slug: string, county_slug?: string, id?: number | null } | null,
 *   locations?: Array<{ id?: number, slug?: string, type?: string, parent_id?: number | null, county_slug?: string }>,
 * }} [options]
 */
/**
 * Same name and newest order as directory listings. Websites have no place,
 * so the location button does not apply.
 *
 * @param {Array<Record<string, any>>} websites
 * @param {string} [sortMode]
 */
export function sortWebsites(websites, sortMode = "title") {
    const newest = sortMode === "newest";
    const list = Array.isArray(websites) ? [...websites] : [];
    list.sort((a, b) => {
        if (newest) {
            const byId = Number(b?.id) - Number(a?.id);
            if (byId !== 0) return byId;
        }
        const byTitle = String(a?.title ?? "")
            .trim()
            .localeCompare(String(b?.title ?? "").trim(), "hu");
        if (byTitle !== 0) return byTitle;
        return Number(a?.id) - Number(b?.id);
    });
    return list;
}

export function sortDirectoryEntries(entries, options = {}) {
    const sortMode = options.sortMode === "newest" ? "newest" : "title";
    const location = options.location?.slug ? options.location : null;
    const locations = options.locations || [];
    const list = Array.isArray(entries) ? [...entries] : [];
    list.sort((a, b) => {
        if (location) {
            const byRing = locationRing(a, location, locations) - locationRing(b, location, locations);
            if (byRing !== 0) return byRing;
        }
        return sortMode === "newest" ? compareEntryNewest(a, b) : compareEntryNames(a, b);
    });
    return list;
}
