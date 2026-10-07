import { get, writable } from "svelte/store";

/**
 * Cookie / local-storage consent.
 *
 * The categories below describe what Lámsza actually stores today. Keep them in
 * sync with /iranyelvek/sutik and with the storage keys listed in
 * CATEGORY_STORAGE_KEYS — the banner is only honest if those lists match.
 *
 * Essential storage is never gated:
 *   - lamsza_session   HttpOnly sign-in cookie set by the backend
 *   - lamsza_auth_session  sessionStorage copy of the signed-in user
 *   - theme            display mode the visitor picked themselves
 *   - user_quick_links / quick_links_display_count  links the visitor configured
 *   - lamsza_cookie_consent  this record
 */
export const CONSENT_STORAGE_KEY = "lamsza_cookie_consent";

/** Bump when the categories change; stored choices then become undecided again. */
export const CONSENT_VERSION = 1;

/** Categories the visitor can switch off. */
export const OPTIONAL_CATEGORIES = /** @type {const} */ (["kenyelmi", "gyorsitotar"]);

/** Exact keys written by each optional category, removed when it is switched off. */
const CATEGORY_STORAGE_KEYS = {
    kenyelmi: ["lamsza_entry_history"],
    gyorsitotar: ["hirek_cache", "promoted_links_cache"],
};

/** Key prefixes written by each optional category (per-settlement caches). */
const CATEGORY_STORAGE_PREFIXES = {
    kenyelmi: [],
    gyorsitotar: ["news_cache", "weather_cache_"],
};

const UNDECIDED = {
    decided: false,
    version: CONSENT_VERSION,
    updatedAt: 0,
    kenyelmi: false,
    gyorsitotar: false,
};

function normalize(raw) {
    if (!raw || typeof raw !== "object") return { ...UNDECIDED };
    if (raw.version !== CONSENT_VERSION) return { ...UNDECIDED };
    return {
        decided: raw.decided === true,
        version: CONSENT_VERSION,
        updatedAt: Number(raw.updatedAt) || 0,
        kenyelmi: raw.kenyelmi === true,
        gyorsitotar: raw.gyorsitotar === true,
    };
}

function readStored() {
    if (typeof localStorage === "undefined") return { ...UNDECIDED };
    try {
        return normalize(JSON.parse(localStorage.getItem(CONSENT_STORAGE_KEY)));
    } catch {
        return { ...UNDECIDED };
    }
}

/** Current choice. `decided` is false until the visitor answers the banner. */
export const consent = writable({ ...UNDECIDED });

/** True once the stored choice has been read in the browser. */
export const consentReady = writable(false);

/** Drives the settings panel; the footer link and the banner both set it. */
export const consentPanelOpen = writable(false);

let loaded = false;

/**
 * Read the stored choice once. Safe to call from anywhere, including before the
 * layout mounts — storageAllowed() relies on that so gated callers never read a
 * stale default just because they ran first.
 */
export function initConsent() {
    if (loaded || typeof localStorage === "undefined") return get(consent);
    return reloadConsent();
}

/** Re-read the stored choice, e.g. after another tab changed it. */
export function reloadConsent() {
    if (typeof localStorage === "undefined") return get(consent);
    const stored = readStored();
    loaded = true;
    consent.set(stored);
    consentReady.set(true);
    return stored;
}

/**
 * May this category write to the browser? Unknown (essential) categories always may.
 * @param {string} category
 */
export function storageAllowed(category) {
    if (!OPTIONAL_CATEGORIES.includes(/** @type {any} */ (category))) return true;
    const state = loaded ? get(consent) : initConsent();
    return state.decided === true && state[category] === true;
}

function clearCategoryData(category) {
    if (typeof localStorage === "undefined") return;
    const exact = CATEGORY_STORAGE_KEYS[category] ?? [];
    const prefixes = CATEGORY_STORAGE_PREFIXES[category] ?? [];
    try {
        const doomed = new Set(exact);
        for (let i = 0; i < localStorage.length; i += 1) {
            const key = localStorage.key(i);
            if (key && prefixes.some((p) => key.startsWith(p))) doomed.add(key);
        }
        for (const key of doomed) localStorage.removeItem(key);
    } catch {
        /* quota / private mode */
    }
}

/**
 * Record a decision. Categories left out keep their current value.
 * @param {Partial<Record<(typeof OPTIONAL_CATEGORIES)[number], boolean>>} choices
 */
export function saveConsent(choices = {}) {
    const previous = loaded ? get(consent) : initConsent();
    const next = {
        decided: true,
        version: CONSENT_VERSION,
        updatedAt: Date.now(),
        kenyelmi: choices.kenyelmi ?? previous.kenyelmi,
        gyorsitotar: choices.gyorsitotar ?? previous.gyorsitotar,
    };
    loaded = true;
    consent.set(next);
    consentReady.set(true);
    if (typeof localStorage !== "undefined") {
        try {
            localStorage.setItem(CONSENT_STORAGE_KEY, JSON.stringify(next));
        } catch {
            /* quota / private mode: the choice then only holds for this page view */
        }
    }
    for (const category of OPTIONAL_CATEGORIES) {
        if (!next[category]) clearCategoryData(category);
    }
    return next;
}

export function acceptAllConsent() {
    return saveConsent({ kenyelmi: true, gyorsitotar: true });
}

export function rejectOptionalConsent() {
    return saveConsent({ kenyelmi: false, gyorsitotar: false });
}

export function openConsentSettings() {
    initConsent();
    consentPanelOpen.set(true);
}

export function closeConsentSettings() {
    consentPanelOpen.set(false);
}
