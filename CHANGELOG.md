# Changelog

All notable changes to lamsza are recorded here, in English, for developers.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

Every notable change adds a line under `[Unreleased]` in the same commit. A line
users would notice ends with `[public]`; at a release those lines are rewritten
in plain Hungarian in `src/lib/publicChangelog.js`, which `/valtozasnaplo` and
the footer version read (WAYS_OF_WORKING R17).

---

## [Unreleased]

### Changed
- The confirm dialog follows the network baseline (Szótár's): a native dialog with a blurred backdrop, Mégse first and the confirm button on the right. [public]
- A refused Google sign-in says "Belépés sikertelen." instead of the server's reply. [public]
- "Székely Gugel" is now "Lámsza" in page titles, descriptions, the news placeholder and the extension manifest. [public]
- Buttons no longer select their label on a double click.
- `AppIcon.svelte` gains the Játszótér icon and round caps on the gear; it, the confirm, notice and sign-in dialogs and the three base stylesheets are now shared with lamsza-admin under the drift guard.
- `docs/network/UI_BASELINE.md` records the owner's 59 UI decisions; the Cursor UI rule points to it.
- `GoogleSignIn` / `SignInDialog` take an optional `signIn(credential)` so an app can sign in through its own API; `AppIcon` gains `trophy`, `login` and `logout`.
- `AppIcon` gains `list` and `add` (Szótár's Lista and Új szó toolbar buttons).

### Fixed
- The Lámsza button tooltip typo ("főódalra"). [public]

---

## [1.3.0] - 2026-10-07

### Added
- Directory catalog v2: browse by main category and subcategory, edit the tree in admin, file websites in a subcategory, list a business without a town. [public]
- Szótár words in unified search, with a Szótár filter pill and result cards. [public]
- Listing claims, joining a listing and suggesting changes; tags and social links on the profile; a short listing profile with hours switches. [public]
- Attraction edit suggestions, nearby sights, activities and photo credits. [public]
- The privacy policy page `/iranyelvek/adatvedelem`. [public]
- The shared network footer and Belépés dialog. [public]
- `/api/health`, which pings Postgres.
- A runnable schema (`backend/schema/`, `scripts/db-bootstrap.sh`, `scripts/db-dump-schema.sh`); the directory and event catalogs ship in the reference dump; `backend/schema/schema_test.go` fails when boot DDL is missing from the dump.
- CI runs the frontend and backend suites against a Postgres service, plus a `ci-status` job that keeps one "CI is red on main" issue.
- `scripts/start-lamsza-network.sh` under version control, with a test.
- `scripts/sync-shared-frontend.sh` and a hash manifest for the 17 modules shared with lamsza-admin; `scripts/sync-cursor-rules.sh` for the network rules.
- `docs/network/`: ways of working (R1-R17), open items, runtime versions.

### Changed
- The admin UI and API moved to the separate lamsza-admin app; local ports renumbered (lamsza: backend 3001, frontend 5174).
- Search results are one compact row per hit. [public]
- Form dialogs are about 70vw wide; notices stay narrow. [public]
- The news page serves a stale cache at once and refreshes it in the background, and the cache is warmed at startup (no more 5-8 s waits). [public]
- Request body limits, server timeouts and a database pool cap.
- Go test suites run only against a `_test` scratch database (`npm run test:go`); an exported variable beats `.env`.
- CI runs on `ubuntu-24.04` with Node from `.nvmrc` (24) and Go from `go.mod` (1.25.7).
- The one-time browser preferences import runs only where the API reports `prefs_imported_at`.
- `npm run build` refuses to build when no backend answers on 3001; prerender no longer calls the backend.

### Fixed
- Retired category URLs show the error page instead of an empty category. [public]
- A tag-query error no longer panics the entry handler.
- `/esemenyek/[id]` no longer fetches during server rendering.
- Csíkszereda weather lookup.

### Security
- CORS answers only allowlisted origins; Markdown and admin-written page HTML are sanitized; a Content-Security-Policy is sent. [public]
- The weather API key is no longer in the public bundle.
- Media folders no longer list their files.
- `EntriesHandler` no longer logs SQL and search queries.

### Removed
- The `/admin` route, the `/api/admin/*` routes and the frontend modules only they used (now in lamsza-admin).
- `scripts/restart_all.sh` and `npm run restart` (they stopped the whole network and the shared database), committed binaries, March one-off scripts; old planning docs moved to `docs/history/`.

---

## [1.2.0] - 2026-09-24

### Added
- Events calendar with venues.
- Cities, villages, counties, and historical seats.
- Google sign-in, the Fiók page, and user settings.
- Favorite places.
- A saved settlement: when set, homepage weather and the events ticker use that settlement.
- Claimed listings can be edited, including photos, opening hours, and ratings.
- Website submission, and a separate Weboldalak list on the index.
- Search can filter services and websites.

---

## [1.1.0] - 2026-03-06

### Added
- **Advanced Search Relevance**: Implemented weighted Full-Text Search using PostgreSQL GIN indexes and `ts_rank_cd`. Results are prioritized by Name, Location (Multilingual), Category, and tags.
- **Service Management Workflow**: Standardized scripts (`scripts/restart_all.sh`) and agent workflows for starting, stopping, and restarting services with background PID tracking and logging.
- **Active Navigation State**: Header toolbar now dynamically highlights the active button and its sub-pages using SvelteKit's `$page` store.
- **Nomenclature Standardization**: Completed a system-wide transition from "Service" to "Entry" across Database, API, and Frontend layers.
- **Improved Reliability**: Robust `.env` loader ensuring configuration consistency regardless of the execution directory.

### Changed
- Refactored frontend pages to use centralized `$lib/api.js` and extracted reusable UI components.
- Simplified CSS by consolidating component-local styles into `global.css`.

---

## [1.0.0] - 2026-03-01

### Added
- Full SvelteKit + Go + PostgreSQL stack
- Local directory search (`/api/directory?q=...`) with Go API and Postgres
- Weather widget (proxied via Go, cached 30 min, OpenWeatherMap)
- RSS news widget (proxied via Go, cached, multiple feeds from DB)
- Quick links grid (admin-manageable, custom bg colors, SVG icons)
- Random mondas on homepage (fetched from Postgres, admin-addable)
- Unified homepage search: directory (green), weather (blue), news (orange)
- Fallback: no local results shows message + external engine links (Google/Bing/DuckDuckGo/Yandex)
- Directory page `/szolgaltatasok` with dynamic category tabs, URL-based filtering
- Admin panel `/admin` with tabs: Quick Links, News Feeds, Local DB, Service Categories, Mondasok
- Unified admin button styling: dark blue submit, orange logout, red delete
- Light / Dark / System theme toggle with localStorage persistence and FOUC prevention
- Changelog page `/valtozasnaplo` (user-facing, Hungarian)
- `changelog.md` (this file, for developers)
- Footer with version + changelog link
- Mobile-first responsive layout

### Changed
- Replaced static JSON data with live Postgres queries
- Search results are not rendered in DOM until search is submitted (Svelte {#if})
- News teaser: 2 freshest items per source partner instead of global top 10
- Search bar: single "Na lamsza!" button, external engine links in results only

### Fixed
- FOUC (Flash of Unstyled Content) on theme load via blocking inline script in app.html
- RSS feed errors due to redirects and user-agent handling in Go proxy
- News feed display inconsistency between homepage and /hirek

---

## [0.9.5] - 2026-02-28

### Added
- Basic layout: header, greeting, search bar, mondas section
- Initial Svelte components and routing
- Mobile responsiveness

---

## [0.9.0] - 2026-02-20

### Added
- Project initialized: SvelteKit + adapter-node
- global.css with Szekely color palette (szekely-red, szekely-brown, szekely-green)
- global.js utilities stub
- Go backend skeleton with Gin router

---

## [0.8.0] - 2026-02-12

### Added
- PostgreSQL schema: services, service_categories, mondasok, quick_links, news_feeds
- Initial Go API endpoints (CRUD stubs)
- CSP header configuration in hooks.server.ts

---

## [0.7.0] - 2026-02-01

### Added
- Project kickoff
- DigitalOcean droplet setup (Ubuntu 24.04 LTS)
- Nginx configuration
- Initial domain setup: lamsza.com
