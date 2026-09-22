import assert from "node:assert/strict";
import { test } from "node:test";
import {
    ENTRY_CATEGORY_BOLT,
    ENTRY_CATEGORY_EGESZSEGUGY,
    ENTRY_CATEGORY_EGYEB,
    ENTRY_CATEGORY_HIVATALOK,
    canonicalEntryCategory,
    canonicalEntryCategoryKey,
    directoryCategoryTabs,
    entryMatchesCategory,
} from "../src/lib/entryCategory.js";

test("canonicalEntryCategory maps Index slugs and aliases", () => {
    assert.equal(canonicalEntryCategory("egeszsegugy"), ENTRY_CATEGORY_EGESZSEGUGY);
    assert.equal(canonicalEntryCategory("Egészségügy"), ENTRY_CATEGORY_EGESZSEGUGY);
    assert.equal(canonicalEntryCategory("Orvosi rendelők"), ENTRY_CATEGORY_EGESZSEGUGY);
    assert.equal(canonicalEntryCategory("hivatalok"), ENTRY_CATEGORY_HIVATALOK);
    assert.equal(canonicalEntryCategory("városháza"), ENTRY_CATEGORY_HIVATALOK);
    assert.equal(canonicalEntryCategory("bolt"), ENTRY_CATEGORY_BOLT);
    assert.equal(canonicalEntryCategory("egyeb"), ENTRY_CATEGORY_EGYEB);
    assert.equal(canonicalEntryCategory(""), "");
});

test("entryMatchesCategory matches URL slug against catalog label", () => {
    assert.equal(
        entryMatchesCategory({ category: "Egészségügy" }, "egeszsegugy"),
        true,
    );
    assert.equal(
        entryMatchesCategory({ category: "egeszsegugy" }, "egeszsegugy"),
        true,
    );
    assert.equal(entryMatchesCategory({ category: "Oktatás" }, "hivatalok"), false);
    assert.equal(entryMatchesCategory({ category: "Hivatalok" }, "osszes"), true);
});

test("directoryCategoryTabs uses slugs in /index URLs", () => {
    const tabs = directoryCategoryTabs([
        { category: "egeszsegugy" },
        { category: "Egészségügy" },
        { category: "hivatalok" },
        { category: "" },
    ]);
    assert.equal(tabs[0].id, "osszes");
    const ids = tabs.map((t) => t.id);
    assert.ok(ids.includes("egeszsegugy"));
    assert.ok(ids.includes("hivatalok"));
    assert.equal(tabs.filter((t) => t.id === "egeszsegugy").length, 1);
    const health = tabs.find((t) => t.id === "egeszsegugy");
    assert.equal(health.label, ENTRY_CATEGORY_EGESZSEGUGY);
    assert.equal(health.url, "/index/egeszsegugy");
});
