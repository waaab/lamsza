package main

import (
	"backend/internal/account"
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/events"
	"backend/internal/handlers"
	"backend/internal/links"
	"backend/internal/middleware"
	"backend/internal/mondasok"
	"backend/internal/news"
	"backend/internal/search"
	"backend/internal/weather"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var testMux *http.ServeMux

func init() {
	config.Load()
	config.AppConfig.GoogleClientID = "test-google-client-id"
	config.AppConfig.AdminGoogleEmails = []string{"admin@test.lamsza"}
	auth.VerifyIDToken = auth.ParseTestIDToken
	db.InitDB()
	mondasok.Migrate()
	handlers.MigrateEntryVerified()
	handlers.MigrateEntryReviews()
	handlers.MigrateEntrySuggestions()
	handlers.MigrateEntryProfileView()
	handlers.MigrateEntryLocationSearch()
	handlers.MigrateSettlementLocationTypes()
	events.Migrate()
	handlers.MigrateAttractions()
	auth.Migrate()
	account.Migrate()
	account.MigrateWebsites()
	handlers.MigrateDirectoryCatalog()

	testMux = http.NewServeMux()
	testMux.HandleFunc("/api/account/preferences", middleware.ApplyCORS(account.HandlePreferences))
	testMux.HandleFunc("/api/account/import", middleware.ApplyCORS(account.HandleImport))
	testMux.HandleFunc("/api/account/links", middleware.ApplyCORS(account.HandleLinks))
	testMux.HandleFunc("/api/account/history", middleware.ApplyCORS(account.HandleHistory))
	testMux.HandleFunc("/api/account/favorites", middleware.ApplyCORS(account.HandleFavorites))
	testMux.HandleFunc("/api/auth/google", middleware.ApplyCORS(auth.HandleGoogleLogin))
	testMux.HandleFunc("/api/auth/me", middleware.ApplyCORS(auth.HandleMe))
	testMux.HandleFunc("/api/auth/logout", middleware.ApplyCORS(auth.HandleLogout))
	testMux.HandleFunc("/api/account/listings/catalog", middleware.ApplyCORS(account.HandleListingCatalog))
	testMux.HandleFunc("/api/account/listings/claim", middleware.ApplyCORS(account.HandleClaimListing))
	testMux.HandleFunc("/api/account/listings/members", middleware.ApplyCORS(account.HandleListingMembers))
	testMux.HandleFunc("/api/account/listings", middleware.ApplyCORS(account.HandleListings))
	testMux.HandleFunc("/api/account/websites/lookup", middleware.ApplyCORS(account.HandleWebsiteLookup))
	testMux.HandleFunc("/api/account/websites", middleware.ApplyCORS(account.HandleAccountWebsites))
	testMux.HandleFunc("/api/websites", middleware.ApplyCORS(account.HandleWebsites))
	testMux.HandleFunc("/api/entries", middleware.ApplyCORS(handlers.EntriesHandler))
	testMux.HandleFunc("/api/directory", middleware.ApplyCORS(handlers.EntriesHandler))
	testMux.HandleFunc("/api/entry-categories", middleware.ApplyCORS(handlers.HandlePublicEntryCategories))
	testMux.HandleFunc("/api/entry", middleware.ApplyCORS(handlers.EntryDetailHandler))
	testMux.HandleFunc("/api/entry/related", middleware.ApplyCORS(handlers.HandleEntryRelated))
	testMux.HandleFunc("/api/entry/reviews", middleware.ApplyCORS(handlers.HandleEntryReviews))
	testMux.HandleFunc("/api/entry/suggestion-form", middleware.ApplyCORS(account.HandleSuggestionForm))
	testMux.HandleFunc("/api/entry/suggestions", middleware.ApplyCORS(account.HandleEntrySuggestions))
	testMux.HandleFunc("/api/locations", middleware.ApplyCORS(handlers.HandlePublicLocations))
	testMux.HandleFunc("/api/settlement_location_types", middleware.ApplyCORS(handlers.HandlePublicSettlementLocationTypes))
	testMux.HandleFunc("/api/events", middleware.ApplyCORS(events.HandleEvents))
	testMux.HandleFunc("/api/news", middleware.ApplyCORS(news.HandleNews))
	testMux.HandleFunc("/api/news/feeds", middleware.ApplyCORS(news.HandlePublicNewsFeeds))
	testMux.HandleFunc("/api/weather/county", middleware.ApplyCORS(weather.HandleCountyWeather))
	testMux.HandleFunc("/api/mondasok", middleware.ApplyCORS(mondasok.HandlePublicMondasok))
	testMux.HandleFunc("/api/quick_links", middleware.ApplyCORS(links.HandlePublicQuickLinks))
	testMux.HandleFunc("/api/proxy", middleware.ApplyCORS(search.ProxyHandler))
	testMux.HandleFunc("/api/autosuggest", middleware.ApplyCORS(search.HandleAutosuggest))
	testMux.HandleFunc("/api/search", middleware.ApplyCORS(search.HandleUnifiedSearch))
}

func mustLogin(email string) *http.Cookie {
	body, _ := json.Marshal(map[string]string{"credential": "test:" + email})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	testMux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		panic("test google login failed: " + rr.Body.String())
	}
	for _, c := range rr.Result().Cookies() {
		if c.Name == auth.SessionCookieName {
			return c
		}
	}
	panic("test google login missing session cookie")
}

