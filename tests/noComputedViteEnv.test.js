import assert from "node:assert/strict";
import { test } from "node:test";
import { readdirSync, readFileSync, statSync, existsSync } from "node:fs";
import { join } from "node:path";

/**
 * Vite replaces `import.meta.env.VITE_X` with the literal value at build time.
 * It can only do that when the property name is written out. A computed lookup
 * (`import.meta.env[key]`) has no name to match, so Vite gives up and inlines the
 * whole env object instead - which put every `VITE_*` value, including an API key,
 * into the public bundle. These two tests keep that door shut.
 */

function walk(dir) {
    /** @type {string[]} */
    const out = [];
    for (const name of readdirSync(dir)) {
        const p = join(dir, name);
        if (statSync(p).isDirectory()) out.push(...walk(p));
        else if (/\.(js|svelte)$/.test(name)) out.push(p);
    }
    return out;
}

/** Drop comments so a doc comment that names the bad pattern is not a hit. */
function stripComments(text) {
    return text.replace(/\/\*[\s\S]*?\*\//g, "").replace(/(^|[^:])\/\/.*$/gm, "$1");
}

test("no source file indexes import.meta.env with a computed key", () => {
    /** `import.meta.env[...]` or `import.meta.env?.[...]` - a dot-name is fine. */
    const computed = /import\s*\.\s*meta\s*\.\s*env\s*(\?\s*\.)?\s*\[/;
    const hits = [];
    for (const file of walk("src")) {
        if (computed.test(stripComments(readFileSync(file, "utf8")))) hits.push(file);
    }
    assert.equal(
        hits.join("\n"),
        "",
        `computed import.meta.env lookup inlines every VITE_* value into the public bundle. Use one static import.meta.env.VITE_NAME per key in:\n${hits.join("\n")}`,
    );
});

test("no VITE_ name carries a secret", () => {
    /** `VITE_*` values are public. Anything key-shaped must use an unprefixed name. */
    const secretish = /^\s*VITE_[A-Z0-9_]*(KEY|SECRET|TOKEN|PASSWORD|CREDENTIAL)/m;
    for (const file of [".env", ".env.example", ".env.test"]) {
        if (!existsSync(file)) continue;
        const text = readFileSync(file, "utf8");
        assert.equal(
            secretish.test(text),
            false,
            `${file} defines a secret under a VITE_ name; every VITE_* value is readable in the public bundle`,
        );
    }
});
