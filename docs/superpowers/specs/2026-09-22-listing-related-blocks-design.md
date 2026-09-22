# Listing page: nearby, related, and browsing history

Date: 2026-09-22

## Problem

A public listing (`/bejegyzes/[slug]`) ends after the profile and events. There is no in-page path to other local businesses or same-category listings, and no record of listings this visitor already opened.

## Goal

On each listing page, after the profile and events:

1. **Közeli helyek** — other local businesses near this listing.
2. **Kapcsolódó bejegyzések** — other listings in the same category (same county, no overlap with nearby).
3. **Böngészési előzmények** — this browser’s recently opened listings (not other people’s views).

Hungarian UI. Hide any block whose list is empty.

## Non-goals

- Real “people also viewed” / co-visit tracking
- Server-side or account-tied history
- Cookie-policy or consent-banner changes
- “See all” links, Index filter deep-links, or map/radius search
- Per-listing GPS; nearby uses settlement, then county
- New carousel npm packages
- Admin UI for these blocks

## Behavior

### Nearby (`nearby`)

- Source: other `entries` with the same settlement (`location_id` / `location_slug`) as the current listing.
- Exclude the current listing.
- Cap **8**.
- If fewer than 8, pad from other listings in the same county (`county_slug`), still excluding the current listing and those already chosen.
- Settlement matches come first; county pads follow.
- If a row was filled from the county (not the same settlement), the UI may show the settlement name next to the link so “nearby” is not misleading.
- If the result is empty, omit the whole **Közeli helyek** block.

### Related (`related`)

- Same `category` as the current listing, same county.
- Exclude the current listing and every id already in `nearby`.
- Cap **8**.
- If empty, omit **Kapcsolódó bejegyzések**.

### Browsing history (client only)

- Storage: `localStorage` key `lamsza_entry_history` on this device only.
- Opening a public listing page prepends `{ slug, name, category, location, photo }` (photo = first stored photo URL, or empty).
- Unique by `slug` (move existing to front).
- Store at most **12**; show at most **8**.
- Never show the listing currently open.
- Corrupt or unreadable JSON: treat as empty, do not throw.
- If after skipping the current slug the list is empty, omit **Böngészési előzmények**.

## Architecture

Three units:

1. **`GET /api/entry/related?slug=`** (public) — returns `{ nearby, related }` arrays of compact listing objects. No history.
2. **`src/lib/entryHistory.js`** — read/write/normalize `localStorage`. No UI, no fetch.
3. **Listing page UI** — fetch related after the entry is known; record history; render the three blocks.

Compact listing object (nearby, related, and stored history):

```json
{
  "id": 150,
  "name": "…",
  "slug": "…",
  "category": "Vendéglő",
  "location": "Kézdivásárhely",
  "location_slug": "kezdivasarhely",
  "county_slug": "kovaszna",
  "photos": []
}
```

History records store a single `photo` string (first image URL or `""`) plus `slug` and `name` (required). Other fields are optional. The related API still returns `photos` as an array so a thumb can use the same helper as Index cards.

Data flow:

```
GET /api/entry?slug=…          → profile
GET /api/entry/related?slug=…  → nearby + related
localStorage                    → history (write on view, read for strip)
```

Unknown `slug` on related: **404**. Listing page only calls related after the entry payload is valid. Related fetch failure: hide nearby/related; profile unchanged.

## Layout

Placement: below `EntryProfile` and `EventsWidget`.

Desktop: two columns of text links.

- Left: **Közeli helyek** — listing name → `/bejegyzes/{slug}`. County-padded rows may include settlement name.
- Right: **Kapcsolódó bejegyzések** — listing name → `/bejegyzes/{slug}`.

Below **992px**: stack, Nearby first.

Then **Böngészési előzmények**: horizontal strip of compact Index-style cards (thumb, name, category). Native overflow scroll; prev/next buttons only if content overflows. No third-party carousel.

## Testing

- Go: nearby prefers same settlement then county pad; never includes self; related is same category + county and disjoint from nearby; 404 for unknown slug; empty arrays when nothing qualifies.
- JS: history prepend, dedupe-by-slug, cap 12, skip current on display, corrupt JSON → `[]`.
- Browser: Manifesto (or any listing with neighbors) shows columns and working links; visiting a second listing then returning shows history without the current page.

## Files (indicative)

- `backend/internal/handlers/public.go` (or a sibling handler) — related endpoint; register in `main.go` and `handlers_test.go`
- `src/lib/entryHistory.js` + `tests/entryHistory.test.js`
- `src/lib/components/` — small presentational blocks for the two link columns and the history strip
- `src/routes/(public)/bejegyzes/[slug]/+page.svelte` — fetch, record, compose
