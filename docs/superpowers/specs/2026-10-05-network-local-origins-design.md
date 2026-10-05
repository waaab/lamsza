# Network local origins (`.test` vs `.com`)

**Date:** 2026-10-05  
**Status:** Approved  
**Apps:** Szótár, Játszótér (Lámsza as network source of truth for the helper; adopt when it gains sister-app links)

## Problem

Satellite apps hardcode `https://lamsza.com` (toolbar, policies, SignInDialog, social-config fetch) and sometimes `https://jatszoter.lamsza.com`. Locally we use `*.lamsza.test` with HTTPS. Those static links send developers to production.

Google OAuth still cannot use `.test` origins; that stays on `localhost` ports and is out of scope.

## Goals

- Cross-app and policy links resolve to local `.test` hosts when developing on `.test`.
- Production builds keep `https://*.lamsza.com` with no accidental `.test` URLs.
- Always use `https://` (never match `http:`).
- Optional env override for explicit control.
- One small helper copied into each app (same pattern as other Lámsza network shared bits).

## Non-goals

- Changing Google Authorized JS origins.
- Rewriting marketing/FAQ copy that merely mentions “lamsza.com”.
- Publishing local nginx / mkcert config into the repos.
- A shared npm package; copy the helper file per app.

## Resolution rules

For each network app key (`lamsza`, `szotar`, `jatszoter`):

1. **Env override** - if `VITE_<APP>_ORIGIN` is set and non-empty, use it (trim trailing `/`).  
   Names: `VITE_LAMSZA_ORIGIN`, `VITE_SZOTAR_ORIGIN`, `VITE_JATSZOTER_ORIGIN`.
2. **Hostname sniff** - if `typeof window !== 'undefined'` and hostname is `lamsza.test` or ends with `.lamsza.test`, use:
   - `https://lamsza.test`
   - `https://szotar.lamsza.test`
   - `https://jatszoter.lamsza.test`
3. **Production default** - otherwise:
   - `https://lamsza.com`
   - `https://szotar.lamsza.com`
   - `https://jatszoter.lamsza.com`

SSR / prerender: pass the request hostname (`$page.url.hostname`) into `lamszaUrl(path, hostname)` (and siblings). Without a hostname, skip step 2 and use env if set, else production defaults. Static production builds on `.com` hosts stay on `.com`.

## API

File (identical in each adopting app): `src/lib/networkOrigins.js`  
(Játszótér / Szótár: under `frontend/src/lib/networkOrigins.js`.)

```js
export function lamszaOrigin() {}
export function szotarOrigin() {}
export function jatszoterOrigin() {}

/** origin + path; path must start with `/` or be empty */
export function lamszaUrl(path = '') {}
export function szotarUrl(path = '') {}
export function jatszoterUrl(path = '') {}
```

Callers replace string literals, e.g. `lamszaUrl('/iranyelvek')`, `jatszoterUrl(\`/jatszok/${slug}\`)`.

## Call sites to update

### Szótár (`szotar/frontend`)

- `src/routes/+layout.svelte` - Lámsza toolbar link; footer Irányelvek / Feltételek / Sütik; social-config fetch URL
- `src/lib/components/SignInDialog.svelte` - policy href
- `src/lib/games.js` - Játszótér game URLs

### Játszótér (`jatszoter/frontend`)

- `src/routes/+layout.svelte` - Lámsza toolbar link; footer policies; social-config fetch URL
- `src/lib/components/SignInDialog.svelte` - policy href
- `src/routes/iranyelvek/**` - links that point at lamsza.com policy pages

### Lámsza

- Add the same helper for future sister-app links; no mandatory call-site sweep unless hardcoded sister URLs appear.

## Local env (optional)

In each satellite app’s gitignored `.env` / `.env.example`:

```env
# Optional. If unset, *.lamsza.test hostnames auto-select local origins.
# VITE_LAMSZA_ORIGIN=https://lamsza.test
# VITE_SZOTAR_ORIGIN=https://szotar.lamsza.test
# VITE_JATSZOTER_ORIGIN=https://jatszoter.lamsza.test
```

Production: leave unset (or set explicitly to `https://…lamsza.com` if desired).

## Verification

1. Open `https://szotar.lamsza.test` - Lámsza toolbar → `https://lamsza.test`; footer Irányelvek → `https://lamsza.test/iranyelvek`.
2. Open `https://jatszoter.lamsza.test` - same for Lámsza links.
3. Szótár game deep-link (if present) → `https://jatszoter.lamsza.test/jatszok/…`.
4. Production build or open via production hostname - links stay on `https://lamsza.com` (and sister `.com` hosts).
5. With `VITE_LAMSZA_ORIGIN=https://lamsza.test` set and Vite restarted, override wins even on odd hostnames.

## Risks

- Env values baked at Vite build time; changing `.env` requires restart / rebuild.
- Wildcard certs cover one DNS label (`*.lamsza.test`); deeper subdomains are out of scope.
- Copying the helper into three repos can drift; keep the file small and document “copy from lamsza” in the network rule when Lámsza adds it.
