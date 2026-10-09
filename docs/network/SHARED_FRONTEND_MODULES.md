# Shared frontend modules: lamsza owns, the apps get generated copies

**Decided on BOG-42, 2026-10-06** for `lamsza-admin`; **extended to
`lamsza-szotar` and `lamsza-jatszoter` on 2026-10-07**, when the owner asked for
one error page for the whole network. This file is the source of truth for the
frontend files that `lamsza` shares with the other apps (the list in
`shared-frontend-modules.json`).

---

## The rule

1. **`lamsza` owns every shared module.** Edit it there, never in an app.
2. **Each consumer app's `frontend/` carries a generated copy of its share.**
   It is a real committed file, not a link or a submodule, so every repo still
   clones, builds and deploys on its own.
3. **Which app gets which module is the `consumers` map** in
   `lamsza/shared-frontend-modules.json`: `"all"` for `lamsza-admin`, a list of
   paths for `lamsza-szotar` and `lamsza-jatszoter`.
4. **`scripts/sync-shared-frontend.sh` in `lamsza` does the copying** and
   rewrites the manifest (path → SHA-256) in `lamsza` and one per app,
   `frontend/shared-frontend-modules.json`, listing only that app's share.
5. **Each repo's own test suite checks its own copies against its own
   manifest.** `tests/sharedFrontendModules.test.js` (itself one of the shared
   files, in every app's share) hashes every listed module and fails on a
   mismatch.

Changing a shared module:

```bash
cd ~/projects/lamsza-network/lamsza
$EDITOR src/lib/entryHours.js
scripts/sync-shared-frontend.sh          # copies + rewrites every manifest
npm test                                 # green here
(cd ../lamsza-admin/frontend && npm test)      # and in each app that got the file
(cd ../lamsza-szotar/frontend && npm test)
(cd ../lamsza-jatszoter/frontend && npm test)
# commit lamsza and every app that changed
```

Adding or removing a shared module: add or delete the path under `modules` in
`lamsza/shared-frontend-modules.json` (any placeholder hash will do), add it to
or drop it from the `consumers` lists that need it, and run the script. It
re-hashes from `lamsza` and writes every manifest. A consumer list that names a
path missing from `modules` is refused before anything is copied.

`scripts/sync-shared-frontend.sh --check` verifies without writing, for when the
repos are checked out side by side: it reports a copy that drifted (`DRIFT`) and
an app manifest that no longer matches lamsza (`STALE`). `LAMSZA_ADMIN_ROOT`,
`LAMSZA_SZOTAR_ROOT` and `LAMSZA_JATSZOTER_ROOT` override where each app lives;
the default is next to `lamsza`. The script's own test is
`scripts/tests/sync-shared-frontend.test.sh`.

## Why this shape and not another

The problem BOG-42 had to solve: 17 byte-identical files in two repos that
`docs/LOCAL_DEV_CHECKS.md` asked people to keep in step **by hand**. Hand-kept
duplication fails quietly — nothing goes red when the two copies diverge.

The constraint that decided it: **drift has to be catchable in single-repo CI.**
Each repo's workflow checks out one repo. A guard that needs both checkouts
would never run where it matters.

| Option | Why not |
|---|---|
| **git submodule / a third `lamsza-shared` repo** | A genuine single source of truth, and the most expensive one: submodule init lands in every clone, both CI workflows and the owner's deploy path, and it is the easiest of these to get wrong. |
| **npm workspace or a `file:../lamsza` dependency** | Breaks the "four independent repos, four independent deploys" property in `WAYS_OF_WORKING.md` §1 and §2. `lamsza-admin` would stop building from its own checkout. |
| **A cross-repo test that diffs the two trees** | Cannot run in single-repo CI — exactly where drift would be caught. Useful as a local extra, which is what `--check` is. |
| **Leave it hand-kept and documented** | The status quo BOG-42 was opened to end. |

What this costs: each app's copy is committed, so a shared-module change touches
up to four repos and four commits. That is the price of keeping the deploys
independent, and the manifest test makes forgetting a commit loud.

## What is shared, and what is not

The lists live in `shared-frontend-modules.json`. Today:

| Consumer | Share |
|---|---|
| `lamsza-admin` | What it imports (narrowed on 2026-10-09 from "everything"): `src/lib/accountPrefs.js`, `entryHistory.js`, `entryHours.js`, `entryPhotos.js`, `entryPublicExtras.js`, `entryType.js`, `eventImage.js`, `quickLinksDisplay.js`, `scheduleActivityTypes.js`, `websiteDomain.js`, `src/lib/stores/{auth,theme}.js`, `src/lib/components/{CategoryMultiSelect,EntryHoursEditor,GoogleSignIn,HuTimeInput}.svelte`, `src/lib/components/{ConfirmDialog,NoticeDialog,SignInDialog,ErrorPage,ErrorShell}.svelte`, `src/lib/icons/{AppIcon,ErrorLantern}.svelte`, `src/styles/{global,typography,component-typography}.css`, `tests/sharedFrontendModules.test.js` and `tests/noEmdash.test.js` |
| `lamsza-szotar`, `lamsza-jatszoter` | `src/lib/icons/{AppIcon,ErrorLantern}.svelte`, `src/lib/components/{ErrorPage,ErrorShell,GoogleSignIn,SignInDialog,AppsLauncher,AccountMenu,AccountPage,AccountDetails,ThemeSettings,Hero}.svelte`, `src/lib/{networkOrigins,networkApps,accountDetails}.js`, `src/styles/global.css`, `tests/sharedFrontendModules.test.js`, `tests/noEmdash.test.js`, `tests/networkApps.test.js` and `tests/accountDetails.test.js` |
| `lamsza-szotar`, `lamsza-jatszoter` | `src/lib/games/catalog.js` and `src/lib/games/GameIcon.svelte` (the games' colours and icons, and `featuredGames` for home pages; UI_BASELINE `game-look`), at the same paths. |

`AppIcon`, the dialogs, `SignInDialog` and the three stylesheets joined on 2026-10-07,
when the owner chose one icon system, one set of dialogs and one base stylesheet for the
network (`docs/network/UI_BASELINE.md`). `ErrorPage`, `ErrorShell` and `ErrorLantern` joined the same day,
together with Szótár and Játszótér, for the one network error page.
`AppsLauncher`, `networkApps.js`, `networkOrigins.js` and the `networkApps` test joined on 2026-10-08
for the apps launcher (UI_BASELINE "tb-apps-launcher"); `networkOrigins.js` already had the same
logic in all three public apps, so one copy replaces three. Each app keeps its own
`tests/networkOrigins.test.js`.
`AccountMenu`, `AccountPage`, `AccountDetails`, `ThemeSettings`, `accountDetails.js` and its test joined the same
day for the shared account menu, Fiók page and Téma panel (UI_BASELINE `tb-account-menu`, `acc-page`).

`global.css` holds only what more than one app uses (2026-10-09): tokens, base elements, buttons,
the toolbar, launcher and account menu, dialogs, the error page, the list layout, info boxes,
skeletons and the hero (`Hero.svelte`, UI_BASELINE "hero"). Rules whose every class is used only
in Lámsza (its home page search and discover, widgets, link cards, index tags, crests…) live in
`src/styles/lamsza.css`, Lámsza's own file, loaded right after `global.css` and not synced. The
exception: a rule stays in `global.css` when moving it after the shared rules would change which
rule wins. A new
rule goes in `global.css` only when another app uses it; otherwise in the app's own stylesheet.
About 60 classes in `global.css` are used by no app; they stay until a separate clean-up checks
for class names built in code.

Deliberately **not** shared, because the apps legitimately differ:

- `src/lib/api.js`: different base URLs and different endpoint sets. The shared
  Svelte files only import `getApiBase` from it, which every app has.
- `typography.css` and `component-typography.css` in Szótár and Játszótér: the
  owner chose "keep each app's own" (UI_BASELINE "typo-files").
- `theme.js` and `openLogin.js` in Szótár and Játszótér: their content differs per
  app (Lámsza's theme store saves to the account, for one).
- Szótár's and Játszótér's confirm dialog (`AppDialog.svelte`, the baseline look
  that Lámsza's `ConfirmDialog` copies, driven by a store).

Paths are relative to each app's frontend root: the repo root in `lamsza`,
`frontend/` in the other three (`WAYS_OF_WORKING.md` §5).

The duplicated **unit tests** for these modules (`accountPrefs.test.js`,
`entryHours.test.js`, …) stay duplicated on purpose. They are cheap, they run
in the repo that would break, and the manifest already guarantees they test the
same source.
