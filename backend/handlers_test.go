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
	"net/http"
	"net/http/httptest"
	"testing"
)

var testMux *http.ServeMux
var testAdminCookie *http.Cookie

func init() {
	config.Load()
	config.AppConfig.GoogleClientID = "test-google-client-id"
	config.AppConfig.AdminGoogleEmails = []string{"admin@test.lamsza"}
	auth.VerifyIDToken = auth.ParseTestIDToken
	db.InitDB()
	mondasok.Migrate()
	handlers.MigrateEntryVerified()
	auth.Migrate()
	account.Migrate()

	admin := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.ApplyCORS(auth.RequireAdmin(h))
	}

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
	testMux.HandleFunc("/api/entries", middleware.ApplyCORS(handlers.EntriesHandler))
	testMux.HandleFunc("/api/directory", middleware.ApplyCORS(handlers.EntriesHandler))
	testMux.HandleFunc("/api/entry", middleware.ApplyCORS(handlers.EntryDetailHandler))
	testMux.HandleFunc("/api/locations", middleware.ApplyCORS(handlers.HandleAdminLocations))
	testMux.HandleFunc("/api/admin/listing-queue", admin(account.HandleListingQueue))
	testMux.HandleFunc("/api/admin/listing-queue/publish", admin(account.HandleListingQueuePublish))
	testMux.HandleFunc("/api/admin/listing-queue/member", admin(account.HandleListingQueueMember))
	testMux.HandleFunc("/api/admin/entries", middleware.ApplyCORS(handlers.HandleAdminEntries))
	testMux.HandleFunc("/api/admin/entry_categories", middleware.ApplyCORS(handlers.HandleAdminEntryCategories))
	testMux.HandleFunc("/api/admin/entry_types", middleware.ApplyCORS(handlers.HandleAdminEntryTypes))
	testMux.HandleFunc("/api/admin/locations", middleware.ApplyCORS(handlers.HandleAdminLocations))
	testMux.HandleFunc("/api/admin/county_seat", middleware.ApplyCORS(handlers.HandleSetCountySeat))
	testMux.HandleFunc("/api/events", middleware.ApplyCORS(events.HandleEvents))
	testMux.HandleFunc("/api/admin/events", middleware.ApplyCORS(events.HandleAdminEvents))
	testMux.HandleFunc("/api/news", middleware.ApplyCORS(news.HandleNews))
	testMux.HandleFunc("/api/admin/news_feeds", middleware.ApplyCORS(news.HandleAdminNewsFeeds))
	testMux.HandleFunc("/api/weather/county", middleware.ApplyCORS(weather.HandleCountyWeather))
	testMux.HandleFunc("/api/mondasok", middleware.ApplyCORS(mondasok.HandlePublicMondasok))
	testMux.HandleFunc("/api/admin/mondasok", middleware.ApplyCORS(mondasok.HandleAdminMondasok))
	testMux.HandleFunc("/api/admin/quick_links", middleware.ApplyCORS(links.HandleAdminQuickLinks))
	testMux.HandleFunc("/api/proxy", middleware.ApplyCORS(search.ProxyHandler))
	testMux.HandleFunc("/api/autosuggest", middleware.ApplyCORS(search.HandleAutosuggest))

	testAdminCookie = mustLogin("admin@test.lamsza")
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

