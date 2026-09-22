import assert from "node:assert/strict";
import { test } from "node:test";
import {
    ENTRY_TYPE_CEG,
    ENTRY_TYPE_EGYEB,
    ENTRY_TYPE_SERVICE,
    canonicalEntryType,
    canonicalEntryTypeKey,
    filterServiceEntries,
    isServiceEntry,
} from "../src/lib/entryType.js";
import {
    buildDirectoryTagCloud,
    entryMatchesAsideFilters,
} from "../src/lib/directoryTagCloud.js";

test("canonical labels match the admin catalog", () => {
    assert.equal(ENTRY_TYPE_SERVICE, "Szolgáltatás");
    assert.equal(ENTRY_TYPE_CEG, "Cég");
    assert.equal(ENTRY_TYPE_EGYEB, "Egyéb");
});

test("canonicalEntryType: service aliases → Szolgáltatás", () => {
    assert.equal(canonicalEntryType("service"), ENTRY_TYPE_SERVICE);
    assert.equal(canonicalEntryType("Service"), ENTRY_TYPE_SERVICE);
    assert.equal(canonicalEntryType("Szolgáltatás"), ENTRY_TYPE_SERVICE);
    assert.equal(canonicalEntryType("  szolgáltatás  "), ENTRY_TYPE_SERVICE);
});

test("canonicalEntryType: company aliases → Cég", () => {
    assert.equal(canonicalEntryType("Cég"), ENTRY_TYPE_CEG);
    assert.equal(canonicalEntryType("ceg"), ENTRY_TYPE_CEG);
    assert.equal(canonicalEntryType("company"), ENTRY_TYPE_CEG);
});

test("canonicalEntryType: leftover entry default → Egyéb", () => {
    assert.equal(canonicalEntryType("entry"), ENTRY_TYPE_EGYEB);
    assert.equal(canonicalEntryType("Egyéb"), ENTRY_TYPE_EGYEB);
    assert.equal(canonicalEntryType("other"), ENTRY_TYPE_EGYEB);
});

test("canonicalEntryType: empty stays empty", () => {
    assert.equal(canonicalEntryType(""), "");
    assert.equal(canonicalEntryType(null), "");
    assert.equal(canonicalEntryType(undefined), "");
});

test("isServiceEntry is true for service aliases only", () => {
    assert.equal(isServiceEntry({ type: "service" }), true);
    assert.equal(isServiceEntry({ type: "Szolgáltatás" }), true);
    assert.equal(isServiceEntry({ type: "Cég" }), false);
    assert.equal(isServiceEntry({ type: "entry" }), false);
    assert.equal(isServiceEntry({ category: "hivatalok" }), false);
});

test("filterServiceEntries keeps services, not companies or leftover entry", () => {
    const rows = [
        { name: "Dr. Papp", type: "service", category: "egeszsegugy" },
        { name: "Gimnázium", type: "Szolgáltatás", category: "oktatas" },
        { name: "Góbé", type: "Cég", category: "egyeb" },
        { name: "FRHG", type: "entry", category: "" },
        { name: "Villanyszerelő", type: "service", category: "mesteremberek" },
    ];
    const names = filterServiceEntries(rows).map((e) => e.name);
    assert.deepEqual(names, ["Dr. Papp", "Gimnázium", "Villanyszerelő"]);
});

test("tag cloud merges service and Szolgáltatás into one type", () => {
    const { typeItems } = buildDirectoryTagCloud([
        { type: "service", tags: [] },
        { type: "Szolgáltatás", tags: [] },
        { type: "Cég", tags: [] },
        { type: "entry", tags: [] },
    ]);
    const byKey = Object.fromEntries(typeItems.map((i) => [i.key, i]));
    assert.equal(byKey[canonicalEntryTypeKey(ENTRY_TYPE_SERVICE)].count, 2);
    assert.equal(byKey[canonicalEntryTypeKey(ENTRY_TYPE_SERVICE)].label, ENTRY_TYPE_SERVICE);
    assert.equal(byKey[canonicalEntryTypeKey(ENTRY_TYPE_CEG)].count, 1);
    assert.equal(byKey[canonicalEntryTypeKey(ENTRY_TYPE_EGYEB)].count, 1);
});

test("aside type filter matches service alias against canonical key", () => {
    const key = canonicalEntryTypeKey(ENTRY_TYPE_SERVICE);
    assert.equal(entryMatchesAsideFilters({ type: "service" }, key, null), true);
    assert.equal(entryMatchesAsideFilters({ type: "Cég" }, key, null), false);
});
