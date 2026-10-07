/**
 * schema.org JSON-LD builders.
 *
 * Every builder is pure: it takes plain data and returns a plain object (or
 * `null` when there is not enough real data to describe). Rendering goes
 * through `$lib/components/JsonLd.svelte`.
 *
 * URLs reuse `$lib/seo.js`, so JSON-LD, canonical and og:url always name the
 * same canonical production host - also from dev and from the browser
 * extension newtab page, where `location.origin` is not a web origin.
 */

import { WEEKDAYS, normalizeHours } from "./entryHours.js";
import { canonicalEntryType, ENTRY_TYPE_SZEMELY } from "./entryType.js";
import { gallerySlides } from "./entryPhotos.js";
import {
    showListingHours,
    showListingPhone,
    showListingPhotos,
    showListingRatings,
    showListingSocial,
} from "./entryPublicExtras.js";
import {
    DEFAULT_SEO_DESCRIPTION,
    SITE_NAME,
    SITE_ORIGIN,
    absoluteUrl,
} from "./seo.js";

export { SITE_NAME, SITE_ORIGIN, absoluteUrl };

export const SITE_ALTERNATE_NAME = "Lámsza.com";
export const SITE_LOGO_URL = absoluteUrl("/icon.png");
/** The homepage search box reads `?q=`, so this template really works. */
export const SEARCH_URL_TEMPLATE = `${SITE_ORIGIN}/?q={search_term_string}`;

const WEEKDAY_SCHEMA_NAME = {
    mon: "Monday",
    tue: "Tuesday",
    wed: "Wednesday",
    thu: "Thursday",
    fri: "Friday",
    sat: "Saturday",
    sun: "Sunday",
};

/** @param {unknown} raw */
function text(raw) {
    return raw == null ? "" : String(raw).trim();
}

/**
 * Absolute URL for a site path, or "" for empty input. Unlike
 * {@link absoluteUrl} an empty value never becomes the bare site origin, so
 * optional fields stay out of the markup instead of pointing at the homepage.
 * @param {unknown} raw
 */
export function nodeUrl(raw) {
    const s = text(raw);
    return s ? absoluteUrl(s) : "";
}

/**
 * Serialize one node or a list of nodes into a `<script type="application/ld+json">`
 * tag. `<` is escaped so the payload can never close the tag early.
 * @param {unknown} data
 */
export function jsonLdScript(data) {
    const nodes = (Array.isArray(data) ? data : [data]).filter(Boolean);
    if (!nodes.length) return "";
    const payload = JSON.stringify(nodes.length === 1 ? nodes[0] : nodes).replace(
        /</g,
        "\\u003c",
    );
    return `<script type="application/ld+json">${payload}</` + `script>`;
}

/** Site-wide `WebSite` node with the search box action. */
export function webSiteJsonLd() {
    return {
        "@context": "https://schema.org",
        "@type": "WebSite",
        "@id": `${SITE_ORIGIN}/#website`,
        url: `${SITE_ORIGIN}/`,
        name: SITE_NAME,
        alternateName: SITE_ALTERNATE_NAME,
        description: DEFAULT_SEO_DESCRIPTION,
        inLanguage: "hu",
        publisher: { "@id": `${SITE_ORIGIN}/#organization` },
        potentialAction: {
            "@type": "SearchAction",
            target: {
                "@type": "EntryPoint",
                urlTemplate: SEARCH_URL_TEMPLATE,
            },
            "query-input": "required name=search_term_string",
        },
    };
}

/**
 * Site-wide `Organization` node. Social profiles come from the public config,
 * so an empty list simply drops `sameAs`.
 * @param {{ socialLinks?: Array<{ label?: string, url?: string }> }} [opts]
 */
