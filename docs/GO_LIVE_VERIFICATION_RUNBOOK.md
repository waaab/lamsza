# Go-Live Verification Runbook — the 7 checks that need server or browser access

**Audience: the owner (`attila`) only.**
**Target:** DigitalOcean droplet `146.190.204.232` (FRA1, Ubuntu 24.04)
**Covers:** items 6, 7, 8, 9, 10, 11 and 13 of the go-live checklist in `docs/PRODUCTION_SERVER_SETUP.md` section 10
**Last updated:** 2026-10-06

> **Agents must not run any command in this document.**
> Agents work only on the local machine. They never get SSH to the droplet, never
> change production config, DNS or domains, and never commit or merge to `main`.
> The owner runs every step here and every deploy. See `docs/AGENT_ENVIRONMENT_POLICY.md`.
>
> The agent-owned companion is `docs/LOCAL_VERIFICATION_RUNBOOK.md`: it proves the same
> behaviour on localhost *before* the owner deploys, so each item below is a confirmation,
> not a discovery.

Each step gives the exact command, the expected result, and what a failure looks like.
Record the real output next to each item. "Looked fine" is not evidence.

**Port note:** this runbook uses the *verified* production ports from
`docs/PRODUCTION_ENVIRONMENT_NOTES.md` — szotar `8081`, jatszoter `8082`, admin `8083`.
`docs/PRODUCTION_SERVER_SETUP.md` section 2 still lists the old dev ports (3000/3010/3001).
That doc is stale; do not use those numbers on the server.

---

## Step 0 (do this first) — the live APIs are down

Checked from outside on 2026-10-06 08:27 UTC:

| URL | Result |
|-----|--------|
| `https://szotar.lamsza.com/` | 200 (static files serve) |
| `https://szotar.lamsza.com/api/health` | **502 Bad Gateway** (nginx) |
| `https://szotar.lamsza.com/api/words` | **502** |
| `https://jatszoter.lamsza.com/` | 200 |
| `https://jatszoter.lamsza.com/api/health` | **502** |
| `https://jatszoter.lamsza.com/api/config/public` | **502** |
| `https://admin.lamsza.com/api/health` | **502** (Phase 2, expected) |

> **Caveat found locally on 2026-10-06:** `/api/health` **does not exist** in the lamsza or
> admin backends — they return `404`, not `200`, even when perfectly healthy. Only szotar
> and jatszoter implement it. So a 404 from `lamsza.com/api/health` or
> `admin.lamsza.com/api/health` is a missing route, not a down service. Use
> `/api/config/public` as the liveness probe for those two until the route is added.
> The 502s above are still real — a 502 is nginx failing to reach the backend at all.

Three retries, same result — not a blip. Nginx and TLS are fine; the Go backends behind
them are not answering. Users get a page shell with no data.

Items 6, 7 and 9 cannot pass until this is fixed, so fix it first.

```bash
ssh attila@146.190.204.232

# 1. What does systemd think?
systemctl status szotar jatszoter --no-pager

# 2. Why did they stop?
journalctl -u szotar -n 80 --no-pager
journalctl -u jatszoter -n 80 --no-pager

# 3. Is anything listening on the proxy targets?
ss -ltnp | grep -E ':(8081|8082|8083)'

# 4. Bring them back
sudo systemctl restart szotar jatszoter
sleep 3
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8081/api/health
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8082/api/health
```

**Expect:** `active (running)` for both, a listener on 8081 and 8082, `200` from both curls.
Then re-check from outside: `curl -s -o /dev/null -w '%{http_code}\n' https://szotar.lamsza.com/api/health` → `200`.

**Failure shapes and what they mean:**

| Symptom | Cause | Fix |
|---|---|---|
| `inactive (dead)` and no recent logs | never started after a reboot | `sudo systemctl enable --now szotar jatszoter` (also closes item 10) |
| `activating (auto-restart)`, log shows `connection refused` to `127.0.0.1:5432` | Postgres is down | `sudo systemctl status postgresql`, restart it, then the apps |
| log shows `FATAL: password authentication failed` | `.env` `DATABASE_URL` wrong | fix `/var/www/<app>/.env`, keep mode `0600` |
| listener is on 3010/3001, not 8081/8082 | `.env` `PORT` does not match the nginx `proxy_pass` | set `PORT` to match the table above, restart |
| no binary at `/var/www/szotar/szotar` | deploy never landed | redeploy per `PRODUCTION_ENVIRONMENT_NOTES.md` section 6 |

