package main

import (
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

func init() {
	config.Load()
	db.InitDB()
	mondasok.Migrate()

	testMux = http.NewServeMux()
	testMux.HandleFunc("/api/entries", middleware.ApplyCORS(handlers.EntriesHandler))
	testMux.HandleFunc("/api/directory", middleware.ApplyCORS(handlers.EntriesHandler))
	testMux.HandleFunc("/api/entry", middleware.ApplyCORS(handlers.EntryDetailHandler))
	testMux.HandleFunc("/api/locations", middleware.ApplyCORS(handlers.HandleAdminLocations))
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
