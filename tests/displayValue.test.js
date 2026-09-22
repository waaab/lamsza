import assert from "node:assert/strict";
import { test } from "node:test";
import {
    EMPTY_PLACEHOLDER,
    displayInitials,
    displayText,
    hasDisplayText,
} from "../src/lib/displayValue.js";

test("displayText: blank values become an em dash", () => {
    assert.equal(displayText(null), EMPTY_PLACEHOLDER);
    assert.equal(displayText(undefined), EMPTY_PLACEHOLDER);
    assert.equal(displayText(""), EMPTY_PLACEHOLDER);
    assert.equal(displayText("   "), EMPTY_PLACEHOLDER);
    assert.equal(displayText([]), EMPTY_PLACEHOLDER);
    assert.equal(displayText(["", " "]), EMPTY_PLACEHOLDER);
});

test("displayText: keeps filled strings and joins arrays", () => {
    assert.equal(displayText("Csíkszereda"), "Csíkszereda");
    assert.equal(displayText(["HU", "RO", "EN"]), "HU, RO, EN");
    assert.equal(hasDisplayText("x"), true);
    assert.equal(hasDisplayText(""), false);
});

test("displayInitials: skips punctuation tokens", () => {
    assert.equal(displayInitials("Manifesto - Csíki Söröző & Étterem"), "MC");
    assert.equal(displayInitials("Dr. Antal Zoltán"), "DA");
    assert.equal(displayInitials("Góbé"), "GÓ");
    assert.equal(displayInitials(""), "?");
});
