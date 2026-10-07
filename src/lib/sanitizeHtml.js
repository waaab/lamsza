// Relative, not `$lib/`: the node:test suite imports this file directly and
// resolves no Vite alias.
import { safeImageSrc, safeLinkHref } from "./markdown.js";

/**
 * HTML sanitizer for the page bodies an admin writes (`pages.content`).
 *
 * The policy pages under `/iranyelvek/*` need real HTML - headings, lists,
 * tables, links - so we cannot escape it the way `markdown.js` does. Instead we
 * rebuild it: the input is tokenized, every tag is checked against an
 * allowlist, and the output tag is written here from a fixed name plus the
 * attributes the allowlist accepts. Nothing from the input is copied into the
 * output as markup, so no attribute, event handler or script can survive.
 *
 * Three rules do the work:
 *   1. A tag not on the allowlist loses its tag but keeps its text, so unknown
 *      markup degrades to prose instead of disappearing.
 *   2. `<script>`, `<style>`, `<iframe>`, `<svg>` and the rest of
 *      DROP_WITH_CONTENT lose their contents too, because those contents are
 *      code and not prose.
 *   3. Only the attributes in the allowlist are written, and each one has to
 *      pass its own check. `href` and `src` go through the same scheme test
 *      `markdown.js` uses, so `javascript:` and `data:` fail shut.
 *
 * This is defence in depth, not the only lock: the CSP in `svelte.config.js`
 * already stops an injected script from running. The content is admin-written,
 * so the realistic threat is an admin pasting markup from somewhere else, or an
 * admin account that is already compromised.
 */

/** Tags we keep. Everything else loses its tag but keeps its text. */
const ALLOWED = new Set([
    "a",
    "abbr",
    "address",
    "article",
    "b",
    "blockquote",
    "br",
    "caption",
    "code",
    "dd",
    "del",
    "div",
    "dl",
    "dt",
    "em",
    "figcaption",
    "figure",
    "h1",
    "h2",
    "h3",
    "h4",
    "h5",
    "h6",
    "hr",
    "i",
    "img",
    "ins",
    "li",
    "mark",
    "ol",
    "p",
    "pre",
    "s",
    "section",
    "small",
    "span",
    "strong",
    "sub",
    "sup",
    "table",
    "tbody",
    "td",
    "tfoot",
    "th",
    "thead",
    "time",
    "tr",
    "u",
    "ul",
]);

/**
 * Tags whose contents are code, markup for another parser, or form state - not
 * prose. We drop the element and everything up to its closing tag.
 */
const DROP_WITH_CONTENT = new Set([
    "applet",
    "base",
    "button",
    "embed",
    "form",
    "frame",
    "frameset",
    "head",
    "iframe",
    "input",
    "link",
    "math",
    "meta",
    "noembed",
    "noframes",
    "noscript",
    "object",
    "script",
    "select",
    "style",
    "svg",
    "template",
    "textarea",
    "title",
    "xmp",
]);

/** Allowed tags that never hold content. */
const VOID = new Set(["br", "hr", "img", "wbr"]);

/** Dropped tags that hold no content, so there is no closing tag to find. */
const VOID_LIKE_DROP = new Set(["base", "frame", "input", "link", "meta"]);

/**
 * Tags whose closing tag the HTML parser infers: a second `<li>` is a sibling
 * of the first, not a child. Each maps to the tags that stop the inference -
 * the `<li>` of an inner list must not close the `<li>` of the outer one.
 */
const IMPLIED_END = new Map([
    ["dd", new Set(["dl"])],
    ["dt", new Set(["dl"])],
    ["li", new Set(["ol", "ul"])],
    [
        "p",
        new Set([
            "blockquote",
            "dd",
            "div",
            "figure",
            "li",
            "section",
            "table",
            "td",
            "th",
        ]),
    ],
    ["td", new Set(["table", "tr"])],
    ["th", new Set(["table", "tr"])],
    ["tr", new Set(["table", "tbody", "tfoot", "thead"])],
]);

const ESCAPES = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
};

/**
 * A character reference the admin wrote on purpose, such as `&nbsp;` or
 * `&amp;`. We keep these as they are: a reference can only ever produce one
 * character, never a tag, so it is safe in text and in an attribute value - and
 * re-escaping the `&` would show `&nbsp;` to the reader as five letters, and
 * would break a `?a=1&amp;b=2` link target.
 */
const CHAR_REF = String.raw`&(?:#\d{1,7};|#[xX][0-9a-fA-F]{1,6};|[a-zA-Z][a-zA-Z0-9]{1,31};)`;
const ESCAPE_PATTERN = new RegExp(`${CHAR_REF}|[&<>"']`, "g");

/**
 * Escape everything that could open a tag or close an attribute, leaving
 * well-formed character references alone.
 *
 * @param {string} value
 * @returns {string}
 */
