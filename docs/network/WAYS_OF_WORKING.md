# Lámsza network — ways of working

How the four Lámsza apps are organised, and the rules that let one Paperclip
board drive four independent repos.

Approved on BOG-14, 2026-10-06. R5 added on BOG-33, 2026-10-06. R7 added on
BOG-54, 2026-10-06. This file is the source of truth. If a task comment and this
file disagree, this file wins until it is changed here.

---

## 1. The four apps

One Paperclip project — **lamsza.com network** — holds all four apps as
separate workspaces. Each app is its own git repo and its own deploy unit.

| App | Folder | Repo | Role |
| --- | --- | --- | --- |
| lamsza | `~/projects/lamsza` | `waaab/lamsza` | Main app: startlap and search. **Primary workspace.** |
| admin | `~/projects/lamsza-admin` | `waaab/lamsza-admin` | Admin: writes the directory data |
| szotar | `~/projects/szotar` | `waaab/lamsza-szotar` | Dictionary |
| jatszoter | `~/projects/jatszoter` | `waaab/lamsza-jatszoter` | Word games |

Three Postgres databases: `lamsza` (shared by the main app and admin), `szotar`,
`jatszoter`.

`lamsza/scripts/start-lamsza-network.sh` starts and stops the whole network on
fixed ports:

| App | Backend | Frontend |
| --- | --- | --- |
| admin | 3000 | 5173 |
| lamsza | 3001 | 5174 |
| szotar | 3002 | 5175 |
| jatszoter | 3003 | 5176 |

**The tracked copy is the only copy.** The script lives in this repo, at
`scripts/start-lamsza-network.sh`. Two symlinks point into it so the habitual
paths keep working — edit neither, edit the repo file:

- `~/projects/start-lamsza-network.sh` → `lamsza/scripts/start-lamsza-network.sh`
- `~/.local/bin/lamsza-network` → `~/projects/start-lamsza-network.sh`

It finds the four apps by resolving its own path through those symlinks and then
walking up until it reaches the directory that contains all four repos — on this
machine, `~/projects`. Set `LAMSZA_PROJECTS_ROOT` to override that, and
`LAMSZA_STATE_DIR` to move the logs and PID files. `status` prints both as
`Apps:` and `Logs:`. Run `scripts/tests/start-lamsza-network.test.sh` after
changing the script.

Until BOG-50 the script was an untracked file on one machine, which is why
BOG-31's fix to it had no commit to point at.

---

## 2. One project, not four

Four boards would split work that belongs together:

- **One go-live.** The whole network moves to the droplet together. Four boards
  give four priority orders and no single answer to "what is next".
- **One shared data contract.** Admin writes the `lamsza` database that the main
  app reads. Change it in admin and the main app can break. That task must live
  on one board.
- **One of everything else.** One domain, one design, one cookie and privacy
  story, one SEO setup, one agent policy.
- **Splitting later is cheap. Merging later is not.**

Give an app its own project only when all three are true: its own release
schedule, its own owner, and almost no shared tasks. No app is there yet.

## 3. The folders stay flat

The four folders sit directly under `~/projects` with no parent folder. **Leave
them there.** A parent folder would give a tidier `ls` and would break four
Paperclip workspace paths, the start script, and every absolute path in the
docs and the task history.

A monorepo is the same trade, bigger, and it also loses the four independent
deploys. Not recommended.

Shared docs go in **`lamsza/docs/network/`** instead. `lamsza` is the primary
workspace, so every agent already has it.

---

## 4. The seven rules

### R1 — Every issue names its app

Start the title with the app tag:

`[lamsza]` · `[admin]` · `[szotar]` · `[jatszoter]` · `[network]`

Use `[network]` when the work crosses two or more apps.

### R2 — One issue, one app, where it can be

If a change fits in one repo, keep it in one issue. One task then stays inside
one workspace and one pull request.

### R3 — Cross-app changes go contract-first: parent, then children

A `[network]` change follows a fixed order: **the data or API contract first
(usually admin or the database), then the apps that read it.**

- The parent issue owns the contract.
- One child issue per consuming app, blocked on the parent.
- No consumer starts before the contract lands.

### R4 — One repo, one run at a time

