import assert from "node:assert/strict";
import { test } from "node:test";
import {
    SEARCH_URL_TEMPLATE,
    SITE_ORIGIN,
    breadcrumbItems,
    breadcrumbListJsonLd,
    eventDateTime,
    eventJsonLd,
    jsonLdScript,
    listingJsonLd,
    nodeUrl,
    organizationJsonLd,
    webSiteJsonLd,
} from "../src/lib/structuredData.js";

test("jsonLdScript escapes < so the payload cannot close the tag", () => {
    const out = jsonLdScript({ name: "</script><img src=x>" });
    assert.ok(out.startsWith('<script type="application/ld+json">'));
    assert.ok(!out.slice(0, -9).includes("</script>"));
    assert.ok(out.includes("\\u003c/script"));
});

test("jsonLdScript drops empty input and keeps lists", () => {
    assert.equal(jsonLdScript(null), "");
    assert.equal(jsonLdScript([]), "");
    assert.equal(jsonLdScript([null, false]), "");
    const many = jsonLdScript([{ a: 1 }, { b: 2 }]);
    assert.ok(many.includes('[{"a":1},{"b":2}]'));
    const one = jsonLdScript([{ a: 1 }]);
    assert.ok(one.includes('{"a":1}'));
});

test("nodeUrl keeps empty values empty instead of pointing at the homepage", () => {
    assert.equal(nodeUrl(""), "");
    assert.equal(nodeUrl(null), "");
    assert.equal(nodeUrl("/index"), `${SITE_ORIGIN}/index`);
    assert.equal(nodeUrl("https://example.com/x"), "https://example.com/x");
});

test("WebSite carries a SearchAction the homepage can actually answer", () => {
    const site = webSiteJsonLd();
    assert.equal(site["@type"], "WebSite");
    assert.equal(site.potentialAction["@type"], "SearchAction");
    assert.equal(site.potentialAction.target.urlTemplate, SEARCH_URL_TEMPLATE);
    assert.equal(
        site.potentialAction["query-input"],
        "required name=search_term_string",
    );
    assert.ok(SEARCH_URL_TEMPLATE.includes("{search_term_string}"));
    assert.ok(SEARCH_URL_TEMPLATE.startsWith(`${SITE_ORIGIN}/?q=`));
});

test("Organization only lists http(s) social profiles, deduped", () => {
    const org = organizationJsonLd({
        socialLinks: [
            { label: "Facebook", url: "https://facebook.com/lamsza" },
            { label: "Facebook", url: "https://facebook.com/lamsza" },
            { label: "Rossz", url: "javascript:alert(1)" },
            { label: "Üres", url: "  " },
            null,
        ],
    });
    assert.deepEqual(org.sameAs, ["https://facebook.com/lamsza"]);
    assert.equal(org.logo["@type"], "ImageObject");
});

test("Organization without social links has no sameAs", () => {
    assert.equal("sameAs" in organizationJsonLd(), false);
    assert.equal("sameAs" in organizationJsonLd({ socialLinks: [] }), false);
});

test("breadcrumbItems mirrors the rendered trail", () => {
    const items = breadcrumbItems({
        label: "Hentes Kft.",
        parentLabel: "Index",
        parentUrl: "/index",
        extraLabel: "Szolgáltatások",
        extraUrl: "/index/szolgaltatasok",
    });
    assert.deepEqual(items, [
        { name: "Főoldal", url: `${SITE_ORIGIN}/` },
        { name: "Index", url: `${SITE_ORIGIN}/index` },
        { name: "Szolgáltatások", url: `${SITE_ORIGIN}/index/szolgaltatasok` },
        { name: "Hentes Kft." },
    ]);
});

test("breadcrumbItems covers county and settlement crumbs", () => {
    const items = breadcrumbItems({
        label: "Vár",
        parentLabel: "",
        parentUrl: "",
        countySlug: "hargita",
        countyName: "Hargita",
        settlementSlug: "csikszereda",
        settlementName: "Csíkszereda",
    });
    assert.deepEqual(items.map((row) => row.name), [
        "Főoldal",
        "Hargita",
        "Csíkszereda",
        "Vár",
    ]);
    assert.equal(items[2].url, `${SITE_ORIGIN}/hargita-megye/csikszereda`);
});

