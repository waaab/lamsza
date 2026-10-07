# Open items

The network's task list since Paperclip was switched off (WAYS_OF_WORKING R1).
One line per item, tagged with its app. Remove an item when it is done and say
in the commit which item it closes. Items marked **owner** are the owner's to
do or decide; agents do not do them.

Last updated: 2026-10-07 (verification review; GA4 item; UI backlog; R18 production security steps; admin move done; R19 time zone and the Mondások move done; lamsza-admin's default branch is `main` and `extract-admin` is deleted; least-privilege CI tokens item; Mondások page layout; reactions idea; apps launcher).

## Production and accounts (owner only; agents never touch production)

- **[network] owner: production upgrade: Node 20.20.2 (end of life) to Node 24 LTS, ports, and the admin app's internal API (R18).**
  A planned task with a backup and a rollback path. Every app already builds
  and tests on Node 24 (`VERSIONS.md`). After it, nothing else needs to change:
  CI and `.nvmrc` already say 24. While on the server, verify the production
  ports: `PRODUCTION_SERVER_SETUP.md` (lamsza 3000, szotar 3010, jatszoter 3001)
  and `PRODUCTION_ENVIRONMENT_NOTES.md` (8081, 8082, 8083) disagree. Correct both
  docs to what the server actually runs, and set lamsza's `SZOTAR_ORIGIN` to
  `http://127.0.0.1:<szotar port>` there (R13: server-to-server calls stay on
  the machine).

  **Also on this visit: production setup for the one admin app (R18)**, before
  or with the deploy of the admin move. Szótár's and Játszótér's internal admin
  API (`/internal/admin/`) stays closed unless all of this holds, so do it in
  this order and check each step on the server:
  1. **Tokens, never in git.** Generate two long random tokens (for example
     `openssl rand -hex 32`). Set `ADMIN_SERVICE_TOKEN` in Szótár's `.env` and
     the same value as `SZOTAR_ADMIN_TOKEN` in admin's `.env`; likewise
     Játszótér's `ADMIN_SERVICE_TOKEN` = admin's `JATSZOTER_ADMIN_TOKEN`. Use a
     different token per app. Without a token an app's internal API answers 404
     to everything (fail closed), and admin's `/dictionary` or `/games` says
     "nincs beállítva".
  2. **Server-to-server on the machine.** Set `SZOTAR_ADMIN_URL` and
     `JATSZOTER_ADMIN_URL` in admin's `.env` to `http://127.0.0.1:<port>` of
     each app's backend (the real ports, checked above). The apps
     accept the internal API only from loopback.
  3. **nginx must not expose `/internal/`.** Every vhost (szotar, jatszoter,
     admin, lamsza) proxies only `location /api/` to a backend; nothing proxies
     `/` or `/internal/`. Each `/api/` block keeps `proxy_set_header X-Real-IP`
     and `X-Forwarded-For` (as in `PRODUCTION_ENVIRONMENT_NOTES.md`): the apps
     refuse an internal call that carries either header, which is what stops a
     request that came through nginx even though nginx connects from
     127.0.0.1. Check: `curl -i https://szotar.lamsza.com/internal/admin/stats`
     and the same on jatszoter must not reach the backend (the SPA page or a
     404, never JSON), and `https://<host>/api/internal/admin/stats` must be 404.
  4. **Backend ports closed from outside.** The backends listen on all
     interfaces (`:<port>`); confirm the firewall (ufw / DigitalOcean) only
     opens 22, 80 and 443, so the backend ports are not reachable from the
     internet. (The apps refuse non-loopback internal calls anyway; this is the
     second lock.)
  5. **Smoke test after deploy, GET only:** admin `/dictionary` and `/games`
     show their dashboards with counts; on the server,
     `curl -s 127.0.0.1:<szotar port>/internal/admin/stats` without the token
     answers 401.
  6. **The old admin settings go:** remove `ADMIN_GOOGLE_EMAILS` from Szótár's
     `.env` and `ADMIN_EMAILS` from Játszótér's (neither app reads them any
     more). Keep admin's `CORS_ALLOWED_ORIGINS=https://admin.lamsza.com`: no
     public app calls the admin app. Check: `https://szotar.lamsza.com/admin`
     and `https://jatszoter.lamsza.com/admin` show the 404 error page.
  7. **Mondások live only in Szótár (WAYS_OF_WORKING R18; done on dev
     2026-10-07).** Production's mondások are test data too: `pg_dump`
     lamsza's `mondasok` and Szótár's `proverbs` (keep the dumps off the
     server's repos), then delete every row in both. Szótár's
     `CORS_ALLOWED_ORIGINS` must include `https://lamsza.com` and
     `https://www.lamsza.com`: Lámsza's home page reads the daily mondás from
     `https://szotar.lamsza.com/api/proverbs/today` in the browser. Lámsza's
     production build needs no setting (its CSP already lists
     `https://szotar.lamsza.com`). Once lamsza and admin without the mondás
     code are deployed, apply `backend/migrations/drop_mondasok.sql`.
- **[network] owner: apply the 13 pending system updates on the droplet**
  (Ubuntu 24.04), as a planned task with a backup (snapshot) and a rollback
  path, ideally together with the Node upgrade.
- **[network] owner: set *Workflow permissions* to "Read and write" in all four
  repos** (GitHub, Settings > Actions > General). Approved on BOG-57; without it
  the `ci-status` job cannot open the "CI is red on main" issue (WAYS_OF_WORKING
  R7).

## Decisions

- **[lamsza] owner: review branch `review/bog-32-seo-consent`** (pushed to origin
  as a backup on 2026-10-07, not merged): the parked BOG-32 SEO, structured
  data, sitemap and cookie-consent work, rebuilt as one commit on its original
  base. It carries that base's old build-only workflow, so its CI run builds
  but runs no tests. Decide whether to rebase and
  finish it or drop it. A rebase onto main conflicts only in `package.json`.
- **[jatszoter] owner: Kaptár needs a bigger word list.** The 475-word Székely
  dictionary supports no board with 10 playable words (best: 5), so no boards
  were seeded and most daily boards repeat. `cmd/seed-kaptar-boards` is ready
  for when the word list grows.
- **[jatszoter] owner: Szókereső has had one published puzzle (2026-09-28).**
  Its daily has been empty since; puzzles are published by hand in the
  admin app's `/games#szokereso`.
- **[jatszoter] `DifficultyPicker.svelte` is unused.** Wire it into the game
  shell or delete it.
- **[network] owner: GA4 analytics, not started; nothing is implemented.**
  Order: first review `review/bog-32-seo-consent` (above), because GA4 may load
  only after cookie consent and that branch holds the consent work. Then decide:
  one GA4 property with cross-subdomain measurement for lamsza.com and its
  subdomains, or one property per app. When GA4 is integrated, the shared
  `ErrorPage.svelte` (lamsza, synced to every app) must send a custom event,
  e.g. `error_page` with `code`, `path`, `referrer` and `app`, on client-side
  navigations too: an error page keeps the original URL, so without the event it
  counts as a normal page view.

## Work

- **[lamsza] The browser extension (`extension/`, git-ignored, built by
  `npm run build:extension`) is not used.** Its March build calls
  `/api/admin/mondasok` on `localhost:3000` and no longer works. A future
  extension must read the daily mondás from Szótár's public API
  (`https://szotar.lamsza.com/api/proverbs?date=`), for example with
  `host_permissions` for Szótár in `manifest.json`; Lámsza has no mondás
  endpoint once the item above is done.
- **[admin] [szotar] [jatszoter] server timeouts.** lamsza got read, write and
  idle timeouts (BOG-18); the other three backends still use a bare
  `http.ListenAndServe`.
- **[admin] CI runs no Postgres.** Admin's workflow has no `postgres:16` service,
  so its DB-bound tests (about 44) skip in CI and run only locally
  (`npm run test:backend`). lamsza, szotar and jatszoter run theirs in CI.
- **[lamsza] [admin] test the `site_settings` contract.** Admin writes the
  `weather_provider_*` and `social_*_url` keys that lamsza reads; no test pins
  the key names on either side.
- **[admin] Type-check errors, to work through gradually.** `npx svelte-check`
  in `frontend/` reports 793 errors in 34 files (2026-10-07). They predate the
  verification fixes: 786 were there before, and the 7 that the shared
  `tests/noEmdash.test.js` adds are the same kind as the rest of the test files
  (no Node type definitions, so `node:` imports do not resolve). Most are in
  `src/routes/+page.svelte` (about 630). Start with the cheap, wide fix (Node
  types for the tests), then the page; do not let the count grow.
- **[lamsza] Catalog seeds burn sequence numbers on every backend start.** The
  boot-time seeds insert with `ON CONFLICT DO NOTHING`, which takes a sequence
  value even when the row already exists, so each start of the lamsza backend
  moves `pages`, `page_faq_sections`, `historical_seats`,
  `catalog_event_types`/`_subtypes`, `settlement_location_types` and `websites`
  ahead (16, 16, 5, 5/16, 5 and 1 on 2026-10-07) with no row change. Seeds in
  `internal/pages/pages.go`, `internal/pagefaq/pagefaq.go`,
  `internal/events/migrate.go`, `internal/handlers/settlement_location_types.go`,
  `internal/account/websites_migrate.go` and `internal/db/seed_historical_seats.go`. Make them
  not burn IDs (insert only `WHERE NOT EXISTS`, or seed once from
  `backend/schema/002_reference.sql`), with a test that a second boot leaves the
  sequences where they were.
- **[szotar] Word and mondás links** (moved from `lamsza-szotar/docs/tasks.md`).
  The word page shows Példamondat and "Székely mondás ezzel a szóval" per sense,
  and the first sense lists mondások whose text mentions the headword (a text
  search on `/api/proverbs`). Still missing: explicit word↔mondás links stored
  in the data, and navigation both ways (from a word in a mondás to its entry).
- **[lamsza] two orphan images** in `backend/data/entry-images/`
  (`990e305eead49770.png`, `verify-manifesto.png`, 2026-09-22): untracked and
  referenced by no row. Delete them or find their owner.
- **[szotar] [jatszoter] Content-Security-Policy.** lamsza and admin send one;
  szotar and jatszoter do not.
- **[jatszoter] archive before the first puzzle** answers "Játék hiba." instead
  of "no puzzle for that day".
- **[jatszoter] Szórejtő plan Task 13, step 5** (share copy to the clipboard)
  needs a browser check.
- **[lamsza] Test claim and membership decisions** (planned in
  `docs/superpowers/plans/2026-09-27-listing-claim-membership.md`, never
  written): the `claim_pending` 409 and member accept/deny have no test.
- **[network] Consider least-privilege CI tokens.** Switch the repos' default
  workflow permissions to read-only and grant `issues: write` only to the
  `ci-status` job in the workflow YAML, then prove the "CI is red on main"
  issue still opens (for example with a deliberately failing test on a
  branch). The "Allow GitHub Actions to create and approve pull requests"
  checkbox stays off. Replaces the owner's "Read and write" item above if
  adopted.
  - The repo setting is the owner's (GitHub, Settings > Actions > General).
  - `ci-status` already declares `contents: read` and `issues: write` in all
    four workflows; add a top-level `permissions: contents: read` so the
    other jobs stay read-only whatever the default is.
  - `ci-status` only runs on a push to `main`, so for the branch test its
    condition has to admit that branch for the test, and is reverted after.
    Check that the issue opens, then closes after the next green run.

## UI backlog

UI changes the owner has asked for but not scheduled yet. Not implemented; each
follows `UI_BASELINE.md` when it is done.

- **[szotar] The add-word button moves from the toolbar to the page title row.**
  Today it is a "+" icon in the top toolbar ("Új szó", signed-in visitors only,
  opens the word dialog). Move it to the far right of the page title row, the
  way Lámsza's Index puts "Új Bejegyzés" there (`index-heading__add` in
  `src/routes/(public)/index/+page.svelte`: a `btn btn-primary btn-lg` with a
  short helper text under it, from `src/lib/indexCreateCopy.js`). Add a helper
  text if one fits Szótár, and remove the add icon from the toolbar. To decide
  when it is done: which pages carry it (the home page at least), and what a
  signed-out visitor sees (the button opening the sign-in dialog, or no button).
