import { bucharestWallToMs, bucharestYMD, serverNowMs } from "./bucharestTime.js";

/**
 * Event timing vs “now” (lamsza WAYS_OF_WORKING R19): event dates and times are
 * Bucharest wall-clock times, and “now” and “today” come from the server's
 * clock in Bucharest, not from the visitor's zone or clock.
 *
 * @typedef {'scheduled' | 'upcoming' | 'ongoing' | 'ending_soon' | 'ended'} EventStatus
 * @typedef {{ start_date?: string, end_date?: string, start_time?: string, end_time?: string }} EventLike
 */

/** Last 48h of the window show “ending soon” (still ongoing). */
const ENDING_SOON_MS = 48 * 60 * 60 * 1000;

/** “Hamarosan” only if the event starts within this window; otherwise “Betervezett”. */
const WITHIN_ONE_MONTH_MS = 30 * 24 * 60 * 60 * 1000;

/** @param {string | undefined} t */
function normalizeTimeStr(t) {
    if (!t || typeof t !== "string") return "00:00:00";
    const s = t.trim();
    if (s.length >= 8 && s.includes(":")) return s.slice(0, 8);
    if (s.length === 5 && s[2] === ":") return `${s}:00`;
    return "00:00:00";
}

/** @param {string | undefined} t */
function isMidnightTime(t) {
    const n = normalizeTimeStr(t);
    return n === "00:00:00" || n.startsWith("00:00:");
}

/** Instant (ms) of a Bucharest date (YYYY-MM-DD) and wall-clock time. */
function parseLocalInstantMs(dateStr, timeStr) {
    if (!dateStr || typeof dateStr !== "string") return null;
    const d = dateStr.trim().slice(0, 10);
    const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(d);
    if (!m) return null;
    const t = normalizeTimeStr(timeStr);
    const [hh, mm, ss] = t.split(":").map((x) => parseInt(x, 10));
    return bucharestWallToMs(Number(m[1]), Number(m[2]), Number(m[3]), hh || 0, mm || 0, ss || 0);
}

/** End of a Bucharest calendar day, inclusive. */
function endOfLocalDayMs(yyyyMmDd) {
    const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(yyyyMmDd.trim().slice(0, 10));
    if (!m) return null;
    return bucharestWallToMs(Number(m[1]), Number(m[2]), Number(m[3]), 23, 59, 59, 999);
}

/**
 * Resolves start/end instants. If the event spans multiple days and the stored
 * end time is exactly midnight, treat that as “through that whole last day”
 * (end = 23:59:59.999 Bucharest time on end_date) - otherwise March 28 00:00 reads as
 * the very start of March 28 and the event wrongly shows as ended by evening.
 *
 * @param {EventLike} ev
 * @returns {{ startMs: number | null, endMs: number | null }}
 */
export function computeEventWindowMs(ev) {
    const sd = ev.start_date?.trim().slice(0, 10);
    const ed = ev.end_date?.trim().slice(0, 10);
    if (!sd || !ed) return { startMs: null, endMs: null };

    let startMs = parseLocalInstantMs(ev.start_date, ev.start_time);
    let endMs = parseLocalInstantMs(ev.end_date, ev.end_time);
    if (startMs == null || endMs == null) return { startMs, endMs };

    const multiDay = ed > sd;
    const sameDay = ed === sd;

    if (multiDay && isMidnightTime(ev.end_time)) {
        const eod = endOfLocalDayMs(ed);
        if (eod != null) endMs = eod;
    } else if (sameDay && isMidnightTime(ev.start_time) && isMidnightTime(ev.end_time)) {
        const eod = endOfLocalDayMs(ed);
        if (eod != null) endMs = eod;
    }

    if (endMs < startMs) {
        endMs = parseLocalInstantMs(ev.end_date, ev.end_time);
    }

    return { startMs, endMs };
}

/**
 * The server's current instant (R19): the visitor's clock, corrected by the
 * Date header of the last API reply.
 * @returns {Date}
 */
export function getReferenceNow() {
    return new Date(serverNowMs());
}

/** Bucharest's calendar “today” (YYYY-MM-DD), for chips like “Ma …”. */
export function referenceTodayYMD() {
    return bucharestYMD(serverNowMs());
}

/**
 * Whether a single schedule activity is in progress at `now` on its day.
 * Uses the end of the Bucharest day when `ends_at` is missing.
 *
 * @param {{ starts_at?: string, ends_at?: string }} act
 * @param {string} scheduleDateKey YYYY-MM-DD
 * @param {Date} [now]
 */
export function isScheduleActivityHappeningNow(act, scheduleDateKey, now) {
    const t = (now ?? getReferenceNow()).getTime();
    const day = String(scheduleDateKey || "").trim().slice(0, 10);
    if (!/^\d{4}-\d{2}-\d{2}$/.test(day)) return false;
    const startMs = parseLocalInstantMs(day, act?.starts_at);
    if (startMs == null) return false;
    let endMs = null;
    if (act?.ends_at && String(act.ends_at).trim()) {
        endMs = parseLocalInstantMs(day, act.ends_at);
    }
    if (endMs == null || endMs < startMs) {
        endMs = endOfLocalDayMs(day);
    }
    if (endMs == null) return false;
    return t >= startMs && t <= endMs;
}

/**
 * @param {EventLike} ev
 * @param {Date} [now]
 * @returns {EventStatus}
 */
export function getEventStatus(ev, now) {
    const t = (now ?? getReferenceNow()).getTime();
    const { startMs, endMs } = computeEventWindowMs(ev);
    if (startMs == null || endMs == null) return "scheduled";

    if (t > endMs) return "ended";
    if (t < startMs) {
        if (startMs - t <= WITHIN_ONE_MONTH_MS) return "upcoming";
        return "scheduled";
    }
    if (endMs - t <= ENDING_SOON_MS) return "ending_soon";
    return "ongoing";
}

/** @type {Record<EventStatus, string>} */
export const EVENT_STATUS_LABELS = {
    scheduled: "Betervezett",
    upcoming: "Hamarosan",
    ongoing: "Folyamatban",
    ending_soon: "Hamarosan véget ér",
    ended: "Lezárult",
};
