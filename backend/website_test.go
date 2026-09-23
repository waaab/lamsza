package main

import (
	"backend/internal/account"
	"backend/internal/db"
	"encoding/json"
	"net/http"
	"strings"
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

func TestWebsiteSubmitPendingRaisesQueue(t *testing.T) {
	defer func() {
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, "submit-example.com"); err != nil {
			t.Errorf("cleanup submit-example.com: %v", err)
		}
	}()

	cookie := mustLogin("website-submit@test.lamsza")
	adminCookie := mustLogin("admin@test.lamsza")
	before := adminQueueCount(t, adminCookie)
	rr := doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain":      "https://www.submit-example.com/a",
		"title":       "<b>Submit Example</b>",
		"description": "A short page.",
	}, cookie)
	if rr.Code != 201 {
		t.Fatalf("submit: %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	if created["domain"] != "submit-example.com" || created["title"] != "Submit Example" || created["status"] != "pending" {
		t.Fatalf("created %#v", created)
	}
	if adminQueueCount(t, adminCookie) != before+1 {
		t.Fatalf("queue count did not rise")
	}
	pub := doRequest(t, "GET", "/api/websites", nil)
	if strings.Contains(pub.Body.String(), "submit-example.com") {
		t.Fatal("pending website was public")
	}
	dup := doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "submit-example.com", "title": "Other", "description": "Other page.",
	}, cookie)
	if dup.Code != 409 || !strings.Contains(dup.Body.String(), "domain_pending") {
		t.Fatalf("dup: %d %s", dup.Code, dup.Body.String())
	}
}

func TestWebsiteSubmitRejectsBannedAndSignedOut(t *testing.T) {
	defer func() {
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, "banned-example.com"); err != nil {
			t.Errorf("cleanup banned-example.com: %v", err)
		}
	}()

	rr := doRequest(t, "POST", "/api/websites", map[string]string{
		"domain": "banned-example.com", "title": "T", "description": "D",
	})
	if rr.Code != 401 {
		t.Fatalf("signed out: %d", rr.Code)
	}
	cookie := mustLogin("website-banned@test.lamsza")
	if _, err := db.DB.Exec(`UPDATE users SET website_banned = true WHERE email = $1`, "website-banned@test.lamsza"); err != nil {
		t.Fatal(err)
	}
	rr = doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "banned-example.com", "title": "T", "description": "D",
	}, cookie)
	if rr.Code != 403 || !strings.Contains(rr.Body.String(), "website_banned") {
		t.Fatalf("banned: %d %s", rr.Code, rr.Body.String())
	}
}

func adminQueueCount(t *testing.T, cookie *http.Cookie) int {
	t.Helper()
	rr := doRequestWithCookie(t, "GET", "/api/auth/me", nil, cookie)
	var me map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &me)
	n, ok := me["admin_queue_count"].(float64)
	if !ok {
		t.Fatalf("no queue count %#v", me)
	}
	return int(n)
}
