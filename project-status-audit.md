# Project Status Audit

**Current check:** 24 September 2026. The March audit below is a historical snapshot of the v1.1.0 modularization work. It is not the product status.

## Current product (v1.2.0)

Shipped and recorded in `src/lib/publicChangelog.js` and `changelog.md`:

- Events calendar, venues, cities, villages, counties, and historical seats.
- Google sign-in, Fiók, and user settings (theme, quick links, one saved settlement).
- Favorite places. They do not set homepage weather.
- Homepage weather and the events ticker follow the saved settlement when one is set.
- Claimed listings: edit, photos, hours, and ratings for the owner.
- Directory websites, with a Weboldalak list on the index.
- Search can filter services and websites.

The footer version is the first entry of `src/lib/publicChangelog.js`. It is not a separate hardcoded string.

`/api/admin/*` requires an admin Google session. `/api/proxy` only fetches crests, attraction images, and news-feed URLs already stored in the database, and it refuses private addresses. Dated specs under `docs/superpowers/` were not rewritten; they are the design records for the features above. A public launch still needs the remaining go-live checks (consent, content pages, SEO, schema, accessibility).

---

**Date of the audit below:** March 2026

This section audits the project against `project-brief.md` as it stood for v1.1.0: modularity, feature toggling, and architecture.

---

## 1. Phase 6 (Theme) Verification
**Status: ✅ Fully Implemented**

- **Dark Mode Implementation:** The theme uses CSS variables from `global.css` and toggles a `data-theme` attribute (`data-theme="dark"` or `data-theme="light"`) on the `<html>` element.
- **Persistence & FOUC:** Preference saved in `localStorage`, with render-blocking script in `<head>` to prevent FOUC.

---

## 2. Phase 7 (Versioning & Changelog) Verification
**Status: ✅ Fully Implemented**

- **Public Route:** The `/valtozasnaplo` route renders an accordion-based user-friendly changelog.
- **Technical Changelog:** The `changelog.md` file resides in the project root with detailed technical updates.

---

## 3. Phase 8 (Modularization & Reliability) Verification
**Status: ✅ Fully Implemented**

- **Package Architecture:** The backend has been completely refactored from a monolithic `main.go` into a modular package structure (`internal/config`, `db`, `models`, `handlers`, `news`, `events`, `weather`, `mondasok`, `links`, `search`).
- **Feature Toggling:** A lightweight configuration-based system allows toggling optional modules (Weather, Events, News, etc.) via `.env` variables (`FEATURE_WEATHER=false`, etc.).
- **Reliability:** Introduced a unit test suite for core utilities (`main_test.go`) and verified that the modular system builds successfully.
- **Env Awareness:** Implemented a robust `.env` loader that resolves configuration relative to the project root, enabling reliable execution from subdirectories.

---

## 4. Feature Inventory & Standards Audit
**Status: ✅ Implemented within Core Principles**

- **Active Frontend Components:** Search Bar, Weather Widget, News Feed, Quick Links, Mondás, Events.
- **Styling Architecture:** 100% compliant. Zero embedded `<style>` blocks; all centralized in `global.css`.
- **Frontend Modularization:** ✅ Fully Implemented. Pages are thin wrappers; all complex logic is delegated to reusable components and centralized `$lib/` utilities.
- **Backend Proxy Verification:** The Go backend properly proxies external resources, bypassing CORS on the client side.

---

## 5. Admin UI State
**Status: ✅ Functional**

The `/admin` route is active and provides full CRUD for:
1. **Mondások**
2. **Gyorslinkek**
3. **News Feeds**
4. **Települések (Locations)**
5. **Szolgáltatás Kategóriák**
6. **Szolgáltatások (Services)**
7. **Események (Events)**

### Final Review (March 2026)
The v1.1.0 modularization goals in this snapshot were met. Later product work is the 24 September 2026 section at the top of this file. The “zero embedded style blocks” line describes that March pass, not the current Svelte pages.
