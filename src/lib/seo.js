import { PAGE_HEADER_FALLBACK } from "./pageHeaderDefaults.js";
import { resolveNetworkOrigin } from "./networkOrigins.js";

export const SITE_NAME = "Lámsza";
export const SITE_LOCALE = "hu_HU";
export const DEFAULT_OG_IMAGE = "/og-image.png";
export const DEFAULT_OG_TYPE = "website";
export const DEFAULT_SEO_TITLE = "Na Lámsza! - Erdélyi magyar startlap és kereső";
export const DEFAULT_SEO_DESCRIPTION =
    "Erdélyi magyar startlap és kereső: gyors keresés, időjárás, hírek, események és gyorslinkek Székelyföldről.";

/**
 * Canonical and og:url are always absolute and always name the production apex
 * (`lamsza.com`, not `www.`) - also from dev and the `.test` network, because a
 * canonical that points at a dev host is worse than no canonical at all.
 * Empty `envOrigin` + empty `hostname` makes `resolveNetworkOrigin` return prod.
 */
export const SITE_ORIGIN = resolveNetworkOrigin("lamsza", {
    envOrigin: "",
    hostname: "",
});

/**
 * `/iranyelvek/sutik/?x=1` → `/iranyelvek/sutik` (no query, no hash, no
 * trailing slash). The root path stays `/`.
 * @param {string} pathname
 */
export function canonicalPath(pathname) {
    const raw = String(pathname || "/").trim();
    const path = raw.split("#")[0].split("?")[0];
    const withSlash = path.startsWith("/") ? path : `/${path}`;
    const trimmed = withSlash.replace(/\/+$/, "");
    return trimmed === "" ? "/" : trimmed;
}

/**
 * Absolute URL on the canonical host. Already absolute input is kept as is.
 * @param {string} pathOrUrl
 */
export function absoluteUrl(pathOrUrl) {
    const value = String(pathOrUrl || "").trim();
    if (/^https?:\/\//i.test(value)) return value;
    if (!value) return SITE_ORIGIN;
    return `${SITE_ORIGIN}${value.startsWith("/") ? value : `/${value}`}`;
}

/**
 * Path → `pages.slug` key of {@link PAGE_HEADER_FALLBACK}, so every page reuses
 * the title and greeting the hero already shows. Unknown deep paths fall back to
 * the nearest parent that has an entry (`/szekek/csikszek` → `szekek`).
 * @param {string} pathname
 * @returns {string} slug key, `home` for the root, `""` when nothing matches
 */
export function pageSlugFromPath(pathname) {
    const path = canonicalPath(pathname);
    if (path === "/") return "home";
    const segments = path.slice(1).split("/");
    for (let i = segments.length; i > 0; i--) {
        const slug = segments.slice(0, i).join("/");
        if (PAGE_HEADER_FALLBACK[slug]) return slug;
    }
    return "";
}

/**
 * Pages with no admin `pages.slug` entry, so no hero title to reuse.
 * Keyed by canonical path, same shape as {@link PAGE_HEADER_FALLBACK}.
 * @type {Record<string, { title: string, greeting: string }>}
 */
const EXTRA_PAGE_SEO = {
    "/beallitasok": {
        title: "Beállítások",
        greeting:
            "Téma, kezdőlap és értesítések - állítsd a Lámszát a saját szokásaidhoz.",
    },
    "/fiok": {
        title: "Fiók",
        greeting: "A fiókod: kedvencek, javaslatok és a saját bejegyzéseid egy helyen.",
    },
};

/** @param {string} title */
function withSiteName(title) {
    const t = String(title || "").trim();
    if (!t) return DEFAULT_SEO_TITLE;
    return t.includes(SITE_NAME) ? t : `${t} - ${SITE_NAME}`;
}

/**
 * Head-tag values for one page. Anything passed in wins over the page defaults.
 *
 * @param {{ pathname?: string, title?: string, description?: string, image?: string, type?: string }} [input]
 * @returns {{ title: string, description: string, canonical: string, image: string, type: string, siteName: string, locale: string }}
 */
export function resolveSeo(input = {}) {
    const path = canonicalPath(input.pathname ?? "/");
    const fallback =
        PAGE_HEADER_FALLBACK[pageSlugFromPath(path)] ?? EXTRA_PAGE_SEO[path];

    const title = String(input.title || "").trim() || fallback?.title || "";
    const description =
        String(input.description || "").trim() ||
        fallback?.greeting ||
        DEFAULT_SEO_DESCRIPTION;

    return {
        title: withSiteName(title),
        description,
        canonical: absoluteUrl(path),
        image: absoluteUrl(input.image || DEFAULT_OG_IMAGE),
        type: String(input.type || "").trim() || DEFAULT_OG_TYPE,
        siteName: SITE_NAME,
        locale: SITE_LOCALE,
    };
}