Two runs editing the same repo at the same time cause conflicts. Before you
start, check that no other run holds that app. Cross-app work is sequential by
R3, not parallel.

### R5 — A task is not done until its code is committed, merged and pushed

Write every agent-driven status change through the script in the `lamsza` repo:

```bash
scripts/paperclip-issue-update.sh done --comment "what shipped"
```

Setting `done` has to clear three gates, in this order:

| Gate | It refuses when | Escape hatch |
|---|---|---|
| 1 — committed | `git status --porcelain` shows anything | `--allow-dirty` |
| 2 — merged | `HEAD` is not an ancestor of `main` | `--allow-unmerged` |
| 3 — pushed | local `main` is ahead of `origin/main` | `--allow-unpushed` |

Each refusal prints what is wrong — the dirty paths, the branch-only commits,
or the unpushed commits — and exits non-zero. On success the script appends the
task's branch-only commits to the comment, so a reviewer can see what shipped.

Gate 1 came from BOG-32: four tasks were marked done with every line of their
code uncommitted, and the work was lost. Gates 2 and 3 came from BOG-38:
BOG-17 and BOG-28 cleared gate 1 and still shipped nothing, because the commits
sat on a branch nobody merged. "Done" means the owner has the fix on
`origin/main`, not that an agent wrote the code.

Use `--base <ref>` when the work is meant to land somewhere other than `main`.
Gate 3 is skipped when the repo has no `origin/<base>`, so a local-only repo is
not a failure.

Pass an escape hatch only when the gate is genuinely wrong for the task —
`--allow-dirty` for a task that produced a brief, a decision or a review, or a
workspace shared with another run; `--allow-unmerged` for work meant to stay on
a branch, such as a snapshot or a spike. Say why in the comment. Each flag must
be typed on purpose; the default is safe.

Why this rule exists: BOG-32 found four tasks (BOG-3, BOG-4, BOG-5, BOG-7)
marked `done` while all of their code sat uncommitted on one feature branch.
None had shipped, and the work was discarded.

Run `scripts/tests/paperclip-issue-update.test.sh` after changing the script.

### R6 — What is true for one app is true for all four

These standards are network-wide:

- **Local-only development.** `docs/AGENT_ENVIRONMENT_POLICY.md` in this repo is
  the standing rule for all four apps. Production is the owner's.
- **Fixed ports**, as the start script defines them (§1).
- **`/api/health` on every backend.** All four have it.
- **A working `npm test` in every repo.** Not true yet — no repo defines a root
  `test` script. Tracked per app: BOG-2 (lamsza), BOG-20 (admin), BOG-21
  (szotar), BOG-22 (jatszoter).
- **`docs/LOCAL_DEV_CHECKS.md` is the definition of "verified":** network up,
  every API smoked, every suite run, output recorded.

### R7 — CI is a gate only if somebody reads it. Today that somebody is you

**What CI is.** GitHub Actions, on every push and every pull request. Two jobs
per repo, named **`frontend`** and **`backend`** — two and not one, because when
it was a single job a frontend failure marked the Go steps "skipped" and the
backend went unbuilt and untested for hours without the summary saying so.

| Repo | Workflow | Status |
| --- | --- | --- |
| `waaab/lamsza` | `.github/workflows/build-test.yml` | on `main` |
| `waaab/lamsza-admin` | `.github/workflows/build.yml` | on `main` |
| `waaab/lamsza-szotar` | `.github/workflows/build-test.yml` | on `main` |
| `waaab/lamsza-jatszoter` | `.github/workflows/build-test.yml` | on `main` |

**Every backend job runs `go test -count=1 ./...`.** No workflow excludes a Go
package by name any more — BOG-53 removed the last three exclusions. Three of
the four backend jobs now run a `postgres:16` service, matching local dev, and
each builds its own schema from the repo:

| Repo | How CI gets a schema |
| --- | --- |
| `lamsza` | `scripts/db-bootstrap.sh` applies the committed `backend/schema/` — the same script and files that create a fresh local database |
| `szotar` | the backend's own `db.Migrate()`, which the suites call |
| `jatszoter` | the backend's own `db.Migrate()`, which the suites call |
| `lamsza-admin` | no database-bound suite; no service needed |

