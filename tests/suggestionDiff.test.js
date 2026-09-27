import { test } from "node:test";
import assert from "node:assert/strict";
import { diffSuggestionFields } from "../src/lib/suggestionDiff.js";

test("diffSuggestionFields omits unchanged fields and keeps a cleared phone", () => {
    const got = diffSuggestionFields(
        { name: "Régi", phone: "111", notes: "szöveg" },
        { name: "Régi", phone: "", notes: "szöveg" },
    );
    assert.deepEqual(got, { phone: "" });
});