func requestBody(t *testing.T, body interface{}) *bytes.Buffer {
	if body == nil {
		return bytes.NewBuffer(nil)
	}
	if r, ok := body.(io.Reader); ok {
		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		return bytes.NewBuffer(data)
	}
	b, _ := json.Marshal(body)
	return bytes.NewBuffer(b)
}

func doRequestWithCookie(t *testing.T, method, path string, body interface{}, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	reqBody := requestBody(t, body)
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", testOrigin)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	testMux.ServeHTTP(rr, req)
	return rr
}

func doAnonRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	return doRequestWithCookie(t, method, path, body, nil)
}

func doRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	reqBody := requestBody(t, body)
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", testOrigin)
	rr := httptest.NewRecorder()
	testMux.ServeHTTP(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// CORS
// ---------------------------------------------------------------------------

// testOrigin is on the allowlist, so the test requests look like calls from one
// of our own pages. The allowlist itself is covered in
// internal/middleware/middleware_test.go.
const testOrigin = "https://lamsza.com"

func TestCORSHeaders(t *testing.T) {
	rr := doRequest(t, "OPTIONS", "/api/entries", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("OPTIONS /api/entries: expected 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, testOrigin)
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Missing Access-Control-Allow-Methods header")
	}
}

func TestCORSRefusesForeignOrigin(t *testing.T) {
	req := httptest.NewRequest("OPTIONS", "/api/entries", nil)
	req.Header.Set("Origin", "https://evil.test")
	rr := httptest.NewRecorder()
	testMux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("preflight from a foreign origin: got %d, want 403", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("foreign origin was reflected back as %q", got)
	}
}

// ---------------------------------------------------------------------------
// Public endpoints
// ---------------------------------------------------------------------------

func TestGetEntries(t *testing.T) {
	rr := doRequest(t, "GET", "/api/entries?q=", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/entries: expected 200, got %d", rr.Code)
	}
	var entries []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &entries); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
	}
}

func TestGetDirectory(t *testing.T) {
	rr := doRequest(t, "GET", "/api/directory?q=", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/directory: expected 200, got %d", rr.Code)
	}
	var entries []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &entries); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
	}
}

func TestGetEntryMissing(t *testing.T) {
	rr := doRequest(t, "GET", "/api/entry?slug=nonexistent-slug-xyz-999", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET /api/entry with bad slug: expected 404, got %d", rr.Code)
	}
}

func TestGetEntryMissingSlugParam(t *testing.T) {
	rr := doRequest(t, "GET", "/api/entry", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/entry without slug: expected 400, got %d", rr.Code)
	}
}

func TestGetLocations(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/locations: expected 200, got %d", rr.Code)
	}
	var locs []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &locs); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
	}
}

