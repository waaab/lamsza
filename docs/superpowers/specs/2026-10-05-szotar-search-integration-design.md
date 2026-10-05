# Szótár in main search

Date: 2026-10-05

## Problem

The main Lámsza search panel covers index content (settlements, services, websites, attractions, venues, events, seats, news) but not Székely Szótár headwords. A user searching for a dialect word sees "Nincs találat" even when Szótár has a match. There is also no way to narrow the live result set to dictionary hits only.

## Goal

1. Include matching Szótár words in the unified search response and show them in the main results panel.
2. Add a **Székely Szótár** filter pill after **Weboldalak** that shows only those dictionary hits (same filter pattern as Index / Szolgáltatások / Weboldalak).
3. Clicking a dictionary row opens the word on the Szótár app in a new tab.

## Non-goals

- Changing Szótár’s own search page, word detail page, or admin.
- Syncing or copying the dictionary into the Lámsza database.
- Browser-direct calls from Lámsza to Szótár (CORS stays Szótár-owned).
- A new card layout, accent color, or toolbar entry for Szótár.
- Suggest / autocomplete changes (`/api/suggest` stays as today).

## Approach

Lámsza’s `HandleUnifiedSearch` fetches Szótár server-side in parallel with the other sources. The browser keeps a single `/api/search` call. Word rows reuse `SearchResultCard` with `external: true`.

## API

`GET /api/search?q=…` gains:

```json
"words": [
  {
    "id": 123,
    "headword": "gagya",
    "definition_hu": "…",
    "speech_types": ["főnév"]
  }
]
```

- Upstream: `{SZOTAR_ORIGIN}/api/words?q={q}` (same filter semantics as Szótár’s list search).
- Cap: 15 hits (aligned with other unified-search limits).
- Empty or missing fields: always return `words` as an array (possibly empty).
- Failure: if `SZOTAR_ORIGIN` is unset, the request times out, or the upstream status is not 2xx, log and return `words: []`. Do not fail the whole unified response.

### Config

`SZOTAR_ORIGIN` (env), no trailing slash. Examples:

- local: `https://szotar.lamsza.test`
- production: `https://szotar.lamsza.com`

No default that hits production from a local backend without the env set; unset means skip the upstream call.

## UI

### Filter pill

In `SearchEngine`’s `result-filters` group, after **Weboldalak**:

| Filter | `resultFilter` | Visible groups |
|---|---|---|
| Lámsza Index | `index` | Existing browse sections + services + websites (unchanged rules) **plus** Székely Szótár when `words` is non-empty |
| Szolgáltatások | `services` | Services only (no words) |
| Weboldalak | `websites` | Websites only (no words) |
| Székely Szótár | `szotar` | Words only |

Label: **Székely Szótár** (accents). Scope line when this filter is active: `szótárban:`.

### Result section

Section title: **Székely Szótár**, with the existing `szotar` app icon from `AppIcon` when practical; otherwise the same neutral heading style as Hírek / Helyszínek.

Card mapping (`searchResultCardModel("word", row)`):

| Field | Source |
|---|---|
| Title | `headword` |
| Meta | `speech_types` joined with `, ` (omit if empty) |
| Description | `definition_hu` |
| Href | `szotarUrl('/szo/' + id)` via `networkOrigins.js` |
| External | `true` (`target="_blank"`, `rel="nofollow noopener"`) |
| Accent | `none` |
| Claimed | not used |

Skip a row with empty `headword` or missing/zero `id`.

### Counts

When the active filter shows words, include them in `hasResults` and `totalCount`. Empty dictionary + empty other visible groups still shows the existing no-results line.

### Section order (Index)

Keep current section order. Place **Székely Szótár** after **Hírek** (last content section before the external-search / filter footer).

## Tests

- Backend: when upstream returns words, unified JSON includes a capped `words` array; when upstream fails or origin is empty, `words` is `[]` and other groups still populate.
- Frontend unit: `searchResultCardModel("word", …)` builds title, meta, description, external href; skips bad rows.
- Filter: `resultFilter === "szotar"` shows only words; Index includes words when present; services/websites hide words.

No browser E2E required for this change.

## Out of scope follow-ups

- Suggest bar including dictionary headwords.
- Deep-linking the filter from a URL query param.
