import { test } from "node:test";
import assert from "node:assert/strict";
import { smoothPath, layoutCurve } from "../src/lib/tempCurve.js";

const box = { pitch: 3.85, card: 3.6, top: 1.4, bottom: 4.6 };

/** Hourly steps from a UTC start, one temperature each. */
function hours(startIso, temps) {
    const t0 = Date.parse(startIso);
    return temps.map((temp, i) => ({ time: new Date(t0 + i * 3600e3).toISOString(), temp }));
}

/** The y of every cubic segment's control points, from path data. */
function controlYs(d) {
    return [...d.matchAll(/C ([-\d.]+) ([-\d.]+) ([-\d.]+) ([-\d.]+)/g)].flatMap((m) => [Number(m[2]), Number(m[4])]);
}

test("smoothPath goes through every point", () => {
    const pts = [{ x: 0, y: 3 }, { x: 1, y: 1 }, { x: 2, y: 2 }];
    const d = smoothPath(pts);
    assert.match(d, /^M 0 3 C .* 1 1 C .* 2 2$/);
    assert.equal(smoothPath([]), "");
    assert.equal(smoothPath([{ x: 1, y: 2 }]), "M 1 2");
});

test("smoothPath never overshoots: control points stay within the data's range", () => {
    const pts = [0, 0, 0, 5, 5, 5, 1, 9, 9].map((y, x) => ({ x, y }));
    for (const y of controlYs(smoothPath(pts))) {
        assert.ok(y >= 0 - 1e-9 && y <= 9 + 1e-9, `control y ${y} leaves 0..9`);
    }
});

test("layoutCurve puts each card's point at its centre, warmest highest", () => {
    const { points, lo, hi } = layoutCurve(hours("2026-10-09T09:00:00Z", [10, 15, 20]), box);
    assert.deepEqual(points.map((p) => p.x), [1.8, 5.65, 9.5]);
    assert.equal(lo, 10);
    assert.equal(hi, 20);
    assert.equal(points[2].y, box.top);
    assert.equal(points[0].y, box.bottom);
});

test("a flat day stays flat: the scale spans at least 4 degrees", () => {
    const { points, lo, hi } = layoutCurve(hours("2026-10-09T09:00:00Z", [10, 11, 10]), box);
    assert.equal(hi - lo, 4);
    assert.ok(points.every((p) => p.y > box.top && p.y < box.bottom));
});

test("each day's high and low are marked, except at the strip's ends", () => {
    // From 01:00Z, hour 20 (21:00Z) is midnight in Bucharest.
    const temps = [9, 8, 7, 8, 12, 18, 21, 19, 15, 12, 11, 10, 9, 8, 6, 5, 7, 11, 16, 20, 22, 21, 17, 13, 12, 11];
    const h = hours("2026-10-09T01:00:00Z", temps);
    const { points, midnights } = layoutCurve(h, box);
    const marked = points.filter((p) => p.extreme).map((p) => [p.day, p.extreme, p.temp]);
    // Day 2's low (11) is the strip's last hour, so it is left unmarked.
    assert.deepEqual(marked, [
        ["2026-10-09", "high", 21],
        ["2026-10-09", "low", 5],
        ["2026-10-10", "high", 22],
    ]);
    assert.deepEqual(midnights.map((m) => [m.i, m.day]), [[20, "2026-10-10"]]);
});

test("a high at the very first or last hour is not marked", () => {
    const { points } = layoutCurve(hours("2026-10-09T09:00:00Z", [20, 15, 10, 12, 14]), box);
    assert.equal(points[0].extreme, null);
    assert.equal(points.find((p) => p.extreme === "low")?.temp, 10);
});

test("hours with no temperature are skipped, not drawn at zero", () => {
    const { points, path } = layoutCurve(hours("2026-10-09T09:00:00Z", [10, null, 12]), box);
    assert.deepEqual(points.map((p) => p.i), [0, 2]);
    assert.ok(path.startsWith("M 1.8 "));
    assert.deepEqual(layoutCurve([], box).points, []);
});
