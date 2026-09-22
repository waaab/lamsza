/** Empty public-field placeholder (same glyph as admin tables). */
export const EMPTY_PLACEHOLDER = "—";

/** @param {unknown} value */
export function displayText(value) {
    if (value == null) return EMPTY_PLACEHOLDER;
    if (Array.isArray(value)) {
        const parts = value.map((x) => String(x).trim()).filter(Boolean);
        return parts.length ? parts.join(", ") : EMPTY_PLACEHOLDER;
    }
    const s = String(value).trim();
    return s || EMPTY_PLACEHOLDER;
}

/** @param {unknown} value */
export function hasDisplayText(value) {
    return displayText(value) !== EMPTY_PLACEHOLDER;
}

/** Two-letter initials from a display name; skips punctuation-only tokens. */
export function displayInitials(name) {
    const parts = String(name ?? "")
        .trim()
        .split(/\s+/)
        .filter((p) => /[\p{L}\p{N}]/u.test(p));
    if (!parts.length) return "?";
    if (parts.length === 1) {
        return parts[0].slice(0, 2).toLocaleUpperCase("hu-HU");
    }
    return (parts[0][0] + parts[1][0]).toLocaleUpperCase("hu-HU");
}
