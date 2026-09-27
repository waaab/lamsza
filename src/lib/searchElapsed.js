/**
 * Hungarian seconds label for a finished search, e.g. "0,4 mp".
 * A search that finished is never shown as zero.
 * @param {number} ms
 */
export function formatSearchElapsed(ms) {
    const seconds = Math.max(0, Number(ms) || 0) / 1000;
    const tenths = Math.max(1, Math.round(seconds * 10));
    const whole = Math.floor(tenths / 10);
    const frac = tenths % 10;
    return `${whole},${frac} mp`;
}
