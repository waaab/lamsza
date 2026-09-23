# Directory websites Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A signed-in user can submit a domain, title, and short description for free; an admin approves, rejects, or bans from one notification; an approved website is a direct-link weboldal until claim creates a normal listing.

**Architecture:** One `websites` row is both the submission and the admin notification. Its `domain_key` is unique against other websites and against listing URLs (backfilled into the same table). Approve flips `status` to `approved`. Claim reuses `POST /api/account/listings` with `website_id`, which creates an unpublished listing, sets the creator as owner, and stores `websites.entry_id`. Search adds a `websites` hit and a `website_query` flag. No new entry type, category, tag, or unclaimed `/bejegyzes/` page.

**Tech Stack:** Go `net/http` + Postgres, Svelte 5 runes for the new form and website card, existing index/search/admin pages stay on `let` / `$:`, Node `node --test`, Go `go test`. No new Go or npm dependencies.

**Spec:** `docs/superpowers/specs/2026-09-23-directory-websites-design.md`

## Global Constraints

- Public action label is exactly `Add your website for free`.
- The form asks only for domain, title, and short description. Title max 120 runes. Short description max 300 runes. HTML is stripped. All three are required.
- Admin notification buttons are exactly `Approve`, `Reject`, and `Ban User`. No email.
- **Claimed** is the existing owner mark (`entries` JSON `claimed`, true when an active owner exists). Reuse the label already rendered by `.entry-profile__owned` in `src/lib/components/EntryProfile.svelte`. Do not add a second badge or a new word.
- Ban blocks only website submit and website claim. Sign-in still works. Existing listings are not deleted.
- An approved unclaimed website has no `/bejegyzes/` page. Its link is `https://` plus the domain key.
- Creating the listing does not rewrite the website title or short description. The listing stays unpublished until the existing publish action. Publishing does not set `verified`.
- A second listing cannot be created for a domain that already has one. After publish, another user joins through the existing claim endpoint.
- New `.svelte` files use Svelte 5 runes. Do not migrate `src/routes/(public)/index/+page.svelte`, `src/lib/components/SearchEngine.svelte`, or `src/routes/admin/+page.svelte` to runes. Run Svelte MCP `svelte-autofixer` on new or edited `.svelte` files until clean (ignore `$:` complaints on those three legacy files).
- Commit only the files listed in that task. Never `git add -A`. `backend/main.go` and `backend/handlers_test.go` already contain unrelated uncommitted work: add the route and migrate lines, but do not stage those files if the commit would include unrelated hunks. Put new Go tests in `backend/website_test.go` (`package main`). If a commit would stage unrelated hunks, leave those files unstaged and say so.
- Do not merge to `main` in these tasks.

## File structure

- `backend/internal/webdomain/domain.go` — canonical key and plain-text limits. No database.
- `src/lib/websiteDomain.js` — the same key rules for the form, tested in Node.
- `backend/internal/account/websites.go` — submit, public list, admin action, search hits, queue rows.
- `backend/migrations/websites.sql` plus `account.MigrateWebsites()` — table, ban flag, listing-URL backfill.
- `src/lib/components/AddWebsiteForm.svelte` — the three-field form.
- `src/lib/components/WebsiteCard.svelte` — title, description, domain, direct link.
- Index, search, and admin pages only gain a section or message. `EntryCard.svelte` gains the existing owner mark.

---

### Task 1: Canonical domain and plain text

**Files:**
- Create: `backend/internal/webdomain/domain.go`
- Create: `backend/internal/webdomain/domain_test.go`
- Create: `src/lib/websiteDomain.js`
- Test: `tests/websiteDomain.test.js`

**Interfaces:**
- Consumes: nothing
- Produces:
  - `func CanonicalDomain(raw string) (string, error)` — registrable domain, or an error when the input has no host
  - `func Plain(raw string, maxRunes int) (string, error)` — HTML stripped, trimmed; error if empty or longer than `maxRunes`
  - `canonicalDomain(raw)` and `plainText(raw, max)` with the same results in `src/lib/websiteDomain.js`

- [ ] **Step 1: Write the failing Go test**

Create `backend/internal/webdomain/domain_test.go`:

```go
package webdomain

import "testing"

func TestCanonicalDomain(t *testing.T) {
	got, err := CanonicalDomain("https://www.kezdisorozo.com/menu")
	if err != nil || got != "kezdisorozo.com" {
		t.Fatalf("got %q err %v", got, err)
	}
	for _, raw := range []string{"kezdisorozo.com", "HTTP://KEZDISOROZO.com", "shop.kezdisorozo.com"} {
		got, err = CanonicalDomain(raw)
		if err != nil || got != "kezdisorozo.com" {
			t.Fatalf("%s -> %q err %v", raw, got, err)
		}
	}
	if _, err := CanonicalDomain("not a domain"); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlain(t *testing.T) {
	got, err := Plain("  <b>Csíki</b>  ", 120)
	if err != nil || got != "Csíki" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := Plain("   ", 120); err == nil {
		t.Fatal("expected empty error")
	}
	if _, err := Plain("abcd", 3); err == nil {
		t.Fatal("expected length error")
	}
}
```

