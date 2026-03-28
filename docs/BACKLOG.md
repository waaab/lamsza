# Product backlog – post–settlements / attractions foundation

**Purpose:** Ordered, actionable steps after the settlements, counties, geo locations, and attractions work landed in code. Items are sequenced so dependencies run first; parallelizable items are noted.

**Related docs:** [PLAN_SETTLEMENTS_LOCATIONS_ATTRACTIONS.md](PLAN_SETTLEMENTS_LOCATIONS_ATTRACTIONS.md), [SETTLEMENTS_AND_LOCATIONS.md](SETTLEMENTS_AND_LOCATIONS.md).

---

## P0 – Discovery and consistency (before new features)

### P0.1 – Verify migration on all environments

- **Do:** Run `backend/migrations/migration_settlements_attractions.sql` (and any prerequisite patches) on staging and production DBs; confirm `locations` is a view, `locations_legacy` exists, seed attraction(s) present if expected.
- **Done when:** App starts, county and settlement pages load, `/api/attractions` returns data where seeded.

### P0.2 – Manual smoke pass

- **Do:** Execute `tests/frontend-test-checklist.md` for navigation, county/settlement/attraction routes, admin látnivalók CRUD.
- **Done when:** Critical paths checked; failures logged as new backlog items or bugs.

---

## P1 – Unified search includes látnivalók ✅ (implemented March 2026)

### P1.1 – Backend: search attractions

- **Done:** `backend/internal/search/unified.go` returns `attractions` and `historical_seats` arrays on `GET /api/search?q=...`.

### P1.2 – Frontend: render attraction results

- **Done:** `SearchEngine.svelte` sections for látnivalók and történelmi székek with correct links.

### P1.3 – Optional: include historical seats in search

- **Done:** Included in unified search and autosuggest (`search.go`).

---

## P2 – Public historical seat (szék) pages ✅

### P2.1 – Route and data loading

- **Done:** `/szekek` index, `/szekek/{slug}` detail; legacy `/{slug}-szek` and `/szek/…` → 301; `GET /api/historical_seats?slug=` for single seat.

### P2.2 – Navigation entry

- **Done:** Header link “Székek”, `/megyek` link to `/szekek`.

---

## P3 – Látnivalók context on settlement pages ✅ (MVP)

### P3.1 – Same-county list (MVP)

- **Done:** Settlement page loads `/api/attractions?county_slug=...` and shows “Látnivalók … megyében” when non-empty.

### P3.2 – Hybrid overrides (optional, from PLAN)

- **Status:** Not implemented; add `attraction_settlements` + admin when needed.

---

## P4 – Admin and content parity ✅

### P4.1 – Counties and historical seats in admin

- **Done:** Megyék tab: Markdown textareas + `PUT /api/admin/counties` and `PUT /api/admin/historical_seats`.

### P4.2 – Markdown preview consistency

- **Done:** County public page renders `counties.content` via `Markdown.svelte`; szék pages render `historical_seats.content`.

---

## P5 – Documentation ✅

### P5.1 – Update SETTLEMENTS_AND_LOCATIONS.md

- **Done:** “Current architecture” section at top; legacy sections retained with note.

### P5.2 – Update PLAN closing section

- **Done:** Section 12 in PLAN points here.

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

1. P0.1 → P0.2 (operator / QA)  
2. P6 as needed  

---

*Last updated: March 2026*
