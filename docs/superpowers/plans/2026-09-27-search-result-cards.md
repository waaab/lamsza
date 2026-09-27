# Search Result Cards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every result in the main search panel the same compact row, with a brown Látnivalók accent and a blue Történelmi székek accent.

**Architecture:** A plain function maps each of the eight search records to a card model, or to nothing when the row cannot be shown. `SearchResultCard` only renders that model. `SearchEngine` is the only caller. `EntryCard` and `WebsiteCard` stay on the other pages.

**Tech Stack:** Svelte 5, existing `canonicalEntryCategory` and `formatDateShort`, `node --test`.

## Global Constraints

- Search panel only. Do not change index, hírek, események, városok, falvak, megyék, székek, homepage widgets, or admin.
- One row: title, optional meta, optional description. Description is the full string from the function and is clipped to one line in CSS.
- Joined meta uses ` · `. One present side is shown alone. Empty lines are omitted.
- Látnivalók: brown heading and brown card border (`--szekely-brown`). Történelmi székek: blue heading and blue card border (`--szekely-blue`). The other six groups are neutral.
- Hover changes the row background and leaves the border color alone.
- Skip a row with an empty title or no href.
- Claim mark only on a claimed website. Websites and news open in a new tab with `rel="nofollow noopener"`.
- Service category label is `canonicalEntryCategory`. Event date is `formatDateShort`; an empty or invalid `start_date` omits the date.
- Do not change the search API, result order, or the Index / Szolgáltatások / Weboldalak filters.
- Tests run with `node --test`.

---

### Task 1: Map each search hit to a card model

**Files:**
- Create: `src/lib/searchResultCard.js`
- Test: `tests/searchResultCard.test.js`

**Interfaces:**
- Consumes: `canonicalEntryCategory` from `src/lib/entryCategory.js`, `formatDateShort` from `src/lib/utils.js`
- Produces: `searchResultCardModel(kind, row) -> null | { href: string, title: string, meta: string, description: string, accent: "none" | "attraction" | "seat", external: boolean, claimed: boolean }`

`kind` is `"website" | "service" | "attraction" | "venue" | "event" | "settlement" | "seat" | "news"`.

- [ ] **Step 1: Write the failing test**

Create `tests/searchResultCard.test.js`:

