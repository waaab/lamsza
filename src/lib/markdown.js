import { Marked } from "marked";

/**
 * Markdown rendering that is safe to drop into `{@html}`.
 *
 * The content comes from people we do not trust. Any signed-in user can file an
 * attraction suggestion, and an admin reading the prose cannot be expected to
 * spot an `onerror=` payload inside it. So we never let raw HTML through: this
 * renderer escapes it into visible text and builds every tag itself, which
 * leaves no place for an attribute or a script to appear.
 *
 * Two rules do the work:
 *   1. Raw HTML in the source becomes text (`<img onerror=...>` shows up as
 *      characters on the page, it does not become an element).
 *   2. Link and image targets must use a scheme we allow, which rules out
 *      `javascript:` and `data:`.
 *
 * Everything else marked emits is built from a fixed set of tags with escaped
 * text inside, so it carries no attacker-controlled markup.
 */

const ESCAPES = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
};

/** Escape every character that could start a tag or break out of an attribute. */
export function escapeHtml(value) {
    if (value === null || value === undefined) return "";
    return String(value).replace(/[&<>"']/g, (c) => ESCAPES[c]);
}

/** Drop whitespace and control characters, which browsers ignore inside a URL. */
function stripUrlNoise(raw) {
    let out = "";
    for (const ch of raw) {
        const code = ch.codePointAt(0);
        if (code <= 0x20 || code === 0x7f) continue;
        if (/\s/.test(ch)) continue;
        out += ch;
    }
    return out;
}

/**
 * Keep a link target only when it is plainly safe.
 *
 * The test reads the literal text, before any HTML entity is decoded. A target
 * that *starts* with a scheme we allow, or with `/`, `#` or `?`, cannot be
 * turned into another scheme by decoding what comes after it. Everything else
 * is dropped, so `javascript:`, `data:` and any unknown scheme fail shut.
 *
 * @param {unknown} raw
 * @param {string[]} schemes allowed schemes, each written with its colon
 * @returns {string|null} the target to use, or null when it is not safe
 */
function safeUrl(raw, schemes) {
    if (typeof raw !== "string") return null;
    // Browsers ignore whitespace and control characters inside a URL, so remove
    // them before the test; otherwise `java\nscript:` would slip past.
    const url = stripUrlNoise(raw);
    if (!url) return null;

    // An entity such as `&#58;` is how an attacker writes a colon without
    // writing one. A real scheme or host never holds an `&`, so refuse one
    // there and the entity-encoded spellings of `javascript:` cannot form.
    const head = url.split(/[/?#]/, 1)[0];
    if (head.includes("&")) return null;

    const lower = url.toLowerCase();
    if (schemes.some((scheme) => lower.startsWith(scheme))) return url;
    // Site-relative and in-page targets.
    if (/^[/#?]/.test(url)) return url;
    // A plain relative path: safe only when no scheme appears before the path.
    if (!head.includes(":")) return url;
    return null;
}

const LINK_SCHEMES = ["http://", "https://", "mailto:", "tel:"];
const IMAGE_SCHEMES = ["http://", "https://"];

export function safeLinkHref(raw) {
    return safeUrl(raw, LINK_SCHEMES);
}

export function safeImageSrc(raw) {
    return safeUrl(raw, IMAGE_SCHEMES);
}

function titleAttr(title) {
    return title ? ` title="${escapeHtml(title)}"` : "";
}

const safeRenderer = {
    /** Raw HTML blocks and inline tags become visible text, never markup. */
    html({ text }) {
        return escapeHtml(text);
    },

    link({ href, title, tokens }) {
        const text = this.parser.parseInline(tokens);
        const target = safeLinkHref(href);
        // A target we refuse keeps its label; the label is already escaped.
        if (!target) return text;
        return `<a href="${escapeHtml(target)}"${titleAttr(title)} rel="nofollow noopener noreferrer">${text}</a>`;
    },

    image({ href, title, text }) {
        const src = safeImageSrc(href);
        if (!src) return escapeHtml(text);
        return `<img src="${escapeHtml(src)}" alt="${escapeHtml(text)}"${titleAttr(title)} loading="lazy" />`;
    },
};

const safeMarked = new Marked({ async: false, gfm: true, breaks: false });
safeMarked.use({ renderer: safeRenderer });

/**
 * Turn untrusted Markdown into HTML that holds no script and no raw markup.
 *
 * @param {unknown} source
 * @returns {string} HTML, or "" when there is nothing to render
 */
export function renderMarkdown(source) {
    if (typeof source !== "string" || source.trim() === "") return "";
    return safeMarked.parse(source, { async: false });
}