- [ ] **Step 2: Run the Go test to verify it fails**

Run: `cd backend && go test -count=1 ./internal/webdomain/`

Expected: FAIL, `CanonicalDomain` undefined.

- [ ] **Step 3: Write the failing Node test**

Create `tests/websiteDomain.test.js`:

```js
import assert from "node:assert/strict";
import test from "node:test";
import { canonicalDomain, plainText } from "../src/lib/websiteDomain.js";

test("www, path, and subdomain share one key", () => {
    for (const raw of ["https://www.kezdisorozo.com/menu", "kezdisorozo.com", "shop.kezdisorozo.com"]) {
        assert.equal(canonicalDomain(raw), "kezdisorozo.com");
    }
    assert.equal(canonicalDomain("not a domain"), "");
});

test("plain text strips HTML and enforces length", () => {
    assert.equal(plainText("  <b>Csíki</b>  ", 120), "Csíki");
    assert.equal(plainText("   ", 120), "");
    assert.equal(plainText("abcd", 3), "");
});
```

- [ ] **Step 4: Run the Node test to verify it fails**

Run: `node --test tests/websiteDomain.test.js`

Expected: FAIL, cannot find `websiteDomain.js`.

- [ ] **Step 5: Implement the Go package**

Create `backend/internal/webdomain/domain.go`:

```go
package webdomain

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var tagRE = regexp.MustCompile(`<[^>]*>`)

var multiSuffix = map[string]bool{
	"co.uk": true,
	"org.uk": true,
	"com.au": true,
}

func CanonicalDomain(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("empty domain")
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return "", errors.New("invalid domain")
	}
	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")
	labels := strings.Split(host, ".")
	if len(labels) < 2 || labels[0] == "" || strings.Contains(host, " ") {
		return "", errors.New("invalid domain")
	}
	suffix := labels[len(labels)-2] + "." + labels[len(labels)-1]
	if multiSuffix[suffix] {
		if len(labels) < 3 {
			return "", errors.New("invalid domain")
		}
		return strings.Join(labels[len(labels)-3:], "."), nil
	}
	return strings.Join(labels[len(labels)-2:], "."), nil
}

func Plain(raw string, maxRunes int) (string, error) {
	s := strings.TrimSpace(tagRE.ReplaceAllString(raw, ""))
	if s == "" {
		return "", errors.New("empty")
	}
	if utf8.RuneCountInString(s) > maxRunes {
		return "", errors.New("too long")
	}
	return s, nil
}
```

- [ ] **Step 6: Implement the JS module with the same rules**

Create `src/lib/websiteDomain.js` exporting `canonicalDomain` and `plainText`. `canonicalDomain` returns `""` on failure. `plainText` returns `""` when empty or over the rune limit. Mirror the Go steps: add `https://` when there is no scheme, take `hostname`, drop a leading `www.`, collapse to the last two labels unless the last two are `co.uk`, `org.uk`, or `com.au` (then the last three). Strip tags with `/<[^>]*>/g` before trimming. Count characters with `[...s].length`.

- [ ] **Step 7: Run both tests**

Run: `cd backend && go test -count=1 ./internal/webdomain/ && cd .. && node --test tests/websiteDomain.test.js`

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/webdomain/domain.go backend/internal/webdomain/domain_test.go src/lib/websiteDomain.js tests/websiteDomain.test.js
git commit -m "$(cat <<'EOF'
Add a shared canonical domain key for directory websites.

EOF
)"
```

---

### Task 2: Website table, ban flag, and listing-URL backfill

**Files:**
- Create: `backend/migrations/websites.sql`
- Modify: `backend/internal/account/prefs.go` (call the new migrate from `account.Migrate`, or add `MigrateWebsites` beside it)
- Modify: `backend/main.go` (call `account.MigrateWebsites()` next to `account.Migrate()`)
- Modify: `backend/handlers_test.go` `init` (same call after `account.Migrate()`)
- Test: `backend/website_test.go`

**Interfaces:**
- Consumes: `webdomain.CanonicalDomain`
- Produces:
  - `websites(id, domain_key UNIQUE, submitted_host, title, description, status, user_id NULL, entry_id NULL, created_at)`
  - `status` is `pending` or `approved`
  - `users.website_banned BOOLEAN NOT NULL DEFAULT false`
  - `func MigrateWebsites()`
  - One approved row per distinct canonical listing URL already stored on `entries.url`, with `entry_id` set, `user_id` null, title = the listing name trimmed to 120 runes, description = listing notes trimmed to 300 runes or `""` when notes are empty

- [ ] **Step 1: Write the failing migration test**

Create `backend/website_test.go`:

```go
package main

import (
	"backend/internal/account"
	"backend/internal/db"
	"testing"
)

