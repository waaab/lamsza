# Open items

The network's task list since Paperclip was switched off (WAYS_OF_WORKING R1).
One line per item, tagged with its app. Remove an item when it is done and say
in the commit which item it closes. Items marked **owner** are the owner's to
do or decide; agents do not do them.

Last updated: 2026-10-09 (weather section; Tájszórejtvény content and follow-ups; verification review; GA4 item; UI backlog; R18 production security steps; admin move done; R19 time zone and the Mondások move done; lamsza-admin's default branch is `main` and `extract-admin` is deleted; least-privilege CI tokens item; Mondások page layout; reactions idea; apps launcher; Workflow permissions set in all four repos; UI batch: one icon system, apps launcher, Szótár add-word button done; Lámsza toolbar overflow noted).

## Production and accounts (owner only; agents never touch production)

- **[lamsza] owner: set `METNO_USER_AGENT` in lamsza's production `.env`**
  before the weather section ships: the app name and a contact, e.g.
  `lamsza.com weather (+https://lamsza.com; <contact email>)`. MET Norway's
  terms require an identifying User-Agent with contact details and may block
  requests without one. Locally the default (no email) is enough.

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
  8. **Szótár's games endpoint (Tájszórejtvény), before that game goes live.**
     Generate a third token and set it as `GAMES_SERVICE_TOKEN` in Szótár's
     `.env` and as `SZOTAR_GAMES_TOKEN` in Játszótér's (a different value from
     the admin tokens). The same nginx rule covers it: nothing proxies
     `/internal/`. Check on the server: `curl -s 127.0.0.1:<szotar
     port>/internal/games/proverbs` without the token answers 401, and
     `https://szotar.lamsza.com/internal/games/proverbs` never returns JSON.
  9. **The account API (Fiók, display name), with the account release.** Generate two more tokens:
     Lámsza's `ACCOUNT_SZOTAR_TOKEN` = Szótár's `LAMSZA_ACCOUNT_TOKEN`, and Lámsza's
     `ACCOUNT_JATSZOTER_TOKEN` = Játszótér's `LAMSZA_ACCOUNT_TOKEN`; set `LAMSZA_ACCOUNT_URL` in both to
     `http://127.0.0.1:<lamsza port>` (`PRODUCTION_SERVER_SETUP.md` §6). The same nginx rule covers it:
     nothing proxies `/internal/`. Check on the server: `curl -s 127.0.0.1:<lamsza port>/internal/account/profile?google_sub=x`
     without a token answers 401, and `https://lamsza.com/internal/account/profile` never returns JSON.
- **[network] owner: nginx must serve the prerendered pages.** The Szótár and Játszótér vhosts in
  `PRODUCTION_ENVIRONMENT_NOTES.md` use `index app.html` and `try_files $uri $uri/ /app.html`, so a
  prerendered page (`/lista` → `lista.html`, `/` → `index.html`) is never served: every page starts from the
  empty shell. Use `index index.html` and `try_files $uri $uri.html $uri/ /app.html` (Lámsza's staged vhost
  already has `$uri.html`, but also `index app.html`). Check: `curl -s https://szotar.lamsza.com/lista` contains
  the page title, not only the shell.
- **[network] owner: apply the 13 pending system updates on the droplet**
  (Ubuntu 24.04), as a planned task with a backup (snapshot) and a rollback
  path, ideally together with the Node upgrade.

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

