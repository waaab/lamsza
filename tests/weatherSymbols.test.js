import { test } from "node:test";
import assert from "node:assert/strict";
import {
    SYMBOL_BASES,
    parseSymbol,
    symbolParts,
    symbolEmoji,
    moonLitPath,
    moonPhaseName,
    compassHU,
    uvLevel,
    sourceCredit,
} from "../src/lib/weatherSymbols.js";

test("there are MET's 41 base codes, each once", () => {
    assert.equal(SYMBOL_BASES.length, 41);
    assert.equal(new Set(SYMBOL_BASES).size, 41);
    assert.ok(SYMBOL_BASES.includes("lightsnowshowersandthunder"));
    assert.ok(SYMBOL_BASES.includes("heavysleet"));
});

test("parseSymbol reads intensity, type, showers, thunder and night", () => {
    assert.deepEqual(parseSymbol("heavyrainshowersandthunder_night"), {
        base: "heavyrainshowersandthunder",
        night: true,
        sky: "partlycloudy",
        precip: "rain",
        intensity: "heavy",
        showers: true,
        thunder: true,
    });
    const s = parseSymbol("lightsnow");
    assert.equal(s.precip, "snow");
    assert.equal(s.intensity, "light");
    assert.equal(s.sky, "cloudy");
    assert.equal(s.night, false);
});

test("parseSymbol fixes MET's two misspelt codes and survives junk", () => {
    assert.equal(parseSymbol("lightssnowshowersandthunder_day").base, "lightsnowshowersandthunder");
    assert.equal(parseSymbol("lightssleetshowersandthunder").base, "lightsleetshowersandthunder");
    assert.equal(parseSymbol("tornado").base, "cloudy");
    assert.equal(parseSymbol(undefined).base, "cloudy");
});

test("every code draws something and has an emoji", () => {
    for (const base of SYMBOL_BASES) {
        for (const code of [base, `${base}_day`, `${base}_night`]) {
            const p = symbolParts(code);
            assert.ok(p.body || p.clouds || p.fog, `${code} draws nothing`);
            assert.ok(symbolEmoji(code), `${code} has no emoji`);
            if (p.precip) assert.ok(p.drops >= 2 && p.drops <= 4, `${code} drops ${p.drops}`);
        }
    }
});

test("symbolParts picks the sun or the moon", () => {
    assert.equal(symbolParts("clearsky_day").body, "sun");
    assert.equal(symbolParts("clearsky_night").body, "moon");
    assert.equal(symbolParts("clearsky_night").stars, true);
    assert.equal(symbolParts("cloudy").body, null);
    assert.equal(symbolParts("cloudy").clouds, 2);
    assert.equal(symbolParts("rainshowers_night").body, "moon");
    assert.equal(symbolParts("heavyrain").dark, true);
    assert.equal(symbolParts("lightrainshowers_day").dark, false);
    assert.equal(symbolParts("fog").fog, true);
    assert.equal(symbolParts("heavysnow").drops, 4);
});

test("moonLitPath: nothing at new moon, a full disc at full moon", () => {
    assert.equal(moonLitPath(0, 10, 10, 8), "");
    assert.equal(moonLitPath(360, 10, 10, 8), "");
    assert.match(moonLitPath(180, 10, 10, 8), /0 1 1/);
});

test("moonLitPath: waxing lights the right, waning the left", () => {
    // Waxing crescent: outer arc through the right (sweep 1), terminator bulging right (sweep 0).
    assert.match(moonLitPath(45, 10, 10, 8), /^M 10 2 A 8 8 0 0 1 10 18 A [\d.]+ 8 0 0 0 10 2 Z$/);
    // Waxing gibbous: terminator bulges left.
    assert.match(moonLitPath(135, 10, 10, 8), /A 8 8 0 0 1 10 18 A [\d.]+ 8 0 0 1 10 2 Z$/);
    // Waning gibbous and crescent: outer arc through the left.
    assert.match(moonLitPath(225, 10, 10, 8), /A 8 8 0 0 0 10 18 A [\d.]+ 8 0 0 0 10 2 Z$/);
    assert.match(moonLitPath(338.85, 10, 10, 8), /A 8 8 0 0 0 10 18 A [\d.]+ 8 0 0 1 10 2 Z$/);
});

test("moon phase names", () => {
    assert.equal(moonPhaseName(0), "újhold");
    assert.equal(moonPhaseName(90), "első negyed");
    assert.equal(moonPhaseName(180), "telihold");
    assert.equal(moonPhaseName(270), "utolsó negyed");
    assert.equal(moonPhaseName(338.85), "fogyó holdsarló");
});

test("compassHU names where the wind comes from", () => {
    assert.equal(compassHU(0), "É");
    assert.equal(compassHU(359), "É");
    assert.equal(compassHU(190), "D");
    assert.equal(compassHU(225), "DNy");
    assert.equal(compassHU(300), "ÉNy");
    assert.equal(compassHU(null), "");
});

test("uvLevel uses the WHO bands", () => {
    assert.equal(uvLevel(0).label, "alacsony");
    assert.equal(uvLevel(4.25).label, "mérsékelt");
    assert.equal(uvLevel(7).level, 2);
    assert.equal(uvLevel(11).label, "extrém");
});

test("sourceCredit gives each licence's wording", () => {
    assert.equal(sourceCredit("metno").href, "https://api.met.no/doc/License");
    assert.equal(sourceCredit("weatherapi_com").linkText, "WeatherAPI.com");
    assert.equal(sourceCredit("openweathermap").linkText, "OpenWeather");
});
