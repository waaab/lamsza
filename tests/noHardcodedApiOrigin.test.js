import assert from "node:assert/strict";
import { test } from "node:test";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

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

test("pages do not hard-code localhost:3000 as the API origin", () => {
    const hits = [];
    for (const file of walk("src")) {
        if (file.replaceAll("\\", "/").endsWith("src/lib/api.js")) continue;
        const text = readFileSync(file, "utf8");
        if (text.includes("localhost:3000")) hits.push(file);
    }
    assert.equal(
        hits.join("\n"),
        "",
        `hardcoded localhost:3000 in:\n${hits.join("\n")}`,
    );
});
