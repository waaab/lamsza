import {
    canonicalEntryType,
    canonicalEntryTypeKey,
} from "./entryType.js";

/** @param {unknown} raw */
export function normalizeTagKey(raw) {
    const s = String(raw ?? "").trim().toLowerCase();
    if (!s) return "";
    return s.startsWith("#") ? s.slice(1) : s;
}

/** @param {unknown} raw */
export function displayTagLabel(raw) {
    const s = String(raw ?? "").trim();
    if (!s) return "";
    return s.startsWith("#") ? s : "#" + s;
}

/**
 * Aggregates entry types and tag counts for the index tag cloud.
 * DB stores tag names **without** `#`; admin and public UI show them hashtag-style (#cimke).
 *
 * @param {Array<{ type?: string, tags?: string[] | null }>} entries
 * @returns {{ typeItems: { key: string, label: string, count: number }[], tagItems: { key: string, label: string, count: number }[] }}
 */
export function buildDirectoryTagCloud(entries) {
    /** @type {Map<string, number>} */
    const typeMap = new Map();
    /** @type {Map<string, string>} */
    const typeLabel = new Map();
    /** @type {Map<string, number>} */
    const tagMap = new Map();
    /** @type {Map<string, string>} */
    const tagFirstRaw = new Map();

    for (const e of entries) {
        const label = canonicalEntryType(e.type);
        if (label) {
            const k = canonicalEntryTypeKey(label);
            typeMap.set(k, (typeMap.get(k) || 0) + 1);
            if (!typeLabel.has(k)) typeLabel.set(k, label);
        }
        const tagList = Array.isArray(e?.tags) ? e.tags : [];
        for (const raw of tagList) {
            const key = normalizeTagKey(raw);
            if (!key) continue;
            if (!tagFirstRaw.has(key)) tagFirstRaw.set(key, String(raw).trim());
            tagMap.set(key, (tagMap.get(key) || 0) + 1);
        }
    }

    const typeItems = Array.from(typeMap.entries())
        .map(([k, count]) => {
            return { key: k, label: typeLabel.get(k) || k, count };
        })
        .sort((a, b) =>
            b.count !== a.count
                ? b.count - a.count
                : a.label.localeCompare(b.label, "hu"),
        );

    const tagItems = Array.from(tagMap.entries())
        .map(([key, count]) => ({
            key,
            label: displayTagLabel(tagFirstRaw.get(key) || key),
            count,
        }))
        .sort((a, b) =>
            b.count !== a.count
                ? b.count - a.count
                : a.label.localeCompare(b.label, "hu"),
        );

    return { typeItems, tagItems };
}

/**
 * @param {unknown} entry
 * @param {string | null} selectedTypeKey
 * @param {string | null} selectedTagKey
 */
export function entryMatchesAsideFilters(entry, selectedTypeKey, selectedTagKey) {
    if (selectedTypeKey) {
        const t = canonicalEntryTypeKey(entry?.type);
        if (t !== selectedTypeKey) return false;
    }
    if (selectedTagKey) {
        const tags = Array.isArray(entry?.tags) ? entry.tags : [];
        let ok = false;
        for (const raw of tags) {
            if (normalizeTagKey(raw) === selectedTagKey) {
                ok = true;
                break;
            }
        }
        if (!ok) return false;
    }
    return true;
}
