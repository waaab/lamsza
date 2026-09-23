/**
 * Homepage weather place.
 * The signed-in user's saved settlement wins. Otherwise the admin site default is used.
 * Favorite settlements are a saved list of places and do not choose this place.
 *
 * @param {{ slug?: string, name?: string }} siteDefault
 * @param {{ slug?: string, name?: string } | null | undefined} [preferred]
 */
export function homepageSettlements(siteDefault, preferred) {
    const preferredSlug = String(preferred?.slug ?? "").trim();
    if (preferredSlug) {
        return [{ slug: preferredSlug, name: preferred?.name || "" }];
    }
    const slug = String(siteDefault?.slug ?? "").trim();
    if (!slug) return [];
    return [{ slug, name: siteDefault?.name || "" }];
}

/**
 * Homepage events place.
 * A signed-in user's saved settlement filters the ticker.
 * With no saved settlement, including a signed-out visit, the ticker lists every location.
 *
 * @param {{ slug?: string, name?: string } | null | undefined} preferred
 * @returns {{ slug: string, name: string } | null}
 */
export function homepageEventPlace(preferred) {
    const slug = String(preferred?.slug ?? "").trim();
    if (!slug) return null;
    return { slug, name: preferred?.name || "" };
}

/** @param {{ attractions?: { slug: string, name?: string, latitude?: number | null, longitude?: number | null }[] }} favorites */
export function homepageAttractionWeather(favorites) {
    const attractions = favorites?.attractions;
    if (!Array.isArray(attractions)) return [];
    return attractions
        .filter((a) => Number.isFinite(a.latitude) && Number.isFinite(a.longitude))
        .map((a) => ({
            slug: a.slug,
            name: a.name || "",
            lat: a.latitude,
            lon: a.longitude,
        }));
}
