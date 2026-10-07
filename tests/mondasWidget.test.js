import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import test from "node:test";

const src = readFileSync(new URL("../src/lib/components/MondasWidget.svelte", import.meta.url), "utf8");

test("the Napi Székely Mondás keeps its markup and CSS byte for byte", () => {
    // Owner's decision (OPEN_ITEMS, Mondások): only the data source changed, from
    // Lámsza's own API to Szótár's. Everything after </script> (the template and
    // the <style>) is pinned to the version before that change.
    const rest = src.slice(src.indexOf("</script>"));
    assert.equal(createHash("sha256").update(rest).digest("hex"), "cefbcb67c58d778c907c7c643b042fa6bbc4ec9a6e1c92d773af2d11f1a421d8");
});

test("the widget asks Szótár for today's mondás and sends no date", () => {
    assert.match(src, /szotarUrl\("\/api\/proverbs\/today"\)/);
    assert.doesNotMatch(src, /\/api\/mondasok|localCalendarISODate|\?date=/);
    assert.match(src, /credentials: "omit"/);
});
