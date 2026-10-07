# Lámsza network — ways of working

How the four Lámsza apps are organised, and the rules that let one task list
drive four independent repos.

Approved on BOG-14, 2026-10-06. R5 added on BOG-33, 2026-10-06. R7 added on
BOG-54 and amended on BOG-57 (branch protection declined), 2026-10-06. Paperclip switched
off and R5 made tool-independent, 2026-10-07. Pushing to `main` made subject to
the owner's confirmation (R5, R7), 2026-10-07. R1-R4 rewritten for work
without a tracker, R5 check 4 and R8-R17 added after the network review,
2026-10-07. This file
is the source of truth. If a task comment and this
file disagree, this file wins until it is changed here.

---

## 1. The four apps

One project — **lamsza.com network** — holds all four apps. Each app is its
own git repo and its own deploy unit.

| App | Folder | Repo | Role |
| --- | --- | --- | --- |
| lamsza | `~/projects/lamsza-network/lamsza` | `waaab/lamsza` | Main app: startlap and search. **Primary workspace.** |
| admin | `~/projects/lamsza-network/lamsza-admin` | `waaab/lamsza-admin` | Admin: writes the directory data |
| szotar | `~/projects/lamsza-network/lamsza-szotar` | `waaab/lamsza-szotar` | Dictionary |
| jatszoter | `~/projects/lamsza-network/lamsza-jatszoter` | `waaab/lamsza-jatszoter` | Word games |

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

- `~/projects/lamsza-network/start-lamsza-network.sh` → `lamsza/scripts/start-lamsza-network.sh`
- `~/.local/bin/lamsza-network` → `~/projects/lamsza-network/start-lamsza-network.sh`

It finds the four apps by resolving its own path through those symlinks and then
walking up until it reaches the directory that contains all four repos — on this
machine, `~/projects/lamsza-network`. Set `LAMSZA_PROJECTS_ROOT` to override that, and
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

## 3. One parent folder, four repos

The four repos sit side by side under `~/projects/lamsza-network/`. That folder
is not a git repo; it holds the repos, the `start-lamsza-network.sh` symlink, a
`CLAUDE.md` symlink to the tracked `docs/network/CLAUDE.md` in this repo, and a
VS Code workspace file.

Until 2026-10-07 the repos sat directly under `~/projects`, and `szotar` and
`jatszoter` had no `lamsza-` prefix. Paths in the task history still use those
names. One thing had to follow the move:

- **The szotar and jatszoter compose files pin `name:`** so their database
  volumes (`szotar_pgdata_szotar`, `jatszoter_pgdata_jatszoter`) survive the
  rename. Do not remove it: compose would derive a new project name from the
  folder and start on an empty volume.

A monorepo is still not recommended: it loses the four independent deploys.

Shared docs go in **`lamsza/docs/network/`** instead. `lamsza` is the primary
workspace, so every agent already has it.

---

## 4. The rules

### R1 — Every task and every commit names its app

Tasks live in **`docs/network/OPEN_ITEMS.md`** (Paperclip is switched off; there
is no other tracker). Each item and each commit subject starts with the app tag:

`[lamsza]` · `[admin]` · `[szotar]` · `[jatszoter]` · `[network]`

Use `[network]` when the work crosses two or more apps.

### R2 — One task, one app, where it can be

If a change fits in one repo, keep it in one task and one commit series. One
task then stays inside one workspace.

### R3 — Cross-app changes go contract-first

A `[network]` change follows a fixed order: **the data or API contract first
(usually the lamsza schema or API), then the apps that read it.** Commit and
verify the contract side before touching a consumer.

Two contracts are easy to miss:

- **A schema change lands with its dump.** Run `scripts/db-dump-schema.sh` and
  commit `backend/schema/` in the same commit as the code that changes the
  schema. `backend/schema/schema_test.go` fails when boot DDL names a table or
  column the dump lacks. *Why:* `admin_audit_log` was created on boot but never
  dumped, so every bootstrapped database (CI, scratch, a new machine) had no
  audit table for admin, and no suite noticed.
- **A shared UI change is a contract too.** A change to lamsza's header, footer,
  sign-in dialog, tokens or theme gets a follow-up in each app that copies it,
  and the UI rule file (R11) changes in the same commit. *Why:* lamsza's
  Adatvédelem footer link (2026-10-06) never reached szotar or jatszoter.

### R4 — One agent per repo at a time

Two agents editing the same working tree cause conflicts. Parallel work in one
repo needs separate worktrees, and each worktree is closed out by R5 check 4.
At the start of a run, look before you edit:

```bash
git worktree list; git branch -a --no-merged main; git stash list; git status --short
```

