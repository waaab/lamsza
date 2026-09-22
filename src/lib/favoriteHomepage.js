/** @param {{ settlements?: { slug: string, name?: string }[] }} favorites @param {{ slug: string, name?: string }} siteDefault */
export function homepageSettlements(favorites, siteDefault) {
    const settlements = favorites?.settlements;
    if (!Array.isArray(settlements) || settlements.length === 0) {
        if (!siteDefault?.slug) return [];
        return [{ slug: siteDefault.slug, name: siteDefault.name || "" }];
    }
    return settlements.map((s) => ({ slug: s.slug, name: s.name || "" }));
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
