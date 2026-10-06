import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";

const read = (relative) =>
    readFileSync(new URL(relative, import.meta.url), "utf-8");

const svelteConfig = read("../svelte.config.js");
const appHtml = read("../src/app.html");

// The policy is the second lock on the Markdown XSS: if script-src ever loosens,
// an injected tag runs again and can post the stolen session anywhere. These
// tests fail loudly before that ships.

test("svelte.config.js declares a Content-Security-Policy", () => {
    assert.match(svelteConfig, /csp:\s*{/);
    assert.match(svelteConfig, /mode:\s*'hash'/);
});

test("script-src allows this origin only, never unsafe-inline or unsafe-eval", () => {
    const scriptSrc = svelteConfig.match(/'script-src':\s*\[([^\]]*)\]/);
    assert.ok(scriptSrc, "script-src is missing from the policy");
    assert.match(scriptSrc[1], /'self'/);
    assert.doesNotMatch(scriptSrc[1], /unsafe-inline/);
    assert.doesNotMatch(scriptSrc[1], /unsafe-eval/);
    assert.doesNotMatch(scriptSrc[1], /\bhttps:'/, "script-src must not allow any https host");
});

test("connect-src is an allowlist, so a stolen session cannot be posted out", () => {
    const connectSrc = svelteConfig.match(/'connect-src':\s*\[([^\]]*)\]/);
    assert.ok(connectSrc, "connect-src is missing from the policy");
    assert.match(connectSrc[1], /'self'/);
    assert.doesNotMatch(connectSrc[1], /\*/);
    assert.doesNotMatch(connectSrc[1], /'https:'/);
});

test("object-src and base-uri are locked down", () => {
    assert.match(svelteConfig, /'object-src':\s*\['none'\]/);
    assert.match(svelteConfig, /'base-uri':\s*\['self'\]/);
});

test("app.html carries no inline script, which script-src 'self' would block", () => {
    for (const match of appHtml.matchAll(/<script([^>]*)>([\s\S]*?)<\/script>/g)) {
        const [, attrs, body] = match;
        assert.ok(
            / src=/.test(attrs),
            `app.html has an inline <script> block; move it to static/ instead:\n${body.trim().slice(0, 120)}`,
        );
    }
});

test("the theme is set from an external file before the first paint", () => {
    assert.match(appHtml, /<script src="%sveltekit\.assets%\/theme-init\.js"><\/script>/);
    // Must stay ahead of the body so no wrong-colour frame is shown.
    assert.ok(appHtml.indexOf("theme-init.js") < appHtml.indexOf("%sveltekit.body%"));
    assert.match(read("../static/theme-init.js"), /localStorage\.getItem\('theme'\)/);
});
