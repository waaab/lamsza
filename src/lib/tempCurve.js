/**
 * The temperature line over the hour-by-hour strip (HourlyStrip.svelte):
 * one point per hour card, a smooth path through them, and each day's high
 * and low. Pure functions, so the geometry is tested without a browser.
 */
import { bucharestYMD } from "./bucharestTime.js";

/**
 * Monotone cubic interpolation (Fritsch-Carlson): the curve goes through
 * every point and never bulges past a neighbour, so it never shows a high or
 * a low that the forecast does not have.
 * @param {{ x: number, y: number }[]} pts
 * @returns {string} SVG path data, "" for no points
 */
export function smoothPath(pts) {
    const n = pts.length;
    if (n === 0) return "";
    const f = (/** @type {number} */ v) => Math.round(v * 1000) / 1000;
    if (n === 1) return `M ${f(pts[0].x)} ${f(pts[0].y)}`;
    const dx = [];
    const slope = [];
    for (let i = 0; i < n - 1; i++) {
        dx[i] = pts[i + 1].x - pts[i].x;
        slope[i] = (pts[i + 1].y - pts[i].y) / dx[i];
    }
    const m = new Array(n);
    m[0] = slope[0];
    m[n - 1] = slope[n - 2];
    for (let i = 1; i < n - 1; i++) {
        m[i] = slope[i - 1] * slope[i] <= 0 ? 0 : (slope[i - 1] + slope[i]) / 2;
    }
    for (let i = 0; i < n - 1; i++) {
        if (slope[i] === 0) {
            m[i] = 0;
            m[i + 1] = 0;
            continue;
        }
        const a = m[i] / slope[i];
        const b = m[i + 1] / slope[i];
        const h = a * a + b * b;
        if (h > 9) {
            const t = 3 / Math.sqrt(h);
            m[i] = t * a * slope[i];
            m[i + 1] = t * b * slope[i];
        }
    }
    let d = `M ${f(pts[0].x)} ${f(pts[0].y)}`;
    for (let i = 0; i < n - 1; i++) {
        const c1x = pts[i].x + dx[i] / 3;
        const c1y = pts[i].y + (m[i] * dx[i]) / 3;
        const c2x = pts[i + 1].x - dx[i] / 3;
        const c2y = pts[i + 1].y - (m[i + 1] * dx[i]) / 3;
        d += ` C ${f(c1x)} ${f(c1y)} ${f(c2x)} ${f(c2y)} ${f(pts[i + 1].x)} ${f(pts[i + 1].y)}`;
    }
    return d;
}

/**
 * @typedef {{
 *   i: number, x: number, y: number, temp: number,
 *   extreme: null | "high" | "low", day: string,
 * }} CurvePoint
 */

/**
 * Lays the hours out on the strip. Card i's centre is at
 * i * pitch + card / 2; temperatures map onto [top, bottom] (warmest at
 * top). Each Bucharest day's warmest and coldest hour is marked, unless it
 * sits at either end of the strip, where the real high or low may lie
 * outside the 48 hours.
 *
 * @param {{ time: string, temp?: number | null }[]} hours
 * @param {{ pitch: number, card: number, top: number, bottom: number }} box
 */
export function layoutCurve(hours, box) {
    /** @type {CurvePoint[]} */
    const points = [];
    hours.forEach((h, i) => {
        if (h.temp == null || Number.isNaN(Number(h.temp))) return;
        points.push({
            i,
            x: i * box.pitch + box.card / 2,
            y: 0,
            temp: Number(h.temp),
            extreme: null,
            day: bucharestYMD(new Date(h.time).getTime()),
        });
    });
    if (points.length === 0) {
        return { points, path: "", area: "", lo: 0, hi: 0, midnights: [] };
    }
    const temps = points.map((p) => p.temp);
    let lo = Math.min(...temps);
    let hi = Math.max(...temps);
    if (hi - lo < 4) {
        // A flat day should look flat, not like a storm of ups and downs.
        const mid = (hi + lo) / 2;
        lo = mid - 2;
        hi = mid + 2;
    }
    for (const p of points) {
        p.y = box.top + ((hi - p.temp) / (hi - lo)) * (box.bottom - box.top);
    }

    /** @type {Map<string, CurvePoint[]>} */
    const byDay = new Map();
    for (const p of points) {
        if (!byDay.has(p.day)) byDay.set(p.day, []);
        byDay.get(p.day)?.push(p);
    }
    const firstI = points[0].i;
    const lastI = points[points.length - 1].i;
    for (const dayPts of byDay.values()) {
        if (dayPts.length < 3) continue;
        let high = dayPts[0];
        let low = dayPts[0];
        for (const p of dayPts) {
            if (p.temp > high.temp) high = p;
            if (p.temp < low.temp) low = p;
        }
        if (high.temp === low.temp) continue;
        if (high.i !== firstI && high.i !== lastI) high.extreme = "high";
        if (low.i !== firstI && low.i !== lastI) low.extreme = "low";
    }

    /** Where a new Bucharest day starts: the edge before its first card. */
    const midnights = [];
    for (let i = 1; i < hours.length; i++) {
        const a = bucharestYMD(new Date(hours[i - 1].time).getTime());
        const b = bucharestYMD(new Date(hours[i].time).getTime());
        if (a !== b) midnights.push({ i, x: i * box.pitch - (box.pitch - box.card) / 2, day: b });
    }

    const path = smoothPath(points);
    const last = points[points.length - 1];
    const area = `${path} L ${last.x} ${box.bottom} L ${points[0].x} ${box.bottom} Z`;
    return { points, path, area, lo, hi, midnights };
}
