/** Canonical `entry_categories.name` values. */
export const ENTRY_CATEGORY_EGESZSEGUGY = "Egészségügy";
export const ENTRY_CATEGORY_OKTATAS = "Oktatás";
export const ENTRY_CATEGORY_MESTEREMBEREK = "Mesteremberek";
export const ENTRY_CATEGORY_HIVATALOK = "Hivatalok";
export const ENTRY_CATEGORY_VENDEGLO = "Vendéglő";
export const ENTRY_CATEGORY_BOLT = "Bolt";
export const ENTRY_CATEGORY_SPORTEGYESULET = "Sportegyesület";
export const ENTRY_CATEGORY_EGYEB = "Egyéb";

/** @param {unknown} raw */
function foldCategory(raw) {
    return String(raw ?? "")
        .trim()
        .toLowerCase()
        .normalize("NFD")
        .replace(/\p{M}/gu, "")
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-|-$/g, "");
}

/**
 * Map stored / legacy category strings onto catalog labels.
 * Unknown values are returned trimmed as-is.
 *
 * @param {unknown} raw
 * @returns {string}
 */
export function canonicalEntryCategory(raw) {
    const folded = foldCategory(raw);
    if (!folded) return "";
    if (
        folded === "egeszsegugy" ||
        folded === "orvosi-rendelok" ||
        folded === "orvosi-rendelo"
    ) {
        return ENTRY_CATEGORY_EGESZSEGUGY;
    }
    if (folded === "oktatas") return ENTRY_CATEGORY_OKTATAS;
	if (folded === "mesteremberek" || folded === "mesterember") {
        return ENTRY_CATEGORY_MESTEREMBEREK;
    }
    if (folded === "hivatalok" || folded === "varoshaza") {
        return ENTRY_CATEGORY_HIVATALOK;
    }
    if (folded === "vendeglo" || folded === "vendeglatas") {
        return ENTRY_CATEGORY_VENDEGLO;
    }
    if (folded === "bolt" || folded === "kereskedelem") {
        return ENTRY_CATEGORY_BOLT;
    }
    if (folded === "sportegyesulet" || folded === "sport") {
        return ENTRY_CATEGORY_SPORTEGYESULET;
    }
    if (folded === "egyeb" || folded === "other") return ENTRY_CATEGORY_EGYEB;
    return String(raw ?? "").trim();
}

/** @param {unknown} raw */
export function canonicalEntryCategoryKey(raw) {
    const label = canonicalEntryCategory(raw);
    return label ? foldCategory(label) : "";
}

/**
 * @param {{ category?: string } | null | undefined} entry
 * @param {string} slug
 */
export function entryMatchesCategory(entry, slug) {
    if (!slug || slug === "osszes") return true;
    return canonicalEntryCategoryKey(entry?.category) === foldCategory(slug);
}

/**
 * @param {Array<{ category?: string }> | null | undefined} entries
 */
export function directoryCategoryTabs(entries) {
    /** @type {Map<string, { id: string, label: string, url: string }>} */
    const seen = new Map();
    for (const e of entries || []) {
        const label = canonicalEntryCategory(e?.category);
        if (!label) continue;
        const id = canonicalEntryCategoryKey(label);
        if (!seen.has(id)) {
            seen.set(id, { id, label, url: `/index/${id}` });
        }
    }
    const generated = [...seen.values()].sort((a, b) =>
        a.label.localeCompare(b.label, "hu"),
    );
    return [{ id: "osszes", label: "Összes", url: "/index" }, ...generated];
}
