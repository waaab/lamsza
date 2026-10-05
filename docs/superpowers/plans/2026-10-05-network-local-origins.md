# Network local origins Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make cross-app and policy links use `https://*.lamsza.test` locally and `https://*.lamsza.com` in production, via env override or hostname sniff.

**Architecture:** A small pure resolver plus thin wrappers in `networkOrigins.js`, copied into Lámsza, Szótár, and Játszótér. Call sites stop hardcoding `lamsza.com` / `jatszoter.lamsza.com`. Optional `VITE_*_ORIGIN` env vars override; otherwise `*.lamsza.test` selects local hosts; else production `.com`. Always `https://`.

**Tech Stack:** SvelteKit / Vite (`import.meta.env`), Node built-in test runner (`node --test`). Spec: `docs/superpowers/specs/2026-10-05-network-local-origins-design.md`.

## Global Constraints

- Always `https://` for network origins (never match page `http:`).
- Resolution order: env `VITE_<APP>_ORIGIN` → hostname sniff (`lamsza.test` or `*.lamsza.test`) → production `.com`.
- App keys: `lamsza`, `szotar`, `jatszoter`.
- Env names: `VITE_LAMSZA_ORIGIN`, `VITE_SZOTAR_ORIGIN`, `VITE_JATSZOTER_ORIGIN`.
- No em dash in copy or comments. Use ` - ` when a dash is required.
- Do not change Google OAuth origins or FAQ marketing copy that only mentions “lamsza.com”.
- Keep existing `localhost` / `127.0.0.1` shortcut for social-config fetch to `http://127.0.0.1:3000/api/config/public`; for every other host use `lamszaUrl('/api/config/public')`.
- Copy the same helper into each app; Lámsza is the source of truth.
- Commit only when the user asks (or when executing a plan that explicitly requires it and the user chose that path).

---

## File structure

- Create `src/lib/networkOrigins.js` (Lámsza) - pure `resolveNetworkOrigin` + `lamszaOrigin` / `szotarOrigin` / `jatszoterOrigin` + `*Url` helpers.
- Create `tests/networkOrigins.test.js` (Lámsza) - unit tests for the resolver.
- Create `frontend/src/lib/networkOrigins.js` (Szótár, Játszótér) - copy of the Lámsza helper.
- Create `frontend/tests/networkOrigins.test.js` (Szótár, Játszótér) - same tests (or import from a copied test file).
- Modify Szótár / Játszótér layouts, SignInDialog, games.js, iranyelvek pages, `.env.example`.
- Optionally document the three `VITE_*_ORIGIN` keys in `docs/PRODUCTION_SERVER_SETUP.md` section 6 (one short note).

---

### Task 1: Helper + tests in Lámsza

**Files:**
- Create: `src/lib/networkOrigins.js`
- Create: `tests/networkOrigins.test.js`

**Interfaces:**
- Produces:
  - `resolveNetworkOrigin(app, { envOrigin, hostname })` → `string`
  - `lamszaOrigin()` / `szotarOrigin()` / `jatszoterOrigin()` → `string`
  - `lamszaUrl(path)` / `szotarUrl(path)` / `jatszoterUrl(path)` → `string`

- [ ] **Step 1: Write the failing test**

Create `tests/networkOrigins.test.js`:

```js
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { resolveNetworkOrigin, joinOriginPath } from '../src/lib/networkOrigins.js';

test('env override wins', () => {
	assert.equal(
		resolveNetworkOrigin('lamsza', {
			envOrigin: 'https://lamsza.test/',
			hostname: 'szotar.lamsza.com'
		}),
		'https://lamsza.test'
	);
});

test('hostname sniff selects .test hosts', () => {
	assert.equal(
		resolveNetworkOrigin('lamsza', { envOrigin: '', hostname: 'szotar.lamsza.test' }),
		'https://lamsza.test'
	);
	assert.equal(
		resolveNetworkOrigin('szotar', { envOrigin: '', hostname: 'lamsza.test' }),
		'https://szotar.lamsza.test'
	);
	assert.equal(
		resolveNetworkOrigin('jatszoter', { envOrigin: '', hostname: 'jatszoter.lamsza.test' }),
		'https://jatszoter.lamsza.test'
	);
});

test('production default when hostname is not .test', () => {
	assert.equal(
		resolveNetworkOrigin('lamsza', { envOrigin: '', hostname: 'szotar.lamsza.com' }),
		'https://lamsza.com'
	);
	assert.equal(
		resolveNetworkOrigin('szotar', { envOrigin: '', hostname: '' }),
		'https://szotar.lamsza.com'
	);
	assert.equal(
		resolveNetworkOrigin('jatszoter', { envOrigin: '', hostname: 'localhost' }),
		'https://jatszoter.lamsza.com'
	);
});

test('joinOriginPath joins path', () => {
	assert.equal(joinOriginPath('https://lamsza.test', '/iranyelvek'), 'https://lamsza.test/iranyelvek');
	assert.equal(joinOriginPath('https://lamsza.test', ''), 'https://lamsza.test');
	assert.equal(joinOriginPath('https://lamsza.test/', 'iranyelvek'), 'https://lamsza.test/iranyelvek');
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test tests/networkOrigins.test.js`  
Expected: FAIL (module not found or exports missing)

