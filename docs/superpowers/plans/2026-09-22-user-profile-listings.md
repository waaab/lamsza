# Listing ownership Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Signed-in users can claim or create listings, own them with members, and see them on **Bejegyzéseim**. Admins publish listings and approve members from an in-app queue. The public page shows **Claimed** and **Ellenőrzött** independently.

**Architecture:** `entries.claimed` migrates to `entries.verified`. `entries.published` defaults true for existing rows. `entry_members` stores one owner and any members (`active` or `pending`). **Claimed** in the public JSON is true only when an active owner exists. The admin queue is those unpublished rows plus pending member rows. No email and no separate notification table.

**Tech Stack:** Go `net/http` + Postgres, Svelte 5 for new profile UI, existing admin page stays on its current style, Go `go test`.

**Spec:** `docs/superpowers/specs/2026-09-22-user-profile-design.md`

**Depends on:** the account plan and the favorites plan, in that order. `profileTabIds` already ends with `kedvencek` before this plan starts.

## Global Constraints

- One active owner per listing. No ownership transfer.
- Claim on a published listing with no owner is immediate. If an owner exists, the claim becomes a pending member. A second pending request is ignored. Reject deletes the pending row.
- Unpublished listings have no public claim or join action, are absent from public search, and are visible to the owner and to admins.
- Create requires name, settlement, category, and type. Creator is the owner. `published` is false. Publishing does not set `verified`.
- Members may edit name, category, type, settlement, URL, phone, address, notes, languages, hours, delivery hours, and photos. Members cannot delete, remove people, change the owner, publish, or set verified.
- Only the owner may delete from the profile. Admin delete still works. Members and pending rows go with the listing.
- Existing `claimed` values copy to `verified`. Existing listings stay `published`.
- Admin header button shows the count of unpublished listings plus pending members. Queue UI is a section in the existing admin panel.
- Hungarian UI. No email.
- Do not commit unless the user explicitly asks. Skip every Commit step unless they have asked.

---

### Task 1: Split verified from ownership

**Files:**
- Create: `backend/migrations/entry_verified_published.sql`
- Modify: `backend/internal/handlers/admin_entries.go` (migrate function next to the other entry migrates)
- Modify: `backend/internal/models/models.go`
- Modify: `backend/internal/handlers/public.go`
- Modify: `src/lib/components/EntryProfile.svelte`
- Test: `backend/handlers_test.go`

**Interfaces:**
- Consumes: `entries.claimed`
- Produces:
  - `entries.verified BOOLEAN NOT NULL DEFAULT false`
  - `entries.published BOOLEAN NOT NULL DEFAULT true`
  - Public entry JSON keeps `claimed` as “has an active owner” (false until Task 2) and adds `verified`
  - `EntryProfile` shows **Ellenőrzött** or **Nem ellenőrzött** from `verified`. It shows **Foglalt** only when `claimed` is true. Do not treat `claimed` as the quality check anymore.

- [ ] **Step 1: Write the failing test**

In `backend/handlers_test.go`, extend the existing claimed-entry test (or add `TestPublicEntryVerifiedSeparateFromClaimed`): create an entry with the admin API field that used to be `claimed: true`. After this task the admin field is `verified: true`. `GET /api/entry?slug=` returns `verified: true` and `claimed: false`.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 120s -run TestPublicEntryVerifiedSeparateFromClaimed`

Expected: FAIL, `verified` missing from JSON or `claimed` still mirrors the old column

- [ ] **Step 3: Write minimal implementation**

SQL migration, also run from a `MigrateEntryVerified()` called in `main.go` beside `MigrateEntryAvailability`:

```sql
ALTER TABLE entries ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE entries ADD COLUMN IF NOT EXISTS published BOOLEAN NOT NULL DEFAULT true;
UPDATE entries SET verified = claimed WHERE verified = false AND claimed = true;
```

Stop writing the quality check to `claimed`. Admin create/update accepts `verified`. Public SELECT computes `claimed` as false until `entry_members` exists (Task 2 replaces the expression). Public list and detail queries add `AND e.published = true`.

Update `Entry` and admin entry models: `Verified bool \`json:"verified"\``. Keep `Claimed` as the computed ownership flag.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestPublicEntryVerifiedSeparateFromClaimed|TestAdminEntriesCRUD'`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 2: Members, claim, and create

