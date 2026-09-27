import { canonicalEntryCategory } from "./entryCategory.js";
import { formatDateShort } from "./utils.js";

function text(value) {
    return String(value ?? "").trim();
}

function joinMeta(parts) {
    return parts.map(text).filter(Boolean).join(" · ");
}

function eventDate(start) {
    const raw = text(start);
    if (!raw) return "";
    const parsed = new Date(raw);
    if (Number.isNaN(parsed.getTime())) return "";
    return formatDateShort(parsed);
}

function card(fields) {
    const title = text(fields.title);
    const href = text(fields.href);
    if (!title || !href) return null;
    return {
        href,
        title,
        meta: text(fields.meta),
        description: text(fields.description),
        accent: fields.accent,
        external: fields.external,
        claimed: fields.claimed,
    };
}

/**
 * @param {"website"|"service"|"attraction"|"venue"|"event"|"settlement"|"seat"|"news"} kind
 * @param {Record<string, unknown>} row
 */
export function searchResultCardModel(kind, row) {
    const item = row && typeof row === "object" ? row : {};
    if (kind === "website") {
        return card({
            href: item.url,
            title: item.title,
            meta: item.domain,
            description: item.description,
            accent: "none",
            external: true,
            claimed: Boolean(item.claimed),
        });
    }
    if (kind === "service") {
        const slug = text(item.slug);
        return card({
            href: slug ? `/bejegyzes/${slug}` : "",
            title: item.name,
            meta: joinMeta([item.location, canonicalEntryCategory(item.category)]),
            description: item.notes,
            accent: "none",
            external: false,
            claimed: Boolean(item.claimed),
        });
    }
    if (kind === "attraction") {
        const slug = text(item.slug);
        const county = text(item.county_slug);
        return card({
            href: slug && county ? `/${county}-megye/${slug}` : "",
            title: item.name,
            meta: item.county_name,
            description: item.description,
            accent: "attraction",
            external: false,
            claimed: null,
        });
    }
    if (kind === "venue") {
        const slug = text(item.slug);
        const county = text(item.county_slug);
        const settlement = text(item.settlement_slug);
        return card({
            href: slug && county && settlement
                ? `/${county}-megye/${settlement}/helyszin/${slug}`
                : "",
            title: item.name,
            meta: joinMeta([item.settlement_name, item.kind_label]),
            description: "",
            accent: "none",
            external: false,
            claimed: null,
        });
    }
    if (kind === "event") {
        const id = text(item.id);
        return card({
            href: id ? `/esemenyek/${id}` : "",
            title: item.title,
            meta: joinMeta([eventDate(item.start_date), item.location_name]),
            description: "",
            accent: "none",
            external: false,
            claimed: null,
        });
    }
    if (kind === "settlement") {
        const slug = text(item.slug);
        const county = text(item.county_slug);
        return card({
            href: slug && county ? `/${county}-megye/${slug}` : "",
            title: item.name,
            meta: item.county,
            description: "",
            accent: "none",
            external: false,
            claimed: null,
        });
    }
    if (kind === "seat") {
        const slug = text(item.slug);
        return card({
            href: slug ? `/szekek/${slug}` : "",
            title: item.name,
            meta: "",
            description: "",
            accent: "seat",
            external: false,
            claimed: null,
        });
    }
    if (kind === "news") {
        return card({
            href: item.link,
            title: item.title,
            meta: item.source,
            description: "",
            accent: "none",
            external: true,
            claimed: null,
        });
    }
    return null;
}
