# Favorite places Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A signed-in user can favorite settlements, látnivalók, listings, and events. The homepage shows weather and events for each favorite settlement, and weather for each favorite látnivaló that has coordinates.

**Later change (24 Sep 2026):** favorite settlements do not choose homepage weather or the events ticker. Those follow the saved settlement. Favorite látnivalók with coordinates can still add a weather widget.

**Architecture:** `user_favorites` stores `(user_id, entity_type, entity_id)`. Public pages toggle one row. `/profil` gains the **Kedvenc helyek** tab. The homepage, when signed in, repeats widgets in `created_at` order and falls back to `my_location_slug` when there is no favorite settlement.

**Tech Stack:** Go `net/http` + Postgres, Svelte 5 runes for the new button, existing pages stay on their current style, Node `node --test`, Go `go test`.

**Spec:** `docs/superpowers/specs/2026-09-22-user-profile-design.md`

**Depends on:** `docs/superpowers/plans/2026-09-22-user-profile-account.md` merged in the working tree.

## Global Constraints

- Favorite types are exactly `settlement`, `attraction`, `entry`, `event`.
- No cap on how many favorites a user may save.
- Favorite listings and events do not add homepage widgets.
- A látnivaló without coordinates stays on the profile only.
- Signed out, or signed in with no favorite settlement, homepage weather and events use the admin `my_location_slug`.
- A favorite whose target row is gone does not appear in `GET /api/account/favorites`.
- Duplicate add is ignored (200 with the existing row).
- Hungarian UI. Tab label **Kedvenc helyek**. Groups: **Települések**, **Látnivalók**, **Bejegyzések**, **Események**.
- New `.svelte` files use Svelte 5 runes. Run Svelte MCP `svelte-autofixer` until clean.
- Do not commit unless the user explicitly asks. Skip every Commit step unless they have asked.

---

### Task 1: Favorites table and API

**Files:**
- Create: `backend/internal/account/favorites.go`
- Modify: `backend/internal/account/prefs_test.go`
- Modify: `backend/main.go`
- Modify: `backend/handlers_test.go`

**Interfaces:**
- Consumes: `auth.UserFromRequest`, `account.Migrate`
- Produces:
  - table `user_favorites(user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, entity_type VARCHAR(20) NOT NULL, entity_id INT NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, entity_type, entity_id))`
  - `func NormalizeFavorite(entityType string, entityID int) (string, int, error)`
  - `GET /api/account/favorites` → `{ settlements: [], attractions: [], entries: [], events: [] }` each item `{ id, name, slug, county_slug, latitude, longitude, created_at }` using the columns that exist for that type. `latitude` and `longitude` are null when the type has none.
  - `POST /api/account/favorites` body `{ "type", "id" }` inserts or returns the existing row
  - `DELETE /api/account/favorites?type=&id=` removes one owned row

- [ ] **Step 1: Write the failing test**

```go
func TestNormalizeFavorite(t *testing.T) {
	typ, id, err := NormalizeFavorite("attraction", 4)
	if err != nil || typ != "attraction" || id != 4 {
		t.Fatalf("got %s %d %v", typ, id, err)
	}
	if _, _, err := NormalizeFavorite("county", 1); err == nil {
		t.Fatal("expected unknown type to fail")
	}
	if _, _, err := NormalizeFavorite("entry", 0); err == nil {
		t.Fatal("expected id 0 to fail")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/account/ -count=1 -run TestNormalizeFavorite`

Expected: FAIL with `undefined: NormalizeFavorite`

- [ ] **Step 3: Write minimal implementation**

`NormalizeFavorite` allows only `settlement`, `attraction`, `entry`, `event`, and `entityID > 0`.

`GET` joins:

- `settlement` → `settlements` joined to `counties` for `county_slug`, and to `geo_locations` on `settlements.location_id` for `latitude` and `longitude`
- `attraction` → `attractions` joined to `counties` for `county_slug`, and to `geo_locations` on `attractions.location_id` for `latitude` and `longitude` (null when `location_id` is null)
- `entry` → `entries` plus settlement name and county slug
- `event` → `events` plus title as `name`

`LEFT JOIN` and drop rows where the target `id` is null. Order each group by `user_favorites.created_at ASC`.

`POST` uses `INSERT ... ON CONFLICT DO NOTHING` then selects the row. Unknown type or missing target → 400.

Register the route in `main.go` and `handlers_test.go`. Create the table from `account.Migrate`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/account/ -count=1 -run TestNormalizeFavorite`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 2: Favorite toggle

**Files:**
- Create: `src/lib/favorites.js`
- Create: `src/lib/components/FavoriteButton.svelte`
- Create: `tests/favorites.test.js`

**Interfaces:**
- Consumes: `POST` and `DELETE /api/account/favorites`, `auth`
- Produces:
  - `export const FAVORITE_TYPES = ["settlement", "attraction", "entry", "event"]`
  - `export function favoriteKey(type, id)` → `"settlement:12"`
  - `export function isFavorite(list, type, id)` → boolean
  - `FavoriteButton` props `type`, `id`, `active`, and callback `ontoggle`

- [ ] **Step 1: Write the failing test**

```js
import assert from "node:assert/strict";
import test from "node:test";
import { favoriteKey, isFavorite } from "../src/lib/favorites.js";

