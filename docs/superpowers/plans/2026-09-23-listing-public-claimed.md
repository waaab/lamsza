# Public listing vs claimed extras Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Unclaimed `/bejegyzes/[slug]` shows directory facts only; claimed listings may show photos, hours, and real ratings/reviews when those extras are filled or enabled.

**Architecture:** `entries.ratings_enabled` (default false) plus `entry_reviews` (one row per user per listing). Public GET strips photos/hours/ratings for unclaimed listings. `POST`/`DELETE /api/entry/reviews` upserts or deletes the current user’s review. Owners and members flip the switch and edit photos/hours on **Bejegyzéseim**. The public URL stays visitor-only.

**Tech Stack:** Go `net/http` + Postgres, Svelte 5 runes for new review UI, existing listing `+page.svelte` stays `let` / `$:`, Node `node --test`, Go `go test`.

**Spec:** `docs/superpowers/specs/2026-09-23-listing-public-claimed-design.md`

## Global Constraints

- Hungarian copy: section **Értékelések**, switch **Értékelések**, save error **A mentés nem sikerült**.
- Unclaimed public listing: directory facts only. No gallery, hours, hero “Nyitvatartás ma”, stars, recommend row, or demo slides — even if admin data exists.
- Claimed extras: photos if stored photos exist; hours if `hoursConfigured`; ratings block if `ratings_enabled` (even when count is 0).
- No demo / staging slides on the public listing page (claimed or not).
- Ratings: integer 1–5, optional plain text max 2000 runes, one row per `(entry_id, user_id)`, HTML stripped.
- Signed-in Google accounts (including owner/members) may post; signed-out may read; unsigned write is 401.
- Unclaimed, unpublished (to strangers), or ratings off: review write 403; unpublished stranger is 404.
- Turning ratings off keeps review rows. Deleting the listing cascades reviews.
- No comments in this version. Do not add comment tables or UI.
- Do not put email, Google sub, or session ids on public review JSON.
- New `.svelte` files: Svelte 5 runes. Do not migrate `src/routes/(public)/bejegyzes/[slug]/+page.svelte` or `src/routes/(public)/profil/+page.svelte` to runes. Run Svelte MCP `svelte-autofixer` on new/edited `.svelte` files until clean (ignore `$:` complaints on the two legacy pages).
- Commit only this task’s files on branch `user-profile`. Never `git add -A`. `handlers_test.go` and `main.go` are dirty with unrelated WIP: add the one-line route/migrate hooks there, but put new tests in `backend/entry_claimed_extras_test.go` (`package main`). Skip a commit step only if it would stage unrelated hunks — then report the files left unstaged.
- Do not merge to `main` in these tasks.

---

### Task 1: Schema and public JSON extras

**Files:**
- Create: `backend/migrations/entry_reviews.sql`
- Create: `backend/internal/handlers/migrate_entry_reviews.go`
- Create: `backend/internal/handlers/entry_public_extras.go`
- Create: `backend/entry_claimed_extras_test.go`
- Modify: `backend/internal/models/models.go` (`Entry`)
- Modify: `backend/internal/handlers/public.go` (call extras helper after scan)
- Modify: `backend/internal/handlers/entry_related.go` (claimed + extras on each row)
- Modify: `backend/main.go` (call `MigrateEntryReviews()` next to `MigrateEntryVerified()`)
- Modify: `backend/handlers_test.go` (same migrate call in `init`)

**Interfaces:**
- Consumes: `Entry.Claimed`, stored hours/photos
- Produces:
  - `entries.ratings_enabled BOOLEAN NOT NULL DEFAULT false`
  - `entry_reviews` table
  - `ApplyPublicEntryExtras(e *models.Entry, viewerUserID int)`
  - Public unclaimed: `photos` `[]`, hours/delivery empty objects, `ratings_enabled` false, no `rating` / `review_count` / `reviews` / `my_review`
  - Public claimed: real photos/hours; `ratings_enabled`; if on, `rating`, `review_count`, `reviews` (cap 50), `my_review` when `viewerUserID > 0`

- [ ] **Step 1: Write the failing tests**

