import assert from "node:assert/strict";
import { test } from "node:test";
import {
    compareEntryNames,
    listingAnchor,
    searchPreferredLocation,
    locationRing,
    sortDirectoryEntries,
} from "../src/lib/directoryListingOrder.js";

const locations = [
    { id: 11, name: "Kézdivásárhely", slug: "kezdivasarhely", type: "municípium", county_slug: "kovaszna", parent_id: null },
    { id: 455, name: "Nyújtód", slug: "nyujtod", type: "falu", county_slug: "kovaszna", parent_id: 11 },
    { id: 456, name: "Kézdioroszfalu", slug: "kezdioroszfalu", type: "falu", county_slug: "kovaszna", parent_id: 11 },
    { id: 10, name: "Kovászna", slug: "kovaszna", type: "város", county_slug: "kovaszna", parent_id: null },
    { id: 1001, name: "Kovászna", slug: "kovaszna", type: "megye", county_slug: "kovaszna", parent_id: null },
    { id: 1, name: "Csíkszereda", slug: "csikszereda", type: "municípium", county_slug: "hargita", parent_id: null },
    { id: 2, name: "Székelyudvarhely", slug: "szekelyudvarhely", type: "municípium", county_slug: "hargita", parent_id: null },
];

const anchor = { slug: "kezdivasarhely", name: "Kézdivásárhely", county_slug: "kovaszna", id: 11 };

test("name order ignores a leading space", () => {
    const imperial = { id: 65, name: " Imperial Medical Center" };
    const doctor = { id: 21, name: "Dr. Antal Zoltán" };
    assert.ok(compareEntryNames(doctor, imperial) < 0);
    const ordered = sortDirectoryEntries([imperial, doctor], { sortMode: "title" });
    assert.deepEqual(
        ordered.map((e) => e.id),
        [21, 65],
    );
});

test("a selected place lists the town, then its outer range, then the county, each by name", () => {
    const entries = [
        { id: 1, name: "Zeta", location_slug: "csikszereda", county_slug: "hargita" },
        { id: 2, name: "Alpha", location_slug: "kovaszna", county_slug: "kovaszna" },
        { id: 3, name: " Imperial Medical Center", location_slug: "kezdivasarhely", county_slug: "kovaszna" },
        { id: 4, name: "Nyújtódi bolt", location_slug: "nyujtod", county_slug: "kovaszna" },
        { id: 5, name: "Manifesto", location_slug: "kezdivasarhely", county_slug: "kovaszna" },
        { id: 6, name: "Dr. Antal Zoltán", location_slug: "szekelyudvarhely", county_slug: "hargita" },
    ];
    const ordered = sortDirectoryEntries(entries, {
        sortMode: "title",
        location: anchor,
        locations,
    });
    assert.deepEqual(
        ordered.map((e) => [e.location_slug, e.name.trim()]),
        [
            ["kezdivasarhely", "Imperial Medical Center"],
            ["kezdivasarhely", "Manifesto"],
            ["nyujtod", "Nyújtódi bolt"],
            ["kovaszna", "Alpha"],
            ["szekelyudvarhely", "Dr. Antal Zoltán"],
            ["csikszereda", "Zeta"],
        ],
    );
    assert.equal(locationRing(entries[3], anchor, locations), 1);
    assert.equal(locationRing(entries[1], anchor, locations), 2);
    assert.equal(locationRing(entries[0], anchor, locations), 3);
});

test("without a location anchor, order stays name A-Z across places", () => {
    const entries = [
        { id: 3, name: " Imperial Medical Center", location_slug: "kezdivasarhely" },
        { id: 6, name: "Dr. Antal Zoltán", location_slug: "szekelyudvarhely" },
    ];
    const ordered = sortDirectoryEntries(entries, { sortMode: "title", locations });
    assert.equal(ordered[0].id, 6);
});

test("a signed-in user's saved place stays available for the location menu", () => {
    const saved = { slug: "csikszereda", name: "Csíkszereda", county_slug: "hargita" };
    assert.equal(searchPreferredLocation(saved, false), null);
    assert.equal(searchPreferredLocation(null, true), null);
    assert.equal(searchPreferredLocation(saved, true)?.slug, "csikszereda");
});

test("listing anchor prefers the saved place, then the site location", () => {
    const site = { slug: "kezdivasarhely", name: "Kézdivásárhely", county_slug: "kovaszna" };
    const saved = { slug: "csikszereda", name: "Csíkszereda", county_slug: "hargita" };
    assert.equal(listingAnchor(site, saved, false)?.slug, "kezdivasarhely");
    assert.equal(listingAnchor(site, saved, true)?.name, "Csíkszereda");
    assert.equal(listingAnchor(site, null, true)?.name, "Kézdivásárhely");
    assert.equal(listingAnchor(null, null, false), null);
});
