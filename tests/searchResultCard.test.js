import assert from "node:assert/strict";
import { test } from "node:test";
import { szotarUrl } from "../src/lib/networkOrigins.js";
import { searchResultCardModel } from "../src/lib/searchResultCard.js";
import { formatDateShort } from "../src/lib/utils.js";

test("website: title, domain, description, claim, new tab", () => {
    const card = searchResultCardModel("website", {
        title: "Háromszéki Ágyúsok",
        description: "Közös együttes.",
        domain: "agyusok.ro",
        url: "https://agyusok.ro",
        claimed: true,
    });
    assert.deepEqual(card, {
        href: "https://agyusok.ro",
        title: "Háromszéki Ágyúsok",
        meta: "agyusok.ro",
        description: "Közös együttes.",
        accent: "none",
        external: true,
        claimed: true,
    });
});

test("service: place and category, notes, listing link", () => {
    const card = searchResultCardModel("service", {
        name: "Szent Anna panzió",
        slug: "szent-anna-panzio",
        location: "Csíkszentimre",
        category: "Szállás",
        notes: "Reggeli a tó mellett.",
    });
    assert.equal(card.href, "/bejegyzes/szent-anna-panzio");
    assert.equal(card.title, "Szent Anna panzió");
    assert.equal(card.meta, "Csíkszentimre · Szállás");
    assert.equal(card.description, "Reggeli a tó mellett.");
    assert.equal(card.accent, "none");
    assert.equal(card.external, false);
    assert.equal(card.claimed, false);
});

test("service: claimed listing keeps the claim flag", () => {
    const card = searchResultCardModel("service", {
        name: "Lámsza.com",
        slug: "lamsza-com",
        location: "Kézdivásárhely",
        category: "Egyéb",
        claimed: true,
    });
    assert.equal(card.claimed, true);
});

// The search API sends entry_categories.name, so the card prints the stored name
// and invents nothing. canonicalEntryCategory stopped mapping legacy aliases such
// as "vendeglo" onto catalog labels when the v2 catalog landed (9ae96d4) - the v2
// tree has no "Vendéglő" node at all.
test("service: the meta line prints the stored category name", () => {
    const card = searchResultCardModel("service", {
        name: "Kőröspatak Étterem",
        slug: "korospatak-etterem",
        location: "Sepsiszentgyörgy",
        category: "Étterem",
        notes: "",
    });
    assert.equal(card.meta, "Sepsiszentgyörgy · Étterem");
});

test("service: an unknown category is passed through, not remapped", () => {
    const card = searchResultCardModel("service", {
        name: "Étterem",
        slug: "etterem",
        location: "Sepsiszentgyörgy",
        category: "vendeglo",
        notes: "",
    });
    assert.equal(card.meta, "Sepsiszentgyörgy · vendeglo");
});

test("service: one side of the meta line has no separator", () => {
    const card = searchResultCardModel("service", {
        name: "Panzió",
        slug: "panzio",
        location: "Csíkszentimre",
        category: "",
        notes: "",
    });
    assert.equal(card.meta, "Csíkszentimre");
    assert.equal(card.description, "");
});

test("attraction: county, description, brown accent, county path", () => {
    const card = searchResultCardModel("attraction", {
        name: "  Szent Anna-tó  ",
        slug: "szent-anna-to",
        county_slug: "hargita",
        county_name: "Hargita",
        description: "Vulkanikus tó.",
    });
    assert.equal(card.href, "/hargita-megye/szent-anna-to");
    assert.equal(card.title, "Szent Anna-tó");
    assert.equal(card.meta, "Hargita");
    assert.equal(card.description, "Vulkanikus tó.");
    assert.equal(card.accent, "attraction");
    assert.equal(card.external, false);
});

test("venue: settlement and kind, venue path", () => {
    const card = searchResultCardModel("venue", {
        name: "Szent Anna kápolna",
        slug: "szent-anna-kapolna",
        county_slug: "hargita",
        settlement_slug: "csikszentimre",
        settlement_name: "Csíkszentimre",
        kind_label: "Kápolna",
    });
    assert.equal(card.href, "/hargita-megye/csikszentimre/helyszin/szent-anna-kapolna");
    assert.equal(card.meta, "Csíkszentimre · Kápolna");
    assert.equal(card.description, "");
    assert.equal(card.accent, "none");
});

