/**
 * The Székelyföld pages (Székek, Megyék, Városok, Falvak; UI_BASELINE
 * "szf-pages"): their navigation, the colour each seat and county has on
 * every one of them, and the honeycomb map's layout.
 */

/** The four pages, in the toolbar's order, with the toolbar's icons. */
export const REGION_PAGES = [
    { key: "szekek", href: "/szekek", label: "Történelmi székek", icon: "szekek" },
    { key: "megyek", href: "/megyek", label: "Székelyföldi megyék", icon: "counties" },
    { key: "varosok", href: "/varosok", label: "Székelyföldi városok", icon: "varosok" },
    { key: "falvak", href: "/falvak", label: "Székelyföldi falvak", icon: "falvak" },
];

const RED = "#c8102e";
const AMBER = "#d4a017";
const TEAL = "#2f7d6d";
const BLUE = "#3a6fbf";
const PURPLE = "#7a4ea3";

/**
 * Seats and counties roughly as they lie on the map, north-east first, each
 * with its colour. A county keeps its colour on every page (the Falvak chips
 * and tiles use it too).
 */
export const SEAT_ORDER = [
    ["csikszek", RED],
    ["haromszek", AMBER],
    ["aranyosszek", TEAL],
    ["marosszek", BLUE],
    ["udvarhelyszek", PURPLE],
];
export const COUNTY_ORDER = [
    ["hargita", RED],
    ["kovaszna", AMBER],
    ["maros", BLUE],
];
const FALLBACK = [RED, AMBER, TEAL, BLUE, PURPLE];

/**
 * Sorts items by a known order (unknown slugs after it, by name) and gives
 * each its colour.
 *
 * @template {{ slug: string, name: string }} T
 * @param {T[]} items
 * @param {Array<[string, string]>} order
 * @returns {Array<T & { color: string }>}
 */
export function inRegionOrder(items, order) {
    const rank = new Map(order.map(([slug], i) => [slug, i]));
    return [...items]
        .sort((a, b) => (rank.get(a.slug) ?? 99) - (rank.get(b.slug) ?? 99) || a.name.localeCompare(b.name, "hu"))
        .map((item, i) => ({
            ...item,
            color: order.find(([slug]) => slug === item.slug)?.[1] ?? FALLBACK[i % FALLBACK.length],
        }));
}

/** @param {string} countySlug */
export function countyColor(countySlug) {
    return COUNTY_ORDER.find(([slug]) => slug === countySlug)?.[1] ?? "var(--text-muted)";
}

/** Hexagon radius and the honeycomb's steps, in viewBox units. */
export const HEX_R = 64;
const STEP_X = 111;
const STEP_Y = 96;
const WIDTH = 380;

/**
 * Honeycomb positions for n cells: rows of two and three in turn, each
 * centred, so 3 cells sit 2 over 1 and 5 sit 2 over 3.
 *
 * @param {number} n
 * @returns {{ cells: Array<{ x: number, y: number }>, width: number, height: number }}
 */
export function honeycomb(n) {
    const cells = [];
    let row = 0;
    while (cells.length < n) {
        const size = Math.min(row % 2 === 0 ? 2 : 3, n - cells.length);
        for (let i = 0; i < size; i++) {
            cells.push({ x: WIDTH / 2 + (i - (size - 1) / 2) * STEP_X, y: HEX_R + 16 + row * STEP_Y });
        }
        row++;
    }
    return { cells, width: WIDTH, height: Math.max(1, row) * STEP_Y + 2 * (HEX_R + 16) - STEP_Y };
}

/** The hexagon's outline around its centre. */
export const HEX_POINTS = "0,-64 55.4,-32 55.4,32 0,64 -55.4,32 -55.4,-32";