func TestMigrateWebsitesBackfillsListingURL(t *testing.T) {
	account.MigrateWebsites()
	var banned bool
	if err := db.DB.QueryRow(`SELECT website_banned FROM users LIMIT 1`).Scan(&banned); err != nil {
		t.Fatal(err)
	}
	id, _ := createEntry(t, "Backfill Sorozo", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)
	if _, err := db.DB.Exec(`UPDATE entries SET url = $1 WHERE id = $2`, "https://www.backfill-sorozo.com/etlap", id); err != nil {
		t.Fatal(err)
	}
	account.MigrateWebsites()
	var key, status string
	var entryID int
	err := db.DB.QueryRow(`SELECT domain_key, status, entry_id FROM websites WHERE domain_key = $1`, "backfill-sorozo.com").Scan(&key, &status, &entryID)
	if err != nil || key != "backfill-sorozo.com" || status != "approved" {
		t.Fatalf("backfill %q %q err %v", key, status, err)
	}
}
```

Use the existing `createEntry`, `mustLocID`, `formatID`, and `doRequest` helpers from `backend/handlers_test.go` and `backend/entry_claimed_extras_test.go`. If `createEntry` does not set `url`, the `UPDATE` above is the fixture.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test -count=1 -run TestMigrateWebsitesBackfillsListingURL .`

Expected: FAIL, `MigrateWebsites` undefined, or the column is missing.

- [ ] **Step 3: Add the migration**

`backend/migrations/websites.sql`:

```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS website_banned BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS websites (
    id SERIAL PRIMARY KEY,
    domain_key VARCHAR(253) NOT NULL UNIQUE,
    submitted_host VARCHAR(253) NOT NULL,
    title VARCHAR(120) NOT NULL DEFAULT '',
    description VARCHAR(300) NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL,
    user_id INTEGER REFERENCES users(id),
    entry_id INTEGER REFERENCES entries(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT websites_status_check CHECK (status IN ('pending', 'approved'))
);
CREATE INDEX IF NOT EXISTS idx_websites_status ON websites (status);
CREATE INDEX IF NOT EXISTS idx_websites_entry ON websites (entry_id);
```

`MigrateWebsites` executes that SQL, then selects `id, name, notes, url` from `entries` where `url` is not empty. For each row, `CanonicalDomain` the URL. Skip failures. `INSERT INTO websites (domain_key, submitted_host, title, description, status, entry_id) VALUES ($1,$1,$2,$3,'approved',$4) ON CONFLICT (domain_key) DO NOTHING`. Title is the name passed through `Plain` when it fits, otherwise the first 120 runes. Description uses `Plain` when notes are non-empty and fit; otherwise `""`. Do not overwrite an existing key.

Call `MigrateWebsites()` from `main()` immediately after `account.Migrate()`, and from `handlers_test.go` `init` immediately after `account.Migrate()`.

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd backend && go test -count=1 -run TestMigrateWebsitesBackfillsListingURL .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/websites.sql backend/internal/account/prefs.go backend/website_test.go
git commit -m "$(cat <<'EOF'
Store directory websites and occupy keys already used by listings.

EOF
)"
```

If `MigrateWebsites` lives in a new file `backend/internal/account/websites_migrate.go`, add that path instead of `prefs.go`. Leave `main.go` and `handlers_test.go` unstaged when they contain unrelated hunks.

---

### Task 3: Submit a website and count the admin notification

**Files:**
- Create: `backend/internal/account/websites.go`
- Modify: `backend/internal/account/queue.go` (`QueueCount` adds pending websites)
- Modify: `backend/internal/auth/auth.go` (add the pending-website count to the inline `admin_queue_count` query. Do not import `account` from `auth`; that import cycle is already closed the other way.)
- Modify: `backend/main.go` and `backend/handlers_test.go` (route `POST/GET /api/websites` — GET is public, POST requires a session)
- Test: `backend/website_test.go`

**Interfaces:**
- Consumes: `webdomain.CanonicalDomain`, `webdomain.Plain`, `auth.UserFromRequest`, `QueueCount`
- Produces:
  - `POST /api/websites` body `{ "domain", "title", "description" }`
  - 201 `{ "id", "domain", "title", "description", "status": "pending" }`
  - 401 when signed out
  - 403 `{ "error": "website_banned" }` when `users.website_banned` is true, and no row is inserted
  - 400 `{ "error": "invalid", "field": "domain"|"title"|"description" }`
  - 409 `{ "error": "domain_pending" }` when a pending row holds the key
  - 409 `{ "error": "domain_taken", "listing": { "name", "slug" } }` when a public listing holds the key
  - 409 `{ "error": "domain_taken", "website": { "title", "domain" } }` when an approved website with no public listing holds the key
  - `QueueCount()` and the `admin_queue_count` query in `auth.HandleMe` both add `SELECT COUNT(*) FROM websites WHERE status = 'pending'`
  - Pending rows are absent from `GET /api/websites`

- [ ] **Step 1: Write the failing tests**

Append to `backend/website_test.go`:

```go
func TestWebsiteSubmitPendingRaisesQueue(t *testing.T) {
	cookie := mustLogin("website-submit@test.lamsza")
	adminCookie := mustLogin("admin@test.lamsza")
	before := adminQueueCount(t, adminCookie)
	rr := doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "https://www.submit-example.com/a",
		"title":  "<b>Submit Example</b>",
		"description": "A short page.",
	}, cookie)
	if rr.Code != 201 {
		t.Fatalf("submit: %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	if created["domain"] != "submit-example.com" || created["title"] != "Submit Example" || created["status"] != "pending" {
		t.Fatalf("created %#v", created)
	}
	if adminQueueCount(t, adminCookie) != before+1 {
		t.Fatalf("queue count did not rise")
	}
	pub := doRequest(t, "GET", "/api/websites", nil)
	if strings.Contains(pub.Body.String(), "submit-example.com") {
		t.Fatal("pending website was public")
	}
	dup := doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "submit-example.com", "title": "Other", "description": "Other page.",
	}, cookie)
	if dup.Code != 409 || !strings.Contains(dup.Body.String(), "domain_pending") {
		t.Fatalf("dup: %d %s", dup.Code, dup.Body.String())
	}
}

