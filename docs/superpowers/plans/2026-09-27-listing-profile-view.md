# Listing profile view Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The public listing page shows the short profile, opens phone and social profiles when Átvéve, and opens photos plus the full services and introduction text when Ellenőrzött.

**Architecture:** `src/lib/entryPublicExtras.js` decides which blocks render. Postgres stores `hours_enabled`, `delivery_enabled`, and `social_links`. Public JSON keeps phone and social profiles only when the listing is claimed, keeps photos only when it is verified, and keeps hours only when the matching switch is on. The page clamps services to two lines and the introduction to four until the listing is Ellenőrzött. Claim, join, and suggestion *behavior* stay as they are today; this plan only moves the controls. Those behaviors are the later plans.

**Tech Stack:** Go `net/http` + Postgres, Svelte 5 runes in `EntryProfile.svelte`, existing `bejegyzes/[slug]/+page.svelte` stays `let` / `$:`, Node `node --test`, Go `go test`.

**Spec:** `docs/superpowers/specs/2026-09-27-listing-profile-design.md`

## Global Constraints

- Hungarian copy: **Gazdátlan**, **Átvéve**, **Nem ellenőrzött**, **Ellenőrzött**, **Szerkesztés**, **Javaslat módosításra**, **Sajátnak jelölöm**, **Tagság kérése**, **Útvonal**, **Nyitvatartás**, **Kiszállítás**.
- Grey marks: Gazdátlan and Nem ellenőrzött. Blue marks: Átvéve and Ellenőrzött (`var(--szekely-blue)`).
- A gazdátlan listing stays on the short view even when `verified` is true.
- Services clamp: 2 lines. Introduction clamp: 4 lines. Ellenőrzött removes both clamps.
- Phone and social profiles render only when `claimed` is true and the value is stored.
- Photos render only when `claimed` and `verified` are both true and at least one photo is stored.
- Full services and introduction text render only when `claimed` and `verified` are both true. A gazdátlan listing stays clamped even when `verified` is true.
- Hours render when `hours_enabled` is true, including an empty week. Delivery renders when `delivery_enabled` is true and `offersDelivery(entry)` is true.
- `offersDelivery`: category **Vendéglő** offers delivery. A listing whose name is **Lámsza.com** does not. Every other category does not.
- Website and languages render when stored, on every listing.
- Town and street render when stored. Map pin and **Útvonal** render when coordinates are stored. Coordinates are not printed.
- The heart is at the top right of the header and only a logged-in user sees it.
- The action row under the profile (`page-actions`) is removed. **Sajátnak jelölöm**, **Tagság kérése**, and **Javaslat módosításra** sit in the sidebar. The active owner sees **Szerkesztés** in the suggestion spot and it opens the existing `ListingFormDialog`.
- Ratings rules from `docs/superpowers/specs/2026-09-23-listing-public-claimed-design.md` stay: ratings show only when claimed and `ratings_enabled`.
- Do not change claim or join POST behavior in this plan. Do not add suggestion storage in this plan.
- New and edited `.svelte` files: run Svelte MCP `svelte-autofixer` until clean. Do not migrate `src/routes/(public)/bejegyzes/[slug]/+page.svelte` to runes.
- Commit only the files named in the task. Never `git add -A`. Leave unrelated `EntryRelatedLinks.svelte` and `EntryHistoryStrip.svelte` work untouched. Do not merge to `main`.

---

### Task 1: Public block rules

**Files:**
- Modify: `src/lib/entryPublicExtras.js`
- Test: `tests/entryPublicExtras.test.js`

**Interfaces:**
- Consumes: entry fields `claimed`, `verified`, `photos`, `hours`, `delivery_hours`, `hours_enabled`, `delivery_enabled`, `category`, `name`, `phone`, `social_links`, `url`, `languages`
- Produces:
  - `offersDelivery(entry) -> boolean`
  - `showListingPhotos(entry) -> boolean`
  - `showListingHours(entry) -> boolean`
  - `showListingDeliveryHours(entry) -> boolean`
  - `showListingRatings(entry) -> boolean`
  - `showListingTodayHours(entry) -> boolean`
  - `showListingPhone(entry) -> boolean`
  - `showListingSocial(entry) -> boolean`
  - `showListingWebsite(entry) -> boolean`
  - `showListingLanguages(entry) -> boolean`
  - `listingTextExpanded(entry) -> boolean`

- [ ] **Step 1: Write the failing tests**

