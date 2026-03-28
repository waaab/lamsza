/**
 * Maps the current pathname to a `page_faq_sections.section_key` row.
 * Keep in sync with backend seeds in `backend/internal/pagefaq/pagefaq.go`.
 */
export function deriveFaqSectionKey(pathname) {
    const p = pathname || "/";
    if (p === "/" || p === "") return "home";
    const parts = p.split("/").filter(Boolean);

    if (parts[0] === "index") return "index";
    if (parts[0] === "hirek") return "hirek";
    if (parts[0] === "megyek") return "megyek";
    if (
        parts.length === 1 &&
        parts[0].endsWith("-megye") &&
        parts[0] !== "szekek"
    ) {
        return "megye";
    }
    if (
        parts.length >= 2 &&
        parts[0].endsWith("-megye") &&
        parts[0] !== "szekek"
    ) {
        return "telepules";
    }
    if (parts[0] === "szekek") {
        return parts.length >= 2 ? "szek" : "szekek";
    }
    if (parts[0] === "varosok" || parts[0] === "varos") return "varosok";
    if (parts[0] === "falvak" || parts[0] === "falu") return "falvak";
    if (parts[0] === "esemenyek") {
        return parts.length >= 2 ? "esemeny" : "esemenyek";
    }
    if (parts[0] === "terkep") return "terkep";
    if (parts[0] === "valtozasnaplo") return "valtozasnaplo";
    if (parts[0] === "bejegyzes") return "bejegyzes";
    if (parts[0] === "iranyelvek") return "iranyelvek";
    return "home";
}