func doRequestWithCookie(t *testing.T, method, path string, body interface{}, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
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
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	testMux.ServeHTTP(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// CORS
// ---------------------------------------------------------------------------

func TestCORSHeaders(t *testing.T) {
	rr := doRequest(t, "OPTIONS", "/api/entries", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("OPTIONS /api/entries: expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Missing Access-Control-Allow-Origin header")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Missing Access-Control-Allow-Methods header")
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
	var ev []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &ev); err != nil {
		t.Fatalf("Response is not valid JSON array: %v", err)
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

func TestSetCountySeat(t *testing.T) {
	// Get a location to use
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB")
	}

	// Find first city-type location
	var locID float64
	for _, l := range locs {
		lt, _ := l["type"].(string)
		if lt == "város" || lt == "municípium" {
			locID, _ = l["id"].(float64)
			break
		}
	}
	if locID == 0 {
		t.Skip("No city-type locations found")
	}

	// Set as county seat
	rr = doRequest(t, "PUT", "/api/admin/county_seat", map[string]interface{}{"location_id": locID})
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT county_seat: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	// Verify it was set
	rr = doRequest(t, "GET", "/api/locations", nil)
	json.Unmarshal(rr.Body.Bytes(), &locs)
	found := false
	for _, l := range locs {
		id, _ := l["id"].(float64)
		isSeat, _ := l["is_county_seat"].(bool)
		if id == locID && isSeat {
			found = true
			break
		}
	}
	if !found {
		t.Error("Location was not marked as county seat after PUT")
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD: Entry Categories
// ---------------------------------------------------------------------------

func TestAdminEntryCategoriesCRUD(t *testing.T) {
	payload := map[string]string{"name": "TestCategory_IntegTest"}

	// CREATE
	rr := doRequest(t, "POST", "/api/admin/entry_categories", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entry_categories: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]

	// READ
	rr = doRequest(t, "GET", "/api/admin/entry_categories", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry_categories: expected 200, got %d", rr.Code)
	}

	// UPDATE
	updated := map[string]interface{}{"id": id, "name": "TestCategory_Updated"}
	rr = doRequest(t, "PUT", "/api/admin/entry_categories", updated)
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT entry_categories: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	// DELETE
	rr = doRequest(t, "DELETE", "/api/admin/entry_categories?id="+formatID(id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE entry_categories: expected 200, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD: Entry Types
// ---------------------------------------------------------------------------

func TestAdminEntryTypesCRUD(t *testing.T) {
	payload := map[string]string{"name": "TestType_IntegTest"}

	rr := doRequest(t, "POST", "/api/admin/entry_types", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entry_types: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]

	rr = doRequest(t, "GET", "/api/admin/entry_types", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry_types: expected 200, got %d", rr.Code)
	}

	rr = doRequest(t, "PUT", "/api/admin/entry_types", map[string]interface{}{"id": id, "name": "TestType_Updated"})
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT entry_types: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, "DELETE", "/api/admin/entry_types?id="+formatID(id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE entry_types: expected 200, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD: Quick Links
// ---------------------------------------------------------------------------

func TestAdminQuickLinksCRUD(t *testing.T) {
	payload := map[string]string{"title": "TestLink", "url": "https://test-integtest.example.com", "bg_color": "#ffffff"}

	rr := doRequest(t, "POST", "/api/admin/quick_links", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST quick_links: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]

	rr = doRequest(t, "GET", "/api/admin/quick_links", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET quick_links: expected 200, got %d", rr.Code)
	}

	rr = doRequest(t, "PUT", "/api/admin/quick_links", map[string]interface{}{"id": id, "title": "TestLink_Updated", "url": "https://test-integtest.example.com", "bg_color": "#000000"})
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT quick_links: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, "DELETE", "/api/admin/quick_links?id="+formatID(id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE quick_links: expected 200, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD: Mondasok
// ---------------------------------------------------------------------------

func TestAdminMondasokCRUD(t *testing.T) {
	payload := map[string]string{
		"text":          "Test mondas for integration testing",
		"display_date":  "2030-06-15",
	}

	rr := doRequest(t, "POST", "/api/admin/mondasok", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST mondasok: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]

	rr = doRequest(t, "GET", "/api/admin/mondasok", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET mondasok: expected 200, got %d", rr.Code)
	}

	rr = doRequest(t, "PUT", "/api/admin/mondasok", map[string]interface{}{"id": id, "text": "Updated mondas", "display_date": "2030-07-01"})
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT mondasok: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, "DELETE", "/api/admin/mondasok?id="+formatID(id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE mondasok: expected 200, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD: News Feeds
// ---------------------------------------------------------------------------

func TestAdminNewsFeedsCRUD(t *testing.T) {
	payload := map[string]string{"title": "TestFeed", "feed_url": "https://test-integtest-feed.example.com/rss", "bg_color": "#ffebd6"}

	rr := doRequest(t, "POST", "/api/admin/news_feeds", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST news_feeds: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]

	rr = doRequest(t, "GET", "/api/admin/news_feeds", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET news_feeds: expected 200, got %d", rr.Code)
	}

	rr = doRequest(t, "PUT", "/api/admin/news_feeds", map[string]interface{}{"id": id, "title": "TestFeed_Updated", "feed_url": "https://test-integtest-feed.example.com/rss", "bg_color": "#000000"})
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT news_feeds: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, "DELETE", "/api/admin/news_feeds?id="+formatID(id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE news_feeds: expected 200, got %d", rr.Code)
	}
}

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

	payload := map[string]interface{}{
		"name":        "Verified Separate Test",
		"location_id": locID,
		"type":        "service",
		"category":    "Egyéb",
		"verified":    true,
	}

	rr = doRequest(t, "POST", "/api/admin/entries", payload)
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

	doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)
}

func TestAdminEntriesCRUD(t *testing.T) {
	// Need a valid location_id; fetch locations first
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test entry CRUD")
	}
	locID := locs[0]["id"]

	payload := map[string]interface{}{
		"name":        "IntegTest Entry",
		"location_id": locID,
		"type":        "entry",
		"category":    "Egyéb",
		"phone":       "0700-000-000",
		"address":     "Test Address 1",
		"notes":       "Integration test entry",
		"languages":   []string{"HU"},
		"tags":        []string{"integtest"},
	}

	// CREATE
	rr = doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]

	// READ
	rr = doRequest(t, "GET", "/api/admin/entries", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entries: expected 200, got %d", rr.Code)
	}

	// UPDATE
	payload["id"] = id
	payload["notes"] = "Updated notes"
	rr = doRequest(t, "PUT", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	// DELETE
	rr = doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE entries: expected 200, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD: Events
// ---------------------------------------------------------------------------

func TestAdminEventsCRUD(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test event CRUD")
	}
	locID := locs[0]["id"]

	payload := map[string]interface{}{
		"title":       "IntegTest Event",
		"location_id": locID,
		"description": "Test event",
		"start_date":  "2026-12-01",
		"start_time":  "10:00",
		"end_date":    "2026-12-01",
		"end_time":    "18:00",
		"event_type":  "cultural",
		"organizer":   "TestOrg",
	}

	rr = doRequest(t, "POST", "/api/admin/events", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST events: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]

	rr = doRequest(t, "GET", "/api/events", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET events: expected 200, got %d", rr.Code)
	}

	payload["id"] = id
	payload["title"] = "IntegTest Event Updated"
	rr = doRequest(t, "PUT", "/api/admin/events", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT events: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, "DELETE", "/api/admin/events?id="+formatID(id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE events: expected 200, got %d", rr.Code)
	}
}

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


func TestAdminPublishDoesNotVerify(t *testing.T) {
	rr := doRequestWithCookie(t, "GET", "/api/locations", nil, testAdminCookie)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test admin publish queue")
	}
	locID := locs[0]["id"]

	var categoryID, typeID int
	if err := db.DB.QueryRow(`SELECT id FROM entry_categories ORDER BY id ASC LIMIT 1`).Scan(&categoryID); err != nil {
		t.Skip("No entry categories in DB; cannot test admin publish queue")
	}
	if err := db.DB.QueryRow(`SELECT id FROM entry_types ORDER BY id ASC LIMIT 1`).Scan(&typeID); err != nil {
		t.Skip("No entry types in DB; cannot test admin publish queue")
	}

	userCookie := mustLogin("queue-owner@test.lamsza")
	createBody := map[string]interface{}{
		"name":        "Queue Publish Test Entry",
		"location_id": locID,
		"category_id": categoryID,
		"type_id":     typeID,
	}
	rr = doRequestWithCookie(t, "POST", "/api/account/listings", createBody, userCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST /api/account/listings: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := created["id"]
	slug, _ := created["slug"].(string)
	if created["published"] != false {
		t.Fatalf("created listing should be unpublished, got %v", created["published"])
	}
	defer doRequestWithCookie(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil, testAdminCookie)

	rr = doRequestWithCookie(t, "GET", "/api/admin/listing-queue", nil, testAdminCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/admin/listing-queue: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var queue map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &queue)
	unpublished, _ := queue["unpublished"].([]interface{})
	found := false
	for _, item := range unpublished {
		m, _ := item.(map[string]interface{})
		if formatID(m["id"]) == formatID(entryID) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("listing-queue unpublished should include created entry, got %#v", unpublished)
	}

	rr = doRequestWithCookie(t, "GET", "/api/auth/me", nil, testAdminCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/auth/me: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var meBefore map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &meBefore)
	countBefore, ok := meBefore["admin_queue_count"].(float64)
	if !ok {
		t.Fatalf("admin_queue_count missing before publish: %#v", meBefore)
	}

	rr = doRequestWithCookie(t, "POST", "/api/admin/listing-queue/publish", map[string]interface{}{
		"entry_id": entryID,
	}, testAdminCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST /api/admin/listing-queue/publish: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("public GET /api/entry: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var pub map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &pub)
	if pub["verified"] != false {
		t.Fatalf("public entry verified should stay false, got %v", pub["verified"])
	}
	if pub["claimed"] != true {
		t.Fatalf("public entry claimed should be true, got %v", pub["claimed"])
	}

	rr = doRequestWithCookie(t, "GET", "/api/auth/me", nil, testAdminCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/auth/me after publish: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var meAfter map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &meAfter)
	countAfter, ok := meAfter["admin_queue_count"].(float64)
	if !ok {
		t.Fatalf("admin_queue_count missing after publish: %#v", meAfter)
	}
	if countAfter != countBefore-1 {
		t.Fatalf("admin_queue_count should drop by 1: before %v after %v", countBefore, countAfter)
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

	payload := map[string]interface{}{
		"name":        "Claim Test Entry",
		"location_id": locID,
		"type":        "entry",
		"category":    "Egyéb",
	}
	rr = doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := created["id"]
	slug, _ := created["slug"].(string)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

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
	if claimResp["status"] != "active" {
		t.Fatalf("claim status: expected active, got %v", claimResp["status"])
	}

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entry: expected 200, got %d", rr.Code)
	}
	var pub map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &pub)
	if pub["claimed"] != true {
		t.Fatalf("public claimed should be true after owner claim, got %v", pub["claimed"])
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

	payload := map[string]interface{}{
		"name":        "Patch Test Entry",
		"location_id": locID,
		"type":        "entry",
		"category":    "Egyéb",
		"verified":    true,
	}
	rr = doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := created["id"]
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	var typeID, categoryID, locationID int
	var published, verified bool
	err := db.DB.QueryRow(`
		SELECT type_id, category_id, location_id, published, verified
		FROM entries WHERE id = $1
	`, int(entryID.(float64))).Scan(&typeID, &categoryID, &locationID, &published, &verified)
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
	_, err = db.DB.Exec(`UPDATE entry_members SET status = 'active' WHERE entry_id = $1 AND user_id = $2`, int(entryID.(float64)), memberUserID)
	if err != nil {
		t.Fatalf("activate member: %v", err)
	}

	patchBody := map[string]interface{}{
		"name":          "Patched Member Name",
		"location_id":   locationID,
		"category_id":   categoryID,
		"type_id":       typeID,
		"photos":        []map[string]interface{}{{"url": "javascript:alert(1)", "alt": "x"}},
	}
	rr = doRequestWithCookie(t, "PATCH", "/api/account/listings?id="+formatID(entryID), patchBody, memberCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("member PATCH listing: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	rr = doRequest(t, "GET", "/api/admin/entries", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET admin entries: expected 200, got %d", rr.Code)
	}
	var entries []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &entries)
	var found map[string]interface{}
	for _, e := range entries {
		if e["id"] == entryID {
			found = e
			break
		}
	}
	if found == nil {
		t.Fatal("patched entry not found in admin list")
	}
	if found["name"] != "Patched Member Name" {
		t.Fatalf("patched name: expected Patched Member Name, got %v", found["name"])
	}

	var gotPublished, gotVerified bool
	var photosJSON string
	err = db.DB.QueryRow(`SELECT published, verified, photos::text FROM entries WHERE id = $1`, int(entryID.(float64))).Scan(&gotPublished, &gotVerified, &photosJSON)
	if err != nil {
		t.Fatalf("entry after patch: %v", err)
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

	payload := map[string]interface{}{
		"name":        "Delete Test Entry",
		"location_id": locID,
		"type":        "entry",
		"category":    "Egyéb",
	}
	rr = doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := created["id"]
	slug, _ := created["slug"].(string)

	ownerCookie := mustLogin("owner@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings/claim", map[string]interface{}{
		"entry_id": entryID,
	}, ownerCookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("claim: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

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
	_, err = db.DB.Exec(`UPDATE entry_members SET status = 'active' WHERE entry_id = $1 AND user_id = $2`, int(entryID.(float64)), memberUserID)
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
	if err := db.DB.QueryRow(`SELECT id FROM entry_categories ORDER BY id ASC LIMIT 1`).Scan(&categoryID); err != nil {
		t.Skip("No entry categories in DB; cannot test listing url validation")
	}
	if err := db.DB.QueryRow(`SELECT id FROM entry_types ORDER BY id ASC LIMIT 1`).Scan(&typeID); err != nil {
		t.Skip("No entry types in DB; cannot test listing url validation")
	}

	userCookie := mustLogin("url-test@test.lamsza")
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
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	var storedURL string
	err := db.DB.QueryRow(`SELECT COALESCE(url, '') FROM entries WHERE id = $1`, int(entryID.(float64))).Scan(&storedURL)
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

	err = db.DB.QueryRow(`SELECT COALESCE(url, '') FROM entries WHERE id = $1`, int(entryID.(float64))).Scan(&storedURL)
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