- [ ] **Step 3: Write minimal implementation**

Create `src/lib/networkOrigins.js`:

```js
const APPS = {
	lamsza: {
		envKey: 'VITE_LAMSZA_ORIGIN',
		local: 'https://lamsza.test',
		prod: 'https://lamsza.com'
	},
	szotar: {
		envKey: 'VITE_SZOTAR_ORIGIN',
		local: 'https://szotar.lamsza.test',
		prod: 'https://szotar.lamsza.com'
	},
	jatszoter: {
		envKey: 'VITE_JATSZOTER_ORIGIN',
		local: 'https://jatszoter.lamsza.test',
		prod: 'https://jatszoter.lamsza.com'
	}
};

/** @param {string} hostname */
function isLocalNetworkHost(hostname) {
	const host = String(hostname || '');
	return host === 'lamsza.test' || host.endsWith('.lamsza.test');
}

/**
 * @param {'lamsza'|'szotar'|'jatszoter'} app
 * @param {{ envOrigin?: string, hostname?: string }} [opts]
 */
export function resolveNetworkOrigin(app, opts = {}) {
	const cfg = APPS[app];
	if (!cfg) throw new Error(`Unknown network app: ${app}`);
	const env = String(opts.envOrigin ?? '').trim().replace(/\/$/, '');
	if (env) return env;
	if (isLocalNetworkHost(opts.hostname)) return cfg.local;
	return cfg.prod;
}

/** @param {string} origin @param {string} [path] */
export function joinOriginPath(origin, path = '') {
	const base = String(origin || '').replace(/\/$/, '');
	const p = String(path || '').trim();
	if (!p) return base;
	return `${base}${p.startsWith('/') ? p : `/${p}`}`;
}

function readEnv(key) {
	try {
		return import.meta.env?.[key];
	} catch {
		return undefined;
	}
}

function currentHostname() {
	if (typeof window !== 'undefined' && window.location?.hostname) {
		return window.location.hostname;
	}
	return '';
}

/** @param {'lamsza'|'szotar'|'jatszoter'} app */
function originFor(app) {
	const cfg = APPS[app];
	return resolveNetworkOrigin(app, {
		envOrigin: readEnv(cfg.envKey),
		hostname: currentHostname()
	});
}

export function lamszaOrigin() {
	return originFor('lamsza');
}
export function szotarOrigin() {
	return originFor('szotar');
}
export function jatszoterOrigin() {
	return originFor('jatszoter');
}

/** @param {string} [path] */
export function lamszaUrl(path = '') {
	return joinOriginPath(lamszaOrigin(), path);
}
/** @param {string} [path] */
export function szotarUrl(path = '') {
	return joinOriginPath(szotarOrigin(), path);
}
/** @param {string} [path] */
export function jatszoterUrl(path = '') {
	return joinOriginPath(jatszoterOrigin(), path);
}
```

- [ ] **Step 4: Run tests and make sure they pass**

Run: `node --test tests/networkOrigins.test.js`  
Expected: PASS (all four tests)

- [ ] **Step 5: Commit (only if the user asked for commits)**

```bash
git add src/lib/networkOrigins.js tests/networkOrigins.test.js
git commit -m "$(cat <<'EOF'
Add networkOrigins helper for local .test vs production .com links.

EOF
)"
```

---

### Task 2: Wire Szótár call sites

**Files:**
- Create: `/home/attila/projects/szotar/frontend/src/lib/networkOrigins.js` (copy from Lámsza Task 1)
- Create: `/home/attila/projects/szotar/frontend/tests/networkOrigins.test.js` (copy from Lámsza; fix import path to `../src/lib/networkOrigins.js`)
- Modify: `/home/attila/projects/szotar/frontend/src/routes/+layout.svelte`
- Modify: `/home/attila/projects/szotar/frontend/src/lib/components/SignInDialog.svelte`
- Modify: `/home/attila/projects/szotar/frontend/src/lib/games.js`
- Create or modify: `/home/attila/projects/szotar/.env.example` (add commented `VITE_*_ORIGIN` lines)