**Files:**
- Create: `backend/internal/account/listings.go`
- Modify: `backend/main.go`
- Modify: `backend/handlers_test.go`
- Modify: `backend/internal/handlers/public.go` (`claimed` subquery)

**Interfaces:**
- Consumes: `auth.UserFromRequest`, `entries.published`, `entries.verified`
- Produces:
  - table `entry_members(entry_id INT NOT NULL REFERENCES entries(id) ON DELETE CASCADE, user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, role VARCHAR(16) NOT NULL, status VARCHAR(16) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (entry_id, user_id))`
  - partial unique index: one `role = 'owner' AND status = 'active'` per `entry_id`
  - `POST /api/account/listings/claim` body `{ "entry_id" }`
  - `POST /api/account/listings` body `{ name, location_id, category_id, type_id, url, phone, address, notes, languages, hours, delivery_hours }`
  - `GET /api/account/listings` → `{ owned: [], member: [], pending: [], unpublished: [] }`
  - Public `claimed` is `EXISTS (SELECT 1 FROM entry_members m WHERE m.entry_id = e.id AND m.role = 'owner' AND m.status = 'active')`

Claim rules inside one transaction:

- 404 if the entry is missing or `published` is false
- if no active owner, insert this user as `owner` / `active`
- if this user is already the owner or an active member, return 200 with no new row
- if a pending row exists for this user, return 200 with that row
- otherwise insert `member` / `pending`

Create rules: require name, `location_id`, `category_id`, `type_id`. Insert `published = false`, `verified = false`, slug the same way admin entries do. Insert the creator as `owner` / `active`.

- [ ] **Step 1: Write the failing test**

`TestClaimFreeListingBecomesOwner` in `handlers_test.go`:

1. Admin creates a published entry.
2. `mustLogin("owner@test.lamsza")` POSTs `/api/account/listings/claim`.
3. Response role is `owner`, status is `active`.
4. Public `GET /api/entry` has `claimed: true`.
5. A second user POSTs claim. Response role is `member`, status is `pending`.
6. Public `claimed` stays true.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 120s -run TestClaimFreeListingBecomesOwner`

Expected: FAIL, route missing

- [ ] **Step 3: Write minimal implementation**

Implement the transaction with `SELECT ... FOR UPDATE` on the entry row so two concurrent claims cannot both insert an owner. The loser takes the pending-member branch.

Register routes. Create the table in `account.Migrate`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test . -count=1 -timeout 120s -run TestClaimFreeListingBecomesOwner`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 3: Edit, kick, and delete

**Files:**
- Modify: `backend/internal/account/listings.go`
- Modify: `backend/handlers_test.go`

**Interfaces:**
- Consumes: `entry_members`, admin entry update field set
- Produces:
  - `PATCH /api/account/listings?id=` updates content fields for an active owner or active member
  - `DELETE /api/account/listings?id=` owner only, deletes the entry
  - `DELETE /api/account/listings/members?entry_id=&user_id=` owner only, deletes that member row. Refuses when the target is the owner.

- [ ] **Step 1: Write the failing test**

`TestMemberCannotDeleteListing`: member session `DELETE /api/account/listings?id=` returns 403 and the public entry still exists. Owner `DELETE /api/account/listings/members` removes the member. Owner `DELETE /api/account/listings?id=` returns 200 and public GET is 404.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 120s -run TestMemberCannotDeleteListing`

Expected: FAIL, handlers missing

- [ ] **Step 3: Write minimal implementation**

`PATCH` allows only the content columns listed in the global constraints. It does not write `published`, `verified`, or `entry_members`. Non-member → 403. `DELETE` listing uses `DELETE FROM entries WHERE id = $1` (members cascade). Kicking the owner → 403.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestMemberCannotDeleteListing|TestClaimFreeListingBecomesOwner'`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 4: Admin queue

