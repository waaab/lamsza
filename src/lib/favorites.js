import { apiFetch } from "./api.js";

export const FAVORITE_TYPES = ["settlement", "attraction", "entry", "event"];

const FAVORITE_PAYLOAD_KEYS = {
    settlement: "settlements",
    attraction: "attractions",
    entry: "entries",
    event: "events",
};

/** @param {string} type @param {number} id */
export function favoriteKey(type, id) {
    return `${type}:${id}`;
}

/** @param {{ type: string, id: number }[]} list @param {string} type @param {number} id */
export function isFavorite(list, type, id) {
    const key = favoriteKey(type, id);
    return list.some((item) => favoriteKey(item.type, item.id) === key);
}

/** @param {Record<string, unknown>} payload */
export function flattenFavoritesPayload(payload) {
    if (!payload || typeof payload !== "object") return [];
    /** @type {{ type: string, id: number }[]} */
    const out = [];
    for (const type of FAVORITE_TYPES) {
        const items = payload[FAVORITE_PAYLOAD_KEYS[type]];
        if (!Array.isArray(items)) continue;
        for (const item of items) {
            const itemId = Number(item?.id);
            if (Number.isFinite(itemId) && itemId > 0) {
                out.push({ type, id: itemId });
            }
        }
    }
    return out;
}

export async function loadFavoriteList() {
    const payload = await apiFetch("/api/account/favorites");
    return flattenFavoritesPayload(payload);
}

/** @param {string} type @param {number} id */
export async function addFavorite(type, id) {
    return apiFetch("/api/account/favorites", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ type, id }),
    });
}

/** @param {string} type @param {number} id */
export async function removeFavorite(type, id) {
    return apiFetch(
        `/api/account/favorites?type=${encodeURIComponent(type)}&id=${encodeURIComponent(String(id))}`,
        { method: "DELETE" },
    );
}

/**
 * @param {{ type: string, id: number }[]} list
 * @param {string} type
 * @param {number} id
 * @param {boolean} active
 */
export function favoriteListWithToggle(list, type, id, active) {
    const key = favoriteKey(type, id);
    if (active) {
        if (isFavorite(list, type, id)) return list;
        return [...list, { type, id }];
    }
    return list.filter((item) => favoriteKey(item.type, item.id) !== key);
}
