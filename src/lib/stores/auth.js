import { writable } from "svelte/store";
import { getApiBase } from "$lib/api.js";
import { buildImportPayload, meToAuthState } from "$lib/accountPrefs.js";
import { readHistory } from "$lib/entryHistory.js";
import { readSlotCount } from "$lib/quickLinksDisplay.js";

const empty = {
    loggedIn: false,
    user: "",
    email: "",
    isAdmin: false,
    picture: "",
    givenName: "",
    familyName: "",
    locale: "",
    googleSub: "",
    lastLoginAt: null,
    createdAt: null,
    theme: null,
    quicklinkSlots: null,
    prefsImportedAt: null,
    adminQueueCount: 0,
};

function clearLegacyStorage() {
    if (typeof window === "undefined") return;
    localStorage.removeItem("admin_auth");
    localStorage.removeItem("admin_user");
    localStorage.removeItem("admin_is_admin");
}

function readBrowserImportInput() {
    let theme = "";
    let links = [];
    if (typeof localStorage === "undefined") {
        return { theme, slots: null, links, history: [] };
    }
    try {
        theme = localStorage.getItem("theme") || "";
    } catch {
        /* private mode / quota */
    }
    try {
        const raw = localStorage.getItem("user_quick_links");
        links = raw ? JSON.parse(raw) : [];
        if (!Array.isArray(links)) links = [];
    } catch {
        links = [];
    }
    return {
        theme,
        slots: readSlotCount(),
        links,
        history: readHistory(),
    };
}

function createAuthStore() {
    const { subscribe, set } = writable(empty);

    async function refresh() {
        if (typeof window === "undefined") {
            set(empty);
            return empty;
        }
        clearLegacyStorage();
        try {
            const res = await fetch(`${getApiBase()}/api/auth/me`, {
                credentials: "include",
            });
            if (!res.ok) {
                set(empty);
                return empty;
            }
            const me = await res.json();
            const next = meToAuthState(me);
            set(next);

            if (next.prefsImportedAt == null && typeof localStorage !== "undefined") {
                try {
                    const importRes = await fetch(`${getApiBase()}/api/account/import`, {
                        method: "POST",
                        credentials: "include",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify(
                            buildImportPayload(readBrowserImportInput()),
                        ),
                    });
                    if (importRes.ok) {
                        try {
                            const meRes = await fetch(`${getApiBase()}/api/auth/me`, {
                                credentials: "include",
                            });
                            if (meRes.ok) {
                                const updated = meToAuthState(await meRes.json());
                                set(updated);
                                return updated;
                            }
                        } catch {
                            /* keep pre-import session */
                        }
                    }
                } catch {
                    /* keep session; prefsImportedAt stays null */
                }
            }

            return next;
        } catch {
            set(empty);
            return empty;
        }
    }

    return {
        subscribe,
        init() {
            return refresh();
        },
        refresh,
        async logout() {
            try {
                await fetch(`${getApiBase()}/api/auth/logout`, {
                    method: "POST",
                    credentials: "include",
                });
            } catch {
                /* still clear locally */
            }
            set(empty);
        },
    };
}

export const auth = createAuthStore();
