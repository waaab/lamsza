# Open items

The network's task list since Paperclip was switched off (WAYS_OF_WORKING R1).
One line per item, tagged with its app. Remove an item when it is done and say
in the commit which item it closes. Items marked **owner** are the owner's to
do or decide; agents do not do them.

Last updated: 2026-10-07 (network review; GA4 item added).

## Production and accounts (owner only; agents never touch production)

- **[network] owner: upgrade production Node 20.20.2 (end of life) to Node 24 LTS.**
  A planned task with a backup and a rollback path. Every app already builds
  and tests on Node 24 (`VERSIONS.md`). After it, nothing else needs to change:
  CI and `.nvmrc` already say 24.
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

- **[admin] server timeouts.** lamsza got read, write and idle timeouts (BOG-18);
  admin still uses a bare `http.ListenAndServe`.
- **[szotar] [jatszoter] Content-Security-Policy.** lamsza and admin send one;
  szotar and jatszoter do not.
- **[jatszoter] archive before the first puzzle** answers "Játék hiba." instead
  of "no puzzle for that day".
- **[jatszoter] Szórejtő plan Task 13, step 5** (share copy to the clipboard)
  needs a browser check.
- **[lamsza] Test claim and membership decisions** (planned in
  `docs/superpowers/plans/2026-09-27-listing-claim-membership.md`, never
  written): the `claim_pending` 409 and member accept/deny have no test.
