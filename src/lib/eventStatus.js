/**
 * Event timing vs “now”, using the same clock as `#datetime` when present.
 *
 * @typedef {'upcoming' | 'ongoing' | 'ending_soon' | 'ended'} EventStatus
 * @typedef {{ start_date?: string, end_date?: string, start_time?: string, end_time?: string }} EventLike
 */

/** Last 48h of the window show “ending soon” (still ongoing). */
const ENDING_SOON_MS = 48 * 60 * 60 * 1000;

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

/** Local instant ms from YYYY-MM-DD + time string. */
function parseLocalInstantMs(dateStr, timeStr) {
    if (!dateStr || typeof dateStr !== "string") return null;
    const d = dateStr.trim().slice(0, 10);
    const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(d);
    if (!m) return null;
    const y = Number(m[1]);
    const mo = Number(m[2]);
    const day = Number(m[3]);
    const t = normalizeTimeStr(timeStr);
    const [hh, mm, ss] = t.split(":").map((x) => parseInt(x, 10));
    const dt = new Date(y, mo - 1, day, hh || 0, mm || 0, ss || 0, 0);
    return dt.getTime();
}

/** End of calendar day (local), inclusive. */
function endOfLocalDayMs(yyyyMmDd) {
    const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(yyyyMmDd.trim().slice(0, 10));
    if (!m) return null;
    const y = Number(m[1]);
    const mo = Number(m[2]);
    const d = Number(m[3]);
    return new Date(y, mo - 1, d, 23, 59, 59, 999).getTime();
}

/**
 * Resolves start/end instants. If the event spans multiple days and the stored
 * end time is exactly midnight, treat that as “through that whole last day”
 * (end = 23:59:59.999 local on end_date) — otherwise March 28 00:00 reads as
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
 * Prefer the homepage clock (`#datetime[data-now-ms]`) when available so badges
 * match the visible time.
 * @returns {Date}
 */
export function getReferenceNow() {
    if (typeof document === "undefined") return new Date();
    const el = document.getElementById("datetime");
    if (el && el.dataset.nowMs) {
        const n = parseInt(el.dataset.nowMs, 10);
        if (!Number.isNaN(n)) return new Date(n);
    }
    return new Date();
}

/**
 * @param {EventLike} ev
 * @param {Date} [now]
 * @returns {EventStatus}
 */
export function getEventStatus(ev, now) {
    const t = (now ?? getReferenceNow()).getTime();
    const { startMs, endMs } = computeEventWindowMs(ev);
    if (startMs == null || endMs == null) return "upcoming";

    if (t > endMs) return "ended";
    if (t < startMs) return "upcoming";
    if (endMs - t <= ENDING_SOON_MS) return "ending_soon";
    return "ongoing";
}

/** @type {Record<EventStatus, string>} */
export const EVENT_STATUS_LABELS = {
    upcoming: "Hamarosan",
    ongoing: "Folyamatban",
    ending_soon: "Hamarosan véget ér",
    ended: "Lezárult",
};
