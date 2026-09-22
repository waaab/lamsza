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
	"backend/internal/settings"
	"backend/internal/weather"
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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
	settings.MigrateSiteSettings()
	weather.MigrateWeatherTranslations()
	events.Migrate()
	handlers.MigrateEntryTypes()
	handlers.MigrateEntryCategories()
	handlers.MigrateEntryAvailability()
	handlers.MigrateEntryVerified()
	auth.Migrate()
	account.Migrate()

	pub := middleware.ApplyCORS
	admin := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.ApplyCORS(auth.RequireAdmin(h))
	}

	testMux = http.NewServeMux()
	testMux.HandleFunc("/api/auth/google", pub(auth.HandleGoogleLogin))
	testMux.HandleFunc("/api/auth/me", pub(auth.HandleMe))
	testMux.HandleFunc("/api/auth/logout", pub(auth.HandleLogout))
	testMux.HandleFunc("/api/account/preferences", pub(account.HandlePreferences))
	testMux.HandleFunc("/api/account/import", pub(account.HandleImport))
	testMux.HandleFunc("/api/account/links", pub(account.HandleLinks))
	testMux.HandleFunc("/api/account/history", pub(account.HandleHistory))
	testMux.HandleFunc("/api/account/favorites", pub(account.HandleFavorites))
	testMux.HandleFunc("/api/account/listings/claim", pub(account.HandleClaimListing))
	testMux.HandleFunc("/api/account/listings/catalog", pub(account.HandleListingCatalog))
	testMux.HandleFunc("/api/account/listings/members", pub(account.HandleListingMembers))
	testMux.HandleFunc("/api/account/listings", pub(account.HandleListings))
	testMux.HandleFunc("/api/entries", pub(handlers.EntriesHandler))
	testMux.HandleFunc("/api/directory", pub(handlers.EntriesHandler))
	testMux.HandleFunc("/api/entry", pub(handlers.EntryDetailHandler))
	testMux.HandleFunc("/api/entry/related", pub(handlers.HandleEntryRelated))
	testMux.HandleFunc("/api/locations", pub(handlers.HandleAdminLocations))
	testMux.HandleFunc("/api/admin/listing-queue", admin(account.HandleListingQueue))
	testMux.HandleFunc("/api/admin/listing-queue/publish", admin(account.HandleListingQueuePublish))
	testMux.HandleFunc("/api/admin/listing-queue/member", admin(account.HandleListingQueueMember))
	testMux.HandleFunc("/api/admin/entries", admin(handlers.HandleAdminEntries))
	testMux.HandleFunc("/api/admin/entry-images", admin(handlers.HandleEntryImageUpload))
	testMux.Handle("/api/media/entry-images/", http.StripPrefix("/api/media/entry-images/", http.FileServer(http.Dir(handlers.EntryImagesDir()))))
	_ = handlers.EnsureEntryImagesDir()
	testMux.HandleFunc("/api/admin/entry_categories", admin(handlers.HandleAdminEntryCategories))
	testMux.HandleFunc("/api/admin/entry_types", admin(handlers.HandleAdminEntryTypes))
	testMux.HandleFunc("/api/admin/locations", admin(handlers.HandleAdminLocations))
	testMux.HandleFunc("/api/admin/county_seat", admin(handlers.HandleSetCountySeat))
	testMux.HandleFunc("/api/events", pub(events.HandleEvents))
	testMux.HandleFunc("/api/admin/events", admin(events.HandleAdminEvents))
	testMux.HandleFunc("/api/news", pub(news.HandleNews))
	testMux.HandleFunc("/api/news/feeds", pub(news.HandlePublicNewsFeeds))
	testMux.HandleFunc("/api/admin/news_feeds", admin(news.HandleAdminNewsFeeds))
	testMux.HandleFunc("/api/weather/county", pub(weather.HandleCountyWeather))
	testMux.HandleFunc("/api/admin/weather_translations", admin(weather.HandleAdminWeatherTranslations))
	testMux.HandleFunc("/api/mondasok", pub(mondasok.HandlePublicMondasok))
	testMux.HandleFunc("/api/admin/mondasok", admin(mondasok.HandleAdminMondasok))
	testMux.HandleFunc("/api/quick_links", pub(links.HandlePublicQuickLinks))
	testMux.HandleFunc("/api/admin/quick_links", admin(links.HandleAdminQuickLinks))
	testMux.HandleFunc("/api/proxy", pub(search.ProxyHandler))
	testMux.HandleFunc("/api/autosuggest", pub(search.HandleAutosuggest))

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

func doRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	return doRequestWithCookie(t, method, path, body, testAdminCookie)
}

func doAnonRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	return doRequestWithCookie(t, method, path, body, nil)
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
		t.Fatalf("GET /api/events: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var payload struct {
		Events []map[string]interface{} `json:"events"`
		Total  int                      `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}
	if payload.Events == nil {
		t.Fatal("GET /api/events: expected events array in payload")
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

func weatherCacheVersion(t *testing.T) string {
	t.Helper()
	var v string
	if err := db.DB.QueryRow("SELECT value FROM site_settings WHERE key = 'weather_cache_version'").Scan(&v); err != nil {
		t.Fatalf("weather_cache_version: %v", err)
	}
	return v
}

func TestAdminWeatherTranslationsBumpCacheVersion(t *testing.T) {
	payload := map[string]string{
		"source_text":     "tdd-weather-cache-bump",
		"lang":            "hu",
		"translated_text": "teszt fordítás cache",
	}
	_, _ = db.DB.Exec("DELETE FROM weather_desc_translations WHERE source_text = $1 AND lang = $2", payload["source_text"], payload["lang"])

	beforePost := weatherCacheVersion(t)
	rr := doRequest(t, "POST", "/api/admin/weather_translations", payload)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST weather_translations: expected 201, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("POST weather_translations JSON: %v", err)
	}
	id := created["id"]
	t.Cleanup(func() {
		_, _ = db.DB.Exec("DELETE FROM weather_desc_translations WHERE source_text = $1 AND lang = $2", payload["source_text"], payload["lang"])
	})
	afterPost := weatherCacheVersion(t)
	if afterPost == beforePost {
		t.Fatalf("POST weather_translations should bump weather_cache_version; still %q", afterPost)
	}

	beforePut := afterPost
	rr = doRequest(t, "PUT", "/api/admin/weather_translations", map[string]interface{}{
		"id":              id,
		"source_text":     payload["source_text"],
		"lang":            payload["lang"],
		"translated_text": "teszt fordítás cache frissítve",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT weather_translations: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	afterPut := weatherCacheVersion(t)
	if afterPut == beforePut {
		t.Fatalf("PUT weather_translations should bump weather_cache_version; still %q", afterPut)
	}

	beforeDelete := afterPut
	rr = doRequest(t, "DELETE", "/api/admin/weather_translations?id="+formatID(id), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE weather_translations: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	afterDelete := weatherCacheVersion(t)
	if afterDelete == beforeDelete {
		t.Fatalf("DELETE weather_translations should bump weather_cache_version; still %q", afterDelete)
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

func TestEntriesCategoryVarcharDropped(t *testing.T) {
	var n int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'entries' AND column_name = 'category'
	`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("entries.category varchar still present")
	}
}

func TestEntryCategoriesNotBloated(t *testing.T) {
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_categories`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n > 20 {
		t.Fatalf("entry_categories bloated: %d rows (want a short catalog)", n)
	}
	var unused int
	if err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM entry_categories c
		LEFT JOIN entries e ON e.category_id = c.id
		WHERE e.id IS NULL
	`).Scan(&unused); err != nil {
		t.Fatal(err)
	}
	// Seed labels may be unused until listings exist; junk leftovers must be gone.
	if unused > 8 {
		t.Fatalf("too many unused entry_categories: %d", unused)
	}
}

