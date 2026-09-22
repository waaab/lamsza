import assert from "node:assert/strict";
import { test } from "node:test";
import { weatherIconEmoji } from "../src/lib/utils.js";

const day = new Date(2026, 8, 22, 12, 0);
const night = new Date(2026, 8, 22, 19, 30);

test("clear sky: sun by day, moon by night", () => {
    assert.equal(weatherIconEmoji("01d", day), "☀️");
    assert.equal(weatherIconEmoji("01d", night), "🌙");
    assert.equal(weatherIconEmoji("01n", night), "🌙");
});

test("cloudy conditions keep a cloud at night, not a moon", () => {
    assert.equal(weatherIconEmoji("02d", night), "☁️");
    assert.equal(weatherIconEmoji("03d", night), "☁️");
    assert.equal(weatherIconEmoji("04d", night), "☁️");
    assert.equal(weatherIconEmoji("02n", night), "☁️");
    assert.equal(weatherIconEmoji("04n", night), "☁️");
});

test("few clouds by day stay a sun-cloud", () => {
    assert.equal(weatherIconEmoji("02d", day), "⛅");
    assert.equal(weatherIconEmoji("03d", day), "☁️");
    assert.equal(weatherIconEmoji("04d", day), "☁️");
});
