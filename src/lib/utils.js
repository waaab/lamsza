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
 * Maps OpenWeatherMap codes to Emojis
 * @param {string} code - OWM icon code
 * @returns {string}
 */
export function weatherIconEmoji(code) {
    const map = {
        "01d": "☀️",
        "01n": "🌙",
        "02d": "⛅",
        "02n": "☁️",
        "03d": "☁️",
        "03n": "☁️",
        "04d": "☁️",
        "04n": "☁️",
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
    return map[code] || "🌡️";
}

/**
 * Format a date to local string (Hungarian)
 * @param {number|string|Date} date 
 * @returns {string}
 */
export function formatDate(date) {
    return new Date(date).toLocaleDateString("hu-HU");
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
