import assert from "node:assert/strict";
import { test } from "node:test";
import {
    WEATHER_DESC_PLACEHOLDER_LENGTH,
    fitWeatherDescPlaceholder,
    WEATHER_SOURCE_PLACEHOLDER_LENGTH,
    weatherDescPlaceholder,
    weatherSourcePlaceholder,
} from "../src/lib/weatherDescPlaceholder.js";

test("weather desc placeholder is 18 characters and ends with an ellipsis when cut", () => {
    assert.equal(Array.from("Béborult esment...").length, WEATHER_DESC_PLACEHOLDER_LENGTH);
    assert.equal(Array.from(weatherDescPlaceholder).length, WEATHER_DESC_PLACEHOLDER_LENGTH);
    assert.equal(weatherDescPlaceholder, "Időjárás adatok...");
    assert.equal(fitWeatherDescPlaceholder("Időjárás betöltése"), "Időjárás betöltése");
    assert.equal(fitWeatherDescPlaceholder("Időjárás adatok betöltése"), "Időjárás adatok...");
});

test("weather source placeholder is at least 14 characters", () => {
    assert.ok(Array.from(weatherSourcePlaceholder).length >= WEATHER_SOURCE_PLACEHOLDER_LENGTH);
    assert.equal(weatherSourcePlaceholder, "adat betöltés...");
});