Anything unexpected is reported, not silently worked around. Cross-app work is
sequential by R3.

### R5 — A task is not done until its code is committed, merged and pushed

"Done" means the owner has the change on `origin/main`, not that an agent wrote
the code. Before calling a task done, check three things in each repo it
touched, in this order:

| Check | Not done while | How to check |
|---|---|---|
| 1 — committed | anything is uncommitted | `git status --porcelain` prints nothing |
| 2 — merged | the work sits only on a branch | `git merge-base --is-ancestor HEAD main` exits 0 |
| 3 — pushed | local `main` is ahead of `origin/main` | `git fetch && git log --oneline origin/main..main` prints nothing |

When you report the task done, list the commits that shipped it, so a reviewer
can see what landed.

Work that is meant to land somewhere other than `main` checks against that
branch instead. A repo with no remote skips check 3.

Check 3 needs a push, and pushing to `main` needs Attila's explicit confirmation
(R7, `docs/AGENT_ENVIRONMENT_POLICY.md`). Until Attila confirms, the task is
committed and merged but not done: report it as ready to push, with the commits.

**Check 4, closed out.** After the merge, remove the worktree and delete the
branch, locally and (with the push) on origin. Unfinished work is committed and
pushed on its own branch, never left uncommitted in a worktree, and never kept
in a stash. *Why:* the review on 2026-10-07 found 3,500 lines of tests
uncommitted in a worktree, a finished branch pushed but never merged, a
2,600-line stash, and about 50 merged branches nobody deleted.

Skip a check only when it is genuinely wrong for the task — check 1 for a task
that produced a brief, a decision or a review; check 2 for work meant to stay on
a branch, such as a snapshot or a spike. Say which check you skipped and why.

Check 1 came from BOG-32: four tasks were marked done with every line of their
code uncommitted, and the work was lost. Checks 2 and 3 came from BOG-38:
BOG-17 and BOG-28 cleared check 1 and still shipped nothing, because the commits
sat on a branch nobody merged.

Until 2026-10-07 these checks were enforced by `scripts/paperclip-issue-update.sh`
when it set a Paperclip issue to `done`. Paperclip is switched off; the script
and its test stay in the repo, unused, by the owner's decision. The checks
stand on their own.

Why this rule exists: BOG-32 found four tasks (BOG-3, BOG-4, BOG-5, BOG-7)
marked `done` while all of their code sat uncommitted on one feature branch.
None had shipped, and the work was discarded.

### R6 — What is true for one app is true for all four

These standards are network-wide:

- **Local-only development.** `docs/AGENT_ENVIRONMENT_POLICY.md` in this repo is
  the standing rule for all four apps. Production is the owner's.
- **Fixed ports**, as the start script defines them (§1).
- **`/api/health` answers `{"ok":true,"db":"up"}`** on every backend, and 503
  with `"db":"down"` when the database does not answer.
- **`npm test` at the repo root runs that repo's suites.** In admin, szotar and
  jatszoter it runs frontend and backend; in lamsza it runs the frontend, and
  `npm run test:go` runs the backend.
- **The same runtime versions,** recorded once in `docs/network/VERSIONS.md`.
- **`docs/LOCAL_DEV_CHECKS.md` is the definition of "verified":** network up,
  the APIs you touched smoked, every suite run, output recorded.

### R7 — CI is a gate only if somebody reads it

**What CI is.** GitHub Actions, on every push and every pull request. Two test
jobs per repo, named **`frontend`** and **`backend`** — two and not one, because
when it was a single job a frontend failure marked the Go steps "skipped" and
the backend went unbuilt and untested for hours without the summary saying so.
All four repos have the same two job names, which is what makes one required
status-check setting work across the network.

| Repo | Workflow | Test jobs | `ci-status` |
| --- | --- | --- | --- |
| `waaab/lamsza` | `.github/workflows/build-test.yml` | every push and pull request | yes |
| `waaab/lamsza-admin` | `.github/workflows/build.yml` | every push and pull request | yes |
| `waaab/lamsza-szotar` | `.github/workflows/build-test.yml` | every push and pull request | yes |
| `waaab/lamsza-jatszoter` | `.github/workflows/build-test.yml` | every push and pull request | yes |

Every job runs on `ubuntu-24.04`, the production droplet's OS, with Node from
`.nvmrc` and Go from `backend/go.mod`; see `docs/network/VERSIONS.md`. The suites
read only `TEST_DATABASE_URL` (R8), which each workflow sets to its service
database.