function escape(value) {
    return value.replace(ESCAPE_PATTERN, (match) =>
        match.length > 1 ? match : ESCAPES[match],
    );
}

/* -------------------------------------------------------------------------- */
/* Attribute checks. Each returns the value to write, or null to drop it.      */
/* -------------------------------------------------------------------------- */

/** Free text. It is escaped before it reaches the output. */
const text = (value) => (value === "" ? null : value);

const className = (value) => {
    const cleaned = value.replace(/[^A-Za-z0-9 _-]+/g, " ").trim();
    return cleaned === "" ? null : cleaned.replace(/\s+/g, " ");
};

const integer = (value) => (/^\d{1,4}$/.test(value) ? value : null);

const oneOf = (allowed) => (value) => {
    const lower = value.trim().toLowerCase();
    return allowed.includes(lower) ? lower : null;
};

const langTag = (value) =>
    /^[A-Za-z]{1,8}(?:-[A-Za-z0-9]{1,8})*$/.test(value.trim())
        ? value.trim()
        : null;

/** An ISO-ish date, time or duration. No other characters are allowed. */
const dateTime = (value) =>
    /^[\dTWPZ:+\-./ ]{1,40}$/i.test(value.trim()) ? value.trim() : null;

const GLOBAL_ATTRS = new Map([
    ["class", className],
    ["dir", oneOf(["ltr", "rtl", "auto"])],
    ["lang", langTag],
    ["title", text],
]);

/**
 * `id` and `name` are deliberately absent: an attacker-chosen id can shadow a
 * `document` or `window` property (DOM clobbering). `style` is absent too, so
 * admin markup cannot cover the page with an invisible overlay.
 */
const TAG_ATTRS = new Map([
    [
        "a",
        new Map([
            ["href", safeLinkHref],
            ["target", oneOf(["_blank", "_self"])],
        ]),
    ],
    [
        "img",
        new Map([
            ["src", safeImageSrc],
            ["alt", (value) => value],
            ["height", integer],
            ["width", integer],
        ]),
    ],
    ["ol", new Map([["start", integer]])],
    ["td", new Map([["colspan", integer], ["rowspan", integer]])],
    [
        "th",
        new Map([
            ["colspan", integer],
            ["rowspan", integer],
            ["scope", oneOf(["row", "col", "rowgroup", "colgroup"])],
        ]),
    ],
    ["time", new Map([["datetime", dateTime]])],
]);

/** Attributes we write ourselves, whatever the input said. */
const FORCED_ATTRS = new Map([
    ["a", new Map([["rel", "nofollow noopener noreferrer"]])],
    ["img", new Map([["loading", "lazy"]])],
]);

/** Allowed tags that are pointless, or broken, without a usable target. */
const REQUIRED_ATTR = new Map([["img", "src"]]);

/* -------------------------------------------------------------------------- */
/* Tokenizer                                                                  */
/* -------------------------------------------------------------------------- */

const WHITESPACE = /\s/;
const NAME_END = /[\s=>/]/;

/**
 * Read the attributes of a start tag.
 *
 * @param {string} src
 * @param {number} start index just after the tag name
 * @returns {{ attrs: [string, string][], next: number, terminated: boolean }}
 *   `terminated` is false when the input ends inside the tag, which is what the
 *   browser's own parser treats as a discarded tag.
 */
function readAttributes(src, start) {
    /** @type {[string, string][]} */
    const attrs = [];
    let i = start;

    while (i < src.length) {
        while (i < src.length && WHITESPACE.test(src[i])) i += 1;
        if (i >= src.length) break;
        if (src[i] === ">") return { attrs, next: i + 1, terminated: true };
        // A `/` here is either the one in `<br/>` or a stray; either way it is
        // not part of a name.
        if (src[i] === "/") {
            i += 1;
            continue;
        }

        const nameStart = i;
        while (i < src.length && !NAME_END.test(src[i])) i += 1;
        const name = src.slice(nameStart, i).toLowerCase();

        while (i < src.length && WHITESPACE.test(src[i])) i += 1;
        let value = "";
        if (src[i] === "=") {
            i += 1;
            while (i < src.length && WHITESPACE.test(src[i])) i += 1;
            const quote = src[i];
            if (quote === '"' || quote === "'") {
                const end = src.indexOf(quote, i + 1);
                if (end < 0) {
                    value = src.slice(i + 1);
                    i = src.length;
                } else {
                    value = src.slice(i + 1, end);
                    i = end + 1;
                }
            } else {
                const valueStart = i;
                while (i < src.length && !/[\s>]/.test(src[i])) i += 1;
                value = src.slice(valueStart, i);
            }
        }

        if (name !== "") attrs.push([name, value]);
    }

    return { attrs, next: src.length, terminated: false };
}