Add these tests to `tests/entryPublicExtras.test.js` and update the imports. Replace the old hours and photos tests that require `claimed` for hours or photos.

```js
test('showListingPhotos is true only when claimed, verified, and a photo is stored', () => {
	assert.equal(showListingPhotos({ verified: false, claimed: true, photos: [{ id: 1 }] }), false);
	assert.equal(showListingPhotos({ verified: true, claimed: false, photos: [{ id: 1 }] }), false);
	assert.equal(showListingPhotos({ verified: true, claimed: true, photos: [{ id: 1 }] }), true);
	assert.equal(showListingPhotos({ verified: true, claimed: true, photos: [] }), false);
});

test('showListingHours follows hours_enabled even when gazdátlan or the week is empty', () => {
	assert.equal(showListingHours({ claimed: false, hours_enabled: true, hours: {} }), true);
	assert.equal(showListingHours({ claimed: true, hours_enabled: false, hours: { mon: { open: '09:00', close: '17:00' } } }), false);
});

test('showListingDeliveryHours requires the switch and a delivery category', () => {
	assert.equal(showListingDeliveryHours({
		delivery_enabled: true,
		category: 'Vendéglő',
		name: 'Példa'
	}), true);
	assert.equal(showListingDeliveryHours({
		delivery_enabled: true,
		category: 'Vendéglő',
		name: 'Lámsza.com'
	}), false);
	assert.equal(showListingDeliveryHours({
		delivery_enabled: true,
		category: 'Bolt',
		name: 'Példa'
	}), false);
});

test('phone and social show only when claimed and stored', () => {
	assert.equal(showListingPhone({ claimed: false, phone: '0700000000' }), false);
	assert.equal(showListingPhone({ claimed: true, phone: '0700000000' }), true);
	assert.equal(showListingPhone({ claimed: true, phone: '  ' }), false);
	assert.equal(showListingSocial({ claimed: true, social_links: [{ label: 'Facebook', url: 'https://facebook.com/a' }] }), true);
	assert.equal(showListingSocial({ claimed: false, social_links: [{ label: 'Facebook', url: 'https://facebook.com/a' }] }), false);
});

test('listingTextExpanded is true only when claimed and verified', () => {
	assert.equal(listingTextExpanded({ verified: false, claimed: true }), false);
	assert.equal(listingTextExpanded({ verified: true, claimed: false }), false);
	assert.equal(listingTextExpanded({ verified: true, claimed: true }), true);
});
```

Keep the existing ratings tests. Delete the old photos and hours tests that expect `claimed` to gate those blocks.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `node --test tests/entryPublicExtras.test.js`

Expected: FAIL. `showListingPhotos` still requires `claimed`. `hours_enabled` and the new exports are missing.

- [ ] **Step 3: Write the rules**

Replace `src/lib/entryPublicExtras.js` with:

```js
import { hoursConfigured } from "./entryHours.js";

export function offersDelivery(entry) {
	if (String(entry?.name ?? "").trim() === "Lámsza.com") return false;
	return String(entry?.category ?? "").trim() === "Vendéglő";
}

export function showListingPhotos(entry) {
	return Boolean(entry?.claimed) && Boolean(entry?.verified) && Array.isArray(entry?.photos) && entry.photos.length > 0;
}

export function showListingHours(entry) {
	return Boolean(entry?.hours_enabled);
}

export function showListingDeliveryHours(entry) {
	return Boolean(entry?.delivery_enabled) && offersDelivery(entry);
}

export function showListingRatings(entry) {
	return Boolean(entry?.claimed) && Boolean(entry?.ratings_enabled);
}

export function showListingTodayHours(entry) {
	return showListingHours(entry);
}

export function showListingPhone(entry) {
	return Boolean(entry?.claimed) && String(entry?.phone ?? "").trim() !== "";
}

export function showListingSocial(entry) {
	return Boolean(entry?.claimed) && Array.isArray(entry?.social_links) && entry.social_links.some((row) => String(row?.url ?? "").trim() !== "");
}

export function showListingWebsite(entry) {
	return String(entry?.url ?? "").trim() !== "";
}

export function showListingLanguages(entry) {
	return Array.isArray(entry?.languages) && entry.languages.some((row) => String(row ?? "").trim() !== "");
}

export function listingTextExpanded(entry) {
	return Boolean(entry?.claimed) && Boolean(entry?.verified);
}

export { hoursConfigured };
```

