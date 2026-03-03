import { writable } from 'svelte/store';

export const theme = writable('system');

const THEMES = ["light", "dark", "system"];
export const LABELS = {
    light: "☀️ Világos mód",
    dark: "🌙 Sötét mód",
    system: "🖥️ Rendszer",
};

export function applyTheme(newTheme) {
    if (newTheme === "system") {
        document.documentElement.removeAttribute("data-theme");
    } else {
        document.documentElement.setAttribute("data-theme", newTheme);
    }
    theme.set(newTheme);
    localStorage.setItem("theme", newTheme);
}

export function cycleTheme(currentTheme) {
    const idx = THEMES.indexOf(currentTheme);
    const next = THEMES[(idx + 1) % THEMES.length];
    applyTheme(next);
}
