package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"backend/internal/db"
)

func TestSuggestionSignedOut(t *testing.T) {
	rr := doAnonRequest(t, "POST", "/api/entry/suggestions", map[string]any{
		"entry_id": 1,
		"fields":   map[string]any{"phone": "0700000000"},
	})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("signed-out POST: expected 401, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestSuggestionOwnerForbidden(t *testing.T) {
	entryID, _ := createEntry(t, "Suggestion Owner", mustLocID(t))
	defer deleteEntryFixture(t, entryID)

	owner := mustLogin("suggestion-owner@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]any{
		"entry_id": entryID,
	}, owner)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	acceptPendingClaim(t, entryID, "suggestion-owner@test.lamsza")

	rr = doRequestWithCookie(t, "POST", "/api/entry/suggestions", map[string]any{
		"entry_id": entryID,
		"fields":   map[string]any{"phone": "0700000001"},
	}, owner)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("owner POST: expected 403, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestSuggestionOnlyOneOpenPerEntry(t *testing.T) {
	entryID, slug := createEntry(t, "Suggestion Accept", mustLocID(t))
	defer deleteEntryFixture(t, entryID)
	id := entryID

	first := mustLogin("suggestion-first@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/entry/suggestions", map[string]any{
		"entry_id": entryID,
		"fields":   map[string]any{"phone": "0700111222", "name": "Suggestion Accept Renamed"},
		"note":     "rossz a telefon",
	}, first)
	if rr.Code != http.StatusOK {
		t.Fatalf("first POST: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]any
	json.Unmarshal(rr.Body.Bytes(), &created)
	if created["status"] != "open" {
		t.Fatalf("status: %v", created["status"])
	}

	second := mustLogin("suggestion-second@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/entry/suggestions", map[string]any{
		"entry_id": entryID,
		"fields":   map[string]any{"phone": "0700333444"},
	}, second)
	if rr.Code != http.StatusConflict {
		t.Fatalf("second POST: expected 409, got %d; body: %s", rr.Code, rr.Body.String())
	}

	// Accepting a suggestion is an admin action and now lives in lamsza-admin
	// (BOG-42). What this app still owns is the open-suggestion gate: exactly
	// one open suggestion per entry, and the stored row keeps the proposal
	// without touching the entry.
	var changes, status string
	if err := db.DB.QueryRow(`SELECT changes::text, status FROM entry_suggestions WHERE entry_id = $1`, id).Scan(&changes, &status); err != nil {
		t.Fatalf("stored suggestion: %v", err)
	}
	if status != "open" {
		t.Fatalf("stored status: %q", status)
	}
	if !strings.Contains(changes, "0700111222") {
		t.Fatalf("stored changes lost the proposal: %s", changes)
	}

	var phone, storedSlug string
	if err := db.DB.QueryRow(`SELECT COALESCE(phone, ''), slug FROM entries WHERE id = $1`, id).Scan(&phone, &storedSlug); err != nil {
		t.Fatalf("entry: %v", err)
	}
	if phone == "0700111222" {
		t.Fatal("an open suggestion must not write the entry")
	}
	if storedSlug != slug {
		t.Fatalf("slug changed from %q to %q", slug, storedSlug)
	}
}

func TestSuggestionDenyAllowsNewPost(t *testing.T) {
	entryID, _ := createEntry(t, "Suggestion Deny", mustLocID(t))
	defer deleteEntryFixture(t, entryID)
	id := entryID

	user := mustLogin("suggestion-deny@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/entry/suggestions", map[string]any{
		"entry_id": entryID,
		"fields":   map[string]any{"phone": "0700555666"},
	}, user)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var suggestionID int
	if err := db.DB.QueryRow(`SELECT id FROM entry_suggestions WHERE entry_id = $1 AND status = 'open'`, id).Scan(&suggestionID); err != nil {
		t.Fatalf("open suggestion: %v", err)
	}
	// The admin decision itself moved to lamsza-admin; this closes the row the
	// way a denial does, so the gate below is still covered here.
	if _, err := db.DB.Exec(`UPDATE entry_suggestions SET status = 'denied' WHERE id = $1`, suggestionID); err != nil {
		t.Fatalf("deny suggestion: %v", err)
	}

	var phone string
	if err := db.DB.QueryRow(`SELECT COALESCE(phone, '') FROM entries WHERE id = $1`, id).Scan(&phone); err != nil {
		t.Fatalf("phone: %v", err)
	}
	if phone == "0700555666" {
		t.Fatal("deny wrote the phone")
	}

	rr = doRequestWithCookie(t, "POST", "/api/entry/suggestions", map[string]any{
		"entry_id": entryID,
		"fields":   map[string]any{"address": "Új utca 1"},
	}, user)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST after deny: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
}
