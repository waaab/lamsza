# Antigravity Development Plan – Székely Gugel v1.0

## Strict Rules (repeat in every major prompt)
- Use only SvelteKit + vanilla CSS + vanilla JavaScript
- No Tailwind, no UnoCSS, no third-party CSS/JS libraries
- All CSS must be in shared global stylesheets (e.g. `global.css` and `admin.css`, or files imported from them) – NEVER duplicate styles in components
- Reuse shared JS utilities from global.js when possible
- Keep bundle small, optimize for low server cost
- Mobile-first responsive design
- No inline styles – always use classes from global.css
- No em dashes in content – use standard hyphens only

## Phases (updated)

### Phase 0 – Initialization (done)
1. Create SvelteKit project (adapter-node), set up global.css & global.js

### Phase 1 – Layout & Core (done)
2. Build root +layout.svelte with header, background, typography
3. Create +page.svelte with search bar + external buttons
4. Add random mondás component (static JSON array)
5. Ensure full responsiveness

### Phase 2 – Data & Search (Postgres + Admin foundation) (done)
6. Set up PostgreSQL schema (tables for entries, entry_categories, entry_types, mondasok, quick_links, news_feeds, tags, entry_tags)
7. Implement Go API endpoints (CRUD for all tables, joined queries for public entries)
8. Create admin panel (/admin) with CRUD for all tables, unified styling, and news refresh
9. Update homepage search to fetch from local DB, integrate weather/news
10. Wire external buttons to open new tabs with custom background colors

### Phase 3 – Widgets (done)
9. Weather widget (proxied via backend)
10. RSS news widget (proxied via backend)
11. Quick links grid

### Phase 4 – Directory Page & Unified Search (done)
12. Create directory routes with dynamic category tabs and URLs (e.g. `/index`, `/index/[category]`, `/index/szolgaltatasok`)
13. Implement real search on homepage → local DB first, weather/news integration
14. Results highlighted by type (green/directory, blue/weather, orange/news)
15. No directory matches → fallback message + external buttons

### Phase 5 – Admin Panel & Quote Management (done)
16. Create /admin route (protected)
17. Add sections for quick links, news feeds (with refresh/timestamps), local DB CRUD, service categories, mondások
18. Unified Admin Styling: Dark blue for submit, orange for logout, red for delete.

### Phase 6 – Theme (Light/Dark Mode)
18. Add toggle button + localStorage persistence
19. Use CSS variables + data-theme attribute

### Phase 7 – Changelog & Versioning
20. Add footer version link to /valtozasnaplo
21. Create user-facing /valtozasnaplo page
22. Maintain technical changelog.md (developers only)

### Phase 8 – Polish & Extension
23. Bundle size & perf check
24. Add footer, meta tags, error states
25. Create manifest.json for extension
26. Set up strict environment variables (.env) for configuration (done)
27. Create GitHub Actions for automated build testing (CI) (done)