# Local Dev Checks — bring the four apps up and prove they work

**Audience:** anyone developing the network on this machine, agents included.
**Target:** `localhost` only. See `docs/AGENT_ENVIRONMENT_POLICY.md`.
**Covers:** starting the network, smoking every API, and the feature checks that are easy
to get wrong (sign-in config, admin allowlists, weather, the Szótár↔Játszótér pair,
restart and data persistence, dump and restore).
**Last updated:** 2026-10-06 — every command below was run and the output recorded.

All four apps are in development. This is the check to run after a change, before calling
it done.

---

## Local port map

From `scripts/start-lamsza-network.sh` in this repo (the source of truth):

| App | Backend | Frontend | Database |
|---|---|---|---|
| admin (`lamsza-admin`) | `3000` | `5173` | shares `lamsza-db` |
| lamsza | `3001` | `5174` | `lamsza-db` `:5433` |
| szotar | `3002` | `5175` | `szotar-db` `:5435` |
| jatszoter | `3003` | `5176` | `jatszoter-db` `:5434` |

> `tests/frontend-test-checklist.md` used to say backend `3131` and frontend `5173` for
> lamsza. Those were pre-renumber values; it now points at this table. This table stays the
> source of truth.

---

## Step 1 — Bring the whole network up

```bash
cd ~/projects/lamsza-network
./start-lamsza-network.sh start      # start | stop | restart | status
```

The script is tracked at `lamsza/scripts/start-lamsza-network.sh`.
`~/projects/lamsza-network/start-lamsza-network.sh` and `~/.local/bin/lamsza-network` are
symlinks into it, so either path, and `lamsza-network start` from anywhere, run
the same tracked file. Change it in the repo, not through a symlink.

Because the symlink points into the working tree, it dangles — `No such file or
directory` — while the `lamsza` checkout is parked on a branch cut before the
script was committed (BOG-50, `5fc03cc`). Any branch cut after that has it. If
you hit it, run the script from a worktree that is on `main`:

```bash
~/projects/lamsza-network/lamsza/.worktrees/<some-branch-on-main>/scripts/start-lamsza-network.sh status
```

Do **not** `git checkout main -- scripts/start-lamsza-network.sh` into someone
else's branch to paper over it: that leaves an uncommitted file in a checkout
another run is using, and their task then fails the "committed" check (WoW R5)
for dirt that is not theirs.

**Expect:** four backends and four frontends listening. Check it yourself rather than
trusting the script's own summary:

```bash
ss -ltn | grep -E ':(3000|3001|3002|3003|5173|5174|5175|5176)'
docker ps --format '{{.Names}}\t{{.Status}}'     # lamsza-db, szotar-db, jatszoter-db
```

**Expect:** eight listeners, three DB containers `Up`.

**Known issues with the script:**

| Symptom | Cause |
|---|---|
| `rg: command not found`, every app reported `down` while it is actually up | the script uses `ripgrep` for its status check. The apps are fine; the *report* is wrong. Verify with the `ss` command above. |
| `cd: …/projects/lamsza-network/lamsza-admin/backend: No such file or directory` | `$HOME` is not `/home/attila` in this shell. Run with `LAMSZA_PROJECTS_ROOT=/home/attila/projects/lamsza-network` |
| a backend never appears | first run compiles Go and downloads modules — jatszoter pulls a large dependency tree. Give it a few minutes and read `${XDG_CACHE_HOME:-$HOME/.cache}/lamsza-network/<app>-backend.log` |

---

## Creating a lamsza database from scratch

Only needed on a new machine, or when you want a throwaway database to test
against. The existing `lamsza-db` container already has one.

```bash
cd ~/projects/lamsza-network/lamsza
scripts/db-bootstrap.sh --create                                  # the dev database
scripts/db-bootstrap.sh --create --url "postgres://lamsza_user:lamsza_password@localhost:5433/lamsza_scratch?sslmode=disable"
```