test("breadcrumbItems keeps the settlement type prefix", () => {
    const items = breadcrumbItems({
        label: "Zetelaka",
        parentLabel: "Falvak",
        parentUrl: "/falvak",
        settlementType: "falu",
    });
    assert.equal(items.at(-1).name, "falu: Zetelaka");
});

test("BreadcrumbList is skipped while the page label is still empty", () => {
    const items = breadcrumbItems({ label: "", parentLabel: "", parentUrl: "" });
    assert.deepEqual(items, [{ name: "Főoldal", url: `${SITE_ORIGIN}/` }]);
    assert.equal(breadcrumbListJsonLd(items), null);
    assert.equal(breadcrumbListJsonLd([]), null);
});

test("BreadcrumbList numbers positions and leaves the last item without item", () => {
    const list = breadcrumbListJsonLd(
        breadcrumbItems({
            label: "Hentes Kft.",
            parentLabel: "Index",
            parentUrl: "/index",
        }),
    );
    assert.equal(list["@type"], "BreadcrumbList");
    assert.deepEqual(
        list.itemListElement.map((row) => row.position),
        [1, 2, 3],
    );
    assert.equal(list.itemListElement[0].item, `${SITE_ORIGIN}/`);
    assert.equal("item" in list.itemListElement[2], false);
});

const BUSINESS = {
    name: "Hentes Kft.",
    slug: "hentes-kft",
    type: "Vállalkozás",
    url: "https://hentes.ro",
    phone: "+40 266 111 222",
    address: "Fő utca 1.",
    location: "Csíkszereda",
    location_county: "Hargita",
    location_coordinates: "46.3597, 25.8017",
    categories: ["Hentes", "Élelmiszer"],
    claimed: true,
    verified: true,
    hours_enabled: true,
    hours: {
        mon: { open: "08:00", close: "16:00", closed: false },
        tue: { open: "", close: "", closed: true },
        wed: { open: "08:00", close: "", closed: false },
    },
    ratings_enabled: true,
    rating: 4.5,
    review_count: 12,
    social_links: [{ label: "Facebook", url: "https://facebook.com/hentes" }],
    photos: [{ url: "/api/media/1.jpg", alt: "Bolt" }],
};

test("listing node describes a claimed business", () => {
    const data = listingJsonLd(BUSINESS);
    assert.equal(data["@type"], "LocalBusiness");
    assert.equal(data["@id"], `${SITE_ORIGIN}/bejegyzes/hentes-kft#listing`);
    assert.equal(data.mainEntityOfPage, `${SITE_ORIGIN}/bejegyzes/hentes-kft`);
    assert.equal(data.url, "https://hentes.ro");
    assert.equal(data.telephone, "+40 266 111 222");
    assert.deepEqual(data.address, {
        "@type": "PostalAddress",
        streetAddress: "Fő utca 1.",
        addressLocality: "Csíkszereda",
        addressRegion: "Hargita",
        addressCountry: "RO",
    });
    assert.deepEqual(data.geo, {
        "@type": "GeoCoordinates",
        latitude: 46.3597,
        longitude: 25.8017,
    });
    assert.deepEqual(data.knowsAbout, ["Hentes", "Élelmiszer"]);
    assert.equal(data.areaServed, "Csíkszereda");
    assert.deepEqual(data.sameAs, ["https://facebook.com/hentes"]);
    assert.deepEqual(data.image, [`${SITE_ORIGIN}/api/media/1.jpg`]);
    assert.deepEqual(data.aggregateRating, {
        "@type": "AggregateRating",
        ratingValue: 4.5,
        reviewCount: 12,
        bestRating: 5,
        worstRating: 1,
    });
});

test("opening hours skip closed and half-filled days", () => {
    const hours = listingJsonLd(BUSINESS).openingHoursSpecification;
    assert.deepEqual(hours, [
        {
            "@type": "OpeningHoursSpecification",
            dayOfWeek: "https://schema.org/Monday",
            opens: "08:00",
            closes: "16:00",
        },
    ]);
});

test("listing node hides what the page itself hides", () => {
    const unclaimed = listingJsonLd({
        ...BUSINESS,
        claimed: false,
        verified: false,
    });
    assert.equal("telephone" in unclaimed, false);
    assert.equal("sameAs" in unclaimed, false);
    assert.equal("image" in unclaimed, false);
    assert.equal("aggregateRating" in unclaimed, false);
    // Hours stay: the page shows them for unclaimed listings too.
    assert.equal(unclaimed.openingHoursSpecification.length, 1);
});

