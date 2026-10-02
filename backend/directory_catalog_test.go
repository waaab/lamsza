package main

import (
	"backend/internal/db"
	"backend/internal/handlers"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

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
