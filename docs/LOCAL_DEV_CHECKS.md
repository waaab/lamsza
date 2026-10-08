# Local Dev Checks — bring the four apps up and prove they work

**Audience:** anyone developing the network on this machine, agents included.
**Target:** `localhost` only. See `docs/AGENT_ENVIRONMENT_POLICY.md`.
**Covers:** starting the network, smoking every API, and the feature checks that are easy
to get wrong (sign-in config, admin allowlists, weather, the Szótár↔Játszótér pair,
restart and data persistence, dump and restore).
Commands and what to expect, not recorded results: run them and compare (WAYS_OF_WORKING R10).
**Never write to dev data while checking** (R8, R9): every check here is a read, a GET or a
test suite on a scratch database.

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

> **All four apps answer `/api/health` with `{"ok":true,"db":"up"}`.** The route pings
> Postgres and returns `503 {"ok":false,"db":"down"}` when the database does not answer,
> so a process that is up with a dead DB reads as down.

---

## Check — Google sign-in

Google sign-in works on the `localhost` ports of every app (`http://localhost:5173` to
`5176`), not on `*.lamsza.test`: Google rejects `.test` origins, so they cannot be in the
OAuth client. What to check:

```bash
curl -s http://127.0.0.1:3001/api/config/public | head -c 300   # lamsza
curl -s http://127.0.0.1:3000/api/config/public | head -c 300   # admin
curl -s http://127.0.0.1:3002/api/config/public | head -c 300   # szotar
curl -s http://127.0.0.1:3003/api/config/public | head -c 300   # jatszoter
```

**Expect:** a non-empty `google_client_id` in every payload. All four frontends read it from
there at runtime; none bakes it in at build time. Then open each app on its `localhost`
port and confirm the **Google button renders** in the sign-in dialog. A missing button with
an empty `google_client_id` means the backend's `GOOGLE_CLIENT_ID` is not set. A button
that renders but fails at Google is an origin-registration problem in the Google console,
not a code bug. Signing in is the owner's to do (R9); agents stop at the button.

---

## Check — Admin allowlists

Only two apps keep an admin list, both named `ADMIN_GOOGLE_EMAILS`. Getting the admin one
wrong is the usual cause of "admin locked out of their own site":

| App | Variable | What it decides |
|---|---|---|
| admin | `ADMIN_GOOGLE_EMAILS` | Who gets into the admin app, all three sections: the network's one admin list (WoW R18) |
| lamsza | `ADMIN_GOOGLE_EMAILS` | Only the `is_admin` flag and the review-queue count in lamsza's `/api/auth/me`; it grants no admin rights |

Szótár and Játszótér have no admin list: they trust only the admin backend, through
`ADMIN_SERVICE_TOKEN` (WoW R18). A leftover `ADMIN_GOOGLE_EMAILS` or `ADMIN_EMAILS` in
their `.env` is read by nothing.

`.env` lives at the repo root in lamsza; admin reads both `.env` and `backend/.env`.

```bash
grep -H '^ADMIN' ~/projects/lamsza-network/{lamsza,lamsza-admin}/.env \
  ~/projects/lamsza-network/lamsza-admin/backend/.env 2>/dev/null
```

What you verify locally is the **gate logic**, not any particular email list:

```bash
cd ~/projects/lamsza-network/lamsza-admin && scripts/test-backend.sh ./internal/auth/...
```

**Expect:** pass. These cover allowlist matching (case-insensitivity, separators).

Where the gate is checked in a browser differs per app, because the **main app has no
`/admin` route** — it was extracted to `lamsza-admin` (see `docs/ADMIN_EXTRACTION.md`):

| App | Admin UI |
|---|---|
| admin (`lamsza-admin`) | `http://localhost:5173/`, `/dictionary` (Szótár), `/games` (Játszótér): the network's one admin UI (WoW R18) |
| szotar | none. `http://localhost:5175/admin` shows the 404 error page |
| jatszoter | none. `http://localhost:5176/admin` shows the 404 error page |
| lamsza | **none.** `http://localhost:5174/admin` must 404 |

An allowlisted account must reach the admin UI; a non-allowlisted one is sent on to
Lámsza, and Szótár's and Játszótér's admin writes under `/api/` answer 403 to anyone. If a non-admin reaches an admin UI locally, stop — that
is a security bug, not a config issue.

---

## Check — The admin session is separate from the public one

`lamsza` and `lamsza-admin` share one database. They do **not** share a session (BOG-45):
the public site uses cookie `lamsza_session` and table `sessions`, admin uses
`lamsza_admin_session` and `admin_sessions`.

This matters most on localhost, where both apps are host `localhost` and **cookies ignore
the port** — a public cookie is offered to `:3000` whatever the ports are.

Prove it with the two boundary suites. They mint real session rows (a random token whose
SHA-256 is the row, the way both backends do it) for a fixture user, but only in a scratch
database, never in the dev `lamsza` database and never for the owner's account (R8, R9):

```bash
cd ~/projects/lamsza-network/lamsza       && scripts/test-go.sh -run 'AdminSession|PublicSession' -v .
cd ~/projects/lamsza-network/lamsza-admin && scripts/test-backend.sh -run 'Session' -v ./internal/auth/...
```

**Expect:** both pass. Between them they check that

- a public token is refused by admin, under the public cookie name *and* renamed to the
  admin one, so neither the browser nor a deliberate caller gets through;
- an admin token works on admin (the sign-in path is intact);
- and that same admin token is refused by the public API.

A failure there means the two session stores have been merged back together. The suites
are `lamsza/backend/admin_session_boundary_test.go` and
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

**Expect:** both `200` with live data, e.g. `"source":"Open-Meteo"`, and a county array of
towns.

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

