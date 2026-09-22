# Listing Nearby / Related / History Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** On `/bejegyzes/[slug]`, show Közeli helyek, Kapcsolódó bejegyzések, and Böngészési előzmények under the profile and events.

**Architecture:** `GET /api/entry/related?slug=` returns `{ nearby, related }` (max 8 each). Nearby is same settlement, then county pad. Related is same category in that county, excluding this listing and nearby ids. History is `localStorage` via `src/lib/entryHistory.js` (no server). The listing page fetches related after the entry loads, records a visit, and hides empty blocks.

**Tech Stack:** Go `net/http` + Postgres, Svelte 5 for new components, listing `+page.svelte` stays on its current `let` / `$:` style, Node `node --test`, no carousel npm packages.

## Global Constraints

- Hungarian copy exactly: **Közeli helyek**, **Kapcsolódó bejegyzések**, **Böngészési előzmények**.
- Nearby cap **8**; related cap **8**; history store **12**, show **8**.
- `localStorage` key **`lamsza_entry_history`**. Device only. No auth. No cookie-policy change.
- Related API 404 if slug missing/unknown. Related fetch failure must not break the profile.
- No “see all”, no GPS radius, no third-party carousel, no admin UI.
- Do not commit unless the user explicitly asks (user git rule). Skip every “Commit” step unless they have asked.
- New `.svelte` files: Svelte 5 runes. Do not migrate `src/routes/(public)/bejegyzes/[slug]/+page.svelte` to runes. Run Svelte MCP `svelte-autofixer` on new/edited `.svelte` files until clean.

---

### Task 1: Browsing-history helper

**Files:**
- Create: `src/lib/entryHistory.js`
- Create: `tests/entryHistory.test.js`

**Interfaces:**
- Consumes: `localStorage` in the browser; no-ops / `[]` when missing
- Produces:
  - `export const ENTRY_HISTORY_STORAGE_KEY = "lamsza_entry_history"`
  - `export const ENTRY_HISTORY_STORE_MAX = 12`
  - `export const ENTRY_HISTORY_SHOW_MAX = 8`
  - `export function normalizeHistory(raw): { slug: string, name: string, category: string, location: string, photo: string }[]`
  - `export function readHistory():` same shape
  - `export function recordHistoryVisit(item):` same shape (returns stored list)
  - `export function historyForDisplay(items, currentSlug):` same shape (skip current, max 8)

- [ ] **Step 1: Write the failing tests**

Create `tests/entryHistory.test.js`:

```js
import assert from "node:assert/strict";
import { afterEach, beforeEach, test } from "node:test";
import {
    ENTRY_HISTORY_SHOW_MAX,
    ENTRY_HISTORY_STORAGE_KEY,
    ENTRY_HISTORY_STORE_MAX,
    historyForDisplay,
    normalizeHistory,
    readHistory,
    recordHistoryVisit,
} from "../src/lib/entryHistory.js";

const store = new Map();

function installLocalStorage() {
    globalThis.localStorage = {
        getItem(key) {
            return store.has(key) ? store.get(key) : null;
        },
        setItem(key, value) {
            store.set(key, String(value));
        },
        removeItem(key) {
            store.delete(key);
        },
    };
}

beforeEach(() => {
    store.clear();
    installLocalStorage();
});

afterEach(() => {
    delete globalThis.localStorage;
});

test("normalizeHistory: empty and corrupt become []", () => {
    assert.deepEqual(normalizeHistory(null), []);
    assert.deepEqual(normalizeHistory("nope"), []);
    assert.deepEqual(normalizeHistory([{ name: "No slug" }]), []);
});

test("recordHistoryVisit: prepends, dedupes by slug, caps store", () => {
    for (let i = 0; i < ENTRY_HISTORY_STORE_MAX + 3; i++) {
        recordHistoryVisit({
            slug: `e-${i}`,
            name: `N${i}`,
            category: "Vendéglő",
            location: "Kézdivásárhely",
            photo: "",
        });
    }
    const all = readHistory();
    assert.equal(all.length, ENTRY_HISTORY_STORE_MAX);
    assert.equal(all[0].slug, `e-${ENTRY_HISTORY_STORE_MAX + 2}`);
    recordHistoryVisit({ slug: "e-0", name: "N0 again" });
    const again = readHistory();
    assert.equal(again[0].slug, "e-0");
    assert.equal(again.filter((x) => x.slug === "e-0").length, 1);
    assert.equal(localStorage.getItem(ENTRY_HISTORY_STORAGE_KEY)?.[0], "[");
});

test("historyForDisplay: skips current and caps show", () => {
    const items = Array.from({ length: 10 }, (_, i) => ({
        slug: `e-${i}`,
        name: `N${i}`,
        category: "",
        location: "",
        photo: "",
    }));
    const shown = historyForDisplay(items, "e-0");
    assert.equal(shown.length, ENTRY_HISTORY_SHOW_MAX);
    assert.equal(shown[0].slug, "e-1");
    assert.ok(!shown.some((x) => x.slug === "e-0"));
});

test("readHistory: corrupt JSON is []", () => {
    localStorage.setItem(ENTRY_HISTORY_STORAGE_KEY, "{not-json");
    assert.deepEqual(readHistory(), []);
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `node --test tests/entryHistory.test.js`  
Expected: `ERR_MODULE_NOT_FOUND` for `../src/lib/entryHistory.js`

- [ ] **Step 3: Write `src/lib/entryHistory.js`**

```js
export const ENTRY_HISTORY_STORAGE_KEY = "lamsza_entry_history";
export const ENTRY_HISTORY_STORE_MAX = 12;
export const ENTRY_HISTORY_SHOW_MAX = 8;

