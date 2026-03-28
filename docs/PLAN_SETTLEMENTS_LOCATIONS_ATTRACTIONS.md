# Plan: Settlements, Locations, Attractions & Content Overhaul

**Status:** Conceptual / Planning  
**Created:** March 2025

---

## Executive Summary

This plan refactors the geographic and administrative model of Lámsza to:

1. **Separate "settlement" (administrative entity) from "location" (geographic position)**
2. **Add historical seats (székek)** above counties in the hierarchy
3. **Introduce attractions (látnivalók)** with their own locations, linkable to nearby settlements
4. **Unify content management** so county, settlement, and attraction pages have editable content

---

## 1. Conceptual Model

### 1.1 Current State (Problems)

| Issue | Current | Problem |
|-------|---------|---------|
| **Overloaded "location"** | One `locations` table for counties, cities, villages | "Location" mixes administrative entity with geographic concept |
| **Flat hierarchy** | Counties and settlements in same table, distinguished by `type` | No historical/regional grouping |
| **No attractions** | – | Can't list Szent Anna-tó, castles, etc. |
| **Content** | `pages` table for static pages only | County/settlement pages have no editable body content |

### 1.2 Target State

```
Historical Seat (szék)     e.g. Csíkszék, Udvarhelyszék, Háromszék
    │
    └── County (megye)     e.g. Hargita, Kovászna, Maros
            │
            └── Settlement (város, falu, község, municípium)
                    │
                    ├── has Location (lat, lon, address)
                    └── hosts Entries, Events

Attraction (látnivaló)     e.g. Szent Anna-tó, Vár
    │
    ├── has Location (lat, lon)
    └── "near" Settlements (by county or radius)
```

**Key distinction:**
- **Settlement** = administrative/place entity (name, type, county, crest, etc.)
- **Location** = geographic point (latitude, longitude, optional address) — used by settlements, attractions, and **default homepage weather**

