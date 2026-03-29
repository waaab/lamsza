/**
 * Formats a timestamp into relative time (Hungarian)
 * @param {number} ts - Timestamp in milliseconds
 * @returns {string}
 */
export function relativeTime(ts) {
    const diff = Date.now() - ts;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 2) return "most";
    if (minutes < 60) return `${minutes} perce`;
    if (hours < 24) return `${hours} órája`;
    if (days === 1) return "1 napja";
    return `${days} napja`;
}

/**
 * Maps OpenWeatherMap icon codes to emojis (day and night variants).
 * OWM uses 9 condition codes (01–04, 09, 10, 11, 13, 50) × "d" (day) / "n" (night) = 18 codes.
 * WeatherAPI.com and Open-Meteo only return "d" codes; we use local time to show night emoji when appropriate.
 * @param {string} code - OWM icon code (e.g. "01d", "02n")
 * @param {Date} [now] - Optional time for day/night; defaults to current local time
 * @returns {string}
 */
export function weatherIconEmoji(code, now = new Date()) {
    const hour = now.getHours();
    const isNight = hour >= 19 || hour < 6;
    const normalized = (code && isNight && code.endsWith("d"))
        ? code.slice(0, -1) + "n"
        : code;
    const map = {
        "01d": "☀️",
        "01n": "🌙",
        "02d": "⛅",
        "02n": "🌙",
        "03d": "☁️",
        "03n": "☁️",
        "04d": "☁️",
        "04n": "🌑",
        "09d": "🌧️",
        "09n": "🌧️",
        "10d": "🌦️",
        "10n": "🌧️",
        "11d": "⛈️",
        "11n": "⛈️",
        "13d": "❄️",
        "13n": "❄️",
        "50d": "🌫️",
        "50n": "🌫️",
    };
    return map[normalized] ?? map[code] ?? "🌡️";
}

/**
 * Format a date to local string (Hungarian), full format: "2026. 03. 07."
 * @param {number|string|Date} date 
 * @returns {string}
 */
export function formatDate(date) {
    return new Date(date).toLocaleDateString("hu-HU");
}

/**
 * Full Hungarian date plus ", HH:MM" when a time string is provided (e.g. API "12:00:00").
 * @param {string | null | undefined} dateStr
 * @param {string | null | undefined} timeStr
 * @returns {string}
 */
export function formatDateWithOptionalTime(dateStr, timeStr) {
    if (!dateStr) return "";
    const datePart = formatDate(dateStr);
    const t = timeStr != null && String(timeStr).trim();
    if (!t) return datePart;
    const hm = t.length >= 5 ? t.slice(0, 5) : t;
    return `${datePart}, ${hm}`;
}

/**
 * Format a date in short Hungarian: "márc. 7."
 * @param {number|string|Date} date 
 * @returns {string}
 */
export function formatDateShort(date) {
    return new Date(date).toLocaleDateString("hu-HU", { month: "short", day: "numeric" });
}

/**
 * Short month label only, matching {@link formatDateShort} style on cards (e.g. "jún.").
 * @param {number} year
 * @param {number} month1to12
 * @returns {string}
 */
export function formatMonthShortLikeCard(year, month1to12) {
    const iso = `${year}-${String(month1to12).padStart(2, "0")}-15`;
    const s = formatDateShort(iso).trim();
    return s.replace(/\d+\.?\s*$/, "").trim();
}

/**
 * Format a time to local string (Hungarian)
 * @param {number|string|Date} date 
 * @returns {string}
 */
export function formatTime(date) {
    return new Date(date).toLocaleTimeString("hu-HU", {
        hour: "2-digit",
        minute: "2-digit"
    });
}

/** Hungarian long date with weekday — same options as the homepage `#datetime` widget. */
export function formatHuDateLong(d) {
    const x = d instanceof Date ? d : new Date(d);
    return x.toLocaleDateString("hu-HU", {
        year: "numeric",
        month: "long",
        day: "numeric",
        weekday: "long",
    });
}

/**
 * Local calendar date as YYYY-MM-DD (matches what the user sees on `#datetime`).
 * @param {Date} [d]
 * @returns {string}
 */
export function localCalendarISODate(d = new Date()) {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
}

/**
 * Format a YYYY-MM-DD string as Hungarian long date (local calendar day, same style as `formatHuDateLong`).
 * @param {string} iso
 * @returns {string}
 */
export function formatHuDateLongFromYMD(iso) {
    if (!iso || typeof iso !== "string") return "";
    const [y, mo, da] = iso.split("-").map((x) => parseInt(x, 10));
    if (!y || !mo || !da) return iso;
    try {
        return formatHuDateLong(new Date(y, mo - 1, da));
    } catch {
        return iso;
    }
}

/**
 * Public venue detail URL under a settlement (matches `/[countySlug]-megye/[slug]/helyszin/[venueSlug]`).
 * @param {{ county_slug?: string, location_slug?: string }} eventLike
 * @param {string | null | undefined} venueSlug
 * @returns {string | null}
 */
export function venuePageUrl(eventLike, venueSlug) {
    const s = String(venueSlug ?? "").trim();
    if (!s) return null;
    const cs = String(eventLike?.county_slug ?? "").trim();
    const ls = String(eventLike?.location_slug ?? "").trim();
    if (!cs || !ls) return null;
    return `/${cs}-megye/${ls}/helyszin/${s}`;
}
