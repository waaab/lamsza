import { test } from "node:test";
import assert from "node:assert/strict";
import { GAMES, LISTED_GAMES, featuredGames, playsWithSzekelyWords, todaysChallenges } from "../src/lib/games/catalog.js";

test("the games played with Szótár's Székely words; Rovásfejtő is not one", () => {
    const slugs = GAMES.filter(playsWithSzekelyWords).map((g) => g.slug);
    assert.deepEqual(slugs, ["kaptar", "szorejto", "szokereso", "akasztofa", "tajszorejtveny"]);
    assert.equal(playsWithSzekelyWords({ slug: "rovasfejto" }), false);
    assert.equal(playsWithSzekelyWords({ slug: "unknown" }), false);
});

test("every released game is listed from the first paint, Tájszórejtvény too", () => {
    assert.equal(LISTED_GAMES.length, GAMES.length);
    assert.ok(LISTED_GAMES.some((g) => g.slug === "tajszorejtveny"));
});

test("featuredGames keeps the flag, so a page can filter the server's list", () => {
    const answer = { games: [
        { slug: "rovasfejto", enabled: true, sort_order: 0 },
        { slug: "kaptar", enabled: true, sort_order: 1 },
        { slug: "szokereso", enabled: false, sort_order: 2 },
    ] };
    const list = featuredGames(answer, Infinity);
    assert.deepEqual(list.map((g) => g.slug), ["rovasfejto", "kaptar"]);
    assert.deepEqual(list.filter(playsWithSzekelyWords).map((g) => g.slug), ["kaptar"]);
});

test("todaysChallenges keeps only the games with a daily today, in Játszótér's order", () => {
    const answer = {
        games: [
            { slug: "szokereso", enabled: true, sort_order: 3, daily_today: false },
            { slug: "kaptar", enabled: true, sort_order: 1, daily_today: true },
            { slug: "rovasfejto", enabled: true, sort_order: 0, daily_today: true },
            { slug: "kviz", enabled: true, sort_order: 2, daily_today: true },
        ],
    };
    assert.deepEqual(todaysChallenges(answer).map((g) => g.slug), ["rovasfejto", "kaptar"]);
});

test("todaysChallenges: none today is an empty list; no games or an old answer are told apart", () => {
    const none = { games: [{ slug: "szokereso", enabled: true, daily_today: false }] };
    assert.deepEqual(todaysChallenges(none), []);
    assert.equal(todaysChallenges({ games: [] }), null);
    assert.equal(todaysChallenges(null), null);
    // A Játszótér without the field yet: every game counts.
    const old = { games: [{ slug: "kaptar", enabled: true }, { slug: "akasztofa", enabled: true }] };
    assert.deepEqual(todaysChallenges(old).map((g) => g.slug), ["kaptar", "akasztofa"]);
});
