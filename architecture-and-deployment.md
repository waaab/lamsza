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
- `CORS_ALLOWED_ORIGINS`: Comma-separated list of the exact browser origins that may call the
  API cross-origin, for example
  `https://lamsza.com,https://www.lamsza.com,https://szotar.lamsza.com,https://jatszoter.lamsza.com`.
  It **replaces** the built-in list, so set it in production to drop the local development hosts.
  See "Cross-origin calls" below.

## Cross-origin calls (CORS)

`backend/internal/middleware/middleware.go` answers a cross-origin call only when the `Origin`
header matches the allowlist exactly, and it sends `Access-Control-Allow-Credentials` only then.

Never echo the request `Origin` back. All four sites sit under one registrable domain, so the
session cookie is same-site for every sibling and `SameSite=Lax` does not hold it back: a
reflected `Origin` plus `Allow-Credentials` let *any* page read the signed-in reply of an admin
who happened to visit it, which is enough to take the account over.

The allowlist doubles as the CSRF guard. A page cannot forge the `Origin` header, so a POST, PUT,
PATCH or DELETE that carries an `Origin` we do not know is refused with `403`. Clients outside a
browser (curl, the deployment scripts) send no `Origin` and are unaffected.

The default list covers the four `*.lamsza.com` sites plus the `*.lamsza.test` and `localhost`
development hosts. Behaviour is covered in `backend/internal/middleware/middleware_test.go`.

The only expected cross-origin caller today is Szótár and Játszótér reading
`https://lamsza.com/api/config/public`. Everything else reaches the API same-origin through the
Vite dev proxy or the production reverse proxy.

## Content-Security-Policy

`svelte.config.js` sets `kit.csp`. The pages are static, so SvelteKit writes the policy into a
`<meta http-equiv="content-security-policy">` tag in every built page and hashes its own inline
scripts for us.

`script-src 'self' https://accounts.google.com/gsi/client` is the part that matters. It is the
second lock on untrusted Markdown (`src/lib/markdown.js` is the first) and `connect-src` names
the only hosts a page may talk to, so even an injected script could not post a stolen session
anywhere. Keep both lists as short as the app allows; `tests/csp.test.js` fails if `script-src`
ever gains `unsafe-inline` or `unsafe-eval`.

Two consequences to keep in mind:

- **No inline `<script>` in `src/app.html`.** The theme preload lives in `static/theme-init.js`
  and is loaded with a `src` attribute. Add new boot scripts the same way.
- **`frame-ancestors` cannot work from a `meta` tag.** The `X-Frame-Options` header that Nginx
  sends stays the control for framing. The same goes for `report-uri` and `sandbox`.

SvelteKit does not apply the policy in `vite dev`, so HMR is unaffected; test the policy against
a real build.

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
