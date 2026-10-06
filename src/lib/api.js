/**
 * API origin for fetches.
 * - In the browser, use same-origin paths (`/api/...`) so Vite’s dev proxy (see vite.config.js)
 *   and production reverse proxies can forward to the Go backend. A hard-coded
 *   `http://localhost:3000` bypasses the proxy and breaks `npm run dev` when the
 *   backend is only reachable via the proxy or another port.
 * - During SSR/prerender (no `window`), fall back to a direct backend URL. Prefer the
 *   server-only `API_BASE_URL`, then build-time `VITE_API_BASE_URL`, then local Go.
 *
 * The browser branch comes first on purpose. `VITE_API_BASE_URL` is baked into the
 * bundle at build time, so while it held a dev value every visitor's browser tried to
 * reach `http://localhost:3001`. It is an SSR/prerender fallback only, and must never
 * hold a dev URL in a production build.
 */
export function getApiBase() {
    /** Same tab as the Svelte app - absolute origin so fetches always resolve (Vite proxy / reverse proxy). */
    if (typeof window !== "undefined" && window.location?.origin) {
        return window.location.origin;
    }
    /** Server-only name, so it is never inlined into the client bundle. */
    const serverEnv = typeof process !== "undefined" ? process.env?.API_BASE_URL : undefined;
    if (serverEnv) return String(serverEnv).replace(/\/$/, "");
    const env = import.meta.env.VITE_API_BASE_URL;
    if (env) return String(env).replace(/\/$/, "");
    return "http://127.0.0.1:3001";
}

/**
 * Interpret an API body. A 2xx response with an empty body is success (`null`).
 * Several account mutations reply with `200` and no JSON; parsing that as JSON
 * made a successful save or delete look like "A mentés nem sikerült".
 * @param {boolean} ok
 * @param {number} status
 * @param {string} text
 */
export function parseApiPayload(ok, status, text) {
    if (!ok) {
        throw new Error(String(text || "").trim() || `API Error: ${status}`);
    }
    const body = String(text ?? "");
    if (!body.trim()) return null;
    return JSON.parse(body);
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
        const response = await fetch(url, { credentials: "include", ...options });
        const text = await response.text();
        return parseApiPayload(response.ok, response.status, text);
    } catch (error) {
        console.error(`Fetch error for ${url}:`, error);
        throw error;
    }
}

/**
 * Same-origin API fetch that returns the raw Response (for admin calls that
 * inspect status / text themselves). Always sends the session cookie.
 * @param {string} path
 * @param {RequestInit} [options]
 */
export function apiCall(path, options = {}) {
    const url = path.startsWith("http") ? path : `${getApiBase()}${path}`;
    return fetch(url, { credentials: "include", ...options });
}

/**
 * Proxy fetch for external resources to bypass CORS
 * @param {string} targetUrl - The external URL to proxy
 * @returns {Promise<any>}
 */
export function proxyFetch(targetUrl) {
    return apiFetch(`/api/proxy?url=${encodeURIComponent(targetUrl)}`);
}
