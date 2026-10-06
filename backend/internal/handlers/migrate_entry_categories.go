package handlers

import (
	"backend/internal/db"
	"database/sql"
	"errors"
	"log"
)

// MigrateEntryCategories is a legacy migrator. It no-ops once directory_catalog_v2 is set.
func MigrateEntryCategories() {
	var flag string
	err := db.DB.QueryRow(`SELECT value FROM site_settings WHERE key = $1`, directoryCatalogV2Key).Scan(&flag)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("MigrateEntryCategories (flag check): %v", err)
		return
	}
	if err == nil && flag == "1" {
		log.Println("skipped: directory_catalog_v2")
		return
	}
}

var errCategoryNotLeaf = errors.New("category must be a subcategory")
