/** Canonical `entry_types.name` values (admin catalog). */
export const ENTRY_TYPE_SZEMELY = "Személy";
export const ENTRY_TYPE_VALLALKOZAS = "Vállalkozás";
export const ENTRY_TYPE_INTEZMENY = "Intézmény";

/** @param {unknown} raw */
function foldType(raw) {
    return String(raw ?? "")
        .trim()
        .toLowerCase()
        .normalize("NFD")
        .replace(/\p{M}/gu, "");
}

/**
 * Map stored type strings onto catalog labels.
 * Unknown values return an empty string.
 *
 * @param {unknown} raw
 * @returns {string}
 */
export function canonicalEntryType(raw) {
    const trimmed = String(raw ?? "").trim();
    if (!trimmed) return "";
    if (
        trimmed === ENTRY_TYPE_SZEMELY ||
        trimmed === ENTRY_TYPE_VALLALKOZAS ||
        trimmed === ENTRY_TYPE_INTEZMENY
    ) {
        return trimmed;
    }
    return "";
}

/** @param {unknown} raw */
export function canonicalEntryTypeKey(raw) {
    const label = canonicalEntryType(raw);
    return label ? foldType(label) : "";
}
