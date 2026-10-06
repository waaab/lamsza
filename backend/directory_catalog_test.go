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
	var maxID int64
	if err := db.DB.QueryRow(`SELECT COALESCE(MAX(id), 0) FROM entries`).Scan(&maxID); err != nil {
		t.Fatal(err)
	}
	if nextID <= maxID {
		t.Fatalf("entries next id = %d, want > max id %d (last_value=%d is_called=%v)", nextID, maxID, lastValue, isCalled)
	}
}

// TestDirectoryCatalogMigrateKeepsRowsWithoutFlag is the BOG-15 guard. The boot
// migrator must never empty the directory, not even when the
// directory_catalog_v2 marker is gone. The wipe lives in
// migrations/0001_directory_catalog_v2.sql and is run by hand.
func TestDirectoryCatalogMigrateKeepsRowsWithoutFlag(t *testing.T) {
	handlers.MigrateDirectoryCatalog()

	var savedFlag sql.NullString
	if err := db.DB.QueryRow(`SELECT value FROM site_settings WHERE key = 'directory_catalog_v2'`).Scan(&savedFlag); err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if !savedFlag.Valid {
			return
		}
		if _, err := db.DB.Exec(`
			INSERT INTO site_settings (key, value) VALUES ('directory_catalog_v2', $1)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, savedFlag.String); err != nil {
			t.Errorf("restore flag: %v", err)
		}
	})

	const probe = "bog15-wipe-probe"
	var entryID int
	if err := db.DB.QueryRow(`
		INSERT INTO entries (name, slug, category_id, type_id, languages)
		VALUES ($1, $1, 39, 2, '{HU}') RETURNING id`, probe).Scan(&entryID); err != nil {
		t.Fatal(err)
	}
	var tagID int
	if err := db.DB.QueryRow(`INSERT INTO tags (name) VALUES ($1) RETURNING id`, probe).Scan(&tagID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.DB.Exec(`DELETE FROM entries WHERE id = $1`, entryID); err != nil {
			t.Errorf("cleanup entry: %v", err)
		}
		if _, err := db.DB.Exec(`DELETE FROM tags WHERE id = $1`, tagID); err != nil {
			t.Errorf("cleanup tag: %v", err)
		}
	})

	var categoriesBefore, typesBefore int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_categories`).Scan(&categoriesBefore); err != nil {
		t.Fatal(err)
	}
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_types`).Scan(&typesBefore); err != nil {
		t.Fatal(err)
	}

	// Lose the marker, the way a dump restore or an admin delete would.
	if _, err := db.DB.Exec(`DELETE FROM site_settings WHERE key = 'directory_catalog_v2'`); err != nil {
		t.Fatal(err)
	}

	handlers.MigrateDirectoryCatalog()

	var stillThere int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entries WHERE id = $1`, entryID).Scan(&stillThere); err != nil {
		t.Fatal(err)
	}
	if stillThere != 1 {
		t.Fatal("the boot migrator deleted an entry with no directory_catalog_v2 flag")
	}
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM tags WHERE id = $1`, tagID).Scan(&stillThere); err != nil {
		t.Fatal(err)
	}
	if stillThere != 1 {
		t.Fatal("the boot migrator deleted a tag with no directory_catalog_v2 flag")
	}

	var categoriesAfter, typesAfter int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_categories`).Scan(&categoriesAfter); err != nil {
		t.Fatal(err)
	}
	if categoriesAfter < categoriesBefore {
		t.Fatalf("entry_categories = %d, was %d", categoriesAfter, categoriesBefore)
	}
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_types`).Scan(&typesAfter); err != nil {
		t.Fatal(err)
	}
	if typesAfter != typesBefore {
		t.Fatalf("entry_types = %d, was %d", typesAfter, typesBefore)
	}

	// The marker stays the hand-run migration's business.
	var flagRows int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM site_settings WHERE key = 'directory_catalog_v2'`).Scan(&flagRows); err != nil {
		t.Fatal(err)
	}
	if flagRows != 0 {
		t.Fatal("the boot migrator wrote directory_catalog_v2 back")
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
	if n != 94 {
		t.Fatalf("categories = %d, want 94", n)
	}
	var sportpalya int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM entry_categories WHERE name = 'Sportpálya'`).Scan(&sportpalya); err != nil {
		t.Fatal(err)
	}
	if sportpalya != 0 {
		t.Fatal("Sportpálya must stay a venue, not a directory category")
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
	// Drop any stray copy first. The name is unique, so the seed row cannot go
	// back while a duplicate is still there.
	if _, err := db.DB.Exec(`
		DELETE FROM entry_categories c
		WHERE c.name = 'Étterem' AND c.id != 11
		AND NOT EXISTS (SELECT 1 FROM entries e WHERE e.category_id = c.id)`); err != nil {
		t.Fatal(err)
	}
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
	// Drop the probe rows even when an assertion below fails, so the next run
	// does not hit unique_category_name on the recreate above.
	t.Cleanup(func() {
		if _, err := db.DB.Exec(`DELETE FROM entries WHERE name = 'Próba étterem'`); err != nil {
			t.Errorf("cleanup entry: %v", err)
		}
		restoreSeedEtteremCategory(t)
	})
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

func TestAdminTags(t *testing.T) {
	handlers.MigrateDirectoryCatalog()
	const name = "admin-tag-probe"
	t.Cleanup(func() {
		db.DB.Exec(`DELETE FROM tags WHERE name = $1 OR name = $2`, name, name+"-2")
	})
	rr := doRequest(t, "POST", "/api/admin/tags", map[string]string{"name": "Bútor"})
	if rr.Code != 400 {
		t.Fatalf("category-shaped tag: %d %s", rr.Code, rr.Body.String())
	}
	rr = doRequest(t, "POST", "/api/admin/tags", map[string]string{"name": name})
	if rr.Code != 201 {
		t.Fatalf("create tag: %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	rr = doRequest(t, "POST", "/api/admin/tags", map[string]string{"name": name})
	if rr.Code != 409 {
		t.Fatalf("duplicate tag: %d %s", rr.Code, rr.Body.String())
	}
	rr = doRequest(t, "PUT", "/api/admin/tags", map[string]interface{}{"id": created.ID, "name": name + "-2"})
	if rr.Code != 200 {
		t.Fatalf("rename tag: %d %s", rr.Code, rr.Body.String())
	}
	rr = doRequest(t, "DELETE", "/api/admin/tags?id="+strconv.Itoa(created.ID), nil)
	if rr.Code != 200 {
		t.Fatalf("delete tag: %d %s", rr.Code, rr.Body.String())
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
