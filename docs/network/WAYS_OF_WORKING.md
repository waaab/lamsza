# Lámsza network — ways of working

How the four Lámsza apps are organised, and the rules that let one Paperclip
board drive four independent repos.

Approved on BOG-14, 2026-10-06. This file is the source of truth. If a task
comment and this file disagree, this file wins until it is changed here.

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

`~/projects/start-lamsza-network.sh` starts and stops the whole network on fixed
ports:

| App | Backend | Frontend |
| --- | --- | --- |
| admin | 3000 | 5173 |
| lamsza | 3001 | 5174 |
| szotar | 3002 | 5175 |
| jatszoter | 3003 | 5176 |

The script finds the apps under `$LAMSZA_PROJECTS_ROOT` (default `~/projects`).

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

## 4. The five rules

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

### R5 — What is true for one app is true for all four

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

---

## 5. Known inconsistency

szotar, jatszoter and admin all use `backend/` plus `frontend/`. The main
`lamsza` repo has `backend/`, but its frontend is at the repo root (`src/`,
`svelte.config.js`, `vite.config.js` at the top level). The start script has to
special-case it, and `npm run dev` means a different thing in lamsza than in the
other three. There is also an empty, untracked `lamsza/frontend/` folder — a
leftover that will mislead any agent that expects the common layout.

Tracked as BOG-23, low priority. Until it is fixed, do not assume `frontend/`
exists in the `lamsza` repo.
