# Architecture & Deployment Guide

## Core Architecture
- **Frontend**: SvelteKit, statically pre-rendered with `@sveltejs/adapter-static` into the `dist` directory.
- **Backend**: Go (1.25.7) compiling to a single binary `sz-gugel-bin`.
- **Database**: PostgreSQL.

## Environment Variables
The application relies on a `.env` file at the root of the project. This file is **not** committed to version control.

### Required Keys:
- `PORT`: The port the Go backend listens on (default: `3000`).
- `DATABASE_URL`: The PostgreSQL connection string. (e.g. `postgres://lamsza_user:lamsza_password@localhost:5433/lamsza?sslmode=disable`)
- `WEATHER_API_KEY`: OpenWeatherMap API key. Backend only. The browser gets weather from `/api/weather`.
- `FEATURE_[WEATHER|EVENTS|NEWS|MONDASOK|QUICKLINKS|SEARCH]`: Toggles for optional modules (default: `true`).

### Never put a secret in a `VITE_*` variable
Vite replaces every `VITE_*` name at build time, so the value is plain text inside the
public JavaScript bundle and any visitor can read it. Secrets belong in plain, unprefixed
names that only the Go backend reads.

`VITE_API_BASE_URL` and `VITE_WEATHER_API_KEY` were removed. The browser reaches the API
same-origin through the Vite dev proxy (`vite.config.js`) or the production reverse proxy.
Set the server-only `API_BASE_URL` if a prerender must reach the backend on another port.

## Continuous Integration (CI)
We use GitHub Actions for continuous integration.
- **Workflow**: `.github/workflows/build-test.yml`
- **Trigger**: Runs on every `push`.
- **Function**: Verifies that both the SvelteKit frontend and Go backend compile successfully without errors.
- **Note**: This workflow does **not** deploy code to the DigitalOcean server. It compiles the frontend and backend. It does not run `go test` or `node --test`.

The public product version lives in `src/lib/publicChangelog.js`. The footer and `/valtozasnaplo` read it. `package.json` stays the private npm package version and is not that number.

## Local Development
1. Ensure your `.env` is configured.
2. Start the backend: `cd backend && go run main.go`
3. Start the frontend: `npm run dev`
