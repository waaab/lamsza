# Changelog

All notable changes to lamsza are recorded here, in English, for developers.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

Every notable change adds a line under `[Unreleased]` in the same commit. A line
users would notice ends with `[public]`; at a release those lines are rewritten
in plain Hungarian in `src/lib/publicChangelog.js`, which `/valtozasnaplo` and
the footer version read (WAYS_OF_WORKING R17).

---

## [Unreleased]

### Added
- Weather section `/idojaras`: every settlement's weather now, by county; `/idojaras/<település>` with the weather now (feels-like, wind and direction, humidity, rain, pressure, UV, clouds, dew point), the next 48 hours, about 9 days ahead, and today's sunrise, sunset, day length, moonrise, moonset and moon phase; `/idojaras/<település>/archivum` with the archived days from launch, a temperature and rain chart, and each day's hours. Every view credits its source as the licence asks. [public]
- Animated weather icons (UI_BASELINE `wx-icons`): `WeatherSymbol` draws all 41 MET Norway codes day and night, `WeatherGlyph` the measurements with their values; theme tokens only, still under reduced motion. The widgets use them with the "svg" icon style. Review page `/idojaras/ikonok` (noindex). [public]
- Weather backend for the new `/idojaras` section: MET Norway (free for commercial use, CC BY 4.0) is the main source, WeatherAPI.com and OpenWeatherMap the fallbacks, used only when MET has been down for three hours. A background worker (`internal/weather/worker.go`) keeps each settlement's and attraction's forecast in `weather_forecast_cache`, honouring MET's `Expires` and `If-Modified-Since`, and stores today's sun and moon times; every route only reads (R14). New routes: `GET /api/weather/forecast` (now, 48 hours, about 9 days), `/api/weather/archive` (from launch day), `/api/weather/places`. One icon vocabulary, MET's symbol codes, for all three providers, with Hungarian text per code. New env `METNO_USER_AGENT`, `WEATHER_WORKER`.
- Weather archive from launch day: `weather_obs_hourly` (MET's value for each hour) and `weather_daily_archive` (one row per settlement and Bucharest day, folded nightly). MET data only, so the series is one source.
- One Fiók and one toolbar for the public apps (UI_BASELINE `tb-account-menu`, `acc-page`). The toolbar's right side is now two buttons: Belépés (signed out) or the profile button with the user's Google photo, which opens a menu with Fiók, Beállítások and Kijelentkezés; then the apps launcher. The Fiók page is built from shared parts (`AccountPage`, `AccountDetails`, `accountDetails.js`), synced to Szótár and Játszótér, with the same user details in every app. [public]
- Internal account API, `GET /internal/account/profile?google_sub=` (`internal/accountapi`): Szótár's and Játszótér's backends read the user's display name by Google ID. Loopback only, no proxy headers, a Bearer token per caller (`ACCOUNT_SZOTAR_TOKEN`, `ACCOUNT_JATSZOTER_TOKEN`); no token means it is off. Route-by-route guard tests.
- UI_BASELINE `ld-reserve-space`: content that arrives late never pushes what is on screen. OPEN_ITEMS: single sign-on as a feature idea; the account-token and nginx prerender steps for production.
- Apps launcher: a nine-dot button at the right end of the toolbar opens a small panel with Lámsza, Szótár and Játszótér, the current app highlighted; keyboard accessible (UI_BASELINE "tb-apps-launcher"). Shared with Szótár and Játszótér with its app list (`networkApps.js`) and `networkOrigins.js`, and a new shared icon, `apps` (nine dots). [public]
- Shared icon `rosette`: the Székely six-petal rosette, for Tájszórejtvény's decorative grid cells. Synced to every app.
- Nine shared icons for Tájszórejtvény and the game tools, synced to every app: `tajszorejtveny` (an arrow-word grid), `zoom_in`, `zoom_out`, `flag`, `text_size`, and `keyboard`, `sound_on`, `sound_off`, `help`, moved from Játszótér's own drawings (same geometry).
- The shared content-dialog shell (`.link-dialog`) takes the screen's width on phones (up to 768px) instead of 70vw, in every app; UI_BASELINE "dlg-content" records it and the OPEN_ITEMS UI-backlog item is closed. [public]
- WAYS_OF_WORKING R19: "today" for all daily content is a Bucharest day, decided by the server. `backend/internal/clock` (Europe/Bucharest with the zone data compiled in, `Today()`, a replaceable `Now`) and `src/lib/bucharestTime.js` (Bucharest wall-clock times, the Bucharest day, the server's "now" from the API's `Date` header), with tests at midnight in summer and winter and on both clock-change days.
- Two shared icons, `words` and `word-suggestions`: the rovás letter G (𐲍), its strokes traced from the `--font-rovas` glyph; the second adds a checkmark. Synced to every app.
- WAYS_OF_WORKING R18: one admin app for the network (`/`, `/dictionary`, `/games`), reaching Szótár and Játszótér only through their internal admin APIs (loopback, a path outside `/api/`, a shared token). UI_BASELINE, the network CLAUDE.md and OPEN_ITEMS follow it.
- `AppIcon` gains `external`, `inbox`, `rovasfejto` and `szokereso` for the admin sections' sidebars.
- `tests/noEmdash.test.js`, shared with every app: `npm test` (and CI) fails on an em dash in code or UI text; Markdown is exempt (rule D5).
- WAYS_OF_WORKING R13: server-to-server calls go to `http://127.0.0.1:<port>`, set by an env var (owner's decision).
- `ErrorShell.svelte`: the network's one minimal error shell (the Lámsza button, a sub-app's own button, the shared `ErrorPage`), shared with every app. Lámsza's root error page uses it and looks the same.
- One error page for the whole network (`ErrorPage.svelte`): a lantern illustration that follows the theme (`icons/ErrorLantern.svelte`), "Hoppácska!", the error code, and a plain Hungarian title and explanation for 400, 401, 403, 404, 408, 410, 429 and any other error, instead of SvelteKit's English "Not Found". Still noindex. [public]
- A `--warm-light` token (the lantern's glow) in both themes.
- `scripts/tests/sync-shared-frontend.test.sh`, the sync script's own test.

### Changed
- Weather widgets: the forecast link sits in the widget title ("Időjárás · Előrejelzés ›", as the events widget does); attraction widgets link to the attraction's own forecast (`/idojaras/<látnivaló>`, which now reads "Látnivaló" and has no archive link). The 48-hour strip drags with the mouse (`lib/dragScroll.js`); touch swipes natively. [public]
- The home, settlement and county weather widgets link to the forecast, credit the real source with a link (the county box always said "OpenWeatherMap"), and show the animated icon or the matching emoji for the MET symbol. [public]
- `/api/weather` and `/api/weather/county` read the cache instead of calling providers per request; the county route's 15-second fan-out is gone. Both add `symbol` and `temp_max`, and `temp_min` is now today's low instead of the current temperature again (the widget's "6°C / 6°C"). [public]
- Home page hero: the title is larger (`--text-hero` up to 7rem) with -8px letter spacing, the greeting under it is muted and pulled up (-30px), and the search box has 3rem above and below. [public]
- The account menu is the same in every app (UI_BASELINE `tb-account-menu`): the profile button always shows the profile icon, never the Google photo (the photo code path and prop are gone), and a highlighted menu item stays inside the panel's padding (`box-sizing: border-box`; it ran 15px past the right edge). The top-right area is pixel-identical in Lámsza, Szótár and Játszótér, signed in and out, menu open and closed, light, dark and phone width. [public]
- Signed out, the site follows the device's theme: signing out, or a session the server no longer knows, clears the saved theme (`clearAccountTheme` in the shared theme store). Choosing a theme stays on Beállítások, for signed-in users. [public]
- The Fiók tab shows the same rows as the other apps: Nyelv is gone, Megjelenített név is listed, dates are on Bucharest's clock. The display-name hint says that Szótár and Játszótér show the name too. [public]
- No more jumping while pages load (layout shift measured on a slowed API, before → after, phone width where worse): the settlement and attraction page reserves a screen's height with a title and lead placeholder (0.357 → 0.009); the "Szűrők betöltése…" placeholder is exactly as tall as the chips (`.btn-md`, button line-height) on /hirek, /index and the category pages; /hirek opens or closes its sources box before fetching (0.144 → 0); /esemenyek shows its two filter buttons' shape while loading (0.023 → 0.002). [public]
- Synced to the other apps: `global.css` with the `--thin-grey` dark value from 36c3f64 (that commit left the copies behind), the account menu and Fiók page CSS, and the launcher's hover and current colour, which used the undefined `--bg-light` and now use `--hover-bg`.
- Lámsza's toolbar draws its 11 icons (Lámsza, Indexelünk, Hírek, Események, Székek, Megyék, Városok, Falvak, Belépés, Kijelentkezés, Beállítások) with the shared `AppIcon` instead of inline SVGs (UI_BASELINE "ic-system"). Same geometry: before and after screenshots of the toolbar are pixel-identical in light, dark and phone width, signed in and out, active and focused. `AppIcon` also gains `plus`, the admin app's bare plus (stroke 2.25), so the admin app can drop its own `AdminPlusIcon`.
- Removed Lámsza's own mondások (OPEN_ITEMS, Mondások: Szótár's `proverbs` are the only store): `/api/mondasok`, `internal/mondasok` with its boot `Migrate()`, `models.Mondas` and `FEATURE_MONDASOK`. The `mondasok` table is dropped (`backend/migrations/drop_mondasok.sql`, applied on dev; `backend/schema/001_schema.sql` regenerated in the same commit, R3). Its rows were test data; dev's are kept in a dump outside the repo.
- The Napi Székely Mondás on the home page comes from Szótár, the network's only mondás store: the widget fetches `szotarUrl('/api/proverbs/today')` without cookies and without a date (Szótár decides today in Europe/Bucharest, R19). Its markup and CSS are unchanged, pinned by `tests/mondasWidget.test.js`; the CSP's `connect-src` admits Szótár (only `https://szotar.lamsza.com` in a production build, checked in `tests/csp.test.js`). `localCalendarISODate` is gone. Before/after screenshots of both home pages are identical in the widget. [public]
- The shared Mondások icon (`AppIcon` `mondasok`) is Google's filled `format_quote` (Material Symbols, without the outlined style's holes), centered and as wide as the other icons; it was a Georgia „ glyph drawn as `<text>` that sat low and small. Fill and stroke are on the path, so the apps need no CSS for it. Synced to every app.
- Events follow Bucharest's day (R19): the list, its filters and the search hide an event at Bucharest midnight after its last day, not at UTC midnight (until 02:00-03:00 Bucharest time before), and the badges read event dates and times as Bucharest times with the server's clock, so a visitor in another time zone, or with a wrong clock, sees the same status. [public]
- `.github/ci-status-issue.*` (identical in all four repos): the comments no longer claim the GitHub API is unreadable or that branch protection was adopted; no em dashes. The Cursor rule's Irányelvek link says `lamszaUrl()`, never lamsza.com in dev. OPEN_ITEMS gains Szótár's word↔mondás links and the letter-kicker decision.
- An unknown entry, event or historical seat (`/bejegyzes/…`, `/esemenyek/…`, `/szekek/…`) shows the network's error page (404, not indexed) instead of an in-page message on a normal page. [public]
- Dark mode: the search box's divider and the "Mégse" hover in dialogs have dark values; the system-dark setting gets the same search seam colour as the explicit dark theme. [public]
- Every JSON API answers `application/json` (several were sniffed as `text/plain`); error lines stay plain text, as the frontend shows them.
- Pages have one `<title>` (the static one in `app.html` is gone). The extension manifest says 1.3.0 with accents.
- The server-side Szótár search goes to `http://127.0.0.1:3002` locally (R13); `SZOTAR_ORIGIN` sets it.
- Shared `global.css` gains `--hover-bg` and dark `--thin-grey`; a dead `.admin-dialog p` rule left `component-typography.css`. Unused CSS removed; no em dash left in code comments (D5).
- `LOCAL_DEV_CHECKS.md` no longer mints sessions for the owner in the dev database (it broke R9): the session-boundary check runs the two boundary suites on scratch databases. The sign-in, CI and restart notes match the code; recorded results and counts are gone (R10).
- `scripts/sync-shared-frontend.sh` also serves lamsza-szotar and lamsza-jatszoter: a `consumers` map in `shared-frontend-modules.json` says which app gets which module, every app gets its own manifest, and `--check` also reports a stale app manifest. Szótár and Játszótér share the icons, the error page, the sign-in dialog and `global.css`.
- `UI_BASELINE.md`: the close-button rule (`dlg-close-label`: "Bezárás" for a dialog that only shows content, "Mégse" only next to an action) and the shared error page (`err-page`).
- The confirm dialog follows the network baseline (Szótár's): a native dialog with a blurred backdrop, Mégse first and the confirm button on the right. [public]
- A refused Google sign-in says "Belépés sikertelen." instead of the server's reply. [public]
- "Székely Gugel" is now "Lámsza" in page titles, descriptions, the news placeholder and the extension manifest. [public]
- Buttons no longer select their label on a double click.
- `AppIcon.svelte` gains the Játszótér icon and round caps on the gear; it, the confirm, notice and sign-in dialogs and the three base stylesheets are now shared with lamsza-admin under the drift guard.
- `docs/network/UI_BASELINE.md` records the owner's 59 UI decisions; the Cursor UI rule points to it.
- `GoogleSignIn` / `SignInDialog` take an optional `signIn(credential)` so an app can sign in through its own API; `AppIcon` gains `trophy`, `login` and `logout`.
- `AppIcon` gains `list` and `add` (Szótár's Lista and Új szó toolbar buttons).

### Removed
- Open-Meteo as a weather source: its free API is for non-commercial use only, and the network will carry ads. The `weather_provider_default` setting is no longer read; `weather_provider_metno_enabled` joins the two fallback switches.

### Fixed
- The events calendar (`/esemenyek`) opens on Bucharest's current year from the server's clock (R19), not the browser's: set again once the API's first reply has fixed the clock, unless a month or day filter is set.
- Docs follow the admin move (R18): PRODUCTION_SERVER_SETUP §6 drops Szótár's and Játszótér's admin lists and adds `ADMIN_SERVICE_TOKEN` and admin's relay variables (§6.4); the LOCAL_DEV_CHECKS allowlist check lists only admin and lamsza; UI_BASELINE `tb-admin-link` says no app links to admin; the OPEN_ITEMS Szókereső item points to `/games#szokereso`.
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
