import assert from "node:assert/strict";
import { test } from "node:test";
import {
    canonicalEntryCategory,
    canonicalEntryCategoryKey,
    DIRECTORY_CATALOG,
    directoryCategoryTabs,
    directoryChildTabs,
    directoryParentSlug,
    entryMatchesCategory,
    parentCategoryTabActive,
} from "../src/lib/entryCategory.js";

test("canonicalEntryCategory returns the stored name", () => {
    assert.equal(canonicalEntryCategory("Bútor"), "Bútor");
    assert.equal(canonicalEntryCategory("  Étterem  "), "Étterem");
    assert.equal(canonicalEntryCategory("egeszsegugy"), "egeszsegugy");
    assert.equal(canonicalEntryCategory("Vendéglő"), "Vendéglő");
    assert.equal(canonicalEntryCategory(""), "");
});

test("entryMatchesCategory matches child slug", () => {
    assert.equal(entryMatchesCategory({ category: "Bútor" }, "butor"), true);
    assert.equal(entryMatchesCategory({ category: "Étterem" }, "etterem"), true);
    assert.equal(entryMatchesCategory({ category: "Oktatás" }, "hivatalok"), false);
    assert.equal(entryMatchesCategory({ category: "Hivatalok" }, "osszes"), true);
});

test("entryMatchesCategory matches parent slug for a child category", () => {
    assert.equal(entryMatchesCategory({ category: "Bútor" }, "vasarlas"), true);
    assert.equal(entryMatchesCategory({ category: "Turbószerviz" }, "auto"), true);
    assert.equal(entryMatchesCategory({ category: "Bútor" }, "auto"), false);
});

test("directoryCategoryTabs shows parent shelves with visible children", () => {
    const tabs = directoryCategoryTabs(
        DIRECTORY_CATALOG,
        [{ category: "Bútor" }],
        [{ category: "Bank", entry_id: 0 }],
    );
    assert.equal(tabs[0].id, "osszes");
    const ids = tabs.map((t) => t.id);
    assert.ok(ids.includes("vasarlas"));
    assert.ok(ids.includes("penzugy"));
    assert.equal(tabs.filter((t) => t.id === "butor").length, 0);
    const shopping = tabs.find((t) => t.id === "vasarlas");
    assert.equal(shopping.label, "Vásárlás");
    assert.equal(shopping.url, "/index/vasarlas");
});

test("directoryChildTabs hides empty child shelves", () => {
    const children = directoryChildTabs(
        "vasarlas",
        DIRECTORY_CATALOG,
        [{ category: "Bútor" }],
        [],
    );
    assert.deepEqual(
        children.map((row) => row.id),
        ["butor"],
    );
});

test("directoryChildTabs resolves a child slug to its parent row", () => {
    const children = directoryChildTabs(
        "butor",
        DIRECTORY_CATALOG,
        [{ category: "Bútor" }],
        [{ category: "Bútor", entry_id: 0 }],
    );
    assert.deepEqual(
        children.map((row) => row.id),
        ["butor"],
    );
});

test("directoryParentSlug maps child slugs to their parent", () => {
    assert.equal(directoryParentSlug("butor", DIRECTORY_CATALOG), "vasarlas");
    assert.equal(directoryParentSlug("vasarlas", DIRECTORY_CATALOG), "vasarlas");
});

test("parentCategoryTabActive keeps the parent tab lit for a child slug", () => {
    assert.equal(parentCategoryTabActive("vasarlas", "butor", DIRECTORY_CATALOG), true);
    assert.equal(parentCategoryTabActive("auto", "butor", DIRECTORY_CATALOG), false);
});
