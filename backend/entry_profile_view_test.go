package main

import (
	"encoding/json"
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
}