function asItem(raw) {
    if (!raw || typeof raw !== "object") return null;
    const slug = String(raw.slug ?? "").trim();
    const name = String(raw.name ?? "").trim();
    if (!slug || !name) return null;
    return {
        slug,
        name,
        category: String(raw.category ?? "").trim(),
        location: String(raw.location ?? "").trim(),
        photo: String(raw.photo ?? "").trim(),
    };
}

export function normalizeHistory(raw) {
    let arr = raw;
    if (typeof raw === "string") {
        try {
            arr = JSON.parse(raw);
        } catch {
            return [];
        }
    }
    if (!Array.isArray(arr)) return [];
    const out = [];
    const seen = new Set();
    for (const row of arr) {
        const item = asItem(row);
        if (!item || seen.has(item.slug)) continue;
        seen.add(item.slug);
        out.push(item);
        if (out.length >= ENTRY_HISTORY_STORE_MAX) break;
    }
    return out;
}

export function readHistory() {
    if (typeof localStorage === "undefined") return [];
    try {
        return normalizeHistory(localStorage.getItem(ENTRY_HISTORY_STORAGE_KEY));
    } catch {
        return [];
    }
}

export function recordHistoryVisit(raw) {
    const item = asItem(raw);
    const prev = readHistory().filter((row) => !item || row.slug !== item.slug);
    const next = item ? [item, ...prev].slice(0, ENTRY_HISTORY_STORE_MAX) : prev;
    if (typeof localStorage !== "undefined") {
        try {
            localStorage.setItem(ENTRY_HISTORY_STORAGE_KEY, JSON.stringify(next));
        } catch {
            /* quota / private mode */
        }
    }
    return next;
}

