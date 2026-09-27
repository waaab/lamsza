package main

import (
	"encoding/json"
	"net/http"
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
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

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

func TestSuggestionOneOpenAndAcceptKeepsSlug(t *testing.T) {
	entryID, slug := createEntry(t, "Suggestion Accept", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)
	id := int(entryID.(float64))

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

	var suggestionID int
	if err := db.DB.QueryRow(`SELECT id FROM entry_suggestions WHERE entry_id = $1 AND status = 'open'`, id).Scan(&suggestionID); err != nil {
		t.Fatalf("open suggestion: %v", err)
	}
	rr = doRequestWithCookie(t, "POST", "/api/admin/listing-queue/suggestion", map[string]any{
		"id":     suggestionID,
		"action": "accept",
	}, testAdminCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("accept: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var phone, name, storedSlug string
	if err := db.DB.QueryRow(`SELECT phone, name, slug FROM entries WHERE id = $1`, id).Scan(&phone, &name, &storedSlug); err != nil {
		t.Fatalf("entry: %v", err)
	}
	if phone != "0700111222" {
		t.Fatalf("phone: %q", phone)
	}
	if name != "Suggestion Accept Renamed" {
		t.Fatalf("name: %q", name)
	}
	if storedSlug != slug {
		t.Fatalf("slug changed from %q to %q", slug, storedSlug)
	}
}

func TestSuggestionDenyAllowsNewPost(t *testing.T) {
	entryID, _ := createEntry(t, "Suggestion Deny", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)
	id := int(entryID.(float64))

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
	rr = doRequestWithCookie(t, "POST", "/api/admin/listing-queue/suggestion", map[string]any{
		"id":     suggestionID,
		"action": "deny",
	}, testAdminCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("deny: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
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
