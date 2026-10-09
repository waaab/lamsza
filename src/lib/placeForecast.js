import { apiFetch } from "$lib/api";

/**
 * A settlement's or attraction's forecast (/api/weather/forecast, the same
 * answer the /idojaras pages use; the server only reads its cache, R14).
 * The page's header symbol and its weather card ask for the same place, so
 * one request serves both: the promise is kept per slug for the page's life.
 */
/** @type {Map<string, Promise<any>>} */
const pending = new Map();

/** @param {string} slug */
export function loadPlaceForecast(slug) {
    if (!slug) return Promise.resolve(null);
    let p = pending.get(slug);
    if (!p) {
        p = apiFetch(`/api/weather/forecast?slug=${encodeURIComponent(slug)}`).catch(() => {
            pending.delete(slug); // a later visit may try again
            return null;
        });
        pending.set(slug, p);
        // Fresh enough for one page view; a later visit fetches again.
        setTimeout(() => pending.delete(slug), 10 * 60 * 1000);
    }
    return p;
}
