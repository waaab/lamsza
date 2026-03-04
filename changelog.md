# Changelog

All notable changes to this project will be documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

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
