import { browser, dev } from "$app/environment";

export const ENTRY_PHOTO_SLOTS = 4;

const PLACEHOLD_COLORS = ["2d6a4f", "1d3557", "9a3412", "4c1d95"];

function isNonProdHost() {
    if (!browser) return false;
    const host = window.location.hostname;
    return (
        host === "localhost" ||
        host === "127.0.0.1" ||
        host.endsWith(".local") ||
        host.includes("staging")
    );
}

/** Demo gallery photos for Vite/local/staging only — never on production hosts. */
export function useDemoEntryPhotos() {
    return dev || isNonProdHost();
}

export function demoEntryPhoto(slug, index) {
    const seed = String(slug || "entry")
        .toLowerCase()
        .replace(/[^a-z0-9-]+/g, "-")
        .replace(/^-|-$/g, "") || "entry";
    const n = Number(index) || 0;
    const color = PLACEHOLD_COLORS[n % PLACEHOLD_COLORS.length];
    return {
        src: `https://picsum.photos/seed/${encodeURIComponent(`${seed}-${n + 1}`)}/800/600`,
        fallback: `https://placehold.co/800x600/${color}/ffffff/png?text=${n + 1}`,
    };
}

/** Demo slides for the public gallery when an entry has no stored photos. */
export function demoGallerySlides(entry) {
    if (!useDemoEntryPhotos()) return [];
    const slug = String(entry?.slug ?? "entry");
    const label = String(entry?.name ?? "").trim() || "Bejegyzés";
    return Array.from({ length: ENTRY_PHOTO_SLOTS }, (_, i) => {
        const demo = demoEntryPhoto(slug, i);
        return {
            url: demo.src,
            src: demo.src,
            fallback: demo.fallback,
            alt: `${label} — fotó ${i + 1}`,
            title: `${label} — fotó ${i + 1}`,
            description: "Példakép (csak fejlesztői és staging környezetben).",
            width: 800,
            height: 600,
            loading: i === 0 ? "eager" : "lazy",
            fetchpriority: i === 0 ? "high" : "low",
        };
    });
}
