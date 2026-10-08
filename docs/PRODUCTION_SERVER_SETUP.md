# Lamsza Network — Production Server Setup (DigitalOcean)

**Audience:** Project manager / ops helper setting up the production server  
**Owner apps (now):** lamsza.com, szotar.lamsza.com, jatszoter.lamsza.com  
**Hosting target:** DigitalOcean Droplet, Ubuntu 24.04 LTS, Nginx  
**Last updated:** 2026-10-06

This document is the single handoff checklist for production server setup and configuration. It covers what must exist before go-live, and what to install/configure on the droplet.

---

## 1. Network map

| Hostname | Purpose | Deploy now? | Repo |
|----------|---------|-------------|------|
| `lamsza.com` | Main startlap / directory / search | **Yes** | `lamsza` |
| `szotar.lamsza.com` | Székely dictionary | **Yes** | `szotar` |
| `jatszoter.lamsza.com` | Word games | **Yes** | `jatszoter` |
| `www.lamsza.com` | Optional alias → redirect to apex | Optional | — |
| `admin.lamsza.com` | Main Lámsza admin app (`lamsza-admin`) | **Phase 2** (DNS/Nginx exist; app deploy later) | `lamsza-admin` |
| `api.lamsza.com` | Shared API gateway | No (TBD) | — |
| `account.lamsza.com` | Central accounts | No (TBD) | — |
| `mink.lamsza.com` | Private social | No (planned) | — |
| `bolt.lamsza.com` | Merch shop | No (planned) | — |

**Rule:** Each live property is a **separate deploy** (own frontend build, own Go backend, own PostgreSQL database), sharing one droplet and one Nginx reverse proxy.

---

## 2. Architecture (production shape)

```
Internet
   │
   ▼
Nginx (TLS :443)
   ├─ lamsza.com          → static files (dist/) + /api → Go :3000
   ├─ szotar.lamsza.com   → static files (dist/) + /api → Go :3010
   └─ jatszoter.lamsza.com→ static files (dist/) + /api → Go :3001
                              │
                              ▼
                    PostgreSQL 16 (3 databases)
```

| Layer | Choice |
|-------|--------|
| OS | Ubuntu 24.04 LTS |
| Web server | Nginx |
| TLS | Let’s Encrypt (Certbot) |
| Frontend | SvelteKit **static** build (`dist/`) — no Node process in prod |
| Backend | Go binary per app (systemd) |
| Database | PostgreSQL 16 (Docker or native) — **one DB per app** |
| CI | GitHub Actions builds/tests only — **does not deploy** |

---

## 3. Prerequisites before touching the server

Complete these offline / in accounts first.

### 3.1 DigitalOcean

- [ ] DO account + billing active
- [ ] Droplet created: **Ubuntu 24.04 LTS**
- [ ] SSH key access working
- [ ] Recommended size: **4 GB RAM** (minimum 2 GB if budget-constrained)
- [ ] Optional: reserved IP, automated backups, DO Cloud Firewall

### 3.2 DNS (must resolve before certificates)

Point these to the droplet IP:

- [ ] `A` / `AAAA` — `lamsza.com`
- [ ] `A` / `AAAA` — `szotar.lamsza.com`
- [ ] `A` / `AAAA` — `jatszoter.lamsza.com`
- [ ] Optional: `www.lamsza.com` → same IP (Nginx will redirect to apex)

### 3.3 Google OAuth (login / admin)

- [ ] Google Cloud OAuth **Web client** ID(s) ready
- [ ] Authorized JavaScript origins include:
  - `https://lamsza.com`
  - `https://szotar.lamsza.com`
  - `https://jatszoter.lamsza.com`
  - `https://www.lamsza.com` (if used)
  - `https://admin.lamsza.com` — **add this now, not at cutover.** Admin is
    Phase 2 (§1), but its whole UI is behind Google sign-in, so a missing origin
    means nobody can log in on day one. Registering an origin for a host that is
    not live yet costs nothing and breaks nothing.
- [ ] Admin email list decided: admin's `ADMIN_GOOGLE_EMAILS` (the only one with admin rights) and lamsza's (§6.1)

### 3.4 External API keys

- [ ] Weather provider key(s) for lamsza (`WEATHER_API_KEY` and/or `WEATHER_API_COM_KEY`)

### 3.5 Secrets to generate (do not reuse local/dev values)

- [ ] Strong Postgres passwords (3× — one per database user)
- [ ] `SESSION_SECRET` for jatszoter (long random string)
- [ ] Confirm Google Client IDs for prod (may be same client for all three if origins are listed)

### 3.6 Source access

- [ ] Git access on the droplet for repos: `lamsza`, `szotar`, `jatszoter`  
  (deploy key or HTTPS credentials)

