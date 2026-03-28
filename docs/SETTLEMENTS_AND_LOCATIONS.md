# Settlements & Locations – Technical Documentation (Draft)

This document describes how locations (settlements, cities, towns, villages, counties) are governed across the Lámsza application: database, backend API, admin interface, and frontend.

---

## 1. Location Types (Hungarian Terminology)

The app uses a single `locations` table for all geographic entities. The `type` column distinguishes:

| Type (HU)   | English      | Description                          |
|-------------|--------------|--------------------------------------|
| **megye**   | county       | Top-level administrative unit        |
| **város**   | city/town    | Urban settlement                     |
| **municípium** | municipality | Special urban status (e.g. county seat) |
| **község**  | commune      | Rural administrative unit             |
| **falu**    | village      | Smallest settlement unit             |

**Counties in scope (Székelyföld):** Hargita, Kovászna, Maros

---

## 2. Database Schema

### 2.1 Base Schema (`backend/migrations/schema.sql`)

```sql
CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    name_ro VARCHAR(255),
    name_de VARCHAR(255),
    county VARCHAR(255),
    type VARCHAR(50),
    CONSTRAINT unique_location UNIQUE (name, county, type)
);
```

### 2.2 Additional Columns (from migrations)

| Column        | Migration                    | Type    | Description                          |
|---------------|------------------------------|---------|--------------------------------------|
| `post_code`   | `patch_locations_metadata.sql` | TEXT    | Postal code (irányítószám)           |
| `coordinates` | `patch_locations_metadata.sql` | TEXT    | Geographic coordinates               |
| `population`  | `patch_locations_metadata.sql` | TEXT    | Population (fő)                      |
| `area`        | `patch_locations_metadata.sql` | TEXT    | Area in km²                          |
| `is_county_seat` | `patch_county_seat.sql`     | BOOLEAN | Whether this settlement is county seat |

### 2.3 Slug and Related Columns (`patch_locations_slugs.sql`)

| Column       | Type    | Purpose                          |
|--------------|---------|----------------------------------|
| `slug`       | VARCHAR(120) | URL-friendly name (e.g. `csikszereda`) |
| `county_slug`| VARCHAR(120) | URL-friendly county (e.g. `hargita`)   |
| `crest`      | TEXT    | Coat of arms image URL           |
| `parent_id`  | INTEGER | FK to parent location (village → town) |

Migration also backfills `slug` and `county_slug` for existing rows and inserts county records (Hargita, Kovászna, Maros) if missing.

### 2.4 Relationships

- **entries** → `location_id` → `locations(id)` – Directory entries belong to a settlement
- **events** → `location_id` → `locations(id)` – Events are tied to a settlement
- **locations** → `parent_id` → `locations(id)` – Villages can link to a parent town

### 2.5 Slug Generation

- **Backend:** `utils.Slugify()` – lowercases, replaces Hungarian accents (á→a, ő→o, etc.), replaces non-alphanumeric with `-`
- **Stored on create/update:** `slug` from `name`, `county_slug` from `county`

---

## 3. Backend API

### 3.1 Location Endpoints

| Method | Endpoint              | Handler               | Description                    |
|--------|------------------------|------------------------|--------------------------------|
| GET    | `/api/locations`       | `HandleAdminLocations` | List locations (optional `?type=`, `?county_slug=`) |
| POST   | `/api/locations`       | `HandleAdminLocations` | Create location                |
| PUT    | `/api/locations`       | `HandleAdminLocations` | Update location                |
| DELETE | `/api/locations?id=`    | `HandleAdminLocations` | Delete location                |
| PUT    | `/api/admin/county_seat`| `HandleSetCountySeat`  | Set county seat                |

**Note:** Both `/api/locations` and `/api/admin/locations` are registered and use the same handler.

**Query params (GET):** Optional `?type=` and `?county_slug=` for server-side filtering.

### 3.2 Directory (Entries) API

**Endpoint:** `GET /api/directory` (alias: `/api/entries`)

**Query parameters:**

| Param           | Description                          |
|-----------------|--------------------------------------|
| `location_slug` | Filter by settlement slug (e.g. `csikszereda`) |
| `county_slug`   | Filter by county slug (e.g. `hargita`)         |
| `q`             | Full-text search                      |
| `category`      | Category filter                        |
| `tag`           | Tag filter                            |

