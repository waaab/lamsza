package main

import (
	"backend/internal/account"
	"backend/internal/db"
	"testing"
)

func TestMigrateWebsitesBackfillsListingURL(t *testing.T) {
	account.MigrateWebsites()
	var banned bool
	if err := db.DB.QueryRow(`SELECT website_banned FROM users LIMIT 1`).Scan(&banned); err != nil {
		t.Fatal(err)
	}
	id, _ := createEntry(t, "Backfill Sorozo", mustLocID(t))
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)
	if _, err := db.DB.Exec(`UPDATE entries SET url = $1 WHERE id = $2`, "https://www.backfill-sorozo.com/etlap", id); err != nil {
		t.Fatal(err)
	}
	account.MigrateWebsites()
	var key, status string
	var entryID int
	err := db.DB.QueryRow(`SELECT domain_key, status, entry_id FROM websites WHERE domain_key = $1`, "backfill-sorozo.com").Scan(&key, &status, &entryID)
	if err != nil || key != "backfill-sorozo.com" || status != "approved" {
		t.Fatalf("backfill %q %q err %v", key, status, err)
	}
}
