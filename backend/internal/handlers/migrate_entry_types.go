package handlers

import (
	"backend/internal/db"
	"database/sql"
	"log"
)

// MigrateEntryTypes is a legacy migrator. It no-ops once directory_catalog_v2 is set.
func MigrateEntryTypes() {
	var flag string
	err := db.DB.QueryRow(`SELECT value FROM site_settings WHERE key = $1`, directoryCatalogV2Key).Scan(&flag)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("MigrateEntryTypes (flag check): %v", err)
		return
	}
	if err == nil && flag == "1" {
		log.Println("skipped: directory_catalog_v2")
		return
	}
}