func TestWebsiteSubmitRejectsBannedAndSignedOut(t *testing.T) {
	rr := doRequest(t, "POST", "/api/websites", map[string]string{
		"domain": "banned-example.com", "title": "T", "description": "D",
	})
	if rr.Code != 401 {
		t.Fatalf("signed out: %d", rr.Code)
	}
	cookie := mustLogin("website-banned@test.lamsza")
	if _, err := db.DB.Exec(`UPDATE users SET website_banned = true WHERE email = $1`, "website-banned@test.lamsza"); err != nil {
		t.Fatal(err)
	}
	rr = doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "banned-example.com", "title": "T", "description": "D",
	}, cookie)
	if rr.Code != 403 || !strings.Contains(rr.Body.String(), "website_banned") {
		t.Fatalf("banned: %d %s", rr.Code, rr.Body.String())
	}
}

func adminQueueCount(t *testing.T, cookie *http.Cookie) int {
	t.Helper()
	rr := doRequestWithCookie(t, "GET", "/api/auth/me", nil, cookie)
	var me map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &me)
	n, ok := me["admin_queue_count"].(float64)
	if !ok {
		t.Fatalf("no queue count %#v", me)
	}
	return int(n)
}
```

Add the imports `encoding/json`, `net/http`, `strings`, and `backend/internal/db` if they are not already in the file.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test -count=1 -run 'TestWebsiteSubmit' .`

Expected: FAIL with 404, because the route is not registered.

- [ ] **Step 3: Implement submit and the queue term**

`HandleWebsites`:
- `GET` returns `{ "websites": [] }` of approved rows only: `id`, `title`, `description`, `domain` (`domain_key`), `url` (`https://` + domain key). No slug, no listing id.
- `POST` requires `auth.UserFromRequest`. Read `website_banned`. Validate with `Plain(title, 120)` and `Plain(description, 300)` and `CanonicalDomain`. On failure return 400 with `field`.
- Look up `websites.domain_key`. Pending → 409 `domain_pending`. Approved with a published `entry_id` → 409 `domain_taken` plus `listing.name` and `listing.slug`. Approved otherwise → 409 `domain_taken` plus `website.title` and `website.domain`.
- Also treat a published entry URL as taken when the backfill row exists. Do not insert a second row.
- Insert `status = 'pending'`, `submitted_host` = the host the user typed (lowercase hostname, not the registrable key), `user_id` = the session user.
- `QueueCount` SQL becomes the current two counts plus pending websites. Add that same third term to the inline query in `auth.HandleMe`. Leave the query inline there.

Register:

```go
mux.HandleFunc("/api/websites", middleware.ApplyCORS(account.HandleWebsites))
```

in `main.go` and on `testMux`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd backend && go test -count=1 -run 'TestWebsiteSubmit|TestMigrateWebsites' .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/account/websites.go backend/internal/account/queue.go backend/website_test.go
git commit -m "$(cat <<'EOF'
Let a signed-in user submit a website that waits for admin approval.

