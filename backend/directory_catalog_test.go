package main

import (
	"backend/internal/db"
	"backend/internal/handlers"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func postWebsite(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	cookie := mustLogin("website-submit@test.lamsza")
	return doRequestWithCookie(t, "POST", "/api/websites", strings.NewReader(body), cookie)
}

func TestWebsiteRequiresLeafCategory(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	defer func() {
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, "mobonline.ro"); err != nil {
			t.Errorf("cleanup mobonline.ro: %v", err)
		}
	}()

	// category 4 is Vásárlás, a parent. category 39 is Bútor.
	rr := postWebsite(t, `{"domain":"mobonline.ro","title":"Mobonline","description":"Bútor webshop","category_id":4}`)
	if rr.Code != 400 {
		t.Fatalf("parent category: %d %s", rr.Code, rr.Body.String())
	}
	rr = postWebsite(t, `{"domain":"mobonline.ro","title":"Mobonline","description":"Bútor webshop","category_id":39}`)
	if rr.Code != 201 {
		t.Fatalf("leaf category: %d %s", rr.Code, rr.Body.String())
	}
}

func TestWebshopListingWithoutTown(t *testing.T) {
	const domain = "mobonline.ro"
	cleanup := func() {
		if _, err := db.DB.Exec(`
			DELETE FROM entries WHERE id IN (
				SELECT entry_id FROM websites WHERE domain_key = $1 AND entry_id IS NOT NULL
			)`, domain); err != nil {
			t.Errorf("cleanup entries: %v", err)
		}
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, domain); err != nil {
			t.Errorf("cleanup website: %v", err)
		}
	}
	cleanup()
	defer cleanup()

	handlers.MigrateDirectoryCatalog()

	rr := postWebsite(t, `{"domain":"mobonline.ro","title":"Mobonline","description":"Bútor webshop","category_id":39}`)
	if rr.Code != 201 {
		t.Fatalf("submit: %d %s", rr.Code, rr.Body.String())
	}
	var submitted map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &submitted); err != nil {
		t.Fatal(err)
	}
	webID := int(submitted["id"].(float64))

	admin := mustLogin("admin@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/admin/websites", map[string]interface{}{
		"id": webID, "action": "approve",
	}, admin)
	if rr.Code != 200 {
		t.Fatalf("approve: %d %s", rr.Code, rr.Body.String())
	}

	user := mustLogin("website-submit@test.lamsza")
	rr = doRequestWithCookie(t, "POST", "/api/account/listings", map[string]interface{}{
		"website_id":  webID,
		"category_id": 0,
		"type_id":     2,
		"location_id": 0,
		"name":        "Mobonline",
		"url":         "https://mobonline.ro",
	}, user)
	if rr.Code != 200 {
		t.Fatalf("create listing: %d %s", rr.Code, rr.Body.String())
	}
	var listing map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	entryID := int(listing["id"].(float64))
	if int(listing["category_id"].(float64)) != 39 {
		t.Fatalf("category_id = %v, want 39", listing["category_id"])
	}

	var locationID sql.NullInt64
	var categoryID int
	if err := db.DB.QueryRow(`SELECT location_id, category_id FROM entries WHERE id = $1`, entryID).Scan(&locationID, &categoryID); err != nil {
		t.Fatal(err)
	}
	if locationID.Valid {
		t.Fatalf("location_id = %d, want NULL", locationID.Int64)
	}
	if categoryID != 39 {
		t.Fatalf("entries.category_id = %d, want 39", categoryID)
	}
}

