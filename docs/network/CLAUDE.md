# Lámsza network workspace

This folder holds the four Lámsza repos side by side. It is not a git repo itself:
each app is its own repo, CI and deploy. For network rules,
`lamsza/docs/network/WAYS_OF_WORKING.md` is the source of truth. If it disagrees
with this file, it wins, except on paths and ports, where the start script wins.

This file is tracked as `lamsza/docs/network/CLAUDE.md`; `~/projects/lamsza-network/CLAUDE.md`
is a symlink to it. Edit it in the lamsza repo.

## Git

Never push to main without my explicit confirmation.

## The four repos

| Repo | What it is | Notes |
|---|---|---|
| `lamsza` | Main app (lamsza.com): startlap, directory, search. **Primary repo.** | Owns the `lamsza` DB schema (`backend/schema/`), the shared frontend modules, the network docs, rules and scripts. Frontend sits at the repo root, with no `frontend/` folder (deliberate, WoW §5). |
| `lamsza-admin` | The network's one admin app (WoW R18): `/` main data, `/dictionary`, `/games` | Shares the `lamsza` DB and **runs no DDL**: schema changes go in `lamsza`. Needs public data? Call `lamsza` on :3001. Szótár and Játszótér data only through their internal admin APIs. |
| `lamsza-szotar` | Székely dictionary (szotar.lamsza.com) | Own DB. Administered only from the admin app's `/dictionary` (WoW R18); no admin pages or admin links of its own. |
| `lamsza-jatszoter` | Word games (jatszoter.lamsza.com) | Own DB. Administered only from the admin app's `/games` (WoW R18); no admin pages or admin links of its own. |

How they relate:
- Admin writes the data that lamsza reads. A cross-app change goes contract-first:
  DB/API first, then the apps that consume it (WoW R3). A schema change lands with
  `scripts/db-dump-schema.sh` output in the same commit.
- The frontend files listed in `lamsza/shared-frontend-modules.json` are shared with the other apps
  (admin gets all of them; szotar and jatszoter get the icons, sign-in, error page and `global.css`).
  Edit them in `lamsza` only, then run `lamsza/scripts/sync-shared-frontend.sh` and commit in every repo that changed.
- `.cursor/rules/lamsza-network.mdc` and `no-emdash.mdc` are canonical in `lamsza`; run
  `lamsza/scripts/sync-cursor-rules.sh` and commit in all four repos (WoW R11).

## Running the network

```bash
./start-lamsza-network.sh start|stop|restart|status
```

- It's a symlink to `lamsza/scripts/start-lamsza-network.sh`, so edit that file and
  run `lamsza/scripts/tests/start-lamsza-network.test.sh` afterwards.
- `start` brings up the DBs (`docker compose up -d`), then `go run .` for each backend
  and `npm run dev` for each frontend. It skips anything already listening.
- `stop` kills the apps but **leaves the databases running**.
- It is the only launcher (WoW R12). Never `pkill -f vite` or `docker compose down`.
- Logs and PID files: `~/.cache/lamsza-network/<app>-{backend,frontend}.log`.

Ports (source: the start script, WoW §1):

| App | Backend | Frontend | Local URL |
|---|---|---|---|
| admin | 3000 | 5173 | https://admin.lamsza.test · http://localhost:5173 |
| lamsza | 3001 | 5174 | https://lamsza.test · http://localhost:5174 |
| szotar | 3002 | 5175 | https://szotar.lamsza.test · http://localhost:5175 |
| jatszoter | 3003 | 5176 | https://jatszoter.lamsza.test · http://localhost:5176 |

Frontends use `strictPort: true`. Never start one on another port. `*.lamsza.test`
resolves through `/etc/hosts` to local nginx (`/etc/nginx/sites-available/lamsza.test`,
mkcert certificate), which isn't in any repo. Google sign-in works only on the
`localhost` ports, for every app: Google rejects `.test` origins, so they can't be
added to the OAuth client.

## Tests and dev data

- **Tests never touch the dev databases** (WoW R8). Go suites read only
  `TEST_DATABASE_URL` and refuse a database whose name does not end in `_test`.
  Run `npm run test:go` in lamsza and `npm run test:backend` (or `npm test`) in the
  others: they rebuild a scratch database first.
- **Never write to dev data through the API, and never as the owner** (WoW R9). Smoke
  tests are GET-only; verify writes with tests on a scratch DB or a throwaway server
  pointed at one.
- Runtime versions (Ubuntu, Node, Go) for production, CI and dev: `lamsza/docs/network/VERSIONS.md`.

## Databases (Docker, Postgres 16)

| Container | Host port | DB | Used by | Compose file |
|---|---|---|---|---|
| `lamsza-db` | 5433 | `lamsza` | lamsza **and** admin | `lamsza/docker-compose.yml` |
| `szotar-db` | 5435 | `szotar` | szotar | `lamsza-szotar/docker-compose.yml` |
| `jatszoter-db` | 5434 | `jatszoter` | jatszoter | `lamsza-jatszoter/docker-compose.yml` |

- Database users are `lamsza_user`, `szotar_user` and `jatszoter_user`; there is no
  `postgres` role in szotar-db or jatszoter-db.
- Admin has no compose file of its own. It connects through `DATABASE_URL` in
  `lamsza-admin/backend/.env`.
- The szotar and jatszoter compose files set `name: szotar` / `name: jatszoter`. That
  pins the project name so the existing volumes (`szotar_pgdata_szotar`,
  `jatszoter_pgdata_jatszoter`) survive the folder rename. **Don't remove `name:`**:
  compose would derive `lamsza-szotar`, create an empty volume and leave the data
  behind. `lamsza` keeps its folder name, so it needs no pin.

## Where shared conventions live

- `lamsza/docs/network/WAYS_OF_WORKING.md`: the rules R1-R19 (tasks and tags, one agent
  per repo, "done" = committed, merged, pushed and closed out, CI, test data, docs,
  changelogs, one admin app, Bucharest "today"), repo layouts.
- `lamsza/docs/network/OPEN_ITEMS.md`: the task list, including what only the owner does.
- `lamsza/docs/network/SHARED_FRONTEND_MODULES.md`: the files lamsza shares with the other apps.
- `lamsza/docs/AGENT_ENVIRONMENT_POLICY.md`: **localhost only, never touch production.**
- `lamsza/docs/LOCAL_DEV_CHECKS.md`: what "verified" means. Run the checks your
  change touches *before* pushing, because CI is not the real gate.
- `lamsza/docs/network/UI_BASELINE.md`: the owner's UI decisions, one per item; the
  reference for any UI change in any app. The Cursor rule `.cursor/rules/lamsza-network.mdc`
  points to it.
- Every repo's `CHANGELOG.md`: a notable change adds a line under `[Unreleased]` in the
  same commit (WoW R17).