func TestEveryEntryHasCatalogCategory(t *testing.T) {
	var missing int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM entries e
		LEFT JOIN entry_categories c ON c.id = e.category_id
		WHERE e.category_id IS NULL OR c.id IS NULL
	`).Scan(&missing)
	if err != nil {
		t.Fatal(err)
	}
	if missing != 0 {
		t.Fatalf("%d entries missing category_id catalog join", missing)
	}
}

func TestMigrateEntryCategoriesPreservesAdminRename(t *testing.T) {
	var id int
	var orig, slug string
	err := db.DB.QueryRow(`
		SELECT id, name, COALESCE(slug, '')
		FROM entry_categories
		WHERE slug = 'mesteremberek'
		   OR name IN ('Mesteremberek', 'Mesterember')
		ORDER BY id
		LIMIT 1
	`).Scan(&id, &orig, &slug)
	if err != nil {
		t.Fatalf("seed mesteremberek row: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.DB.Exec(`UPDATE entry_categories SET name = $1, slug = $2 WHERE id = $3`, orig, slug, id)
	})
	if _, err := db.DB.Exec(`UPDATE entry_categories SET name = 'Mesterember' WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	handlers.MigrateEntryCategories()
	var got string
	if err := db.DB.QueryRow(`SELECT name FROM entry_categories WHERE id = $1`, id).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != "Mesterember" {
		t.Fatalf("startup seed overwrote admin rename: got %q, want Mesterember", got)
	}
}

