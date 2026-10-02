import assert from "node:assert/strict";
import { test } from "node:test";
import {
    canonicalEntryCategory,
    canonicalEntryCategoryKey,
    directoryCategoryTabs,
    entryMatchesCategory,
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

test("directoryCategoryTabs uses slugs in /index URLs", () => {
    const tabs = directoryCategoryTabs([
        { category: "Bútor" },
        { category: "Bútor" },
        { category: "Bank" },
        { category: "" },
    ]);
    assert.equal(tabs[0].id, "osszes");
    const ids = tabs.map((t) => t.id);
    assert.ok(ids.includes("butor"));
    assert.ok(ids.includes("bank"));
    assert.equal(tabs.filter((t) => t.id === "butor").length, 1);
    const furniture = tabs.find((t) => t.id === "butor");
    assert.equal(furniture.label, "Bútor");
    assert.equal(furniture.url, "/index/butor");
});
