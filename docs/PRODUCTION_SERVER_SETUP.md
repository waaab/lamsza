# Lamsza Network — Production Server Setup (DigitalOcean)

**Audience:** Project manager / ops helper setting up the production server  
**Owner apps (now):** lamsza.com, szotar.lamsza.com, jatszoter.lamsza.com  
**Hosting target:** DigitalOcean Droplet, Ubuntu 24.04 LTS, Nginx  
**Last updated:** 2026-10-04

This document is the single handoff checklist for production server setup and configuration. It covers what must exist before go-live, and what to install/configure on the droplet.

---

## 1. Network map

| Hostname | Purpose | Deploy now? | Repo |
|----------|---------|-------------|------|
| `lamsza.com` | Main startlap / directory / search | **Yes** | `lamsza` |
| `szotar.lamsza.com` | Székely dictionary | **Yes** | `szotar` |
| `jatszoter.lamsza.com` | Word games | **Yes** | `jatszoter` |
| `www.lamsza.com` | Optional alias → redirect to apex | Optional | — |
| `admin.lamsza.com` | Network admin dashboard | No (not started) | — |
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
- [ ] Admin email allowlists decided for each app

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
| `WEATHER_API_KEY` | Yes* | OpenWeather-style key |
| `WEATHER_API_COM_KEY` | Optional | Alternate weather provider |
| `VITE_API_BASE_URL` | Yes (build-time) | Public API base URL for frontend build |
| `VITE_WEATHER_API_KEY` | If used by FE build | Build-time |
| `FEATURE_WEATHER` | Optional | Default `true` |
| `FEATURE_EVENTS` | Optional | Default `true` |
| `FEATURE_NEWS` | Optional | Default `true` |
| `FEATURE_MONDASOK` | Optional | Default `true` |
| `FEATURE_QUICKLINKS` | Optional | Default `true` |
| `FEATURE_SEARCH` | Optional | Default `true` |
| `DATA_API` | Optional | Default `true` |
| `SZOTAR_ORIGIN` | Optional | Base URL for Szótár API, no trailing slash. Local `https://szotar.lamsza.test`, prod `https://szotar.lamsza.com`. Unset skips dictionary hits in `/api/search`. |

### 6.2 szotar.lamsza.com

| Variable | Required | Notes |
|----------|----------|-------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `PORT` | Yes | `3010` |
| `GOOGLE_CLIENT_ID` | Yes | Google OAuth |
| `ADMIN_GOOGLE_EMAILS` | Yes | Comma-separated admin emails |

Also needed for content: dictionary import source / dump for initial seed (`szotar_db1` import path).

### 6.3 jatszoter.lamsza.com

| Variable | Required | Notes |
|----------|----------|-------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `PORT` | Yes | `3001` |
| `GOOGLE_CLIENT_ID` | Yes | Backend |
| `VITE_GOOGLE_CLIENT_ID` | Yes (build-time) | Same value as `GOOGLE_CLIENT_ID` |
| `SESSION_SECRET` | Yes | Long random prod secret |
| `ADMIN_EMAILS` | Yes | Comma-separated admin emails |
| `DICTIONARY_SOURCE` | Yes | `local` until szotar is live; then can switch |
| `DICTIONARY_DATA_DIR` | If local | e.g. `data/dictionary` |
| `SZOTAR_BASE_URL` | Yes when not local | e.g. `https://szotar.lamsza.com` or `http://127.0.0.1:3010` |
| `APP_VERSION` | Optional | Display / config version |

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
4. **Szótár CORS** — current code allowlists local Vite (`localhost:5174`). Prefer same-origin Nginx proxy (`https://szotar.lamsza.com/api` → `:3010`) so browsers do not need cross-origin calls. If the frontend calls the API on another origin, CORS must be updated.
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
| lamsza | 5173 | 3000 | 5433 |
| szotar | 5174 | 3010 | 5435 |
| jatszoter | 5175 | 3001 | 5434 |

---

*Source of truth for network rules: `jatszoter/docs/lamsza-network.md`. App stack details: each repo’s README / config.*
