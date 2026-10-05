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
    directoryCatalogFromApi,
    directoryGroups,
    directoryGroupsForEntries,
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

test("entryMatchesCategory matches every shelf on the listing", () => {
    const entry = {
        category: "Webfejlesztés",
        categories: ["Webfejlesztés", "Webdizájn", "E-kereskedelem", "Tanácsadás"],
    };
    assert.equal(entryMatchesCategory(entry, "webfejlesztes"), true);
    assert.equal(entryMatchesCategory(entry, "webdizajn"), true);
    assert.equal(entryMatchesCategory(entry, "e-kereskedelem"), true);
    assert.equal(entryMatchesCategory(entry, "tanacsadas"), true);
    assert.equal(entryMatchesCategory(entry, "informatika-es-tavkozles"), true);
    assert.equal(entryMatchesCategory(entry, "szakmai-szolgaltatasok"), true);
    assert.equal(entryMatchesCategory(entry, "auto"), false);
});

test("entryMatchesCategory matches parent slug for a child category", () => {
    assert.equal(entryMatchesCategory({ category: "Bútor" }, "vasarlas"), true);
    assert.equal(entryMatchesCategory({ category: "Turbószerviz" }, "auto"), true);
    assert.equal(entryMatchesCategory({ category: "Bútor" }, "auto"), false);
});

test("directoryCategoryTabs lists every main category", () => {
    const tabs = directoryCategoryTabs(
        DIRECTORY_CATALOG,
        [{ category: "Bútor" }],
        [{ category: "Bank", entry_id: 0 }],
    );
    assert.equal(tabs[0].id, "osszes");
    assert.equal(tabs.length, 13);
    assert.equal(tabs[1].id, "etkezes");
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

test("directoryGroupsForEntries hides subcategories with no entry", () => {
    const groups = directoryGroupsForEntries(DIRECTORY_CATALOG, [{ category: "Bútor" }]);
    const shopping = groups.find((group) => group.slug === "vasarlas");
    const food = groups.find((group) => group.slug === "etkezes");
    assert.deepEqual(shopping.children.map((child) => child.name), ["Bútor"]);
    assert.equal(food.children.length, 0);
    assert.equal(groups.length, 12);
});

test("directoryGroups lists every parent with its subcategories", () => {
    const groups = directoryGroups(DIRECTORY_CATALOG);
    assert.equal(groups.length, 12);
    const food = groups.find((group) => group.slug === "etkezes");
    assert.deepEqual(
        food.children.map((child) => child.name),
        ["Étterem", "Kávézó", "Cukrászda", "Pékség", "Söröző"],
    );
    assert.equal(groups.some((group) => group.name === "Sportpálya"), false);
});

test("directoryCatalogFromApi keeps an admin-added child on its parent shelf", () => {
    const catalog = directoryCatalogFromApi([
        { id: 1, name: "Étkezés", slug: "etkezes", parent_id: null, sort_order: 1 },
        { id: 99, name: "Reggeli bár", slug: "reggeli-bar", parent_id: 1, sort_order: 9 },
    ]);
    assert.equal(entryMatchesCategory({ category: "Reggeli bár" }, "etkezes", catalog), true);
    const tabs = directoryCategoryTabs(catalog, [{ category: "Reggeli bár" }], []);
    assert.equal(tabs.some((tab) => tab.id === "etkezes"), true);
});