With the pair working, the payload reads `"dictionary":{"ok":true,"source":"http",...}` with
Szótár's word count.

If `source` is `local` and you want the pair, set in `~/projects/lamsza-network/lamsza-jatszoter/.env`:

```
DICTIONARY_SOURCE=http
SZOTAR_BASE_URL=http://127.0.0.1:3002
```

**Failure shapes:** `ok:false` with `source:"http"` → Szótár is down or the port is wrong.
`word_count: 0` with `source:"local"` → `DICTIONARY_DATA_DIR` points nowhere.

---

## Check — Tájszórejtvény (Játszótér)

The game reads Szótár's mondások server to server, so it never repeats the daily mondás within 14 days. That
needs one token pair, the same value on both sides. Without it, Játszótér uses its own proverb list and says
nothing else is wrong:

| App | Variable | |
|---|---|---|
| szotar | `GAMES_SERVICE_TOKEN` | opens `GET /internal/games/proverbs` (loopback, no proxy headers, Bearer); empty → 404 |
| jatszoter | `SZOTAR_GAMES_TOKEN` | the same value; the base URL is `SZOTAR_BASE_URL` |

Smokes, GET only (WoW R9). The game row ships switched off, so until it is on (admin `/games`, Tájszórejtvény,
"Bekapcsolás"):

```bash
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:3003/api/games/tajszorejtveny/week        # 404 while off
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:3003/internal/admin/stats               # 401: no token
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:3002/internal/games/proverbs            # 401 with the token set, 404 without
```

When it is on, `week` answers 200 with `today` (Bucharest), the week's days and which are playable; `daily/<today>`
200 once the week's puzzles exist; `daily/<a future day>` 404. Next week's puzzles are made on Wednesday at
22:00 and go live on Monday at 0:00; the admin's `/games` tab shows them and has "Generálás most". Writes (play,
reports, admin actions) are verified with the Go suites or a throwaway backend on a scratch database, never
against the dev data.

---

## Check — Restart and data persistence

Prove the apps come back cleanly after a full stop, and that no data lives only in memory.

```bash
# record the counts
docker exec -i szotar-db    psql -U szotar_user -d szotar    -c 'select count(*) from words;'
docker exec -i jatszoter-db psql -U jatszoter_user -d jatszoter -c 'select count(*) from users;'

# full restart of the apps, through the launcher (WAYS_OF_WORKING R12)
cd ~/projects/lamsza-network && ./start-lamsza-network.sh stop
# the one sanctioned container step: restart (never `down`) one database container, to
# prove its data lives in the named volume and not in the container
docker restart szotar-db
./start-lamsza-network.sh start

# same counts?
docker exec -i szotar-db    psql -U szotar_user -d szotar    -c 'select count(*) from words;'
docker exec -i jatszoter-db psql -U jatszoter_user -d jatszoter -c 'select count(*) from users;'
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

docker exec -i szotar-db pg_dump -U szotar_user -Fc szotar > /tmp/lamsza-backup-test/szotar-$STAMP.dump
ls -lh /tmp/lamsza-backup-test/

docker exec -i szotar-db createdb -U szotar_user szotar_restore_test
docker exec -i szotar-db pg_restore -U szotar_user -d szotar_restore_test < /tmp/lamsza-backup-test/szotar-$STAMP.dump

docker exec -i szotar-db psql -U szotar_user -d szotar              -c 'select count(*) from words;'
docker exec -i szotar-db psql -U szotar_user -d szotar_restore_test -c 'select count(*) from words;'

docker exec -i szotar-db dropdb -U szotar_user szotar_restore_test
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
packages with "TEST_DATABASE_URL is not set" (admin's DB-bound tests skip instead)
instead of writing to dev data, which is what it used to do: on 2026-10-07, 43 of
the 46 users in the dev `lamsza` database were `*@test.lamsza` fixtures.

> `node --test tests/` (directory form) crashes with `MODULE_NOT_FOUND` on Node 24.
> Use the glob `'tests/*.test.js'`.

Every frontend suite (lamsza, admin, szotar, jatszoter) includes
`tests/sharedFrontendModules.test.js`, the drift guard for the frontend files that
`lamsza` shares with the other apps (listed in `shared-frontend-modules.json`, with a
`consumers` map of who gets what). `lamsza` owns them and
`scripts/sync-shared-frontend.sh` copies them over; each repo checks its own copies
against its own committed hash manifest, so a drifted module goes red in whichever
repo drifted. Do **not** keep the copies in step by hand: run the script. With all
four repos checked out, `scripts/sync-shared-frontend.sh --check` verifies them in
one go, and `scripts/tests/sync-shared-frontend.test.sh` tests the script itself.
The rule is `docs/network/SHARED_FRONTEND_MODULES.md`.

Every frontend suite also runs `tests/noEmdash.test.js` (shared the same way): no em dash
in code or UI text anywhere in the repo, Markdown exempt (`.cursor/rules/no-emdash.mdc`).

> The lamsza, szotar and jatszoter Go suites also run in CI against a `postgres:16`
> service (since BOG-53); CI sets `TEST_DATABASE_URL` to that service database.
> Admin's CI has no Postgres service, so its DB-bound tests skip there and run only
> locally (`OPEN_ITEMS.md`). The local scripts above run
> the suites the way CI does: on an empty database built from the repo. A suite
> that passes on your dev data and fails on an empty database is the failure CI
> used to be blind to.

---

## What a complete local pass looks like

1. Eight listeners up, three DB containers up.
2. Every API in Step 2 returns `200`.
3. The feature checks above verified, with the output pasted into the task, not summarised.
4. Restart test: counts unchanged, everything back without manual help.
5. Dump and restore: matching counts.
6. All test suites green.
