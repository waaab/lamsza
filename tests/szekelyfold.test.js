import test from "node:test";
import assert from "node:assert/strict";
import { COUNTY_ORDER, SEAT_ORDER, countyColor, honeycomb, inRegionOrder } from "../src/lib/szekelyfold.js";

test("honeycomb puts five cells two over three and three cells two over one", () => {
    const five = honeycomb(5);
    assert.deepEqual(
        five.cells.map((c) => [Math.round(c.x), c.y]),
        [[135, 80], [246, 80], [79, 176], [190, 176], [301, 176]],
    );
    assert.equal(five.height, 256);
    const three = honeycomb(3);
    assert.deepEqual(three.cells.map((c) => Math.round(c.x)), [135, 246, 190]);
    assert.equal(honeycomb(0).cells.length, 0);
});

test("inRegionOrder follows the map's order and colours, unknown ones last", () => {
    const seats = [
        { slug: "udvarhelyszek", name: "Udvarhelyszék" },
        { slug: "ujszek", name: "Újszék" },
        { slug: "csikszek", name: "Csíkszék" },
    ];
    const out = inRegionOrder(seats, SEAT_ORDER);
    assert.deepEqual(out.map((s) => s.slug), ["csikszek", "udvarhelyszek", "ujszek"]);
    assert.equal(out[0].color, SEAT_ORDER[0][1]);
    assert.match(out[2].color, /^#/);
});

test("countyColor matches the Megyék map, muted for an unknown county", () => {
    assert.equal(countyColor("hargita"), COUNTY_ORDER[0][1]);
    assert.equal(countyColor("feher"), "var(--text-muted)");
});
