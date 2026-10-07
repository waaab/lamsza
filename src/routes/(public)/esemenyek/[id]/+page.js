import { error } from "@sveltejs/kit";

/** Client-only, like szekek/[slug]: the API is reached via same-origin `/api`. */
export const ssr = false;

/**
 * Load the event here, not in the page, so an unknown id is a real 404 on the
 * network's error page (ErrorShell, noindex) instead of an in-page message on a
 * 200 page (verification review, 2026-10-07).
 */
export async function load({ params, fetch }) {
    const id = Number(params.id);
    if (!Number.isInteger(id) || id <= 0) {
        error(404, "Az esemény nem található.");
    }
    const res = await fetch(`/api/events/detail?id=${id}`, { credentials: "include" });
    if (res.status === 404) {
        error(404, "Az esemény nem található.");
    }
    if (!res.ok) {
        error(502, "Nem sikerült betölteni az eseményt.");
    }
    return { id, event: await res.json() };
}