export function historyForDisplay(items, currentSlug) {
    const skip = String(currentSlug ?? "").trim();
    return normalizeHistory(items)
        .filter((row) => row.slug !== skip)
        .slice(0, ENTRY_HISTORY_SHOW_MAX);
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `node --test tests/entryHistory.test.js`  
Expected: 4 passing

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 2: `GET /api/entry/related`

**Files:**
- Create: `backend/internal/handlers/entry_related.go`
- Modify: `backend/main.go` — register next to `/api/entry`
- Modify: `backend/handlers_test.go` — register on `testMux`; add `TestEntryRelated`

**Interfaces:**
- Consumes: `slug` query; tables `entries`, `settlements`, `counties`, `entry_categories`
- Produces: JSON `{ "nearby": RelatedEntry[], "related": RelatedEntry[] }`
- `RelatedEntry`: `id` (JSON number or string matching existing entry `id` encoding — use the same `models` id type as `Entry.ID` which is `string`), `name`, `slug`, `category`, `location`, `location_slug`, `county_slug`, `photos` (`json.RawMessage`, sanitized array)
- Constants: `relatedEntryLimit = 8`
- 400 if `slug` empty; 404 if not found; GET only

- [ ] **Step 1: Register a stub route and write the failing test**

In `backend/main.go` after `/api/entry`:

```go
mux.HandleFunc("/api/entry/related", pub(handlers.HandleEntryRelated))
```

In `handlers_test.go` `init()`, after `/api/entry`:

```go
testMux.HandleFunc("/api/entry/related", pub(handlers.HandleEntryRelated))
```

Add `TestEntryRelated` at the end of `handlers_test.go` (before `formatID`). Logic:

1. `doAnonRequest GET /api/entry/related` → 400  
2. `doAnonRequest GET /api/entry/related?slug=does-not-exist-xyz` → 404  
3. GET `/api/locations`, pick two settlements with the **same** `county` (or `county_slug`). If fewer than two, `t.Skip`.  
4. POST three admin entries (unique names with `RelatedInteg` prefix):
   - `A` at loc1, category `Egyéb`
   - `B` at loc1, category `Egyéb` (same town)
   - `C` at loc2, category `Egyéb` (same county, other town)
5. GET `/api/entry/related?slug=` + A’s slug (anon is fine):
   - `nearby` contains B (same settlement) before any county pad
   - `nearby` must not contain A
   - if C appears in `nearby`, it is after same-settlement rows
   - `related` must not contain A or any `nearby` id
   - `related` items (if any) share A’s category
   - `len(nearby) <= 8`, `len(related) <= 8`
6. DELETE A, B, C by id

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 120s -run TestEntryRelated`  
Expected: FAIL (undefined `HandleEntryRelated` or 404/empty)

- [ ] **Step 3: Implement `HandleEntryRelated`**

Create `backend/internal/handlers/entry_related.go`:

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/db"
)

const relatedEntryLimit = 8

type relatedEntry struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	Category     string          `json:"category"`
	Location     string          `json:"location"`
	LocationSlug string          `json:"location_slug"`
	CountySlug   string          `json:"county_slug"`
	Photos       json.RawMessage `json:"photos"`
}

func HandleEntryRelated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Missing slug", http.StatusBadRequest)
		return
	}

	var id, locationID, categoryID int
	var countySlug string
	err := db.DB.QueryRow(`
		SELECT e.id, e.location_id, COALESCE(e.category_id, 0), c.slug
		FROM entries e
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id
		WHERE e.slug = $1`, slug).Scan(&id, &locationID, &categoryID, &countySlug)
	if err != nil {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	nearby := queryRelatedEntries(`
		SELECT e.id::text, e.name, COALESCE(e.slug,''), COALESCE(ec.name,''),
			s.name, s.slug, c.slug, COALESCE(e.photos, '[]'::jsonb)
		FROM entries e
		JOIN settlements s ON e.location_id = s.id
		JOIN counties c ON s.county_id = c.id
		LEFT JOIN entry_categories ec ON e.category_id = ec.id
		WHERE e.location_id = $1 AND e.id <> $2
		ORDER BY LOWER(e.name) ASC, e.id ASC
		LIMIT $3`, locationID, id, relatedEntryLimit)

	if len(nearby) < relatedEntryLimit {
		exclude := relatedIDs(nearby)
		exclude = append(exclude, id)
		pad := queryRelatedEntries(`
			SELECT e.id::text, e.name, COALESCE(e.slug,''), COALESCE(ec.name,''),
				s.name, s.slug, c.slug, COALESCE(e.photos, '[]'::jsonb)
			FROM entries e
			JOIN settlements s ON e.location_id = s.id
			JOIN counties c ON s.county_id = c.id
			LEFT JOIN entry_categories ec ON e.category_id = ec.id
			WHERE c.slug = $1 AND e.id <> ALL($2)
			ORDER BY LOWER(e.name) ASC, e.id ASC
			LIMIT $3`, countySlug, exclude, relatedEntryLimit-len(nearby))
		nearby = append(nearby, pad...)
	}

	relExclude := relatedIDs(nearby)
	relExclude = append(relExclude, id)
	related := []relatedEntry{}
	if categoryID > 0 {
		related = queryRelatedEntries(`
			SELECT e.id::text, e.name, COALESCE(e.slug,''), COALESCE(ec.name,''),
				s.name, s.slug, c.slug, COALESCE(e.photos, '[]'::jsonb)
			FROM entries e
			JOIN settlements s ON e.location_id = s.id
			JOIN counties c ON s.county_id = c.id
			LEFT JOIN entry_categories ec ON e.category_id = ec.id
			WHERE e.category_id = $1 AND c.slug = $2 AND e.id <> ALL($3)
			ORDER BY LOWER(e.name) ASC, e.id ASC
			LIMIT $4`, categoryID, countySlug, relExclude, relatedEntryLimit)
	}

	if nearby == nil {
		nearby = []relatedEntry{}
	}
	if related == nil {
		related = []relatedEntry{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"nearby":  nearby,
		"related": related,
	})
}
```

Also add `queryRelatedEntries(sql string, args ...any) []relatedEntry` in the same file: scan eight columns, `Photos: sanitizePhotos(photos)`, skip failed rows, return empty slice not nil.

Use `strconv.Atoi` on `row.ID` when building exclude ids. For `<> ALL`, pass `pq.Array(ids)` as `$n::int[]`. If the exclude list would be empty, pass `[]int{0}` so Postgres still gets a non-null array (real ids start at 1). Callers must pass `pq.Array(relExclude)` into the query, not a raw `[]int`.

- [ ] **Step 4: Run tests**

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestEntryRelated|TestAdminEntriesCRUD'`  
Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 3: Presentational UI

**Files:**
- Create: `src/lib/components/EntryRelatedLinks.svelte`
- Create: `src/lib/components/EntryHistoryStrip.svelte`

**Interfaces:**
- Consumes:
  - Related links: `{ nearby: relatedEntry[], related: relatedEntry[], currentLocationSlug: string }`
  - History strip: `{ items: historyItem[] }` already passed through `historyForDisplay`
- Produces: visible Hungarian headings; hide a column/section when its array is empty; hide the whole links component when both arrays empty

`relatedEntry` shape matches the API. For county-padded nearby rows (`location_slug !== currentLocationSlug`), render settlement name after the link.

History cards: map each item to `EntryCard` with `layout="grid"` and

```js
{
  slug: item.slug,
  name: item.name,
  category: item.category,
  location: item.location,
  photos: item.photo
    ? [{ url: item.photo, alt: item.name, title: item.name, width: 160, height: 120 }]
    : [],
}
```

Strip: `overflow-x: auto`, flex row, gap. Prev/next `button` type=button, `aria-label` Előző / Következő, `scrollBy({ left: ±240, behavior: "smooth" })` on a bound scroller. Show those buttons only if `scrollWidth > clientWidth` (check on mount and `svelte:window onresize`). No npm carousel.

Breakpoint: stack the two link columns below **992px**.

- [ ] **Step 1: Create `EntryRelatedLinks.svelte`**

```svelte
<script>
    let { nearby = [], related = [], currentLocationSlug = "" } = $props();
    let here = $derived(String(currentLocationSlug ?? "").trim());
    let showNearby = $derived(Array.isArray(nearby) && nearby.length > 0);
    let showRelated = $derived(Array.isArray(related) && related.length > 0);
</script>

{#if showNearby || showRelated}
    <div class="entry-related">
        {#if showNearby}
            <section class="entry-related__col" aria-labelledby="entry-nearby-title">
                <h2 id="entry-nearby-title" class="entry-related__title">Közeli helyek</h2>
                <ul class="entry-related__list">
                    {#each nearby as row (row.slug || row.id)}
                        <li>
                            <a href="/bejegyzes/{row.slug}">{row.name}</a>
                            {#if here && row.location_slug && row.location_slug !== here && row.location}
                                <span class="entry-related__place">{row.location}</span>
                            {/if}
                        </li>
                    {/each}
                </ul>
            </section>
        {/if}
        {#if showRelated}
            <section class="entry-related__col" aria-labelledby="entry-related-title">
                <h2 id="entry-related-title" class="entry-related__title">Kapcsolódó bejegyzések</h2>
                <ul class="entry-related__list">
                    {#each related as row (row.slug || row.id)}
                        <li>
                            <a href="/bejegyzes/{row.slug}">{row.name}</a>
                        </li>
                    {/each}
                </ul>
            </section>
        {/if}
    </div>
{/if}

<style>
    .entry-related {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 1.5rem 2rem;
        padding-top: 0.5rem;
        border-top: 1px solid var(--border-color);
    }
    @media (max-width: 991px) {
        .entry-related {
            grid-template-columns: 1fr;
        }
    }
    .entry-related__title {
        margin: 0 0 0.65rem;
        font-size: 1.05rem;
    }
    .entry-related__list {
        list-style: none;
        margin: 0;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }
    .entry-related__list a {
        color: var(--szekely-blue, #1d4ed8);
        font-weight: 600;
        text-decoration: none;
    }
    .entry-related__list a:hover {
        text-decoration: underline;
    }
    .entry-related__place {
        margin-left: 0.35rem;
        color: var(--text-muted);
        font-weight: 400;
    }
</style>
```

- [ ] **Step 2: Create `EntryHistoryStrip.svelte`** with `EntryCard`, scroller `bind:this`, overflow-gated nav, heading **Böngészési előzmények**. Return nothing if `items.length === 0`. Card wrapper min-width ~16rem so they sit in a row.

- [ ] **Step 3: Autofixer**

Run Svelte MCP `svelte-autofixer` on both files (`desired_svelte_version: 5`) until issues is empty.

- [ ] **Step 4: Commit**

Skip unless the user asked to commit.

---

### Task 4: Wire the listing page

**Files:**
- Modify: `src/routes/(public)/bejegyzes/[slug]/+page.svelte`

**Interfaces:**
- Consumes: `GET /api/entry/related?slug=` via `apiFetch`; Task 1 helpers; Task 3 components
- Produces: blocks under `EventsWidget`; history recorded once per successful entry load

- [ ] **Step 1: After `entry` is set in `fetchEntry`, call related + record history**

Keep existing `$:` / `let` style.

```js
import EntryRelatedLinks from "$lib/components/EntryRelatedLinks.svelte";
import EntryHistoryStrip from "$lib/components/EntryHistoryStrip.svelte";
import { normalizePhotos } from "$lib/entryPhotos.js";
import {
    historyForDisplay,
    readHistory,
    recordHistoryVisit,
} from "$lib/entryHistory.js";

let nearby = [];
let related = [];
let historyItems = [];

// inside fetchEntry success, after entry = data:
nearby = [];
related = [];
try {
    const rel = await apiFetch(
        `/api/entry/related?slug=${encodeURIComponent(slug)}`,
    );
    nearby = Array.isArray(rel?.nearby) ? rel.nearby : [];
    related = Array.isArray(rel?.related) ? rel.related : [];
} catch {
    nearby = [];
    related = [];
}
recordHistoryVisit({
    slug: data.slug,
    name: data.name,
    category: data.category,
    location: data.location,
    photo: normalizePhotos(data.photos)[0]?.url || "",
});
historyItems = historyForDisplay(readHistory(), data.slug);
```

On error/not-found, set `nearby = []`, `related = []`, `historyItems = []` and do **not** record.

- [ ] **Step 2: Render after EventsWidget**

```svelte
<EventsWidget organizerName={entry.name} />
<EntryRelatedLinks {nearby} {related} currentLocationSlug={entry.location_slug} />
<EntryHistoryStrip items={historyItems} />
```

Do not add extra skeleton for these blocks in v1 (they appear after the profile is already shown).

- [ ] **Step 3: Autofixer on `+page.svelte` if the MCP reports issues** — fix real issues; ignore unrelated pre-existing ones.

- [ ] **Step 4: Commit**

Skip unless the user asked to commit.

---

### Task 5: Browser verification

**Files:** none (manual)

- [ ] **Step 1: Restart Go if it is serving old binary** so `/api/entry/related` exists. Vite HMR is enough for the page.

- [ ] **Step 2: Open a listing that has same-town neighbors (e.g. Manifesto).** Confirm **Közeli helyek** links go to `/bejegyzes/…` and work. Confirm **Kapcsolódó bejegyzések** is same category, no name repeated from Nearby. County-padded nearby rows show the other settlement.

- [ ] **Step 3: Open a second listing, then go back to the first.** **Böngészési előzmények** shows the second (and any others), not the current page. Cards link correctly. If the strip overflows, prev/next appear.

- [ ] **Step 4: `curl -s -o /dev/null -w '%{http_code}' 'http://127.0.0.1:3000/api/entry/related'`** Expected: `400`. Unknown slug: `404`.

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

## Spec coverage

| Spec | Task |
|---|---|
| Nearby same settlement then county pad, cap 8, exclude self | 2 |
| Related same category + county, disjoint from nearby, cap 8 | 2 |
| 404 unknown slug; fetch failure hides blocks | 2, 4 |
| History localStorage key, prepend, dedupe, store 12, show 8, skip current | 1, 4 |
| Hungarian headings; hide empty | 3 |
| Two columns / stack at 992px; history strip + native scroll | 3 |
| Placement below profile + events | 4 |
| Browser check | 5 |
| No GPS, no see-all, no carousel lib, no people-also-viewed | Global / omitted |
