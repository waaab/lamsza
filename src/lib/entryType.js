/** Canonical `entry_types.name` values (admin catalog). */
export const ENTRY_TYPE_SERVICE = "Szolgáltatás";
export const ENTRY_TYPE_CEG = "Cég";
export const ENTRY_TYPE_EGYEB = "Egyéb";

/** @param {unknown} raw */
function foldType(raw) {
    return String(raw ?? "")
        .trim()
        .toLowerCase()
        .normalize("NFD")
        .replace(/\p{M}/gu, "");
}

/**
 * Map stored / legacy `entries.type` strings onto catalog labels.
 * Unknown values are returned trimmed as-is.
 *
 * @param {unknown} raw
 * @returns {string}
 */
export function canonicalEntryType(raw) {
    const folded = foldType(raw);
    if (!folded) return "";
    if (folded === "service" || folded === "szolgaltatas") {
        return ENTRY_TYPE_SERVICE;
    }
    if (folded === "ceg" || folded === "company") {
        return ENTRY_TYPE_CEG;
    }
    if (folded === "entry" || folded === "egyeb" || folded === "other") {
        return ENTRY_TYPE_EGYEB;
    }
    const trimmed = String(raw ?? "").trim();
    return trimmed;
}

/** @param {unknown} raw */
export function canonicalEntryTypeKey(raw) {
    const label = canonicalEntryType(raw);
    return label ? foldType(label) : "";
}

/**
 * @param {{ type?: string } | null | undefined} entry
 */
export function isServiceEntry(entry) {
    return canonicalEntryType(entry?.type) === ENTRY_TYPE_SERVICE;
}

/**
 * @template {{ type?: string }} T
 * @param {T[] | null | undefined} entries
 * @returns {T[]}
 */
export function filterServiceEntries(entries) {
    if (!Array.isArray(entries)) return [];
    return entries.filter(isServiceEntry);
}
