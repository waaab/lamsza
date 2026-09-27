export function diffSuggestionFields(before, after) {
    const out = {};
    const next = after && typeof after === "object" ? after : {};
    const prev = before && typeof before === "object" ? before : {};
    for (const key of Object.keys(next)) {
        if (JSON.stringify(prev[key]) !== JSON.stringify(next[key])) {
            out[key] = next[key];
        }
    }
    return out;
}
