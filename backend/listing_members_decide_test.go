package main

import (
	"backend/internal/db"
	"encoding/json"
	"net/http"
	"testing"
)

// The owner decides member requests (planned in
// docs/superpowers/plans/2026-09-27-listing-claim-membership.md, never
// written): only the active owner may accept or deny, only a pending member
// request can be decided, accept activates it and deny removes it.
func TestOwnerDecidesMemberRequests(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB")
	}
	entryID, _ := createEntry(t, "Decide Members Test Entry", locs[0]["id"])
	t.Cleanup(func() { deleteEntryFixture(t, entryID) })

	const ownerEmail, aEmail, bEmail, otherEmail = "decide-owner@test.lamsza", "decide-a@test.lamsza", "decide-b@test.lamsza", "decide-other@test.lamsza"
	owner := mustLogin(ownerEmail)
	claim := func(cookie *http.Cookie, who string) {
		t.Helper()
		if rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{"entry_id": entryID}, cookie); rr.Code != http.StatusOK {
			t.Fatalf("%s claim: %d %s", who, rr.Code, rr.Body.String())
		}
	}
	claim(owner, "owner")
	acceptPendingClaim(t, entryID, ownerEmail)
	memberA, memberB := mustLogin(aEmail), mustLogin(bEmail)
	mustLogin(otherEmail)
	claim(memberA, "member A")
	claim(memberB, "member B")

	userID := func(email string) int {
		t.Helper()
		var id int
		if err := db.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&id); err != nil {
			t.Fatalf("user %s: %v", email, err)
		}
		return id
	}
	status := func(uid int) string {
		t.Helper()
		var s string
		if err := db.DB.QueryRow(`SELECT status FROM entry_members WHERE entry_id = $1 AND user_id = $2`, entryID, uid).Scan(&s); err != nil {
			return "none"
		}
		return s
	}
	decide := func(cookie *http.Cookie, uid int, action string) int {
		t.Helper()
		return doRequestWithCookie(t, "POST", "/api/account/listings/members",
			map[string]interface{}{"entry_id": entryID, "user_id": uid, "action": action}, cookie).Code
	}
	a, b := userID(aEmail), userID(bEmail)

	if got := decide(memberA, b, "accept"); got != http.StatusForbidden {
		t.Errorf("a pending member accepting another: %d, want 403", got)
	}
	if got := decide(owner, a, "maybe"); got != http.StatusBadRequest {
		t.Errorf("an unknown action: %d, want 400", got)
	}
	if got := decide(owner, userID(otherEmail), "accept"); got != http.StatusNotFound {
		t.Errorf("deciding for someone who never asked: %d, want 404", got)
	}
	if got := decide(owner, userID(ownerEmail), "deny"); got != http.StatusForbidden {
		t.Errorf("deciding on the owner row: %d, want 403", got)
	}

	if got := decide(owner, a, "accept"); got != http.StatusOK {
		t.Fatalf("owner accepts A: %d, want 200", got)
	}
	if s := status(a); s != "active" {
		t.Errorf("A after accept: %s, want active", s)
	}
	if got := decide(owner, a, "deny"); got != http.StatusForbidden {
		t.Errorf("deciding again on an active member: %d, want 403", got)
	}

	if got := decide(owner, b, "deny"); got != http.StatusOK {
		t.Fatalf("owner denies B: %d, want 200", got)
	}
	if s := status(b); s != "none" {
		t.Errorf("B after deny: %s, want no row", s)
	}
}