**Interfaces:**
- Consumes: `lamszaUrl`, `jatszoterUrl` from `$lib/networkOrigins.js`

- [ ] **Step 1: Copy helper + tests into Szótár**

Copy Lámsza `src/lib/networkOrigins.js` → `szotar/frontend/src/lib/networkOrigins.js`.  
Copy tests → `szotar/frontend/tests/networkOrigins.test.js` with import `../src/lib/networkOrigins.js`.

Run: `cd /home/attila/projects/szotar/frontend && node --test tests/networkOrigins.test.js`  
Expected: PASS

- [ ] **Step 2: Update `+layout.svelte`**

In `<script>`, add:

```js
import { lamszaUrl } from '$lib/networkOrigins.js';
```

Replace `networkConfigUrl`:

```js
function networkConfigUrl() {
	const host = location.hostname;
	if (host === 'localhost' || host === '127.0.0.1') {
		return 'http://127.0.0.1:3000/api/config/public';
	}
	return lamszaUrl('/api/config/public');
}
```

Replace toolbar href:

```svelte
<a href={lamszaUrl()} class="btn nav-btn" title="Lámsza" rel="noopener">
```

Replace footer policy links:

```svelte
<a href={lamszaUrl('/iranyelvek')} title="Irányelvek">Irányelvek</a>
<a href={lamszaUrl('/iranyelvek/feltetelek')} title="Feltételek">Feltételek</a>
<a href={lamszaUrl('/iranyelvek/sutik')} title="Sütik">Sütik</a>
```

- [ ] **Step 3: Update SignInDialog default**

```js
import { lamszaUrl } from '$lib/networkOrigins.js';

let {
	open = false,
	appName = 'Székely szótár',
	clientId = '',
	configReady = true,
	onClose = () => {},
	onSignedIn = () => {},
	policyHref = lamszaUrl('/iranyelvek')
} = $props();
```

- [ ] **Step 4: Update `games.js`**

```js
import { jatszoterUrl } from '$lib/networkOrigins.js';

/** @param {string} slug */
export function gameHref(slug) {
	return jatszoterUrl(`/jatszok/${slug}`);
}
```

- [ ] **Step 5: Document optional env in `.env.example`**

Create `/home/attila/projects/szotar/.env.example` if missing, with at least:

```env
# Optional network origins (default: sniff *.lamsza.test, else *.lamsza.com)
# VITE_LAMSZA_ORIGIN=https://lamsza.test
# VITE_SZOTAR_ORIGIN=https://szotar.lamsza.test
# VITE_JATSZOTER_ORIGIN=https://jatszoter.lamsza.test
```

If the file already exists with other keys, only append these commented lines.

- [ ] **Step 6: Manual check**

With apps running, open `https://szotar.lamsza.test` (or `http://` if TLS not ready).  
Hover Lámsza toolbar and Irányelvek: targets must be `https://lamsza.test` and `https://lamsza.test/iranyelvek`.  
Restart Vite after any `.env` change.

- [ ] **Step 7: Commit in szotar repo (only if the user asked)**

```bash
cd /home/attila/projects/szotar
git add frontend/src/lib/networkOrigins.js frontend/tests/networkOrigins.test.js \
  frontend/src/routes/+layout.svelte frontend/src/lib/components/SignInDialog.svelte \
  frontend/src/lib/games.js .env.example
git commit -m "$(cat <<'EOF'
Resolve Lámsza and Játszótér links for local .test vs production.

EOF
)"
```

---

### Task 3: Wire Játszótér call sites

**Files:**
- Create: `/home/attila/projects/jatszoter/frontend/src/lib/networkOrigins.js`
- Create: `/home/attila/projects/jatszoter/frontend/tests/networkOrigins.test.js`
- Modify: `/home/attila/projects/jatszoter/frontend/src/routes/+layout.svelte`
- Modify: `/home/attila/projects/jatszoter/frontend/src/lib/components/SignInDialog.svelte`
- Modify: `/home/attila/projects/jatszoter/frontend/src/routes/iranyelvek/+page.svelte`
- Modify: `/home/attila/projects/jatszoter/frontend/src/routes/iranyelvek/feltetelek/+page.svelte`
- Modify: `/home/attila/projects/jatszoter/frontend/src/routes/iranyelvek/sutik/+page.svelte`
- Modify: `/home/attila/projects/jatszoter/.env.example`

**Interfaces:**
- Consumes: `lamszaUrl` from `$lib/networkOrigins.js`

- [ ] **Step 1: Copy helper + tests**

Same as Task 2 Step 1 into Játszótér paths.

Run: `cd /home/attila/projects/jatszoter/frontend && node --test tests/networkOrigins.test.js`  
Expected: PASS

- [ ] **Step 2: Update `+layout.svelte`**

