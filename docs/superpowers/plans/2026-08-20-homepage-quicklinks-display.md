# Homepage Gyorslinkek Display Width Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let each homepage visitor set how many quicklink slots wide the Gyorslinkek widget is (default 7 including `+`), so date/time and weather stay beside it until they choose more than 7, while every card stays visible and wraps.

**Architecture:** A tiny `src/lib/quickLinksDisplay.js` helper owns min/max/default, clamp, and `localStorage`. The homepage sets `--quicklink-slots` and an expanded class. CSS caps widget width, wraps cards, packs the three widgets with `flex-start`, and uses `flex-basis: 100%` on Gyorslinkek only when `N > 7`. Promoted-link fetch/cache is unchanged.

**Tech Stack:** Svelte 5 (keep this page in legacy `let` / `$: ` / `on:click` — do not migrate the file to runes), SvelteKit, `localStorage`, Node built-in test runner (`node --test`), existing `global.css`.

## Global Constraints

- Default / min slot count: **7** (includes the `+` / Új card). Max: **14**.
- `localStorage` key: **`quick_links_display_count`**. Missing / non-numeric → 7; out of range → clamp to `[7, 14]`.
- `N` is widget **width in cards**, not a hide limit. Order is always `+`, then personal (`user_quick_links`), then admin/promoted.
- Compact (`N = 7`): three widgets one row, packed (`justify-content: flex-start`). Expanded (`N > 7`): Gyorslinkek `flex-basis: 100%` so date/time and weather wrap below, adjacent; card strip stays `N` slots wide, left-aligned — do not set `width: 100%` on `#gyorslinkek`.
- Slot width math: `calc(var(--quicklink-slots, 7) * (60px + 0.5rem))`.
- Skeletons while `promotedLoading`: count = `max(0, 7 - 1 - userLinks.length)`.
- No arrows, no drag-scroll. No admin UI or API changes. Visitors cannot edit/delete admin cards.
- Do not commit unless the user explicitly asks (user git rule). Skip every “Commit” step unless they have asked.

---

### Task 1: Slot-count helper

**Files:**
- Create: `src/lib/quickLinksDisplay.js`
- Create: `tests/quickLinksDisplay.test.js`

**Interfaces:**
- Consumes: `localStorage` (browser) or none (SSR / Node without store)
- Produces:
  - `export const DEFAULT_QUICKLINK_SLOTS = 7`
  - `export const MIN_QUICKLINK_SLOTS = 7`
  - `export const MAX_QUICKLINK_SLOTS = 14`
  - `export const QUICKLINK_SLOTS_STORAGE_KEY = "quick_links_display_count"`
  - `export function clampSlotCount(value): number`
  - `export function readSlotCount(): number`
  - `export function writeSlotCount(n): number` (clamps, persists, returns clamped value)

- [ ] **Step 1: Write the failing tests**

Create `tests/quickLinksDisplay.test.js`:

```js
import assert from "node:assert/strict";
import { afterEach, beforeEach, test } from "node:test";
import {
    DEFAULT_QUICKLINK_SLOTS,
    MAX_QUICKLINK_SLOTS,
    MIN_QUICKLINK_SLOTS,
    QUICKLINK_SLOTS_STORAGE_KEY,
    clampSlotCount,
    readSlotCount,
    writeSlotCount,
} from "../src/lib/quickLinksDisplay.js";

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

test("constants", () => {
    assert.equal(DEFAULT_QUICKLINK_SLOTS, 7);
    assert.equal(MIN_QUICKLINK_SLOTS, 7);
    assert.equal(MAX_QUICKLINK_SLOTS, 14);
    assert.equal(QUICKLINK_SLOTS_STORAGE_KEY, "quick_links_display_count");
});

test("clampSlotCount: missing and non-numeric → default 7", () => {
    assert.equal(clampSlotCount(null), 7);
    assert.equal(clampSlotCount(undefined), 7);
    assert.equal(clampSlotCount(""), 7);
    assert.equal(clampSlotCount("abc"), 7);
    assert.equal(clampSlotCount(NaN), 7);
});

test("clampSlotCount: out of range", () => {
    assert.equal(clampSlotCount(3), 7);
    assert.equal(clampSlotCount(20), 14);
    assert.equal(clampSlotCount("6"), 7);
    assert.equal(clampSlotCount("15"), 14);
});

test("clampSlotCount: in range", () => {
    assert.equal(clampSlotCount(7), 7);
    assert.equal(clampSlotCount(10), 10);
    assert.equal(clampSlotCount(14), 14);
    assert.equal(clampSlotCount("8"), 8);
});

test("readSlotCount uses storage and clamps", () => {
    assert.equal(readSlotCount(), 7);
    localStorage.setItem(QUICKLINK_SLOTS_STORAGE_KEY, "11");
    assert.equal(readSlotCount(), 11);
    localStorage.setItem(QUICKLINK_SLOTS_STORAGE_KEY, "2");
    assert.equal(readSlotCount(), 7);
});

test("writeSlotCount persists clamped value", () => {
    assert.equal(writeSlotCount(9), 9);
    assert.equal(localStorage.getItem(QUICKLINK_SLOTS_STORAGE_KEY), "9");
    assert.equal(writeSlotCount(99), 14);
    assert.equal(localStorage.getItem(QUICKLINK_SLOTS_STORAGE_KEY), "14");
});

test("readSlotCount without localStorage returns default", () => {
    delete globalThis.localStorage;
    assert.equal(readSlotCount(), 7);
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `node --test tests/quickLinksDisplay.test.js`

Expected: FAIL with `ERR_MODULE_NOT_FOUND` for `../src/lib/quickLinksDisplay.js`.

- [ ] **Step 3: Write the helper**

Create `src/lib/quickLinksDisplay.js`:

```js
export const DEFAULT_QUICKLINK_SLOTS = 7;
export const MIN_QUICKLINK_SLOTS = 7;
export const MAX_QUICKLINK_SLOTS = 14;
export const QUICKLINK_SLOTS_STORAGE_KEY = "quick_links_display_count";

/** @param {unknown} value */
export function clampSlotCount(value) {
    if (value == null || value === "") return DEFAULT_QUICKLINK_SLOTS;
    const n = typeof value === "number" ? value : Number.parseInt(String(value), 10);
    if (!Number.isFinite(n)) return DEFAULT_QUICKLINK_SLOTS;
    return Math.min(MAX_QUICKLINK_SLOTS, Math.max(MIN_QUICKLINK_SLOTS, n));
}

export function readSlotCount() {
    if (typeof localStorage === "undefined") return DEFAULT_QUICKLINK_SLOTS;
    try {
        return clampSlotCount(localStorage.getItem(QUICKLINK_SLOTS_STORAGE_KEY));
    } catch {
        return DEFAULT_QUICKLINK_SLOTS;
    }
}

