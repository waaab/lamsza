import { error } from "@sveltejs/kit";

/** Client-only: API is reached via same-origin `/api` (Vite proxy in dev). */
export const ssr = false;

export async function load({ params, fetch }) {
    const slug = params.slug?.toLowerCase()?.trim();
    if (!slug) {
        error(404, "A szék nem található.");
    }

    const res = await fetch(
        `/api/historical_seats?slug=${encodeURIComponent(slug)}`,
    );

    // An unknown seat is a real 404 on the network's error page, not an
    // in-page message on a 200 page (verification review, 2026-10-07).
    if (res.status === 404) {
        error(404, "A szék nem található.");
    }

    if (!res.ok) {
        error(502, "Nem sikerült betölteni a szék adatait.");
    }

    const seat = await res.json();
    if (!seat?.id) {
        error(404, "A szék nem található.");
    }

    return { seat };
}
