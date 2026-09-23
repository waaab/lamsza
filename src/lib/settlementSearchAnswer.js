/**
 * Spoken-style settlement summary for search, built only from stored facts.
 */

export function foldPlace(value) {
    return String(value ?? "")
        .trim()
        .toLowerCase()
        .normalize("NFD")
        .replace(/\p{M}/gu, "")
        .replace(/[^a-z0-9]+/g, "");
}

/**
 * The settlement the query is actually about.
 * A partial that matches more than one town returns null.
 * @param {Array<Record<string, unknown>> | null | undefined} locations
 * @param {string} query
 */
export function pickSettlement(locations, query) {
    const q = foldPlace(query);
    if (!q) return null;
    const towns = (locations || []).filter((loc) => String(loc?.type || "") !== "megye");
    const exact = towns.filter((loc) => {
        const keys = [loc.slug, loc.name, loc.name_ro, loc.name_de];
        return keys.some((key) => foldPlace(key) === q);
    });
    if (exact.length === 1) return exact[0];
    if (exact.length > 1) return exact[0];
    return null;
}

function clean(value) {
    const text = String(value ?? "").trim();
    if (!text || text === "-") return "";
    return text;
}

function formatPopulation(value) {
    const text = clean(value);
    if (!text) return "";
    if (/fő/i.test(text)) return text;
    return `${text} fő`;
}

function formatArea(value) {
    const text = clean(value);
    if (!text) return "";
    const pretty = text.replace(".", ",");
    if (/km/i.test(pretty)) return pretty;
    return `${pretty} km²`;
}

function weatherSentence(weather) {
    if (!weather || weather.temp == null || weather.temp === "") return "";
    const desc = clean(weather.desc);
    const emoji = clean(weather.emoji);
    const temp = `${weather.temp} °C`;
    const status = desc ? `${desc}, ${temp}` : temp;
    return `Az időjárás éppen: ${status}${emoji ? ` ${emoji}` : ""}.`;
}

function eventsSentence(events) {
    const titles = (events || [])
        .map((event) => clean(event?.title))
        .filter(Boolean);
    if (titles.length === 0) return "";
    if (titles.length === 1) return `Most zajló esemény: ${titles[0]}.`;
    return `Most zajló események: ${titles.join(", ")}.`;
}

/**
 * @param {Record<string, unknown> | null | undefined} settlement
 * @param {{ weather?: { temp?: number|string, desc?: string, emoji?: string } | null, events?: Array<{ title?: string }> }} [extras]
 */
export function buildSettlementAnswer(settlement, extras = {}) {
    const name = clean(settlement?.name);
    if (!name) return "";

    const aliases = [];
    const ro = clean(settlement.name_ro);
    const de = clean(settlement.name_de);
    if (ro) aliases.push(`románul ${ro}`);
    if (de) aliases.push(`németül ${de}`);

    const sentences = [aliases.length ? `${name}, ${aliases.join(", ")}` : name];
    const post = clean(settlement.post_code);
    if (post) sentences.push(`Irányítószám: ${post}`);
    const population = formatPopulation(settlement.population);
    if (population) sentences.push(`Lakosság: ${population}`);
    const area = formatArea(settlement.area);
    if (area) sentences.push(`Terület: ${area}`);
    const type = clean(settlement.type);
    if (type) sentences.push(`Közigazgatási forma: ${type}`);

    const parts = [`${sentences.join(". ")}.`, weatherSentence(extras.weather), eventsSentence(extras.events)];
    return parts.filter(Boolean).join(" ");
}

export function settlementFoundLine(name) {
    const label = clean(name);
    return label ? `Erre találtam ${label} kapcsán:` : "";
}
