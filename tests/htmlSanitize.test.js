import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";
import { sanitizeHtml } from "../src/lib/sanitizeHtml.js";

/* -------------------------------------------------------------------------- */
/* What must survive: the markup an admin really writes on a policy page.     */
/* -------------------------------------------------------------------------- */

test("ordinary prose markup survives", () => {
    const html = sanitizeHtml(
        "<h2>Adatvédelem</h2><p>Egy <strong>fontos</strong> és <em>dőlt</em> szó.</p>",
    );
    assert.equal(
        html,
        "<h2>Adatvédelem</h2><p>Egy <strong>fontos</strong> és <em>dőlt</em> szó.</p>",
    );
});

test("lists, tables and blockquotes survive with their structure", () => {
    assert.equal(
        sanitizeHtml("<ul><li>egy<ul><li>kettő</li></ul></li></ul>"),
        "<ul><li>egy<ul><li>kettő</li></ul></li></ul>",
    );
    assert.equal(
        sanitizeHtml(
            '<table><thead><tr><th scope="col">A</th></tr></thead>' +
                '<tbody><tr><td colspan="2">B</td></tr></tbody></table>',
        ),
        '<table><thead><tr><th scope="col">A</th></tr></thead>' +
            '<tbody><tr><td colspan="2">B</td></tr></tbody></table>',
    );
    assert.equal(
        sanitizeHtml("<blockquote><p>Idézet</p></blockquote>"),
        "<blockquote><p>Idézet</p></blockquote>",
    );
});

test("a safe link keeps its target and gains rel", () => {
    assert.equal(
        sanitizeHtml('<a href="https://lamsza.com/x" title="Cím">Link</a>'),
        '<a href="https://lamsza.com/x" title="Cím" rel="nofollow noopener noreferrer">Link</a>',
    );
    assert.match(sanitizeHtml('<a href="/iranyelvek">Belső</a>'), /href="\/iranyelvek"/);
    assert.match(sanitizeHtml('<a href="mailto:a@b.hu">Mail</a>'), /href="mailto:a@b\.hu"/);
});

test("a character reference the admin wrote is left alone", () => {
    // Re-escaping the `&` would show five letters to the reader, and would turn
    // a `&amp;` in a link target into a literal "&amp;".
    assert.equal(sanitizeHtml("<p>egy&nbsp;kettő &amp; három</p>"), "<p>egy&nbsp;kettő &amp; három</p>");
    assert.match(sanitizeHtml('<a href="/x?a=1&amp;b=2">L</a>'), /href="\/x\?a=1&amp;b=2"/);
    // A lone `&` is still escaped, so it cannot start anything.
    assert.equal(sanitizeHtml("<p>a & b</p>"), "<p>a &amp; b</p>");
});

test("implied closing tags do not nest the wrong way", () => {
    assert.equal(sanitizeHtml("<p>egy<p>kettő"), "<p>egy</p><p>kettő</p>");
    assert.equal(sanitizeHtml("<ul><li>egy<li>kettő</ul>"), "<ul><li>egy</li><li>kettő</li></ul>");
});

/* -------------------------------------------------------------------------- */
/* What must not survive.                                                     */
/* -------------------------------------------------------------------------- */

test("a script never comes out, and neither does its body", () => {
    for (const payload of [
        "<script>alert(1)</script>",
        "<SCRIPT>alert(1)</SCRIPT>",
        '<script src="https://evil.test/x.js"></script>',
        "<script>var a = '</p>'; alert(1)</script>",
        "<style>body{background:url(https://evil.test)}</style>",
        "<template><img src=x onerror=alert(1)></template>",
        "<noscript><img src=x onerror=alert(1)></noscript>",
    ]) {
        const html = sanitizeHtml(payload);
        assert.doesNotMatch(html, /alert/i, payload);
        assert.doesNotMatch(html, /<script|<style|evil\.test/i, payload);
    }
});

test("an event handler is never written out", () => {
    for (const payload of [
        '<img src="https://lamsza.com/a.png" onerror="alert(1)">',
        '<p onclick="alert(1)">x</p>',
        '<p ONCLICK="alert(1)">x</p>',
        "<p onmouseover=alert(1)>x</p>",
        '<div onfocus="alert(1)" autofocus>x</div>',
        '<a href="/x" onclick="alert(1)">x</a>',
        // No quote, no space, no closing quote: the parser forgives, we do not.
        "<img src=x onerror=alert(1)//",
        '<p title="a" onclick=alert(1)>x</p>',
    ]) {
        const html = sanitizeHtml(payload);
        assert.doesNotMatch(html, /on[a-z]+\s*=/i, payload);
        assert.doesNotMatch(html, /alert/i, payload);
    }
});

test("an unsafe link or image target is dropped", () => {
    for (const payload of [
        '<a href="javascript:alert(1)">x</a>',
        '<a href="JaVaScRiPt:alert(1)">x</a>',
        '<a href="java\nscript:alert(1)">x</a>',
        '<a href="&#106;avascript:alert(1)">x</a>',
        '<a href="data:text/html,<script>alert(1)</script>">x</a>',
        '<a href="vbscript:msgbox(1)">x</a>',
    ]) {
        const html = sanitizeHtml(payload);
        assert.doesNotMatch(html, /href=/i, payload);
        assert.doesNotMatch(html, /alert|msgbox/i, payload);
    }
    // An image with no usable target loses the whole element.
    assert.equal(sanitizeHtml('<img src="javascript:alert(1)" alt="a">'), "");
});

