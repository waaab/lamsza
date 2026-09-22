import { get, writable } from 'svelte/store';
import { getApiBase } from '$lib/api.js';
import { auth } from './auth.js';

export const theme = writable('system');

const THEMES = ["light", "dark", "system"];
export const LABELS = {
    light: "☀️ Világos mód",
    dark: "🌙 Sötét mód",
    system: "🖥️ Rendszer",
};

function persistThemeToAccount(newTheme) {
    const state = get(auth);
    if (!state.loggedIn || typeof window === "undefined") return;
    fetch(`${getApiBase()}/api/account/preferences`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: newTheme }),
    }).catch(() => {});
}

export function applyTheme(newTheme) {
    if (newTheme === "system") {
        document.documentElement.removeAttribute("data-theme");
    } else {
        document.documentElement.setAttribute("data-theme", newTheme);
    }
    theme.set(newTheme);
    localStorage.setItem("theme", newTheme);
    persistThemeToAccount(newTheme);
}

export function cycleTheme(currentTheme) {
    const idx = THEMES.indexOf(currentTheme);
    const next = THEMES[(idx + 1) % THEMES.length];
    applyTheme(next);
}