It applies the two committed files in `backend/schema/` — the full schema, then
the reference rows (counties, settlements, venue types, weather translations).
Postgres 16, as local dev and CI both run. The script works with
`postgresql-client` on `PATH` or, as on this machine, by going through the
`lamsza-db` container; it picks whichever is available.

This is the same script and the same two files the CI backend job runs, so a
green CI means this path works. **`backend/migrations/` is not that path** — its
41 files are a hand-applied historical record with no runnable order. Do not try
to build a database out of them.

After a schema change, regenerate and commit:

```bash
scripts/db-dump-schema.sh
```

See `backend/schema/README.md` for what is in each file and which tables are
deliberately not dumped (anything with personal data or configuration).

---

## Step 2 — Smoke every API

```bash
probe(){ printf '%-48s %s\n' "$1" "$(curl -s -o /dev/null -w '%{http_code}' --max-time 6 "$1")"; }

probe http://127.0.0.1:3000/api/health                 # admin
probe http://127.0.0.1:3001/api/health                 # lamsza
probe http://127.0.0.1:3002/api/health                 # szotar
probe http://127.0.0.1:3002/api/words
probe http://127.0.0.1:3003/api/health                 # jatszoter
probe http://127.0.0.1:3003/api/config/public
for p in 5173 5174 5175 5176; do probe http://127.0.0.1:$p/; done
```

**Expect:** `200` everywhere.

> **All four apps answer `/api/health`** since 2026-10-06. On lamsza and admin the route
> pings Postgres and returns `503 {"ok":false,"db":"down"}` when the database does not
> answer, so a process that is up with a dead DB reads as down. szotar and jatszoter return
> a plain `{"ok":true}`.

---

## Check — Google sign-in

Sign-in cannot be fully proved on localhost — Google OAuth needs the dev origins
registered. What you **can** prove:

```bash
curl -s http://127.0.0.1:3002/api/config/public | head -c 300   # szotar
curl -s http://127.0.0.1:3003/api/config/public | head -c 300   # jatszoter
```

**Expect:** the payload reports a Google client id is configured (non-empty), not an empty
string. Then, in a browser at `http://localhost:5175` and `http://localhost:5176`, confirm
the **Google button renders**. A missing button means `VITE_GOOGLE_CLIENT_ID` was empty at
build time — a code/config bug the agent fixes. A button that renders but fails at Google
is an origin-registration problem in the Google console, not a code bug.

`http://localhost:<port>` must be in the dev OAuth client's authorized origins for the
button to complete locally.

---

## Check — Admin allowlists

The variable name differs per app. Getting this wrong is the usual cause of "admin locked
out of their own site":

| App | Variable |
|---|---|
| lamsza | `ADMIN_GOOGLE_EMAILS` |
| szotar | `ADMIN_GOOGLE_EMAILS` |
| admin | `ADMIN_GOOGLE_EMAILS` |
| jatszoter | `ADMIN_EMAILS` |

```bash
grep -H '^ADMIN' ~/projects/lamsza-network/{lamsza,lamsza-szotar,lamsza-jatszoter,lamsza-admin}/backend/.env 2>/dev/null
```

What you verify locally is the **gate logic**, not any particular email list:

```bash
cd ~/projects/lamsza-network/lamsza-szotar && scripts/test-backend.sh ./internal/auth/...
cd ~/projects/lamsza-network/lamsza-admin  && scripts/test-backend.sh ./internal/auth/...
```

**Expect:** pass. These cover allowlist matching (case-insensitivity, separators).

Where the gate is checked in a browser differs per app, because the **main app has no
`/admin` route** — it was extracted to `lamsza-admin` (see `docs/ADMIN_EXTRACTION.md`):

| App | Admin UI |
|---|---|
| admin (`lamsza-admin`) | `http://localhost:5173/` — the whole app is the admin UI |
| szotar | `http://localhost:5175/admin` |
| jatszoter | `http://localhost:5176/admin` |
| lamsza | **none.** `http://localhost:5174/admin` must 404 |

