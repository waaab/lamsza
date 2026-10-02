# Listing suggestions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A logged-in user who is not the owner or a member can send one suggestion of changed listing fields, and an admin can preview it and accept or deny the whole request.

**Architecture:** `entry_suggestions` stores one open row per listing. `POST /api/entry/suggestions` rejects a second open row. Accept copies the stored field values onto `entries` and does not change `slug`. Public `GET /api/entry` stays stripped; `GET /api/entry/suggestion-form` returns the stored editable fields to an eligible signed-in user.

**Tech Stack:** Go `net/http` + Postgres, Svelte 5 for the new suggestion dialog, existing listing page stays `let` / `$:`, Go `go test`, Node `node --test`.

**Spec:** `docs/superpowers/specs/2026-09-27-listing-profile-design.md`

**Depends on:** `docs/superpowers/plans/2026-09-27-listing-profile-view.md` (sidebar button and `hours_enabled` / `social_links`).

## Global Constraints

- Audience: logged-in, not an active owner, not an active member. A pending member counts as not a member and may send a suggestion when none is open.
- The active owner sees **Szerkesztés**, never this form. A waiting suggestion does not replace **Szerkesztés**.
- One open suggestion per listing. While it is open, every eligible user sees **Javaslat elküldve** and the control is disabled. The sender cannot withdraw it.
- A second POST while one is open returns **409** with `{"error":"suggestion_pending"}`.
- The form is prefilled from stored values. Only changed fields are stored. A cleared field is a change and accept writes the empty value.
- Fields: `name`, `tags`, `notes`, `location_id`, `address`, `hours`, `delivery_hours`, `url`, `phone`, `social_links`, `languages`, plus optional `note`.
- Hours are included only when `hours_enabled` is true. Delivery hours are included only when `delivery_enabled` is true and `offersDelivery` is true.
- Photos are not in the form or the stored diff.
- Accept writes those fields and keeps the existing slug. Deny deletes the open request by setting `status` to `denied`.
- Admin queue count includes open suggestions.
- Hungarian: **Javaslat módosításra**, **Javaslat elküldve**, **Elfogad**, **Elutasít**, empty diff **Nincs módosított mező.**, save error **A mentés nem sikerült**.
- Commit only the files named in the task. Never `git add -A`. Do not merge to `main`.

---

### Task 1: Suggestion table and diff helper

**Files:**
- Create: `backend/migrations/entry_suggestions.sql`
- Create: `backend/internal/handlers/migrate_entry_suggestions.go`
- Create: `backend/internal/account/suggestions.go`
- Test: `backend/internal/account/suggestion_diff_test.go`
- Modify: `backend/main.go`
- Modify: `backend/handlers_test.go`

**Interfaces:**
- Produces table `entry_suggestions` (`id`, `entry_id`, `user_id`, `changes JSONB`, `note TEXT`, `status TEXT`, `created_at`). Partial unique index on `entry_id` where `status = 'open'`.
- Produces `diffSuggestion(before, after map[string]any) map[string]any` in `suggestions.go`.

- [ ] **Step 1: Write the failing test**

`backend/internal/account/suggestion_diff_test.go`:

```go
package account

import "testing"

func TestDiffSuggestionKeepsOnlyChanges(t *testing.T) {
	before := map[string]any{"name": "Régi", "phone": "111", "notes": "szöveg"}
	after := map[string]any{"name": "Régi", "phone": "", "notes": "szöveg"}
	got := diffSuggestion(before, after)
	if _, ok := got["name"]; ok {
		t.Fatal("unchanged name must be omitted")
	}
	if got["phone"] != "" {
		t.Fatalf("cleared phone: %v", got["phone"])
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd backend && go test ./internal/account/ -count=1 -run TestDiffSuggestionKeepsOnlyChanges`

Expected: FAIL, `diffSuggestion` is undefined.

- [ ] **Step 3: Add the table and the function**

`backend/migrations/entry_suggestions.sql`:

```sql
CREATE TABLE IF NOT EXISTS entry_suggestions (
    id SERIAL PRIMARY KEY,
    entry_id INT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    changes JSONB NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'accepted', 'denied')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS entry_suggestions_one_open
    ON entry_suggestions (entry_id)
    WHERE status = 'open';
```

Wire `MigrateEntrySuggestions()` the same way as `MigrateEntryReviews()`. Call it from `main.go` and the test init in `handlers_test.go`.

```go
func diffSuggestion(before, after map[string]any) map[string]any {
	out := map[string]any{}
	for key, next := range after {
		prev, ok := before[key]
		if !ok || !suggestionValuesEqual(prev, next) {
			out[key] = next
		}
	}
	return out
}
```

`suggestionValuesEqual` compares the JSON encoding of both values.

- [ ] **Step 4: Run the test**

Run: `cd backend && go test ./internal/account/ -count=1 -run TestDiffSuggestionKeepsOnlyChanges`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/entry_suggestions.sql backend/internal/handlers/migrate_entry_suggestions.go backend/internal/account/suggestions.go backend/internal/account/suggestion_diff_test.go backend/main.go backend/handlers_test.go
git commit -m "Store one open listing suggestion per entry."
```

---

### Task 2: Create, preview, accept, and deny

**Files:**
- Modify: `backend/internal/account/suggestions.go`
- Modify: `backend/internal/account/queue.go`
- Modify: `backend/internal/auth/auth.go`
- Modify: `backend/internal/handlers/public.go`
- Modify: `backend/main.go`
- Modify: `backend/handlers_test.go`
- Test: `backend/entry_suggestions_test.go`

**Interfaces:**
- `GET /api/entry/suggestion-form?slug=` - 401 signed out, 403 owner or active member, 200 editable snapshot for an eligible user. The snapshot includes phone and social links even when public `GET /api/entry` hides them.
- `POST /api/entry/suggestions` body `{ "entry_id", "fields", "note" }` - 401, 403, 409 `{"error":"suggestion_pending"}`, 400 `{"error":"empty_diff"}`, 200 `{ "id", "status": "open" }`.
- `POST /api/admin/listing-queue/suggestion` body `{ "id", "action": "accept" | "deny" }`. Accept writes the changed columns and does not change `slug`. Deny sets `status` to `denied`.
- `GET /api/admin/listing-queue` adds `suggestions`: `{ "id", "entry_id", "entry_name", "email", "changes", "note" }` for open rows.
- `GET /api/entry` adds `suggestion_pending: true` when an open row exists.
- `admin_queue_count` adds the open-suggestion count in both `queue.go` `QueueCount` and `auth.go`.

Allowed field keys: `name`, `tags`, `notes`, `location_id`, `address`, `hours`, `delivery_hours`, `url`, `phone`, `social_links`, `languages`. Drop every other key, including `slug`, `photos`, `verified`, and `claimed`. Drop `hours` when `hours_enabled` is false. Drop `delivery_hours` when `delivery_enabled` is false or the category is not **Vendéglő** or the name is **Lámsza.com**.

- [ ] **Step 1: Write the failing API tests**

In `backend/entry_suggestions_test.go` (`package main`), using the cookie helpers in `handlers_test.go`:

1. Signed-out POST returns 401.
2. Active owner POST returns 403.
3. Eligible user POST with a changed phone returns 200. A second eligible user POST returns 409.
4. Admin accept changes `entries.phone` and leaves `entries.slug` unchanged when `fields.name` differs.
5. After deny, a new POST returns 200.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd backend && go test . -count=1 -timeout 180s -run TestSuggestion`

Expected: FAIL, route not registered.

- [ ] **Step 3: Implement the handlers**

Register the routes in `main.go` and `handlers_test.go`. Accept locks the open row, updates only the keys in `changes`, then sets `status = 'accepted'`. A `name` change updates `entries.name` only. Deny sets `status = 'denied'` and does not write `entries`.