**Logic:** Entries are joined with `locations` via `location_id`. Filtering uses `l.slug` and `l.county_slug`.

### 3.3 Other Location-Related Endpoints

| Endpoint                    | Params              | Purpose                              |
|----------------------------|---------------------|--------------------------------------|
| `GET /api/weather`         | `slug`              | Weather for settlement by slug       |
| `GET /api/weather/county`   | `slug` (county)     | Weather for county cities            |
| `GET /api/events`          | `location_slug`, `county_slug` | Events by location/county   |
| `GET /api/search`          | `q`                 | Unified search (includes locations)  |
| `GET /api/config/public`   | –                   | Includes `my_location_slug`, `my_location_county_slug` |

---

## 4. Admin Interface

### 4.1 Locations Tab (`src/routes/admin/+page.svelte`)

**Create/Edit form fields:**

| Field (HU)           | DB Column   | Notes                                  |
|----------------------|-------------|----------------------------------------|
| Település neve (HU)  | `name`      | Required                                |
| Név (RO)             | `name_ro`   | Optional                                |
| Név (DE)             | `name_de`   | Optional                                |
| Megye                | `county`    | Select: Hargita, Kovászna, Maros       |
| Típus                | `type`      | Select: város, község, falu, megye, municípium |
| Posta kód            | `post_code` |                                        |
| Koordináták          | `coordinates` |                                    |
| Lakosság (fő)       | `population` |                                    |
| Terület (km²)       | `area`       |                                    |
| Címer URL            | `crest`      |                                        |
| Kapcsolt település   | `parent_id`  | Select from existing locations         |

**Table columns:** ID, Név (HU/RO/DE), Megye, Típus, Irányítószám, Koordináták, Lakosság, Terület, Címer, Szülő település, Szerk., Törlés

### 4.2 County Seat Management (Megyeszékhelyek)

- **Tab:** Counties
- **Logic:** For each county, lists settlements (excluding type `megye`) with radio buttons
- **API:** `PUT /api/admin/county_seat` with `{ "location_id": <id> }`
- **Backend:** Clears `is_county_seat` for all in that county, then sets it for the selected location

### 4.3 Site Settings – Default Location

- **Setting:** `my_location_slug` (e.g. `csikszereda`)
- **Used for:** Homepage weather, “my location” behavior
- **Admin:** Dropdown of all locations

---

## 5. Frontend Display

### 5.1 URL Structure

| Pattern                    | Example              | Purpose        |
|----------------------------|----------------------|----------------|
| `/{countySlug}-megye`      | `/hargita-megye`     | County page    |
| `/{countySlug}-megye/{slug}` | `/hargita-megye/csikszereda` | Settlement page |
| `/megyek`                  | –                    | County list    |
| `/varosok`                 | –                    | City list      |
| `/falvak`                  | –                    | Village list   |

### 5.2 List Pages

| Route      | File                          | Filter (client-side)                    |
|------------|-------------------------------|-----------------------------------------|
| `/megyek`  | `megyek/+page.svelte`         | `GET /api/locations?type=megye` (server-side) |
| `/varosok` | `varosok/+page.svelte`        | `type` in ["város", "municípium"]       |
| `/falvak`  | `falvak/+page.svelte`         | `type` in ["falu", "község"]            |

**Data source:** `/megyek` uses `?type=megye`; `/varosok` and `/falvak` fetch all and filter client-side (multiple types).

**Links:**
- Counties: `/{slug}-megye` (e.g. `/hargita-megye`)
- Cities/Villages: `/{county_slug}-megye/{slug}` (e.g. `/hargita-megye/csikszereda`)

### 5.3 County Page (`[countySlug]-megye/+page.svelte`)

**Displayed:**
- County name (HU/RO/DE)
- Irányítószám, koordináták, lakosság, terület
- Közigazgatási forma
- Megyeszékhely link
- County crest
- Weather for county cities
- Events widget
- News ticker
- Child settlements grid
- Directory entries