An allowlisted account must reach the admin UI of the app it belongs to; a
non-allowlisted one must get 403. If a non-admin reaches an admin UI locally, stop — that
is a security bug, not a config issue.

---

## Check — The admin session is separate from the public one

`lamsza` and `lamsza-admin` share one database. They do **not** share a session (BOG-45):
the public site uses cookie `lamsza_session` and table `sessions`, admin uses
`lamsza_admin_session` and `admin_sessions`.

This matters most on localhost, where both apps are host `localhost` and **cookies ignore
the port** — a public cookie is offered to `:3000` whatever the ports are.

Google sign-in cannot complete from a script, so mint the session rows the way the two
backends do — a session *is* a random token whose SHA-256 is a row. Both apps hash the
same way, so this is the real thing, not a stand-in.

```bash
psql(){ docker exec -i lamsza-db psql -U lamsza_user -d lamsza "$@"; }

# the admin session table exists — the main lamsza backend creates it; admin runs no DDL
psql -c '\d admin_sessions'

# an account on ADMIN_GOOGLE_EMAILS, so the only thing under test is the session store
EMAIL=$(grep -h '^ADMIN_GOOGLE_EMAILS' ~/projects/lamsza-network/lamsza-admin/backend/.env 2>/dev/null \
        | head -1 | cut -d= -f2 | cut -d, -f1)
USERID=$(psql -qtAc "SELECT id FROM users WHERE lower(email) = lower('$EMAIL')")
echo "admin user: $EMAIL -> id $USERID"

mint(){ # mint <table> -> prints the raw token
  local t h; t=$(openssl rand -hex 32); h=$(printf %s "$t" | sha256sum | cut -d' ' -f1)
  psql -qtAc "INSERT INTO $1 (token_hash, user_id, expires_at)
              VALUES ('$h', $USERID, NOW() + INTERVAL '1 hour')" >/dev/null
  echo "$t"
}
PUB=$(mint sessions); ADM=$(mint admin_sessions)
code(){ curl -s -o /dev/null -w '%{http_code}' --max-time 6 -H "Cookie: $2" "$1"; }

printf 'public token -> admin /api/auth/me        %s\n' "$(code http://127.0.0.1:3000/api/auth/me     "lamsza_session=$PUB")"
printf 'public token -> admin /api/admin/users    %s\n' "$(code http://127.0.0.1:3000/api/admin/users "lamsza_admin_session=$PUB")"
printf 'admin token  -> admin /api/admin/users    %s\n' "$(code http://127.0.0.1:3000/api/admin/users "lamsza_admin_session=$ADM")"
printf 'admin token  -> public /api/auth/me       %s\n' "$(code http://127.0.0.1:3001/api/auth/me     "lamsza_session=$ADM")"

psql -qc "DELETE FROM sessions WHERE expires_at < NOW() + INTERVAL '2 hours'" \
        -qc "DELETE FROM admin_sessions WHERE expires_at < NOW() + INTERVAL '2 hours'"
```

**Expect** `401`, `401`, `200`, `401`. In words:

- a public token is refused by admin — under the public cookie name *and* renamed to the
  admin one, so neither the browser nor a deliberate caller gets through;
- an admin token works on admin (the sign-in path is intact);
- and that same admin token is refused by the public API.

A `200` on any of the three `401` lines means the two session stores have been merged back
together. The `\d admin_sessions` failing means the table is missing: run the main `lamsza`
backend, or `lamsza/backend/migrations/admin_sessions.sql`.

The suites covering this: `lamsza/backend/admin_session_boundary_test.go` and
`lamsza-admin/backend/internal/auth/session_boundary_test.go`.

---

## Check — Weather

Both endpoints need a `slug`:

```bash
curl -s 'http://127.0.0.1:3001/api/weather?slug=csikszereda' | head -c 300
curl -s 'http://127.0.0.1:3001/api/weather/county?slug=hargita' | head -c 300
```

> `/api/weather` with **no** `slug` returns `400 Missing slug`. That is correct behaviour,
> not a failure — pass a slug.