Import `lamszaUrl`. Replace `networkConfigUrl` the same way as Szótár (localhost → `:3000`, else `lamszaUrl('/api/config/public')`).

Toolbar:

```svelte
<a href={lamszaUrl()} class="btn nav-btn" title="Lámsza" rel="noopener" target="_blank">
```

Footer:

```svelte
<a href={lamszaUrl('/iranyelvek')} title="Irányelvek">Irányelvek</a>
<a href={lamszaUrl('/iranyelvek/feltetelek')} title="Feltételek">Feltételek</a>
<a href={lamszaUrl('/iranyelvek/sutik')} title="Sütik">Sütik</a>
```

- [ ] **Step 3: Update SignInDialog**

```js
import { lamszaUrl } from '$lib/networkOrigins.js';
// ...
policyHref = lamszaUrl('/iranyelvek')
```

- [ ] **Step 4: Update iranyelvek pages**

`iranyelvek/+page.svelte`:

```svelte
<script>
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import { lamszaUrl } from '$lib/networkOrigins.js';
</script>
<!-- ... -->
<a href={lamszaUrl('/iranyelvek')}>lamsza.com/iranyelvek</a>.
```

(Keep visible label `lamsza.com/...` as marketing text; only the `href` is dynamic.)

Same pattern for feltetelek (`/iranyelvek/feltetelek`) and sutik (`/iranyelvek/sutik`).

- [ ] **Step 5: Append commented `VITE_*_ORIGIN` lines to `.env.example`**

Same three commented lines as Szótár.

- [ ] **Step 6: Manual check**

Open `https://jatszoter.lamsza.test` - Lámsza and Irányelvek point at `https://lamsza.test…`.

- [ ] **Step 7: Commit in jatszoter repo (only if the user asked)**

```bash
cd /home/attila/projects/jatszoter
git add frontend/src/lib/networkOrigins.js frontend/tests/networkOrigins.test.js \
  frontend/src/routes/+layout.svelte frontend/src/lib/components/SignInDialog.svelte \
  frontend/src/routes/iranyelvek/+page.svelte \
  frontend/src/routes/iranyelvek/feltetelek/+page.svelte \
  frontend/src/routes/iranyelvek/sutik/+page.svelte \
  .env.example
git commit -m "$(cat <<'EOF'
Resolve Lámsza policy and toolbar links for local .test vs production.

EOF
)"
```

---

### Task 4: Production notes + grep hygiene

**Files:**
- Modify: `docs/PRODUCTION_SERVER_SETUP.md` (section 6, short note under each app or one shared bullet)
- Optional: mention in `docs/superpowers/specs/2026-10-05-network-local-origins-design.md` that the plan is `docs/superpowers/plans/2026-10-05-network-local-origins.md`

- [ ] **Step 1: Add a short note to PRODUCTION_SERVER_SETUP.md**

Under section 6 (env vars), add:

```markdown
**Network origins (frontend build):** Optional `VITE_LAMSZA_ORIGIN`, `VITE_SZOTAR_ORIGIN`, `VITE_JATSZOTER_ORIGIN`. Leave unset in production so links default to `https://*.lamsza.com`. Local `.test` hostnames auto-select `https://*.lamsza.test` without these vars.
```

- [ ] **Step 2: Grep for leftover hardcodes in satellite frontends**

Run:

```bash
rg -n 'https://lamsza\.com|https://jatszoter\.lamsza\.com|https://szotar\.lamsza\.com' \
  /home/attila/projects/szotar/frontend/src \
  /home/attila/projects/jatszoter/frontend/src
```

Expected: no matches in `href` / URL builders (visible text `lamsza.com/...` in iranyelvek pages is OK).

- [ ] **Step 3: Commit lamsza docs + helper (only if the user asked)**

```bash
cd /home/attila/projects/lamsza
git add src/lib/networkOrigins.js tests/networkOrigins.test.js \
  docs/PRODUCTION_SERVER_SETUP.md \
  docs/superpowers/specs/2026-10-05-network-local-origins-design.md \
  docs/superpowers/plans/2026-10-05-network-local-origins.md
git commit -m "$(cat <<'EOF'
Add network origin helper and document local .test vs .com links.

EOF
)"
```

---

## Plan self-review

1. **Spec coverage:** Env + sniff + prod defaults, always https, full network (Lámsza + Játszótér links), call sites listed, optional `.env.example`, SSR-safe empty hostname → prod. Covered.
2. **Placeholders:** None; full helper and call-site snippets included.
3. **Type consistency:** `resolveNetworkOrigin` / `joinOriginPath` / `lamszaUrl` names match across tasks. Localhost social-config exception documented in Global Constraints and Tasks 2–3.
