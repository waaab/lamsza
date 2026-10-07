import assert from "node:assert/strict";
import { test } from "node:test";
import {
    absoluteUrl,
    canonicalPath,
    DEFAULT_SEO_DESCRIPTION,
    pageSlugFromPath,
    resolveSeo,
    SITE_ORIGIN,
} from "../src/lib/seo.js";

test("canonical host is the production apex, never www or a dev host", () => {
    assert.equal(SITE_ORIGIN, "https://lamsza.com");
});

test("canonicalPath drops query, hash and trailing slash", () => {
    assert.equal(canonicalPath("/"), "/");
    assert.equal(canonicalPath("/hirek/"), "/hirek");
    assert.equal(canonicalPath("/iranyelvek/sutik/?utm_source=fb"), "/iranyelvek/sutik");
    assert.equal(canonicalPath("/megyek#lista"), "/megyek");
    assert.equal(canonicalPath("megyek"), "/megyek");
});

test("absoluteUrl keeps absolute input and prefixes the rest", () => {
    assert.equal(absoluteUrl("/og-image.png"), "https://lamsza.com/og-image.png");
    assert.equal(absoluteUrl("og-image.png"), "https://lamsza.com/og-image.png");
    assert.equal(
        absoluteUrl("https://cdn.example.com/a.png"),
        "https://cdn.example.com/a.png",
    );
});

test("pageSlugFromPath reuses the hero page keys, parent first", () => {
    assert.equal(pageSlugFromPath("/"), "home");
    assert.equal(pageSlugFromPath("/hirek"), "hirek");
    assert.equal(pageSlugFromPath("/iranyelvek/sutik"), "iranyelvek/sutik");
    // Unknown deep path falls back to the nearest parent that has defaults.
    assert.equal(pageSlugFromPath("/szekek/csikszek"), "szekek");
    assert.equal(pageSlugFromPath("/nincs-ilyen"), "");
});

test("resolveSeo takes title and description from the page defaults", () => {
    const seo = resolveSeo({ pathname: "/hirek" });
    assert.equal(seo.title, "Friss hírek erdélyi forrásból - Lámsza");
    assert.equal(seo.description, "Helyi hírcsatornák legfrissebb hírei időrendben.");
    assert.equal(seo.canonical, "https://lamsza.com/hirek");
    assert.equal(seo.image, "https://lamsza.com/og-image.png");
    assert.equal(seo.type, "website");
    assert.equal(seo.locale, "hu_HU");
});

test("resolveSeo does not repeat the site name when the title has it", () => {
    assert.equal(resolveSeo({ pathname: "/" }).title, "Na Lámsza!");
});

test("resolveSeo lets a page override every field", () => {
    const seo = resolveSeo({
        pathname: "/esemenyek/12",
        title: "Csíkszeredai néptánctalálkozó",
        description: "Kétnapos találkozó a Mikó-várnál.",
        image: "/api/media/event-images/12.png",
        type: "article",
    });
    assert.equal(seo.title, "Csíkszeredai néptánctalálkozó - Lámsza");
    assert.equal(seo.description, "Kétnapos találkozó a Mikó-várnál.");
    assert.equal(seo.canonical, "https://lamsza.com/esemenyek/12");
    assert.equal(seo.image, "https://lamsza.com/api/media/event-images/12.png");
    assert.equal(seo.type, "article");
});

test("an unknown path still gets a usable title and description", () => {
    const seo = resolveSeo({ pathname: "/nincs-ilyen" });
    assert.equal(seo.title, "Na Lámsza! - Erdélyi magyar startlap és kereső");
    assert.equal(seo.description, DEFAULT_SEO_DESCRIPTION);
    assert.equal(seo.canonical, "https://lamsza.com/nincs-ilyen");
});
