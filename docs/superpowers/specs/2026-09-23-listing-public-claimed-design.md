# Public listing vs claimed extras

Date: 2026-09-23

Related: `docs/superpowers/specs/2026-09-22-user-profile-design.md`, `docs/superpowers/specs/2026-09-22-listing-related-blocks-design.md`

## Problem

`/bejegyzes/[slug]` renders a full business profile for every listing: photo gallery (including staging demo slides), hours, star ratings, and “Ajánlja ezt a vállalkozást?” even when nobody owns the listing and those fields are empty. Directory facts and claimed extras are mixed. Ratings are placeholders (always 0,0). Comments do not exist.

Owners already edit photos and hours on **Bejegyzéseim**. That must stay off the public URL.

## Goal

An unclaimed public listing is a directory card. A claimed listing may add photos, hours, and real ratings/reviews when those extras are filled or enabled. Owners and members control extras on **Bejegyzéseim**. Hungarian UI.

## Non-goals

- Comments of any kind (listing-wide or per-service). See **Later: comments**.
- Owner/member moderation of other people’s reviews
- Half-star scores, review replies, photos on reviews, sort/filter, helpful votes
- Recommend Igen/Nem/Talán
- Editors on the public listing URL
- Changing claim, join, publish, verified, or Bejegyzéseim membership rules
- Cookie-policy changes

## Public directory facts (always, if present)

Name, type, category, settlement, address, phone, website, languages, notes, tags (Szolgáltatások chips), **Ellenőrzött** / **Nem ellenőrzött**, **Foglalt** when there is an active owner, claim/join when the viewer may use them, favorite, suggest-edit, events widget, nearby, related, browsing history.

Address stays visible without hours. Do not show “Nyitvatartás ma” in the hero unless hours are shown (claimed and configured).

## Claimed extras (only if claimed)

Show a block only when it has work to do:

| Block | Show when |
|---|---|
| Photo gallery | At least one stored photo. No demo / staging slides on the public page. |
| Hours (and delivery hours) | `hoursConfigured` is true for that set. Location address can show without hours. |
| Ratings | `ratings_enabled` is true. Show even if the count is 0 (form or sign-in prompt + empty list). |

If the listing is not claimed, those three stay off the page even if the database still has photos, hours, or old reviews (admin-entered data is not public until someone owns the listing).

Index cards, search hits, related/history thumbs, and map popups follow the same rule: no fake 0,0 stars, no demo photos, no hours line for unclaimed listings. A claimed listing with ratings off omits stars on cards. A claimed listing with no photos omits the thumb photo (initials are fine).

## Ratings

- Listing column `ratings_enabled`, boolean, default **false**.
- Owner and members set it on **Bejegyzéseim** (same PATCH as hours/photos). Label: **Értékelések**.
- Off: public ratings block hidden; POST/PATCH/DELETE review returns 403; existing review rows are kept so turning it on restores them.
- A review: listing, user, integer score **1–5**, optional plain text, created/updated timestamps. At most one row per user per listing. Saving again updates. The author may delete their own row. Average and count ignore deleted rows. The author’s own review counts in the average.
- Display name and avatar come from the account at read time. Do not put email in the public JSON.
- Signed-out visitors may read reviews. To write, they sign in (existing Google dialog). Owner and members may review their own listing.
- Score is required to save. Empty text is allowed. Strip HTML; store plain text. Max length **2000** characters.
- “Ajánlja ezt a vállalkozást?” is removed.

## APIs

Public `GET /api/entry` (and list/search payloads used by cards):

- Always: directory facts, `claimed`, `verified`, `published` as today.
- If **not** claimed: `photos` is `[]`, `hours` and `delivery_hours` are empty objects, `ratings_enabled` is false, and `rating`, `review_count`, `reviews`, and `my_review` are omitted. The page keys off `claimed`, not leftover admin data.
- If claimed: include stored photos and hours; include `ratings_enabled`. If ratings are on, include `rating` (average, one decimal, `0` if count is 0), `review_count`, `reviews` (newest first, cap **50**), and `my_review` when the request has a session. If ratings are off, omit `reviews` and `my_review`.

`POST /api/entry/reviews` body `{ "slug" or "entry_id", "score", "text" }` — upsert the current user’s review.  
`DELETE /api/entry/reviews?slug=` — delete the current user’s review.

Member `PATCH` of a listing accepts `ratings_enabled`. Non-members keep 403.

## Failures and bugs to close

**Auth and rights**

- No session on write: **401**. Public page shows sign-in, not a dead form submit.
- Unclaimed, unpublished, or ratings off: write **403**. Do not leak unpublished listings through the reviews URL (404 if the viewer may not see the listing).
- Non-member PATCH `ratings_enabled`: **403**.
- A user cannot update or delete someone else’s review: **403**.

**Validation**

- Score missing, not an integer, or outside 1–5: **400**.
- Text over 2000 characters: **400**.
- Text is sanitized to plain text (no stored HTML, no `javascript:` URLs in text).
- Two tabs posting at once: unique `(entry_id, user_id)` so the second save updates, it does not create two rows.

**Data lifetime**

- Turning ratings off does not delete reviews.
- Deleting the listing deletes its reviews.
- Unclaim does not happen in this version (no ownership transfer). If `claimed` becomes false, public extras hide; rows remain.
- Unpublished listing: not on the public site; reviews API 404 for strangers; owner/members still edit extras on the profile.

**Display leaks (current bugs)**

- Unclaimed page must not render gallery, hours tables, hero “Nyitvatartás ma”, stars, or recommend buttons.
- Staging `demoGallerySlides` must not run on the public listing page (claimed or not).
- Entry cards and history/related thumbs must not show 0,0 stars when ratings are off or the listing is unclaimed.
- Public JSON must not send other users’ emails, Google ids, or session ids on reviews.
- Average is computed server-side from remaining rows after a delete; do not leave a stale cached 0,0 vs a non-zero count.

**Save UX**

- Failed save keeps the previous review and shows **A mentés nem sikerült**.
- After success, average, count, and the list refresh from the response (no full page reload required).
- Member flips ratings off while another tab submits: **403**; the public block disappears on next load.

## Later: comments

Out of this version. When we return:

- Comments are a **separate** system from ratings/reviews, with a separate enable switch (or switches).
- Intended shape: a listing-wide thread **and** optional threads on Szolgáltatások the owner allows (“different types of services”).
- Switch nesting was not chosen. Options were: (A) one master comments switch, listing thread on, plus ticks per service; (B) no master, each thread independent; (C) master off kills all threads, on still requires ticks for listing thread and each service.
- Signed-in posting (including owner/members) was already chosen for ratings; revisit for comments if needed.

## Tests

- Unclaimed `GET /api/entry`: no public photos/hours/ratings fields that the UI would show; page hides those blocks.
- Claimed, ratings off, hours configured: hours visible, ratings hidden; review POST 403.
- Claimed, ratings on: empty list allowed; signed-in upsert; second POST updates the same row; average/count match; DELETE own review updates average; DELETE someone else’s 403.
- Unsigned POST 401. Unpublished slug POST 404 for a stranger.
- Member PATCH `ratings_enabled` true/false. Stranger PATCH 403.
- Review text with tags is stored without HTML.
- Duplicate concurrent upsert does not create two rows.
- Index/card payload for an unclaimed listing has no displayable rating.

## Failures (product)

Profile extras save already uses **A mentés nem sikerült**. Reuse it for the ratings switch and for review save. A hidden block is not an error.
