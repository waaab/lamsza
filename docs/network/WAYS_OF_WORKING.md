# Lámsza network — ways of working

How the four Lámsza apps are organised, and the rules that let one Paperclip
board drive four independent repos.

Approved on BOG-14, 2026-10-06. R5 added on BOG-33, 2026-10-06. R7 added on
BOG-54 and amended on BOG-57 (branch protection declined), 2026-10-06. This file
is the source of truth. If a task comment and this
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

### R7 — CI is a gate only if somebody reads it

**What CI is.** GitHub Actions, on every push and every pull request. Two test
jobs per repo, named **`frontend`** and **`backend`** — two and not one, because
when it was a single job a frontend failure marked the Go steps "skipped" and
the backend went unbuilt and untested for hours without the summary saying so.
All four repos have the same two job names, which is what makes one required
status-check setting work across the network.

| Repo | Workflow | Test jobs | `ci-status` |
| --- | --- | --- | --- |
| `waaab/lamsza` | `.github/workflows/build-test.yml` | on `main` | yes |
| `waaab/lamsza-admin` | `.github/workflows/build.yml` | on `main` | yes |
| `waaab/lamsza-szotar` | `.github/workflows/build-test.yml` | on `main` | yes |
| `waaab/lamsza-jatszoter` | `.github/workflows/build-test.yml` | on `main` | yes |

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

**How a red CI reaches a person: a notification, and no branch protection.**
Attila chose both on BOG-54, then on BOG-57 (2026-10-06) decided against the
branch-protection half once its full cost was priced. The notification is what
ships. CI that nobody reads is not a gate — `lamsza` CI stayed red across four
pushes in 17 minutes, three of them straight onto `main`, before BOG-44 noticed.

**Half one, shipped: the `ci-status` job.** A third job in every workflow mirrors
the state of CI on `main` into one GitHub issue, titled exactly
**`CI is red on main`**. It opens that issue when `main` goes red, comments on it
on each further red push instead of filing duplicates, and **closes it on the
next green run**. So the issue existing *is* the answer to "is `main` green right
now", readable without credentials by anyone who can see the repo.

It had to be pushed from inside Actions rather than polled from here, because
nothing on this machine can read CI: no `gh` CLI, git auth is SSH-key only,
unauthenticated `api.github.com` reads from this IP are rate-limited to zero, and
two of the four repos (`lamsza-szotar`, `lamsza-jatszoter`) are private. Inside
Actions, `GITHUB_TOKEN` already exists, so this needs no secret configured and
works the same in a private repo.

- Logic: `.github/ci-status-issue.sh`, the same file in all four repos.
- It only fires on pushes to `main`: a pull request shows its own checks, and a
  feature branch being red mid-work is normal and must not notify anyone.
- A cancelled or skipped run is not reported as a failure.
- After editing it, run `bash .github/ci-status-issue.test.sh` (needs `jq`, no
  network, no credential — it fakes `gh` and checks the real `--jq` filter).
- **That test also runs as a step of the `frontend` job, in all four repos**, so a
  broken receiver fails a check somebody watches instead of failing silently. It
  has to be guarded from outside itself: if `ci-status-issue.sh` breaks, the only
  symptom is the `ci-status` job going red on `main` — and the thing that reports
  a red `main` is that same script. It sits in `frontend` because `backend` runs
  from `backend/` and waits on a database; in the three repos whose `frontend` job
  runs from `frontend/` the step needs `working-directory: .`, because the test
  lives at the repo root.
- **One repo setting can still switch it off, and it says so.** *Settings >
  Actions > General > Workflow permissions* is a per-repo ceiling on
  `GITHUB_TOKEN`. While it reads "Read repository contents and packages
  permissions", the `issues` scope is dropped no matter what the job's
  `permissions:` block asks for, every call 403s, and the receiver is dead — the
  BOG-54 failure one level up. The job's `permissions: issues: write` is
  necessary but not sufficient.

  So the script diagnoses its own refusal rather than dying with a bare `HTTP
  403`. It prints which setting to change, in which repo, to the step log **and
  to `$GITHUB_STEP_SUMMARY`**, which renders on the run's summary page — the one
  screen a person actually opens. An unrecognised failure is quoted verbatim
  instead of guessed at. It still exits non-zero in every case: a receiver that
  cannot write must leave the job red, never look quiet.

  **Attila approved setting it to *Read and write permissions* in all four
  repos, on BOG-57, 2026-10-06.** No repo content needs to change. It was not
  applied in that run: the setting is owner-console or token work and this
  machine has neither, so it is still pending on BOG-57.

- **The job has been seen running; its write path has not.** On 2026-10-06 the
  `ci-status` job ran and passed on both public repos (`lamsza` run
  `37489216167`, `lamsza-admin` run `37489231652`), which answers the older
  "never observed" note. But `main` was green and neither repo had an open
  issue, so the script only made its `gh issue list` **read** — which succeeds
  under the read-only default. The write path, the one the setting above gates,
  has still never executed. The setting is therefore unverified, not
  verified-good, and the first push that matters is the first red `main`. That
  is exactly the case the self-diagnosis above is built for: a wrong setting
  costs one red run to discover, not silence.

**Half two: branch protection on `main` was considered and declined.** Decided
by Attila on BOG-57, 2026-10-06, after BOG-54 had provisionally asked for it.
There is no branch protection on any of the four repos and none is planned.

The reason is the cost of the one setting that would have made it work.
"Include administrators" has to be on, because agents push with the owner's SSH
key — to GitHub every agent push is an admin push, and admins bypass required
checks, so with it off the rule is decorative. On it also stops Attila's own
direct pushes to `main`. For a network with one engineering agent and one owner
that buys little and costs the owner a working path to his own repos. Two of the
four repos (`lamsza-szotar`, `lamsza-jatszoter`) are private on a personal
account, which would additionally have needed a paid plan or rulesets.

`docs/network/apply-branch-protection.sh` stays in the repo as the written-down
spec of what was declined — **do not run it.** If the decision is ever revisited,
one thing in it must survive the revisit: the required contexts are `frontend`
and `backend` only, **never `ci-status`**, which is gated on `refs/heads/main`,
never reports on a pull-request branch, and as a required check would block every
pull request forever.

**So the git grant does not change.** `docs/AGENT_ENVIRONMENT_POLICY.md` stands
as written: agents commit, merge and push, `main` included. Pushing to `main` is
still how work lands, and nothing mechanical stops a red push. The whole weight
of "do not break `main`" sits on the rules below and on the notification.

**So the rule is: the local gate is the real gate.** This is now permanent, not
a holding position — nothing is coming to replace it:

1. Run the `docs/LOCAL_DEV_CHECKS.md` checks your change touches **before** you
   push, not after. `ci-status` tells someone that `main` broke; it does not
   tell them about what CI never ran — the real `npm run build` against a live
   backend, and anything that only breaks on real content rather than on the
   empty schema CI builds.
2. Never merge to `main` with a check you already know is red. A red local check
   is a blocker whether or not anything reports it.
3. `npm run build` in `lamsza` only counts with the backend up on 3001. Without
   it the prerender fails with `ECONNREFUSED`, the build still exits 0, and the
   error states get baked into the pages.

**Open.** One thing: *Workflow permissions* is approved but not yet set to read
and write in the four repos, which needs the owner console or a GitHub token —
tracked on BOG-57. Branch protection is closed, not open: declined. Everything
else in R7 is live.

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
