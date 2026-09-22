# Homepage Gyorslinkek display width and widget layout

Date: 2026-08-20

## Problem

The homepage widgets row (`Gyorslinkek`, date/time, weather) is a wrapping flex row with `justify-content: space-between`. Gyorslinkek grows with every card (`flex-wrap: nowrap`, no width cap, `#gyorslinkek { min-width: 33% }`). When the visitor has many links, the strip takes most of the row, date/time and weather wrap, and `space-between` pins them to opposite edges.

`LINKS_BEFORE_ARROWS` (7) does not affect layout. Scroll arrows are unused and are not wanted.

## Goal

Let each visitor choose how many quicklink **slots wide** the Gyorslinkek widget is. Default compact layout is three widgets in one row, packed next to each other. A higher choice makes Gyorslinkek take the top row so date/time and weather sit together on the row below. Every card stays visible; extras wrap inside the widget. Admin (promoted) links are always in the list. While promoted links load, the compact row already shows 7 slots (including `+`).

## Non-goals

- Admin UI or site-wide setting for display count
- Hiding, paging, or scrolling extra cards
- Scroll arrows or drag-to-scroll
- Changing quicklink CRUD, cache versioning, or the `quick_links` API
- Letting visitors edit or delete admin cards

## Behavior

### Slot count `N`

- `N` is the widget **width in cards**, including the `+` (Új) card. It is not a hide limit.
- Default / min: **7**. Max: **14**.
- Stored in `localStorage` key `quick_links_display_count` (integer). Missing, non-numeric, or out-of-range values clamp to `[7, 14]` (missing → 7).
- Visitor control: `−` / `+` stepper on the right of the Gyorslinkek `widget-header`. `−` disabled at 7, `+` disabled at 14. Buttons have accessible names (e.g. fewer / more slots). Current `N` may be shown between the buttons or only via `aria`.
- Below 992px widgets already stack in a column; `N` still sets wrap width (`max-width: 100%`).

### Layout

- All personal and admin cards always render. Order: `+`, personal (`user_quick_links`), then admin/promoted. Extras wrap to additional rows inside the widget.
- **`N = 7` (compact):** Gyorslinkek width ≈ 7 card slots. Date/time and weather stay on the same row, immediately after Gyorslinkek (`flex-start`, not `space-between`).
- **`N > 7` (expanded):** Gyorslinkek uses `flex-basis: 100%` so it occupies the whole first row (date/time and weather wrap below, still adjacent and left-aligned). The card strip itself stays **N slots wide** and left-aligned; `N` still controls wrap. Do not stretch cards to the viewport.

### Loading skeletons

- While `promotedLoading` is true:
  - Always show the real `+` card.
  - Show already-loaded personal links.
  - Fill remaining slots up to **7** with `.link-card--skeleton` placeholders.
  - Skeleton count = `max(0, 7 - 1 - userLinks.length)`.
- When promoted fetch finishes (success or error), remove skeletons.
- On success, show every promoted card after personal ones.
- On error, keep today’s behavior: hide promoted cards (`promotedError`); `+` and personal links remain; `N` still applies.

## Architecture

Three units:

1. **`src/lib/quickLinksDisplay.js`** — `DEFAULT_QUICKLINK_SLOTS = 7`, `MIN = 7`, `MAX = 14`, `STORAGE_KEY`, `clampSlotCount(value)`, `readSlotCount()`, `writeSlotCount(n)`. No UI, no fetch.
2. **Homepage `+page.svelte`** — owns stepper, `--quicklink-slots`, expanded class, skeleton `{#each}`, card lists. Drops `LINKS_BEFORE_ARROWS`, `showArrows`, arrow buttons, `dragScroll` / `scrollLinks` / scroll state used only for arrows.
3. **`src/styles/global.css`** — slot width, wrap, compact vs expanded row. No JS.

Data flow:

```
localStorage (display count + personal links)
  → clampSlotCount → N
  → style --quicklink-slots and expanded class on the three-widget row
  → CSS width / wrap / flex-basis

GET /api/config/public + GET /api/admin/quick_links (unchanged)
  → promotedLinks / promotedError
```

## CSS (normative)

Card slot for width math: **60px** (title `min-width`) **+ 0.5rem** gap, matching `#gyorslinkek .quick-links { gap: 0.5rem }`.

```css
#gyorslinkek {
  /* remove min-width: 33% */
  width: calc(var(--quicklink-slots, 7) * (60px + 0.5rem));
  max-width: 100%;
  flex: 0 1 auto;
}

#gyorslinkek .quick-links {
  display: flex;
  flex-wrap: wrap; /* was nowrap */
  overflow-x: visible; /* was auto */
  /* no grab cursor / hidden scrollbar needed for layout */
}

.widgets-box--three-col {
  justify-content: flex-start; /* was space-between */
}

.widgets-box--three-col.widgets-box--quicklinks-expanded #gyorslinkek {
  flex-basis: 100%;
  /* keep width: calc(N * slot); do not set width: 100% */
}
```

Existing `.link-card--skeleton` styles are reused. Stepper styling follows existing `btn btn-xs` / widget header patterns.

## Errors

| Case | Result |
| --- | --- |
| Invalid `quick_links_display_count` | Use 7 |
| Promoted fetch fails | Skeletons off; no admin cards; personal + `+` remain; `N` unchanged |
| `localStorage` quota / parse error on personal links | Same as today: `userLinks = []` |
| `FEATURE_QUICKLINKS` off | Out of scope (API missing); homepage error path already covers failed fetch |

## Tests

No new backend tests. Extend `tests/frontend-test-checklist.md`:

- Default: 7 slots, three widgets in one row, packed together.
- `N > 7`: Gyorslinkek full top row; date/time and weather adjacent below.
- `−` / `+` clamp at 7 and 14; value persists across reload.
- All personal and admin cards remain visible; extras wrap; no arrows.
- Loading: `+` plus skeletons totaling 7 slots when the visitor has no personal links.
- Admin cards cannot be edited or deleted from the homepage.

## Files

| File | Change |
| --- | --- |
| `src/lib/quickLinksDisplay.js` | Create |
| `src/routes/(public)/+page.svelte` | Stepper, `N`, skeletons, drop arrows/drag-scroll |
| `src/styles/global.css` | Width, wrap, `flex-start`, expanded modifier |
| `tests/frontend-test-checklist.md` | Checklist items above |

Admin `+page.svelte` is unchanged.
