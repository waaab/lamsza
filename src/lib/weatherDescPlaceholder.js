/** Same visible length as a short description such as "Béborult esment...". */
export const WEATHER_DESC_PLACEHOLDER_LENGTH = 18;

const WEATHER_DESC_LOADING = "Időjárás adatok betöltése";

/**
 * Weather status placeholder. Longer copy is cut to 18 characters and ends with "...".
 * @param {string} [text]
 * @param {number} [max]
 */
export function fitWeatherDescPlaceholder(text = WEATHER_DESC_LOADING, max = WEATHER_DESC_PLACEHOLDER_LENGTH) {
    const chars = Array.from(text);
    if (chars.length <= max) return text;
    return chars.slice(0, max - 3).join("") + "...";
}

export const weatherDescPlaceholder = fitWeatherDescPlaceholder();

/** Source name after "Forrás:". At least as long as a 14-character label. */
export const WEATHER_SOURCE_PLACEHOLDER_LENGTH = 14;

export const weatherSourcePlaceholder = "adat betöltés...";
