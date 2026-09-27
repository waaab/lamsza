package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"backend/internal/db"
)

func TestGazdátlanVerifiedEntryKeepsShortPublicFields(t *testing.T) {
	id, slug := createEntry(t, "Short View A", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)
	_, err := db.DB.Exec(`
		UPDATE entries
		SET verified = true,
		    phone = '0700111222',
		    social_links = '[{"label":"Facebook","url":"https://facebook.com/a"}]'::jsonb,
		    photos = '["https://example.com/a.jpg"]'::jsonb,
		    hours_enabled = false,
		    hours = '{"mon":{"open":"09:00","close":"17:00","closed":false}}'::jsonb
		WHERE id = $1
	`, id)
	if err != nil {
		t.Fatal(err)
	}
	rr := doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != 200 {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	var got map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["phone"] != "" {
		t.Fatalf("phone visible on gazdátlan listing: %v", got["phone"])
	}
	photos, _ := got["photos"].([]interface{})
	if len(photos) != 0 {
		t.Fatalf("photos visible before Ellenőrzött: %v", got["photos"])
	}
	socialLinks, _ := got["social_links"].([]interface{})
	if len(socialLinks) != 0 {
		t.Fatalf("social_links visible on gazdátlan listing: %v", got["social_links"])
	}
	hours, _ := got["hours"].(map[string]interface{})
	if len(hours) != 0 {
		t.Fatalf("hours visible when hours_enabled false: %v", got["hours"])
	}
	if got["verified"] != true || got["claimed"] == true {
		t.Fatalf("marks: verified %v claimed %v", got["verified"], got["claimed"])
	}
}

func TestClaimedUnverifiedEntryShowsPhoneAndHoursEnabled(t *testing.T) {
	id, slug := createEntry(t, "Short View B", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	var ownerUserID int
	err := db.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, "owner@test.lamsza").Scan(&ownerUserID)
	if err != nil {
		t.Fatal(err)
	}
	entryIDInt := int(id.(float64))
	_, err = db.DB.Exec(`
		INSERT INTO entry_members (entry_id, user_id, role, status)
		VALUES ($1, $2, 'owner', 'active')
	`, entryIDInt, ownerUserID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.DB.Exec(`
		UPDATE entries
		SET verified = false,
		    phone = '0700333444',
		    hours_enabled = true,
		    hours = '{}'::jsonb
		WHERE id = $1
	`, entryIDInt)
	if err != nil {
		t.Fatal(err)
	}

	rr := doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != 200 {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	var got map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["phone"] != "0700333444" {
		t.Fatalf("phone: expected stored number, got %v", got["phone"])
	}
	if got["hours_enabled"] != true {
		t.Fatalf("hours_enabled: expected true, got %v", got["hours_enabled"])
	}
	photos, _ := got["photos"].([]interface{})
	if len(photos) != 0 {
		t.Fatalf("photos visible before Ellenőrzött: %v", got["photos"])
	}
}

func TestAdminEntriesStoreHoursAndDeliverySwitches(t *testing.T) {
	locID := mustLocID(t)
	id, _ := createEntry(t, "Admin Switch Entry", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	payload := map[string]interface{}{
		"id":               id,
		"name":             "Admin Switch Entry",
		"location_id":      locID,
		"type":             "entry",
		"category":         "Egyéb",
		"hours_enabled":    true,
		"delivery_enabled": true,
	}
	rr := doRequest(t, "PUT", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, "GET", "/api/admin/entries", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entries: expected 200, got %d", rr.Code)
	}
	var entries []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &entries)
	var found map[string]interface{}
	for _, entry := range entries {
		if entry["id"] == id {
			found = entry
			break
		}
	}
	if found == nil {
		t.Fatalf("entry %v not found in admin list", id)
	}
	if found["hours_enabled"] != true {
		t.Fatalf("hours_enabled: expected true on gazdátlan listing, got %v", found["hours_enabled"])
	}
	if found["delivery_enabled"] != false {
		t.Fatalf("delivery_enabled: expected false for non-Vendéglő, got %v", found["delivery_enabled"])
	}

	rr = doRequest(t, "GET", "/api/admin/entry_categories", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry_categories: expected 200, got %d", rr.Code)
	}
	var categories []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &categories)
	var vendegloID interface{}
	for _, cat := range categories {
		if cat["name"] == "Vendéglő" {
			vendegloID = cat["id"]
			break
		}
	}
	if vendegloID == nil {
		t.Skip("Vendéglő category missing; cannot test delivery_enabled")
	}

	payload["category_id"] = vendegloID
	payload["category"] = "Vendéglő"
	payload["delivery_enabled"] = true
	rr = doRequest(t, "PUT", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT vendéglő entry: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, "GET", "/api/admin/entries", nil)
	json.Unmarshal(rr.Body.Bytes(), &entries)
	found = nil
	for _, entry := range entries {
		if entry["id"] == id {
			found = entry
			break
		}
	}
	if found == nil {
		t.Fatalf("entry %v not found after vendéglő update", id)
	}
	if found["delivery_enabled"] != true {
		t.Fatalf("delivery_enabled: expected true for Vendéglő, got %v", found["delivery_enabled"])
	}
}

func TestPublicHoursSwitch(t *testing.T) {
	locID := mustLocID(t)
	id, slug := createEntry(t, "Public Hours Switch", locID)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	entryIDInt := int(id.(float64))
	_, err := db.DB.Exec(`
		UPDATE entries
		SET hours = '{"mon":{"open":"09:00","close":"17:00","closed":false}}'::jsonb,
		    hours_enabled = false
		WHERE id = $1
	`, entryIDInt)
	if err != nil {
		t.Fatal(err)
	}

	payload := map[string]interface{}{
		"id":            id,
		"name":          "Public Hours Switch",
		"location_id":   locID,
		"type":          "entry",
		"category":      "Egyéb",
		"hours_enabled": true,
		"hours":         map[string]interface{}{},
	}
	rr := doRequest(t, "PUT", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT entries hours on: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry hours on: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var got map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["hours_enabled"] != true {
		t.Fatalf("hours_enabled: expected true, got %v", got["hours_enabled"])
	}
	hours, ok := got["hours"].(map[string]interface{})
	if !ok {
		t.Fatalf("hours: expected JSON object, got %T (%v)", got["hours"], got["hours"])
	}
	if len(hours) != 0 {
		t.Fatalf("hours: expected empty object when switch on, got %v", got["hours"])
	}

	payload["hours_enabled"] = false
	rr = doRequest(t, "PUT", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT entries hours off: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry hours off: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["hours_enabled"] != false {
		t.Fatalf("hours_enabled: expected false, got %v", got["hours_enabled"])
	}
	hours, ok = got["hours"].(map[string]interface{})
	if !ok {
		t.Fatalf("hours: expected JSON object, got %T (%v)", got["hours"], got["hours"])
	}
	if len(hours) != 0 {
		t.Fatalf("hours: expected empty object when switch off, got %v", got["hours"])
	}
}
