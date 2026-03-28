/**
 * API origin for fetches.
 * - In the browser, use same-origin paths (`/api/...`) so Vite’s dev proxy (see vite.config.js)
 *   and production reverse proxies can forward to the Go backend. A hard-coded
 *   `http://localhost:3000` bypasses the proxy and breaks `npm run dev` when the
 *   backend is only reachable via the proxy or another port.
 * - During SSR/prerender (no `window`), fall back to a direct backend URL unless
 *   `VITE_API_BASE_URL` is set.
 */
export function getApiBase() {
    const env = import.meta.env.VITE_API_BASE_URL;
    if (env) return String(env).replace(/\/$/, "");
    if (typeof window !== "undefined") return "";
    return "http://localhost:3000";
}

/**
 * Enhanced fetch wrapper for the Lamsza API
 * @param {string} endpoint - The relative endpoint (e.g. '/api/directory')
 * @param {RequestInit} options - Standard fetch options
 * @returns {Promise<any>}
 */
export async function apiFetch(endpoint, options = {}) {
    const base = getApiBase();
    const url = endpoint.startsWith("http")
        ? endpoint
        : `${base}${endpoint}`;

    try {
        const response = await fetch(url, options);
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || `API Error: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error(`Fetch error for ${url}:`, error);
        throw error;
    }
}

/**
 * Proxy fetch for external resources to bypass CORS
 * @param {string} targetUrl - The external URL to proxy
 * @returns {Promise<any>}
 */
export function proxyFetch(targetUrl) {
    return apiFetch(`/api/proxy?url=${encodeURIComponent(targetUrl)}`);
}