`hoursConfigured` stays exported because `ListingFormDialog.svelte` and older callers may import it from `entryHours.js` directly. Do not switch hours visibility back to `hoursConfigured`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `node --test tests/entryPublicExtras.test.js`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add src/lib/entryPublicExtras.js tests/entryPublicExtras.test.js
git commit -m "Gate listing blocks by verification and the hours switches."
```

---

### Task 2: Store the switches and social links

**Files:**
- Create: `backend/migrations/entry_profile_view.sql`
- Create: `backend/internal/handlers/migrate_entry_profile_view.go`
- Modify: `backend/internal/models/models.go` (`Entry` struct)
- Modify: `backend/main.go` (call migrate next to `MigrateEntryReviews()`)
- Modify: `backend/handlers_test.go` (same migrate call in test init)

**Interfaces:**
- Produces columns on `entries`:
  - `hours_enabled BOOLEAN NOT NULL DEFAULT false`
  - `delivery_enabled BOOLEAN NOT NULL DEFAULT false`
  - `social_links JSONB NOT NULL DEFAULT '[]'`
- Produces `Entry` fields: `HoursEnabled bool \`json:"hours_enabled"\``, `DeliveryEnabled bool \`json:"delivery_enabled"\``, `SocialLinks json.RawMessage \`json:"social_links"\``

- [ ] **Step 1: Write the migration**

`backend/migrations/entry_profile_view.sql`:

```sql
ALTER TABLE entries ADD COLUMN IF NOT EXISTS hours_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE entries ADD COLUMN IF NOT EXISTS delivery_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE entries ADD COLUMN IF NOT EXISTS social_links JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE entries
SET hours_enabled = true
WHERE hours_enabled = false
  AND hours IS NOT NULL
  AND hours::text NOT IN ('{}', 'null');
```

The UPDATE turns the switch on for listings that already have hours stored, so existing claimed hours do not disappear. An empty `{}` stays off.

- [ ] **Step 2: Apply it on startup**

Follow `MigrateEntryReviews` in `backend/internal/handlers/migrate_entry_reviews.go`. Read the SQL file and execute it. Call `MigrateEntryProfileView()` from `backend/main.go` and from the test `init` in `backend/handlers_test.go`, next to the reviews migrate.

- [ ] **Step 3: Add the struct fields**

On `models.Entry`, next to `Hours`:

```go
HoursEnabled    bool            `json:"hours_enabled"`
DeliveryEnabled bool            `json:"delivery_enabled"`
SocialLinks     json.RawMessage `json:"social_links"`
```

- [ ] **Step 4: Run the backend tests that already boot the database**

Run: `cd backend && go test . -count=1 -timeout 180s -run TestUnclaimedPublicEntryHidesExtras`