EOF
)"
```

Leave `main.go`, `handlers_test.go`, and `auth.go` unstaged if they contain unrelated hunks. `auth.go` is also already dirty; include only the `QueueCount` call when the rest of the file is unrelated, otherwise report it left unstaged.

---

### Task 4: Approve, Reject, and Ban User

**Files:**
- Modify: `backend/internal/account/websites.go`
- Modify: `backend/internal/account/queue.go` (include pending websites on `GET /api/admin/listing-queue`)
- Modify: `backend/main.go` and `backend/handlers_test.go` (route `POST /api/admin/websites`)
- Test: `backend/website_test.go`

**Interfaces:**
- Consumes: pending `websites` rows, `QueueCount`
- Produces:
  - `GET /api/admin/listing-queue` adds `websites`: `{ "id", "domain", "title", "description", "submitter" }` where `submitter` is the user's email. Pending only.
  - `POST /api/admin/websites` body `{ "id", "action": "approve"|"reject"|"ban" }`, admin only
  - Approve sets `status = 'approved'` and the row disappears from the queue
  - Reject deletes the row. The key can be submitted again
  - Ban deletes the row and sets `users.website_banned = true` for that `user_id`. A later submit from that account is 403 and does not create a notification. `mustLogin` for that email still returns a session cookie.

- [ ] **Step 1: Write the failing tests**

```go
func TestWebsiteApproveRejectAndBan(t *testing.T) {
	user := mustLogin("website-review@test.lamsza")
	admin := mustLogin("admin@test.lamsza")
	id := submitWebsite(t, user, "review-example.com", "Review", "A page.")
	q := queueWebsites(t, admin)
	if len(q) != 1 || q[0]["domain"] != "review-example.com" || q[0]["title"] != "Review" || q[0]["submitter"] == "" {
		t.Fatalf("queue %#v", q)
	}
	rr := doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": id, "action": "approve"}, admin)
	if rr.Code != 200 {
		t.Fatalf("approve %d %s", rr.Code, rr.Body.String())
	}
	if strings.Contains(queueBody(t, admin), "review-example.com") {
		t.Fatal("approved website stayed in the queue")
	}
	pub := doRequest(t, "GET", "/api/websites", nil)
	if !strings.Contains(pub.Body.String(), "review-example.com") {
		t.Fatal("approved website was not public")
	}

	id = submitWebsite(t, user, "reject-example.com", "Reject", "Gone.")
	rr = doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": id, "action": "reject"}, admin)
	if rr.Code != 200 {
		t.Fatalf("reject %d", rr.Code)
	}
	again := submitWebsite(t, user, "reject-example.com", "Reject", "Again.")
	if again <= 0 {
		t.Fatal("rejected key was not freed")
	}
	doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": again, "action": "reject"}, admin)

	banID := submitWebsite(t, user, "ban-example.com", "Ban", "Nope.")
	rr = doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": banID, "action": "ban"}, admin)
	if rr.Code != 200 {
		t.Fatalf("ban %d %s", rr.Code, rr.Body.String())
	}
	rr = doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "after-ban.com", "title": "T", "description": "D",
	}, user)
	if rr.Code != 403 {
		t.Fatalf("banned resubmit %d", rr.Code)
	}
	mustLogin("website-review@test.lamsza")
}
```

`submitWebsite` POSTs and returns the numeric `id`. `queueWebsites` GETs `/api/admin/listing-queue` and returns the `websites` array. After the ban, call `mustLogin("website-review@test.lamsza")` again. A returned cookie means sign-in still works. Do not post a raw Google credential.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test -count=1 -run TestWebsiteApproveRejectAndBan .`

Expected: FAIL, route 404 or `websites` missing from the queue JSON.

- [ ] **Step 3: Implement the three actions**

Add `websites` to `queueResponse`, queried as pending rows joined to `users.email`. Empty slice when there are none, never JSON `null`.

`HandleAdminWebsite`:
- Admin route, POST only.
- `approve`: `UPDATE websites SET status = 'approved' WHERE id = $1 AND status = 'pending'`. Zero rows → 404.
- `reject`: `DELETE FROM websites WHERE id = $1 AND status = 'pending'`.
- `ban`: in one transaction, read `user_id`, delete the pending row, `UPDATE users SET website_banned = true WHERE id = $1`. A null `user_id` (backfill) cannot be banned; return 400.
- Unknown action → 400.

Register `POST /api/admin/websites` with the admin wrapper in `main.go` and `testMux`.

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd backend && go test -count=1 -run 'TestWebsite' .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/account/websites.go backend/internal/account/queue.go backend/website_test.go
git commit -m "$(cat <<'EOF'
Approve, reject, or ban a website from the admin queue.

EOF
)"
```

---

### Task 5: Search hits for a domain query and a name query

**Files:**
- Modify: `backend/internal/account/websites.go`
- Modify: `backend/internal/search/unified.go`
- Modify: `backend/main.go` and `backend/handlers_test.go` only if `/api/search` is missing from `testMux` (it is missing today; add `search.HandleUnifiedSearch`)
- Test: `backend/website_test.go`

**Interfaces:**
- Consumes: approved `websites` rows, published `entries`, `entry_members` owner rows
- Produces on `GET /api/search?q=`:
  - `websites`: `{ "id", "title", "description", "domain", "url" }`. `url` is `https://` + domain. Empty array, not null.
  - `website_query`: true only when the whole query canonicalizes to a stored `domain_key`
  - Domain query: `websites` has that one row first. `entries` is only the published listing with `websites.entry_id`, or empty. That entry's `claimed` is true when an active owner exists.
  - Any other query: `websites` contains approved rows whose title or description matches and whose listing is missing or unpublished. A website with a published listing is omitted. `entries` stays the existing published-listing search.
  - Pending and rejected domains never appear.

- [ ] **Step 1: Write the failing tests**

```go
func TestWebsiteSearchDomainAndTitle(t *testing.T) {
	user := mustLogin("website-search@test.lamsza")
	admin := mustLogin("admin@test.lamsza")
	id := submitWebsite(t, user, "search-example.com", "Search Title", "Unique blurb zzq")
	doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": id, "action": "approve"}, admin)

	domain := searchJSON(t, "https://www.search-example.com/path")
	if domain["website_query"] != true {
		t.Fatalf("expected domain query %#v", domain["website_query"])
	}
	sites := domain["websites"].([]interface{})
	if len(sites) != 1 {
		t.Fatalf("sites %#v", sites)
	}
	hit := sites[0].(map[string]interface{})
	if hit["url"] != "https://search-example.com" || hit["title"] != "Search Title" {
		t.Fatalf("hit %#v", hit)
	}
	if ents, ok := domain["entries"].([]interface{}); ok && len(ents) != 0 {
		t.Fatalf("unclaimed domain query returned entries %#v", ents)
	}

	title := searchJSON(t, "Unique blurb zzq")
	if title["website_query"] == true {
		t.Fatal("title query was treated as a domain")
	}
	if len(title["websites"].([]interface{})) != 1 {
		t.Fatalf("title sites %#v", title["websites"])
	}
}
```