func TestDirectoryCatalogSeedIds(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	var flag string
	if err := db.DB.QueryRow(`SELECT value FROM site_settings WHERE key = 'directory_catalog_v2'`).Scan(&flag); err != nil {
		t.Fatal(err)
	}
	if flag != "1" {
		t.Fatalf("flag = %q", flag)
	}
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("entries = %d, want 0", n)
	}
	var name string
	if err := db.DB.QueryRow(`SELECT name FROM entry_types WHERE id = 2`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Vállalkozás" {
		t.Fatalf("type 2 = %q", name)
	}
	if err := db.DB.QueryRow(`SELECT name FROM entry_categories WHERE id = 39`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Bútor" {
		t.Fatalf("category 39 = %q", name)
	}
	var parent int
	if err := db.DB.QueryRow(`SELECT parent_id FROM entry_categories WHERE id = 45`).Scan(&parent); err != nil {
		t.Fatal(err)
	}
	if parent != 5 {
		t.Fatalf("Turbószerviz parent = %d, want 5 Autó", parent)
	}
	var towns int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM settlements`).Scan(&towns); err != nil {
		t.Fatal(err)
	}
	if towns == 0 {
		t.Fatal("settlements were wiped")
	}
	var seqName string
	if err := db.DB.QueryRow(`SELECT pg_get_serial_sequence('entries', 'id')`).Scan(&seqName); err != nil {
		t.Fatal(err)
	}
	if seqName == "" {
		t.Fatal("entries has no serial sequence")
	}
	var lastValue int64
	var isCalled bool
	if err := db.DB.QueryRow(`SELECT last_value, is_called FROM `+seqName).Scan(&lastValue, &isCalled); err != nil {
		t.Fatal(err)
	}
	nextID := lastValue
	if isCalled {
		nextID++
	}
	if nextID != 1 {
		t.Fatalf("entries next id = %d, want 1 (last_value=%d is_called=%v)", nextID, lastValue, isCalled)
	}
}

func TestLegacyCategoryMigrateDoesNotPruneTree(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	handlers.MigrateEntryCategories()
	handlers.MigrateEntryTypes()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_categories`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 68 {
		t.Fatalf("categories = %d, want 68", n)
	}
	var types int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_types`).Scan(&types); err != nil {
		t.Fatal(err)
	}
	if types != 3 {
		t.Fatalf("types = %d, want 3", types)
	}
}

func restoreSeedEtteremCategory(t *testing.T) {
	t.Helper()
	var hasSeed bool
	if err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM entry_categories WHERE id = 11)`).Scan(&hasSeed); err != nil {
		t.Fatal(err)
	}
	if !hasSeed {
		if _, err := db.DB.Exec(`
			INSERT INTO entry_categories (id, name, slug, parent_id, sort_order)
			VALUES (11, 'Étterem', 'etterem', 1, 1)`); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.DB.Exec(`
		DELETE FROM entry_categories c
		WHERE c.name = 'Étterem' AND c.id != 11
		AND NOT EXISTS (SELECT 1 FROM entries e WHERE e.category_id = c.id)`); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCategoryRequiresMove(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	restoreSeedEtteremCategory(t)
	// Étterem is 11. Attach nothing. Delete succeeds.
	rr := doRequest(t, "DELETE", "/api/admin/entry_categories?id=11", nil)
	if rr.Code != 200 {
		t.Fatalf("empty delete: %d %s", rr.Code, rr.Body.String())
	}
	// Recreate Étterem under Étkezés (1) so later tests still have a leaf.
	rr = doRequest(t, "POST", "/api/admin/entry_categories", strings.NewReader(`{"name":"Étterem","parent_id":1}`))
	if rr.Code != 200 {
		t.Fatalf("recreate: %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`
		INSERT INTO entries (name, category_id, type_id, languages)
		VALUES ('Próba étterem', $1, 2, '{HU}')`, created.ID); err != nil {
		t.Fatal(err)
	}
	rr = doRequest(t, "DELETE", "/api/admin/entry_categories?id="+strconv.Itoa(created.ID), nil)
	if rr.Code != 409 {
		t.Fatalf("blocked delete: %d", rr.Code)
	}
	rr = doRequest(t, "DELETE", "/api/admin/entry_categories?id="+strconv.Itoa(created.ID)+"&move_to=12", nil)
	if rr.Code != 200 {
		t.Fatalf("move delete: %d %s", rr.Code, rr.Body.String())
	}
	var cat int
	if err := db.DB.QueryRow(`SELECT category_id FROM entries WHERE name = 'Próba étterem'`).Scan(&cat); err != nil {
		t.Fatal(err)
	}
	if cat != 12 {
		t.Fatalf("moved category = %d, want 12 Kávézó", cat)
	}
	_, _ = db.DB.Exec(`DELETE FROM entries WHERE name = 'Próba étterem'`)
	restoreSeedEtteremCategory(t)
}

func TestCategoryBrowseIncludesWebshop(t *testing.T) {
	const domain = "browse-webshop-test.ro"
	cleanup := func() {
		if _, err := db.DB.Exec(`
			DELETE FROM entries WHERE slug IN ('butor-bolt-browse-test', 'butor-webshop-browse-test')
		`); err != nil {
			t.Errorf("cleanup entries: %v", err)
		}
		if _, err := db.DB.Exec(`DELETE FROM websites WHERE domain_key = $1`, domain); err != nil {
			t.Errorf("cleanup website: %v", err)
		}
	}
	cleanup()
	defer cleanup()

	handlers.MigrateDirectoryCatalog()

	locID := mustLocID(t)
	var townEntryID int
	if err := db.DB.QueryRow(`
		INSERT INTO entries (type_id, location_id, category_id, cat_name, name, slug, languages, published)
		VALUES (2, $1, 39, 'Bútor', 'Bútor bolt', 'butor-bolt-browse-test', '{HU}', true)
		RETURNING id
	`, locID).Scan(&townEntryID); err != nil {
		t.Fatal(err)
	}
	var webshopEntryID int
	if err := db.DB.QueryRow(`
		INSERT INTO entries (type_id, location_id, category_id, cat_name, name, slug, languages, published)
		VALUES (2, NULL, 39, 'Bútor', 'Bútor webshop', 'butor-webshop-browse-test', '{HU}', true)
		RETURNING id
	`).Scan(&webshopEntryID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`
		INSERT INTO websites (domain_key, submitted_host, title, description, status, category_id)
		VALUES ($1, $1, 'Browse Test Shop', 'Bútor webshop', 'approved', 39)
	`, domain); err != nil {
		t.Fatal(err)
	}

	entryIDsFrom := func(path string) map[int]bool {
		rr := doRequest(t, "GET", path, nil)
		if rr.Code != 200 {
			t.Fatalf("%s: %d %s", path, rr.Code, rr.Body.String())
		}
		var rows []map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
			t.Fatalf("%s json: %v", path, err)
		}
		ids := map[int]bool{}
		for _, row := range rows {
			switch v := row["id"].(type) {
			case float64:
				ids[int(v)] = true
			case string:
				n, _ := strconv.Atoi(v)
				if n > 0 {
					ids[n] = true
				}
			}
		}
		return ids
	}

	butor := entryIDsFrom("/api/entries?category=butor")
	if !butor[townEntryID] || !butor[webshopEntryID] {
		t.Fatalf("category=butor: got ids %v, want %d and %d", butor, townEntryID, webshopEntryID)
	}

	vasarlas := entryIDsFrom("/api/entries?category=vasarlas")
	if !vasarlas[townEntryID] || !vasarlas[webshopEntryID] {
		t.Fatalf("category=vasarlas: got ids %v, want %d and %d", vasarlas, townEntryID, webshopEntryID)
	}

	townOnly := entryIDsFrom("/api/entries?location_id=" + formatID(locID))
	if !townOnly[townEntryID] || townOnly[webshopEntryID] {
		t.Fatalf("location_id filter: got ids %v, want only %d", townOnly, townEntryID)
	}

	rr := doRequest(t, "GET", "/api/websites", nil)
	if rr.Code != 200 {
		t.Fatalf("GET /api/websites: %d %s", rr.Code, rr.Body.String())
	}
	var webPayload struct {
		Websites []struct {
			Domain   string `json:"domain"`
			Category string `json:"category"`
		} `json:"websites"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &webPayload); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, w := range webPayload.Websites {
		if w.Domain == domain && w.Category == "Bútor" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("approved unlinked website missing from list: %+v", webPayload.Websites)
	}
}

func TestEntryTypesStayClosed(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	rr := doRequest(t, "POST", "/api/admin/entry_types", strings.NewReader(`{"name":"Weboldal"}`))
	if rr.Code != 403 {
		t.Fatalf("create type: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "A típuslista zárt.") {
		t.Fatalf("create type body: %s", rr.Body.String())
	}
	rr = doRequest(t, "PUT", "/api/admin/entry_types", strings.NewReader(`{"id":1,"name":"Más"}`))
	if rr.Code != 403 {
		t.Fatalf("update type: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "A típuslista zárt.") {
		t.Fatalf("update type body: %s", rr.Body.String())
	}
	rr = doRequest(t, "DELETE", "/api/admin/entry_types?id=1", nil)
	if rr.Code != 403 {
		t.Fatalf("delete type: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "A típuslista zárt.") {
		t.Fatalf("delete type body: %s", rr.Body.String())
	}
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_types`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("types = %d", n)
	}
}

func TestPublicEntryCategories(t *testing.T) {
	rr := doAnonRequest(t, "GET", "/api/entry-categories", nil)
	if rr.Code != 200 {
		t.Fatalf("public categories: %d %s", rr.Code, rr.Body.String())
	}
	var rows []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		ParentID *int   `json:"parent_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	var butor, vasarlas bool
	for _, row := range rows {
		if row.Name == "Bútor" && row.Slug == "butor" && row.ParentID != nil && *row.ParentID == 4 {
			butor = true
		}
		if row.Name == "Vásárlás" && row.ParentID == nil {
			vasarlas = true
		}
	}
	if !butor || !vasarlas {
		t.Fatalf("catalog missing Bútor or Vásárlás: %d rows", len(rows))
	}
	rr = doRequest(t, "POST", "/api/admin/entries", map[string]interface{}{
		"name": "Parent Shelf Entry", "category_id": 1, "type": "Vállalkozás",
	})
	if rr.Code != 400 {
		t.Fatalf("parent category on entry: %d %s", rr.Code, rr.Body.String())
	}
}