func TestGetEvents(t *testing.T) {
	rr := doRequest(t, "GET", "/api/events", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/events: expected 200, got %d", rr.Code)
	}
	var payload struct {
		Events []map[string]interface{} `json:"events"`
		Total  int                      `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Response is not {events,total}: %v body=%s", err, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Autosuggest
// ---------------------------------------------------------------------------

func TestAutosuggestTooShort(t *testing.T) {
	rr := doRequest(t, "GET", "/api/autosuggest?q=ab", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Autosuggest short query: expected 200, got %d", rr.Code)
	}
	var results []string
	json.Unmarshal(rr.Body.Bytes(), &results)
	if len(results) != 0 {
		t.Errorf("Expected empty results for short query, got %d", len(results))
	}
}

func TestAutosuggestValid(t *testing.T) {
	rr := doRequest(t, "GET", "/api/autosuggest?q=csik", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Autosuggest: expected 200, got %d", rr.Code)
	}
	var results []string
	if err := json.Unmarshal(rr.Body.Bytes(), &results); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
	}
}

// ---------------------------------------------------------------------------
// News (parsed RSS)
// ---------------------------------------------------------------------------

func TestGetNews(t *testing.T) {
	rr := doRequest(t, "GET", "/api/news?limit=5", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/news: expected 200, got %d", rr.Code)
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &items); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
	}
}

func TestGetPublicNewsFeeds(t *testing.T) {
	rr := doRequest(t, "GET", "/api/news/feeds", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/news/feeds: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin == "" {
		t.Fatal("GET /api/news/feeds: missing CORS header")
	}
	var feeds []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &feeds); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
	}
}

// ---------------------------------------------------------------------------
// County Weather
// ---------------------------------------------------------------------------

func TestCountyWeather(t *testing.T) {
	rr := doRequest(t, "GET", "/api/weather/county?slug=hargita", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/weather/county: expected 200, got %d", rr.Code)
	}
	var results []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &results); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
	}
	if len(results) == 0 {
		t.Log("Warning: no weather results for Hargita county (may be API key issue)")
	}
}

// ---------------------------------------------------------------------------
// Set County Seat
// ---------------------------------------------------------------------------

// Real megyeszékhely for each county. Tests must leave these in place.
var canonicalCountySeats = []struct {
	countySlug     string
	settlementSlug string
}{
	{"hargita", "csikszereda"},
	{"kovaszna", "sepsiszentgyorgy"},
	{"maros", "marosvasarhely"},
}

// ---------------------------------------------------------------------------
// Admin CRUD: Entry Categories
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Admin CRUD: Entry Types
// ---------------------------------------------------------------------------

func TestPreferredSettlementDoesNotFollowFilters(t *testing.T) {
	cookie := mustLogin("preferred-place@test.lamsza")
	rr := doRequest(t, "GET", "/api/locations", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("locations: %d %s", rr.Code, rr.Body.String())
	}
	var locs []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &locs); err != nil {
		t.Fatal(err)
	}
	var settlementID int
	var slug string
	for _, loc := range locs {
		if strings.EqualFold(fmt.Sprint(loc["type"]), "megye") {
			continue
		}
		id, _ := loc["id"].(float64)
		s, _ := loc["slug"].(string)
		if id > 0 && s != "" {
			settlementID = int(id)
			slug = s
			break
		}
	}
	if settlementID == 0 {
		t.Skip("no settlement to save")
	}

	rr = doRequestWithCookie(t, "PUT", "/api/account/preferences", map[string]int{
		"preferred_settlement_id": settlementID,
	}, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("save preferred: %d %s", rr.Code, rr.Body.String())
	}
	var me map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	pref, _ := me["preferred_location"].(map[string]interface{})
	if pref == nil || pref["slug"] != slug {
		t.Fatalf("preferred_location = %#v, want slug %s", me["preferred_location"], slug)
	}

	rr = doRequestWithCookie(t, "PUT", "/api/account/preferences", map[string]int{
		"preferred_settlement_id": 0,
	}, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("clear preferred: %d %s", rr.Code, rr.Body.String())
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me["preferred_location"] != nil {
		t.Fatalf("expected cleared preferred_location, got %#v", me["preferred_location"])
	}
}

func TestGetPublicQuickLinks(t *testing.T) {
	rr := doRequest(t, "GET", "/api/quick_links", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/quick_links: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin == "" {
		t.Fatal("GET /api/quick_links: missing CORS header")
	}
	var links []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &links); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD: Quick Links
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Admin CRUD: Mondasok
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Admin CRUD: News Feeds
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Admin CRUD: Entries (full cycle)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Admin CRUD: Entries (full cycle)
// ---------------------------------------------------------------------------

func TestPublicEntryVerifiedSeparateFromClaimed(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test verified entry")
	}
	locID := locs[0]["id"]

	id, slug := createEntryFixture(t, entryFixture{
		Name:       "Verified Separate Test",
		LocationID: locID,
		Verified:   true,
	})
	defer deleteEntryFixture(t, id)

	rr = doRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var got map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got["verified"] != true {
		t.Fatalf("public entry verified should be true, got %v", got["verified"])
	}
	if got["claimed"] != false {
		t.Fatalf("public entry claimed should be false until ownership exists, got %v", got["claimed"])
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD: Events
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Search via FTS
// ---------------------------------------------------------------------------

func TestSearchFTS(t *testing.T) {
	rr := doRequest(t, "GET", "/api/entries?q=csikszereda", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("FTS search: expected 200, got %d", rr.Code)
	}
	var entries []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &entries); err != nil {
		t.Fatalf("FTS search response not valid JSON: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Proxy requires url param
// ---------------------------------------------------------------------------

func TestProxyMissingURL(t *testing.T) {
	rr := doRequest(t, "GET", "/api/proxy", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Proxy without url: expected 400, got %d", rr.Code)
	}
}

func TestProxyRejectsArbitraryAndPrivateURLs(t *testing.T) {
	rr := doAnonRequest(t, "GET", "/api/proxy?url="+url.QueryEscape("http://example.com"), nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("example.com: expected 403, got %d body %s", rr.Code, rr.Body.String())
	}
	rr = doAnonRequest(t, "GET", "/api/proxy?url="+url.QueryEscape("http://127.0.0.1/secret"), nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("loopback: expected 403, got %d body %s", rr.Code, rr.Body.String())
	}
	rr = doAnonRequest(t, "POST", "/api/proxy?url="+url.QueryEscape("http://example.com"), nil)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST proxy: expected 405, got %d", rr.Code)
	}
}

func TestProxyAllowsStoredFeedURL(t *testing.T) {
	const feed = "https://feeds.example.test/only-stored"
	var id int
	err := db.DB.QueryRow(
		`INSERT INTO news_feeds (title, feed_url, bg_color) VALUES ($1, $2, $3) RETURNING id`,
		"Proxy Allow Test", feed, "#ffffff",
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	defer db.DB.Exec(`DELETE FROM news_feeds WHERE id = $1`, id)

	rr := doAnonRequest(t, "GET", "/api/proxy?url="+url.QueryEscape(feed), nil)
	if rr.Code == http.StatusForbidden || rr.Code == http.StatusBadRequest {
		t.Fatalf("stored feed should pass the allowlist, got %d %s", rr.Code, rr.Body.String())
	}
}

func TestPublicLocationWriteRejected(t *testing.T) {
	rr := doAnonRequest(t, "POST", "/api/locations", map[string]string{"name": "Nope"})
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/locations: expected 405, got %d body %s", rr.Code, rr.Body.String())
	}
}

func TestClaimFreeListingBecomesOwner(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test claim")
	}
	locID := locs[0]["id"]

	entryID, slug := createEntry(t, "Claim Test Entry", locID)
	defer deleteEntryFixture(t, entryID)

	ownerCookie := mustLogin("owner@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var claimResp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &claimResp)
	if claimResp["role"] != "owner" {
		t.Fatalf("claim role: expected owner, got %v", claimResp["role"])
	}
	if claimResp["status"] != "pending" {
		t.Fatalf("claim status: expected pending, got %v", claimResp["status"])
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry: expected 200, got %d", rr.Code)
	}
	var pub map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &pub)
	if pub["claimed"] != false {
		t.Fatalf("public claimed should stay false until an admin accepts, got %v", pub["claimed"])
	}
	if pub["claim_pending"] != true {
		t.Fatalf("claim_pending should be true, got %v", pub["claim_pending"])
	}

	otherCookie := mustLogin("other-claim@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, otherCookie)
	if rr.Code != http.StatusConflict {
		t.Fatalf("second claim while pending: expected 409, got %d; body: %s", rr.Code, rr.Body.String())
	}

	acceptPendingClaim(t, entryID, "owner@test.lamsza")

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	json.Unmarshal(rr.Body.Bytes(), &pub)
	if pub["claimed"] != true {
		t.Fatalf("public claimed should be true after admin accepts, got %v", pub["claimed"])
	}
	if pub["claim_pending"] != false {
		t.Fatalf("claim_pending should be false after accept, got %v", pub["claim_pending"])
	}

	memberCookie := mustLogin("member@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, memberCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("second claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var claim2 map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &claim2)
	if claim2["role"] != "member" {
		t.Fatalf("second claim role: expected member, got %v", claim2["role"])
	}
	if claim2["status"] != "pending" {
		t.Fatalf("second claim status: expected pending, got %v", claim2["status"])
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	json.Unmarshal(rr.Body.Bytes(), &pub)
	if pub["claimed"] != true {
		t.Fatalf("public claimed should stay true with pending member, got %v", pub["claimed"])
	}
}

func TestMemberCanPatchListing(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test patch")
	}
	locID := locs[0]["id"]

	entryID, _ := createEntryFixture(t, entryFixture{
		Name:       "Patch Test Entry",
		LocationID: locID,
		Verified:   true,
	})
	defer deleteEntryFixture(t, entryID)

	var typeID, categoryID, locationID int
	var published, verified bool
	err := db.DB.QueryRow(`
		SELECT type_id, category_id, location_id, published, verified
		FROM entries WHERE id = $1
	`, entryID).Scan(&typeID, &categoryID, &locationID, &published, &verified)
	if err != nil {
		t.Fatalf("entry baseline: %v", err)
	}

	ownerCookie := mustLogin("owner@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	acceptPendingClaim(t, entryID, "owner@test.lamsza")

	memberCookie := mustLogin("member@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, memberCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("member claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var memberUserID int
	err = db.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, "member@test.lamsza").Scan(&memberUserID)
	if err != nil {
		t.Fatalf("member user id: %v", err)
	}
	_, err = db.DB.Exec(`UPDATE entry_members SET status = 'active' WHERE entry_id = $1 AND user_id = $2`, entryID, memberUserID)
	if err != nil {
		t.Fatalf("activate member: %v", err)
	}

	patchBody := map[string]interface{}{
		"name":        "Patched Member Name",
		"location_id": locationID,
		"category_id": categoryID,
		"type_id":     typeID,
		"photos":      []map[string]interface{}{{"url": "javascript:alert(1)", "alt": "x"}},
	}
	rr = doRequestWithCookie(t, "PATCH", "/api/account/listings?id="+formatID(entryID), patchBody, memberCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("member PATCH listing: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var patchedName string
	var gotPublished, gotVerified bool
	var photosJSON string
	err = db.DB.QueryRow(`SELECT name, published, verified, photos::text FROM entries WHERE id = $1`, entryID).Scan(&patchedName, &gotPublished, &gotVerified, &photosJSON)
	if err != nil {
		t.Fatalf("entry after patch: %v", err)
	}
	if patchedName != "Patched Member Name" {
		t.Fatalf("patched name: expected Patched Member Name, got %q", patchedName)
	}
	if gotPublished != published {
		t.Fatalf("published should stay %v, got %v", published, gotPublished)
	}
	if gotVerified != verified {
		t.Fatalf("verified should stay %v, got %v", verified, gotVerified)
	}
	if photosJSON != "[]" {
		t.Fatalf("javascript photo should be dropped, got photos %s", photosJSON)
	}

	strangerCookie := mustLogin("stranger@test.lamsza")
	rr = doRequestWithCookie(t, "PATCH", "/api/account/listings?id="+formatID(entryID), patchBody, strangerCookie)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("non-member PATCH listing: expected 403, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestMemberCannotDeleteListing(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test delete")
	}
	locID := locs[0]["id"]

	entryID, slug := createEntry(t, "Delete Test Entry", locID)

	ownerCookie := mustLogin("owner@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	acceptPendingClaim(t, entryID, "owner@test.lamsza")

	memberCookie := mustLogin("member@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, memberCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("member claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var memberUserID int
	err := db.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, "member@test.lamsza").Scan(&memberUserID)
	if err != nil {
		t.Fatalf("member user id: %v", err)
	}
	_, err = db.DB.Exec(`UPDATE entry_members SET status = 'active' WHERE entry_id = $1 AND user_id = $2`, entryID, memberUserID)
	if err != nil {
		t.Fatalf("activate member: %v", err)
	}

	rr = doRequestWithCookie(t, "DELETE", "/api/account/listings?id="+formatID(entryID), nil, memberCookie)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("member DELETE listing: expected 403, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("public GET after member delete attempt: expected 200, got %d", rr.Code)
	}

	rr = doRequestWithCookie(t, "DELETE", "/api/account/listings/members?entry_id="+formatID(entryID)+"&user_id="+java(memberUserID), nil, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("owner kick member: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequestWithCookie(t, "DELETE", "/api/account/listings?id="+formatID(entryID), nil, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("owner DELETE listing: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("public GET after owner delete: expected 404, got %d", rr.Code)
	}
}

func TestListingRejectsInvalidURL(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test listing url validation")
	}
	locID := locs[0]["id"]

	var categoryID, typeID int
	if err := db.DB.QueryRow(`SELECT id FROM entry_categories WHERE parent_id IS NOT NULL ORDER BY id ASC LIMIT 1`).Scan(&categoryID); err != nil {
		t.Skip("No entry categories in DB; cannot test listing url validation")
	}
	if err := db.DB.QueryRow(`SELECT id FROM entry_types ORDER BY id ASC LIMIT 1`).Scan(&typeID); err != nil {
		t.Skip("No entry types in DB; cannot test listing url validation")
	}

	userCookie := mustLogin("url-test@test.lamsza")
	// Creating a listing with a URL also reserves that domain. A previous run
	// leaves the website row behind after the entry is deleted.
	if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, "example.com"); err != nil {
		t.Fatalf("cleanup reserved example.com: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, "example.com"); err != nil {
			t.Errorf("cleanup reserved example.com: %v", err)
		}
	})
	createBody := map[string]interface{}{
		"name":        "URL Test Entry",
		"location_id": locID,
		"category_id": categoryID,
		"type_id":     typeID,
		"url":         "javascript:alert(1)",
	}
	rr = doRequestWithCookie(t, "POST", "/api/account/listings", createBody, userCookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("POST with javascript url: expected 400, got %d; body: %s", rr.Code, rr.Body.String())
	}

	createBody["url"] = "https://example.com"
	rr = doRequestWithCookie(t, "POST", "/api/account/listings", createBody, userCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST with https url: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := created["id"]
	defer deleteEntryFixture(t, entryID)

	var storedURL string
	err := db.DB.QueryRow(`SELECT COALESCE(url, '') FROM entries WHERE id = $1`, entryID).Scan(&storedURL)
	if err != nil {
		t.Fatalf("select url: %v", err)
	}
	if storedURL != "https://example.com" {
		t.Fatalf("stored url: expected https://example.com, got %q", storedURL)
	}

	patchBody := map[string]interface{}{
		"name":        "URL Test Entry",
		"location_id": locID,
		"category_id": categoryID,
		"type_id":     typeID,
		"url":         "javascript:alert(1)",
	}
	rr = doRequestWithCookie(t, "PATCH", "/api/account/listings?id="+formatID(entryID), patchBody, userCookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("PATCH with javascript url: expected 400, got %d; body: %s", rr.Code, rr.Body.String())
	}

	err = db.DB.QueryRow(`SELECT COALESCE(url, '') FROM entries WHERE id = $1`, entryID).Scan(&storedURL)
	if err != nil {
		t.Fatalf("select url after patch reject: %v", err)
	}
	if storedURL != "https://example.com" {
		t.Fatalf("url should be unchanged after rejected PATCH, got %q", storedURL)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func formatID(v interface{}) string {
	switch id := v.(type) {
	case float64:
		return java(int(id))
	case int:
		return java(id)
	default:
		return "0"
	}
}

func java(n int) string {
	s := ""
	if n == 0 {
		return "0"
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
