package main

import (
	"backend/internal/account"
	"backend/internal/db"
	"encoding/json"
	"net/http"
	"net/url"
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

func submitWebsite(t *testing.T, cookie *http.Cookie, domain, title, description string) int {
	t.Helper()
	rr := doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": domain, "title": title, "description": description,
	}, cookie)
	if rr.Code != 201 {
		t.Fatalf("submit %s: %d %s", domain, rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id, ok := created["id"].(float64)
	if !ok || id <= 0 {
		t.Fatalf("submit id %#v", created)
	}
	return int(id)
}

func queueBody(t *testing.T, cookie *http.Cookie) string {
	t.Helper()
	rr := doRequestWithCookie(t, "GET", "/api/admin/listing-queue", nil, cookie)
	if rr.Code != 200 {
		t.Fatalf("queue: %d %s", rr.Code, rr.Body.String())
	}
	return rr.Body.String()
}

func queueWebsites(t *testing.T, cookie *http.Cookie) []map[string]interface{} {
	t.Helper()
	rr := doRequestWithCookie(t, "GET", "/api/admin/listing-queue", nil, cookie)
	if rr.Code != 200 {
		t.Fatalf("queue: %d %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	raw, ok := resp["websites"].([]interface{})
	if !ok {
		t.Fatalf("websites missing %#v", resp)
	}
	out := make([]map[string]interface{}, len(raw))
	for i, item := range raw {
		out[i] = item.(map[string]interface{})
	}
	return out
}

func TestWebsiteApproveRejectAndBan(t *testing.T) {
	cleanupReviewTestDomains := func() {
		for _, d := range []string{"review-example.com", "reject-example.com", "ban-example.com", "after-ban.com"} {
			if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, d); err != nil {
				t.Errorf("cleanup %s: %v", d, err)
			}
		}
		if _, err := db.DB.Exec(`UPDATE users SET website_banned = false WHERE email = $1`, "website-review@test.lamsza"); err != nil {
			t.Errorf("cleanup ban flag: %v", err)
		}
	}
	cleanupReviewTestDomains()
	defer cleanupReviewTestDomains()

	user := mustLogin("website-review@test.lamsza")
	admin := mustLogin("admin@test.lamsza")
	id := submitWebsite(t, user, "review-example.com", "Review", "A page.")
	q := queueWebsites(t, admin)
	if len(q) != 1 || q[0]["domain"] != "review-example.com" || q[0]["title"] != "Review" || q[0]["submitter"] == "" {
		t.Fatalf("queue %#v", q)
	}
	rr := doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": id, "action": "approve"}, admin)
	if rr.Code != 200 {
		t.Fatalf("approve %d %s", rr.Code, rr.Body.String())
	}
	if strings.Contains(queueBody(t, admin), "review-example.com") {
		t.Fatal("approved website stayed in the queue")
	}
	pub := doRequest(t, "GET", "/api/websites", nil)
	if !strings.Contains(pub.Body.String(), "review-example.com") {
		t.Fatal("approved website was not public")
	}

	id = submitWebsite(t, user, "reject-example.com", "Reject", "Gone.")
	rr = doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": id, "action": "reject"}, admin)
	if rr.Code != 200 {
		t.Fatalf("reject %d", rr.Code)
	}
	again := submitWebsite(t, user, "reject-example.com", "Reject", "Again.")
	if again <= 0 {
		t.Fatal("rejected key was not freed")
	}
	doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": again, "action": "reject"}, admin)

	banID := submitWebsite(t, user, "ban-example.com", "Ban", "Nope.")
	rr = doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": banID, "action": "ban"}, admin)
	if rr.Code != 200 {
		t.Fatalf("ban %d %s", rr.Code, rr.Body.String())
	}
	rr = doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "after-ban.com", "title": "T", "description": "D",
	}, user)
	if rr.Code != 403 {
		t.Fatalf("banned resubmit %d", rr.Code)
	}
	mustLogin("website-review@test.lamsza")
}

func searchJSON(t *testing.T, q string) map[string]interface{} {
	t.Helper()
	rr := doRequest(t, "GET", "/api/search?q="+url.QueryEscape(q), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("search %q: %d %s", q, rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("search decode: %v", err)
	}
	return resp
}

func TestWebsiteSearchDomainAndTitle(t *testing.T) {
	defer func() {
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, "search-example.com"); err != nil {
			t.Errorf("cleanup search-example.com: %v", err)
		}
	}()

	user := mustLogin("website-search@test.lamsza")
	admin := mustLogin("admin@test.lamsza")
	id := submitWebsite(t, user, "search-example.com", "Search Title", "Unique blurb zzq")
	doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": id, "action": "approve"}, admin)

	domain := searchJSON(t, "https://www.search-example.com/path")
	if domain["website_query"] != true {
		t.Fatalf("expected domain query %#v", domain["website_query"])
	}
	sites := domain["websites"].([]interface{})
	if len(sites) != 1 {
		t.Fatalf("sites %#v", sites)
	}
	hit := sites[0].(map[string]interface{})
	if hit["url"] != "https://search-example.com" || hit["title"] != "Search Title" {
		t.Fatalf("hit %#v", hit)
	}
	if ents, ok := domain["entries"].([]interface{}); ok && len(ents) != 0 {
		t.Fatalf("unclaimed domain query returned entries %#v", ents)
	}

	title := searchJSON(t, "Unique blurb zzq")
	if title["website_query"] == true {
		t.Fatal("title query was treated as a domain")
	}
	if len(title["websites"].([]interface{})) != 1 {
		t.Fatalf("title sites %#v", title["websites"])
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