func TestSeedHistoricalSeatsPreservesAdminContent(t *testing.T) {
	const marker = "admin-edit-do-not-clobber"
	var orig string
	err := db.DB.QueryRow(`SELECT COALESCE(content, '') FROM historical_seats WHERE slug = 'csikszek'`).Scan(&orig)
	if err != nil {
		t.Fatalf("csikszek: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.DB.Exec(`UPDATE historical_seats SET content = $1 WHERE slug = 'csikszek'`, orig)
	})
	if _, err := db.DB.Exec(`UPDATE historical_seats SET content = $1 WHERE slug = 'csikszek'`, marker); err != nil {
		t.Fatal(err)
	}
	db.SeedHistoricalSeatsContent()
	var got string
	if err := db.DB.QueryRow(`SELECT content FROM historical_seats WHERE slug = 'csikszek'`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != marker {
		t.Fatalf("startup seed overwrote admin seat content: got %q", got)
	}
}

func TestAdminEntryCategoryPersistsViaCatalogFK(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test entry category FK")
	}

	rr = doRequest(t, "POST", "/api/admin/entries", map[string]interface{}{
		"name":        "IntegTest CategoryFK Entry",
		"location_id": locs[0]["id"],
		"type":        "service",
		"category":    "hivatalok",
		"languages":   []string{"HU"},
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	var catID int
	var catName string
	err := db.DB.QueryRow(`
		SELECT e.category_id, c.name
		FROM entries e
		JOIN entry_categories c ON c.id = e.category_id
		WHERE e.id = $1
	`, int(id.(float64))).Scan(&catID, &catName)
	if err != nil {
		t.Fatalf("entry category_id join failed: %v", err)
	}
	if catName != "Hivatalok" {
		t.Fatalf("catalog category name = %q, want Hivatalok", catName)
	}
	if created["category"] != "Hivatalok" {
		t.Fatalf("POST response category = %v, want Hivatalok", created["category"])
	}
}

func TestCannotDeleteInUseEntryCategory(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test in-use category delete")
	}

	rr = doRequest(t, "POST", "/api/admin/entry_categories", map[string]string{"name": "FKGuardCategory_IntegTest"})
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entry_categories: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var createdCat map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &createdCat)
	catID := createdCat["id"]

	rr = doRequest(t, "POST", "/api/admin/entries", map[string]interface{}{
		"name":        "IntegTest Category FK Guard",
		"location_id": locs[0]["id"],
		"type":        "service",
		"category":    "FKGuardCategory_IntegTest",
		"languages":   []string{"HU"},
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := created["id"]
	defer doRequest(t, "DELETE", "/api/admin/entry_categories?id="+formatID(catID), nil)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	rr = doRequest(t, "DELETE", "/api/admin/entry_categories?id="+formatID(catID), nil)
	if rr.Code != http.StatusConflict {
		t.Fatalf("DELETE in-use entry category: expected 409, got %d; body: %s", rr.Code, rr.Body.String())
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

func TestEntriesTypeIdIsForeignKey(t *testing.T) {
	var conname string
	err := db.DB.QueryRow(`
		SELECT c.conname
		FROM pg_constraint c
		JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)
		WHERE c.conrelid = 'public.entries'::regclass
		  AND c.contype = 'f'
		  AND a.attname = 'type_id'
		LIMIT 1
	`).Scan(&conname)
	if err != nil {
		t.Fatalf("entries.type_id FOREIGN KEY missing: %v", err)
	}
}

func TestAdminEntryTypePersistsViaCatalogFK(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test entry type FK")
	}

	rr = doRequest(t, "POST", "/api/admin/entries", map[string]interface{}{
		"name":        "IntegTest TypeFK Entry",
		"location_id": locs[0]["id"],
		"type":        "service",
		"languages":   []string{"HU"},
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(id), nil)

	var typeID int
	var typeName string
	err := db.DB.QueryRow(`
		SELECT e.type_id, et.name
		FROM entries e
		JOIN entry_types et ON et.id = e.type_id
		WHERE e.id = $1
	`, int(id.(float64))).Scan(&typeID, &typeName)
	if err != nil {
		t.Fatalf("entry type_id join failed: %v", err)
	}
	if typeID == 0 {
		t.Fatal("type_id is 0")
	}
	if typeName != "Szolgáltatás" {
		t.Fatalf("catalog type name = %q, want Szolgáltatás", typeName)
	}
	if created["type"] != "Szolgáltatás" {
		t.Fatalf("POST response type = %v, want Szolgáltatás", created["type"])
	}
}

func TestCannotDeleteInUseEntryType(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)
	if len(locs) == 0 {
		t.Skip("No locations in DB; cannot test in-use entry type delete")
	}

	rr = doRequest(t, "POST", "/api/admin/entry_types", map[string]string{"name": "FKGuardType_IntegTest"})
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entry_types: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var createdType map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &createdType)
	typeID := createdType["id"]

	rr = doRequest(t, "POST", "/api/admin/entries", map[string]interface{}{
		"name":        "IntegTest FK Guard Entry",
		"location_id": locs[0]["id"],
		"type":        "FKGuardType_IntegTest",
		"languages":   []string{"HU"},
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	entryID := created["id"]
	defer doRequest(t, "DELETE", "/api/admin/entry_types?id="+formatID(typeID), nil)
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	rr = doRequest(t, "DELETE", "/api/admin/entry_types?id="+formatID(typeID), nil)
	if rr.Code != http.StatusConflict {
		t.Fatalf("DELETE in-use entry type: expected 409, got %d; body: %s", rr.Code, rr.Body.String())
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
		"text":         "Test mondas for integration testing",
		"display_date": "2030-06-15",
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
	slug, _ := created["slug"].(string)
	if slug == "" {
		t.Fatal("POST entries missing slug")
	}
	id := created["id"]

	rr = doAnonRequest(t, "GET", "/api/entry?slug="+slug, nil)
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
		"verified":    true,
		"hours": map[string]interface{}{
			"mon": map[string]interface{}{"open": "09:00", "close": "17:00"},
		},
		"photos": []map[string]interface{}{
			{
				"url":         "https://cdn.example.com/integ-entry.jpg",
				"alt":         "Udvar",
				"title":       "Udvar",
				"description": "Nyári fény",
				"width":       800,
				"height":      600,
			},
		},
	}

	// CREATE
	rr = doRequest(t, "POST", "/api/admin/entries", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entries: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	id := created["id"]
	if created["type"] != "Egyéb" {
		t.Fatalf("POST type \"entry\" should persist as Egyéb, got %v", created["type"])
	}
	if created["verified"] != true {
		t.Fatalf("POST verified should persist as true, got %v", created["verified"])
	}
	if slug, ok := created["slug"].(string); ok && slug != "" {
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
		photos, _ := got["photos"].([]interface{})
		if len(photos) != 1 {
			t.Fatalf("public entry photos should have 1 item, got %#v", got["photos"])
		}
		photo, _ := photos[0].(map[string]interface{})
		if photo["url"] != "https://cdn.example.com/integ-entry.jpg" || photo["alt"] != "Udvar" {
			t.Fatalf("public entry photo mismatch: %#v", photo)
		}
	}

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

func TestEntryImageUpload(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatal(err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "yard.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(pngBuf.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/entry-images", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.AddCookie(testAdminCookie)
	rr := httptest.NewRecorder()
	testMux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("upload: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var data map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	url, _ := data["url"].(string)
	if url == "" || !strings.HasPrefix(url, "/api/media/entry-images/") {
		t.Fatalf("bad url: %#v", data)
	}
	if data["width"] != float64(4) || data["height"] != float64(3) {
		t.Fatalf("size mismatch: %#v", data)
	}

	rr = doAnonRequest(t, "GET", url, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET media: expected 200, got %d", rr.Code)
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
// Admin auth (Google session)
// ---------------------------------------------------------------------------

func TestAdminEntriesRequiresAuth(t *testing.T) {
	rr := doAnonRequest(t, "GET", "/api/admin/entries", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated GET /api/admin/entries: expected 401, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestPublicDirectoryUnauthenticated(t *testing.T) {
	rr := doAnonRequest(t, "GET", "/api/directory", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("public GET /api/directory: expected 200, got %d", rr.Code)
	}
}

func TestGoogleLoginAdminSession(t *testing.T) {
	rr := doRequest(t, "GET", "/api/auth/me", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/auth/me: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var me map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &me); err != nil {
		t.Fatalf("me json: %v", err)
	}
	if me["email"] != "admin@test.lamsza" {
		t.Fatalf("me email: got %#v", me["email"])
	}
	if me["is_admin"] != true {
		t.Fatalf("me is_admin: got %#v", me["is_admin"])
	}
	rr = doRequest(t, "GET", "/api/admin/entries", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("admin session GET /api/admin/entries: expected 200, got %d", rr.Code)
	}
}

func TestGoogleLoginNonAdminForbidden(t *testing.T) {
	cookie := mustLogin("visitor@example.com")
	rr := doRequestWithCookie(t, "GET", "/api/auth/me", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("visitor /api/auth/me: expected 200, got %d", rr.Code)
	}
	var me map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &me)
	if me["is_admin"] == true {
		t.Fatal("visitor must not be admin")
	}
	rr = doRequestWithCookie(t, "GET", "/api/admin/entries", nil, cookie)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("visitor GET /api/admin/entries: expected 403, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestLogoutClearsAdminSession(t *testing.T) {
	cookie := mustLogin("admin@test.lamsza")
	rr := doRequestWithCookie(t, "POST", "/api/auth/logout", map[string]bool{"ok": true}, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("logout: expected 200, got %d", rr.Code)
	}
	rr = doRequestWithCookie(t, "GET", "/api/admin/entries", nil, cookie)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("after logout GET /api/admin/entries: expected 401, got %d", rr.Code)
	}
}

func TestPublicNewsFeedsUnauthenticated(t *testing.T) {
	rr := doAnonRequest(t, "GET", "/api/news/feeds", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("public GET /api/news/feeds: expected 200, got %d", rr.Code)
	}
}

func TestPublicQuickLinksUnauthenticated(t *testing.T) {
	rr := doAnonRequest(t, "GET", "/api/quick_links", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("public GET /api/quick_links: expected 200, got %d", rr.Code)
	}
}

func TestEntryRelated(t *testing.T) {
	rr := doAnonRequest(t, "GET", "/api/entry/related", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/entry/related without slug: expected 400, got %d", rr.Code)
	}

	rr = doAnonRequest(t, "GET", "/api/entry/related?slug=does-not-exist-xyz", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET /api/entry/related with bad slug: expected 404, got %d", rr.Code)
	}

	rr = doAnonRequest(t, "GET", "/api/locations", nil)
	var locs []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &locs)

	countyGroups := make(map[string][]map[string]interface{})
	for _, loc := range locs {
		cs, _ := loc["county_slug"].(string)
		if cs == "" {
			cs, _ = loc["county"].(string)
		}
		if cs != "" {
			countyGroups[cs] = append(countyGroups[cs], loc)
		}
	}

	var loc1, loc2 map[string]interface{}
	for _, group := range countyGroups {
		if len(group) >= 2 {
			loc1 = group[0]
			loc2 = group[1]
			break
		}
	}
	if loc1 == nil || loc2 == nil {
		t.Skip("Need at least two settlements in the same county")
	}

	createEntry := func(name string, locID interface{}) (id interface{}, slug string) {
		payload := map[string]interface{}{
			"name":        name,
			"location_id": locID,
			"type":        "entry",
			"category":    "Egyéb",
		}
		rr := doRequest(t, "POST", "/api/admin/entries", payload)
		if rr.Code != http.StatusOK {
			t.Fatalf("POST entry %s: expected 200, got %d; body: %s", name, rr.Code, rr.Body.String())
		}
		var created map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &created)
		slug, _ = created["slug"].(string)
		return created["id"], slug
	}

	idA, slugA := createEntry("RelatedInteg A", loc1["id"])
	idB, _ := createEntry("RelatedInteg B", loc1["id"])
	idC, _ := createEntry("RelatedInteg C", loc2["id"])

	defer func() {
		doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(idA), nil)
		doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(idB), nil)
		doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(idC), nil)
	}()

	rr = doAnonRequest(t, "GET", "/api/entry/related?slug="+slugA, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/entry/related: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	nearby, _ := resp["nearby"].([]interface{})
	related, _ := resp["related"].([]interface{})

	entryID := func(m map[string]interface{}) string {
		switch v := m["id"].(type) {
		case string:
			return v
		case float64:
			return formatID(v)
		default:
			return ""
		}
	}

	idAStr := formatID(idA)
	idBStr := formatID(idB)
	idCStr := formatID(idC)

	bIdx, cIdx := -1, -1
	for i, item := range nearby {
		m, _ := item.(map[string]interface{})
		eid := entryID(m)
		if eid == idBStr {
			bIdx = i
		}
		if eid == idCStr {
			cIdx = i
		}
		if eid == idAStr {
			t.Fatal("nearby must not contain A")
		}
	}
	if bIdx < 0 {
		t.Fatalf("nearby should contain B (same settlement), got %#v", nearby)
	}
	if cIdx >= 0 && cIdx <= bIdx {
		t.Fatalf("if C appears in nearby, it should be after same-settlement rows (B at %d, C at %d)", bIdx, cIdx)
	}

	nearbyIDs := make(map[string]bool)
	for _, item := range nearby {
		m, _ := item.(map[string]interface{})
		nearbyIDs[entryID(m)] = true
	}
	for _, item := range related {
		m, _ := item.(map[string]interface{})
		eid := entryID(m)
		if eid == idAStr {
			t.Fatal("related must not contain A")
		}
		if nearbyIDs[eid] {
			t.Fatal("related must not contain any nearby id")
		}
		cat, _ := m["category"].(string)
		if cat != "Egyéb" {
			t.Fatalf("related items should share A's category (Egyéb), got %q", cat)
		}
	}

	if len(nearby) > 8 {
		t.Fatalf("len(nearby) should be <= 8, got %d", len(nearby))
	}
	if len(related) > 8 {
		t.Fatalf("len(related) should be <= 8, got %d", len(related))
	}
}

func TestAdminPublishDoesNotVerify(t *testing.T) {
	rr := doRequest(t, "GET", "/api/locations", nil)
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
	defer doRequest(t, "DELETE", "/api/admin/entries?id="+formatID(entryID), nil)

	rr = doRequest(t, "GET", "/api/admin/listing-queue", nil)
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

	rr = doRequest(t, "GET", "/api/auth/me", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/auth/me: expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var meBefore map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &meBefore)
	countBefore, ok := meBefore["admin_queue_count"].(float64)
	if !ok {
		t.Fatalf("admin_queue_count missing before publish: %#v", meBefore)
	}

	rr = doRequest(t, "POST", "/api/admin/listing-queue/publish", map[string]interface{}{
		"entry_id": entryID,
	})
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

	rr = doRequest(t, "GET", "/api/auth/me", nil)
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