export function organizationJsonLd(opts = {}) {
    const sameAs = (Array.isArray(opts.socialLinks) ? opts.socialLinks : [])
        .map((row) => text(row?.url))
        .filter((url) => /^https?:\/\//i.test(url));
    /** @type {Record<string, unknown>} */
    const data = {
        "@context": "https://schema.org",
        "@type": "Organization",
        "@id": `${SITE_ORIGIN}/#organization`,
        name: SITE_NAME,
        alternateName: SITE_ALTERNATE_NAME,
        url: `${SITE_ORIGIN}/`,
        description: DEFAULT_SEO_DESCRIPTION,
        logo: {
            "@type": "ImageObject",
            url: SITE_LOGO_URL,
        },
    };
    if (sameAs.length) data.sameAs = [...new Set(sameAs)];
    return data;
}

/**
 * Breadcrumb trail in the same order the `Breadcrumbs` markup renders it.
 * The last item is the current page and carries no URL.
 *
 * @param {{
 *   label?: string,
 *   parentLabel?: string,
 *   parentUrl?: string,
 *   extraLabel?: string,
 *   extraUrl?: string,
 *   countySlug?: string,
 *   countyName?: string,
 *   settlementSlug?: string,
 *   settlementName?: string,
 *   settlementType?: string,
 * }} props
 * @returns {Array<{ name: string, url?: string }>}
 */
export function breadcrumbItems(props = {}) {
    const label = text(props.label);
    const parentLabel = text(props.parentLabel);
    const parentUrl = text(props.parentUrl);
    const extraLabel = text(props.extraLabel);
    const extraUrl = text(props.extraUrl);
    const countySlug = text(props.countySlug);
    const countyName = text(props.countyName);
    const settlementSlug = text(props.settlementSlug);
    const settlementName = text(props.settlementName);
    const settlementType = text(props.settlementType);

    /** @type {Array<{ name: string, url?: string }>} */
    const items = [{ name: "Főoldal", url: `${SITE_ORIGIN}/` }];
    if (parentLabel && parentUrl) {
        items.push({ name: parentLabel, url: nodeUrl(parentUrl) });
    }
    if (extraLabel && extraUrl) {
        items.push({ name: extraLabel, url: nodeUrl(extraUrl) });
    }
    if (countyName && countySlug) {
        items.push({ name: countyName, url: `${SITE_ORIGIN}/${countySlug}-megye` });
    }
    if (settlementSlug && settlementName) {
        items.push({
            name: settlementName,
            url: `${SITE_ORIGIN}/${countySlug}-megye/${settlementSlug}`,
        });
        if (label) items.push({ name: label });
        return items;
    }
    if (!label) return items;
    items.push({ name: settlementType ? `${settlementType}: ${label}` : label });
    return items;
}

/**
 * `BreadcrumbList` for a trail. Returns `null` for a trail that is only
 * "Főoldal" (nothing to show) so pages still loading emit nothing.
 * @param {Array<{ name: string, url?: string }>} items
 */
export function breadcrumbListJsonLd(items) {
    const list = (Array.isArray(items) ? items : []).filter((row) => text(row?.name));
    if (list.length < 2) return null;
    return {
        "@context": "https://schema.org",
        "@type": "BreadcrumbList",
        itemListElement: list.map((row, i) => {
            /** @type {Record<string, unknown>} */
            const element = {
                "@type": "ListItem",
                position: i + 1,
                name: text(row.name),
            };
            const url = nodeUrl(row.url);
            if (url) element.item = url;
            return element;
        }),
    };
}

/** @param {Record<string, any> | null | undefined} entry */
function openingHours(entry) {
    if (!showListingHours(entry)) return [];
    const hours = normalizeHours(entry?.hours);
    /** @type {Array<Record<string, string>>} */
    const out = [];
    for (const { key } of WEEKDAYS) {
        const slot = hours[key];
        // Closed days are left out; a missing day already means "not open".
        if (!slot || slot.closed || !slot.open || !slot.close) continue;
        out.push({
            "@type": "OpeningHoursSpecification",
            dayOfWeek: `https://schema.org/${WEEKDAY_SCHEMA_NAME[key]}`,
            opens: slot.open,
            closes: slot.close,
        });
    }
    return out;
}

/** @param {Record<string, any> | null | undefined} entry */
function entrySchemaType(entry) {
    return canonicalEntryType(entry?.type) === ENTRY_TYPE_SZEMELY
        ? "Person"
        : "LocalBusiness";
}

/**
 * Listing node for an index entry. Only data the page itself shows is
 * described, so the markup never claims more than the visible page.
 *
 * @param {Record<string, any> | null | undefined} entry
 */
export function listingJsonLd(entry) {
    const name = text(entry?.name);
    if (!name) return null;

    const schemaType = entrySchemaType(entry);
    const isPerson = schemaType === "Person";
    const slug = text(entry?.slug);
    const pageUrl = slug ? `${SITE_ORIGIN}/bejegyzes/${slug}` : "";
    const website = text(entry?.url);

    /** @type {Record<string, unknown>} */
    const data = {
        "@context": "https://schema.org",
        "@type": schemaType,
        name,
    };
    if (pageUrl) {
        data["@id"] = `${pageUrl}#listing`;
        data.mainEntityOfPage = pageUrl;
    }
    const canonical = website || pageUrl;
    if (canonical) data.url = canonical;
    if (showListingPhone(entry) && text(entry?.phone)) {
        data.telephone = text(entry.phone);
    }

    /** @type {Record<string, string>} */
    const postal = { "@type": "PostalAddress" };
    const address = text(entry?.address);
    const locality = text(entry?.location);
    const region = text(entry?.location_county);
    if (address) postal.streetAddress = address;
    if (locality) postal.addressLocality = locality;
    if (region) postal.addressRegion = region;
    if (Object.keys(postal).length > 1) {
        postal.addressCountry = "RO";
        data.address = postal;
    }

    const coords = parseCoordinates(entry?.location_coordinates);
    if (coords && !isPerson) {
        data.geo = {
            "@type": "GeoCoordinates",
            latitude: coords.lat,
            longitude: coords.lng,
        };
    }

    const categories = listingCategories(entry);
    if (categories.length) data.knowsAbout = categories;
    if (locality && !isPerson) data.areaServed = locality;

    const hours = isPerson ? [] : openingHours(entry);
    if (hours.length) data.openingHoursSpecification = hours;

    if (showListingPhotos(entry)) {
        const images = gallerySlides(entry)
            .map((slide) => nodeUrl(slide?.src))
            .filter(Boolean);
        if (images.length) data.image = images;
    }

    if (showListingSocial(entry)) {
        const sameAs = (Array.isArray(entry?.social_links) ? entry.social_links : [])
            .map((row) => text(row?.url))
            .filter((url) => /^https?:\/\//i.test(url));
        if (sameAs.length) data.sameAs = [...new Set(sameAs)];
    }

    const rating = aggregateRating(entry);
    if (rating) data.aggregateRating = rating;

    return data;
}

/** @param {Record<string, any> | null | undefined} entry */
function listingCategories(entry) {
    const many = Array.isArray(entry?.categories)
        ? entry.categories.map((c) => text(c)).filter(Boolean)
        : [];
    if (many.length) return [...new Set(many)];
    const one = text(entry?.category);
    return one ? [one] : [];
}

/** @param {Record<string, any> | null | undefined} entry */
function aggregateRating(entry) {
    if (!showListingRatings(entry)) return null;
    const value = Number(entry?.rating);
    const count = Number(entry?.review_count);
    if (!Number.isFinite(value) || value <= 0) return null;
    if (!Number.isFinite(count) || count < 1) return null;
    return {
        "@type": "AggregateRating",
        ratingValue: Math.min(5, Math.max(0, value)),
        reviewCount: Math.floor(count),
        bestRating: 5,
        worstRating: 1,
    };
}

/** @param {unknown} raw "lat,lng" */
export function parseCoordinates(raw) {
    const parts = text(raw)
        .split(",")
        .map((part) => part.trim());
    if (parts.length !== 2) return null;
    const lat = Number(parts[0]);
    const lng = Number(parts[1]);
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null;
    return { lat, lng };
}

/**
 * ISO-8601 local date-time. Times are local to the event, and the source data
 * carries no zone, so no offset is written.
 * @param {unknown} date YYYY-MM-DD (or an ISO string)
 * @param {unknown} time HH:MM(:SS)
 */
export function eventDateTime(date, time) {
    const d = text(date);
    if (!d) return "";
    const day = d.length >= 10 ? d.slice(0, 10) : d;
    if (!/^\d{4}-\d{2}-\d{2}$/.test(day)) return "";
    const t = text(time);
    if (!t) return day;
    const hhmm = t.slice(0, 5);
    if (!/^\d{2}:\d{2}$/.test(hhmm)) return day;
    return `${day}T${hhmm}`;
}

/**
 * `Event` node for an event detail page.
 *
 * @param {Record<string, any> | null | undefined} event
 * @param {{ imageUrl?: string, venues?: Array<{ name?: string }> }} [opts]
 */
export function eventJsonLd(event, opts = {}) {
    const name = text(event?.title);
    const startDate = eventDateTime(event?.start_date, event?.start_time);
    if (!name || !startDate) return null;

    const id = text(event?.id);
    const locality = text(event?.location_name);
    const county = text(event?.county);
    const venueName = text((Array.isArray(opts.venues) ? opts.venues : [])[0]?.name);

    /** @type {Record<string, unknown>} */
    const place = {
        "@type": "Place",
        name: venueName || locality || SITE_NAME,
    };
    /** @type {Record<string, string>} */
    const postal = { "@type": "PostalAddress" };
    if (locality) postal.addressLocality = locality;
    if (county) postal.addressRegion = `${county} megye`;
    if (Object.keys(postal).length > 1) {
        postal.addressCountry = "RO";
        place.address = postal;
    }

    /** @type {Record<string, unknown>} */
    const data = {
        "@context": "https://schema.org",
        "@type": "Event",
        name,
        startDate,
        eventStatus: "https://schema.org/EventScheduled",
        eventAttendanceMode: "https://schema.org/OfflineEventAttendanceMode",
        location: place,
    };

    // A missing end time on a single-day event means "no end known": writing a
    // date-only endDate there would claim the event ends before it starts.
    const startDay = text(event?.start_date).slice(0, 10);
    const endDay = text(event?.end_date).slice(0, 10) || startDay;
    const endTime = text(event?.end_time);
    const endDate =
        endDay !== startDay || endTime ? eventDateTime(endDay, endTime) : "";
    if (endDate && endDate !== startDate) data.endDate = endDate;
    if (id) {
        data["@id"] = `${SITE_ORIGIN}/esemenyek/${id}#event`;
        data.url = `${SITE_ORIGIN}/esemenyek/${id}`;
    }
    const description = text(event?.description);
    if (description) data.description = description;
    const image = nodeUrl(opts.imageUrl);
    if (image) data.image = [image];
    const organizer = text(event?.organizer);
    if (organizer) {
        data.organizer = { "@type": "Organization", name: organizer };
    }
    // `entry_price` is free text ("500 lej / felnőtt"), so no `offers` node:
    // Google needs a parsed price and currency, and guessing would be wrong.
    return data;
}
