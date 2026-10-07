# Open items

The network's task list since Paperclip was switched off (WAYS_OF_WORKING R1).
One line per item, tagged with its app. Remove an item when it is done and say
in the commit which item it closes. Items marked **owner** are the owner's to
do or decide; agents do not do them.

Last updated: 2026-10-07 (verification review; GA4 item; UI backlog).

## Production and accounts (owner only; agents never touch production)

- **[network] owner: upgrade production Node 20.20.2 (end of life) to Node 24 LTS.**
  A planned task with a backup and a rollback path. Every app already builds
  and tests on Node 24 (`VERSIONS.md`). After it, nothing else needs to change:
  CI and `.nvmrc` already say 24. While on the server, verify the production
  ports: `PRODUCTION_SERVER_SETUP.md` (lamsza 3000, szotar 3010, jatszoter 3001)
  and `PRODUCTION_ENVIRONMENT_NOTES.md` (8081, 8082, 8083) disagree. Correct both
  docs to what the server actually runs, and set lamsza's `SZOTAR_ORIGIN` to
  `http://127.0.0.1:<szotar port>` there (R13: server-to-server calls stay on
  the machine).
- **[network] owner: apply the 13 pending system updates on the droplet**
  (Ubuntu 24.04), as a planned task with a backup (snapshot) and a rollback
  path, ideally together with the Node upgrade.
- **[network] owner: set *Workflow permissions* to "Read and write" in all four
  repos** (GitHub, Settings > Actions > General). Approved on BOG-57; without it
  the `ci-status` job cannot open the "CI is red on main" issue (WAYS_OF_WORKING
  R7).
- **[admin] owner: set the default branch of `waaab/lamsza-admin` to `main`.**
  It is `extract-admin`, which holds lamsza's history, so the repo's front page,
  new pull requests and fresh clones point at the wrong code. Then delete
  `extract-admin` (`git push origin --delete extract-admin`; GitHub refuses
  while it is the default).

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
  Its daily has been empty since; puzzles are published by hand in
  `/admin/szokereso`.
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

- **[admin] [szotar] [jatszoter] server timeouts.** lamsza got read, write and
  idle timeouts (BOG-18); the other three backends still use a bare
  `http.ListenAndServe`.
- **[admin] CI runs no Postgres.** Admin's workflow has no `postgres:16` service,
  so its DB-bound tests (about 44) skip in CI and run only locally
  (`npm run test:backend`). lamsza, szotar and jatszoter run theirs in CI.
- **[lamsza] [admin] test the `site_settings` contract.** Admin writes the
  `weather_provider_*` and `social_*_url` keys that lamsza reads; no test pins
  the key names on either side.
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
- **[lamsza] [admin] One icon system everywhere** (UI_BASELINE "ic-system").
  Lámsza's own toolbar (`src/routes/(public)/+layout.svelte`) still draws 11
  inline SVGs that `AppIcon` already has, and admin's `AdminPlusIcon` duplicates
  `AppIcon`'s `add`. Switch them to `AppIcon`.