```javascript
import assert from "node:assert/strict";
import { test } from "node:test";
import { searchResultCardModel } from "../src/lib/searchResultCard.js";
import { formatDateShort } from "../src/lib/utils.js";

test("website: title, domain, description, claim, new tab", () => {
    const card = searchResultCardModel("website", {
        title: "Háromszéki Ágyúsok",
        description: "Közös együttes.",
        domain: "agyusok.ro",
        url: "https://agyusok.ro",
        claimed: true,
    });
    assert.deepEqual(card, {
        href: "https://agyusok.ro",
        title: "Háromszéki Ágyúsok",
        meta: "agyusok.ro",
        description: "Közös együttes.",
        accent: "none",
        external: true,
        claimed: true,
    });
});

test("service: place and category, notes, listing link", () => {
    const card = searchResultCardModel("service", {
        name: "Szent Anna panzió",
        slug: "szent-anna-panzio",
        location: "Csíkszentimre",
        category: "Szállás",
        notes: "Reggeli a tó mellett.",
    });
    assert.equal(card.href, "/bejegyzes/szent-anna-panzio");
    assert.equal(card.title, "Szent Anna panzió");
    assert.equal(card.meta, "Csíkszentimre · Szállás");
    assert.equal(card.description, "Reggeli a tó mellett.");
    assert.equal(card.accent, "none");
    assert.equal(card.external, false);
    assert.equal(card.claimed, false);
});

test("service: category aliases use the catalog label", () => {
    const card = searchResultCardModel("service", {
        name: "Étterem",
        slug: "etterem",
        location: "Sepsiszentgyörgy",
        category: "vendeglo",
        notes: "",
    });
    assert.equal(card.meta, "Sepsiszentgyörgy · Vendéglő");
});

test("service: one side of the meta line has no separator", () => {
    const card = searchResultCardModel("service", {
        name: "Panzió",
        slug: "panzio",
        location: "Csíkszentimre",
        category: "",
        notes: "",
    });
    assert.equal(card.meta, "Csíkszentimre");
    assert.equal(card.description, "");
});

test("attraction: county, description, brown accent, county path", () => {
    const card = searchResultCardModel("attraction", {
        name: "  Szent Anna-tó  ",
        slug: "szent-anna-to",
        county_slug: "hargita",
        county_name: "Hargita",
        description: "Vulkanikus tó.",
    });
    assert.equal(card.href, "/hargita-megye/szent-anna-to");
    assert.equal(card.title, "Szent Anna-tó");
    assert.equal(card.meta, "Hargita");
    assert.equal(card.description, "Vulkanikus tó.");
    assert.equal(card.accent, "attraction");
    assert.equal(card.external, false);
});

test("venue: settlement and kind, venue path", () => {
    const card = searchResultCardModel("venue", {
        name: "Szent Anna kápolna",
        slug: "szent-anna-kapolna",
        county_slug: "hargita",
        settlement_slug: "csikszentimre",
        settlement_name: "Csíkszentimre",
        kind_label: "Kápolna",
    });
    assert.equal(card.href, "/hargita-megye/csikszentimre/helyszin/szent-anna-kapolna");
    assert.equal(card.meta, "Csíkszentimre · Kápolna");
    assert.equal(card.description, "");
    assert.equal(card.accent, "none");
});

test("venue: kind without settlement has no separator", () => {
    const card = searchResultCardModel("venue", {
        name: "Kápolna",
        slug: "kapolna",
        county_slug: "hargita",
        settlement_slug: "csikszentimre",
        settlement_name: "",
        kind_label: "Kápolna",
    });
    assert.equal(card.meta, "Kápolna");
});

test("event: short date and place", () => {
    const card = searchResultCardModel("event", {
        id: 12,
        title: "Szent Anna-napi búcsú",
        start_date: "2026-07-26",
        location_name: "Csíkszentimre",
    });
    assert.equal(card.href, "/esemenyek/12");
    assert.equal(card.meta, `${formatDateShort("2026-07-26")} · Csíkszentimre`);
    assert.equal(card.description, "");
    assert.equal(card.accent, "none");
});

test("event: invalid date omits the date side", () => {
    const card = searchResultCardModel("event", {
        id: 12,
        title: "Búcsú",
        start_date: "not-a-date",
        location_name: "Csíkszentimre",
    });
    assert.equal(card.meta, "Csíkszentimre");
});

test("settlement: county and town path", () => {
    const card = searchResultCardModel("settlement", {
        name: "Csíkszentimre",
        slug: "csikszentimre",
        county_slug: "hargita",
        county: "Hargita",
    });
    assert.equal(card.href, "/hargita-megye/csikszentimre");
    assert.equal(card.meta, "Hargita");
    assert.equal(card.description, "");
    assert.equal(card.accent, "none");
});

test("seat: name only, blue accent", () => {
    const card = searchResultCardModel("seat", {
        name: "Csíkszék",
        slug: "csikszek",
    });
    assert.deepEqual(card, {
        href: "/szekek/csikszek",
        title: "Csíkszék",
        meta: "",
        description: "",
        accent: "seat",
        external: false,
        claimed: false,
    });
});

test("news: title and source, new tab", () => {
    const card = searchResultCardModel("news", {
        title: "Kapaszkodik felfelé a tabellán",
        source: "Bihar Napló",
        link: "https://example.com/hir",
    });
    assert.equal(card.href, "https://example.com/hir");
    assert.equal(card.meta, "Bihar Napló");
    assert.equal(card.description, "");
    assert.equal(card.external, true);
    assert.equal(card.claimed, false);
    assert.equal(card.accent, "none");
});

test("missing link or empty title is skipped", () => {
    assert.equal(searchResultCardModel("website", { title: "Név", url: "" }), null);
    assert.equal(searchResultCardModel("website", { title: "   ", url: "https://example.com" }), null);
    assert.equal(searchResultCardModel("service", { name: "Név", slug: "" }), null);
    assert.equal(searchResultCardModel("attraction", { name: "Tó", slug: "to", county_slug: "" }), null);
    assert.equal(
        searchResultCardModel("venue", {
            name: "Kápolna",
            slug: "kapolna",
            county_slug: "hargita",
            settlement_slug: "",
        }),
        null,
    );
    assert.equal(searchResultCardModel("event", { id: "", title: "Búcsú" }), null);
    assert.equal(searchResultCardModel("settlement", { name: "Falu", slug: "falu", county_slug: "" }), null);
    assert.equal(searchResultCardModel("seat", { name: "Szék", slug: "" }), null);
    assert.equal(searchResultCardModel("news", { title: "Hír", link: "" }), null);
    assert.equal(searchResultCardModel("other", { title: "X" }), null);
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `node --test tests/searchResultCard.test.js`

Expected: FAIL because `src/lib/searchResultCard.js` does not exist.

- [ ] **Step 3: Write the mapper**

Create `src/lib/searchResultCard.js`:

```javascript
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
            claimed: false,
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
            claimed: false,
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
            claimed: false,
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
            claimed: false,
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
            claimed: false,
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
            claimed: false,
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
            claimed: false,
        });
    }
    return null;
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `node --test tests/searchResultCard.test.js`

