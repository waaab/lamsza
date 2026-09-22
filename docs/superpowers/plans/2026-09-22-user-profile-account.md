# User profile account Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Signed-in `/profil` saves Google identity, theme, homepage slot count, quick links, and listing history on the account, and the header plus homepage follow that account.

**Architecture:** Extend `users` and refresh known Google fields on every sign-in, clearing any known field the token omits. Links, history, and a one-time browser import are separate tables and `/api/account/*` routes. `/profil` shows Profil, Beállítások, Linkjeim, and Előzmények. Bejegyzéseim and Kedvenc helyek are added by the later plans.

**Tech Stack:** Go `net/http` + Postgres, Svelte 5 runes for new components, existing public layout stays on `let` / `$:` , Node `node --test`, Go `go test`.

**Spec:** `docs/superpowers/specs/2026-09-22-user-profile-design.md`

**This is plan 1 of 3.** Plan 2 is favorites. Plan 3 is listing ownership. Do not start plan 2 until this plan’s tasks pass.

## Global Constraints

- Hungarian UI. Route is `/profil`. No public profile.
- Known Google fields: photo URL, given name, family name, display name, email, locale, `sub`. Every sign-in overwrites them. An omitted known field is cleared. Extra claims and the raw token are not stored.
- Theme values are exactly `light`, `dark`, `system`. Slot count stays in **7–14**, default **7**.
- History item shape is `{ slug, name, category, location, photo }`. Store at most **12**. Listing page shows at most **8** and hides the open listing.
- One-time import runs only while `prefs_imported_at` is null, then sets it even if the browser was empty. A failed import leaves `prefs_imported_at` null. Favorites are not imported.
- Signed out, theme, links, slot count, and history stay in this browser (`theme`, `user_quick_links`, `quick_links_display_count`, `lamsza_entry_history`).
- Signed-out visit to `/profil` opens the existing Google sign-in dialog.
- New `.svelte` files use Svelte 5 runes. Run Svelte MCP `svelte-autofixer` on new or edited `.svelte` files until clean.
- Do not commit unless the user explicitly asks. Skip every Commit step unless they have asked.

---

### Task 1: Persist known Google fields

**Files:**
- Modify: `backend/internal/auth/auth.go`
- Modify: `backend/migrations/schema.sql` (users table)
- Modify: `backend/migrations/users_sessions.sql`
- Test: `backend/internal/auth/auth_test.go`

**Interfaces:**
- Consumes: existing `VerifyIDToken`, `users`, `sessions`
- Produces:
  - `GoogleIdentity` fields `Sub`, `Email`, `Name`, `GivenName`, `FamilyName`, `Picture`, `Locale` (all `string`)
  - `User` JSON from `GET /api/auth/me`: `email`, `name`, `is_admin`, `given_name`, `family_name`, `picture`, `locale`, `google_sub`, `last_login_at`, `created_at`, `theme` (`null` or string), `quicklink_slots` (`null` or number), `prefs_imported_at` (`null` or RFC3339)

- [ ] **Step 1: Write the failing test**

Add to `backend/internal/auth/auth_test.go`:

```go
func TestApplyGoogleProfileClearsOmittedFields(t *testing.T) {
	got := ApplyGoogleProfile(UserProfile{
		GivenName: "Anna",
		Picture:   "https://example.test/old.jpg",
		Locale:    "hu",
	}, GoogleIdentity{
		Sub:   "sub-1",
		Email: "anna@example.test",
		Name:  "Anna",
	})
	if got.GivenName != "" || got.Picture != "" || got.Locale != "" {
		t.Fatalf("omitted fields must clear, got %+v", got)
	}
	if got.Email != "anna@example.test" || got.Name != "Anna" || got.Sub != "sub-1" {
		t.Fatalf("present fields must overwrite, got %+v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/auth/ -count=1 -run TestApplyGoogleProfileClearsOmittedFields`

Expected: FAIL with `undefined: ApplyGoogleProfile`

- [ ] **Step 3: Write minimal implementation**

In `auth.go`, extend the structs and add:

```go
type UserProfile struct {
	Sub        string
	Email      string
	Name       string
	GivenName  string
	FamilyName string
	Picture    string
	Locale     string
}

func ApplyGoogleProfile(prev UserProfile, ident GoogleIdentity) UserProfile {
	return UserProfile{
		Sub:        ident.Sub,
		Email:      ident.Email,
		Name:       ident.Name,
		GivenName:  ident.GivenName,
		FamilyName: ident.FamilyName,
		Picture:    ident.Picture,
		Locale:     ident.Locale,
	}
}
```

