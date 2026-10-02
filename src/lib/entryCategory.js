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

/** @type {Map<string, string>} child slug -> parent slug */
const CATEGORY_PARENT_SLUG = new Map(
    DIRECTORY_TREE.flatMap(([parent, children]) => {
        const parentSlug = foldCategory(parent);
        return children.map((child) => [foldCategory(child), parentSlug]);
    }),
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