Create `backend/entry_claimed_extras_test.go`:

```go
package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestUnclaimedPublicEntryHidesExtras(t *testing.T) {
	id, slug := createEntry("UnclaimedExtras A", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	doRequest(t, "PUT", "/api/admin/entries", map[string]interface{}{
		"id": id, "name": "UnclaimedExtras A", "slug": slug,
		"hours": map[string]any{"mon": map[string]any{"open": "09:00", "close": "17:00", "closed": false}},
	})

	rr := doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", rr.Code, rr.Body.String())
	}
	var got map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["claimed"] == true {
		t.Fatal("expected unclaimed")
	}
	if _, ok := got["rating"]; ok {
		t.Fatal("unclaimed must omit rating")
	}
	if _, ok := got["reviews"]; ok {
		t.Fatal("unclaimed must omit reviews")
	}
	if got["ratings_enabled"] != false {
		t.Fatalf("ratings_enabled: %v", got["ratings_enabled"])
	}
	photos, _ := got["photos"].([]interface{})
	if len(photos) != 0 {
		t.Fatalf("unclaimed photos must be empty, got %v", got["photos"])
	}
}
```

Use the same `createEntry` / `formatID` helpers already in `handlers_test.go`. If PUT hours is awkward, insert hours with SQL in the test via the admin payload fields that already exist (`hours` on admin update). If admin PUT cannot set hours, `db.DB.Exec` `UPDATE entries SET hours = $1 WHERE id = $2` is allowed in this test file.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 120s -run TestUnclaimedPublicEntryHidesExtras`

Expected: FAIL (`ratings_enabled` missing, photos/hours still present)

- [ ] **Step 3: Write the migration**

`backend/migrations/entry_reviews.sql`:

```sql
ALTER TABLE entries ADD COLUMN IF NOT EXISTS ratings_enabled BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS entry_reviews (
    id SERIAL PRIMARY KEY,
    entry_id INT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score SMALLINT NOT NULL CHECK (score >= 1 AND score <= 5),
    body TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (entry_id, user_id)
);
```

`MigrateEntryReviews()` runs those statements like `MigrateEntryVerified`. Call it from `main.go` and `handlers_test.go` `init` immediately after `MigrateEntryVerified()`.

- [ ] **Step 4: Models + ApplyPublicEntryExtras**

On `models.Entry` add:

```go
RatingsEnabled bool            `json:"ratings_enabled"`
Rating         *float64        `json:"rating,omitempty"`
ReviewCount    int             `json:"review_count,omitempty"`
Reviews        []PublicReview  `json:"reviews,omitempty"`
MyReview       *PublicReview   `json:"my_review,omitempty"`
```

```go
type PublicReview struct {
    ID          string  `json:"id"`
    AuthorName  string  `json:"author_name"`
    AuthorPhoto string  `json:"author_photo"`
    Score       int     `json:"score"`
    Text        string  `json:"text"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}
```

`ApplyPublicEntryExtras` in `entry_public_extras.go`:

- If `!e.Claimed`: set `Photos` to `[]`, `Hours` and `DeliveryHours` to `{}`, `RatingsEnabled = false`, leave rating/reviews unset.
- If claimed: keep stored photos/hours; scan `ratings_enabled`. If false, stop. If true, fill average (one decimal), count, up to 50 reviews newest first (join `users` for `COALESCE(NULLIF(name,''), 'Felhasználó')` and photo URL only — never email), and `MyReview` when `viewerUserID > 0`.

Call it at the end of `EntryDetailHandler` and each list row in `EntriesHandler`. Pass `0` for anonymous. For a signed-in public GET, read the session user id if `auth.UserFromRequest` already exists; if that helper is awkward on the public handler, pass `0` in this task and attach `my_review` in Task 3. Prefer attaching here if `UserFromRequest` is one call.

On related rows, add `Claimed bool`. After scan, if `!Claimed`, set `Photos` to `[]`. Related links do not show stars; stripping photos is enough.

- [ ] **Step 5: Run tests**

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestUnclaimedPublicEntryHidesExtras|TestPublicEntryVerifiedSeparateFromClaimed|TestClaimFreeListingBecomesOwner'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/migrations/entry_reviews.sql \
  backend/internal/handlers/migrate_entry_reviews.go \
  backend/internal/handlers/entry_public_extras.go \
  backend/entry_claimed_extras_test.go \
  backend/internal/models/models.go \
  backend/internal/handlers/public.go \
  backend/internal/handlers/entry_related.go
# stage only the migrate/route hunks in main.go and handlers_test.go
git commit -m "$(cat <<'EOF'
Hide unclaimed listing extras on the public entry API.

EOF
)"
```

---

### Task 2: Review write API

**Files:**
- Create: `backend/internal/handlers/entry_reviews.go`
- Modify: `backend/main.go` (`mux.HandleFunc("/api/entry/reviews", pub(handlers.HandleEntryReviews))`)
- Modify: `backend/handlers_test.go` (`testMux.HandleFunc("/api/entry/reviews", ...)`)
- Modify: `backend/entry_claimed_extras_test.go` (add write tests)
- Modify: `backend/internal/handlers/entry_public_extras.go` if `my_review` still missing

**Interfaces:**
- Consumes: session, `entries.ratings_enabled`, `entry_members` (claimed = active owner exists), `published`
- Produces: `HandleEntryReviews`
  - POST `{ "slug": string, "score": int, "text": string }` upserts; 200 body is `{ "my_review", "rating", "review_count", "reviews" }` same shapes as GET
  - DELETE `?slug=` deletes the current user’s row; same summary body
  - GET not required (detail GET already lists)

- [ ] **Step 1: Write the failing tests**

Append to `backend/entry_claimed_extras_test.go` (use `mustLogin` and claim like `TestClaimFreeListingBecomesOwner`):

1. Unsigned POST `/api/entry/reviews` → 401
2. Unclaimed published listing POST as signed-in user → 403
3. Claim listing, ratings still default false, POST → 403
4. Enable ratings with SQL `UPDATE entries SET ratings_enabled = true WHERE id = $1` (Task 3 adds PATCH; SQL is OK here), POST `{slug, score:5, text:"<b>hi</b>"}` → 200, stored text `"hi"`, `review_count` 1, `rating` 5
5. Same user POST score 3 → still one row, rating 3
6. Second user POST 5 → count 2, average 4.0
7. First user DELETE → count 1
8. First user DELETE second user’s review by any trick → 403 (DELETE has no target id; only own row). Score 0 or 6 → 400. Text of 2001 runes → 400. Unpublished listing (create via account, do not publish) stranger POST → 404

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestReview'`

