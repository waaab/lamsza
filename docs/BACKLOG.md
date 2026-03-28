# Product backlog – post–settlements / attractions foundation

**Purpose:** Ordered, actionable steps after the settlements, counties, geo locations, and attractions work landed in code. Items are sequenced so dependencies run first; parallelizable items are noted.

**Related docs:** [PLAN_SETTLEMENTS_LOCATIONS_ATTRACTIONS.md](PLAN_SETTLEMENTS_LOCATIONS_ATTRACTIONS.md), [SETTLEMENTS_AND_LOCATIONS.md](SETTLEMENTS_AND_LOCATIONS.md) (needs sync with current schema).

---

## P0 – Discovery and consistency (before new features)

### P0.1 – Verify migration on all environments

- **Do:** Run `backend/migrations/migration_settlements_attractions.sql` (and any prerequisite patches) on staging and production DBs; confirm `locations` is a view, `locations_legacy` exists, seed attraction(s) present if expected.
- **Done when:** App starts, county and settlement pages load, `/api/attractions` returns data where seeded.

### P0.2 – Manual smoke pass

- **Do:** Execute `tests/frontend-test-checklist.md` for navigation, county/settlement/attraction routes, admin látnivalók CRUD.
- **Done when:** Critical paths checked; failures logged as new backlog items or bugs.

---

## P1 – Unified search includes látnivalók

### P1.1 – Backend: search attractions

- **Do:** Extend `backend/internal/search/unified.go` to query `attractions` joined with `counties` (name fields, slug, county slug). Respect the same `LIMIT` / relevance patterns as other entity searches. Add a new JSON field on the unified response, e.g. `attractions: []` (define a small struct or reuse `handlers.Attraction`-shaped subset).
- **Done when:** `GET /api/search?q=...` (or the unified search endpoint in use) returns matching attractions for a known query; empty array when none.

### P1.2 – Frontend: render attraction results

- **Do:** In `SearchEngine.svelte` (and any shared types/helpers), add a section for látnivalók with title, link pattern `/{countySlug}-megye/{slug}`, and distinct styling (reuse existing result-type patterns: directory / weather / news).
- **Done when:** Homepage search shows attraction hits with working links; no console errors.

### P1.3 – Optional: include historical seats in search

- **Do:** If product agrees, add `historical_seats` ILIKE search and a small result section linking to future szék routes (or hide until P2 exists).
- **Done when:** Documented as skipped or implemented with links.

---

## P2 – Public historical seat (szék) pages

### P2.1 – Route and data loading

- **Do:** Add a public route (e.g. `/[szekSlug]-szek` or `/szek/[slug]` — align with PLAN URL table). Load content from `GET /api/historical_seats` or a dedicated `GET /api/historical_seats?slug=` if list-only is inefficient.
- **Done when:** Each szék in DB has a readable page (title, RO/DE names if present, Markdown `content`).

### P2.2 – Navigation entry

- **Do:** Link székek from a sensible place (footer, `/megyek`, or a new “Székely székek” index listing from API).
- **Done when:** Users can reach szék pages without typing URLs.

---

## P3 – Látnivalók context on settlement pages

### P3.1 – Same-county list (MVP)

- **Do:** On `[countySlug]-megye/[slug]/+page.svelte`, when the page is a **settlement** (not attraction), fetch `GET /api/attractions?county_slug=...` and show a compact “Látnivalók a megyében” block (reuse card styling from county page where possible).
- **Done when:** Settlement pages list attractions in the same county; links go to attraction detail URLs.

### P3.2 – Hybrid overrides (optional, from PLAN)

- **Do:** Add `attraction_settlements` junction migration; admin UI to attach/detach; filter settlement page list: show junction-linked first, then county default, or hide per rules you define.
- **Done when:** Documented behavior matches admin controls; migration applied.

---

## P4 – Admin and content parity

### P4.1 – Counties and historical seats in admin

- **Do:** Ensure admin can edit `counties.content` and `historical_seats.content` (Markdown) if not already complete; list views where helpful.
- **Done when:** Content saved in DB appears on public county and szék pages.

### P4.2 – Markdown preview consistency

- **Do:** Reuse `Markdown.svelte` for county and szék bodies where applicable; match attraction styling tokens if needed.
- **Done when:** No raw Markdown leakage; headings accessible.

---

## P5 – Documentation

### P5.1 – Update SETTLEMENTS_AND_LOCATIONS.md

- **Do:** Replace single-table `locations` description with: `locations` view, `settlements`, `counties`, `geo_locations`, `attractions`, FKs from entries/events to `settlements`.
- **Done when:** New developer can follow data flow without reading migrations first.

### P5.2 – Update PLAN closing section

- **Do:** Replace stale “Next step: Phase 1…” with pointer to this backlog or a short “Implemented / In progress” summary.
- **Done when:** PLAN no longer contradicts repo state.

---

## P6 – Later (separate initiatives)

| Item | Source | Notes |
|------|--------|--------|
| Multi-provider weather, caching, admin defaults | `weather-and-widgets-plan.md` | Own design pass; env keys |
| Shared icon module | `docs/icon-module-later.md` | Replace inline SVGs incrementally |
| Google Sign-In / accounts | `project-brief.md` | v1.2 scope |
| TypeScript types for Location / API | SETTLEMENTS recommendations | Nice-to-have |

---

## Suggested execution order

1. P0.1 → P0.2  
2. P1.1 → P1.2 → (P1.3 optional)  
3. P2.1 → P2.2  
4. P3.1 → (P3.2 if needed)  
5. P4.1 → P4.2  
6. P5.1 → P5.2  

P6 items are independent; schedule when priorities allow.

---

*Last updated: March 2026*
