import assert from "node:assert/strict";
import test from "node:test";
import { homepageAttractionWeather, homepageEventPlace, homepageSettlements } from "../src/lib/favoriteHomepage.js";

test("saved place replaces the site default", () => {
    const rows = homepageSettlements(
        { slug: "kezdivasarhely", name: "Kézdivásárhely" },
        { slug: "sepsiszentgyorgy", name: "Sepsiszentgyörgy" },
    );
    assert.deepEqual(rows, [{ slug: "sepsiszentgyorgy", name: "Sepsiszentgyörgy" }]);
});

test("favorite settlements do not replace the site default", () => {
    const rows = homepageSettlements(
        { slug: "csikszereda", name: "Csíkszereda" },
        null,
    );
    assert.deepEqual(rows, [{ slug: "csikszereda", name: "Csíkszereda" }]);
});

test("a saved place wins over the site default", () => {
    const rows = homepageSettlements(
        { slug: "csikszereda", name: "Csíkszereda" },
        { slug: "sepsiszentgyorgy", name: "Sepsiszentgyörgy" },
    );
    assert.deepEqual(rows.map((r) => r.slug), ["sepsiszentgyorgy"]);
});

test("missing site default yields no homepage place", () => {
    assert.deepEqual(homepageSettlements({}, null), []);
});

test("homepage events use a saved place", () => {
    assert.deepEqual(
        homepageEventPlace({ slug: "csikszereda", name: "Csíkszereda" }),
        { slug: "csikszereda", name: "Csíkszereda" },
    );
});

test("homepage events list every location when no place is saved", () => {
    assert.equal(homepageEventPlace(null), null);
    assert.equal(homepageEventPlace({ slug: "  ", name: "Csíkszereda" }), null);
});

test("attraction weather skips missing coordinates", () => {
    const rows = homepageAttractionWeather({
        attractions: [
            { slug: "to", name: "Tó", latitude: 46.1, longitude: 25.8 },
            { slug: "var", name: "Vár", latitude: null, longitude: null },
        ],
    });
    assert.equal(rows.length, 1);
    assert.equal(rows[0].slug, "to");
});
