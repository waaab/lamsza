# Listing claim and membership Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Sajátnak jelölöm asks an admin to approve one owner, and Tagság kérése can be accepted or denied by the owner or by an admin.

**Architecture:** A gazdátlan claim inserts `entry_members` with `role = 'owner'` and `status = 'pending'`. The listing stays Gazdátlan until an admin sets that row to `active`. A second claim while that row exists returns 409. Membership stays `role = 'member'` and `status = 'pending'`, and both the active owner and an admin may set it to `active` or delete it. Creating a listing still inserts an active owner immediately.

**Tech Stack:** Go `net/http` + Postgres, existing Svelte listing page and **Bejegyzéseim**, Go `go test`.

**Spec:** `docs/superpowers/specs/2026-09-27-listing-profile-design.md`

**Depends on:** `docs/superpowers/plans/2026-09-27-listing-profile-view.md` (sidebar buttons). Suggestion queue may land in parallel; this plan only adds the claim count beside it.

## Global Constraints

- **Sajátnak jelölöm** has no note. The admin preview shows the user and the listing, then **Elfogad** or **Elutasít**.
- Accept makes that user the active owner and the listing becomes Átvéve. Deny deletes the pending owner row. The listing stays Gazdátlan until accept.
- Only one claim can be open. While it is open, every eligible logged-in user sees disabled **Átvételre vár**. The sender cannot withdraw it. A second POST returns **409** `{"error":"claim_pending"}`.
- **Tagság kérése** has no note. Several may be open. The sender sees disabled **Tagságkérés elküldve**. Other eligible users still see **Tagság kérése**.
- The owner accepts or denies from **Bejegyzéseim**. An admin accepts or denies from the existing members queue.
- An accepted member edits the same fields as the owner, including `hours_enabled` and `delivery_enabled`. Only the owner can delete the listing. Admins keep their existing tools.
- A pending member may still send **Javaslat módosításra** when no suggestion is open.
- A listing created by a user is Átvéve at once. Do not route that create through the claim queue.
- Hungarian: **Sajátnak jelölöm**, **Átvételre vár**, **Tagság kérése**, **Tagságkérés elküldve**, **Elfogad**, **Elutasít**.
- `admin_queue_count` includes open claims (pending owner rows) and open suggestions. Pending members are already counted.
- Commit only the files named in the task. Never `git add -A`. Do not merge to `main`.

---

### Task 1: Claim waits for an admin

**Files:**
- Modify: `backend/internal/account/listings.go` (`HandleClaimListing`)
- Modify: `backend/internal/account/queue.go`
- Modify: `backend/internal/auth/auth.go`
- Test: `backend/entry_claim_request_test.go`

**Interfaces:**
- `POST /api/account/listings/claim` with `{ "entry_id" }` on a gazdátlan listing inserts `role = 'owner', status = 'pending'` and returns `{ "entry_id", "role": "owner", "status": "pending" }`.
- The same POST while a pending owner exists returns 409 `{"error":"claim_pending"}`.
- `claimed` on public `GET /api/entry` stays false until an active owner exists. The existing `EXISTS (... role = 'owner' AND status = 'active')` query stays.
- `GET /api/entry` adds `claim_pending: true` while a pending owner row exists.
- `GET /api/admin/listing-queue` adds `claims`: `{ "entry_id", "entry_name", "user_id", "email" }`.
- `POST /api/admin/listing-queue/claim` body `{ "entry_id", "user_id", "action": "accept" | "deny" }`.
- The members query in `HandleListingQueue` gains `AND m.role = 'member'` so a pending owner is not listed as a membership request.
- `QueueCount` and the count in `auth.go` add pending owner rows. Do not count those rows twice through the unfiltered pending-member count: the member count becomes `status = 'pending' AND role = 'member'`, plus a separate `role = 'owner' AND status = 'pending'`.

- [ ] **Step 1: Write the failing tests**

`backend/entry_claim_request_test.go` (`package main`):