Expected: PASS, 13 tests.

- [ ] **Step 5: Commit**

```bash
git add src/lib/searchResultCard.js tests/searchResultCard.test.js
git commit -m "$(cat <<'EOF'
Map each search hit onto one compact card model.

EOF
)"
```

---

### Task 2: Render the card

**Files:**
- Create: `src/lib/components/SearchResultCard.svelte`

**Interfaces:**
- Consumes: the object returned by `searchResultCardModel`
- Produces: `SearchResultCard` props `href`, `title`, `meta`, `description`, `accent`, `external`, `claimed`

The row text is already tested. This task only paints it.

- [ ] **Step 1: Add the component**

Create `src/lib/components/SearchResultCard.svelte`:

```svelte
<script>
    import ClaimMark from "$lib/components/ClaimMark.svelte";

    let {
        href,
        title,
        meta = "",
        description = "",
        accent = "none",
        external = false,
        claimed = false,
    } = $props();
</script>

<a
    class="search-result-card"
    class:search-result-card--attraction={accent === "attraction"}
    class:search-result-card--seat={accent === "seat"}
    {href}
    target={external ? "_blank" : undefined}
    rel={external ? "nofollow noopener" : undefined}
>
    <span class="search-result-card__title">
        {title}
        {#if claimed}
            <ClaimMark claimed={true} />
        {/if}
    </span>
    {#if meta}
        <span class="search-result-card__meta">{meta}</span>
    {/if}
    {#if description}
        <span class="search-result-card__desc">{description}</span>
    {/if}
</a>

<style>
    .search-result-card {
        display: block;
        padding: 0.6rem 0.8rem;
        background: var(--bg-body);
        border-radius: 8px;
        border: 1px solid var(--border-color);
        color: var(--text-primary);
        text-decoration: none;
        transition: background 0.2s;
    }

    .search-result-card:hover {
        background: var(--tab-hover-bg);
    }

    .search-result-card--attraction {
        border-color: var(--szekely-brown, #8d6e63);
    }

    .search-result-card--seat {
        border-color: var(--szekely-blue, #42a5f5);
    }

    .search-result-card__title {
        display: block;
        font-weight: 500;
    }

    .search-result-card__meta,
    .search-result-card__desc {
        display: block;
        font-size: var(--text-sm);
    }

    .search-result-card__meta {
        color: var(--text-faint);
    }

    .search-result-card__desc {
        color: var(--text-muted);
        margin-top: 0.25rem;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
</style>
```

- [ ] **Step 2: Compile the component**

Run:

```bash
node --input-type=module -e "
import { compile } from 'svelte/compiler';
import fs from 'fs';
const source = fs.readFileSync('src/lib/components/SearchResultCard.svelte','utf8');
const result = compile(source, { filename: 'SearchResultCard.svelte', generate: 'client' });
if (result.warnings.length) {
  console.log(result.warnings);
  process.exit(1);
}
console.log('compiled');
"
```

Expected: `compiled` and exit 0.

- [ ] **Step 3: Commit**

```bash
git add src/lib/components/SearchResultCard.svelte
git commit -m "$(cat <<'EOF'
Draw the shared search result row.

EOF
)"
```

---

### Task 3: Use the card in the search panel

**Files:**
- Modify: `src/lib/components/SearchEngine.svelte`
- Modify: `src/styles/component-typography.css` (search discover rules around the `.discover-section-title` block)

**Interfaces:**
- Consumes: `searchResultCardModel` and `SearchResultCard` from Tasks 1 and 2

- [ ] **Step 1: Swap the result markup**

In `src/lib/components/SearchEngine.svelte`:

Replace the `EntryCard` and `WebsiteCard` imports with:

```javascript
import SearchResultCard from "$lib/components/SearchResultCard.svelte";
import { searchResultCardModel } from "$lib/searchResultCard.js";
```

Change the utils import from `formatDateShort, weatherIconEmoji` to `weatherIconEmoji`.

Delete `locationToEntry`.

Replace each result loop. Keep the section headings, including `discover-section-title--attractions` and `discover-section-title--szek`. Both website blocks (the website-query block and the later website block) use the same loop:

```svelte
<div class="discover-result-list">
    {#each shownWebsites as website (website.id)}
        {@const card = searchResultCardModel("website", website)}
        {#if card}
            <SearchResultCard {...card} />
        {/if}
    {/each}
</div>
```

Services:

```svelte
<div class="discover-result-list">
    {#each indexEntries as entry (entry.id)}
        {@const card = searchResultCardModel("service", entry)}
        {#if card}
            <SearchResultCard {...card} />
        {/if}
    {/each}
</div>
```

Events, venues, attractions, settlements, seats, and news follow the same list. The kinds and keys are:

- events: `searchResultCardModel("event", ev)`, key `ev.id`
- venues: `searchResultCardModel("venue", venue)`, key `venue.id`
- attractions: `searchResultCardModel("attraction", att)`, key `att.id`
- settlements: `searchResultCardModel("settlement", loc)`, key `loc.id`. Do not call `locationToEntry`.
- seats: `searchResultCardModel("seat", seat)`, key `seat.id`
- news: `searchResultCardModel("news", item)`, key `item.link`

Remove the inner `<span class="discover-event-title">` markup. The card replaces the whole `<a class="discover-*-card">`.

- [ ] **Step 2: Replace the old row CSS**

In the `<style>` block of `SearchEngine.svelte`, delete `.discover-event-list` through `.discover-szek-title`, including the attraction and seat card border rules.

Keep, or add if the delete removed them:

```css
.discover-section-title--attractions {
    color: var(--szekely-brown, #6d4c41);
}

.discover-section-title--szek {
    color: var(--szekely-blue, #1565c0);
}

.discover-result-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}
```

In `src/styles/component-typography.css`, delete these rules and leave `.discover-section-title` in place:

```css
.discover-event-meta,
.discover-news-source,
.discover-attraction-meta {
  font-size: var(--text-sm);
}

.discover-attraction-desc {
  font-size: var(--text-sm);
}

.discover-szek-title {
  font-size: var(--text-base);
}
```

- [ ] **Step 3: Confirm the old cards are gone from search**

Run:

```bash
node --test tests/searchResultCard.test.js
rg -n "EntryCard|WebsiteCard|locationToEntry|discover-event-card|discover-attraction-card|discover-szek-card" src/lib/components/SearchEngine.svelte
```

Expected: the test file passes. `rg` prints no matches.

- [ ] **Step 4: Commit**

```bash
git add src/lib/components/SearchEngine.svelte src/styles/component-typography.css
git commit -m "$(cat <<'EOF'
Show every search hit as the same compact row.

EOF
)"
```
