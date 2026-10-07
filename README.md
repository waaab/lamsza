# Lamsza Platform

A high-performance, culturally authentic startpage and directory for the Szekely region.

## Overview
Lamsza is a full-stack platform: a local directory of services and websites, bilingual search, weather, news, events, and a daily quote ("mondás"). Signed-in people have a Fiók page, settings, and favorite places. The published product version is the first entry in `src/lib/publicChangelog.js`. The footer and `/valtozasnaplo` both read that list.

## Repo layout
This repo has **no `frontend/` folder**. The Go backend is in `backend/`; the
SvelteKit frontend lives at the repo root (`src/`, `static/`, `tests/`,
`svelte.config.js`, `vite.config.js`, root `package.json`), and `npm run build`
writes to `dist/`.

The other three network apps (`lamsza-admin`, `lamsza-szotar`, `lamsza-jatszoter`) do use
`backend/` + `frontend/`. The difference is deliberate — see §5 of
[docs/network/WAYS_OF_WORKING.md](docs/network/WAYS_OF_WORKING.md). Do not create
a `frontend/` folder here.

## Tech Stack
- **Frontend**: SvelteKit (Vanilla CSS & JS)
- **Backend**: Go (net/http, PostgreSQL)
- **Database**: PostgreSQL (with Full-Text Search)
- **Deployment**: Docker Compose

## Quick Start
1. Configure your `.env` file in the repo root (`DATABASE_URL`, `PORT`, `GOOGLE_CLIENT_ID`, ...).
2. Start the whole network (all four apps and their databases) from the parent folder:
   ```bash
   ~/projects/lamsza-network/start-lamsza-network.sh start|stop|restart|status
   ```
   `stop` stops the apps and leaves the databases running. Logs and PID files are in
   `~/.cache/lamsza-network/`. Never kill other apps' Vite servers or run
   `docker compose down` here: `lamsza-db` is shared with lamsza-admin.

## Development
- Tests: `npm test` (frontend) and `npm run test:go` (backend, on a scratch database);
  see `docs/LOCAL_DEV_CHECKS.md`.

## Architecture
See [docs/history/architecture_overview.md](docs/history/architecture_overview.md) for an older, partly outdated breakdown.