`searchJSON` GETs `/api/search?q=` + `url.QueryEscape(q)` and decodes the object. Register `/api/search` on `testMux` before running this test.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test -count=1 -run TestWebsiteSearchDomainAndTitle .`

Expected: FAIL, `websites` missing.

- [ ] **Step 3: Implement search selection**

Add `account.SearchWebsites(q string) (hits []WebsiteHit, domainQuery bool, publishedEntryID int)`.

- If `CanonicalDomain(q)` equals a `domain_key`, set `domainQuery` true, return that approved row, and set `publishedEntryID` when `entry_id` points at a published entry. Ignore pending rows for public hits; a pending key still means this is not a public domain hit.
- Otherwise select approved websites where `unaccent(lower(title))` or `unaccent(lower(description))` ILIKE the query, and (`entry_id` is null or the entry is unpublished). Limit 8.

In `HandleUnifiedSearch`, call this before the entry goroutine. When `domainQuery` is true, skip the fuzzy entry search and, when `publishedEntryID > 0`, load that one published entry with the same public entry shape the fuzzy query returns, including:

```sql
EXISTS (
    SELECT 1 FROM entry_members m
    WHERE m.entry_id = e.id AND m.role = 'owner' AND m.status = 'active'
) AS claimed
```

Add that `claimed` expression to the fuzzy entry SELECT as well, and scan it into `Entry.Claimed`. Put the expression in the `GROUP BY` if Postgres requires it. Add `Websites` and `WebsiteQuery` to `UnifiedSearchResult`. The empty-query response includes `"websites": []` and `"website_query": false`.

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd backend && go test -count=1 -run 'TestWebsiteSearch|TestWebsiteApprove' .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/account/websites.go backend/internal/search/unified.go backend/website_test.go
git commit -m "$(cat <<'EOF'
Return a website hit for a domain query and a title query.

EOF
)"
```

---

### Task 6: Claiming a website creates the existing listing

**Files:**
- Modify: `backend/internal/account/listings.go` (`handleCreateListing`)
- Test: `backend/website_test.go`

**Interfaces:**
- Consumes: approved `websites` row with `entry_id` null, existing create-listing insert and `entry_members` owner insert
- Produces:
  - `POST /api/account/listings` accepts optional `website_id`
  - When set: the user must not be `website_banned` (403 `website_banned`). The website must be approved and unlinked (404 otherwise). Required body fields stay `name`, `location_id`, `category_id`, `type_id`. `url` is forced to `https://` + `domain_key`. Creator is the owner. `published` and `verified` are false. `websites.entry_id` is set. `websites.title` and `websites.description` are unchanged.
  - A second `website_id` create for that row is 409 `domain_taken`.
  - While the listing is unpublished, `GET /api/search` for the domain still returns only the website hit.
  - After `POST /api/admin/listing-queue/publish`, a domain query returns the website hit and that one entry with `claimed: true`. A query for the listing name returns the entry and does not include the website. `verified` stays false.
  - `POST /api/account/listings/claim` by a second user on that published listing returns a pending member, using the existing handler.

- [ ] **Step 1: Write the failing test**

