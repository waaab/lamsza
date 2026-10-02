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

/** @type {[string, string[]][]} parent name, child names */
const DIRECTORY_TREE = [
    ["Étkezés", ["Étterem", "Kávézó", "Cukrászda", "Pékség", "Söröző"]],
    ["Szállás", ["Szálloda", "Motel", "Panzió", "Apartman", "Kemping"]],
    [
        "Egészség és szépség",
        [
            "Orvos",
            "Fogászat",
            "Bőrgyógyászat",
            "Optika",
            "Csontkovács",
            "Lábgyógyászat",
            "Gyógytorna",
            "Masszázs",
            "Gyógyszertár",
            "Kórház",
            "Állatorvos",
            "Fodrász",
            "Borbély",
            "Körömszalon",
            "Spa",
        ],
    ],
    ["Vásárlás", ["Élelmiszer", "Ruházat", "Műszaki bolt", "Bútor", "Piac"]],
    [
        "Autó",
        [
            "Autószerviz",
            "Karosszéria",
            "Olajcsere",
            "Gumiszerviz",
            "Turbószerviz",
            "Autómentés",
            "Autómosó",
            "Autókozmetika",
            "Parkoló",
            "Autókereskedés",
            "Autóbontó",
            "Autóalkatrész",
            "Benzinkút",
        ],
    ],
    [
        "Mesteremberek",
        ["Villanyszerelő", "Vízvezeték-szerelő", "Asztalos", "Takarítás", "Építkezés"],
    ],
    ["Oktatás", ["Óvoda", "Iskola", "Egyetem"]],
    ["Hivatalok", ["Polgármesteri hivatal", "Megyei intézmény", "Posta"]],
    ["Sport és szabadidő", ["Sportegyesület", "Sportpálya"]],
    ["Pénzügy", ["Bank", "Biztosító"]],
];

/** @type {Array<{ id: string, name: string, slug: string, parent_id: string | null, sort_order: number }>} */
export const DIRECTORY_CATALOG = DIRECTORY_TREE.flatMap(([parentName, children], parentIndex) => {
    const parentSlug = foldCategory(parentName);
    const parent = {
        id: parentSlug,
        name: parentName,
        slug: parentSlug,
        parent_id: null,
        sort_order: parentIndex + 1,
    };
    const childRows = children.map((childName, childIndex) => ({
        id: foldCategory(childName),
        name: childName,
        slug: foldCategory(childName),
        parent_id: parentSlug,
        sort_order: childIndex + 1,
    }));
    return [parent, ...childRows];
});

/** @type {Map<string, string>} child slug -> parent slug */
const CATEGORY_PARENT_SLUG = new Map(
    DIRECTORY_CATALOG
        .filter((row) => row.parent_id)
        .map((row) => [row.slug, row.parent_id]),
);

/**
 * Return the stored category name.
 *
 * @param {unknown} raw
 * @returns {string}
 */
export function canonicalEntryCategory(raw) {
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
    const catKey = canonicalEntryCategoryKey(entry?.category);
    const filterKey = foldCategory(slug);
    if (catKey === filterKey) return true;
    return CATEGORY_PARENT_SLUG.get(catKey) === filterKey;
}

/** @param {Array<{ category?: string, category_id?: number }>} rows */
function visibleChildSlugs(rows) {
    /** @type {Set<string>} */
    const visible = new Set();
    for (const row of rows || []) {
        const slug = canonicalEntryCategoryKey(row?.category);
        if (slug) visible.add(slug);
    }
    return visible;
}

/**
 * @param {Array<{ id: string, name: string, slug: string, parent_id: string | null, sort_order: number }>} catalog
 * @param {Array<{ category?: string }>} entries
 * @param {Array<{ category?: string, category_id?: number, entry_id?: number }>} [websites]
 */
export function directoryCategoryTabs(catalog, entries, websites = []) {
    const visible = visibleChildSlugs([
        ...(entries || []),
        ...(websites || []).filter((site) => !site?.entry_id),
    ]);
    const parents = (catalog || DIRECTORY_CATALOG).filter((row) => !row.parent_id);
    const tabs = parents
        .filter((parent) => {
            const children = (catalog || DIRECTORY_CATALOG).filter(
                (row) => row.parent_id === parent.slug,
            );
            return children.some((child) => visible.has(child.slug));
        })
        .map((parent) => ({
            id: parent.slug,
            label: parent.name,
            url: `/index/${parent.slug}`,
        }))
        .sort((a, b) => a.label.localeCompare(b.label, "hu"));
    return [{ id: "osszes", label: "Összes", url: "/index" }, ...tabs];
}

/**
 * @param {string} parentSlug
 * @param {Array<{ id: string, name: string, slug: string, parent_id: string | null, sort_order: number }>} catalog
 * @param {Array<{ category?: string }>} entries
 * @param {Array<{ category?: string, category_id?: number, entry_id?: number }>} [websites]
 */
export function directoryChildTabs(parentSlug, catalog, entries, websites = []) {
    const parentKey = foldCategory(parentSlug);
    if (!parentKey || parentKey === "osszes") return [];
    const visible = visibleChildSlugs([
        ...(entries || []),
        ...(websites || []).filter((site) => !site?.entry_id),
    ]);
    return (catalog || DIRECTORY_CATALOG)
        .filter((row) => row.parent_id === parentKey && visible.has(row.slug))
        .map((child) => ({
            id: child.slug,
            label: child.name,
            url: `/index/${child.slug}`,
        }))
        .sort((a, b) => a.label.localeCompare(b.label, "hu"));
}

/** @param {{ location_slug?: string } | null | undefined} entry */
export function entryHasTown(entry) {
    return Boolean(String(entry?.location_slug || "").trim());
}