test("venue: kind without settlement has no separator", () => {
    const card = searchResultCardModel("venue", {
        name: "Kápolna",
        slug: "kapolna",
        county_slug: "hargita",
        settlement_slug: "csikszentimre",
        settlement_name: "",
        kind_label: "Kápolna",
    });
    assert.equal(card.meta, "Kápolna");
});

test("event: short date and place", () => {
    const card = searchResultCardModel("event", {
        id: 12,
        title: "Szent Anna-napi búcsú",
        start_date: "2026-07-26",
        location_name: "Csíkszentimre",
    });
    assert.equal(card.href, "/esemenyek/12");
    assert.equal(card.meta, `${formatDateShort("2026-07-26")} · Csíkszentimre`);
    assert.equal(card.description, "");
    assert.equal(card.accent, "none");
});

test("event: invalid date omits the date side", () => {
    const card = searchResultCardModel("event", {
        id: 12,
        title: "Búcsú",
        start_date: "not-a-date",
        location_name: "Csíkszentimre",
    });
    assert.equal(card.meta, "Csíkszentimre");
});

test("settlement: county and town path", () => {
    const card = searchResultCardModel("settlement", {
        name: "Csíkszentimre",
        slug: "csikszentimre",
        county_slug: "hargita",
        county: "Hargita",
    });
    assert.equal(card.href, "/hargita-megye/csikszentimre");
    assert.equal(card.meta, "Hargita");
    assert.equal(card.description, "");
    assert.equal(card.accent, "none");
});

test("seat: name only, blue accent", () => {
    const card = searchResultCardModel("seat", {
        name: "Csíkszék",
        slug: "csikszek",
    });
    assert.deepEqual(card, {
        href: "/szekek/csikszek",
        title: "Csíkszék",
        meta: "",
        description: "",
        accent: "seat",
        external: false,
        claimed: null,
    });
});

test("news: title and source, new tab", () => {
    const card = searchResultCardModel("news", {
        title: "Kapaszkodik felfelé a tabellán",
        source: "Bihar Napló",
        link: "https://example.com/hir",
    });
    assert.equal(card.href, "https://example.com/hir");
    assert.equal(card.meta, "Bihar Napló");
    assert.equal(card.description, "");
    assert.equal(card.external, true);
    assert.equal(card.claimed, null);
    assert.equal(card.accent, "none");
});

test("word: headword, speech types, definition, szotar link, new tab", () => {
    const card = searchResultCardModel("word", {
        id: 42,
        headword: "  gagya  ",
        definition_hu: "ügyetlen",
        speech_types: ["melléknév", "főnév"],
    });
    assert.deepEqual(card, {
        href: szotarUrl("/szo/42"),
        title: "gagya",
        meta: "melléknév, főnév",
        description: "ügyetlen",
        accent: "none",
        external: true,
        claimed: null,
    });
});

test("word: no speech types omits meta", () => {
    const card = searchResultCardModel("word", {
        id: 7,
        headword: "csángó",
        definition_hu: "",
        speech_types: [],
    });
    assert.equal(card.meta, "");
    assert.equal(card.description, "");
    assert.equal(card.href, szotarUrl("/szo/7"));
});

test("word: missing id or empty headword is skipped", () => {
    assert.equal(searchResultCardModel("word", { id: 0, headword: "x" }), null);
    assert.equal(searchResultCardModel("word", { id: 1, headword: "  " }), null);
    assert.equal(searchResultCardModel("word", { headword: "x" }), null);
});

test("missing link or empty title is skipped", () => {
    assert.equal(searchResultCardModel("website", { title: "Név", url: "" }), null);
    assert.equal(searchResultCardModel("website", { title: "   ", url: "https://example.com" }), null);
    assert.equal(searchResultCardModel("service", { name: "Név", slug: "" }), null);
    assert.equal(searchResultCardModel("attraction", { name: "Tó", slug: "to", county_slug: "" }), null);
    assert.equal(
        searchResultCardModel("venue", {
            name: "Kápolna",
            slug: "kapolna",
            county_slug: "hargita",
            settlement_slug: "",
        }),
        null,
    );
    assert.equal(searchResultCardModel("event", { id: "", title: "Búcsú" }), null);
    assert.equal(searchResultCardModel("settlement", { name: "Falu", slug: "falu", county_slug: "" }), null);
    assert.equal(searchResultCardModel("seat", { name: "Szék", slug: "" }), null);
    assert.equal(searchResultCardModel("news", { title: "Hír", link: "" }), null);
    assert.equal(searchResultCardModel("other", { title: "X" }), null);
});
