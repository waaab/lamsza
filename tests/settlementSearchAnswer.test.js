import assert from "node:assert/strict";
import { test } from "node:test";
import {
    buildSettlementAnswer,
    pickSettlement,
    settlementFoundLine,
} from "../src/lib/settlementSearchAnswer.js";

const kezdi = {
    name: "Kézdivásárhely",
    name_ro: "Târgu Secuiesc",
    name_de: "Seklerneumarkt",
    slug: "kezdivasarhely",
    type: "municípium",
    post_code: "53111",
    population: "2353",
    area: "37.5",
};

const oroszfalu = {
    name: "Kézdioroszfalu",
    slug: "kezdioroszfalu",
    type: "falu",
};

test("pickSettlement matches the town by slug, accented name, or other-language name", () => {
    const locations = [kezdi, oroszfalu, { name: "Kovászna", slug: "kovaszna", type: "megye" }];
    assert.equal(pickSettlement(locations, "kezdivasarhely")?.slug, "kezdivasarhely");
    assert.equal(pickSettlement(locations, "Kézdivásárhely")?.slug, "kezdivasarhely");
    assert.equal(pickSettlement(locations, "Târgu Secuiesc")?.slug, "kezdivasarhely");
    assert.equal(pickSettlement(locations, "kezdi"), null);
});

test("buildSettlementAnswer speaks the stored facts and skips empty ones", () => {
    const text = buildSettlementAnswer({
        ...kezdi,
        name_de: "",
        area: "",
    });
    assert.equal(
        text,
        "Kézdivásárhely, románul Târgu Secuiesc. Irányítószám: 53111. Lakosság: 2353 fő. Közigazgatási forma: municípium.",
    );
});

test("buildSettlementAnswer adds weather and ongoing events when they exist", () => {
    const text = buildSettlementAnswer(kezdi, {
        weather: { temp: 5, desc: "köd", emoji: "🌫️" },
        events: [{ title: "Könyvvásár" }, { title: "Vásár" }],
    });
    assert.match(text, /Terület: 37,5 km²\./);
    assert.match(text, /Az időjárás éppen: köd, 5 °C 🌫️\./);
    assert.match(text, /Most zajló események: Könyvvásár, Vásár\./);
});

test("settlementFoundLine names the town", () => {
    assert.equal(settlementFoundLine("Kézdivásárhely"), "Erre találtam Kézdivásárhely kapcsán:");
});