test("markup for another parser is dropped with its contents", () => {
    for (const payload of [
        "<svg><script>alert(1)</script></svg>",
        "<svg/onload=alert(1)>",
        '<math><mtext><script>alert(1)</script></mtext></math>',
        '<iframe src="https://evil.test"></iframe>',
        '<iframe srcdoc="<script>alert(1)</script>"></iframe>',
        '<object data="https://evil.test/x.swf"></object>',
        '<embed src="https://evil.test/x">',
        '<form action="https://evil.test"><input name="a"></form>',
        '<base href="https://evil.test/">',
        '<meta http-equiv="refresh" content="0;url=https://evil.test">',
        '<link rel="stylesheet" href="https://evil.test/x.css">',
        '<textarea></textarea><img src=x onerror=alert(1)>',
    ]) {
        const html = sanitizeHtml(payload);
        assert.doesNotMatch(html, /alert|evil\.test/i, payload);
        assert.doesNotMatch(html, /<svg|<math|<iframe|<object|<embed|<form|<base|<meta|<link/i, payload);
    }
});

test("style and id attributes are dropped", () => {
    // `style` would let admin markup lay an invisible box over the page, and a
    // chosen `id` can shadow a document property (DOM clobbering).
    const html = sanitizeHtml(
        '<div style="position:fixed;inset:0" id="cookie" name="x" data-x="1" class="note">x</div>',
    );
    assert.equal(html, '<div class="note">x</div>');
});

test("an unknown tag loses its tag but keeps its text", () => {
    assert.equal(sanitizeHtml("<font color=red>Piros</font>"), "Piros");
    assert.equal(sanitizeHtml("<marquee>Fut</marquee>"), "Fut");
    assert.equal(sanitizeHtml("<my-widget>Szöveg</my-widget>"), "Szöveg");
});

test("a comment never comes out", () => {
    assert.equal(sanitizeHtml("<p>a<!-- <script>alert(1)</script> -->b</p>"), "<p>ab</p>");
    assert.equal(sanitizeHtml("<!--[if IE]><script>alert(1)</script><![endif]-->"), "");
    assert.equal(sanitizeHtml("<!DOCTYPE html><p>a</p>"), "<p>a</p>");
});

/* -------------------------------------------------------------------------- */
/* Shape of the output.                                                       */
/* -------------------------------------------------------------------------- */

test("the output is always balanced", () => {
    for (const payload of [
        "<div><p>a",
        "</div></p>a",
        "<p>a</span></p>",
        "<div><p>a</div>b",
        "<em><strong>a",
    ]) {
        const html = sanitizeHtml(payload);
        const opened = [...html.matchAll(/<([a-z0-9]+)(?:\s[^>]*)?>/g)]
            .filter((m) => !/\/>$/.test(m[0]))
            .map((m) => m[1]);
        const closed = [...html.matchAll(/<\/([a-z0-9]+)>/g)].map((m) => m[1]);
        assert.deepEqual(
            opened.slice().sort(),
            closed.slice().sort(),
            `${payload} -> ${html}`,
        );
    }
});

test("an unterminated tag takes nothing live with it", () => {
    assert.equal(sanitizeHtml('<p>a</p><img src=x onerror="alert(1)'), "<p>a</p>");
    assert.equal(sanitizeHtml("<p>a</p><div class='b"), "<p>a</p>");
});

test("nothing in, nothing out", () => {
    assert.equal(sanitizeHtml(""), "");
    assert.equal(sanitizeHtml("   "), "");
    assert.equal(sanitizeHtml(null), "");
    assert.equal(sanitizeHtml(undefined), "");
    assert.equal(sanitizeHtml(42), "");
    assert.equal(sanitizeHtml({ content: "<p>a</p>" }), "");
});

test("a prototype-shaped attribute name does not reach a checker", () => {
    const html = sanitizeHtml('<p constructor="x" __proto__="y" toString="z">a</p>');
    assert.equal(html, "<p>a</p>");
});

/* -------------------------------------------------------------------------- */
/* The wiring: the sanitizer has to sit where the content enters the app.     */
/* -------------------------------------------------------------------------- */

test("loadPageMeta sanitizes the content it hands out", () => {
    const source = readFileSync("src/lib/loadPageMeta.js", "utf8");
    assert.match(source, /import \{ sanitizeHtml \}/);
    assert.match(source, /content: sanitizeHtml\(page\?\.content\)/);
    assert.doesNotMatch(source, /content: page\?\.content/);
});

test("no page reads /api/pages past the sanitizer", () => {
    const pages = [
        "src/routes/(public)/iranyelvek/+page.svelte",
        "src/routes/(public)/iranyelvek/adatvedelem/+page.svelte",
        "src/routes/(public)/iranyelvek/feltetelek/+page.svelte",
        "src/routes/(public)/iranyelvek/sutik/+page.svelte",
    ];
    for (const path of pages) {
        const source = readFileSync(path, "utf8");
        assert.match(source, /\{@html page\.content\}/, path);
        assert.match(source, /loadPageMeta\(/, path);
        assert.doesNotMatch(source, /\/api\/pages/, path);
    }
});