- **[lamsza] owner: enter coordinates for the 27 settlements** in the admin
  app's settlement form ("46.3593, 25.8017"). Until then the weather worker finds each settlement by name
  (OpenWeatherMap's geocoder), which can land a village on a namesake.

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
  checkbox stays off. Today every repo is set to "Read and write" (owner,
  2026-10-08).
  - The repo setting is the owner's (GitHub, Settings > Actions > General).
  - `ci-status` already declares `contents: read` and `issues: write` in all
    four workflows; add a top-level `permissions: contents: read` so the
    other jobs stay read-only whatever the default is.
  - `ci-status` only runs on a push to `main`, so for the branch test its
    condition has to admit that branch for the test, and is reverted after.
    Check that the issue opens, then closes after the next green run.

## Tájszórejtvény (content and follow-ups)

The game is built and pushed, and stays switched off until go-live (plan:
`lamsza-jatszoter/docs/superpowers/plans/2026-10-07-tajszorejtveny.md`, decisions in its spec §16; the owner's
go-live checklist is at the end of the plan). These are the agreed tasks outside that plan.

- **[jatszoter] owner: grow the proverb list toward about 230.** 28 puzzles a week need about 230
  proverbs to avoid a repeat within 8 weeks; today there are 44 fallback proverbs (and Szótár's
  mondások, which count too). Until then a reused proverb is allowed and counted (spec §16.2.10).
- **[jatszoter] owner: mark more words as familiarity 1** in admin `/games` (word metadata). Könnyű uses
  21 well-known words a week and 62-77 fit a grid, so each returns every 3-4 weeks (spec §16.2.11).
- **[jatszoter] owner: grow the filler words** from 453 to several thousand, with plenty of 7-8 letter
  words, not only short ones. Fewer empty cells and more variety (spec §16.2.3).
- **[jatszoter] owner: review the word metadata** (familiarity, clue overrides, block list) and the 78
  example sentences written for the concept; until one is marked reviewed it is not used (spec §16.2.7).
- **[szotar] owner: fix the definition typos the concept found**, in admin `/dictionary`: katulya
  "Dobpz", pityóka "Burgyona", laska "nyujtott", bikfic "fiu", igír "Igér", szakajtó "Szakitó", mihót
  "Miota", nyiszitel "Vág,rossz", tanyitt "Tanit", cseszle "Szunyog", összegurucsálódik "Összegyürödik",
  kalorifer "Fütőtest", garzon "Egszobás"; megcsemelettem is defined as "Kályha" (misaligned).
- **[network] No email sender.** Tájszórejtvény's review reminders are in-app only (the admin
  dashboard). Add a sender if the owner wants email.
- **[jatszoter] Összesített ranglista with placement points per game** (100 × (players − rank + 1) /
  players), with Hét/Hónap/Év/Összes tabs; it also fixes Szórejtő missing from the overall board and
  Szókereső's all-time sum. Caveat accepted by the owner: a game with few players weighs the same as a
  busy one. Tájszórejtvény joins the overall board only then.
- **[jatszoter] Tájszórejtvény admin v1.1:** swap a single word (partial re-fill), push a report's fix
  to Szótár, a spelling/consistency check on clues.
- **[jatszoter] Move the other games' tile drawings** from `GameIcon.svelte` into the shared `AppIcon`
  (UI_BASELINE ic-system); Tájszórejtvény's tile already uses `AppIcon`.
- **[network] owner: decide a network text-size setting.** Tájszórejtvény brings a Játszótér-only
  Betűméret (normál / nagyobb / legnagyobb); it could become a UI_BASELINE item for every app.
  Owner, 2026-10-08: it stays game-local for now; decide after real players have tried it.

## UI backlog

UI changes the owner has asked for but not scheduled yet. Not implemented; each
follows `UI_BASELINE.md` when it is done.

- **[lamsza] The toolbar does not fit between 391 and about 1225 px.** The
  labels hide only at 768px and below and the row never wraps, so Fiók,
  Kijelentkezés and Beállítások are cut off on phones, tablets and small
  laptops; with the apps launcher the signed-in limit is 1280px. Owner,
  2026-10-08: keep it as it is for now; the toolbar changes before go-live.
- **[szotar] The Mondások page follows the default page layout.** Align
  `/mondasok` with the network's standard structure (title, lead line, main
  area with content and sidebar), the same way the other Szótár list pages
  and Lámsza's list pages are built. Today it is inside `SidebarLayout` but
  has only a title: no lead line under it, and `page-lead` is used for the
  "Elérted a lista végét." line at the bottom instead.

## Feature ideas

Ideas the owner wants kept, not scheduled. Each needs a plan before any work.

- **[lamsza] Erdélyi Hírek redesign** (the owner's design artifact
  ESGiDrAtMHEyqYYXRzxANz). Deferred by the owner on 2026-10-09 after the other
  redesign phases (home pages, Székelyföld pages) shipped. Its regions, topics,
  read counts and story clustering need backend work first.
- **[network] Single sign-on across *.lamsza.com.** Today each app has its own Google OAuth client and its own
  sign-in; users are matched across apps by Google ID, Lámsza's display name is copied over the internal
  account API, and each app stores its own theme (UI_BASELINE `acc-page`, `set-theme-store`). One sign-in for
  the whole network (a shared session on `.lamsza.com`, or one auth service) would make that copying and the
  per-app theme unnecessary. Needs a plan: cookie domain and SameSite, the four OAuth clients, the admin app's
  separate session, CSRF.

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