**Data fetching:**
1. `GET /api/locations?county_slug=${town}` – returns county record + all settlements in county
2. County data: `locations.find(l => l.slug === town)` (county has type=megye)
3. Child settlements: `locations.filter(l => l.type !== "megye")`
4. **Directory:** `GET /api/directory?county_slug=${town}` ✓

### 5.4 Settlement Page (`[countySlug]-megye/[slug]/+page.svelte`)

**Displayed:**
- Settlement name (HU/RO/DE)
- Irányítószám, koordináták, lakosság, terület
- Közigazgatási forma
- Parent settlement link
- County link
- Crest
- Weather, Events, News widgets
- Directory entries

**Data fetching:**
1. `GET /api/locations` – find by `slug`
2. `GET /api/directory?location_slug=${slug}` ✓ Correct

### 5.5 Components Using Location Data

| Component       | Usage                                                                 |
|-----------------|-----------------------------------------------------------------------|
| `EntryCard`     | Links to `/{county_slug}-megye/{location_slug}` for entry location   |
| `Breadcrumbs`   | `countySlug`, `countyName`, `settlementSlug`, `settlementName`, `settlementType` |
| `NewsWidget`    | `settlementSlug` for local news                                      |
| `EventsWidget`  | `settlementSlug`, `countySlug` for filtering                          |
| `WeatherWidget` | `settlementSlug` for single-settlement weather                       |

---

## 6. Naming Conventions (Settlement vs Location)

The codebase uses different terms for the same concept. Summary:

| Layer    | Term(s) Used | Context |
|----------|--------------|---------|
| **Database** | `locations` (table), `location_id` (FK) | Technical |
| **Backend**  | `Location` (struct), `location_slug`, `location_id`, `LocationName` | API, models |
| **Admin**    | "Települések" (tab), "Település" (form), `locations`, `editingLocation` | UI (HU) + vars |
| **Frontend** | `settlementData`, `settlementSlug`, `settlementName`, `childSettlements` | Page-level vars |
| **Frontend** | `entry.location`, `entry.location_slug`, `locationName` (prop) | Entry/Event data |
| **Search**   | `entity_type: "settlement"`, "Települések" (section title) | Search results |

**Canonical technical term:** `location` (DB, API, backend models)

**User-facing term:** `Település` / `Települések` (Hungarian)

**Frontend convention:** `settlement*` for page/component props (settlement page, widgets); `location*` when mirroring API/entry data.

---

## 7. Resolved Issues (March 2025)

- ✅ **County directory:** Fixed to use `county_slug` instead of `slug`+`type`
- ✅ **Locations migration:** Added `patch_locations_slugs.sql` for slug, county_slug, crest, parent_id
- ✅ **API filtering:** GET `/api/locations` supports `?type=` and `?county_slug=`
- ✅ **County names:** Standardized schema1.sql to Hungarian (Hargita, Kovászna, Maros)
- ✅ **Type fix:** schema1.sql `falvak` → `falu` (valid type)

---

## 8. Data Flow Summary

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           DATABASE (PostgreSQL)                         │
│  locations: id, name, name_ro, name_de, county, type, slug,             │
│             county_slug, post_code, coordinates, population, area,      │
│             crest, parent_id, is_county_seat                             │
│  entries.location_id → locations.id                                     │
│  events.location_id → locations.id                                      │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           BACKEND (Go)                                   │
│  /api/locations          → CRUD, optional ?type=, ?county_slug=         │
│  /api/admin/county_seat  → Set county seat                              │
│  /api/directory          → location_slug, county_slug (entries filter)  │
│  /api/weather, /api/events, /api/search, /api/config/public              │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           FRONTEND (SvelteKit)                           │
│  Admin: Locations tab, Counties tab, Site settings                     │
│  Public: /megyek, /varosok, /falvak, /[county]-megye, /[county]-megye/[slug] │
│  Filtering: Server-side for megyek; client-side for varosok/falvak      │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 9. Recommendations

1. **Add TypeScript types:** Define `Location` and related interfaces for frontend type safety.
2. **Optional:** Add `?slug=` to GET `/api/locations` for single-location lookup (settlement page optimization).
3. **Run migration:** Apply `patch_locations_slugs.sql` on existing databases.

---

*Document version: Draft – March 2025 (updated with fixes)*
