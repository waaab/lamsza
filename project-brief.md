# Székely Gugel Revival – Project Brief (v1.0)

## Goal
Revive lamsza.com as a fast, culturally authentic new-tab extension + web page for Hungarian speakers in Székelyföld / Transylvania.  
Primary focus: a real, browseable local directory of services (doctors, schools, craftsmen, offices, etc.) with quick dashboard features (search, weather, news).

## Target Users
Hungarian speakers in Hargita, Kovászna, and eastern Maros counties + diaspora.  
Primary use case: every new browser tab + browsing local services.

## Core Principles (must be followed strictly)
- Fast, simple, reliable
- Minimal server costs
- 100% vanilla CSS + vanilla JavaScript (no Tailwind, no libraries)
- All styling in one global.css file – no component-local CSS duplication
- Reuse global.js utilities when possible
- Small bundle size
- Works perfectly as Chrome/Edge/Firefox new-tab override

## Tech Stack (locked)
- Frontend: SvelteKit + vanilla CSS + vanilla JS
- Backend: Go 1.25.7+ + PostgreSQL
- Web server: Nginx
- Hosting: DigitalOcean Droplet (Ubuntu 24.04 LTS)
- CI/CD: GitHub Actions (Build & Test only)
- Extension: Manifest V3

## Versioning & Scope

**v1.0 – MVP (Target: 3–4 weeks) – Anonymous only**
- Székely-themed layout (light/dark mode toggle)
- Central search bar → primarily searches local directory (in-place results in #results)
  - No directory matches → fallback message + buttons for Google/Bing/Duck/Yandex
- Dedicated directory page: /szolgaltatasok
  - Dynamic category tabs/filters (admin-manageable via Service Categories)
  - URL updates: /szolgaltatasok/egeszsegugy, /szolgaltatasok/oktatas, etc.
  - Full browseable list with cards (name, location, phone, address, notes, and optional URL)
- Random daily Székely mondás (from DB, admin-addable)
- Weather widget (proxied via backend, cached 30 min, timestamp, skeleton)
- RSS news feed (proxied via backend, cached, with manual refresh and timestamps in admin)
- Quick links grid (admin-addable, custom background colors)
- Footer with version link to /valtozasnaplo
- Fully responsive (mobile-first)

**Search Engine & Directory Design**
- Primary search ("Keresés a Gugelben") searches the local directory first → results in #results on homepage
- Integrated weather and news in search:
  - Weather: if query contains "idő", "időjárás", "milyen", "hőmérséklet" + city name (e.g. "milyen az ido Csikba?") → show in blue section
  - News: keyword search in titles/descriptions → show in orange section
- Highlighting: Different colors for directory (green), weather (blue), news (orange)
- Backend: PostgreSQL + Go API for dynamic queries (/api/directory?q=...&category=...)

**Admin Panel Design**
- Protected page: /admin (basic password or Google Sign-In in v1.2)
- Tabs/sections:
  - Quick Links (add/edit/delete)
  - News Feeds (add/edit RSS URLs + names, manual refresh trigger, last update timestamp)
  - Local Database (CRUD for services: name, URL, location, phone, address, notes, linked to Service Categories)
  - Service Categories (add/edit/delete categories for the directory)
  - Mondások (add/edit new quotes)
- Unified Styling: Dark blue for submit/login, orange for logout, red for delete.
- Vanilla forms + tables

**Changelog & Versioning**
- User-facing changelog page: /valtozasnaplo (simple list of versions + changes in Hungarian)
- Technical changelog.md (in root or docs/, for developers only – not public)
- Footer shows current version with link to changelog (e.g. v1.0.0 | Változásnapló)

**Non-Goals for v1.0**
- Google Sign-In / user accounts (defer to v1.2)
- Custom user links / saved favorites
- Dark mode toggle (defer to v1.2)

**Success Criteria for v1.0**
- Extension overrides new tab correctly
- Directory page loads and filters by category
- Homepage search returns real directory results (or fallback with weather/news)
- Page loads under 1.5 seconds
- All widgets work reliably
- Visual identity feels proudly Székely