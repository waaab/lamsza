import assert from "node:assert/strict";
import { test } from "node:test";
import {
    ENTRY_TYPE_INTEZMENY,
    ENTRY_TYPE_SZEMELY,
    ENTRY_TYPE_VALLALKOZAS,
    canonicalEntryType,
    canonicalEntryTypeKey,
} from "../src/lib/entryType.js";
import {
    buildDirectoryTagCloud,
    entryMatchesAsideFilters,
} from "../src/lib/directoryTagCloud.js";

test("canonical labels match the admin catalog", () => {
    assert.equal(ENTRY_TYPE_SZEMELY, "Személy");
    assert.equal(ENTRY_TYPE_VALLALKOZAS, "Vállalkozás");
    assert.equal(ENTRY_TYPE_INTEZMENY, "Intézmény");
});

test("canonicalEntryType accepts only the three catalog names", () => {
    assert.equal(canonicalEntryType("Személy"), ENTRY_TYPE_SZEMELY);
    assert.equal(canonicalEntryType("Vállalkozás"), ENTRY_TYPE_VALLALKOZAS);
    assert.equal(canonicalEntryType("Intézmény"), ENTRY_TYPE_INTEZMENY);
    assert.equal(canonicalEntryType("service"), "");
    assert.equal(canonicalEntryType("Szolgáltatás"), "");
    assert.equal(canonicalEntryType("Cég"), "");
    assert.equal(canonicalEntryType("Egyéb"), "");
});

test("canonicalEntryType: empty stays empty", () => {
    assert.equal(canonicalEntryType(""), "");
    assert.equal(canonicalEntryType(null), "");
    assert.equal(canonicalEntryType(undefined), "");
});

test("tag cloud counts the three catalog types", () => {
    const { typeItems } = buildDirectoryTagCloud([
        { type: "Személy", tags: [] },
        { type: "Vállalkozás", tags: [] },
        { type: "Intézmény", tags: [] },
        { type: "service", tags: [] },
    ]);
    const byKey = Object.fromEntries(typeItems.map((i) => [i.key, i]));
    assert.equal(byKey[canonicalEntryTypeKey(ENTRY_TYPE_SZEMELY)].count, 1);
    assert.equal(byKey[canonicalEntryTypeKey(ENTRY_TYPE_VALLALKOZAS)].count, 1);
    assert.equal(byKey[canonicalEntryTypeKey(ENTRY_TYPE_INTEZMENY)].count, 1);
    assert.equal(typeItems.length, 3);
});

test("aside type filter matches canonical key", () => {
    const key = canonicalEntryTypeKey(ENTRY_TYPE_VALLALKOZAS);
    assert.equal(entryMatchesAsideFilters({ type: "Vállalkozás" }, key, null), true);
    assert.equal(entryMatchesAsideFilters({ type: "Személy" }, key, null), false);
});
