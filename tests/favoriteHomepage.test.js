import assert from "node:assert/strict";
import test from "node:test";
import { homepageAttractionWeather, homepageSettlements } from "../src/lib/favoriteHomepage.js";

test("no favorite settlement uses the site default", () => {
    const rows = homepageSettlements({ settlements: [] }, { slug: "csikszereda", name: "Csíkszereda" });
    assert.deepEqual(rows, [{ slug: "csikszereda", name: "Csíkszereda" }]);
});

test("favorite settlements keep added order and skip the default", () => {
    const rows = homepageSettlements(
        { settlements: [{ slug: "udvarhely", name: "Székelyudvarhely" }, { slug: "csikszereda", name: "Csíkszereda" }] },
        { slug: "csikszereda", name: "Csíkszereda" },
    );
    assert.deepEqual(rows.map((r) => r.slug), ["udvarhely", "csikszereda"]);
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