/**
 * Index just past the closing tag of a dropped element. The end of the input
 * counts as a close, so an unclosed `<script>` swallows the rest.
 *
 * @param {string} src
 * @param {string} name a name from DROP_WITH_CONTENT, so it is `[a-z]` only
 * @param {number} from
 * @returns {number}
 */
function skipElement(src, name, from) {
    if (VOID_LIKE_DROP.has(name)) return from;
    const match = new RegExp(`</${name}(?=[\\s/>])`, "i").exec(src.slice(from));
    if (!match) return src.length;
    const gt = src.indexOf(">", from + match.index);
    return gt < 0 ? src.length : gt + 1;
}

/**
 * Write one start tag from a fixed name and the attributes that pass the
 * allowlist.
 *
 * @param {string} name a name from ALLOWED
 * @param {[string, string][]} attrs
 * @returns {string|null} the tag, or null when the element has to go
 */
function buildTag(name, attrs) {
    const perTag = TAG_ATTRS.get(name);
    const written = new Map();

    for (const [attrName, raw] of attrs) {
        if (written.has(attrName)) continue;
        const check = perTag?.get(attrName) ?? GLOBAL_ATTRS.get(attrName);
        if (typeof check !== "function") continue;
        const value = check(raw, name);
        if (value === null || value === undefined) continue;
        written.set(attrName, String(value));
    }

    const required = REQUIRED_ATTR.get(name);
    if (required && !written.has(required)) return null;

    for (const [attrName, value] of FORCED_ATTRS.get(name) ?? []) {
        written.set(attrName, value);
    }

    let html = `<${name}`;
    for (const [attrName, value] of written) {
        html += ` ${attrName}="${escape(value)}"`;
    }
    return VOID.has(name) ? `${html} />` : `${html}>`;
}

/**
 * Turn admin-written HTML into HTML that is safe to drop into `{@html}`.
 *
 * The output is always balanced: an end tag with nothing open is ignored, and
 * anything still open at the end is closed here.
 *
 * @param {unknown} input
 * @returns {string} HTML, or "" when there is nothing to render
 */
export function sanitizeHtml(input) {
    if (typeof input !== "string" || input.trim() === "") return "";

    const src = input;
    /** @type {string[]} tags we have written and not yet closed */
    const open = [];
    let out = "";
    let i = 0;

    while (i < src.length) {
        const lt = src.indexOf("<", i);
        if (lt < 0) {
            out += escape(src.slice(i));
            break;
        }
        if (lt > i) out += escape(src.slice(i, lt));
        i = lt;

        // A comment, a doctype or a processing instruction. None of them render.
        if (src.startsWith("<!--", i)) {
            const end = src.indexOf("-->", i + 4);
            i = end < 0 ? src.length : end + 3;
            continue;
        }
        if (src[i + 1] === "!" || src[i + 1] === "?") {
            const end = src.indexOf(">", i);
            i = end < 0 ? src.length : end + 1;
            continue;
        }

        const endTag = /^<\/([a-zA-Z][a-zA-Z0-9]*)/.exec(src.slice(i));
        if (endTag) {
            const gt = src.indexOf(">", i);
            i = gt < 0 ? src.length : gt + 1;
            const name = endTag[1].toLowerCase();
            const depth = open.lastIndexOf(name);
            // An end tag for something we never opened is noise. One that
            // skips over open tags closes them too, so the output stays
            // balanced.
            if (depth >= 0) {
                while (open.length > depth) out += `</${open.pop()}>`;
            }
            continue;
        }

        const startTag = /^<([a-zA-Z][a-zA-Z0-9]*)/.exec(src.slice(i));
        if (!startTag) {
            // A bare `<` is text.
            out += "&lt;";
            i += 1;
            continue;
        }

        const name = startTag[1].toLowerCase();
        const tag = readAttributes(src, i + startTag[0].length);
        if (!tag.terminated) {
            // The input ends inside the tag. The browser discards such a tag,
            // and everything after it was inside the tag anyway.
            break;
        }
        i = tag.next;

        if (DROP_WITH_CONTENT.has(name)) {
            i = skipElement(src, name, i);
            continue;
        }
        // Not on the allowlist: drop the tag, keep the text inside it.
        if (!ALLOWED.has(name)) continue;

        const html = buildTag(name, tag.attrs);
        if (html === null) continue;

        // `<p>` after `<p>` is a sibling, not a child; the parser would close
        // the first one, so we do too - unless a tag in between puts the two
        // in different lists, rows or sections.
        const barriers = IMPLIED_END.get(name);
        if (barriers) {
            const depth = open.lastIndexOf(name);
            const nested = open
                .slice(depth + 1)
                .some((tag) => barriers.has(tag));
            if (depth >= 0 && !nested) {
                while (open.length > depth) out += `</${open.pop()}>`;
            }
        }

        out += html;
        if (!VOID.has(name)) open.push(name);
    }

    while (open.length) out += `</${open.pop()}>`;
    return out;
}
