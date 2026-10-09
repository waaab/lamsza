import test from "node:test";
import assert from "node:assert/strict";
import { attractionFacts } from "../src/lib/attractionFacts.js";

test("attractionFacts formats the set facts the Hungarian way", () => {
    assert.deepEqual(
        attractionFacts({ elevation_m: 946, area_km2: 0.22, depth_m: 7 }).map((f) => `${f.value} ${f.label}`),
        ["946 m magasság", "0,22 km² felszín", "7 m mélység"],
    );
});

test("attractionFacts leaves out what is not set", () => {
    assert.deepEqual(attractionFacts({ area_km2: 1234.5 }).map((f) => f.value), ["1234,5 km²"]);
    assert.deepEqual(attractionFacts({ elevation_m: null }), []);
    assert.deepEqual(attractionFacts(null), []);
});