/** @param {unknown} n */
export function writeSlotCount(n) {
    const v = clampSlotCount(n);
    if (typeof localStorage === "undefined") return v;
    try {
        localStorage.setItem(QUICKLINK_SLOTS_STORAGE_KEY, String(v));
    } catch {
        /* ignore quota / private-mode failures */
    }
    return v;
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `node --test tests/quickLinksDisplay.test.js`

Expected: all tests PASS.

- [ ] **Step 5: Commit** (skip unless the user asked to commit)

```bash
git add src/lib/quickLinksDisplay.js tests/quickLinksDisplay.test.js
git commit -m "$(cat <<'EOF'
Add homepage quicklink slot-count helper with clamp and localStorage.

EOF
)"
```

---

### Task 2: Widget row and Gyorslinkek CSS

**Files:**
- Modify: `src/styles/global.css` (`#gyorslinkek` ~942–1017, `.widgets-box--three-col` ~525–533, mobile `#gyorslinkek` ~2742–2744; add stepper rules near widget-header)

**Interfaces:**
- Consumes: `--quicklink-slots` set by the homepage; class `widgets-box--quicklinks-expanded` on `.widgets-box--three-col`
- Produces: compact 7-slot width; wrap instead of horizontal scroll; packed widgets; expanded first-row occupancy without stretching the card strip to 100% width

- [ ] **Step 1: Replace `#gyorslinkek` and strip CSS**

Replace the `#gyorslinkek` block that currently has `min-width: 33%` with:

```css
#gyorslinkek {
  display: flex;
  flex-direction: column;
  width: calc(var(--quicklink-slots, 7) * (60px + 0.5rem));
  max-width: 100%;
  flex: 0 1 auto;
}
```

Replace `#gyorslinkek .quick-links` through `#gyorslinkek .quick-links.active` (the nowrap / overflow-x / grab / scrollbar-hidden / `.active` grabbing rules) with:

```css
#gyorslinkek .quick-links {
  display: flex;
  flex-wrap: wrap;
  overflow-x: visible;
  gap: 0.5rem;
  padding: 0.25rem 0;
}

#gyorslinkek .quick-links > * {
  flex-shrink: 0;
}
```

Keep `.quick-links-wrapper` as a simple relative flex container (needed until Task 3 simplifies markup). Remove the `::before` / `::after` edge fades and `.can-left` / `.can-right` rules (they only served horizontal scroll). If those selectors remain unused after Task 3, deleting them in this task is correct.

- [ ] **Step 2: Pack the three-widget row and add expanded modifier**

Change `.widgets-box--three-col` `justify-content: space-between` to `justify-content: flex-start`.

Immediately after `.widgets-box--three-col { ... }` add:

```css
.widgets-box--three-col.widgets-box--quicklinks-expanded #gyorslinkek {
  flex-basis: 100%;
}
```

Do **not** set `width: 100%` on that expanded rule. The existing `@media (max-width: 992px) { #gyorslinkek { width: 100%; } }` stays (stacked column on small screens).

- [ ] **Step 3: Stepper header styles**

After `.widget-header` (around line 853) add:

```css
.quicklinks-slot-stepper {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-shrink: 0;
}

.quicklinks-slot-count {
  min-width: 1.5rem;
  text-align: center;
  font-size: var(--text-sm);
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}
```

- [ ] **Step 4: Visual check**

Run: `npm run dev` (if not already running). Open `/`.

Expected: Gyorslinkek no longer stretches to fill leftover row space; date/time and weather sit immediately to its right (may still look wide until Task 3 sets `--quicklink-slots` and removes nowrap via the CSS already applied). No JS errors.

- [ ] **Step 5: Commit** (skip unless the user asked to commit)

```bash
git add src/styles/global.css
git commit -m "$(cat <<'EOF'
Cap Gyorslinkek width and pack homepage widgets instead of space-between.

EOF
)"
```

---

### Task 3: Homepage stepper, skeletons, drop arrows

**Files:**
- Modify: `src/routes/(public)/+page.svelte`

**Interfaces:**
- Consumes: `DEFAULT_QUICKLINK_SLOTS`, `MIN_QUICKLINK_SLOTS`, `MAX_QUICKLINK_SLOTS`, `readSlotCount()`, `writeSlotCount(n)` from `$lib/quickLinksDisplay.js`
- Produces: `slotCount`, `--quicklink-slots`, `widgets-box--quicklinks-expanded` when `slotCount > DEFAULT_QUICKLINK_SLOTS`, skeleton placeholders while `promotedLoading`

Keep **legacy** Svelte syntax in this file (`let`, `$:`, `on:click`). Do not introduce runes.

- [ ] **Step 1: Script — import, state, drop scroll helpers**

Add import:

```js
import {
    DEFAULT_QUICKLINK_SLOTS,
    MAX_QUICKLINK_SLOTS,
    MIN_QUICKLINK_SLOTS,
    readSlotCount,
    writeSlotCount,
} from "$lib/quickLinksDisplay.js";
```

Remove: `canScrollLeft`, `canScrollRight`, `quickLinksContainer`, `LINKS_BEFORE_ARROWS`, `allLinksCount`, `showArrows`, `checkScroll`, `dragScroll`, `scrollLinks`.

Add:

```js
let slotCount = DEFAULT_QUICKLINK_SLOTS;

$: quicklinksExpanded = slotCount > DEFAULT_QUICKLINK_SLOTS;
$: skeletonCount = promotedLoading
    ? Math.max(0, DEFAULT_QUICKLINK_SLOTS - 1 - userLinks.length)
    : 0;

function setSlotCount(next) {
    slotCount = writeSlotCount(next);
}

function decreaseSlots() {
    setSlotCount(slotCount - 1);
}

function increaseSlots() {
    setSlotCount(slotCount + 1);
}
```

In `onMount`, immediately after `loadUserLinks();` add:

```js
slotCount = readSlotCount();
```

`loadUserLinks` stays first so `skeletonCount` can use personal links on the first client render.

- [ ] **Step 2: Markup — header stepper, wrap strip, skeletons**

Replace the Gyorslinkek widget (from `<div id="gyorslinkek"` through its closing `</div>` before `<DateTimeWidget />`) with:

```svelte
        <div
            id="gyorslinkek"
            class="widget"
            style:--quicklink-slots={slotCount}
        >
            <div class="widget-header">
                <h3 class="widget-title">Gyorslinkek</h3>
                <div
                    class="quicklinks-slot-stepper"
                    role="group"
                    aria-label="Megjelenített gyorslinkek száma"
                >
                    <button
                        type="button"
                        class="btn btn-xs"
                        disabled={slotCount <= MIN_QUICKLINK_SLOTS}
                        aria-label="Kevesebb hely"
                        on:click={decreaseSlots}
                    >−</button>
                    <span class="quicklinks-slot-count" aria-live="polite">{slotCount}</span>
                    <button
                        type="button"
                        class="btn btn-xs"
                        disabled={slotCount >= MAX_QUICKLINK_SLOTS}
                        aria-label="Több hely"
                        on:click={increaseSlots}
                    >+</button>
                </div>
            </div>
            <div class="quick-links-wrapper">
                <div class="quick-links widget-content">
                    <button
                        type="button"
                        class="link-card link-card-add"
                        on:click={openAddLink}
                        title="Új gyorslink hozzáadása"
                        aria-label="Új gyorslink hozzáadása"
                    >
                        <span class="link-card-icon link-card-icon-add">
                            <span class="link-card-add-plus">+</span>
                        </span>
                        <span class="link-card-title">Új</span>
                    </button>

                    {#each userLinks as q (q.id)}
                        <div class="link-card">
                            <a
                                href={q.url}
                                target="_blank"
                                rel="nofollow noopener"
                                class="link-card-link"
                                title={q.title}
                            >
                                <span class="link-card-icon" style:background={q.bg_color || "#2f4f4f"}>
                                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--border-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                                </span>
                                <span class="link-card-title">{truncateTitle(q.title)}</span>
                            </a>
                            <button
                                type="button"
                                class="link-card-edit"
                                on:click={(e) => openEditLink(q, e)}
                                title="Szerkesztés"
                                aria-label="Szerkesztés"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
                                </svg>
                            </button>
                        </div>
                    {/each}

                    {#if promotedLoading}
                        {#each { length: skeletonCount } as _, i (i)}
                            <div class="link-card link-card--skeleton" aria-hidden="true">
                                <span class="link-card-icon skeleton"></span>
                                <span class="link-card-title skeleton"></span>
                            </div>
                        {/each}
                    {:else if !promotedError}
                        {#each promotedLinks as q (q.id)}
                            <div class="link-card link-card--promoted">
                                <a
                                    href={q.url}
                                    target="_blank"
                                    rel="nofollow noopener"
                                    class="link-card-link"
                                    title={q.title}
                                >
                                    <span class="link-card-icon" style:background={q.bg_color || "#2f4f4f"}>
                                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--border-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                                    </span>
                                    <span class="link-card-title">{truncateTitle(q.title)}</span>
                                </a>
                                <span class="link-card-promoted-badge" title="Promóciós link" aria-label="Promóciós link">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon></svg>
                                </span>
                            </div>
                        {/each}
                    {/if}
                </div>
            </div>
        </div>
```

On the parent row, replace:

```svelte
    <div class="widgets-box--three-col">
```

with:

```svelte
    <div
        class="widgets-box--three-col"
        class:widgets-box--quicklinks-expanded={quicklinksExpanded}
    >
```

- [ ] **Step 3: Validate the Svelte file**

Run svelte-autofixer on `src/routes/(public)/+page.svelte` (MCP `svelte-autofixer` or `npx @sveltejs/mcp svelte-autofixer src/routes/(public)/+page.svelte`). Fix any reported issues and re-run until clean. Do not convert the file to runes unless the autofixer requires it for a real error.

- [ ] **Step 4: Manual check on `/`**

With `npm run dev`:

1. Default: stepper shows `7`; `−` disabled; Gyorslinkek ~7 cards wide; date/time and weather on the same row, next to it (not opposite edges).
2. Click `+` until `8+`: Gyorslinkek occupies the first row; date/time and weather sit together on the next row; cards wrap at `N` (not stretched full viewport).
3. `+` disables at `14`. Reload: `N` persists.
4. All personal and admin cards visible; extras wrap; no scroll arrows.
5. Throttle network: with no personal links, Új + 6 skeletons (7 slots) until promoted load finishes.

- [ ] **Step 5: Commit** (skip unless the user asked to commit)

```bash
git add src/routes/(public)/+page.svelte
git commit -m "$(cat <<'EOF'
Let visitors set Gyorslinkek width and keep homepage widgets packed.

EOF
)"
```

---

### Task 4: Frontend checklist

**Files:**
- Modify: `tests/frontend-test-checklist.md` (section 1.1 Homepage)

**Interfaces:**
- Consumes: behavior from Tasks 1–3
- Produces: manual test items covering stepper, layout, wrap, skeletons, persistence

- [ ] **Step 1: Add checklist items under 1.1 Homepage**

After the existing “Clicking a quick link opens the URL in a new tab” item, insert:

```markdown
- [ ] Gyorslinkek default slot count is 7 (including Új); − is disabled; date/time and weather sit on the same row immediately beside it
- [ ] Increasing slots above 7 moves date/time and weather to the next row, still adjacent (not opposite edges)
- [ ] Slot stepper clamps at 7 and 14; chosen N persists after reload (`localStorage` `quick_links_display_count`)
- [ ] All personal and admin quicklinks remain visible; extras wrap inside the widget; no scroll arrows
- [ ] While promoted links load and the visitor has no personal links, Új plus 6 skeletons fill 7 slots
- [ ] Admin/promoted cards show the star badge and cannot be edited or deleted on the homepage
```

- [ ] **Step 2: Re-run helper tests**

Run: `node --test tests/quickLinksDisplay.test.js`

Expected: PASS.

- [ ] **Step 3: Commit** (skip unless the user asked to commit)

```bash
git add tests/frontend-test-checklist.md
git commit -m "$(cat <<'EOF'
Document homepage Gyorslinkek slot-width layout checks.

EOF
)"
```

---

## Spec coverage (self-review)

| Spec requirement | Task |
| --- | --- |
| Helper constants, clamp, read/write, storage key | 1 |
| SSR-safe `localStorage` missing → 7 | 1 |
| CSS width, wrap, flex-start, expanded flex-basis without width 100% | 2 |
| Stepper UI, `--quicklink-slots`, expanded class | 3 |
| Drop arrows / drag-scroll | 3 |
| Skeletons `max(0, 7 - 1 - userLinks.length)` | 3 |
| All cards always shown; admin after personal | 3 |
| Promoted error: no admin cards, N unchanged | 3 (existing `promotedError` branch kept) |
| Manual checklist | 4 |
| Admin page / API unchanged | no task (intentional) |
