package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"backend/internal/db"
)

func TestOwnerPatchUpdatesTags(t *testing.T) {
	entryID, slug := createEntry(t, "Owner Tags", mustLocID(t))
	defer deleteEntryFixture(t, entryID)

	var typeID, categoryID, locationID int
	err := dbQueryIDs(t, entryID, &typeID, &categoryID, &locationID)
	if err != nil {
		t.Fatal(err)
	}

	owner := mustLogin("tags-owner@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]any{
		"entry_id": entryID,
	}, owner)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: %d %s", rr.Code, rr.Body.String())
	}
	acceptPendingClaim(t, entryID, "tags-owner@test.lamsza")

	rr = doRequestWithCookie(t, "PATCH", "/api/account/listings?id="+formatID(entryID), map[string]any{
		"name":        "Owner Tags",
		"location_id": locationID,
		"category_id": categoryID,
		"type_id":     typeID,
		"tags":        []string{"#uj-cimke", "masik"},
	}, owner)
	if rr.Code != http.StatusOK {
		t.Fatalf("patch: %d %s", rr.Code, rr.Body.String())
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("public: %d %s", rr.Code, rr.Body.String())
	}
	var pub map[string]any
	json.Unmarshal(rr.Body.Bytes(), &pub)
	got := map[string]bool{}
	for _, tag := range pub["tags"].([]any) {
		got[tag.(string)] = true
	}
	if !got["uj-cimke"] || !got["masik"] {
		t.Fatalf("tags: %#v", pub["tags"])
	}
}

func dbQueryIDs(t *testing.T, entryID any, typeID, categoryID, locationID *int) error {
	t.Helper()
	return db.DB.QueryRow(`
		SELECT type_id, category_id, location_id FROM entries WHERE id = $1
	`, entryID).Scan(typeID, categoryID, locationID)
}
