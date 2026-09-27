import assert from "node:assert/strict";
import { test } from "node:test";
import { formatSearchElapsed } from "../src/lib/searchElapsed.js";

test("formats elapsed search time in Hungarian seconds", () => {
    assert.equal(formatSearchElapsed(170), "0,2 mp");
    assert.equal(formatSearchElapsed(40), "0,1 mp");
    assert.equal(formatSearchElapsed(848), "0,8 mp");
    assert.equal(formatSearchElapsed(1200), "1,2 mp");
    assert.equal(formatSearchElapsed(12360), "12,4 mp");
});
