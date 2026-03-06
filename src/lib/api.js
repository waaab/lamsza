const apiBase = import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

/**
 * Enhanced fetch wrapper for the Lamsza API
 * @param {string} endpoint - The relative endpoint (e.g., '/api/directory')
 * @param {RequestInit} options - Standard fetch options
 * @returns {Promise<any>}
 */
export async function apiFetch(endpoint, options = {}) {
    const url = endpoint.startsWith('http') ? endpoint : `${apiBase}${endpoint}`;

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
 * @param {string} targetUrl - The external URL to fetch
 * @returns {Promise<any>}
 */
export async function proxyFetch(targetUrl) {
    return apiFetch(`/api/proxy?url=${encodeURIComponent(targetUrl)}`);
}
