import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const src = readFileSync(new URL("../src/routes/(public)/esemenyek/+page.svelte", import.meta.url), "utf8");

test("the events calendar opens on Bucharest's year from the server's clock (R19)", () => {
    // The browser's clock and zone must not pick the year: near New Year a
    // visitor outside Romania, or with a wrong clock, would see the wrong one.
    assert.doesNotMatch(src, /new Date\(\)/);
    assert.match(src, /import \{ referenceTodayYMD \} from "\$lib\/eventStatus"/);
    assert.match(src, /let calendarViewYear = bucharestYear\(\);/);
});