---

## 4. Software to install on the droplet

| Package / tool | Purpose |
|----------------|---------|
| `nginx` | Reverse proxy + static hosting |
| `certbot` + nginx plugin | HTTPS certificates |
| Docker **or** PostgreSQL 16 | Databases |
| Go **1.25.7+** | Compile / run backends |
| Node.js **20+** + npm | Build frontends (`vite build`) |
| `git` | Pull code |
| `ufw` (or DO firewall) | Restrict public ports to 22 / 80 / 443 |

Public firewall should **not** expose Postgres or Go backend ports.

---

## 5. Port and database plan

Backends listen on **localhost only**. Nginx is the only public HTTP(S) entry.

| App | Hostname | Go port | DB name | DB user | Local compose DB port (dev reference) |
|-----|----------|---------|---------|---------|----------------------------------------|
| Lamsza | `lamsza.com` | `3000` | `lamsza` | `lamsza_user` | `5433` |
| Játszótér | `jatszoter.lamsza.com` | `3001` | `jatszoter` | `jatszoter_user` | `5434` |
| Szótár | `szotar.lamsza.com` | `3010` | `szotar` | `szotar_user` | `5435` |

**Decision needed:** one Postgres instance with three databases, or three containers (as in local docker-compose). Either is fine; passwords must be unique and strong.

---

## 6. Environment variables (prod `.env` per app)

Each app needs its own `.env` on the server (never commit these).

### 6.1 lamsza.com

