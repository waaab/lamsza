package main

import (
	"backend/internal/account"
	"backend/internal/db"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestMigrateWebsitesBackfillsListingURL(t *testing.T) {
	if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, "backfill-sorozo.com"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, "backfill-sorozo.com"); err != nil {
			t.Errorf("cleanup backfill-sorozo.com: %v", err)
		}
	}()

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
	wantEntryID := int(id.(float64))
	if entryID != wantEntryID {
		t.Fatalf("entry_id %d want %d", entryID, wantEntryID)
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
	var approver string
	var approvedAt sql.NullTime
	if err := db.DB.QueryRow(`
		SELECT COALESCE(u.email, ''), w.approved_at
		FROM websites w
		LEFT JOIN users u ON u.id = w.approved_by
		WHERE w.domain_key = $1
	`, "review-example.com").Scan(&approver, &approvedAt); err != nil {
		t.Fatal(err)
	}
	if approver != "admin@test.lamsza" || !approvedAt.Valid {
		t.Fatalf("approval record approver %q at %v", approver, approvedAt)
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

func cleanupClaimTestDomains(t *testing.T) {
	for _, d := range []string{"claim-example.com", "claim-free.com", "evil.example"} {
		if _, err := db.DB.Exec(`DELETE FROM entries WHERE url = $1`, "https://"+d); err != nil {
			t.Errorf("cleanup entry url %s: %v", d, err)
		}
		if _, err := db.DB.Exec(`
			DELETE FROM entries WHERE id IN (SELECT entry_id FROM websites WHERE domain_key = $1 AND entry_id IS NOT NULL)
		`, d); err != nil {
			t.Errorf("cleanup entries %s: %v", d, err)
		}
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, d); err != nil {
			t.Errorf("cleanup website %s: %v", d, err)
		}
	}
	for _, name := range []string{"Manifesto", "Second", "Banned claim"} {
		if _, err := db.DB.Exec(`DELETE FROM entries WHERE name = $1`, name); err != nil {
			t.Errorf("cleanup entry name %s: %v", name, err)
		}
	}
	if _, err := db.DB.Exec(`UPDATE users SET website_banned = false WHERE email = $1`, "website-claim-ban@test.lamsza"); err != nil {
		t.Errorf("cleanup ban flag: %v", err)
	}
}

func TestClaimWebsiteCreatesUnpublishedListing(t *testing.T) {
	cleanupClaimTestDomains(t)
	defer cleanupClaimTestDomains(t)

	owner := mustLogin("website-owner@test.lamsza")
	other := mustLogin("website-joiner@test.lamsza")
	admin := mustLogin("admin@test.lamsza")
	webID := submitWebsite(t, owner, "claim-example.com", "Claim Title", "Claim blurb")
	doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": webID, "action": "approve"}, admin)

	locID := mustLocID(t)
	var catID, typeID int
	db.DB.QueryRow(`SELECT id FROM entry_categories ORDER BY id ASC LIMIT 1`).Scan(&catID)
	db.DB.QueryRow(`SELECT id FROM entry_types ORDER BY id ASC LIMIT 1`).Scan(&typeID)
	rr := doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"website_id": webID, "name": "Manifesto", "location_id": locID,
		"category_id": catID, "type_id": typeID, "url": "https://evil.example",
	}, owner)
	if rr.Code != 200 {
		t.Fatalf("create %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	if created["published"] != false || created["verified"] != false || created["role"] != "owner" {
		t.Fatalf("created %#v", created)
	}
	var title, url string
	var published, verified bool
	db.DB.QueryRow(`SELECT w.title, e.url, e.published, e.verified FROM websites w JOIN entries e ON e.id = w.entry_id WHERE w.id = $1`, webID).Scan(&title, &url, &published, &verified)
	if title != "Claim Title" || url != "https://claim-example.com" || published || verified {
		t.Fatalf("stored title %q url %q published %v verified %v", title, url, published, verified)
	}
	before := searchJSON(t, "claim-example.com")
	if len(before["entries"].([]interface{})) != 0 || len(before["websites"].([]interface{})) != 1 {
		t.Fatalf("before publish %#v", before)
	}
	entryID := int(created["id"].(float64))
	doRequestWithCookie(t, "POST", "/api/admin/listing-queue/publish", map[string]interface{}{"entry_id": entryID}, admin)
	after := searchJSON(t, "www.claim-example.com")
	ents := after["entries"].([]interface{})
	if len(ents) != 1 || ents[0].(map[string]interface{})["claimed"] != true {
		t.Fatalf("domain entries %#v", ents)
	}
	if after["websites"].([]interface{})[0].(map[string]interface{})["url"] != "https://claim-example.com" {
		t.Fatal("website hit missing")
	}
	byName := searchJSON(t, "Manifesto")
	if byName["website_query"] == true || len(byName["websites"].([]interface{})) != 0 {
		t.Fatalf("name query leaked website %#v", byName["websites"])
	}
	if byName["entries"].([]interface{})[0].(map[string]interface{})["claimed"] != true {
		t.Fatal("name query missing claimed")
	}
	rr = doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"website_id": webID, "name": "Second", "location_id": locID,
		"category_id": catID, "type_id": typeID,
	}, other)
	if rr.Code != 409 {
		t.Fatalf("second listing %d", rr.Code)
	}
	banned := mustLogin("website-claim-ban@test.lamsza")
	db.DB.Exec(`UPDATE users SET website_banned = true WHERE email = $1`, "website-claim-ban@test.lamsza")
	freeID := submitWebsite(t, owner, "claim-free.com", "Free", "Free page.")
	doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": freeID, "action": "approve"}, admin)
	rr = doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"website_id": freeID, "name": "Banned claim", "location_id": locID,
		"category_id": catID, "type_id": typeID,
	}, banned)
	if rr.Code != 403 {
		t.Fatalf("banned claim %d %s", rr.Code, rr.Body.String())
	}
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{"entry_id": entryID}, other)
	var member map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &member)
	if member["role"] != "member" || member["status"] != "pending" {
		t.Fatalf("join %#v", member)
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

func TestWebsiteListingCreateReservesDomain(t *testing.T) {
	const domain = "listing-reserve-example.com"
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

	owner := mustLogin("listing-reserve@test.lamsza")
	locID := mustLocID(t)
	var catID, typeID int
	if err := db.DB.QueryRow(`SELECT id FROM entry_categories ORDER BY id ASC LIMIT 1`).Scan(&catID); err != nil {
		t.Fatal(err)
	}
	if err := db.DB.QueryRow(`SELECT id FROM entry_types ORDER BY id ASC LIMIT 1`).Scan(&typeID); err != nil {
		t.Fatal(err)
	}

	rr := doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"name":        "Reserve Listing",
		"location_id": locID,
		"category_id": catID,
		"type_id":     typeID,
		"url":         "https://www.listing-reserve-example.com/menu",
		"notes":       "Fresh notes here.",
	}, owner)
	if rr.Code != 200 {
		t.Fatalf("create listing: %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := int(created["id"].(float64))
	defer doRequestWithCookie(t, "DELETE", "/api/account/listings?id="+formatID(entryID), nil, owner)

	var key, status, title, description string
	var linkedEntryID, ownerID int
	err := db.DB.QueryRow(`
		SELECT domain_key, status, title, description, entry_id, user_id
		FROM websites WHERE domain_key = $1
	`, domain).Scan(&key, &status, &title, &description, &linkedEntryID, &ownerID)
	if err != nil {
		t.Fatalf("website row missing: %v", err)
	}
	if key != domain || status != "approved" || linkedEntryID != entryID {
		t.Fatalf("website %#v entry %d", map[string]interface{}{
			"key": key, "status": status, "entry_id": linkedEntryID,
		}, entryID)
	}
	if title != "Reserve Listing" || description != "Fresh notes here." {
		t.Fatalf("website text title=%q description=%q", title, description)
	}

	dup := doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": domain, "title": "Other Site", "description": "Taken.",
	}, owner)
	if dup.Code != 409 || !strings.Contains(dup.Body.String(), "domain_taken") {
		t.Fatalf("duplicate submit: %d %s", dup.Code, dup.Body.String())
	}

	rr = doRequestWithCookie(t, "PATCH", "/api/account/listings?id="+formatID(entryID), map[string]interface{}{
		"name":        "Reserve Listing",
		"location_id": locID,
		"category_id": catID,
		"type_id":     typeID,
		"url":         "https://other-domain.com",
	}, owner)
	if rr.Code != 409 || !strings.Contains(rr.Body.String(), "domain_taken") {
		t.Fatalf("domain change: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAccountWebsitesListsOnlyTheSignedInUser(t *testing.T) {
	for _, domain := range []string{"account-mine.example", "account-other.example"} {
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, domain); err != nil {
			t.Fatal(err)
		}
	}
	defer func() {
		for _, domain := range []string{"account-mine.example", "account-other.example"} {
			if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, domain); err != nil {
				t.Errorf("cleanup %s: %v", domain, err)
			}
		}
	}()

	mine := mustLogin("website-mine@test.lamsza")
	other := mustLogin("website-other@test.lamsza")
	created := doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "account-mine.example", "title": "Mine", "description": "My page.",
	}, mine)
	if created.Code != 201 {
		t.Fatalf("mine submit: %d %s", created.Code, created.Body.String())
	}
	otherCreated := doRequestWithCookie(t, "POST", "/api/websites", map[string]string{
		"domain": "account-other.example", "title": "Other", "description": "Their page.",
	}, other)
	if otherCreated.Code != 201 {
		t.Fatalf("other submit: %d %s", otherCreated.Code, otherCreated.Body.String())
	}

	denied := doRequest(t, "GET", "/api/account/websites", nil)
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("signed out: %d", denied.Code)
	}

	rr := doRequestWithCookie(t, "GET", "/api/account/websites", nil, mine)
	if rr.Code != http.StatusOK {
		t.Fatalf("list: %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "account-mine.example") || !strings.Contains(rr.Body.String(), `"status":"pending"`) {
		t.Fatalf("missing own pending website: %s", rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "account-other.example") {
		t.Fatalf("listed another user's website: %s", rr.Body.String())
	}
}

func TestWebsiteLookupFindsAnExistingDomain(t *testing.T) {
	const domain = "lookup-example.com"
	if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, domain); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, domain); err != nil {
			t.Errorf("cleanup %s: %v", domain, err)
		}
	}()

	rr := doRequest(t, "GET", "/api/account/websites/lookup?url=https://www."+domain, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("signed out: %d", rr.Code)
	}

	owner := mustLogin("website-lookup@test.lamsza")
	rr = doRequestWithCookie(t, "GET", "/api/account/websites/lookup?url="+url.QueryEscape(domain), nil, owner)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"website":null`) {
		t.Fatalf("missing domain: %d %s", rr.Code, rr.Body.String())
	}

	id := submitWebsite(t, owner, domain, "Lookup Title", "Lookup page.")
	admin := mustLogin("admin@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{"id": id, "action": "approve"}, admin)
	if rr.Code != http.StatusOK {
		t.Fatalf("approve: %d %s", rr.Code, rr.Body.String())
	}

	rr = doRequestWithCookie(t, "GET", "/api/account/websites/lookup?url="+url.QueryEscape("https://www."+domain+"/menu"), nil, owner)
	if rr.Code != http.StatusOK {
		t.Fatalf("lookup: %d %s", rr.Code, rr.Body.String())
	}
	var body struct {
		Website map[string]interface{} `json:"website"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Website["title"] != "Lookup Title" || body.Website["domain"] != domain || body.Website["claimed"] != false {
		t.Fatalf("lookup body %#v", body.Website)
	}
	if body.Website["entry_id"].(float64) != 0 {
		t.Fatalf("unclaimed website has entry %#v", body.Website["entry_id"])
	}
}