test("no aggregateRating without real reviews", () => {
    const zero = listingJsonLd({ ...BUSINESS, rating: 0, review_count: 0 });
    assert.equal("aggregateRating" in zero, false);
    const noCount = listingJsonLd({ ...BUSINESS, review_count: 0 });
    assert.equal("aggregateRating" in noCount, false);
});

test("a person listing is a Person without business-only fields", () => {
    const data = listingJsonLd({ ...BUSINESS, type: "Személy" });
    assert.equal(data["@type"], "Person");
    assert.equal("geo" in data, false);
    assert.equal("areaServed" in data, false);
    assert.equal("openingHoursSpecification" in data, false);
    assert.equal(data.telephone, "+40 266 111 222");
});

test("a listing without a website falls back to its own page URL", () => {
    const data = listingJsonLd({ ...BUSINESS, url: "" });
    assert.equal(data.url, `${SITE_ORIGIN}/bejegyzes/hentes-kft`);
});

test("a nameless listing emits nothing", () => {
    assert.equal(listingJsonLd(null), null);
    assert.equal(listingJsonLd({ name: "  " }), null);
});

test("eventDateTime joins date and time, and rejects junk", () => {
    assert.equal(eventDateTime("2026-07-04", "19:30:00"), "2026-07-04T19:30");
    assert.equal(eventDateTime("2026-07-04T00:00:00Z", "19:30"), "2026-07-04T19:30");
    assert.equal(eventDateTime("2026-07-04", ""), "2026-07-04");
    assert.equal(eventDateTime("2026-07-04", "este"), "2026-07-04");
    assert.equal(eventDateTime("", "19:30"), "");
    assert.equal(eventDateTime("nincs", "19:30"), "");
});

const EVENT = {
    id: 42,
    title: "Ezer Székely Leány Napja",
    description: "Néptánc és népzene.",
    start_date: "2026-07-04",
    start_time: "10:00",
    end_date: "2026-07-05",
    end_time: "22:00",
    location_name: "Csíksomlyó",
    location_slug: "csiksomlyo",
    county: "Hargita",
    county_slug: "hargita",
    organizer: "Hargita Megye Tanácsa",
    entry_price: "ingyenes",
};

test("Event node has the fields Google needs", () => {
    const data = eventJsonLd(EVENT, {
        imageUrl: "/images/event.jpg",
        venues: [{ name: "Nyergestető" }],
    });
    assert.equal(data["@type"], "Event");
    assert.equal(data.name, EVENT.title);
    assert.equal(data.startDate, "2026-07-04T10:00");
    assert.equal(data.endDate, "2026-07-05T22:00");
    assert.equal(data.url, `${SITE_ORIGIN}/esemenyek/42`);
    assert.equal(data.eventStatus, "https://schema.org/EventScheduled");
    assert.equal(data.location.name, "Nyergestető");
    assert.deepEqual(data.location.address, {
        "@type": "PostalAddress",
        addressLocality: "Csíksomlyó",
        addressRegion: "Hargita megye",
        addressCountry: "RO",
    });
    assert.deepEqual(data.image, [`${SITE_ORIGIN}/images/event.jpg`]);
    assert.deepEqual(data.organizer, {
        "@type": "Organization",
        name: "Hargita Megye Tanácsa",
    });
    // Free-text entry_price is never guessed into an Offer.
    assert.equal("offers" in data, false);
});

test("Event falls back to the settlement as place name", () => {
    const data = eventJsonLd(EVENT);
    assert.equal(data.location.name, "Csíksomlyó");
});

test("single-day Event has no endDate repeat", () => {
    const data = eventJsonLd({
        ...EVENT,
        end_date: "",
        end_time: "",
    });
    assert.equal("endDate" in data, false);
});

test("single-day Event with an end time still ends the same day", () => {
    const data = eventJsonLd({ ...EVENT, end_date: "" });
    assert.equal(data.endDate, "2026-07-04T22:00");
});

test("an event without a start date emits nothing", () => {
    assert.equal(eventJsonLd(null), null);
    assert.equal(eventJsonLd({ ...EVENT, start_date: "" }), null);
    assert.equal(eventJsonLd({ ...EVENT, title: "" }), null);
});