**Verified 2026-10-06:** both `200` with live data from Open-Meteo —
`{"temp":15,"desc":"Vannak felhők es...","source":"Open-Meteo",...}` and a county array
including Tusnádfürdő and Gyergyószentmiklós.

Then open `http://localhost:5174` and confirm the widget **renders** that temperature.
API data with a blank widget is a frontend bug.

---

## Check — Játszótér reaches the dictionary

```bash
curl -s http://127.0.0.1:3003/api/config/public
```

**Expect:** `dictionary.ok: true`, and read `source`:

- `source: "http"` — Játszótér is really calling Szótár. This is the pair that matters.
- `source: "local"` — bundled word files. `ok:true` here says nothing about reaching Szótár.

**Verified 2026-10-06:** `{"dictionary":{"ok":true,"source":"http","word_count":475},"version":"0.1.0"}`
— the live pair works locally, against Szótár on `:3002`.

If `source` is `local` and you want the pair, set in `~/projects/lamsza-network/lamsza-jatszoter/backend/.env`:

```
DICTIONARY_SOURCE=http
SZOTAR_BASE_URL=http://127.0.0.1:3002
```

**Failure shapes:** `ok:false` with `source:"http"` → Szótár is down or the port is wrong.
`word_count: 0` with `source:"local"` → `DICTIONARY_DATA_DIR` points nowhere.

---

## Check — Restart and data persistence

Prove the apps come back cleanly after a full stop, and that no data lives only in memory.

```bash
# record the counts
docker exec -i szotar-db    psql -U postgres -d szotar    -c 'select count(*) from words;'
docker exec -i jatszoter-db psql -U postgres -d jatszoter -c 'select count(*) from users;'

# full restart of apps and containers
cd ~/projects/lamsza-network && ./start-lamsza-network.sh stop
docker compose -f ~/projects/lamsza-network/lamsza-szotar/docker-compose.yml restart
./start-lamsza-network.sh start

# same counts?
docker exec -i szotar-db    psql -U postgres -d szotar    -c 'select count(*) from words;'
docker exec -i jatszoter-db psql -U postgres -d jatszoter -c 'select count(*) from users;'
```

**Expect:** identical counts, and every app back to `200` in Step 2 with no manual fixing.

**The thing to actually look for:** each `docker-compose.yml` must mount a **named volume
or bind mount** for Postgres data. A container with no volume loses everything on restart.
Check it in the compose file, not by hoping.

---

## Check — Backup and restore

Prove the dump/restore *commands* work, so nobody is debugging `pg_restore` syntax on the
day they need it.

```bash
STAMP=$(date +%F-%H%M)
mkdir -p /tmp/lamsza-backup-test

docker exec -i szotar-db pg_dump -U postgres -Fc szotar > /tmp/lamsza-backup-test/szotar-$STAMP.dump
ls -lh /tmp/lamsza-backup-test/

docker exec -i szotar-db createdb -U postgres szotar_restore_test
docker exec -i szotar-db pg_restore -U postgres -d szotar_restore_test < /tmp/lamsza-backup-test/szotar-$STAMP.dump

docker exec -i szotar-db psql -U postgres -d szotar              -c 'select count(*) from words;'
docker exec -i szotar-db psql -U postgres -d szotar_restore_test -c 'select count(*) from words;'

docker exec -i szotar-db dropdb -U postgres szotar_restore_test
```

**Expect:** a dump of real size (not a few hundred bytes), `pg_restore` with no errors, and
two identical counts. Missing-role errors are usually harmless (`--no-owner`); missing
tables or constraint violations mean the dump is bad.

---

## Step 3 — The test suites (run these on every change)

```bash
# lamsza frontend (node:test)
cd ~/projects/lamsza-network/lamsza && npm test

# lamsza backend: drops and recreates the scratch database lamsza_test from
# backend/schema/, then runs every Go package against it
cd ~/projects/lamsza-network/lamsza && npm run test:go

# the other three: frontend suite, then backend suite on a scratch database
for d in lamsza-szotar lamsza-jatszoter lamsza-admin; do
  (cd ~/projects/lamsza-network/$d && npm test)
done
```

