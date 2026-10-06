# Shared frontend modules — lamsza owns, lamsza-admin gets a generated copy

**Decided on BOG-42, 2026-10-06.** This file is the source of truth for the
17 frontend files that exist in both `lamsza` and `lamsza-admin`.

---

## The rule

1. **`lamsza` owns every shared module.** Edit it there, never in `lamsza-admin`.
2. **`lamsza-admin/frontend` carries a generated copy.** It is a real committed
   file, not a link or a submodule, so the admin repo still clones, builds and
   deploys on its own.
3. **`scripts/sync-shared-frontend.sh` in `lamsza` does the copying** and
   rewrites `shared-frontend-modules.json` — a path → SHA-256 manifest — in
   **both** repos.
4. **Each repo's own test suite checks its own copies against that manifest.**
   `tests/sharedFrontendModules.test.js` (which is itself one of the shared
   files) hashes every listed module and fails on a mismatch.

Changing a shared module:

```bash
cd ~/projects/lamsza
$EDITOR src/lib/entryHours.js
scripts/sync-shared-frontend.sh          # copies + rewrites both manifests
node --test 'tests/*.test.js'            # green here
(cd ../lamsza-admin/frontend && npm test)  # green there
# commit both repos
```

Adding or removing a shared module: add or delete the path under `modules` in
`lamsza/shared-frontend-modules.json` (any placeholder hash will do) and run
the script — it re-hashes from `lamsza` and writes both manifests.

`scripts/sync-shared-frontend.sh --check` verifies without writing, for when
both repos are checked out. `LAMSZA_ADMIN_ROOT` overrides where `lamsza-admin`
lives; the default is `../lamsza-admin`.

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

What this costs: the admin copy is committed, so a shared-module change touches
two repos and two commits. That is the price of keeping the two deploys
independent, and the manifest test makes forgetting the second commit loud.

## What is shared, and what is not

The list lives in `shared-frontend-modules.json`. Today: `src/lib/accountPrefs.js`,
`entryHistory.js`, `entryHours.js`, `entryPhotos.js`, `entryPublicExtras.js`,
`entryType.js`, `eventImage.js`, `quickLinksDisplay.js`, `scheduleActivityTypes.js`,
`websiteDomain.js`, `src/lib/stores/{auth,theme}.js`,
`src/lib/components/{CategoryMultiSelect,EntryHoursEditor,GoogleSignIn,HuTimeInput}.svelte`
and `tests/sharedFrontendModules.test.js`.

Deliberately **not** shared, because the two apps legitimately differ:

- `src/lib/api.js` — different base URLs and different endpoint sets.
- `src/lib/icons/AppIcon.svelte` — the admin app's icon set is its own.

Paths are relative to each app's frontend root: the repo root in `lamsza`,
`frontend/` in `lamsza-admin` (`WAYS_OF_WORKING.md` §5).

The duplicated **unit tests** for these modules (`accountPrefs.test.js`,
`entryHours.test.js`, …) stay duplicated on purpose. They are cheap, they run
in the repo that would break, and the manifest already guarantees they test the
same source.