**Files:**
- Create: `backend/internal/account/queue.go`
- Modify: `backend/main.go` (wrap with `auth.RequireAdmin`)
- Modify: `backend/handlers_test.go`
- Modify: `backend/internal/auth/auth.go` (`HandleMe` adds `admin_queue_count` for admins)
- Modify: `src/routes/(public)/+layout.svelte` (badge on the admin button)
- Modify: `src/routes/admin/+page.svelte` (queue section)

**Interfaces:**
- Consumes: unpublished entries, pending `entry_members`
- Produces:
  - `GET /api/admin/listing-queue` → `{ unpublished: [{id,name,slug,owner_email}], members: [{entry_id,entry_name,user_id,email}] }`
  - `POST /api/admin/listing-queue/publish` body `{ "entry_id" }` sets `published = true` and does not change `verified`
  - `POST /api/admin/listing-queue/member` body `{ "entry_id", "user_id", "action": "approve"|"reject" }` — approve sets `status = active`; reject deletes the pending row
  - `HandleMe` field `admin_queue_count` (number) for admins, omitted otherwise
  - Header admin button shows the count when it is greater than 0
  - Admin section title **Bejegyzés-jóváhagyások**

- [ ] **Step 1: Write the failing test**

`TestAdminPublishDoesNotVerify`: user creates a listing (`published` false). `GET /api/admin/listing-queue` includes it. Admin publish. Public GET returns the entry with `verified: false` and `claimed: true`. `admin_queue_count` on admin `GET /api/auth/me` drops by one.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 120s -run TestAdminPublishDoesNotVerify`

Expected: FAIL, route missing

- [ ] **Step 3: Write minimal implementation**

Count query:

```sql
SELECT
  (SELECT COUNT(*) FROM entries WHERE published = false) +
  (SELECT COUNT(*) FROM entry_members WHERE status = 'pending')
```

Publish and member actions are admin-only. Reject of a missing pending row returns 404. The admin page section lists both arrays with **Közzététel**, **Elfogadás**, and **Elutasítás** buttons.

Header: next to the admin icon, show the count from `$auth.adminQueueCount` when `> 0`. Map `admin_queue_count` in `meToAuthState`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestAdminPublishDoesNotVerify|TestClaimFreeListingBecomesOwner|TestMemberCannotDeleteListing'`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.

---

### Task 5: Bejegyzéseim tab

**Files:**
- Modify: `src/lib/accountPrefs.js`
- Modify: `src/routes/(public)/profil/+page.svelte`
- Modify: `src/routes/(public)/bejegyzes/[slug]/+page.svelte`
- Modify: `tests/accountPrefs.test.js`

**Interfaces:**
- Consumes: listing routes from Tasks 2–3, public entry `claimed`
- Produces: `profileTabIds` includes `"bejegyzeseim"` after `"beallitasok"`

- [ ] **Step 1: Write the failing test**

Set the expected `profileTabIds` to:

```js
["profil", "beallitasok", "bejegyzeseim", "linkjeim", "elozmenyek", "kedvencek"]
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test tests/accountPrefs.test.js`

Expected: FAIL, array mismatch

- [ ] **Step 3: Write minimal implementation**

**Bejegyzéseim** loads `GET /api/account/listings` and shows four groups: **Saját**, **Tagság**, **Jóváhagyásra vár**, **Közzétételre vár**. Create form fields: name, settlement, category, type, and the optional content fields. Submit POSTs `/api/account/listings`. Owner rows have **Törlés** and, for active members, **Eltávolítás**. Content edit uses `PATCH`. Failures keep the previous list and show **A mentés nem sikerült**.

On the public listing page, when logged in and `published`: if `claimed` is false, button **Sajátnak jelölöm** POSTs claim. If `claimed` is true and the user is not owner or member, button **Tagság kérése** POSTs claim. Hide both when the listing is unpublished (it will not be on the public page).

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test tests/accountPrefs.test.js`

Run: `cd backend && go test . -count=1 -timeout 120s -run 'TestClaimFreeListingBecomesOwner|TestMemberCannotDeleteListing|TestAdminPublishDoesNotVerify|TestPublicEntryVerifiedSeparateFromClaimed'`

Expected: PASS

- [ ] **Step 5: Commit**

Skip unless the user asked to commit.
