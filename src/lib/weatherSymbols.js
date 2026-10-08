/**
 * MET Norway symbol codes (the backend maps every provider onto them) and the
 * small helpers the weather icons and pages share: which parts an icon is
 * drawn from, the moon's lit shape, compass and UV words.
 *
 * Code shape: an optional "light"/"heavy", then clearsky | fair | partlycloudy
 * | cloudy | fog, or rain | sleet | snow with an optional "showers" and
 * "andthunder", then "_day" / "_night" for the codes the sun or moon shows
 * through. Examples: "partlycloudy_night", "heavyrainshowersandthunder_day",
 * "lightsnow".
 */

const SKIES = ["clearsky", "fair", "partlycloudy", "cloudy", "fog"];
const PRECIPS = ["rain", "sleet", "snow"];

/** Every base code, in MET's legend order (41). */
export const SYMBOL_BASES = (() => {
    const out = [...SKIES];
    for (const thunder of ["", "andthunder"]) {
        for (const p of PRECIPS) {
            for (const i of ["light", "", "heavy"]) out.push(`${i}${p}showers${thunder}`);
        }
    }
    for (const thunder of ["", "andthunder"]) {
        for (const p of PRECIPS) {
            for (const i of ["light", "", "heavy"]) out.push(`${i}${p}${thunder}`);
        }
    }
    return out;
})();

const TYPOS = {
    lightssleetshowersandthunder: "lightsleetshowersandthunder",
    lightssnowshowersandthunder: "lightsnowshowersandthunder",
};

/**
 * @typedef {{
 *   base: string,
 *   night: boolean,
 *   sky: "clearsky" | "fair" | "partlycloudy" | "cloudy" | "fog",
 *   precip: null | "rain" | "sleet" | "snow",
 *   intensity: "light" | "normal" | "heavy",
 *   showers: boolean,
 *   thunder: boolean,
 * }} ParsedSymbol
 */

/**
 * Reads a symbol code. Anything unknown reads as "cloudy", never as an error.
 * @param {string | null | undefined} code
 * @returns {ParsedSymbol}
 */
export function parseSymbol(code) {
    let raw = String(code ?? "").trim().toLowerCase();
    let variant = "";
    const us = raw.lastIndexOf("_");
    if (us >= 0) {
        variant = raw.slice(us + 1);
        raw = raw.slice(0, us);
    }
    const base = TYPOS[raw] ?? raw;
    const night = variant === "night";
    if (SKIES.includes(base)) {
        return { base, night, sky: /** @type {any} */ (base), precip: null, intensity: "normal", showers: false, thunder: false };
    }
    const m = /^(light|heavy)?(rain|sleet|snow)(showers)?(andthunder)?$/.exec(base);
    if (!m) {
        return { base: "cloudy", night: false, sky: "cloudy", precip: null, intensity: "normal", showers: false, thunder: false };
    }
    const showers = Boolean(m[3]);
    return {
        base,
        night,
        // Showers come and go, so the sun or the moon shows between them.
        sky: showers ? "partlycloudy" : "cloudy",
        precip: /** @type {any} */ (m[2]),
        intensity: m[1] === "light" ? "light" : m[1] === "heavy" ? "heavy" : "normal",
        showers,
        thunder: Boolean(m[4]),
    };
}

/**
 * What an icon is drawn from, back to front.
 * @param {string | null | undefined} code
 * @returns {{
 *   sky: ParsedSymbol["sky"],
 *   body: null | "sun" | "moon",
 *   bodyFull: boolean,
 *   clouds: 0 | 1 | 2,
 *   dark: boolean,
 *   precip: null | "rain" | "sleet" | "snow",
 *   drops: number,
 *   showers: boolean,
 *   thunder: boolean,
 *   fog: boolean,
 *   stars: boolean,
 * }}
 */
export function symbolParts(code) {
    const s = parseSymbol(code);
    const body = s.sky === "clearsky" || s.sky === "fair" || s.sky === "partlycloudy" ? (s.night ? "moon" : "sun") : null;
    const drops = s.precip ? { light: 2, normal: 3, heavy: 4 }[s.intensity] : 0;
    return {
        sky: s.sky,
        body,
        // Alone in the sky (clear) the sun or moon is drawn large and centred.
        bodyFull: s.sky === "clearsky",
        clouds: s.sky === "clearsky" || s.sky === "fog" ? 0 : s.sky === "cloudy" && !s.precip ? 2 : 1,
        dark: Boolean(s.precip) && (!s.showers || s.intensity === "heavy" || s.thunder),
        precip: s.precip,
        drops,
        showers: s.showers,
        thunder: s.thunder,
        fog: s.sky === "fog",
        stars: s.sky === "clearsky" && s.night,
    };
}

/**
 * The emoji for the "emoji" icon style.
 * @param {string | null | undefined} code
 */