- **[network] Content dialogs on phones.** The shared `.link-dialog` shell in
  `global.css` is `min(70vw, 100vw - 2rem)` wide, about 273px on a 390px phone,
  in every app. Admin already overrides it to the screen width on small screens
  (`admin.css`); make that the shell's own small-screen rule in lamsza and sync.
- **[szotar] The letter above the headword.** The word page shows the entry's
  first letter ("cs") as a small link above the headword. The 2026-10-05
  word-entry-polish plan says to remove it; it was never done. Owner to
  confirm: remove it, or keep it and close the plan item.
- **[lamsza] [admin] One icon system everywhere** (UI_BASELINE "ic-system").
  Lámsza's own toolbar (`src/routes/(public)/+layout.svelte`) still draws 11
  inline SVGs that `AppIcon` already has, and admin's `AdminPlusIcon` duplicates
  `AppIcon`'s `add`. Switch them to `AppIcon`.
- **[szotar] The Mondások page follows the default page layout.** Align
  `/mondasok` with the network's standard structure (title, lead line, main
  area with content and sidebar), the same way the other Szótár list pages
  and Lámsza's list pages are built. Today it is inside `SidebarLayout` but
  has only a title: no lead line under it, and `page-lead` is used for the
  "Elérted a lista végét." line at the bottom instead.