Expected: FAIL (404 on `/api/entry/reviews`)

- [ ] **Step 3: Implement HandleEntryReviews**

```go
func HandleEntryReviews(w http.ResponseWriter, r *http.Request) {
    user := auth.UserFromRequest(r)
    if user == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    slug := r.URL.Query().Get("slug")
    // POST: also accept slug in JSON body
    // load entry: id, published, ratings_enabled, claimed (owner exists)
    // if !published && !canSeeUnpublished(user, entry): 404
    // if !claimed || !ratings_enabled: 403
    // POST: validate score 1–5; sanitizeReviewText; utf8.RuneCountInString <= 2000
    // INSERT ... ON CONFLICT (entry_id, user_id) DO UPDATE
    // DELETE: DELETE FROM entry_reviews WHERE entry_id=$1 AND user_id=$2
    // write JSON summary via ApplyPublicEntryExtras
}
```

`sanitizeReviewText`: trim, strip `<[^>]*>`, reject or strip `javascript:` substrings, return plain text.

Unpublished + not owner/member → 404 (do not reveal it exists). Unpublished + owner/member may write only if ratings on (profile still uses this API).

- [ ] **Step 4: Run tests**

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestUnclaimedPublicEntryHidesExtras|TestReview'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git commit -m "$(cat <<'EOF'
Add signed-in listing review upsert and delete.