Extend `GoogleIdentity` with `GivenName`, `FamilyName`, `Picture`, `Locale`. In `VerifyGoogleIDToken`, decode `given_name`, `family_name`, `picture`, `locale` from tokeninfo and copy them onto the identity. Do not persist the raw body.

In `Migrate`, add columns if missing: `given_name VARCHAR(255) NOT NULL DEFAULT ''`, `family_name VARCHAR(255) NOT NULL DEFAULT ''`, `picture TEXT NOT NULL DEFAULT ''`, `locale VARCHAR(35) NOT NULL DEFAULT ''`, `theme VARCHAR(16)`, `quicklink_slots INTEGER`, `prefs_imported_at TIMESTAMP`. Mirror those columns in `schema.sql` and `users_sessions.sql`.

Login upsert writes the seven Google fields on insert and on conflict, including empty strings. `HandleMe` selects them plus `theme`, `quicklink_slots`, `prefs_imported_at`, `last_login_at`, `created_at`, `google_sub`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/auth/ -count=1 -run 'TestApplyGoogleProfileClearsOmittedFields|TestParseTestIDToken|TestIsAdminAllowlist'`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 2: Account preferences and one-time import

**Files:**
- Create: `backend/internal/account/prefs.go`
- Create: `backend/internal/account/prefs_test.go`
- Modify: `backend/main.go` (register routes after `auth.Migrate`)
- Modify: `backend/handlers_test.go` (register the same routes on `testMux`)

**Interfaces:**
- Consumes: `auth.UserFromRequest`, `auth.ApplyGoogleProfile` is not used here
- Produces:
  - `func HandlePreferences(w http.ResponseWriter, r *http.Request)` — `PUT /api/account/preferences` body `{ "theme": "light"|"dark"|"system", "quicklink_slots": 7..14 }`. Omit a key to leave it unchanged. Invalid theme or slot → 400.
  - `func HandleImport(w http.ResponseWriter, r *http.Request)` — `POST /api/account/import` body `{ "theme": "", "quicklink_slots": null, "links": [], "history": [] }`. If `prefs_imported_at` is already set, return the current account and do not write. Otherwise copy only empty account values, set `prefs_imported_at = NOW()`, return `GET` me-shaped prefs.

- [ ] **Step 1: Write the failing test**

`backend/internal/account/prefs_test.go` tests pure helpers (no database):

```go
func TestNormalizeTheme(t *testing.T) {
	if _, err := NormalizeTheme("blue"); err == nil {
		t.Fatal("expected invalid theme")
	}
	got, err := NormalizeTheme("dark")
	if err != nil || got != "dark" {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestClampSlots(t *testing.T) {
	if ClampSlots(3) != 7 || ClampSlots(20) != 14 || ClampSlots(9) != 9 {
		t.Fatal("slot clamp mismatch")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/account/ -count=1 -run 'TestNormalizeTheme|TestClampSlots'`

Expected: FAIL, package or functions missing

- [ ] **Step 3: Write minimal implementation**

```go
func NormalizeTheme(v string) (string, error) {
	switch v {
	case "light", "dark", "system":
		return v, nil
	default:
		return "", fmt.Errorf("invalid theme")
	}
}

func ClampSlots(n int) int {
	if n < 7 {
		return 7
	}
	if n > 14 {
		return 14
	}
	return n
}
```

`HandlePreferences` and `HandleImport` use `auth.UserFromRequest`. Missing session → 401. This task’s import writes only `theme` (when the column is null and the body has a valid theme) and `quicklink_slots` (when the column is null and the body has a number). It accepts `links` and `history` in the JSON and ignores them until Tasks 3 and 4. On success it sets `prefs_imported_at = NOW()` even when both values were empty. On database error it returns 500 and does not set `prefs_imported_at`. Call `account.Migrate()` from `main.go` next to `auth.Migrate()`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/account/ -count=1 -run 'TestNormalizeTheme|TestClampSlots'`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 3: Account links

**Files:**
- Create: `backend/internal/account/links.go`
- Modify: `backend/internal/account/prefs_test.go` (add `TestNormalizeLink`)
- Modify: `backend/main.go`, `backend/handlers_test.go`

**Interfaces:**
- Consumes: `auth.UserFromRequest`
- Produces:
  - table `user_links(id SERIAL PK, user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, title VARCHAR(100) NOT NULL, url VARCHAR(255) NOT NULL, bg_color VARCHAR(50) NOT NULL DEFAULT '#e6f0ff', position INT NOT NULL)`
  - `GET /api/account/links` → `[{id,title,url,bg_color,position}]` ordered by `position`, `id`
  - `POST /api/account/links` body `{title,url,bg_color}` appends at max position + 1
  - `PUT /api/account/links` body `{links:[{id,title,url,bg_color}]}` replaces order for this user only
  - `DELETE /api/account/links?id=` deletes one row owned by the user
  - `func NormalizeLink(title, url, color string) (title, url, color string, err error)` — title and url required after trim; empty color becomes `#e6f0ff`

- [ ] **Step 1: Write the failing test**

```go
func TestNormalizeLink(t *testing.T) {
	title, url, color, err := NormalizeLink("  RMDSZ ", "https://rmdsz.ro", "")
	if err != nil || title != "RMDSZ" || url != "https://rmdsz.ro" || color != "#e6f0ff" {
		t.Fatalf("got %q %q %q %v", title, url, color, err)
	}
	if _, _, _, err := NormalizeLink("", "https://rmdsz.ro", ""); err == nil {
		t.Fatal("expected empty title to fail")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/account/ -count=1 -run TestNormalizeLink`

Expected: FAIL with `undefined: NormalizeLink`

- [ ] **Step 3: Write minimal implementation**

Implement `NormalizeLink` and the four handlers. Every query filters `user_id` from the session. Another user’s id in `PUT` or `DELETE` is ignored (that row is not updated). `PUT` rewrites `position` as the array index starting at 0.

Extend `HandleImport` in the same transaction, before it sets `prefs_imported_at`: when that column is still null and this user has zero `user_links`, insert the payload links in array order. Still ignore `history` until Task 4.

Register routes in `main.go` and `handlers_test.go`:

- `GET/POST /api/account/links` → `HandleLinks`
- `PUT /api/account/links` → `HandleLinks`
- `DELETE /api/account/links` → `HandleLinks`
- `PUT /api/account/preferences` → `HandlePreferences`
- `POST /api/account/import` → `HandleImport`

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/account/ -count=1 -run TestNormalizeLink`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 4: Account listing history

**Files:**
- Create: `backend/internal/account/history.go`
- Modify: `backend/internal/account/prefs_test.go`
- Modify: `backend/main.go`, `backend/handlers_test.go`

**Interfaces:**
- Consumes: `auth.UserFromRequest`
- Produces:
  - table `user_entry_history(user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, slug VARCHAR(255) NOT NULL, name VARCHAR(255) NOT NULL, category VARCHAR(255) NOT NULL DEFAULT '', location VARCHAR(255) NOT NULL DEFAULT '', photo TEXT NOT NULL DEFAULT '', viewed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, slug))`
  - `GET /api/account/history` → newest first, max 12
  - `POST /api/account/history` body `{slug,name,category,location,photo}` upserts and moves `viewed_at` to now, then deletes rows beyond 12 for that user
  - `DELETE /api/account/history?slug=` removes one
  - `DELETE /api/account/history` with no slug clears the user
  - `func NormalizeHistoryItem(slug, name, category, location, photo string) (item HistoryItem, err error)` — slug and name required

- [ ] **Step 1: Write the failing test**

```go
func TestNormalizeHistoryItem(t *testing.T) {
	item, err := NormalizeHistoryItem("kavezo", "Kávézó", "Vendéglő", "Csíkszereda", "")
	if err != nil || item.Slug != "kavezo" || item.Name != "Kávézó" {
		t.Fatalf("got %+v %v", item, err)
	}
	if _, err := NormalizeHistoryItem("", "Kávézó", "", "", ""); err == nil {
		t.Fatal("expected missing slug to fail")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/account/ -count=1 -run TestNormalizeHistoryItem`

Expected: FAIL with `undefined: NormalizeHistoryItem`

- [ ] **Step 3: Write minimal implementation**

Implement the helper and `HandleHistory`. Cap constant `historyStoreMax = 12`. After upsert:

```sql
DELETE FROM user_entry_history
WHERE user_id = $1 AND slug NOT IN (
  SELECT slug FROM user_entry_history WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 12
)
```

Extend `HandleImport` in that same transaction, before `prefs_imported_at` is set: when the user has zero history rows, insert the payload history (slug and name required, cap 12). Register `GET/POST/DELETE /api/account/history`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/account/ -count=1 -run 'TestNormalizeHistoryItem|TestNormalizeLink|TestNormalizeTheme|TestClampSlots'`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 5: Browser account client

**Files:**
- Create: `src/lib/accountPrefs.js`
- Create: `tests/accountPrefs.test.js`
- Modify: `src/lib/stores/auth.js`

**Interfaces:**
- Consumes: `GET /api/auth/me`, `localStorage` keys `theme`, `user_quick_links`, `quick_links_display_count`, `lamsza_entry_history`
- Produces:
  - `export function buildImportPayload({ theme, slots, links, history })` → `{ theme, quicklink_slots, links, history }`
  - `export function meToAuthState(me)` → `{ loggedIn: true, user, email, isAdmin, picture, givenName, familyName, locale, googleSub, lastLoginAt, createdAt, theme, quicklinkSlots, prefsImportedAt }`
  - `auth.refresh()` stores that shape. `auth` empty state includes the new fields as `""` or `null`.

- [ ] **Step 1: Write the failing test**

Create `tests/accountPrefs.test.js`:

```js
import assert from "node:assert/strict";
import test from "node:test";
import { buildImportPayload, meToAuthState } from "../src/lib/accountPrefs.js";

test("buildImportPayload copies browser prefs", () => {
    const payload = buildImportPayload({
        theme: "dark",
        slots: 9,
        links: [{ title: "RMDSZ", url: "https://rmdsz.ro", bg_color: "#fff" }],
        history: [{ slug: "kavezo", name: "Kávézó", category: "", location: "", photo: "" }],
    });
    assert.equal(payload.theme, "dark");
    assert.equal(payload.quicklink_slots, 9);
    assert.equal(payload.links.length, 1);
    assert.equal(payload.history[0].slug, "kavezo");
});

test("meToAuthState maps google fields", () => {
    const state = meToAuthState({
        name: "Anna",
        email: "anna@example.test",
        is_admin: false,
        picture: "https://example.test/a.jpg",
        given_name: "Anna",
        family_name: "Kiss",
        locale: "hu",
        google_sub: "sub-1",
        theme: null,
        quicklink_slots: null,
        prefs_imported_at: null,
    });
    assert.equal(state.loggedIn, true);
    assert.equal(state.picture, "https://example.test/a.jpg");
    assert.equal(state.theme, null);
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test tests/accountPrefs.test.js`

Expected: FAIL, cannot find module

- [ ] **Step 3: Write minimal implementation**

Implement the two functions in `src/lib/accountPrefs.js`. `buildImportPayload` drops links missing title or url and history rows missing slug or name. `meToAuthState` uses `me.name || me.email` for `user`.

Update `src/lib/stores/auth.js` so `refresh` sets `meToAuthState(me)` and the empty object has the new keys. After a successful `refresh`, if `prefsImportedAt` is null and `localStorage` exists, `POST /api/account/import` with `buildImportPayload` from the current browser keys, then `refresh` once more. If that POST fails, keep the session and leave `prefsImportedAt` null.

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test tests/accountPrefs.test.js`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 6: Profile page

**Files:**
- Create: `src/routes/(public)/profil/+page.svelte`
- Modify: `src/routes/(public)/+layout.svelte` (name links to `/profil`)
- Modify: `src/lib/stores/theme.js` (`applyTheme` writes the account when signed in)

**Interfaces:**
- Consumes: `auth`, `PUT /api/account/preferences`, links and history routes from Tasks 3–4, `applyTheme`, `LABELS`
- Produces: `/profil` with tabs **Profil**, **Beállítások**, **Linkjeim**, **Előzmények**. Header `.nav-admin-user` is an `<a href="/profil">`.

- [ ] **Step 1: Write the failing test**

Add to `tests/accountPrefs.test.js`:

```js
test("profileTabIds are the account tabs", () => {
    assert.deepEqual(profileTabIds, ["profil", "beallitasok", "linkjeim", "elozmenyek"]);
});
```

Import `profileTabIds` from `src/lib/accountPrefs.js`.

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test tests/accountPrefs.test.js`

Expected: FAIL, `profileTabIds` missing

- [ ] **Step 3: Write minimal implementation**

Export `profileTabIds` from `accountPrefs.js`.

`profil/+page.svelte` (Svelte 5 runes): on mount read `$auth`. If `loggedIn` is false, call the layout’s existing login dialog by dispatching the same event the header login button uses, or navigate nowhere and show a **Belépés** button that calls `openLoginDialog` duplicated via a shared function. Prefer exporting `openLogin` from a tiny `src/lib/openLogin.js` only if the layout already cannot be called. Simpler rule: the page shows **Jelentkezz be a profilodhoz** and a button that clicks the same path as the header login button. Read `+layout.svelte` `openLoginDialog` and move that function into `src/lib/openLogin.js` if both the layout and the page need it. Keep one implementation.

**Profil** tab renders only non-empty: picture (`<img alt="">`), given name, family name, display name, email, locale, Google account id, last login, created. Labels in Hungarian: **Keresztnév**, **Vezetéknév**, **Név**, **E-mail**, **Nyelv**, **Google azonosító**, **Utolsó belépés**, **Fiók létrehozva**.

**Beállítások** has three theme buttons using `LABELS` and a slot control that calls the existing `writeSlotCount` clamp (7–14). Save calls `PUT /api/account/preferences` then `applyTheme`. On failure, leave the previous selection and show **A mentés nem sikerült**.

**Linkjeim** lists account links with add, edit, delete, and up/down reorder. Reorder sends `PUT /api/account/links` with the full list. Save failures keep the previous list and show the same error.

**Előzmények** lists account history newest first with per-row remove and **Összes törlése**.

`applyTheme` in `theme.js`: when `auth` is logged in, `PUT` the theme after writing `localStorage`, so a signed-in header change and the profile buttons hit the same account column. When logged out, only `localStorage`.

Header name: replace the `<span class="nav-admin-user">` with `<a class="nav-admin-user" href="/profil">`.

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test tests/accountPrefs.test.js tests/quickLinksDisplay.test.js`

Expected: PASS

Run Svelte autofixer on `src/routes/(public)/profil/+page.svelte` and any new `.svelte` file until clean.

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 7: Homepage and listing history follow the account

**Files:**
- Modify: `src/routes/(public)/+page.svelte`
- Modify: `src/routes/(public)/bejegyzes/[slug]/+page.svelte`
- Modify: `src/lib/entryHistory.js`

**Interfaces:**
- Consumes: `GET /api/account/links`, `GET /api/auth/me` `quicklink_slots`, `POST /api/account/history`, existing `recordHistoryVisit`
- Produces:
  - `export async function recordAccountHistory(item)` in `entryHistory.js` — if `auth` is logged in, `POST /api/account/history` and also call `recordHistoryVisit`; if logged out, only `recordHistoryVisit`
  - Homepage, when `$auth.loggedIn`, loads links from the account and slot count from `quicklinkSlots` (fallback `readSlotCount()` when null)

- [ ] **Step 1: Write the failing test**

Add to `tests/entryHistory.test.js` a test that `recordAccountHistory` is a function. Behavior with `fetch` mocked:

```js
test("recordAccountHistory posts when logged in", async () => {
    const calls = [];
    globalThis.fetch = async (url, opts) => {
        calls.push({ url: String(url), body: opts?.body });
        return { ok: true, json: async () => [] };
    };
    await recordAccountHistory(
        { slug: "kavezo", name: "Kávézó", category: "", location: "", photo: "" },
        { loggedIn: true },
    );
    assert.equal(calls.length, 1);
    assert.match(calls[0].url, /\/api\/account\/history$/);
});
```

Pass auth state as the second argument so the test does not import the store. Signature: `recordAccountHistory(item, authState)`.

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test tests/entryHistory.test.js`

Expected: FAIL, `recordAccountHistory` missing

- [ ] **Step 3: Write minimal implementation**

Implement `recordAccountHistory`. Logged-in POST failure still calls `recordHistoryVisit` so the browser copy remains. The listing page calls `recordAccountHistory(item, get(auth))` instead of `recordHistoryVisit` alone.

Homepage `onMount`: if logged in, `GET /api/account/links` replaces `userLinks`, and `quicklinkSlots` from auth replaces `slotCount` when it is a number. Slot buttons `PUT /api/account/preferences`. Link dialog saves go to the account routes. Signed out, keep the current `localStorage` path.

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test tests/entryHistory.test.js tests/accountPrefs.test.js`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.