**Weather:** Locations drive weather display. The default homepage weather uses a site setting (e.g. `my_location_slug` → settlement's location). **Attractions with coordinates can also have weather** — e.g. Szent Anna-tó (46.126, 25.888). Weather APIs (Open-Meteo, etc.) accept lat/lon directly, so we can show weather for any entity with a `location_id`.

---

## 2. Proposed Data Model

### 2.1 New / Renamed Tables

| Table | Purpose |
|-------|---------|
| `locations` | **New meaning:** Geographic coordinates. `id`, `latitude`, `longitude`, `address`, `elevation` (optional) |
| `settlements` | **Replaces** current `locations` for cities/villages. Administrative entities. |
| `counties` | **New.** Counties (megye) as first-class entities. |
| `historical_seats` | **New.** Székek (Csíkszék, Udvarhelyszék, Háromszék). |
| `attractions` | **New.** Látnivalók (Szent Anna-tó, etc.). |

### 2.2 Entity Relationships

```
historical_seats
    id, name, name_ro, name_de, slug, content (Markdown)

counties
    id, name, name_ro, name_de, slug
    location_id (FK) — county "center" for map, default weather
    content

county_historical_seats (junction, many-to-many)
    county_id, historical_seat_id
    — Hargita → Csíkszék, Udvarhelyszék; Kovászna → Háromszék; Maros → Marosszék

settlements
    id, county_id (FK), name, name_ro, name_de, slug, type (város, falu, község, municípium)
    location_id (FK) — settlement coordinates
    parent_id (FK) — village → town
    post_code, population, area, crest, is_county_seat
    content

locations (geographic)
    id, latitude, longitude, address, elevation

attractions
    id, county_id (FK) — attraction belongs to a county
    name, name_ro, name_de, slug, description (short)
    location_id (FK) — coordinates for map + weather
    content (Markdown)
    featured_image (banner)
    — gallery: attraction_images table (id, attraction_id, url, sort_order)

attraction_settlements (junction, optional)
    attraction_id, settlement_id
    — explicit "show this attraction on this settlement's page"
    — alternative: derive "nearby" by county or radius
```

### 2.3 Entries & Events

- **entries** → `settlement_id` (was `location_id`)
- **events** → `settlement_id` (was `location_id`)

---

## 3. Attractions: "Nearby" Logic

**Options for "display attraction X on settlement Y's page":**

| Approach | Pros | Cons |
|----------|------|------|
| **A. Same county** | Simple, no extra tables | Attraction in Hargita shows for all Hargita settlements |
| **B. Radius (km)** | Geographically accurate | Need PostGIS or manual haversine; more compute |
| **C. Explicit links** | Full control, no false positives | Admin must manually link each attraction–settlement pair |
| **D. Hybrid** | County + optional explicit links | Best of both; can override "nearby" per settlement |

**Recommendation:** Start with **D (Hybrid)**:
- Default: attraction in county X → shown on all settlement pages in county X
- Optional: `attraction_settlements` for "also show on settlement Y" or "hide from settlement Z"

---

## 3.1 Historical Seats (Székek) — Mapping

Based on [Wikipedia: Székelyföld történelmi székei](https://hu.wikipedia.org/wiki/Székelyföld_történelmi_székei):

| Szék | Modern county (approx.) | Notes |
|------|-------------------------|-------|
| **Udvarhelyszék** | Hargita | Anyaszék (central seat); Székelyudvarhely area |
| **Csíkszék** | Hargita | Csíkszereda area; Gyergyószék, Kászonszék fiúszékek |
| **Háromszék** | Kovászna | Sepsiszentgyörgy, Kézdi, Orbai, Sepsi; Miklósvárszék fiúszék |
| **Marosszék** | Maros | Marosvásárhely area; Szeredaszék fiúszék |
| **Aranyosszék** | (Alba/Fehér) | Exclave, west; for future Transylvania expansion |

**For current 3 counties:** Hargita has both Csíkszék and Udvarhelyszék; Kovászna → Háromszék; Maros → Marosszék. Design for expansion (more counties, more székek).

---

## 4. Content Management

### 4.1 Current

- `pages` table: slug, title, content — for static pages (iranyelvek, sütik, feltetelek)
- County/settlement pages: no editable body content; data comes from DB fields only

### 4.2 Proposed

**Option A: Content column on each entity**
- `counties.content`, `settlements.content`, `attractions.content`
- Simple, one place per entity
- Rich text (Markdown or HTML) stored in TEXT

**Option B: Unified entity_content table**
- `entity_type` (county, settlement, attraction), `entity_id`, `content`
- Easier to add new entity types later
- Slightly more complex queries

**Option C: Extend pages**
- `pages` gains `entity_type`, `entity_id` (nullable)
- Entity pages: entity_type + entity_id set
- Static pages: both null, use slug only
- Reuses existing pages admin UI

**Recommendation:** **Option A** for clarity and simplicity. Each entity owns its content. Admin has dedicated forms per entity type. **Content format: Markdown.**

---

## 5. Admin Structure

### 5.1 Separate Sections

| Section | Entities | CRUD | Content |
|---------|----------|------|---------|
| **Historical seats** | Csíkszék, Udvarhelyszék, Háromszék | Create, edit, delete | Rich text body |
| **Counties** | Hargita, Kovászna, Maros | Create, edit, delete | Rich text body, county seat |
| **Settlements** | Cities, towns, villages | Create, edit, delete | Rich text body |
| **Attractions** | Szent Anna-tó, etc. | Create, edit, delete | Rich text body, location |
| **Entries** | (unchanged) | Directory entries | – |
| **Events** | (unchanged) | Events | – |
| **Pages** | Static pages | (existing) | – |

### 5.2 Admin UX

- **Sidebar:** Historical seats | Counties | Settlements | Attractions | Entries | Events | …
- **List + detail:** Each section has list view and edit form
- **Content editor:** Markdown for `content` fields
- **Location picker:** For settlements and attractions — map or lat/lon input

---

## 6. URL Structure (Proposed)

| Page type | URL | Example |
|-----------|-----|---------|
| County | `/{countySlug}-megye` | `/hargita-megye` |
| Settlement | `/{countySlug}-megye/{slug}` | `/hargita-megye/csikszereda` |
| Attraction | `/{countySlug}-megye/{slug}` | `/hargita-megye/szent-anna-to` |
| Historical seat | `/{slug}-szek` (legacy) or `/szekek/{slug}` | `/szekek/csikszek` |

**Settlements and attractions share** `/{countySlug}-megye/{slug}`. Resolution: look up slug in settlements for that county first, then in attractions. **Slug uniqueness:** enforce unique (slug, county_id) within settlements and within attractions; no overlap between settlement and attraction slugs in the same county (admin responsibility or DB constraint).

**Breadcrumbs** must respect hierarchy and parent relationships:
- Settlement: `Főoldal › Csíkszék › Hargita megye › Csíkszereda`
- Attraction: `Főoldal › Hargita megye › Szent Anna-tó`
- County: `Főoldal › Csíkszék › Hargita megye`
- Historical seat: `Főoldal › Csíkszék`

---

## 7. Migration Strategy

### 7.1 Phases

| Phase | Scope | Risk |
|-------|-------|------|
| **1. Add locations table** | New `locations` table; backfill from `coordinates` in current `locations` | Low |
| **2. Add counties, historical_seats** | New tables; migrate county data from current `locations` (type=megye) | Medium |
| **3. Rename locations → settlements** | Create `settlements`, migrate data, add `location_id`, `county_id` | High |
| **4. Update entries, events** | `location_id` → `settlement_id` | Low |
| **5. Add attractions** | New table, admin, frontend | Low |
| **6. Add content columns** | `content` on counties, settlements, attractions | Low |
| **7. Admin overhaul** | New sections, content editor | Medium |

### 7.2 Backward Compatibility

- **API:** Introduce new endpoints (`/api/settlements`, `/api/counties`, etc.); deprecate `/api/locations` gradually
- **Frontend:** Update routes and components in one release
- **Data:** Migration scripts to transform `locations` → `settlements` + `counties` + `locations` (geo)

---

## 8. Resolved Decisions

| Question | Decision |
|----------|----------|
| **Historical seats** | Székely-related székek. Map from [Wikipedia: Székelyföld történelmi székei](https://hu.wikipedia.org/wiki/Székelyföld_történelmi_székei). Support expansion to Transylvania. |
| **County–szék mapping** | Many-to-many: `county_historical_seats` junction. Hargita → Csíkszék, Udvarhelyszék; Kovászna → Háromszék; Maros → Marosszék. |
| **Attraction imagery** | `featured_image` (banner) + `attraction_images` gallery table. |
| **Content format** | Markdown. |
| **Attraction URLs** | `/{countySlug}-megye/{slug}` — same pattern as settlements, linked to county. |
| **Breadcrumbs** | Respect hierarchy: Home › [Szék] › County › [Settlement/Attraction]. |

---

## 9. Recommended Approach

### Phase 1: Foundation (low risk)
1. Create `locations` table (id, latitude, longitude, address)
2. Create `historical_seats` and `counties` tables
3. Migrate county rows from current `locations` into `counties`
4. Add `content` column to counties (and wire admin)

### Phase 2: Settlements (medium risk)
1. Create `settlements` table with `county_id`, `location_id`
2. Migrate non-county rows from `locations` → `settlements`
3. Create `locations` rows from existing `coordinates` where possible
4. Update `entries` and `events` to use `settlement_id`
5. Deprecate old `locations` table (or rename to `locations_legacy` temporarily)

### Phase 3: Attractions (low risk)
1. Create `attractions` table with `location_id`
2. Build admin CRUD
3. Build frontend: attraction detail page, "nearby" on settlement pages

### Phase 4: Admin & content
1. Add `content` to settlements and attractions
2. Restructure admin: Historical seats | Counties | Settlements | Attractions
3. Markdown content editor
4. Weather: extend to support `location_id` / coordinates (for attractions, default homepage)

---

## 10. Summary

| Concept | Before | After |
|---------|--------|-------|
| **Location** | Administrative entity (county, city, village) | Geographic point (lat, lon, address) |
| **Settlement** | Mixed with "location" | City, town, village — belongs to county |
| **County** | Row in locations (type=megye) | First-class entity, belongs to historical seat |
| **Historical seat** | – | New; above county |
| **Attraction** | – | New; has location (weather!), county, featured_image + gallery |
| **Content** | Static pages only | County, settlement, attraction pages have Markdown content |

---

## 11. Weather Integration

| Use case | Source | Notes |
|----------|--------|-------|
| **Homepage default** | Site setting `my_location_slug` → settlement → `location_id` → lat/lon | Existing; ensure location has coordinates |
| **Settlement page** | Settlement's `location_id` → lat/lon | Current flow: slug → name → geocode. **Improve:** use stored lat/lon when available (more reliable) |
| **Attraction page** | Attraction's `location_id` → lat/lon | **New.** Weather APIs accept lat/lon; Szent Anna-tó (46.126, 25.888) works. |

**Implementation:** Add optional `?lat=&lon=` or `?location_id=` to `/api/weather`. When coordinates provided, skip geocoding and fetch directly. Fallback to name-based geocode for backward compatibility.

---

## 12. Implementation status (March 2026)

Foundation migration, attractions, county/settlement public pages, unified search (including látnivalók and székek), `/szekek` and `/szekek/{slug}` pages, settlement-side látnivalók list, admin Markdown for counties and historical seats, and doc updates are tracked in [BACKLOG.md](BACKLOG.md). Optional follow-ups (e.g. `attraction_settlements` junction, multi-provider weather) remain in that backlog or [weather-and-widgets-plan.md](../weather-and-widgets-plan.md).

*Next steps: see [BACKLOG.md](BACKLOG.md) (P6 and manual verification items).*
