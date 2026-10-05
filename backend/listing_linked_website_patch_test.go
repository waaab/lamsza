package main

import (
	"backend/internal/db"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPatchListingRatingsWithLinkedWebsite(t *testing.T) {
	const domain = "patch-linked-ratings.example"
	cleanup := func() {
		if _, err := db.DB.Exec(`DELETE FROM entries WHERE id IN (SELECT entry_id FROM websites WHERE domain_key = $1)`, domain); err != nil {
			t.Errorf("cleanup entries: %v", err)
		}
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, domain); err != nil {
			t.Errorf("cleanup website: %v", err)
		}
	}
	cleanup()
	defer cleanup()

	owner := mustLogin("patch-linked-ratings@test.lamsza")
	locID := mustLocID(t)
	catID := testLeafCategoryID
	var typeID int
	if err := db.DB.QueryRow(`SELECT id FROM entry_types ORDER BY id ASC LIMIT 1`).Scan(&typeID); err != nil {
		t.Fatal(err)
	}

	rr := doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"name":        "Linked Ratings Biz",
		"location_id": locID,
		"category_id": catID,
		"type_id":     typeID,
		"url":         "https://" + domain,
	}, owner)
	if rr.Code != http.StatusOK {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := int(created["id"].(float64))

	var linked int
	if err := db.DB.QueryRow(`SELECT entry_id FROM websites WHERE domain_key=$1`, domain).Scan(&linked); err != nil || linked != entryID {
		t.Fatalf("website link: err=%v entry=%d want=%d", err, linked, entryID)
	}

	patch := map[string]interface{}{
		"name":             "Linked Ratings Biz",
		"location_id":      locID,
		"category_id":      catID,
		"category_ids":     []int{catID},
		"type_id":          typeID,
		"url":              "https://" + domain,
		"phone":            "0700000000",
		"address":          "Teszt utca",
		"notes":            "",
		"languages":        []string{"HU", "RO", "EN"},
		"hours":            map[string]interface{}{},
		"hours_enabled":    false,
		"delivery_hours":   map[string]interface{}{},
		"delivery_enabled": false,
		"photos":           []interface{}{},
		"ratings_enabled":  true,
		"tags":             []string{},
		"social_links":     []interface{}{},
	}
	rr = doRequestWithCookie(t, "PATCH", "/api/account/listings?id="+formatID(entryID), patch, owner)
	if rr.Code != http.StatusOK {
		t.Fatalf("PATCH ratings with linked website: %d %s", rr.Code, rr.Body.String())
	}
}