**Evidence to record:** the `systemctl status` lines, the root-cause log line, and the
three `200`s after the fix.

---

## Item 6 — Google sign-in works

Needs: a browser, plus the production OAuth client in Google Cloud. Needs Step 0 done.

1. Open `https://szotar.lamsza.com` in a normal (not incognito-blocked) window. Click sign in.
2. **Expect:** Google's account chooser appears, you return to the site signed in, and your
   name/avatar shows.
3. Confirm the server agrees — in the same browser, open
   `https://szotar.lamsza.com/api/auth/me`.
   **Expect:** JSON with your email. Not `401`.
4. Repeat 1–3 on `https://jatszoter.lamsza.com`.

**Failure shapes:**

| Symptom | Cause | Fix |
|---|---|---|
| Google error `redirect_uri_mismatch` / `origin_mismatch` | the prod origin is not registered | Google Cloud Console → Credentials → the prod OAuth client → add `https://szotar.lamsza.com` and `https://jatszoter.lamsza.com` to **Authorized JavaScript origins** |
| no Google button at all on Játszótér | `VITE_GOOGLE_CLIENT_ID` was empty at build time | rebuild the frontend with the var set, redeploy `public/`. `VITE_*` is baked in at build time — a server-side change does nothing |
| button appears, sign-in returns 500 | backend `GOOGLE_CLIENT_ID` missing | set it in `/var/www/<app>/.env`, restart the service |
| signs in, then logged out on reload | `SESSION_SECRET` changes per restart, or cookie not `Secure` on HTTPS | set a fixed long `SESSION_SECRET` in `.env` |

**Evidence:** a screenshot of the signed-in header, plus the `/api/auth/me` JSON body.

---

## Item 7 — Admin allowlists correct

Needs: server access (the lists live in `.env`), then a browser to confirm.
The variable name differs per app — this is an easy thing to get wrong:

| App | Variable |
|---|---|
| lamsza | `ADMIN_GOOGLE_EMAILS` |
| szotar | `ADMIN_GOOGLE_EMAILS` |
| jatszoter | `ADMIN_EMAILS` |

```bash
sudo grep -H '^ADMIN' /var/www/szotar/.env /var/www/jatszoter/.env
# if the main app is deployed:
sudo grep -H '^ADMIN' /var/www/lamsza/.env
```

**Expect:** a comma-separated list containing the owner's real Google account email, and
nobody who should not be an admin. Matching is case-insensitive, so case does not matter.
Watch for: stray spaces are tolerated, but a trailing comma, quotes around the value, or a
semicolon separator are not.

Then confirm it actually gates:

1. Signed in as an allowlisted account, open `https://szotar.lamsza.com/admin`
   and `https://jatszoter.lamsza.com/admin/rovasfejto`.
   **Expect:** the admin UI loads.
2. Signed in as (or logged out to) a non-allowlisted account, open the same URLs.
   **Expect:** refused — 403 or a "not authorized" screen. **Not** the admin UI.

**Failure shapes:** admin UI loads for a non-admin → the gate is not applied, stop and
treat as a security bug. Admin UI refuses the owner → check for a typo and that the email
in the list is the same Google account you signed in with (a `@gmail.com` vs custom-domain
mix-up is the usual cause).

**Evidence:** the `grep` output with emails, plus the result of both browser checks.

---

## Item 8 — Weather widget works on lamsza

**This item cannot be closed yet, and the reason is not the weather code.**

`lamsza.com` does not point at the droplet. As of 2026-10-06 it serves a **301 to
`https://www.lamsza.com`** with `server: Squarespace`. The main app is not live on its own
domain at all. (`PRODUCTION_ENVIRONMENT_NOTES.md` says it is on Google AppEngine — that is
also out of date.)

So: either the DNS cutover for `lamsza.com` happens first, or this item stays open.
Do not tick it from a localhost test; the checklist item is about the live domain.

After the cutover, verify like this:

```bash
# on the server
sudo grep -H -E '^(WEATHER_API_KEY|WEATHER_API_COM_KEY|FEATURE_WEATHER)' /var/www/lamsza/.env
curl -s 'http://127.0.0.1:8080/api/weather?slug=csikszereda' | head -c 400
curl -s 'http://127.0.0.1:8080/api/weather/county?slug=hargita' | head -c 400
# from outside
curl -s 'https://lamsza.com/api/weather?slug=csikszereda' | head -c 400
```

> Both endpoints **require** `?slug=`. Without it you get `400 Missing slug` — that is the
> endpoint working, not a failure. Verified on localhost 2026-10-06.

**Expect:** HTTP 200 and JSON with a real current temperature and a location name.
Then open `https://lamsza.com` and confirm the widget shows that same temperature —
the API passing is not enough, the widget has to render.

**Failure shapes:** `502` → backend down (Step 0). `500` → `WEATHER_API_KEY` missing.
`200` but empty/null values → the provider rejected the key (new OpenWeather keys take
up to a couple of hours to activate) or the key is for the wrong product tier.
Widget blank while the API returns data → frontend build issue, check `VITE_*` weather vars
were set at build time.

---

## Item 9 — Játszótér can reach the dictionary

One request tells you everything, once Step 0 is done:

```bash
curl -s https://jatszoter.lamsza.com/api/config/public
```

**Expect:** `{"dictionary":{"ok":true,"source":"...","word_count":<a number well above 0>},...}`

Read `source` carefully — it is the decision point:

- `source: "local"` — Játszótér is reading its own bundled word files. `ok:true` here does
  **not** mean it can reach Szótár. This is an acceptable go-live state (it is the documented
  default), but tick the item as "local, by choice" and say so.
- `source: "http"` — it is calling Szótár over HTTP. This is the real pair.

To prove the live request pair (the thing the checklist is asking for), do both halves on
the server:

```bash
# half 1: Szótár answers the endpoint Játszótér uses
curl -s http://127.0.0.1:8081/api/words | head -c 300

# half 2: Játszótér reports the dictionary healthy over that link
curl -s http://127.0.0.1:8082/api/config/public
```

To switch Játszótér from `local` to the live Szótár, set in `/var/www/jatszoter/.env`:

```
DICTIONARY_SOURCE=http
SZOTAR_BASE_URL=http://127.0.0.1:8081
```

then `sudo systemctl restart jatszoter` and re-run the curl above.
Use the loopback address, not `https://szotar.lamsza.com` — it avoids a public round trip,
TLS cost and a CORS question, and it keeps working if DNS breaks.

**Failure shapes:** `ok:false` with `source:"http"` → Szótár is down or the port is wrong.
`word_count: 0` with `source:"local"` → `DICTIONARY_DATA_DIR` points nowhere; the data files
were not deployed. Nothing at all → Step 0.

**Evidence:** both JSON bodies, verbatim.

---

## Items 10 + 11 — Backends and Postgres survive a reboot

**These two need a scheduled reboot window.** `szotar.lamsza.com`,
`jatszoter.lamsza.com` and `admin.lamsza.com` are live and serving real users, so announce
the window before you reboot. Expect roughly 60–90 seconds of downtime on a 2 GB droplet.

### Before the reboot

```bash
# item 10 — will they come back by themselves?
systemctl is-enabled szotar jatszoter postgresql nginx

# item 11 — record the numbers you will compare against
sudo -u postgres psql -d szotar -c 'select count(*) from words;'
sudo -u postgres psql -d szotar -c 'select count(*) from definitions;'
sudo -u postgres psql -d jatszoter -c 'select count(*) from users;'
sudo -u postgres psql -d jatszoter -c 'select count(*) from game_results;'

# and prove the data lives on disk, not in a container that will vanish
sudo -u postgres psql -c 'show data_directory;'
```

**Expect:** `enabled` four times, four counts you write down, and a data directory under
`/var/lib/postgresql/16/main` (native install). If Postgres runs in Docker instead, check
the compose file uses a **named volume or bind mount** — a container with no volume loses
everything on reboot, which is exactly what item 11 is there to catch.

If any service says `disabled`:

```bash
sudo systemctl enable szotar jatszoter postgresql nginx
```