test("isFavorite matches type and id", () => {
    const list = [{ type: "event", id: 9 }];
    assert.equal(isFavorite(list, "event", 9), true);
    assert.equal(isFavorite(list, "event", 3), false);
    assert.equal(favoriteKey("attraction", 4), "attraction:4");
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test tests/favorites.test.js`

Expected: FAIL, module missing

- [ ] **Step 3: Write minimal implementation**

Implement the helpers. `FavoriteButton.svelte` is a button with `aria-pressed={active}` and label **Kedvenchez adás** when inactive, **Kedvenc eltávolítása** when active. It does not call fetch itself. The parent passes `ontoggle`.

Replace the non-functional heart in `src/routes/(public)/esemenyek/[id]/+page.svelte` with `FavoriteButton` `type="event"`. On click, if logged out, open the existing sign-in dialog. If logged in, POST or DELETE, then set `active` from the result.

Add the same button to:

- `src/routes/(public)/[countySlug]-megye/[slug]/+page.svelte` — `settlement` when `settlementData` is set, `attraction` when `attractionData` is set
- `src/routes/(public)/bejegyzes/[slug]/+page.svelte` — `entry`

Load `GET /api/account/favorites` once when logged in and pass `active` from `isFavorite`.

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test tests/favorites.test.js`

Expected: PASS

Run Svelte autofixer on `FavoriteButton.svelte` and the three edited pages until clean.

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 3: Kedvenc helyek tab

**Files:**
- Modify: `src/lib/accountPrefs.js`
- Modify: `src/routes/(public)/profil/+page.svelte`
- Modify: `tests/accountPrefs.test.js`

**Interfaces:**
- Consumes: `GET/DELETE /api/account/favorites`, `profileTabIds`
- Produces: `profileTabIds` includes `"kedvencek"` after `"elozmenyek"`

- [ ] **Step 1: Write the failing test**

Update `tests/accountPrefs.test.js` so the account tabs stay in place and **Kedvenc helyek** is last:

```js
assert.deepEqual(profileTabIds, [
    "profil",
    "beallitasok",
    "linkjeim",
    "elozmenyek",
    "kedvencek",
]);
```

Run this plan before the listings plan. The listings plan inserts `bejegyzeseim` and updates this assertion.

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test tests/accountPrefs.test.js`

Expected: FAIL, array mismatch

- [ ] **Step 3: Write minimal implementation**

Append `"kedvencek"` in `profileTabIds`. The tab title is **Kedvenc helyek**. Render four groups from the GET payload. Each row links to the public URL:

- settlement and attraction: `/${county_slug}-megye/${slug}`
- entry: `/bejegyzes/${slug}`
- event: `/esemenyek/${id}`

A **Eltávolítás** button calls DELETE and removes the row. Empty group: hide that heading.

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test tests/accountPrefs.test.js tests/favorites.test.js`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 4: Homepage widgets follow favorites

**Files:**
- Modify: `src/routes/(public)/+page.svelte`
- Create: `src/lib/favoriteHomepage.js`
- Create: `tests/favoriteHomepage.test.js`

**Interfaces:**
- Consumes: favorites payload from Task 1, `my_location_slug` from `/api/config/public`
- Produces:
  - `export function homepageSettlements(favorites, siteDefault)` → `[{ slug, name }]`
  - `export function homepageAttractionWeather(favorites)` → `[{ slug, name, lat, lon }]` only when both coordinates are finite numbers
  - Homepage renders `WeatherWidget` + `EventsWidget` once per settlement from `homepageSettlements`
  - Homepage renders `WeatherWidget` with `lat` and `lon` once per `homepageAttractionWeather` item

- [ ] **Step 1: Write the failing test**

```js
import assert from "node:assert/strict";
import test from "node:test";
import { homepageAttractionWeather, homepageSettlements } from "../src/lib/favoriteHomepage.js";

test("no favorite settlement uses the site default", () => {
    const rows = homepageSettlements({ settlements: [] }, { slug: "csikszereda", name: "Csíkszereda" });
    assert.deepEqual(rows, [{ slug: "csikszereda", name: "Csíkszereda" }]);
});

test("favorite settlements keep added order and skip the default", () => {
    const rows = homepageSettlements(
        { settlements: [{ slug: "udvarhely", name: "Székelyudvarhely" }, { slug: "csikszereda", name: "Csíkszereda" }] },
        { slug: "csikszereda", name: "Csíkszereda" },
    );
    assert.deepEqual(rows.map((r) => r.slug), ["udvarhely", "csikszereda"]);
});

test("attraction weather skips missing coordinates", () => {
    const rows = homepageAttractionWeather({
        attractions: [
            { slug: "to", name: "Tó", latitude: 46.1, longitude: 25.8 },
            { slug: "var", name: "Vár", latitude: null, longitude: null },
        ],
    });
    assert.equal(rows.length, 1);
    assert.equal(rows[0].slug, "to");
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test tests/favoriteHomepage.test.js`

Expected: FAIL, module missing

- [ ] **Step 3: Write minimal implementation**

Implement the two functions. In `+page.svelte`, when `$auth.loggedIn`, load favorites and replace the single `WeatherWidget` / `EventsWidget` pair with `{#each homepageSettlements(...) as place}` and `{#each homepageAttractionWeather(...) as place}`. When logged out, keep the current single site-default pair. Do not render widgets for `entries` or `events`.

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test tests/favoriteHomepage.test.js tests/favorites.test.js`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.