**Every backend job runs `go test -count=1 ./...`.** No workflow excludes a Go
package by name any more — BOG-53 removed the last three exclusions. Three of
the four backend jobs now run a `postgres:16` service, matching local dev, and
each builds its own schema from the repo:

| Repo | How CI gets a schema |
| --- | --- |
| `lamsza` | `scripts/db-bootstrap.sh` applies the committed `backend/schema/` — the same script and files that create a fresh local database |
| `szotar` | the backend's own `db.Migrate()`, which the suites call |
| `jatszoter` | the backend's own `db.Migrate()`, which the suites call |
| `lamsza-admin` | no service: its DB-bound tests skip in CI and run locally on a scratch database built from lamsza's schema (`npm run test:backend`) |

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
this machine cannot read CI for all four repos: there is no `gh` CLI, git auth
is SSH-key only, and two of the four repos (`lamsza-szotar`, `lamsza-jatszoter`)
are private. Inside Actions, `GITHUB_TOKEN` already exists, so this needs no
secret configured and works the same in a private repo.

**The two public repos can be read from here.** When BOG-54 was written,
unauthenticated `api.github.com` reads from this IP were rate-limited to zero.
On 2026-10-07 they worked again, within GitHub's normal unauthenticated limit
of 60 requests an hour:

```bash
curl -s "https://api.github.com/repos/waaab/lamsza/actions/runs?branch=main&per_page=3" \
  | jq -r '.workflow_runs[] | "\(.head_sha[0:7]) \(.status)/\(.conclusion)"'
```

Use the same call for `lamsza-admin`. For the two private repos it returns
`Not Found`, so the `CI is red on main` issue is still the only signal for them.

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

**The git grant: commit and merge locally, push only on confirmation.** Agents
commit and merge locally, `main` included. Pushing to `main`, or merging a pull
request into it, always needs Attila's explicit confirmation first (narrowed on
2026-10-07; see `docs/AGENT_ENVIRONMENT_POLICY.md`). Pushing to `main` is still
how work lands, and nothing on GitHub stops a red push. The whole weight of "do
not break `main`" sits on that confirmation, the rules below and the notification.

**So the rule is: the local gate is the real gate.** This is now permanent, not
a holding position — nothing is coming to replace it:

1. Run the `docs/LOCAL_DEV_CHECKS.md` checks your change touches **before** you
   push, not after. `ci-status` tells someone that `main` broke; it does not
   tell them about what CI never ran — the real `npm run build` against a live
   backend, and anything that only breaks on real content rather than on the
   empty schema CI builds.
2. Never merge to `main` with a check you already know is red. A red local check
   is a blocker whether or not anything reports it.
3. `npm run build` in `lamsza` runs `scripts/preflight-api.js` first, which
   refuses to build (exit 1) when no backend answers on 3001; component fetches
   are gated on `canReachApi()`, so `dist/` does not depend on the backend.
   CI uses `npm run build:no-preflight`.

**Open.** *Workflow permissions* is approved but not yet set to read and write in
the four repos, which needs the owner console or a GitHub token; it is listed in
`docs/network/OPEN_ITEMS.md`. Branch protection is closed, not open: declined.
Everything else in R7 is live.

### R8 — Tests never touch dev data

DB-bound Go tests read only `TEST_DATABASE_URL`, never `DATABASE_URL` or `.env`,
and refuse a local database whose name does not end in `_test`
(`backend/internal/db/testguard.go` in each repo). `npm run test:go` (lamsza) and
`npm run test:backend` (the others) rebuild a scratch database first. An
exported variable beats `.env` in every backend. *Why:* 43 of the 46 users in
the dev `lamsza` database were test fixtures, a szotar test deleted whatever
proverb sat on today+40, and the documented "scratch" recipe silently used the
dev database because `.env` overrode it.

### R9 — Never act as the owner on dev data

Smoke tests are GET-only. Write paths are verified by tests on a scratch
database or a throwaway server pointed at one, never through the dev API, and
never with a hand-minted session for Attila's account. *Why:* 17 audit rows
recorded agent probes as the owner's actions in an append-only log.

### R10 — Docs move with code; one fact, one home

- When a commit changes a script, port, route, env var or test command,
  `git grep` all four repos for the old fact and fix it in the same change.
- Docs give the command, not a result count; a status table is replaced, never
  amended below.
- A fact has one home and the rest link to it: ports in the start script and §1,
  versions in `VERSIONS.md`, open work in `OPEN_ITEMS.md`.

*Why:* R6's "no repo has npm test", LOCAL_DEV_CHECKS' status tables, admin's
route count and szotar's "no database needed" were all wrong within hours of
the code that changed them, and a port table existed in eight places.

### R11 — Rule files have one canonical copy