EOF
)"
```

---

### Task 3: Bejegyzéseim ratings switch

**Files:**
- Modify: `backend/internal/account/listings.go` (`updateListingBody`, `listingDetailResponse`, GET detail SELECT, PATCH UPDATE)
- Modify: `src/routes/(public)/profil/+page.svelte` (`emptyListingForm`, `listingRequestBody`, load detail, dialog checkbox)
- Modify: `backend/entry_claimed_extras_test.go` (member PATCH)

**Interfaces:**
- Consumes: existing PATCH `/api/account/listings?id=`
- Produces: `ratings_enabled` boolean on GET detail and PATCH body. Members and owners who can already PATCH hours/photos can flip it. Strangers 403.

- [ ] **Step 1: Write the failing test**

Claim as user A. PATCH `/api/account/listings?id=` with the same fields the existing member PATCH test sends, plus `"ratings_enabled": true`. GET `/api/entry?slug=` → `ratings_enabled: true`. User B (not a member) PATCH → 403 and the flag stays true.

Follow the payload shape of `TestMemberCannotDeleteListing` / listing PATCH tests already in the repo (copy required name/location_id/category_id/type_id).

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 120s -run TestMemberPatchRatingsEnabled`

Expected: FAIL

- [ ] **Step 3: Backend PATCH/GET**

Add `RatingsEnabled bool \`json:"ratings_enabled"\`` to `updateListingBody` and `listingDetailResponse`. SELECT/UPDATE `e.ratings_enabled`. Create listing leaves the column default false.

- [ ] **Step 4: Profile dialog**

`emptyListingForm` adds `ratings_enabled: false`. When loading detail, set it from the GET. `listingRequestBody` includes `ratings_enabled: Boolean(form.ratings_enabled)`.

In the dialog, after photos, add:

```svelte
<label class="profile-listing-flag">
    <input type="checkbox" bind:checked={listingForm.ratings_enabled} />
    Értékelések
</label>
```

Create form includes the same checkbox (default off). Save failures already show **A mentés nem sikerült**.

Run Svelte autofixer on `profil/+page.svelte`. Do not convert it to runes.

- [ ] **Step 5: Run tests**

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestMemberPatchRatingsEnabled|TestReview'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git commit -m "$(cat <<'EOF'
Let listing owners toggle public ratings from the profile.

EOF
)"
```

---

### Task 4: Public listing UI and cards

**Files:**
- Create: `src/lib/components/EntryReviews.svelte`
- Modify: `src/lib/components/EntryProfile.svelte`
- Modify: `src/lib/components/EntryCard.svelte`
- Modify: `src/routes/(public)/bejegyzes/[slug]/+page.svelte` (pass session into profile only if needed; do not migrate to runes)
- Test: `tests/entryPublicExtras.test.js` (pure helpers if you extract `showListingHours` / `showListingPhotos` / `showListingRatings`; otherwise skip Node test and rely on Go + browser)

**Interfaces:**
- Consumes: public entry JSON from Task 1–3, `hoursConfigured` from `src/lib/entryHours.js`, `openLogin` from `$lib/openLogin.js`, `auth` store
- Produces: `EntryProfile` hides extras unless claimed; `EntryReviews` posts to `/api/entry/reviews`

- [ ] **Step 1: Write a small helper test (optional but preferred)**

`src/lib/entryPublicExtras.js`:

```js
import { hoursConfigured } from "./entryHours.js";