```go
func TestClaimWebsiteCreatesUnpublishedListing(t *testing.T) {
	owner := mustLogin("website-owner@test.lamsza")
	other := mustLogin("website-joiner@test.lamsza")
	admin := mustLogin("admin@test.lamsza")
	webID := submitWebsite(t, owner, "claim-example.com", "Claim Title", "Claim blurb")
	doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": webID, "action": "approve"}, admin)

	locID := mustLocID(t)
	var catID, typeID int
	db.DB.QueryRow(`SELECT id FROM entry_categories ORDER BY id ASC LIMIT 1`).Scan(&catID)
	db.DB.QueryRow(`SELECT id FROM entry_types ORDER BY id ASC LIMIT 1`).Scan(&typeID)
	rr := doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"website_id": webID, "name": "Manifesto", "location_id": locID,
		"category_id": catID, "type_id": typeID, "url": "https://evil.example",
	}, owner)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("create %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	if created["published"] != false || created["verified"] != false || created["role"] != "owner" {
		t.Fatalf("created %#v", created)
	}
	var title, url string
	var published, verified bool
	db.DB.QueryRow(`SELECT w.title, e.url, e.published, e.verified FROM websites w JOIN entries e ON e.id = w.entry_id WHERE w.id = $1`, webID).Scan(&title, &url, &published, &verified)
	if title != "Claim Title" || url != "https://claim-example.com" || published || verified {
		t.Fatalf("stored title %q url %q published %v verified %v", title, url, published, verified)
	}
	before := searchJSON(t, "claim-example.com")
	if len(before["entries"].([]interface{})) != 0 || len(before["websites"].([]interface{})) != 1 {
		t.Fatalf("before publish %#v", before)
	}
	entryID := int(created["id"].(float64))
	doRequestWithCookie(t, "POST", "/api/admin/listing-queue/publish", map[string]interface{}{"entry_id": entryID}, admin)
	after := searchJSON(t, "www.claim-example.com")
	ents := after["entries"].([]interface{})
	if len(ents) != 1 || ents[0].(map[string]interface{})["claimed"] != true {
		t.Fatalf("domain entries %#v", ents)
	}
	if after["websites"].([]interface{})[0].(map[string]interface{})["url"] != "https://claim-example.com" {
		t.Fatal("website hit missing")
	}
	byName := searchJSON(t, "Manifesto")
	if byName["website_query"] == true || len(byName["websites"].([]interface{})) != 0 {
		t.Fatalf("name query leaked website %#v", byName["websites"])
	}
	if byName["entries"].([]interface{})[0].(map[string]interface{})["claimed"] != true {
		t.Fatal("name query missing claimed")
	}
	rr = doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"website_id": webID, "name": "Second", "location_id": locID,
		"category_id": catID, "type_id": typeID,
	}, other)
	if rr.Code != 409 {
		t.Fatalf("second listing %d", rr.Code)
	}
	banned := mustLogin("website-claim-ban@test.lamsza")
	db.DB.Exec(`UPDATE users SET website_banned = true WHERE email = $1`, "website-claim-ban@test.lamsza")
	freeID := submitWebsite(t, owner, "claim-free.com", "Free", "Free page.")
	doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": freeID, "action": "approve"}, admin)
	rr = doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"website_id": freeID, "name": "Banned claim", "location_id": locID,
		"category_id": catID, "type_id": typeID,
	}, banned)
	if rr.Code != 403 {
		t.Fatalf("banned claim %d %s", rr.Code, rr.Body.String())
	}
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{"entry_id": entryID}, other)
	var member map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &member)
	if member["role"] != "member" || member["status"] != "pending" {
		t.Fatalf("join %#v", member)
	}
}
```

Confirm `HandleListings` POST status code from `handleCreateListing` (it writes 200 with no explicit status). Assert that actual code.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test -count=1 -run TestClaimWebsiteCreatesUnpublishedListing .`

Expected: FAIL, listing URL stays `https://evil.example` or `website_id` is ignored.

- [ ] **Step 3: Link the create path**

In `handleCreateListing`, when `website_id > 0`:
- 403 if the user is website-banned.
- `SELECT domain_key, status, entry_id FROM websites WHERE id = $1 FOR UPDATE`.
- 404 unless status is `approved` and `entry_id` is null.
- Set the inserted URL to `https://` + `domain_key`.
- After the owner insert, `UPDATE websites SET entry_id = $1 WHERE id = $2 AND entry_id IS NULL`. If zero rows, roll back and return 409 `domain_taken`.
- Do not update `title` or `description`.

When `website_id` is 0 and `url` canonicalizes to an existing `domain_key`, return 409 `domain_taken` so a normal create cannot occupy a submitted website.

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd backend && go test -count=1 -run 'TestClaimWebsite|TestWebsiteSearch' .`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/account/listings.go backend/website_test.go
git commit -m "$(cat <<'EOF'
Create a normal unpublished listing when a website is claimed.

EOF
)"
```

---

### Task 7: Add form, index rows, and search cards

**Files:**
- Create: `src/lib/components/AddWebsiteForm.svelte`
- Create: `src/lib/components/WebsiteCard.svelte`
- Modify: `src/routes/(public)/index/+page.svelte`
- Modify: `src/lib/components/SearchEngine.svelte`
- Modify: `src/lib/components/EntryCard.svelte`
- Modify: `src/styles/global.css` only for classes the new markup needs, next to existing entry-card rules
- Test: `tests/websiteDomain.test.js` (form field list is covered by the component contract below; domain tests already exist)

**Interfaces:**
- Consumes: `POST /api/websites`, `GET /api/websites`, `GET /api/search` `websites` and `website_query`, `openLogin` from `src/lib/openLogin.js`, `auth` store
- Produces:
  - On `/index`, a button labeled `Add your website for free` that opens the form
  - Signed out: `openLogin()`. The form stays mounted. When `$auth.loggedIn` becomes true, the form is still open
  - Fields named `domain`, `title`, `description` only. Submit posts those three. 201 replaces the form with the text `Waiting for admin approval.`
  - 400 keeps the form open and shows the `field` name. 409 `domain_taken` shows the listing name when `listing` is present, otherwise the website title and domain. 409 `domain_pending` shows `This domain is already waiting for approval.` 403 shows `You cannot add a website.`
  - `/index` with category `osszes` loads `GET /api/websites` and renders `WebsiteCard` rows in a section titled `Weboldalak`. Category filters do not render that section. A card’s only link is the external `url` (`target="_blank"`, `rel="noopener"`). It shows title, description, and domain. It does not link to `/bejegyzes/`
  - Search: when `website_query` is true, the `Weboldalak` section is above the `Index` entries section. Otherwise it is below that section. `EntryCard` shows the existing owner mark when `entry.claimed` is true

- [ ] **Step 1: Write WebsiteCard and AddWebsiteForm**