**No test suite touches a dev database.** In all four repos the Go suites read
only `TEST_DATABASE_URL`, never `DATABASE_URL` or `.env`, and refuse a local
database whose name does not end in `_test` (`backend/internal/db/testguard.go`).
The scripts behind the commands above drop and recreate the scratch database on
every run:

| Repo | Command | Scratch database | Schema from |
|---|---|---|---|
| lamsza | `npm run test:go` (`scripts/test-go.sh`) | `lamsza_test` on `lamsza-db` | `backend/schema/` via `scripts/db-bootstrap.sh` |
| admin | `npm run test:backend` (`scripts/test-backend.sh`) | `lamsza_admin_test` on `lamsza-db` | lamsza's `backend/schema/`, same script |
| szotar | `npm run test:backend` (`scripts/test-backend.sh`) | `szotar_test` on `szotar-db` | the backend's own `db.Migrate()` |
| jatszoter | `npm run test:backend` (`scripts/test-backend.sh`) | `jatszoter_test` on `jatszoter-db` | the backend's own `db.Migrate()` |

Each script passes extra arguments to `go test`, e.g. `scripts/test-backend.sh
./internal/auth/...`. A plain `go test ./...` in `backend/` fails the DB-bound
packages with "TEST_DATABASE_URL is not set" (admin skips its two DB tests)
instead of writing to dev data, which is what it used to do: on 2026-10-07, 43 of
the 46 users in the dev `lamsza` database were `*@test.lamsza` fixtures.

> `node --test tests/` (directory form) crashes with `MODULE_NOT_FOUND` on Node 24.
> Use the glob `'tests/*.test.js'`.

The lamsza suite includes `tests/sharedFrontendModules.test.js`, the drift guard for
the 17 frontend files that `lamsza` and `lamsza-admin` share. `lamsza` owns them and
`scripts/sync-shared-frontend.sh` copies them over; each repo checks its own copies
against the committed hash manifest, so a drifted module goes red in whichever repo
drifted. Do **not** keep the copies in step by hand — run the script. The rule is
`docs/network/SHARED_FRONTEND_MODULES.md`.

**Status 2026-10-06:**

| Suite | Result |
|---|---|
| lamsza frontend | **151 / 153** — 2 failing: `entryHistory.test.js:45`, `searchResultCard.test.js:54` |
| lamsza backend | **2 failing** in `directory_catalog_test.go` (123, 223) — both look like dirty-DB test isolation, not product bugs |
| szotar backend | pass |
| jatszoter backend | pass |
| lamsza-admin backend | pass |

**Update 2026-10-06, BOG-42:** the table above is stale. Measured on the BOG-42
branch: lamsza frontend **197 / 197**, lamsza backend green. The two
`directory_catalog_test.go` failures were dirty-DB isolation caused by the admin
CRUD tests in the same package; those tests went with the admin handlers they
covered. The two frontend failures were fixed by their own task.

> Three of these suites talk to Postgres, and since BOG-53 all three also run in
> CI against a `postgres:16` service; nothing is excluded by name any more. CI
> sets `TEST_DATABASE_URL` to that service database. The local scripts above run
> the suites the way CI does: on an empty database built from the repo. A suite
> that passes on your dev data and fails on an empty database is the failure CI
> used to be blind to.

`tests/networkOrigins.test.js` was failing against stale pre-renumber ports (5173/5174/5175)
and is fixed. The remaining failures are tracked as their own task.

---

## What a complete local pass looks like

1. Eight listeners up, three DB containers up.
2. Every API in Step 2 returns `200`.
3. The feature checks above verified, with the output pasted into the task, not summarised.
4. Restart test: counts unchanged, everything back without manual help.
5. Dump and restore: matching counts.
6. All test suites green.