export function showListingPhotos(entry) {
    return Boolean(entry?.claimed) && Array.isArray(entry?.photos) && entry.photos.length > 0;
}
export function showListingHours(entry) {
    return Boolean(entry?.claimed) && hoursConfigured(entry?.hours);
}
export function showListingDeliveryHours(entry) {
    return Boolean(entry?.claimed) && hoursConfigured(entry?.delivery_hours);
}
export function showListingRatings(entry) {
    return Boolean(entry?.claimed) && Boolean(entry?.ratings_enabled);
}
export function showListingTodayHours(entry) {
    return showListingHours(entry);
}
```

`tests/entryPublicExtras.test.js` asserts unclaimed → all false; claimed empty photos → photos false; claimed ratings_enabled → ratings true.

Run: `node --test tests/entryPublicExtras.test.js`  
Expected first run: FAIL (module missing)

- [ ] **Step 2: Implement helpers + EntryProfile**

- Remove `demoGallerySlides` from the public profile `slides` derived. If there are no real slides, do not render the Fotók section (`showListingPhotos`).
- Hero: keep Ellenőrzött / Foglalt. Render “Nyitvatartás ma” + open status only when `showListingTodayHours(entry)`.
- Hero stars: only when `showListingRatings(entry)`.
- Delete `{#snippet recommendBlock}` and all Igen/Nem/Talán markup.
- Helyszín section: always show address (directory). Hours tables / delivery tables only when `showListingHours` / `showListingDeliveryHours`.
- Replace the placeholder Értékelések card with `<EntryReviews {entry} />` when `showListingRatings(entry)`.

- [ ] **Step 3: EntryReviews.svelte (runes)**

Props: `entry`. Read `$auth` for `loggedIn`. Local state: `score`, `text`, `error`, `busy`, plus lists from `entry` that refresh after save.

- Heading **Értékelések**, average, count.
- List `entry.reviews` (author_name, author_photo, score, text). Empty: **Még nincs értékelés.**
- If logged in: star buttons 1–5, textarea, submit **Mentés**, if `entry.my_review` also **Törlés**. POST `/api/entry/reviews` with `{ slug: entry.slug, score, text }`. DELETE `/api/entry/reviews?slug=`. On failure keep fields and set error **A mentés nem sikerült**. On success, apply JSON (`rating`, `review_count`, `reviews`, `my_review`) onto a local copy so the list updates without a full reload.
- If logged out: button **Belépés** that calls `openLogin()`.

- [ ] **Step 4: EntryCard**

Show the rating row and “Még nincs értékelés.” quote only when `showListingRatings(entry)`. Thumb image only from real photos (already via `gallerySlides`); do not use demo photos. When ratings are hidden, omit the 0,0 row entirely.

- [ ] **Step 5: Autofixer + tests**

Run Svelte autofixer on `EntryReviews.svelte`, `EntryProfile.svelte`, `EntryCard.svelte`.

Run: `node --test tests/entryPublicExtras.test.js`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git commit -m "$(cat <<'EOF'
Show listing photos, hours, and ratings only when the listing is claimed.

EOF
)"
```

---

### Task 5: Browser verification

**Files:** none (manual)

- [ ] **Step 1: Restart Go** so migrations and `/api/entry/reviews` exist. Vite HMR is enough for Svelte.

- [ ] **Step 2: Unclaimed listing (e.g. Turbomaster if still unclaimed).** Confirm no Fotók gallery, no hours tables, no Nyitvatartás ma, no stars, no recommend, no demo slides. Directory facts and claim button remain.

- [ ] **Step 3: Claim it (or open a listing you own).** On **Bejegyzéseim**, add hours and a photo URL, leave Értékelések off, save. Public page shows photo + hours, not ratings. POST review (curl with session) returns 403.

- [ ] **Step 4: Enable Értékelések.** Public page shows Értékelések with empty list. Signed-out sees Belépés. Signed-in save a 4-star review with text; average updates; save again changes the same review. Delete removes it.

- [ ] **Step 5: Index/history card for an unclaimed listing has no 0,0 stars.**

- [ ] **Step 6: Commit** only if Task 5 produced code fixes. Otherwise skip.

---

## Spec coverage

| Spec | Task |
|---|---|
| Unclaimed directory-only JSON | 1 |
| Claimed photos/hours/ratings_enabled | 1 |
| Reviews table, upsert, delete, sanitize, 401/403/404/400 | 2 |
| Bejegyzéseim Értékelések switch | 3 |
| Public profile hide extras, no demo, no recommend | 4 |
| EntryCard no fake 0,0 | 4 |
| Browser | 5 |
| Comments deferred | omitted; remains in the spec |

## Placeholder / type check

- `PublicReview` and `ApplyPublicEntryExtras` names are stable across tasks.
- JSON field for review text is `text`; DB column is `body`.
- `showListingHours` uses existing `hoursConfigured`.
