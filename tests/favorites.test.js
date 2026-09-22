import assert from "node:assert/strict";
import test from "node:test";
import { favoriteKey, isFavorite } from "../src/lib/favorites.js";

test("isFavorite matches type and id", () => {
    const list = [{ type: "event", id: 9 }];
    assert.equal(isFavorite(list, "event", 9), true);
    assert.equal(isFavorite(list, "event", 3), false);
    assert.equal(favoriteKey("attraction", 4), "attraction:4");
});
