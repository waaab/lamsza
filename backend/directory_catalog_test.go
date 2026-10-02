package main

import (
	"backend/internal/db"
	"backend/internal/handlers"
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
