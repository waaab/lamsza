import { error } from "@sveltejs/kit";

/** Client-only: API is reached via same-origin `/api` (Vite proxy in dev). */
export const ssr = false;

export async function load({ params, fetch }) {
    const slug = params.slug?.toLowerCase()?.trim();
    if (!slug) {
        throw error(404, "Not found");
    }

    const res = await fetch(
        `/api/historical_seats?slug=${encodeURIComponent(slug)}`,
    );

    if (res.status === 404) {
        return { seat: null, notFound: true };
    }

    if (!res.ok) {
        throw error(502, "Nem sikerült betölteni a szék adatait.");
    }

    const seat = await res.json();
    if (!seat?.id) {
        return { seat: null, notFound: true };
    }

    return { seat, notFound: false };
}