- **[lamsza] [szotar] [jatszoter] Apps launcher in the toolbar.** A new button
  at the top right of the toolbar in every public app (Lámsza, Szótár,
  Játszótér; not the admin app).
  - The icon is our own nine-dot grid SVG in the shared icon set (`AppIcon`),
    clearly different from the old admin icon.
  - It opens a small panel with each network app's icon and name, the
    current app highlighted, with links built from `networkOrigins.js`.
  - The app list comes from one shared place, so a new app appears everywhere
    automatically. `networkOrigins.js` is not in
    `shared-frontend-modules.json` today (Szótár and Játszótér keep their own
    copies), so the list and the origins likely join the shared modules.
  - One shared component, keyboard accessible, works on phones.
  - Check it against UI_BASELINE `tb-cross-links` (the sister-app links each
    toolbar has today) when planning.

## Feature ideas

Ideas the owner wants kept, not scheduled. Each needs a plan before any work.

- **[network] Reactions, starting with Mondások**, inspired by IMDb's reaction
  bar: thumbs up and thumbs down with counts, and a smiley button that opens
  a small set of emoji reactions with counts.
  - Only signed-in users can react. Signed-out visitors see the counts, and
    clicking opens the sign-in dialog.
  - Our own SVG icons for thumbs up, thumbs down and the smiley, in the
    shared icon set (`AppIcon`).
  - One reusable shared component, so it can later be used for words, games
    and other content across the network.
  - To decide when planning: one or several reactions per user, toggling a
    reaction off, the emoji set, and native emojis or our own SVGs (native
    emojis look different on each OS).
