import { error } from "@sveltejs/kit";

/** Client-only, like szekek/[slug]: the API is reached via same-origin `/api`. */
export const ssr = false;

/**
 * Load the entry here, not in onMount, so an unknown slug is a real 404 on the
 * network's error page (ErrorShell, noindex) instead of an in-page message on a
 * 200 page (verification review, 2026-10-07).
 */
export async function load({ params, fetch }) {
    const slug = params.slug?.trim();
    if (!slug) {
        error(404, "A bejegyzés nem található.");
    }
    const res = await fetch(`/api/entry?slug=${encodeURIComponent(slug)}`, {
        credentials: "include",
    });
    if (res.status === 404) {
        error(404, "A bejegyzés nem található.");
    }
    if (!res.ok) {
        error(502, "Nem sikerült betölteni a bejegyzést.");
    }
    const entry = await res.json();
    if (!entry?.name) {
        error(404, "A bejegyzés nem található.");
    }
    return { slug, entry };
}
