/**
 * The network's time zone for everything that depends on "today" (lamsza
 * WAYS_OF_WORKING R19): Europe/Bucharest, never the visitor's zone or clock.
 *
 * - Event dates and times are Bucharest wall-clock times; bucharestWallToMs
 *   turns them into instants.
 * - "Now" comes from the server: every API reply carries an HTTP Date header,
 *   and noteServerDate keeps the difference to this device's clock, so a wrong
 *   visitor clock does not move a day boundary. serverNowMs applies it.
 * - bucharestYMD gives the Bucharest calendar day of an instant.
 */

export const ZONE = "Europe/Bucharest";

const parts = new Intl.DateTimeFormat("en-GB", {
    timeZone: ZONE,
    hourCycle: "h23",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
});

/** @param {number} ms @returns {Record<string, string>} */
function partsAt(ms) {
    return Object.fromEntries(parts.formatToParts(new Date(ms)).map((p) => [p.type, p.value]));
}

/** Bucharest's offset from UTC at an instant, in ms (+3h in summer, +2h in winter). */
function offsetAt(/** @type {number} */ ms) {
    const p = partsAt(ms);
    const wall = Date.UTC(+p.year, +p.month - 1, +p.day, +p.hour, +p.minute, +p.second);
    return wall - Math.floor(ms / 1000) * 1000;
}

/**
 * The instant of a Bucharest wall-clock time. In the spring gap (03:00-03:59 on
 * the change day) it lands an hour later, as a clock that jumps would.
 */
export function bucharestWallToMs(y, mo, d, hh = 0, mm = 0, ss = 0, msPart = 0) {
    const wall = Date.UTC(y, mo - 1, d, hh, mm, ss, msPart);
    const first = wall - offsetAt(wall);
    return wall - offsetAt(first);
}

/** The Bucharest calendar day of an instant, as YYYY-MM-DD. */
export function bucharestYMD(/** @type {number} */ ms) {
    const p = partsAt(ms);
    return `${p.year}-${p.month}-${p.day}`;
}

let skewMs = 0;

/**
 * Remember the server's time from an HTTP Date header (second precision). The
 * API replies call it; anything that is not a date is ignored.
 * @param {string | null | undefined} header
 */
export function noteServerDate(header) {
    const t = Date.parse(String(header || ""));
    if (Number.isNaN(t)) return;
    // The header is truncated to the second: assume the middle of it.
    skewMs = t + 500 - Date.now();
}

/** The server's current instant, as far as this page knows it. */
export function serverNowMs() {
    return Date.now() + skewMs;
}