### The reboot

```bash
sudo reboot
```

### After the reboot (wait ~90s, then reconnect)

```bash
ssh attila@146.190.204.232
uptime                       # confirm it really rebooted
systemctl status szotar jatszoter postgresql nginx --no-pager | grep -E 'Active:|●'

# item 10 passes when these are 200 with no manual intervention
curl -s -o /dev/null -w 'szotar %{http_code}\n'    https://szotar.lamsza.com/api/health
curl -s -o /dev/null -w 'jatszoter %{http_code}\n' https://jatszoter.lamsza.com/api/health

# item 11 passes when every count matches what you wrote down
sudo -u postgres psql -d szotar -c 'select count(*) from words;'
sudo -u postgres psql -d szotar -c 'select count(*) from definitions;'
sudo -u postgres psql -d jatszoter -c 'select count(*) from users;'
sudo -u postgres psql -d jatszoter -c 'select count(*) from game_results;'
```

**Item 10 fails if** you had to start anything by hand. Touching a service to get the 200
means the item does not pass — fix the `enable`, reboot again.
**Item 11 fails if** any count dropped, or a table is missing. Stop and do not go live;
that is data loss, and item 13 is the only thing that can save you.

**Evidence:** both sets of counts side by side, and the post-reboot `Active: active (running)`
lines with no manual restart in between.

---

## Item 13 — Backup method documented and tested once

A backup you have never restored is not a backup. This item is the restore, not the dump.

```bash
sudo mkdir -p /var/backups/lamsza && sudo chown attila:attila /var/backups/lamsza
STAMP=$(date +%F-%H%M)

# 1. Dump both live databases
sudo -u postgres pg_dump -Fc szotar    > /var/backups/lamsza/szotar-$STAMP.dump
sudo -u postgres pg_dump -Fc jatszoter > /var/backups/lamsza/jatszoter-$STAMP.dump
ls -lh /var/backups/lamsza/

# 2. Restore into a scratch database — this is the part that counts
sudo -u postgres createdb szotar_restore_test
sudo -u postgres pg_restore -d szotar_restore_test /var/backups/lamsza/szotar-$STAMP.dump

# 3. Compare the restored copy against live
sudo -u postgres psql -d szotar              -c 'select count(*) from words;'
sudo -u postgres psql -d szotar_restore_test -c 'select count(*) from words;'

# 4. Clean up the scratch database (leave the dumps)
sudo -u postgres dropdb szotar_restore_test
```

**Expect:** both dump files non-trivial in size (not a few hundred bytes), `pg_restore`
finishing with no errors, and identical `words` counts in step 3.

**Then get the dump off the droplet.** A backup sitting on the machine it protects dies with
that machine. Pick one and do it now, not later:

```bash
# simplest: pull it to your laptop
scp attila@146.190.204.232:/var/backups/lamsza/szotar-*.dump ~/lamsza-backups/
```

or enable DigitalOcean droplet backups in the control panel (weekly snapshots, paid), or
push to DO Spaces with a cron job. Whichever you choose, write it into
`PRODUCTION_SERVER_SETUP.md` — the checklist item says "documented **and** tested".

**Failure shapes:** `pg_restore` errors about missing roles are usually harmless
(`--no-owner` fixes them); errors about missing tables or constraint violations are not —
the dump is bad. A dump of a few hundred bytes means `pg_dump` hit the wrong database.

**Evidence:** `ls -lh` of the dumps, the two matching counts, and one sentence naming where
the off-server copy now lives.

---

## Summary of what blocks what

| Item | Blocked by |
|---|---|
| 6 Google sign-in | Step 0 (APIs down) + prod OAuth origins registered |
| 7 Admin allowlists | server access; Step 0 for the browser half |
| 8 Weather on lamsza | **`lamsza.com` DNS cutover** — it is on Squarespace today |
| 9 Játszótér → dictionary | Step 0 |
| 10 Reboot survival | SSH + announced reboot window |
| 11 Postgres persistence | same reboot window as item 10 |
| 13 Backup tested | SSH + somewhere off-server to keep the dump |

Items 10, 11 and 13 should be done in one session: record counts → dump → reboot →
verify → restore test. One announced window, three items closed.