Do not add a CI-only copy of a schema. If CI needs something the repo cannot
build, that is a gap in the bootstrap, and the bootstrap is what to fix.

**What CI still does not cover.** There is no running backend, so `lamsza` runs
`npm run build:no-preflight` — `npm run build` refuses to build when no Go
backend answers on 3001. CI also runs against an *empty* database seeded only
from the repo, which is stricter than dev data in the ways that matter and
weaker in one: it will not catch something that only breaks on real content.

A green CI therefore still does **not** mean the localhost checks passed. It is
the weaker of the two gates, not the stronger.

**Nothing reads CI automatically.** As of 2026-10-06, all four of these are true:

- no branch protection on `main` in any of the four repos, so no required status
  check;
- agents merge to `main` with `git push`, not through a pull request, so nobody
  ever sees a red check in a PR UI;
- nothing routes an Actions failure to this board or to the owner;
- and polling it from this machine does not work either. There is no `gh` CLI
  here, git auth is SSH-key only, unauthenticated `api.github.com` reads from
  this machine's IP are rate-limited and currently return `403`, and two of the
  four repos (`lamsza-szotar`, `lamsza-jatszoter`) are private.

That is how `lamsza` CI stayed red across four pushes in 17 minutes, three of
them straight onto `main`, before BOG-44 noticed.

**So the rule is: the local gate is the real gate.** Until that changes,

1. Run the `docs/LOCAL_DEV_CHECKS.md` checks your change touches **before** you
   push, not after. CI will not catch what you skipped — nobody is watching it.
2. Never merge to `main` with a check you already know is red. A red local check
   is a blocker whether or not anything reports it.
3. `npm run build` in `lamsza` only counts with the backend up on 3001. Without
   it the prerender fails with `ECONNREFUSED`, the build still exits 0, and the
   error states get baked into the pages.

**Open.** How a red CI reaches a person without an agent going to look — a
notification, branch protection with `frontend` and `backend` as required
checks, or both — is the owner's decision on BOG-54, because required status
checks would change the standing git grant in
`docs/AGENT_ENVIRONMENT_POLICY.md`. Record the answer here when it lands.

---

## 5. Repo layout — two shapes, both correct

Three apps use a `backend/` plus `frontend/` split. The main `lamsza` repo keeps
its frontend at the repo root.

| App | Backend | Frontend | Build output |
| --- | --- | --- | --- |
| admin | `backend/` | `frontend/` | `frontend/dist/` |
| szotar | `backend/` | `frontend/` | `frontend/dist/` |
| jatszoter | `backend/` | `frontend/` | `frontend/dist/` |
| **lamsza** | `backend/` | **repo root** (`src/`, `static/`, `svelte.config.js`, `vite.config.js`) | `dist/` |

**The root layout in `lamsza` is accepted, not a defect.** Decided on BOG-23. Do
not "fix" it by moving files into `frontend/`.

Why it stays:

- The cost of the move is real and reaches production. It renames `src/`,
  `static/`, `tests/`, `scripts/`, `package.json`, `package-lock.json`,
  `svelte.config.js`, `vite.config.js`, `jsconfig.json`, `manifest.json` and
  `.npmrc`, then needs matching edits in `.gitignore`, the CI workflow, the
  extension build, `docs/PRODUCTION_SERVER_SETUP.md`,
  `docs/PRODUCTION_ENVIRONMENT_NOTES.md` and the owner's deploy path for
  `dist/`. A wrong `dist/` path breaks a live site.
- The benefit is small. The only mechanical cost today is one `.` in the start
  script's app table (§1), which already works.
- `npm run dev` at the repo root starts the frontend dev server in all four
  apps. In the other three the root `package.json` is a thin proxy
  (`npm run dev --prefix frontend`); in lamsza the root `package.json` *is* the
  frontend package. Same command, same result.

What this means for you:

- **In `lamsza`, there is no `frontend/`.** Frontend paths are `src/…`,
  `static/…`, `tests/…`. The empty `lamsza/frontend/` leftover was deleted on
  BOG-23. If you ever see it again, something created it by mistake — delete it,
  do not fill it.
- Run frontend commands from the repo root in `lamsza`, and from `frontend/` (or
  through the root proxy script) in the other three.