Expected: PASS or the existing unclaimed-hours assertion still passes because the new default is false and that test does not set `hours_enabled`. If the test fails only because hours remain visible, stop and fix the SELECT in Task 3 before continuing. Do not weaken the test to hide a missing column.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/entry_profile_view.sql backend/internal/handlers/migrate_entry_profile_view.go backend/internal/models/models.go backend/main.go backend/handlers_test.go
git commit -m "Store listing hours switches and social profile links."
```

---

### Task 3: Public JSON matches the short view

**Files:**
- Modify: `backend/internal/handlers/public.go` (entry SELECT lists)
- Modify: `backend/internal/handlers/entry_public_extras.go`
- Modify: `backend/internal/account/listings.go` (owner/member PATCH and GET detail)
- Test: `backend/entry_profile_view_test.go`

**Interfaces:**
- Consumes: `hours_enabled`, `delivery_enabled`, `social_links`, `verified`, `claimed`
- Produces public `GET /api/entry` JSON:
  - `phone` is `""` when `claimed` is false
  - `social_links` is `[]` when `claimed` is false
  - `photos` is `[]` unless `claimed` and `verified` are both true
  - `hours` is `{}` when `hours_enabled` is false
  - `delivery_hours` is `{}` when `delivery_enabled` is false
  - `notes` and `tags` stay complete; the page clamps them
- Owner/member GET of a listing returns the stored phone, social links, photos, and hours even when the public page hides them.

- [ ] **Step 1: Write the failing test**

Create `backend/entry_profile_view_test.go` in `package main`. Use `createEntry`, `doAnonRequest`, and `db.DB.Exec` the same way as `backend/entry_claimed_extras_test.go`.

```go
func TestGazdátlanVerifiedEntryKeepsShortPublicFields(t *testing.T) {
	id, slug := createEntry("Short View A", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)
	_, err := db.DB.Exec(`
		UPDATE entries
		SET verified = true,
		    phone = '0700111222',
		    social_links = '[{"label":"Facebook","url":"https://facebook.com/a"}]'::jsonb,
		    photos = '["https://example.com/a.jpg"]'::jsonb,
		    hours_enabled = false,
		    hours = '{"mon":{"open":"09:00","close":"17:00","closed":false}}'::jsonb
		WHERE id = $1
	`, id)
	if err != nil {
		t.Fatal(err)
	}
	rr := doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != 200 {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	var got map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["phone"] != "" {
		t.Fatalf("phone visible on gazdátlan listing: %v", got["phone"])
	}
	photos, _ := got["photos"].([]interface{})
	if len(photos) != 0 {
		t.Fatalf("photos visible before Ellenőrzött: %v", got["photos"])
	}
	if got["verified"] != true || got["claimed"] == true {
		t.Fatalf("marks: verified %v claimed %v", got["verified"], got["claimed"])
	}
}
```

Add a second test that sets `claimed` by inserting an active owner, sets `verified = false`, `hours_enabled = true`, and an empty hours object, and asserts `hours_enabled` is true and `phone` is the stored number.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 180s -run 'TestGazd'`

Expected: FAIL because phone and photos are still present.

- [ ] **Step 3: Select and strip the fields**

Add `e.hours_enabled`, `e.delivery_enabled`, `e.social_links`, and settlement `locations.coordinates` as `location_coordinates` to every public entry SELECT in `public.go` that already selects `verified` and `hours`. Scan them into `models.Entry`. Add `LocationCoordinates string \`json:"location_coordinates"\`` on `models.Entry`.

In `ApplyPublicEntryExtrasMode`, replace the unclaimed strip with:

```go
if !e.Claimed || !e.Verified {
	e.Photos = json.RawMessage("[]")
}
if !e.Claimed {
	e.Phone = ""
	e.SocialLinks = json.RawMessage("[]")
	e.RatingsEnabled = false
}
if !e.HoursEnabled {
	e.Hours = json.RawMessage("{}")
}
if !e.DeliveryEnabled {
	e.DeliveryHours = json.RawMessage("{}")
}
if !e.Claimed {
	return
}
```

Leave the ratings query on the claimed path as it is. Do not clear `notes` or `tags`.

Update the owner/member listing SELECT and PATCH in `listings.go` so GET returns the three new columns and PATCH writes `hours_enabled`, `delivery_enabled`, and `social_links`. A member PATCH that changes those columns is allowed. A non-member PATCH stays 403.

- [ ] **Step 4: Run the tests**

Run: `cd backend && go test . -count=1 -timeout 180s -run 'TestGazd|TestUnclaimedPublicEntryHidesExtras|TestClaimed'`

Expected: PASS. Update an old test only when it still asserts that unclaimed listings hide hours that now have `hours_enabled = false`. Do not assert that a switched-on gazdátlan listing hides hours.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handlers/public.go backend/internal/handlers/entry_public_extras.go backend/internal/account/listings.go backend/entry_profile_view_test.go
git commit -m "Hide unclaimed phone and unverified photos on the public entry."
```

---

### Task 4: Render the short profile

**Files:**
- Modify: `src/lib/components/EntryProfile.svelte`
- Modify: `src/routes/(public)/bejegyzes/[slug]/+page.svelte`

**Interfaces:**
- Consumes: the functions from Task 1 and the JSON from Task 3
- Produces: header marks, clamped text, sidebar rows, heart in the header, no `.page-actions` row

- [ ] **Step 1: Marks**

In `EntryProfile.svelte`, the verified badge uses `entry-profile__claim--verified` when `verified` is true and `entry-profile__claim--unverified` when it is false. Set verified color to `var(--szekely-blue)`. Set unverified color to `var(--text-faint)`. Leave `ClaimMark` as it is: grey when gazdátlan, blue when Átvéve.

- [ ] **Step 2: Clamp the text**

On the services `ul`, add class `entry-profile__services--clamp` when `listingTextExpanded(entry)` is false. On the introduction `p`, add `entry-profile__about--clamp` in that same case.

```css
.entry-profile__services--clamp {
	display: -webkit-box;
	-webkit-line-clamp: 2;
	-webkit-box-orient: vertical;
	overflow: hidden;
}
.entry-profile__about--clamp {
	display: -webkit-box;
	-webkit-line-clamp: 4;
	-webkit-box-orient: vertical;
	overflow: hidden;
}
```

Empty services and an empty introduction keep the existing empty placeholder.

- [ ] **Step 3: Hours, delivery, phone, social, website, languages**

Render the hours block when `showListingHours(entry)` is true, even if every day is empty. Render delivery when `showListingDeliveryHours(entry)` is true. Render the phone row when `showListingPhone(entry)` is true. Render one sidebar row per social link when `showListingSocial(entry)` is true: the label is the link text and `href` is the URL, with `target="_blank"` and `rel="noopener noreferrer"`. Website and languages stay, and they are omitted when `showListingWebsite` or `showListingLanguages` is false.

The place block keeps town and street. Public entry JSON adds `location_coordinates` from `locations.coordinates` (the settlement column, a `"lat,lng"` string). When that string parses as two numbers, show a map-pin link and point **Útvonal** at `https://www.google.com/maps/search/?api=1&query=LAT,LNG`. When it does not parse, keep today's **Útvonal** built from street and town, and omit the pin. Do not print the coordinate string.

- [ ] **Step 4: Move the heart and the sidebar actions**

In `+page.svelte`, delete the `.page-actions` block that contains `FavoriteButton`, **Sajátnak jelölöm**, and **Tagság kérése**.

Pass these props into `EntryProfile`: `loggedIn`, `isFavorite`, `onFavorite`, `showClaim`, `showJoin`, `claimBusy`, `onClaim`, `onJoin`, `isOwner`, `onEdit`.

`EntryProfile` places `FavoriteButton` at the top right of `.entry-profile__hero` only when `loggedIn` is true. Keep the current `showClaimButton` and `showJoinButton` conditions and render those buttons in the sidebar. The suggestion spot shows **Szerkesztés** when the viewer is the active owner. It shows **Javaslat módosításra** when the viewer is logged in and is not the active owner or an active member. Waiting labels for a pending claim or suggestion are added by the later plans. **Szerkesztés** calls `onEdit`.

`onEdit` in `+page.svelte` opens the existing `ListingFormDialog` in edit mode for `entry.id`, the same dialog `fiok/+page.svelte` uses. Do not create a second editor.

The suggestion button stays a button with no submit handler in this plan.

- [ ] **Step 5: Check the page**

Run: `node --test tests/entryPublicExtras.test.js`

Expected: PASS

With the dev server running, open a gazdátlan listing and an Átvéve listing. Confirm the heart is in the header for a logged-in user only, the old action row is gone, and **Szerkesztés** shows for the owner.

Run Svelte MCP `svelte-autofixer` on `EntryProfile.svelte` until it reports no issues.

- [ ] **Step 6: Commit**

```bash
git add src/lib/components/EntryProfile.svelte src/routes/(public)/bejegyzes/[slug]/+page.svelte
git commit -m "Show the short listing profile and move its actions into the sidebar."
```

---

### Task 5: Hours switches in the editor

**Files:**
- Modify: `src/lib/components/ListingFormDialog.svelte`
- Modify: the admin entry editor in `src/routes/admin/+page.svelte` where hours are saved

**Interfaces:**
- Consumes: `hours_enabled` and `delivery_enabled` on the listing payload
- Produces: two switches saved with the listing. Filling a day does not turn a switch on. Turning a switch on with an empty week stays on.

- [ ] **Step 1: Stop deriving the switch from filled days**

In `ListingFormDialog.svelte`, `hoursEnabled` is initialized from `listing.hours_enabled`, not from `hoursConfigured(listing.hours)`. Add `deliveryEnabled` from `listing.delivery_enabled`. The hours grid stays editable while the switch is on, including when every day is empty. The delivery grid is shown only when `offersDelivery(listing)` is true.

Include both booleans on the PATCH body next to `hours` and `delivery_hours`.

- [ ] **Step 2: Admin can switch any listing**

On the admin entry form, add the same two switches and send them on admin save. An admin can turn hours on for a gazdátlan listing. The delivery switch is saved only when `offersDelivery` is true; otherwise send `delivery_enabled: false`.

- [ ] **Step 3: Check a save**

Edit an Átvéve listing, turn **Nyitvatartás** on, leave the week empty, and save. Reload the public page. The hours block is visible. Turn the switch off and save. The block is gone.

Run Svelte MCP `svelte-autofixer` on `ListingFormDialog.svelte` until it reports no issues.

- [ ] **Step 4: Commit**

```bash
git add src/lib/components/ListingFormDialog.svelte src/routes/admin/+page.svelte
git commit -m "Let owners and admins switch listing hours independently of filled days."
```
