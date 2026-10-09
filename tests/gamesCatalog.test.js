import test from "node:test";
import assert from "node:assert/strict";
import { featuredGames, LISTED_GAMES } from "../src/lib/games/catalog.js";

test("featuredGames keeps enabled games in the server's order, with the catalog's look", () => {
    const answer = {
        games: [
            { slug: "kaptar", enabled: true, sort_order: 2, title_hu: "Kaptár!" },
            { slug: "rovasfejto", enabled: true, sort_order: 1 },
            { slug: "szorejto", enabled: false, sort_order: 0 },
            { slug: "ismeretlen", enabled: true, sort_order: 3 },
            { slug: "szokereso", enabled: true, sort_order: 4 },
        ],
    };
    const list = featuredGames(answer, 2);
    assert.deepEqual(list?.map((g) => g.slug), ["rovasfejto", "kaptar"]);
    assert.equal(list?.[1].title_hu, "Kaptár!");
    assert.match(list?.[0].color ?? "", /^#/);
});

test("featuredGames answers null when nothing is listed, so the caller keeps its games", () => {
    assert.equal(featuredGames({ games: [] }, 4), null);
    assert.equal(featuredGames(null, 4), null);
    assert.equal(featuredGames({ games: [{ slug: "kaptar", enabled: false }] }, 4), null);
    assert.ok(LISTED_GAMES.length >= 4);
});