| Variable | Required | Notes |
|----------|----------|-------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `PORT` | Yes | `3000` |
| `GOOGLE_CLIENT_ID` | Yes | Google OAuth |
| `ADMIN_GOOGLE_EMAILS` | Yes | Comma-separated admin emails |
| `METNO_USER_AGENT` | Yes | Identifies us to MET Norway, the main weather source: app name plus a contact (site or email), e.g. `lamsza.com weather (+https://lamsza.com; <contact email>)`. Their terms require it. Default `lamsza.com weather (+https://lamsza.com)` |
| `WEATHER_API_KEY` | Optional | OpenWeatherMap key: second fallback, and the geocoder for settlements with no coordinates yet |
| `WEATHER_API_COM_KEY` | Optional | WeatherAPI.com key: first fallback |
| `WEATHER_WORKER` | Optional | Default `true`. The background refresh that fills the weather cache and archive; `false` on a throwaway server |
| `API_BASE_URL` | Optional (build-time) | Backend URL for prerender only. Server-only name, never in the bundle. Default `http://127.0.0.1:3001` |
| `FEATURE_WEATHER` | Optional | Default `true` |
| `FEATURE_EVENTS` | Optional | Default `true` |
| `FEATURE_NEWS` | Optional | Default `true` |
| `FEATURE_QUICKLINKS` | Optional | Default `true` |
| `FEATURE_SEARCH` | Optional | Default `true` |
| `DATA_API` | Optional | Default `true` |
| `SZOTAR_ORIGIN` | Optional | Szótár's backend for the server-side word search, no trailing slash. Server-to-server, so `http://127.0.0.1:<szotar port>` (WAYS_OF_WORKING R13): locally `http://127.0.0.1:3002`; on the server, Szótár's backend port there (to verify, `OPEN_ITEMS.md`). Unset skips dictionary hits in `/api/search`. |
| `ACCOUNT_SZOTAR_TOKEN` | Yes, with the account release | Opens the internal account API (`/internal/account/profile`, the user's display name by Google ID) to Szótár's backend. Long random secret, equal to Szótár's `LAMSZA_ACCOUNT_TOKEN` (§6.2). Unset (with the next one unset too) turns the API off. |
| `ACCOUNT_JATSZOTER_TOKEN` | Yes, with the account release | The same for Játszótér's backend; equal to Játszótér's `LAMSZA_ACCOUNT_TOKEN` (§6.3), a different value from Szótár's. |
| `DB_MAX_OPEN_CONNS` | Optional | Default `25`. Pool cap. The `lamsza` database is shared with admin and Postgres allows 100 connections in total, so keep lamsza plus admin under that. |
| `DB_MAX_IDLE_CONNS` | Optional | Default `10`. Must not be above `DB_MAX_OPEN_CONNS`; the code lowers it if it is. |

### 6.2 szotar.lamsza.com

| Variable | Required | Notes |
|----------|----------|-------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `PORT` | Yes | `3010` |
| `GOOGLE_CLIENT_ID` | Yes | Google OAuth |
| `ADMIN_SERVICE_TOKEN` | Yes | Opens the internal admin API to the admin app (WAYS_OF_WORKING R18). Long random secret, equal to admin's `SZOTAR_ADMIN_TOKEN` (§6.4). Unset turns the API off. Szótár has no admin list of its own. |
| `LAMSZA_ACCOUNT_URL` | Yes, with the account release | Lámsza's backend, server to server: `http://127.0.0.1:<lamsza port>` (WAYS_OF_WORKING R13). |
| `LAMSZA_ACCOUNT_TOKEN` | Yes, with the account release | Equal to Lámsza's `ACCOUNT_SZOTAR_TOKEN` (§6.1). Szótár copies the user's display name from Lámsza at sign-in and on Fiók. Unset: Google's name is shown. |
| `GAMES_SERVICE_TOKEN` | Yes, before Tájszórejtvény goes live | Opens `/internal/games/` (the full mondás list, future days included) to Játszótér's backend. Long random secret, equal to Játszótér's `SZOTAR_GAMES_TOKEN`; a different value from `ADMIN_SERVICE_TOKEN`. Unset turns it off, and Tájszórejtvény uses only its own proverb list. |

Also needed for content: dictionary import source / dump for initial seed (`szotar_db1` import path).

### 6.3 jatszoter.lamsza.com

| Variable | Required | Notes |
|----------|----------|-------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `PORT` | Yes | `3001` |
| `GOOGLE_CLIENT_ID` | Yes | Backend |
| `VITE_GOOGLE_CLIENT_ID` | Yes (build-time) | Same value as `GOOGLE_CLIENT_ID` |
| `SESSION_SECRET` | Yes | Long random prod secret |
| `ADMIN_SERVICE_TOKEN` | Yes | Opens the internal admin API to the admin app (WAYS_OF_WORKING R18). Long random secret, equal to admin's `JATSZOTER_ADMIN_TOKEN` (§6.4). Unset turns the API off. Játszótér has no admin list of its own. |
| `DICTIONARY_SOURCE` | Yes | `local` until szotar is live; then can switch |
| `DICTIONARY_DATA_DIR` | If local | e.g. `data/dictionary` |
| `SZOTAR_BASE_URL` | Yes when not local | e.g. `https://szotar.lamsza.com` or `http://127.0.0.1:3010` |
| `SZOTAR_GAMES_TOKEN` | Yes, before Tájszórejtvény goes live | Equal to Szótár's `GAMES_SERVICE_TOKEN` (§6.2): Tájszórejtvény reads every mondás with its date from `SZOTAR_BASE_URL/internal/games/proverbs`. Unset: only the game's own proverb list is used. |
| `LAMSZA_ACCOUNT_URL` | Yes, with the account release | Lámsza's backend, server to server: `http://127.0.0.1:<lamsza port>` (WAYS_OF_WORKING R13). |
| `LAMSZA_ACCOUNT_TOKEN` | Yes, with the account release | Equal to Lámsza's `ACCOUNT_JATSZOTER_TOKEN` (§6.1). Játszótér copies the user's display name from Lámsza at sign-in and on Fiók, and the leaderboards show it. Unset: Google's name is shown. |
| `APP_VERSION` | Optional | Display / config version |

### 6.4 admin.lamsza.com (Phase 2)

The full list, with comments, is `lamsza-admin/.env.example`. The ones that tie it to the other apps:

| Variable | Required | Notes |
|----------|----------|-------|
| `ADMIN_GOOGLE_EMAILS` | Yes | Comma-separated admin emails: the network's one admin list (R18). Unset falls back to the owner's address. |
| `SZOTAR_ADMIN_URL` | Yes | Szótár's backend, server-to-server: `http://127.0.0.1:<szotar port>` |
| `SZOTAR_ADMIN_TOKEN` | Yes | Equal to Szótár's `ADMIN_SERVICE_TOKEN`. Unset turns `/dictionary` off (503). |
| `JATSZOTER_ADMIN_URL` | Yes | Játszótér's backend, server-to-server: `http://127.0.0.1:<jatszoter port>` |
| `JATSZOTER_ADMIN_TOKEN` | Yes | Equal to Játszótér's `ADMIN_SERVICE_TOKEN`. Unset turns `/games` off (503). |

Generate each token with `openssl rand -hex 32`, one per pair. nginx must not proxy `/internal/` on any vhost; the steps are in `docs/network/OPEN_ITEMS.md`, the production upgrade item.

**Important:** All `VITE_*` variables are baked in at **frontend build time**. Changing them requires a rebuild + redeploy of `dist/`.

**Network origins (frontend build):** Optional `VITE_LAMSZA_ORIGIN`, `VITE_SZOTAR_ORIGIN`, `VITE_JATSZOTER_ORIGIN`. Leave unset in production so links default to `https://*.lamsza.com`. Local `.test` hostnames auto-select `https://*.lamsza.test` without these vars.

---

## 7. Suggested filesystem layout

```text
/var/www/lamsza/
  dist/          # frontend static build
  bin/           # Go binary
  .env

/var/www/szotar/
  dist/
  bin/
  .env

/var/www/jatszoter/
  dist/
  bin/
  .env
```

Use systemd units such as `lamsza-api.service`, `szotar-api.service`, `jatszoter-api.service` to keep backends running and restarting on boot.

---

## 8. Nginx responsibilities (per hostname)

For each of the three hosts:

1. Serve HTTPS (Certbot)
2. Serve static files from that app’s `dist/`
3. SPA fallback to `app.html` for client routes
4. Proxy `/api` (and auth-related API paths) to the local Go port
5. Do **not** expose Go ports publicly

Optional:

- Redirect `http` → `https`
- Redirect `www.lamsza.com` → `lamsza.com`

For `admin.lamsza.com` (Phase 2), the vhost must additionally send

```nginx
add_header X-Robots-Tag "noindex, nofollow, noarchive" always;
```

Admin has no public page and no application layer to add the header itself —
Nginx serves `dist/` as static files. `robots.txt` in that app is already
`Disallow: /`, but that only asks a crawler not to fetch; it does not keep a
URL someone linked out of an index. The ready-to-copy vhost is in
`lamsza-admin/docs/ARCHITECTURE.md`, "Production (Phase 2)".

---

## 9. Setup order (recommended)

1. Create droplet + SSH harden + firewall (22/80/443)
2. Install Nginx, Docker/Postgres, Go, Node, Certbot, Git
3. Create three databases + users + strong passwords
4. Clone repos (or copy release artifacts)
5. Place prod `.env` files
6. Run DB migrations / seeds (per app)
7. Build Go binaries; build frontends with prod `VITE_*` values
8. Install systemd units; start APIs; confirm localhost health
9. Configure Nginx vhosts; obtain Certbot certificates
10. Smoke-test each hostname (home page + `/api` health/login paths)
11. Confirm Google login origins work on all three sites
12. Enable backups for Postgres volumes/data

---

## 10. Go-live checklist

- [ ] `https://lamsza.com` loads (static + API)
- [ ] `https://szotar.lamsza.com` loads (static + API)
- [ ] `https://jatszoter.lamsza.com` loads (static + API)
- [ ] Valid TLS on all three (no browser warnings)
- [ ] HTTP redirects to HTTPS
- [ ] Google sign-in works where enabled
- [ ] Admin allowlists correct
- [ ] Weather widget works on lamsza
- [ ] Játszótér can reach dictionary (`local` or szotar URL)
- [ ] Backends survive reboot (systemd)
- [ ] Postgres data persists across reboot
- [ ] Firewall blocks DB/API ports from the public internet
- [ ] Backup method documented and tested once

---

## 11. Known gotchas for the setup person

1. **No auto-deploy** — GitHub Actions does not push to DigitalOcean; deploy is manual (or a script you add later).
2. **Static frontends** — Node is for building only; Nginx serves `dist/`.
3. **Build-time env** — `VITE_*` must be set when running `npm run build`, not only in the backend `.env`.
4. **Szótár CORS** — the code allowlists the local frontends and the `*.lamsza.com` origins (`CORS_ALLOWED_ORIGINS` in szotar's `.env.example`). Prefer same-origin Nginx proxy (`https://szotar.lamsza.com/api` → `:3010`) so browsers do not need cross-origin calls. If the frontend calls the API on another origin, CORS must be updated.
5. **Do not reuse local compose passwords** in production.
6. **Three separate databases** — do not share one DB across apps.
7. Future subdomains (`admin`, `api`, `account`, etc.) are out of scope for this first production pass.

---

## 12. Contacts / ownership (fill in)

| Role | Name | Notes |
|------|------|-------|
| Product / owner | | |
| Server setup (PM/ops) | | |
| App deploy / secrets | | |
| DNS access | | |
| Google Cloud project | | |
| DigitalOcean account | | |

---

## 13. Quick reference — default local ports (dev only)

Useful when comparing to local docker-compose; **not** for public exposure.

| App | Frontend (Vite) | Backend | Postgres |
|-----|-----------------|---------|----------|
| admin (`lamsza-admin`) | 5173 | 3000 | shared with lamsza (:5433) |
| lamsza | 5174 | 3001 | 5433 |
| szotar | 5175 | 3002 | 5435 |
| jatszoter | 5176 | 3003 | 5434 |

---

*Source of truth for network rules: `lamsza/docs/network/WAYS_OF_WORKING.md`; for the UI, `lamsza/docs/network/UI_BASELINE.md`. App stack details: each repo’s README / config. The production ports in this file are to be verified on the server (`docs/network/OPEN_ITEMS.md`).*