`WebsiteCard.svelte` props: `website` with `title`, `description`, `domain`, `url`. One external anchor. No internal route.

`AddWebsiteForm.svelte` props: `onClose`. Local state for the three fields, `error`, and `pending`. On submit, if `!$auth.loggedIn`, call `openLogin()` and return. Otherwise `apiFetch("/api/websites", { method: "POST", body })`. Map the status codes above. Disable the button while the request is in flight.

Both files use `$props` and `$state`. No extra fields.

- [ ] **Step 2: Run the Svelte autofixer**

Call the Svelte MCP `svelte-autofixer` on both new files and fix until it reports no issues.

- [ ] **Step 3: Mount them on the index and in search**

In `index/+page.svelte`, add the button in the page hero actions. Opening it sets `addWebsiteOpen`. Render `AddWebsiteForm` when that flag is true. Fetch `/api/websites` in the existing load path. Render the `Weboldalak` section only when `currentCategory === "osszes"` and the array is non-empty.

In `SearchEngine.svelte`, read `searchResults.websites`. Render `WebsiteCard` for each. Place that block immediately before the `Index` entries block when `searchResults.website_query` is true, and immediately after it otherwise. Include website length in `totalCount` so a website-only result is visible.

In `EntryCard.svelte`, next to the title, when `entry?.claimed` is true, render:

```svelte
<span class="entry-profile__owned">Foglalt</span>
```

Only if that is still the text inside `.entry-profile__owned` in `EntryProfile.svelte`. If that span’s text has changed, copy the current text. Do not introduce a different word.

- [ ] **Step 4: Check the domain unit test still passes**

Run: `node --test tests/websiteDomain.test.js`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add src/lib/components/AddWebsiteForm.svelte src/lib/components/WebsiteCard.svelte src/routes/(public)/index/+page.svelte src/lib/components/SearchEngine.svelte src/lib/components/EntryCard.svelte src/styles/global.css
git commit -m "$(cat <<'EOF'
Show the free website form, weboldal rows, and search hits.

EOF
)"
```

Omit `global.css` from the commit if no rules were added.

---

### Task 8: Admin notification buttons

**Files:**
- Modify: `src/routes/admin/+page.svelte`
- Modify: `src/lib/accountPrefs.js` only if the queue count already refreshes from `/api/auth/me` after admin actions; if it does, do not change it

**Interfaces:**
- Consumes: `GET /api/admin/listing-queue` field `websites`, `POST /api/admin/websites` `{ id, action }`
- Produces: on the admin welcome **Üzenetek** list, one message per pending website. The text includes the submitter email, domain, title, and short description. The message has three buttons, `Approve`, `Reject`, and `Ban User`. Each posts the matching action and then refetches the queue. The message disappears after a 200. `Ban User` does not add a confirm dialog beyond the button itself.

- [ ] **Step 1: Render one message per pending website**

Extend `fetchListingQueue` to store `data.websites` on `listingQueueWebsites` (default `[]`). In `buildDashboardMessages`, after the existing queue summary, push one object per website:

```js
{
    id: "website-" + site.id,
    level: "info",
    text: `${site.submitter} added ${site.domain}: ${site.title}. ${site.description}`,
    action: "website",
    websiteId: site.id,
}
```

In the `{#each dashboardMessages}` block, when `msg.action === "website"`, render the three buttons. Each calls `reviewWebsite(msg.websiteId, "approve"|"reject"|"ban")`, which POSTs `/api/admin/websites` and then `fetchListingQueue()`. Show `listingQueueError` on a non-200 body.

Include `listingQueueWebsites.length` in the waiting count that drives the queue summary, so the summary is not “nothing is waiting” while a website message is on screen.

- [ ] **Step 2: Run the Svelte autofixer**

Run it on the edited admin page. Ignore `$:` complaints. Fix any issue the tool reports in the new button block.

- [ ] **Step 3: Commit**

```bash
git add src/routes/admin/+page.svelte
git commit -m "$(cat <<'EOF'
Let an admin approve, reject, or ban a submitted website from the notification.

EOF
)"
```

---

## Self-review

Spec coverage:

- Domain key, including `www`, path, and subdomain: Task 1 and the duplicate checks in Task 3.
- Three-field free add, sign-in return, pending confirmation: Task 7. Server validation: Task 3.
- Admin notification with Approve, Reject, Ban User, queue count: Tasks 3, 4, and 8.
- Ban refuses later submit and claim, sign-in still works, listings stay: Tasks 4 and 6.
- Unclaimed website has no listing page and a direct link: Tasks 5 and 7.
- Domain query shows website then the published listing; name query shows the listing with Claimed and omits the website: Tasks 5, 6, and 7.
- Claim uses the existing create and join flows, listing stays unpublished, title is not rewritten, `verified` stays false: Task 6.
- Existing listing URLs occupy the key: Task 2.

Placeholder scan: no TBD, no “similar to task N”, no unspecified error handling. UI copy that the spec named is written out.

Type consistency: `website_id`, `domain_key`, `website_query`, `websites[].url`, actions `approve` / `reject` / `ban`, and errors `website_banned`, `domain_pending`, `domain_taken` are the same names in Tasks 3 through 8.