export function symbolEmoji(code) {
    const s = parseSymbol(code);
    if (s.thunder) return "⛈️";
    if (s.precip === "snow") return "🌨️";
    if (s.precip === "sleet") return "🌨️";
    if (s.precip === "rain") return s.showers && !s.night ? "🌦️" : "🌧️";
    switch (s.sky) {
        case "clearsky":
            return s.night ? "🌙" : "☀️";
        case "fair":
            return s.night ? "🌙" : "🌤️";
        case "partlycloudy":
            return s.night ? "☁️" : "⛅";
        case "fog":
            return "🌫️";
        default:
            return "☁️";
    }
}

/**
 * The lit part of the moon as an SVG path, seen from the northern
 * hemisphere (waxing lights the right side).
 * @param {number} phase MET's moon phase in degrees: 0 new, 90 first quarter, 180 full, 270 last quarter
 * @param {number} cx
 * @param {number} cy
 * @param {number} r
 * @returns {string} "" for a new moon
 */
export function moonLitPath(phase, cx, cy, r) {
    const p = ((Number(phase) % 360) + 360) % 360;
    const lit = (1 - Math.cos((p * Math.PI) / 180)) / 2;
    if (lit < 0.01) return "";
    const top = `${cx} ${cy - r}`;
    const bottom = `${cx} ${cy + r}`;
    if (lit > 0.99) {
        return `M ${top} A ${r} ${r} 0 1 1 ${bottom} A ${r} ${r} 0 1 1 ${top} Z`;
    }
    const rx = Math.abs(Math.cos((p * Math.PI) / 180)) * r;
    const waxing = p < 180;
    const gibbous = lit > 0.5;
    // Outer edge, top to bottom: through the right side when waxing (sweep 1
    // is clockwise on screen), through the left when waning.
    const outerSweep = waxing ? 1 : 0;
    // Terminator, bottom back to top: it bulges towards the lit side for a
    // crescent and away from it for a gibbous moon.
    const innerSweep = waxing === gibbous ? 1 : 0;
    const f = (/** @type {number} */ n) => Math.round(n * 1000) / 1000;
    return `M ${top} A ${r} ${r} 0 0 ${outerSweep} ${bottom} A ${f(rx)} ${r} 0 0 ${innerSweep} ${top} Z`;
}

/**
 * The Hungarian name of a moon phase.
 * @param {number} phase degrees
 */
export function moonPhaseName(phase) {
    const p = ((Number(phase) % 360) + 360) % 360;
    if (p < 11.25 || p >= 348.75) return "újhold";
    if (p < 78.75) return "növekvő holdsarló";
    if (p < 101.25) return "első negyed";
    if (p < 168.75) return "növekvő hold";
    if (p < 191.25) return "telihold";
    if (p < 258.75) return "fogyó hold";
    if (p < 281.25) return "utolsó negyed";
    return "fogyó holdsarló";
}

const COMPASS = ["É", "ÉK", "K", "DK", "D", "DNy", "Ny", "ÉNy"];

/**
 * Where the wind blows from, as a Hungarian compass point ("ÉNy").
 * @param {number | null | undefined} deg
 */
export function compassHU(deg) {
    if (deg == null || Number.isNaN(Number(deg))) return "";
    const i = Math.round((((Number(deg) % 360) + 360) % 360) / 45) % 8;
    return COMPASS[i];
}

/**
 * The UV index's level in Hungarian (WHO bands).
 * @param {number | null | undefined} uv
 * @returns {{ label: string, level: 0 | 1 | 2 | 3 | 4 }}
 */
export function uvLevel(uv) {
    const v = Number(uv ?? 0);
    if (v < 3) return { label: "alacsony", level: 0 };
    if (v < 6) return { label: "mérsékelt", level: 1 };
    if (v < 8) return { label: "magas", level: 2 };
    if (v < 11) return { label: "nagyon magas", level: 3 };
    return { label: "extrém", level: 4 };
}

/**
 * Attribution the source's licence asks for, shown next to the data.
 * @param {string} source "metno" | "weatherapi_com" | "openweathermap"
 * @returns {{ text: string, href: string, linkText: string }}
 */
export function sourceCredit(source) {
    switch (source) {
        case "weatherapi_com":
            return { text: "Powered by", href: "https://www.weatherapi.com/", linkText: "WeatherAPI.com" };
        case "openweathermap":
            return { text: "Weather data ©", href: "https://openweathermap.org/", linkText: "OpenWeather" };
        default:
            return {
                text: "Időjárás-adatok: MET Norway,",
                href: "https://api.met.no/doc/License",
                linkText: "CC BY 4.0",
            };
    }
}

/**
 * The credit link for a source as the API names it ("MET Norway",
 * "WeatherAPI.com", "OpenWeather").
 * @param {string} name
 */
export function sourceHrefByName(name) {
    switch (name) {
        case "WeatherAPI.com":
            return sourceCredit("weatherapi_com").href;
        case "OpenWeather":
            return sourceCredit("openweathermap").href;
        case "MET Norway":
            return sourceCredit("metno").href;
        default:
            return "";
    }
}
