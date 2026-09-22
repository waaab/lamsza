/** @param {unknown} raw */
export function eventEntryPriceText(raw) {
    const s = raw == null ? "" : String(raw).trim();
    return s === "" ? "-" : s;
}

/** @param {unknown} raw */
export function eventEntryPriceIsEmpty(raw) {
    const s = raw == null ? "" : String(raw).trim();
    return s === "";
}
