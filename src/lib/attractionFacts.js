/**
 * An attraction's facts as the page's stat tiles (UI_BASELINE "szf-pages"):
 * only the ones that are set, in Hungarian number format ("0,22 km²").
 */
const NUMBER = new Intl.NumberFormat("hu-HU", { maximumFractionDigits: 3 });

const FACTS = [
    ["elevation_m", "m", "magasság"],
    ["area_km2", "km²", "felszín"],
    ["depth_m", "m", "mélység"],
];

/**
 * @param {{ elevation_m?: number | null, area_km2?: number | null, depth_m?: number | null } | null | undefined} attraction
 * @returns {Array<{ key: string, value: string, label: string }>}
 */
export function attractionFacts(attraction) {
    if (!attraction) return [];
    return FACTS.flatMap(([key, unit, label]) => {
        const n = /** @type {Record<string, unknown>} */ (attraction)[key];
        return typeof n === "number" && Number.isFinite(n) ? [{ key, value: `${NUMBER.format(n)} ${unit}`, label }] : [];
    });
}
