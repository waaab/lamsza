package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func mustLocID(t *testing.T) interface{} {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test")
	}
	return locs[0]["id"]
}

func createEntry(t *testing.T, name string, locID interface{}) (interface{}, string) {
	payload := map[string]interface{}{
		"name":        name,
		"location_id": locID,
		"type":        "entry",
		"category":    "Egyéb",
	}
	rr := doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	slug, _ := created["slug"].(string)
	if slug == "" {
		t.Fatalf("POST entries missing slug, got: %v", created)
	}
	return created["id"], slug
}

func TestUnclaimedPublicEntryHidesExtras(t *testing.T) {
	payload := map[string]interface{}{
		"name":        "UnclaimedExtras A",
		"location_id": mustLocID(t),
		"type":        "entry",
		"category":    "Egyéb",
		"hours":       map[string]any{"mon": map[string]any{"open": "09:00", "close": "17:00", "closed": false}},
	}
	rr := doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	slug, _ := created["slug"].(string)
	if slug == "" {
		t.Fatal("POST entries missing slug")
	}
	id := created["id"]
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s (slug=%s, id=%v)", rr.Code, rr.Body.String(), slug, id)
	}
	var got map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["claimed"] == true {
		t.Fatal("expected unclaimed")
	}
	if _, ok := got["rating"]; ok {
		t.Fatal("unclaimed must omit rating")
	}
	if _, ok := got["reviews"]; ok {
		t.Fatal("unclaimed must omit reviews")
	}
	if got["ratings_enabled"] != false {
		t.Fatalf("ratings_enabled: %v", got["ratings_enabled"])
	}
	photos, _ := got["photos"].([]interface{})
	if len(photos) != 0 {
		t.Fatalf("unclaimed photos must be empty, got %v", got["photos"])
	}
}