`.cursor/rules/*.mdc` live in this repo and are copied to the other three by
`scripts/sync-cursor-rules.sh`, which also has `--check`. Never replace a rule
file's content in a bulk edit. *Why:* a port renumber on 2026-10-06 replaced
the UI-consistency rule in szotar and jatszoter with a port table, and admin
never had the rules.

### R12 — One launcher

`start-lamsza-network.sh` starts and stops the apps; no repo has its own
restart script. *Why:* lamsza's old `restart_all.sh` killed every Vite server on
the machine and stopped the database admin shares.

### R13 — Localhost-only covers code paths

No dev code path calls a `*.lamsza.com` host; network origins come from
`networkOrigins.js`, which resolves to localhost or `*.lamsza.test` in dev.
*Why:* jatszoter fetched production's `/api/config/public` on every page load
in dev.

**Server-to-server calls go to `http://127.0.0.1:<port>`, set by an env var.** A
backend that calls another app's backend (lamsza → szotar, jatszoter → szotar)
reads the target from its `.env` (`SZOTAR_ORIGIN`, `SZOTAR_BASE_URL`, …), and
the dev value is the other backend on `127.0.0.1` with its port from §1, never a
`*.lamsza.test` host. *Why:* lamsza reached szotar through
`https://szotar.lamsza.test`, which depends on the local nginx and mkcert setup
that lives in no repo; jatszoter already called `127.0.0.1:3002`. Owner's
decision, 2026-10-07.

### R14 — A GET never creates data

Except today's lazily created daily puzzle. *Why:* a GET for a future date
created Kaptár dailies ahead of time, which also made "GET-only" smoke tests
unsafe.

### R15 — Plans record their outcome

When a `docs/superpowers` plan is executed, tick its boxes or append an
"Outcome" note saying what was done and what was skipped, and why. *Why:* none
of about 580 plan checkboxes was ticked, and Kaptár Task 10 and Szórejtő
Task 13 were dropped without a word.

### R16 — Commits name their task and explain themselves

A subject with the R1 tag, a body that says what changed and why, and the
verification that was run. Never a snapshot commit that sweeps in ignored or
generated files. *Why:* several commits had no body, one credited the wrong
task, and a snapshot committed `server.pid`.

### R17 — Changelogs

Every repo has a `CHANGELOG.md` (Keep a Changelog, English). Every notable
change adds a line under `## [Unreleased]` in the same commit; a line users
would notice ends with `[public]`. At a release, `[Unreleased]` becomes
`## [x.y.z] - date`, and the `[public]` lines are rewritten in plain Hungarian
as a new entry in the app's `publicChangelog.js`, which the `/valtozasnaplo`
page and the footer version read. Admin has no public page. *Why:* lamsza's
changelog stopped at the admin extraction, szotar and jatszoter had none, and
the public pages listed nothing from the last two weeks.

### R18 — One admin app for the network

Every app is administered from the admin app (`lamsza-admin`): `/` for the main
Lámsza data, `/dictionary` for Szótár, `/games` for Játszótér. Routes are
English, the UI is Hungarian. The admin app's `ADMIN_GOOGLE_EMAILS` is the
network's one admin list; a person signs in as admin only there. Decided on
2026-10-07; the move runs in batches (OPEN_ITEMS), and until it ends Szótár and
Játszótér still carry their own `/admin` pages.

How the admin app reaches another app's data, without touching its database:

- The other app keeps owning its database and its rules (R3). It offers its
  admin features as an **internal admin API** under `/internal/admin/...`.
- The admin backend relays its own `/api/admin/dictionary/...` and
  `/api/admin/games/...` routes to that API, server-to-server on
  `http://127.0.0.1:<port>` from an env var (`SZOTAR_ADMIN_URL`,
  `JATSZOTER_ADMIN_URL`; R13). The admin browser only talks to its own backend,
  every relayed write lands in `admin_audit_log`, and the acting admin's email
  goes along in `X-Admin-Email`.
- The internal API trusts the admin backend, not a person, and only when all
  three hold: the path is outside `/api/` (production nginx proxies only
  `/api/`, so `/internal/` is never public); the request comes from loopback
  with no `X-Forwarded-For` or `X-Real-IP` header (anything through nginx has
  one); and it carries `Authorization: Bearer <token>`, a secret per app pair
  (`ADMIN_SERVICE_TOKEN` in the app, `SZOTAR_ADMIN_TOKEN` /
  `JATSZOTER_ADMIN_TOKEN` in admin), compared in constant time. No token
  configured means the API is off.

*Why:* three sign-ins, three admin lists and three admin UIs, and only one of
them kept an audit trail.

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