```go
func TestClaimStaysGazdátlanUntilAdminAccepts(t *testing.T) {
	id, slug := createEntry("Claim Wait", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]any{
		"entry_id": id,
	}, userCookie)
	if rr.Code != 200 {
		t.Fatalf("claim status %d %s", rr.Code, rr.Body.String())
	}
	var membership map[string]any
	json.Unmarshal(rr.Body.Bytes(), &membership)
	if membership["role"] != "owner" || membership["status"] != "pending" {
		t.Fatalf("membership: %#v", membership)
	}

	pub := doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	var entry map[string]any
	json.Unmarshal(pub.Body.Bytes(), &entry)
	if entry["claimed"] == true {
		t.Fatal("listing became Átvéve before accept")
	}
	if entry["claim_pending"] != true {
		t.Fatal("claim_pending missing")
	}

	again := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]any{
		"entry_id": id,
	}, otherCookie)
	if again.Code != 409 {
		t.Fatalf("second claim: %d %s", again.Code, again.Body.String())
	}
}
```

`userCookie` and `otherCookie` are the two non-admin cookies already built in `handlers_test.go` and `website_test.go`. Copy the second-user setup from `website_test.go` if this file cannot see that cookie. Add a second test in the same file: admin POST `action: "accept"` sets public `claimed` to true; a fresh listing's deny deletes the pending row and a later claim returns 200.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 180s -run TestClaimStays`

Expected: FAIL. Today's handler inserts `status = 'active'`, so `claimed` is already true.

- [ ] **Step 3: Change the claim insert**

In `HandleClaimListing`, when no active owner exists, insert `owner` / `pending` instead of `owner` / `active`. Before the insert, if a pending owner row exists, return 409 `claim_pending`. Keep the member insert as `member` / `pending` when an active owner exists.

Add `claim_pending` to the public entry SELECT in `public.go`:

```sql
EXISTS (
  SELECT 1 FROM entry_members m
  WHERE m.entry_id = e.id AND m.role = 'owner' AND m.status = 'pending'
)
```

Add `HandleListingQueueClaim`. Accept updates that pending owner row to `active`. Deny deletes it. Reject accept when another active owner already exists.

Filter the members queue with `AND m.role = 'member'`. Split the queue count as specified above. Register `/api/admin/listing-queue/claim` in `main.go` and `handlers_test.go`.

- [ ] **Step 4: Run the tests**

Run: `cd backend && go test . -count=1 -timeout 180s -run 'TestClaimStays|TestClaim|TestAdminQueue'`

