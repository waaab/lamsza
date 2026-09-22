export const KIND_LABELS = {
    sports_arena: "Sportcsarnok / pálya",
    indoor_hall: "Fedett csarnok",
    outdoor_area: "Szabadtéri terület",
    market_square: "Piac / tér",
    park: "Park",
    street: "Utca / felvonulás",
    mixed: "Több helyszín",
    temporary: "Ideiglenes",
    other: "Egyéb",
};

/**
 * Prefer label from API; fall back to static map or slug.
 * @param {string | null | undefined} kind
 * @param {string | null | undefined} labelFromApi
 */
export function kindLabel(kind, labelFromApi) {
    if (labelFromApi && String(labelFromApi).trim()) {
        return String(labelFromApi).trim();
    }
    const k = String(kind || "").trim();
    if (!k) return "";
    return KIND_LABELS[/** @type {keyof typeof KIND_LABELS} */ (k)] || k;
}