`HandleListingQueue` appends open suggestions. Both queue-count queries add `SELECT COUNT(*) FROM entry_suggestions WHERE status = 'open'`.

- [ ] **Step 4: Run the tests**

Run: `cd backend && go test . -count=1 -timeout 180s -run 'TestSuggestion|TestAdminQueue'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/account/suggestions.go backend/internal/account/queue.go backend/internal/auth/auth.go backend/internal/handlers/public.go backend/main.go backend/handlers_test.go backend/entry_suggestions_test.go
git commit -m "Let an admin accept or deny one open listing suggestion."
```

---

### Task 3: Suggestion dialog and admin preview

**Files:**
- Create: `src/lib/components/SuggestionFormDialog.svelte`
- Create: `src/lib/suggestionDiff.js`
- Test: `tests/suggestionDiff.test.js`
- Modify: `src/routes/(public)/bejegyzes/[slug]/+page.svelte`
- Modify: `src/lib/components/EntryProfile.svelte`
- Modify: `src/routes/admin/+page.svelte`

**Interfaces:**
- `diffSuggestionFields(before, after) -> object` in `src/lib/suggestionDiff.js`.
- `EntryProfile` prop `suggestionState`: `"hidden" | "open" | "waiting"`.
- `waiting` renders a disabled **Javaslat elküldve**.
- Admin queue renders each open suggestion with the changed fields, the note, **Elfogad**, and **Elutasít**.

- [ ] **Step 1: Write the failing diff test**

```js
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { diffSuggestionFields } from '../src/lib/suggestionDiff.js';

test('diffSuggestionFields omits unchanged fields and keeps a cleared phone', () => {
	const got = diffSuggestionFields(
		{ name: 'Régi', phone: '111', notes: 'szöveg' },
		{ name: 'Régi', phone: '', notes: 'szöveg' }
	);
	assert.deepEqual(got, { phone: '' });
});
```

- [ ] **Step 2: Run it to verify it fails**

Run: `node --test tests/suggestionDiff.test.js`

Expected: FAIL, module missing.

- [ ] **Step 3: Implement the dialog**

`diffSuggestionFields` returns only keys whose `JSON.stringify` value changed.

`SuggestionFormDialog` loads `GET /api/entry/suggestion-form?slug=`. Hours inputs stay hidden when `hours_enabled` is false. Delivery inputs stay hidden unless the category is **Vendéglő**, the name is not **Lámsza.com**, and `delivery_enabled` is true. Submit POSTs `{ entry_id, fields: diff, note }`. An empty diff shows **Nincs módosított mező.** and does not POST.

`+page.svelte` sets `suggestionState` to `hidden` when signed out, when the viewer is the active owner, or when the viewer is an active member. It sets `waiting` when `entry.suggestion_pending` is true. Otherwise it sets `open` for an eligible user. The button opens the dialog.

In `fetchListingQueue` inside `src/routes/admin/+page.svelte`, store `data.suggestions`. Under the existing members block, render each row. **Elfogad** POSTs `{ id, action: "accept" }` to `/api/admin/listing-queue/suggestion`. **Elutasít** POSTs `action: "deny"`. Then call `fetchListingQueue()` again.

Run Svelte MCP `svelte-autofixer` on `SuggestionFormDialog.svelte` and `EntryProfile.svelte` until clean.

- [ ] **Step 4: Run the test and click through**

Run: `node --test tests/suggestionDiff.test.js`

Expected: PASS

As an eligible user on a gazdátlan listing, change the phone and send the suggestion. The button becomes disabled **Javaslat elküldve**. As admin, accept it. The stored phone changes and the slug does not.

- [ ] **Step 5: Commit**

```bash
git add src/lib/suggestionDiff.js tests/suggestionDiff.test.js src/lib/components/SuggestionFormDialog.svelte src/lib/components/EntryProfile.svelte 'src/routes/(public)/bejegyzes/[slug]/+page.svelte' src/routes/admin/+page.svelte
git commit -m "Show the listing suggestion form and the admin preview."
```
