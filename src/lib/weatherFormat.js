/**
 * Formatting for the weather pages. Times are shown in Bucharest time
 * whatever the visitor's clock says (WAYS_OF_WORKING R19).
 */
import { ZONE, bucharestYMD } from "./bucharestTime.js";

const hourFmt = new Intl.DateTimeFormat("hu-HU", { timeZone: ZONE, hour: "2-digit", minute: "2-digit", hourCycle: "h23" });
const weekdayFmt = new Intl.DateTimeFormat("hu-HU", { timeZone: "UTC", weekday: "long" });
const weekdayShortFmt = new Intl.DateTimeFormat("hu-HU", { timeZone: "UTC", weekday: "short" });
const monthDayFmt = new Intl.DateTimeFormat("hu-HU", { timeZone: "UTC", month: "short", day: "numeric" });

/**
 * "07:25" in Bucharest time.
 * @param {string | number | Date | null | undefined} t
 */
export function fmtClock(t) {
    if (t == null || t === "") return "";
    const d = new Date(t);
    return Number.isNaN(d.getTime()) ? "" : hourFmt.format(d);
}

/** @param {string} ymd "2026-10-09" */
function ymdToUtcDate(ymd) {
    const [y, m, d] = ymd.split("-").map(Number);
    return new Date(Date.UTC(y, m - 1, d));
}

/**
 * "Ma", "Holnap" or the weekday ("szombat").
 * @param {string} ymd
 * @param {number} [nowMs]
 */
export function dayName(ymd, nowMs = Date.now()) {
    const today = bucharestYMD(nowMs);
    if (ymd === today) return "Ma";
    if (ymd === bucharestYMD(nowMs + 24 * 3600 * 1000)) return "Holnap";
    const w = weekdayFmt.format(ymdToUtcDate(ymd));
    return w.charAt(0).toUpperCase() + w.slice(1);
}

/**
 * "szo" style short weekday.
 * @param {string} ymd
 */
export function dayNameShort(ymd) {
    return weekdayShortFmt.format(ymdToUtcDate(ymd)).replace(/\.$/, "");
}

/**
 * "okt. 9."
 * @param {string} ymd
 */
export function monthDay(ymd) {
    return monthDayFmt.format(ymdToUtcDate(ymd));
}

/**
 * A temperature as "8°", or "-" when missing.
 * @param {number | null | undefined} v
 */
export function deg(v) {
    if (v == null || Number.isNaN(Number(v))) return "-";
    const r = Math.round(Number(v));
    return `${Object.is(r, -0) ? 0 : r}°`;
}

/**
 * Minutes between two instants as "11 óra 16 perc".
 * @param {string | null | undefined} from
 * @param {string | null | undefined} to
 */
export function duration(from, to) {
    if (!from || !to) return "";
    const mins = Math.round((new Date(to).getTime() - new Date(from).getTime()) / 60000);
    if (!Number.isFinite(mins) || mins <= 0) return "";
    return `${Math.floor(mins / 60)} óra ${mins % 60} perc`;
}

/**
 * The hour of a step in Bucharest time ("14").
 * @param {string} t
 */
export function hourLabel(t) {
    return fmtClock(t).slice(0, 2);
}