Expected: PASS. Update an older test that expected claim to set `claimed` immediately so it accepts through the admin route first, or so it asserts the new pending status. Do not leave a test that requires the immediate claim.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/account/listings.go backend/internal/account/queue.go backend/internal/auth/auth.go backend/internal/handlers/public.go backend/main.go backend/handlers_test.go backend/entry_claim_request_test.go
git commit -m "Hold a listing claim until an admin accepts it."
```

---

### Task 2: Owner decides membership

**Files:**
- Modify: `backend/internal/account/listings.go`
- Modify: `backend/main.go`
- Modify: `backend/handlers_test.go`
- Test: `backend/entry_membership_decision_test.go`
- Modify: `src/routes/(public)/fiok/+page.svelte`
- Modify: `src/routes/(public)/bejegyzes/[slug]/+page.svelte`
- Modify: `src/lib/components/EntryProfile.svelte`
- Modify: `src/routes/admin/+page.svelte`

**Interfaces:**
- `GET /api/account/listings/members?entry_id=` already exists and returns an array of active members. Extend that array to include pending member rows, each with `status`. Do not change it to an object. `fiok/+page.svelte` already treats the body as an array.
- `POST /api/account/listings/members` body `{ "entry_id", "user_id", "action": "accept" | "deny" }` — active owner only. Accept sets that pending member to `active`. Deny deletes the pending member row. 403 if the target role is `owner`. The route is already registered in `main.go` and `handlers_test.go`. Add the POST case to `HandleListingMembers`. Keep the existing DELETE.
- Public page: `showClaim` is false while `claim_pending` is true; the sidebar shows disabled **Átvételre vár** for every eligible logged-in user. `showJoin` stays true for eligible users who have no membership row. A user whose own row is `member` / `pending` sees disabled **Tagságkérés elküldve** instead.
- **Bejegyzéseim** lists pending members for each owned listing with **Elfogad** and **Elutasít**.
- Admin members queue keeps **Elfogad** and **Elutasít** and now lists only `role = 'member'`. Add a claims block bound to `data.claims`.

- [ ] **Step 1: Write the failing owner-decision test**

```go
func TestOwnerAcceptsMember(t *testing.T) {
	owner := userCookie
	member := otherCookie
	rr := doRequestWithCookie(t, "POST", "/api/account/listings", map[string]any{
		"name": "Owned Place",
		"type_id": typeID, "category_id": categoryID, "location_id": mustLocID(t),
	}, owner)
	if rr.Code != 200 {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]any
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := created["id"]
	defer doRequestWithCookie(t, "DELETE", "/api/account/listings?id="+formatID(entryID), nil, owner)

	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]any{"entry_id": entryID}, member)
	if rr.Code != 200 {
		t.Fatalf("join: %d %s", rr.Code, rr.Body.String())
	}

	rr = doRequestWithCookie(t, "POST", "/api/account/listings/members", map[string]any{
		"entry_id": entryID, "user_id": memberUserID, "action": "accept",
	}, member)
	if rr.Code != 403 {
		t.Fatalf("member accept: %d", rr.Code)
	}

	rr = doRequestWithCookie(t, "POST", "/api/account/listings/members", map[string]any{
		"entry_id": entryID, "user_id": memberUserID, "action": "accept",
	}, owner)
	if rr.Code != 200 {
		t.Fatalf("owner accept: %d %s", rr.Code, rr.Body.String())
	}

	rr = doRequestWithCookie(t, "PATCH", "/api/account/listings?id="+formatID(entryID), map[string]any{
		"notes": "tag szerkesztette",
	}, member)
	if rr.Code != 200 {
		t.Fatalf("member patch: %d %s", rr.Code, rr.Body.String())
	}
}
```

Copy `typeID` and `categoryID` from the create body in `handlers_test.go` around the existing `POST /api/account/listings` test. `memberUserID` is the numeric id of `otherCookie`'s user. Assert a third user's `DELETE /api/account/listings?id=` returns 403.

- [ ] **Step 2: Run it to verify it fails**

Run: `cd backend && go test . -count=1 -timeout 180s -run TestOwnerAcceptsMember`

Expected: FAIL, POST returns 405 because `HandleListingMembers` has no POST branch.

- [ ] **Step 3: Implement the owner route and the labels**

Add the POST branch to the existing `HandleListingMembers`. Change `handleListListingMembers` so the SELECT returns `role = 'member'` rows with `status` in `('pending', 'active')`.

On the listing page, pass `claimPending` and `membershipStatus` into `EntryProfile`. Render the disabled waiting labels from the Global Constraints. Do not call claim again when the label is a waiting state.

On **Bejegyzéseim**, for each owned listing, GET the members endpoint and render the pending rows. **Elfogad** and **Elutasít** POST the owner route, then reload.

In `fetchListingQueue`, store `data.claims`. Render that block with the same accept/deny buttons, posting to `/api/admin/listing-queue/claim`.

Run Svelte MCP `svelte-autofixer` on `EntryProfile.svelte` until clean. The fiok and admin pages are legacy; fix only issues introduced by the new blocks.

- [ ] **Step 4: Run the tests and click through**

Run: `cd backend && go test . -count=1 -timeout 180s -run 'TestOwnerAcceptsMember|TestClaimStays'`

Expected: PASS

As a logged-in non-member, send **Sajátnak jelölöm** on a gazdátlan listing. The mark stays **Gazdátlan** and the button becomes **Átvételre vár** for a second account too. Accept as admin and confirm **Átvéve**. On that listing, send **Tagság kérése** from a third account and accept it from **Bejegyzéseim**.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/account/listings.go backend/main.go backend/handlers_test.go backend/entry_membership_decision_test.go 'src/routes/(public)/fiok/+page.svelte' 'src/routes/(public)/bejegyzes/[slug]/+page.svelte' src/lib/components/EntryProfile.svelte src/routes/admin/+page.svelte
git commit -m "Let the owner and an admin decide listing membership."
```
